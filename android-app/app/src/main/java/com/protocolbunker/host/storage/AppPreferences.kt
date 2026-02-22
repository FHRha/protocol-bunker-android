package com.protocolbunker.host.storage

import android.content.Context

class AppPreferences(context: Context) {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun port(): Int = prefs.getInt(KEY_PORT, DEFAULT_PORT)

    fun devMode(): Boolean = prefs.getBoolean(KEY_DEV_MODE, false)

    fun themeMode(): Int = prefs.getInt(KEY_THEME_MODE, THEME_DARK)

    fun autoRestartEnabled(): Boolean = prefs.getBoolean(KEY_AUTO_RESTART, false)

    fun save(port: Int, devMode: Boolean) {
        prefs.edit()
            .putInt(KEY_PORT, port)
            .putBoolean(KEY_DEV_MODE, devMode)
            .apply()
    }

    fun setThemeMode(mode: Int) {
        prefs.edit().putInt(KEY_THEME_MODE, mode).apply()
    }

    fun setAutoRestartEnabled(value: Boolean) {
        prefs.edit().putBoolean(KEY_AUTO_RESTART, value).apply()
    }

    companion object {
        const val DEFAULT_PORT = 8080
        const val THEME_LIGHT = 0
        const val THEME_DARK = 1

        private const val PREFS_NAME = "protocol_bunker_host_config"
        private const val KEY_PORT = "PORT"
        private const val KEY_DEV_MODE = "DEV_MODE"
        private const val KEY_THEME_MODE = "THEME_MODE"
        private const val KEY_AUTO_RESTART = "AUTO_RESTART"
    }
}
