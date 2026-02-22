package com.protocolbunker.host.server

import android.content.Context
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.BufferedReader
import java.io.BufferedWriter
import java.io.InputStreamReader
import java.io.OutputStreamWriter
import java.net.InetSocketAddress
import java.net.ServerSocket
import java.net.Socket
import java.nio.charset.StandardCharsets

internal class MiniHttpServerBackend : ServerBackend {
    override val name: String = "Встроенный JVM mock"

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var acceptJob: Job? = null
    private var serverSocket: ServerSocket? = null

    override suspend fun start(
        context: Context,
        port: Int,
        devMode: Boolean,
        logger: (String) -> Unit,
        onProcessTerminated: ((Int) -> Unit)?
    ) = withContext(Dispatchers.IO) {
        if (serverSocket != null) {
            logger("Mock сервер уже запущен")
            return@withContext
        }

        val socket = ServerSocket()
        socket.reuseAddress = true
        socket.bind(InetSocketAddress(port))
        serverSocket = socket

        acceptJob = scope.launch {
            logger("Mock сервер слушает 0.0.0.0:$port")
            while (isActive) {
                val client = try {
                    socket.accept()
                } catch (_: Exception) {
                    if (socket.isClosed) break
                    continue
                }
                launch { handleClient(client, port, logger) }
            }
        }
    }

    override suspend fun stop(logger: (String) -> Unit) = withContext(Dispatchers.IO) {
        serverSocket?.close()
        serverSocket = null
        acceptJob?.cancelAndJoin()
        acceptJob = null
        logger("Mock сервер остановлен")
    }

    private fun handleClient(socket: Socket, port: Int, logger: (String) -> Unit) {
        socket.use { client ->
            val reader = BufferedReader(InputStreamReader(client.getInputStream(), StandardCharsets.UTF_8))
            val writer = BufferedWriter(OutputStreamWriter(client.getOutputStream(), StandardCharsets.UTF_8))

            val requestLine = reader.readLine() ?: return
            val path = requestLine.split(" ").getOrNull(1)?.substringBefore('?') ?: "/"
            while (true) {
                val header = reader.readLine() ?: break
                if (header.isBlank()) break
            }

            val response = when (path) {
                "/" -> HttpResponse(
                    200,
                    "text/html; charset=utf-8",
                    "<!doctype html><html><body><h2>server works</h2></body></html>"
                )

                "/health" -> HttpResponse(
                    200,
                    "application/json; charset=utf-8",
                    """{"status":"ok","service":"protocol-bunker-host","port":$port,"mode":"lan_only"}"""
                )

                "/api/scenarios" -> HttpResponse(
                    200,
                    "application/json; charset=utf-8",
                    """[{"id":"classic","name":"Classic Bunker","description":"mock"}]"""
                )

                else -> HttpResponse(
                    404,
                    "application/json; charset=utf-8",
                    """{"error":"not_found","path":"${escapeJson(path)}"}"""
                )
            }

            val bodyBytes = response.body.toByteArray(StandardCharsets.UTF_8)
            writer.write("HTTP/1.1 ${response.code} ${reasonPhrase(response.code)}\r\n")
            writer.write("Content-Type: ${response.contentType}\r\n")
            writer.write("Content-Length: ${bodyBytes.size}\r\n")
            writer.write("Connection: close\r\n")
            writer.write("\r\n")
            writer.write(response.body)
            writer.flush()

            logger("HTTP $path -> ${response.code}")
        }
    }

    private fun reasonPhrase(code: Int): String = when (code) {
        200 -> "OK"
        404 -> "Not Found"
        else -> "OK"
    }

    private fun escapeJson(value: String): String =
        value.replace("\\", "\\\\").replace("\"", "\\\"")

    private data class HttpResponse(val code: Int, val contentType: String, val body: String)
}
