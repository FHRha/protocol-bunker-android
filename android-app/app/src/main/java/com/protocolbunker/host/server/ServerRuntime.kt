package com.protocolbunker.host.server

import android.content.Context
import com.protocolbunker.host.util.NetworkUtils
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.time.LocalTime
import java.time.format.DateTimeFormatter

data class ServerState(
    val running: Boolean = false,
    val status: String = "остановлен",
    val port: Int = 8080,
    val lanUrl: String = "http://127.0.0.1:8080",
    val backendName: String = "не выбран"
)

object ServerRuntime {
    private const val MAX_LOG_LINES = 300

    private val formatter = DateTimeFormatter.ofPattern("HH:mm:ss")
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val mutex = Mutex()

    private val _state = MutableStateFlow(ServerState())
    val state = _state.asStateFlow()

    private val _logs = MutableStateFlow<List<String>>(emptyList())
    val logs = _logs.asStateFlow()

    private var activeBackend: ServerBackend? = null
    private var startInProgress: Boolean = false

    fun start(context: Context, port: Int, devMode: Boolean) {
        scope.launch {
            mutex.withLock {
                if (_state.value.running || startInProgress) {
                    appendLog("Сервер уже запущен или запускается")
                    return@withLock
                }
                startInProgress = true

                _state.value = _state.value.copy(
                    running = false,
                    status = "запуск...",
                    port = port,
                    lanUrl = buildLanUrl(port)
                )

                val backend = try {
                    resolveBackend(context, devMode)
                } catch (e: Exception) {
                    val reason = classifyStartFailure(e)
                    startInProgress = false
                    _state.value = _state.value.copy(
                        running = false,
                        status = "ошибка",
                        port = port,
                        lanUrl = buildLanUrl(port),
                        backendName = "не выбран"
                    )
                    appendLog("Ошибка подготовки backend: $reason")
                    return@withLock
                }

                appendLog("Запуск backend: ${backend.name} на порту $port")

                try {
                    backend.start(
                        context = context,
                        port = port,
                        devMode = devMode,
                        logger = ::appendLog,
                        onProcessTerminated = { code ->
                            handleUnexpectedProcessExit(code, backend.name, port)
                        }
                    )
                    activeBackend = backend
                    _state.value = _state.value.copy(
                        running = true,
                        status = "запущен",
                        port = port,
                        lanUrl = buildLanUrl(port),
                        backendName = if (devMode) "${backend.name} (DEV)" else backend.name
                    )
                    appendLog("Сервер запущен")
                } catch (e: Exception) {
                    val reason = classifyStartFailure(e)
                    activeBackend = null
                    _state.value = _state.value.copy(
                        running = false,
                        status = "ошибка",
                        port = port,
                        lanUrl = buildLanUrl(port),
                        backendName = backend.name
                    )
                    appendLog("Ошибка запуска: $reason")
                } finally {
                    startInProgress = false
                }
            }
        }
    }

    fun stop() {
        scope.launch {
            mutex.withLock {
                startInProgress = false
                val backend = activeBackend
                if (backend == null) {
                    _state.value = _state.value.copy(running = false, status = "остановлен")
                    appendLog("Сервер уже остановлен")
                    return@withLock
                }

                appendLog("Остановка backend: ${backend.name}")
                try {
                    backend.stop(::appendLog)
                } catch (e: Exception) {
                    appendLog("Ошибка при остановке: ${e.message}")
                } finally {
                    activeBackend = null
                    _state.value = _state.value.copy(
                        running = false,
                        status = "остановлен",
                        backendName = "не выбран"
                    )
                    appendLog("Сервер остановлен")
                }
            }
        }
    }

    fun refreshLanUrl() {
        scope.launch {
            mutex.withLock {
                val current = _state.value
                if (!current.running) return@withLock
                val nextUrl = buildLanUrl(current.port)
                if (nextUrl != current.lanUrl) {
                    _state.value = current.copy(lanUrl = nextUrl)
                    appendLog("Обновлен LAN URL: $nextUrl")
                }
            }
        }
    }

    private fun handleUnexpectedProcessExit(exitCode: Int, backendName: String, port: Int) {
        scope.launch {
            mutex.withLock {
                if (activeBackend == null) return@withLock
                activeBackend = null
                startInProgress = false
                _state.value = _state.value.copy(
                    running = false,
                    status = "ошибка",
                    port = port,
                    lanUrl = buildLanUrl(port),
                    backendName = backendName
                )
                appendLog("Процесс сервера завершился аварийно. Код: $exitCode")
            }
        }
    }

    private suspend fun resolveBackend(context: Context, devMode: Boolean): ServerBackend {
        val installed = runCatching { ServerBinaryInstaller.ensureInstalled(context, ::appendLog) }.getOrElse {
            if (!devMode) {
                throw IllegalStateException("Не удалось установить Go-бинарь: ${it.message}", it)
            }
            appendLog("DEV_MODE: установка Go-бинаря не удалась, включаю аварийный mock fallback")
            return MiniHttpServerBackend()
        }
        if (!installed) {
            if (devMode) {
                appendLog("DEV_MODE: Go-бинарь недоступен для текущей ABI, включаю аварийный mock fallback")
                return MiniHttpServerBackend()
            }
            throw IllegalStateException("Go-бинарь недоступен для текущей ABI.")
        }
        if (!ExternalProcessServerBackend.isBinaryAvailable(context)) {
            if (devMode) {
                appendLog("DEV_MODE: бинарь не найден после установки, включаю аварийный mock fallback")
                return MiniHttpServerBackend()
            }
            throw IllegalStateException("Go-бинарь не найден после установки.")
        }
        if (devMode) {
            appendLog("DEV_MODE: запускаю реальный Go backend (dev_tab + dev scenarios)")
        }
        return ExternalProcessServerBackend()
    }

    private fun buildLanUrl(port: Int): String {
        val ip = NetworkUtils.findLanIpv4() ?: "127.0.0.1"
        return "http://$ip:$port"
    }

    private fun appendLog(message: String) {
        val timestamp = LocalTime.now().format(formatter)
        val newLine = "$timestamp | $message"
        val next = _logs.value.toMutableList()
        next.add(newLine)
        if (next.size > MAX_LOG_LINES) {
            repeat(next.size - MAX_LOG_LINES) { next.removeAt(0) }
        }
        _logs.value = next
    }

    private fun classifyStartFailure(error: Throwable): String {
        val text = error.message.orEmpty()
        val lower = text.lowercase()
        return when {
            "address already in use" in lower || "bind:" in lower ->
                "Порт занят. Выберите другой PORT."
            "exec format error" in lower || "not executable" in lower ->
                "Go-бинарь не совместим с ABI устройства или повреждён."
            "permission denied" in lower ->
                "Нет прав на запуск бинаря."
            text.isNotBlank() -> text
            else -> "Неизвестная ошибка запуска."
        }
    }
}
