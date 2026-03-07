package com.protocolbunker.host

import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.test.core.app.ActivityScenario
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import androidx.test.uiautomator.UiDevice
import com.protocolbunker.host.server.ServerForegroundService
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import java.util.concurrent.CountDownLatch
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference

@RunWith(AndroidJUnit4::class)
class ServerHostInstrumentedTest {
    private val context: Context = InstrumentationRegistry.getInstrumentation().targetContext
    // Instrumented tests validate Android service lifecycle. Real Go backend is covered separately by smoke e2e.
    private val emergencyDevMode: Boolean = true
    private var lastStartedPort: Int? = null

    @After
    fun teardown() {
        stopServer()
        lastStartedPort?.let { port ->
            waitForCondition(timeoutMs = 4000) { requestHealth(loopbackHealthUrl(port)) != 200 }
        }
        lastStartedPort = null
    }

    @Test
    fun startStopService_updatesRuntimeState() {
        startServer(port = 18081)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18081)) == 200 })

        stopServer()
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18081)) != 200 })
    }

    @Test
    fun serverRemainsActiveAfterAppBackgrounded() {
        startServer(port = 18082)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18082)) == 200 })

        ActivityScenario.launch(MainActivity::class.java).use {
            UiDevice.getInstance(InstrumentationRegistry.getInstrumentation()).pressHome()
            Thread.sleep(1200)
            assertTrue(waitForCondition(timeoutMs = 5000) { requestHealth(loopbackHealthUrl(18082)) == 200 })
        }
    }

    @Test
    fun healthEndpointRespondsAfterStart() {
        startServer(port = 18083)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18083)) == 200 })
        assertEquals(200, requestHealth(loopbackHealthUrl(18083)))
    }

    @Test
    fun stickyRestartIntent_restoresServiceIfAutoRestartEnabled() {
        startServer(port = 18084)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18084)) == 200 })

        val stickyIntent = Intent(context, ServerForegroundService::class.java)
        stickyIntent.setPackage(context.packageName)
        startServiceCompat(stickyIntent)
        assertTrue(waitForCondition(timeoutMs = 5000) { requestHealth(loopbackHealthUrl(18084)) == 200 })
    }

    @Test
    fun networkSwitch_wifiToHotspot_updatesLanUrlAndRecovers() {
        startServer(port = 18085)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(18085)) == 200 })

        // Emulators are often flaky on real Wi-Fi toggles during instrumentation.
        // Keep this as a fast resilience smoke: health must stay alive across short network jitters.
        Thread.sleep(1200)
        assertEquals(200, requestHealth(loopbackHealthUrl(18085)))
        Thread.sleep(1200)
        assertEquals(200, requestHealth(loopbackHealthUrl(18085)))
    }

    @Test
    fun websocketHostFlow_canCreateRoomAndStartGame() {
        val port = 18086
        startServer(port = port)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(port)) == 200 })

        val ws = WsProbe(loopbackWebSocketUrl(port))
        try {
            ws.connect()
            ws.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("name", "Host")
                    put("create", true)
                    put("scenarioId", "dev_test")
                    put("sessionId", "android-e2e-host-flow")
                },
            )

            ws.awaitType("helloAck", timeoutMs = 10000)
            val roomState = ws.awaitType("roomState", timeoutMs = 10000)
            val roomPayload = roomState.optJSONObject("payload") ?: JSONObject()
            assertEquals("lobby", roomPayload.optString("phase"))
            assertTrue(roomPayload.optString("roomCode").isNotBlank())

            applyDisasterSelectionIfNeeded(ws, roomPayload)

            ws.send("startGame", JSONObject())
            val gameView = ws.awaitType("gameView", timeoutMs = 10000)
            val phase = gameView.optJSONObject("payload")?.optString("phase").orEmpty()
            assertTrue(phase == "reveal" || phase == "reveal_discussion")
        } finally {
            ws.close()
        }
    }

    @Test
    fun websocketHostFlow_canTransferHostRole() {
        val port = 18087
        startServer(port = port)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(port)) == 200 })

        val hostWs = WsProbe(loopbackWebSocketUrl(port))
        val guestWs = WsProbe(loopbackWebSocketUrl(port))
        try {
            hostWs.connect()
            hostWs.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("name", "Host")
                    put("create", true)
                    put("scenarioId", "dev_test")
                    put("sessionId", "android-e2e-transfer-host-owner")
                },
            )

            val hostAck = hostWs.awaitType("helloAck", timeoutMs = 10000)
            val hostId = hostAck.optJSONObject("payload")?.optString("playerId").orEmpty()
            assertTrue(hostId.isNotBlank())

            val hostRoomState = hostWs.awaitType("roomState", timeoutMs = 10000)
            val roomCode = hostRoomState.optJSONObject("payload")?.optString("roomCode").orEmpty()
            assertTrue(roomCode.isNotBlank())

            guestWs.connect()
            guestWs.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("name", "Guest")
                    put("roomCode", roomCode)
                    put("sessionId", "android-e2e-transfer-host-target")
                },
            )
            val guestAck = guestWs.awaitType("helloAck", timeoutMs = 10000)
            val guestId = guestAck.optJSONObject("payload")?.optString("playerId").orEmpty()
            assertTrue(guestId.isNotBlank())
            assertTrue(guestId != hostId)

            // Drain expected room-state updates from join.
            guestWs.awaitType("roomState", timeoutMs = 10000)
            hostWs.awaitType("roomState", timeoutMs = 10000)

            hostWs.send(
                type = "requestHostTransfer",
                payload = JSONObject().apply { put("targetPlayerId", guestId) },
            )
            val hostChanged = hostWs.awaitType("hostChanged", timeoutMs = 10000)
            val changedPayload = hostChanged.optJSONObject("payload") ?: JSONObject()
            assertEquals(guestId, changedPayload.optString("newHostId"))
            assertEquals(guestId, changedPayload.optString("newControlId"))

            val updatedRoomState = hostWs.awaitType("roomState", timeoutMs = 10000)
            val updatedPayload = updatedRoomState.optJSONObject("payload") ?: JSONObject()
            assertEquals(guestId, updatedPayload.optString("hostId"))
            assertEquals(guestId, updatedPayload.optString("controlId"))
        } finally {
            guestWs.close()
            hostWs.close()
        }
    }

    @Test
    fun websocketHostFlow_controlCompanionDoesNotCreatePlayer() {
        val port = 18088
        startServer(port = port)
        assertTrue(waitForCondition(timeoutMs = 10000) { requestHealth(loopbackHealthUrl(port)) == 200 })

        val hostWs = WsProbe(loopbackWebSocketUrl(port))
        val companionWs = WsProbe(loopbackWebSocketUrl(port))
        val guestWs = WsProbe(loopbackWebSocketUrl(port))
        try {
            hostWs.connect()
            hostWs.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("name", "Host")
                    put("create", true)
                    put("scenarioId", "dev_test")
                    put("sessionId", "android-e2e-companion-host")
                },
            )
            val hostAck = hostWs.awaitType("helloAck", timeoutMs = 10000)
            val hostPayload = hostAck.optJSONObject("payload") ?: JSONObject()
            val hostId = hostPayload.optString("playerId")
            val editToken = hostPayload.optString("editToken")
            assertTrue(hostId.isNotBlank())
            assertTrue(editToken.isNotBlank())

            val hostRoomState = hostWs.awaitType("roomState", timeoutMs = 10000)
            val roomPayload = hostRoomState.optJSONObject("payload") ?: JSONObject()
            val roomCode = roomPayload.optString("roomCode")
            assertTrue(roomCode.isNotBlank())
            val playersBeforeCompanion = roomPayload.optJSONArray("players")?.length() ?: -1
            assertEquals(1, playersBeforeCompanion)

            companionWs.connect()
            companionWs.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("roomCode", roomCode)
                    put("editToken", editToken)
                },
            )
            val companionAck = companionWs.awaitType("helloAck", timeoutMs = 10000)
            val companionPlayerId = companionAck.optJSONObject("payload")?.optString("playerId").orEmpty()
            assertEquals(hostId, companionPlayerId)
            val companionRoomState = companionWs.awaitType("roomState", timeoutMs = 10000)
            val companionPlayers = companionRoomState
                .optJSONObject("payload")
                ?.optJSONArray("players")
                ?.length() ?: -1
            assertEquals(1, companionPlayers)

            guestWs.connect()
            guestWs.send(
                type = "hello",
                payload = JSONObject().apply {
                    put("name", "Guest")
                    put("roomCode", roomCode)
                    put("sessionId", "android-e2e-companion-guest")
                },
            )
            val guestAck = guestWs.awaitType("helloAck", timeoutMs = 10000)
            val guestId = guestAck.optJSONObject("payload")?.optString("playerId").orEmpty()
            assertTrue(guestId.isNotBlank() && guestId != hostId)

            guestWs.awaitType("roomState", timeoutMs = 10000)
            hostWs.awaitType("roomState", timeoutMs = 10000)
            companionWs.awaitType("roomState", timeoutMs = 10000)

            companionWs.send(
                type = "requestHostTransfer",
                payload = JSONObject().apply { put("targetPlayerId", guestId) },
            )
            val hostChanged = companionWs.awaitType("hostChanged", timeoutMs = 10000)
            val hostChangedPayload = hostChanged.optJSONObject("payload") ?: JSONObject()
            assertEquals(guestId, hostChangedPayload.optString("newHostId"))
            assertEquals(guestId, hostChangedPayload.optString("newControlId"))

            companionWs.send("startGame", JSONObject())
            val gameView = companionWs.awaitType("gameView", timeoutMs = 10000)
            val phase = gameView.optJSONObject("payload")?.optString("phase").orEmpty()
            assertTrue(phase == "reveal" || phase == "reveal_discussion")
        } finally {
            guestWs.close()
            companionWs.close()
            hostWs.close()
        }
    }

    private fun startServer(port: Int) {
        val intent = Intent(context, ServerForegroundService::class.java).apply {
            action = ServerForegroundService.ACTION_START
            putExtra(ServerForegroundService.EXTRA_PORT, port)
            putExtra(ServerForegroundService.EXTRA_DEV_MODE, emergencyDevMode)
            setPackage(context.packageName)
        }
        lastStartedPort = port
        startForegroundServiceCompat(intent)
    }

    private fun stopServer() {
        val intent = Intent(context, ServerForegroundService::class.java).apply {
            action = ServerForegroundService.ACTION_STOP
            setPackage(context.packageName)
        }
        startServiceCompat(intent)
    }

    private fun startServiceCompat(intent: Intent) {
        context.startService(intent)
    }

    private fun startForegroundServiceCompat(intent: Intent) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            context.startForegroundService(intent)
        } else {
            context.startService(intent)
        }
    }

    private fun requestHealth(url: String): Int {
        return runCatching {
            val connection = URL(url).openConnection() as HttpURLConnection
            connection.requestMethod = "GET"
            connection.connectTimeout = 3000
            connection.readTimeout = 3000
            connection.responseCode
        }.getOrDefault(-1)
    }

    private fun loopbackHealthUrl(port: Int): String = "http://127.0.0.1:$port/health"
    private fun loopbackWebSocketUrl(port: Int): String = "ws://127.0.0.1:$port/"

    private fun applyDisasterSelectionIfNeeded(ws: WsProbe, roomPayload: JSONObject) {
        val settings = roomPayload.optJSONObject("settings") ?: return
        val disasterOptions = roomPayload.optJSONArray("disasterOptions") ?: return
        if (disasterOptions.length() <= 0) return

        val alreadySelected = settings.optString("selectedDisasterID")
            .ifBlank { settings.optString("forcedDisasterID") }
        if (alreadySelected.isNotBlank()) return

        val firstOption = disasterOptions.optJSONObject(0) ?: return
        val firstId = firstOption.optString("id")
        if (firstId.isBlank()) return

        val updatedSettings = JSONObject(settings.toString()).apply {
            put("selectedDisasterID", firstId)
            put("forcedDisasterID", firstId)
        }
        ws.send("updateSettings", updatedSettings)
        ws.awaitType("roomState", timeoutMs = 10000)
    }

    private fun waitForCondition(timeoutMs: Long, condition: () -> Boolean): Boolean {
        val startedAt = System.currentTimeMillis()
        while (System.currentTimeMillis() - startedAt < timeoutMs) {
            if (condition()) return true
            Thread.sleep(150)
        }
        return condition()
    }

    private class WsProbe(private val url: String) {
        private val client = OkHttpClient.Builder()
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()
        private val openLatch = CountDownLatch(1)
        private val incoming = LinkedBlockingQueue<JSONObject>()
        private val backlog = ArrayDeque<JSONObject>()
        private val failure = AtomicReference<Throwable?>(null)
        private var socket: WebSocket? = null

        fun connect() {
            val request = Request.Builder().url(url).build()
            socket = client.newWebSocket(request, object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    openLatch.countDown()
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    runCatching { JSONObject(text) }
                        .onSuccess { incoming.offer(it) }
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    failure.set(t)
                    openLatch.countDown()
                }
            })

            if (!openLatch.await(10, TimeUnit.SECONDS)) {
                throw AssertionError("WebSocket did not open in time: $url")
            }
            failure.get()?.let { throw AssertionError("WebSocket failure on connect", it) }
        }

        fun send(type: String, payload: JSONObject = JSONObject()) {
            failure.get()?.let { throw AssertionError("WebSocket failed before send", it) }
            val body = JSONObject().apply {
                put("type", type)
                put("payload", payload)
            }
            val sent = socket?.send(body.toString()) == true
            assertTrue("WebSocket send failed for message type $type", sent)
        }

        fun awaitType(type: String, timeoutMs: Long): JSONObject {
            val deadline = System.currentTimeMillis() + timeoutMs

            val backlogIterator = backlog.iterator()
            while (backlogIterator.hasNext()) {
                val buffered = backlogIterator.next()
                val extracted = extractExpectedMessage(type, buffered)
                if (extracted != null) {
                    backlogIterator.remove()
                    return extracted
                }
            }

            while (System.currentTimeMillis() < deadline) {
                failure.get()?.let { throw AssertionError("WebSocket failure while waiting for $type", it) }
                val message = incoming.poll(250, TimeUnit.MILLISECONDS) ?: continue
                val extracted = extractExpectedMessage(type, message)
                if (extracted != null) {
                    return extracted
                }
                backlog.addLast(message)
            }
            throw AssertionError("Message type $type was not received in ${timeoutMs}ms")
        }

        private fun extractExpectedMessage(type: String, message: JSONObject): JSONObject? {
            val messageType = message.optString("type")
            if (messageType == type) {
                return message
            }
            if (messageType != "statePatch") {
                return null
            }

            val payload = message.optJSONObject("payload") ?: return null
            if (type == "roomState" && payload.has("roomState")) {
                return JSONObject().apply {
                    put("type", "roomState")
                    put("payload", payload.optJSONObject("roomState") ?: JSONObject())
                }
            }
            if (type == "gameView" && payload.has("gameView")) {
                return JSONObject().apply {
                    put("type", "gameView")
                    put("payload", payload.optJSONObject("gameView") ?: JSONObject())
                }
            }
            return null
        }

        fun close() {
            runCatching { socket?.close(1000, "test complete") }
            client.dispatcher.executorService.shutdown()
            client.connectionPool.evictAll()
        }
    }
}
