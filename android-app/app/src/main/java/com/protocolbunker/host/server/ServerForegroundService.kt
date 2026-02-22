package com.protocolbunker.host.server

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat
import com.protocolbunker.host.MainActivity
import com.protocolbunker.host.R
import com.protocolbunker.host.storage.AppPreferences
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class ServerForegroundService : Service() {
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    private lateinit var prefs: AppPreferences
    private var commandReceived: Boolean = false

    override fun onCreate() {
        super.onCreate()
        prefs = AppPreferences(applicationContext)
        createNotificationChannel()
        serviceScope.launch {
            ServerRuntime.state.collectLatest { state ->
                if (state.running) {
                    notificationManager().notify(
                        NOTIFICATION_ID,
                        buildRunningNotification()
                    )
                } else {
                    if (!commandReceived) return@collectLatest
                    val keepForeground = prefs.autoRestartEnabled() && state.status.startsWith("запуск")
                    if (keepForeground) {
                        notificationManager().notify(
                            NOTIFICATION_ID,
                            buildStartingNotification()
                        )
                        return@collectLatest
                    }
                    stopForeground(STOP_FOREGROUND_REMOVE)
                    clearRunningNotification()
                    if (state.status.startsWith("ошибка")) {
                        prefs.setAutoRestartEnabled(false)
                        stopSelf()
                        return@collectLatest
                    }
                    if (!prefs.autoRestartEnabled()) {
                        stopSelf()
                    }
                }
            }
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        commandReceived = true
        when (intent?.action) {
            ACTION_START -> {
                val portExtra = intent.getIntExtra(EXTRA_PORT, prefs.port())
                val port = portExtra.coerceIn(1, 65535)
                val devMode = intent.getBooleanExtra(EXTRA_DEV_MODE, prefs.devMode())

                prefs.save(port, devMode)
                prefs.setAutoRestartEnabled(true)

                if (!startForegroundSafe()) {
                    prefs.setAutoRestartEnabled(false)
                    return START_NOT_STICKY
                }

                ServerRuntime.start(applicationContext, port, devMode)
                return START_STICKY
            }

            ACTION_STOP -> {
                Log.i(TAG, "ACTION_STOP from notification")
                prefs.setAutoRestartEnabled(false)
                ServerRuntime.stop()
                stopForeground(STOP_FOREGROUND_REMOVE)
                clearRunningNotification()
                stopSelf()
                return START_NOT_STICKY
            }

            else -> {
                if (prefs.autoRestartEnabled()) {
                    val port = prefs.port().coerceIn(1, 65535)
                    val devMode = prefs.devMode()

                    if (!startForegroundSafe()) {
                        prefs.setAutoRestartEnabled(false)
                        return START_NOT_STICKY
                    }

                    ServerRuntime.start(applicationContext, port, devMode)
                    return START_STICKY
                }
            }
        }
        return START_NOT_STICKY
    }

    override fun onTaskRemoved(rootIntent: Intent?) {
        prefs.setAutoRestartEnabled(false)
        ServerRuntime.stop()
        stopForeground(STOP_FOREGROUND_REMOVE)
        clearRunningNotification()
        stopSelf()
        super.onTaskRemoved(rootIntent)
    }

    override fun onDestroy() {
        clearRunningNotification()
        serviceScope.cancel()
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun startForegroundSafe(): Boolean {
        return runCatching {
            startForeground(NOTIFICATION_ID, buildStartingNotification())
        }.onFailure {
            Log.e(TAG, "Failed to start foreground notification", it)
        }.isSuccess
    }

    private fun buildStartingNotification(): Notification {
        return buildBaseNotification(
            text = getString(R.string.notification_starting_text)
        )
    }

    private fun buildRunningNotification(): Notification {
        return buildBaseNotification(
            text = getString(R.string.notification_text)
        )
    }

    private fun buildBaseNotification(text: String, subText: String? = null): Notification {
        val openIntent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            openIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        val stopIntent = Intent(this, ServerForegroundService::class.java).apply {
            action = ACTION_STOP
        }
        val stopPendingIntent = PendingIntent.getService(
            this,
            101,
            stopIntent,
            PendingIntent.FLAG_CANCEL_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        val builder = NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle(getString(R.string.notification_title))
            .setContentText(text)
            .setSmallIcon(R.drawable.ic_notification_small)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setSilent(true)
            .setShowWhen(false)
            .setCategory(NotificationCompat.CATEGORY_SERVICE)
            .setContentIntent(pendingIntent)
            .addAction(0, getString(R.string.notification_action_stop), stopPendingIntent)
        if (!subText.isNullOrBlank()) {
            builder.setSubText(subText)
        }
        return builder.build()
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val channel = NotificationChannel(
            CHANNEL_ID,
            getString(R.string.notification_channel_name),
            NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = getString(R.string.notification_channel_description)
        }
        notificationManager().createNotificationChannel(channel)
    }

    private fun notificationManager(): NotificationManager =
        getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

    private fun clearRunningNotification() {
        notificationManager().cancel(NOTIFICATION_ID)
    }

    companion object {
        const val ACTION_START = "com.protocolbunker.host.START_SERVER"
        const val ACTION_STOP = "com.protocolbunker.host.STOP_SERVER"

        const val EXTRA_PORT = "extra_port"
        const val EXTRA_DEV_MODE = "extra_dev_mode"

        const val DEFAULT_PORT = 8080

        private const val CHANNEL_ID = "protocol_bunker_host_channel"
        private const val NOTIFICATION_ID = 1001
        private const val TAG = "ServerForegroundSvc"
    }
}
