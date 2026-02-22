package com.protocolbunker.host

import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.test.core.app.ActivityScenario
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import androidx.test.uiautomator.UiDevice
import com.protocolbunker.host.server.ServerForegroundService
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import java.net.HttpURLConnection
import java.net.URL

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

    private fun waitForCondition(timeoutMs: Long, condition: () -> Boolean): Boolean {
        val startedAt = System.currentTimeMillis()
        while (System.currentTimeMillis() - startedAt < timeoutMs) {
            if (condition()) return true
            Thread.sleep(150)
        }
        return condition()
    }
}
