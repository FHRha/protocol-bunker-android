package com.protocolbunker.host.server

import android.content.Context
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.util.Locale
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicReference

internal class ExternalProcessServerBackend : ServerBackend {
    override val name: String = "Go бинарь"

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var process: Process? = null
    private var outJob: Job? = null
    private var errJob: Job? = null
    private var waitJob: Job? = null
    private val stopping = AtomicBoolean(false)
    private val lastErrLine = AtomicReference("")

    override suspend fun start(
        context: Context,
        port: Int,
        devMode: Boolean,
        logger: (String) -> Unit,
        onProcessTerminated: ((Int) -> Unit)?
    ) = withContext(Dispatchers.IO) {
        val binary = binaryFile(context)
        require(binary.exists()) {
            "Go бинарь не найден: ${binary.absolutePath}"
        }

        if (!binary.canExecute()) {
            check(binary.setExecutable(true, false)) {
                "Не удалось выдать execute права бинарю: ${binary.absolutePath}"
            }
        }

        val assetsRoot = ServerBinaryInstaller.runtimeAssetsRoot(context)
        val clientDistRoot = ServerBinaryInstaller.runtimeClientDistRoot(context)
        val scenariosRoot = ServerBinaryInstaller.runtimeScenariosRoot(context)
        require(File(assetsRoot, "decks").isDirectory) {
            "Игровые ресурсы не найдены: ${File(assetsRoot, "decks").absolutePath}"
        }
        require(File(clientDistRoot, "index.html").isFile) {
            "Client dist не найден: ${File(clientDistRoot, "index.html").absolutePath}"
        }
        require(File(scenariosRoot, "classic/SPECIAL_CONDITIONS.json").isFile) {
            "SPECIAL_CONDITIONS.json не найден: ${File(scenariosRoot, "classic/SPECIAL_CONDITIONS.json").absolutePath}"
        }

        stopping.set(false)
        lastErrLine.set("")

        val processArgs = mutableListOf(
            binary.absolutePath,
            "-port", port.toString(),
            "-assets-root", assetsRoot.absolutePath,
            "-client-dist", clientDistRoot.absolutePath,
            "-scenarios-root", scenariosRoot.absolutePath,
            "-specials-file", File(scenariosRoot, "classic/SPECIAL_CONDITIONS.json").absolutePath,
        )
        if (devMode) {
            processArgs += "-enable-dev-scenarios"
        }
        val processBuilder = ProcessBuilder(processArgs)
        processBuilder.environment()["PORT"] = port.toString()
        processBuilder.environment()["BUNKER_ENABLE_DEV_SCENARIOS"] = if (devMode) "true" else "false"
        processBuilder.environment()["BUNKER_IDENTITY_MODE"] = if (devMode) "dev_tab" else "prod"
        processBuilder.redirectErrorStream(false)
        processBuilder.directory(context.filesDir)

        logger("Go backend config: port=$port, devMode=$devMode")
        logger("Go backend paths: assets=${assetsRoot.absolutePath}, client=${clientDistRoot.absolutePath}, scenarios=${scenariosRoot.absolutePath}")
        logger("Go command: ${processArgs.joinToString(" ")}")

        process = processBuilder.start()
        logger(
            "Go процесс запущен: ${binary.absolutePath} (mode=${if (devMode) "dev_tab" else "prod"})"
        )

        val runningProcess = process ?: return@withContext
        outJob = scope.launch {
            runCatching {
                runningProcess.inputStream.bufferedReader().forEachLine { line ->
                    logger("[go] $line")
                }
            }.onFailure { err ->
                if (!stopping.get()) {
                    logger("Ошибка чтения stdout Go процесса: ${err.message}")
                }
            }
        }
        errJob = scope.launch {
            runCatching {
                runningProcess.errorStream.bufferedReader().forEachLine { line ->
                    lastErrLine.set(line)
                    logger("[go-err] $line")
                }
            }.onFailure { err ->
                if (!stopping.get()) {
                    logger("Ошибка чтения stderr Go процесса: ${err.message}")
                }
            }
        }

        // Detect early crash (port busy, wrong ABI, corrupted binary, missing assets).
        delay(450)
        if (!runningProcess.isAlive) {
            val code = runCatching { runningProcess.exitValue() }.getOrDefault(-1)
            val stderr = lastErrLine.get()
            val hint = classifyEarlyExitHint(stderr)
            process = null
            outJob?.cancelAndJoin()
            outJob = null
            errJob?.cancelAndJoin()
            errJob = null
            throw IllegalStateException(
                buildString {
                    append("Go процесс завершился сразу (код $code)")
                    if (hint.isNotBlank()) append(": $hint")
                    if (stderr.isNotBlank()) append(". stderr: $stderr")
                }
            )
        }

        waitJob = scope.launch {
            val code = runCatching { runningProcess.waitFor() }.getOrElse { err ->
                if (!stopping.get()) {
                    logger("Ошибка ожидания Go процесса: ${err.message}")
                }
                return@launch
            }
            logger("Go процесс завершился. Код: $code")
            if (!stopping.get() && code != 0) {
                val stderr = lastErrLine.get()
                if (stderr.isNotBlank()) {
                    logger("Последняя строка stderr перед завершением: $stderr")
                }
            }
            if (!stopping.get()) {
                onProcessTerminated?.invoke(code)
            }
        }
    }

    override suspend fun stop(logger: (String) -> Unit) = withContext(Dispatchers.IO) {
        stopping.set(true)
        val runningProcess = process
        if (runningProcess == null) {
            logger("Go процесс уже остановлен")
            return@withContext
        }

        runningProcess.destroy()
        val exited = runCatching { runningProcess.waitFor(2, TimeUnit.SECONDS) }.getOrDefault(false)
        if (!exited) {
            logger("Go процесс не завершился за 2с, выполняю kill")
            runningProcess.destroyForcibly()
            runCatching { runningProcess.waitFor(1500, TimeUnit.MILLISECONDS) }
        }

        waitJob?.cancelAndJoin()
        waitJob = null
        outJob?.cancelAndJoin()
        outJob = null
        errJob?.cancelAndJoin()
        errJob = null
        process = null
        logger("Go процесс остановлен")
    }

    companion object {
        private const val BIN_RELATIVE_PATH = "server-go/server-go"

        fun isBinaryAvailable(context: Context): Boolean = binaryFile(context).exists()

        private fun binaryFile(context: Context): File = File(context.filesDir, BIN_RELATIVE_PATH)

        private fun classifyEarlyExitHint(stderr: String): String {
            val normalized = stderr.lowercase(Locale.ROOT)
            return when {
                normalized.contains("address already in use") || normalized.contains("bind:") ->
                    "порт занят другим процессом"
                normalized.contains("permission denied") ->
                    "нет прав на запуск бинаря"
                normalized.contains("exec format error") || normalized.contains("not executable") ->
                    "бинарь не подходит для ABI или повреждён"
                normalized.contains("assets decks path error") || normalized.contains("assets catalog is empty") ->
                    "не найдены игровые decks в assets"
                normalized.contains("client dist root does not contain index.html") ->
                    "не найден client/dist/index.html"
                normalized.isNotBlank() -> stderr
                else -> ""
            }
        }
    }
}
