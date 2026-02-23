package com.protocolbunker.host

import android.Manifest
import android.content.ActivityNotFoundException
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Intent
import android.content.pm.PackageManager
import android.content.res.ColorStateList
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.view.View
import android.widget.Button
import android.widget.FrameLayout
import android.widget.ImageButton
import android.widget.PopupMenu
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.appcompat.app.AppCompatDelegate
import androidx.core.content.ContextCompat
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.lifecycleScope
import androidx.lifecycle.repeatOnLifecycle
import com.google.android.material.bottomsheet.BottomSheetBehavior
import com.google.android.material.bottomsheet.BottomSheetDialog
import com.google.android.material.chip.Chip
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import com.google.android.material.switchmaterial.SwitchMaterial
import com.google.android.material.textfield.TextInputEditText
import com.protocolbunker.host.server.ServerForegroundService
import com.protocolbunker.host.server.ServerRuntime
import com.protocolbunker.host.server.ServerState
import com.protocolbunker.host.storage.AppPreferences
import com.protocolbunker.host.util.NetworkUtils
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {
    private lateinit var prefs: AppPreferences

    private lateinit var rootContainer: View
    private lateinit var navButton: ImageButton
    private lateinit var startButton: Button
    private lateinit var startHintText: TextView
    private lateinit var openBrowserButton: Button
    private lateinit var openInAppButton: Button
    private lateinit var copyLanUrlButton: Button
    private lateinit var gameInfoBlock: View
    private lateinit var connectionActionsRow: View
    private lateinit var statusText: TextView
    private lateinit var lanUrlText: TextView
    private lateinit var lanModeText: TextView
    private lateinit var backendText: TextView
    private var latestLogs: List<String> = emptyList()

    private val notificationsPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) {}

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        prefs = AppPreferences(this)
        applySavedTheme()
        setContentView(R.layout.activity_main)

        bindViews()
        applySystemInsets()
        requestNotificationsPermissionIfNeeded()

        navButton.setOnClickListener { anchor ->
            showContextMenu(anchor)
        }
        startButton.setOnClickListener {
            if (ServerRuntime.state.value.running) onStopClicked() else onStartClicked()
        }
        copyLanUrlButton.setOnClickListener {
            val url = resolveOpenUrl()
            copyTextToClipboard(url)
            Toast.makeText(this, getString(R.string.lan_url_copied), Toast.LENGTH_SHORT).show()
        }
        openBrowserButton.setOnClickListener { openInBrowser(resolveOpenUrl()) }
        openInAppButton.setOnClickListener { openInApp(resolveOpenUrl()) }

        observeRuntime()
        renderState(ServerRuntime.state.value)
        renderLogs(ServerRuntime.logs.value)
    }

    override fun onResume() {
        super.onResume()
        ServerRuntime.refreshLanUrl()
    }

    private fun bindViews() {
        rootContainer = findViewById(R.id.rootContainer)
        navButton = findViewById(R.id.navButton)
        startButton = findViewById(R.id.startButton)
        startHintText = findViewById(R.id.startHintText)
        openBrowserButton = findViewById(R.id.openBrowserButton)
        openInAppButton = findViewById(R.id.openInAppButton)
        copyLanUrlButton = findViewById(R.id.copyLanUrlButton)
        gameInfoBlock = findViewById(R.id.gameInfoBlock)
        connectionActionsRow = findViewById(R.id.connectionActionsRow)
        statusText = findViewById(R.id.statusText)
        lanUrlText = findViewById(R.id.lanUrlText)
        lanModeText = findViewById(R.id.lanModeText)
        backendText = findViewById(R.id.backendText)
    }

    private fun applySystemInsets() {
        val baseLeft = rootContainer.paddingLeft
        val baseTop = rootContainer.paddingTop
        val baseRight = rootContainer.paddingRight
        val baseBottom = rootContainer.paddingBottom
        ViewCompat.setOnApplyWindowInsetsListener(rootContainer) { view, insets ->
            val bars = insets.getInsets(WindowInsetsCompat.Type.systemBars())
            view.setPadding(
                baseLeft + bars.left,
                baseTop + bars.top,
                baseRight + bars.right,
                baseBottom + bars.bottom
            )
            insets
        }
        ViewCompat.requestApplyInsets(rootContainer)
    }

    private fun observeRuntime() {
        lifecycleScope.launch {
            repeatOnLifecycle(Lifecycle.State.STARTED) {
                launch { ServerRuntime.state.collect { renderState(it) } }
                launch { ServerRuntime.logs.collect { renderLogs(it) } }
            }
        }
    }

    private fun onStartClicked() {
        val port = prefs.port()
        if (port !in 1..65535) {
            Toast.makeText(this, getString(R.string.invalid_port), Toast.LENGTH_SHORT).show()
            return
        }
        val devMode = prefs.devMode()
        if (devMode) {
            Toast.makeText(this, getString(R.string.dev_mode_warning), Toast.LENGTH_SHORT).show()
        }

        val startIntent = Intent(this, ServerForegroundService::class.java).apply {
            action = ServerForegroundService.ACTION_START
            putExtra(ServerForegroundService.EXTRA_PORT, port)
            putExtra(ServerForegroundService.EXTRA_DEV_MODE, devMode)
        }
        ContextCompat.startForegroundService(this, startIntent)
    }

    private fun onStopClicked() {
        val stopIntent = Intent(this, ServerForegroundService::class.java).apply {
            action = ServerForegroundService.ACTION_STOP
        }
        runCatching { startService(stopIntent) }.onFailure {
            ServerRuntime.stop()
            runCatching { stopService(Intent(this, ServerForegroundService::class.java)) }
            Toast.makeText(this, getString(R.string.stop_fallback_notice), Toast.LENGTH_SHORT).show()
        }
    }

    override fun onDestroy() {
        if (isFinishing && !isChangingConfigurations) {
            onStopClicked()
        }
        super.onDestroy()
    }

    private fun renderState(state: ServerState) {
        statusText.text = getString(R.string.status_prefix, state.status)
        val statusColor = when {
            state.status.startsWith("ошибка", ignoreCase = true) -> R.color.status_error
            state.running -> R.color.status_running
            else -> R.color.status_stopped
        }
        statusText.setTextColor(ContextCompat.getColor(this, statusColor))
        lanUrlText.text = state.lanUrl
        lanModeText.text = getString(R.string.lan_only_mode)
        gameInfoBlock.visibility = if (state.running) View.GONE else View.VISIBLE
        connectionActionsRow.visibility = if (state.running) View.VISIBLE else View.GONE
        backendText.text = when {
            state.running && prefs.devMode() -> getString(R.string.server_mode_dev)
            state.running -> getString(R.string.server_mode_ready)
            else -> getString(R.string.server_mode_stopped)
        }

        startButton.text = if (state.running) getString(R.string.stop_button) else getString(R.string.start_button)
        startHintText.text = if (state.running) getString(R.string.start_hint_running) else getString(R.string.start_hint_idle)
        val startButtonBackgroundColor = if (state.running) R.color.status_running else android.R.color.white
        val startButtonTextColor = if (state.running) android.R.color.white else R.color.brand_primary
        startButton.backgroundTintList = ColorStateList.valueOf(
            ContextCompat.getColor(this, startButtonBackgroundColor)
        )
        startButton.setTextColor(ContextCompat.getColor(this, startButtonTextColor))

        openBrowserButton.isEnabled = state.running
        openInAppButton.isEnabled = state.running

        if (!state.running) {
            val port = prefs.port().takeIf { it in 1..65535 } ?: AppPreferences.DEFAULT_PORT
            val lanPreview = "http://${NetworkUtils.findLanIpv4() ?: "127.0.0.1"}:$port"
            lanUrlText.text = lanPreview
        }
    }

    private fun renderLogs(logLines: List<String>) {
        latestLogs = logLines.takeLast(250)
    }

    private fun resolveOpenUrl(): String {
        val state = ServerRuntime.state.value
        if (state.running) return state.lanUrl
        val port = prefs.port().takeIf { it in 1..65535 } ?: AppPreferences.DEFAULT_PORT
        return "http://${NetworkUtils.findLanIpv4() ?: "127.0.0.1"}:$port"
    }

    private fun showContextMenu(anchor: View) {
        val popup = PopupMenu(this, anchor)
        popup.menuInflater.inflate(R.menu.main_context_menu, popup.menu)
        popup.setOnMenuItemClickListener { item ->
            when (item.itemId) {
                R.id.action_settings -> {
                    showSettingsDialog()
                    true
                }
                R.id.action_logs -> {
                    showLogsDialog()
                    true
                }
                R.id.action_open_browser -> {
                    openInBrowser(resolveOpenUrl())
                    true
                }
                R.id.action_open_in_app -> {
                    openInApp(resolveOpenUrl())
                    true
                }
                R.id.action_about -> {
                    showAboutDialog()
                    true
                }
                else -> false
            }
        }
        popup.show()
    }

    private fun showSettingsDialog() {
        val sheet = BottomSheetDialog(this)
        val content = layoutInflater.inflate(R.layout.sheet_settings, null)
        sheet.setContentView(content)
        sheet.setOnShowListener {
            val bottomSheet =
                sheet.findViewById<FrameLayout>(com.google.android.material.R.id.design_bottom_sheet)
                    ?: return@setOnShowListener
            val behavior = BottomSheetBehavior.from(bottomSheet)
            behavior.isFitToContents = true
            behavior.skipCollapsed = true
            behavior.state = BottomSheetBehavior.STATE_EXPANDED
        }

        val portInput = content.findViewById<TextInputEditText>(R.id.settingsPortInput)
        val themeSwitch = content.findViewById<SwitchMaterial>(R.id.settingsThemeSwitch)
        val devModeSwitch = content.findViewById<SwitchMaterial>(R.id.settingsDevModeSwitch)
        val chip8080 = content.findViewById<Chip>(R.id.chipPort8080)
        val chip9090 = content.findViewById<Chip>(R.id.chipPort9090)
        val chip18080 = content.findViewById<Chip>(R.id.chipPort18080)
        val cancelButton = content.findViewById<Button>(R.id.settingsCancelButton)
        val saveButton = content.findViewById<Button>(R.id.settingsSaveButton)

        val previousPort = prefs.port()
        val previousDevMode = prefs.devMode()
        val previousTheme = prefs.themeMode()

        portInput.setText(previousPort.toString())
        themeSwitch.isChecked = previousTheme == AppPreferences.THEME_DARK
        devModeSwitch.isChecked = previousDevMode
        chip8080.setOnClickListener { portInput.setText("8080") }
        chip9090.setOnClickListener { portInput.setText("9090") }
        chip18080.setOnClickListener { portInput.setText("18080") }

        cancelButton.setOnClickListener {
            sheet.dismiss()
        }
        saveButton.setOnClickListener {
            val parsedPort = parsePort(portInput.text?.toString().orEmpty())
            if (parsedPort == null) {
                portInput.error = getString(R.string.invalid_port)
                return@setOnClickListener
            }

            val nextTheme = if (themeSwitch.isChecked) AppPreferences.THEME_DARK else AppPreferences.THEME_LIGHT
            val nextDevMode = devModeSwitch.isChecked

            prefs.save(parsedPort, nextDevMode)
            if (previousTheme != nextTheme) {
                prefs.setThemeMode(nextTheme)
                applySavedTheme()
            }

            if (ServerRuntime.state.value.running &&
                (parsedPort != previousPort || nextDevMode != previousDevMode)
            ) {
                Toast.makeText(this, getString(R.string.settings_apply_after_restart), Toast.LENGTH_SHORT).show()
            }
            renderState(ServerRuntime.state.value)
            Toast.makeText(this, getString(R.string.settings_saved), Toast.LENGTH_SHORT).show()
            sheet.dismiss()
        }
        sheet.show()
    }

    private fun showLogsDialog() {
        val body = if (latestLogs.isEmpty()) getString(R.string.logs_empty)
        else latestLogs.joinToString(separator = "\n")
        val content = layoutInflater.inflate(R.layout.dialog_logs, null)
        val logsText = content.findViewById<TextView>(R.id.logsDialogText)
        logsText.text = body
        MaterialAlertDialogBuilder(this)
            .setTitle(R.string.logs_title)
            .setView(content)
            .setNeutralButton(R.string.logs_copy) { _, _ ->
                copyTextToClipboard(body)
                Toast.makeText(this, getString(R.string.logs_copied), Toast.LENGTH_SHORT).show()
            }
            .setPositiveButton(android.R.string.ok, null)
            .show()
    }

    private fun copyTextToClipboard(text: String) {
        val clipboard = getSystemService(ClipboardManager::class.java) ?: return
        clipboard.setPrimaryClip(ClipData.newPlainText("protocol_bunker_logs", text))
    }

    private fun showAboutDialog() {
        MaterialAlertDialogBuilder(this)
            .setTitle(R.string.menu_about)
            .setMessage(getString(R.string.about_text))
            .setPositiveButton(android.R.string.ok, null)
            .show()
    }

    private fun openInBrowser(url: String) {
        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
        runCatching { startActivity(intent) }.onFailure {
            Toast.makeText(this, getString(R.string.open_url_error), Toast.LENGTH_SHORT).show()
        }
    }

    private fun openInApp(url: String) {
        try {
            val intent = Intent(this, WebViewActivity::class.java).apply {
                putExtra(WebViewActivity.EXTRA_URL, url)
            }
            startActivity(intent)
        } catch (_: ActivityNotFoundException) {
            Toast.makeText(this, getString(R.string.open_url_error), Toast.LENGTH_SHORT).show()
        }
    }

    private fun parsePort(value: String): Int? {
        val normalized = value.trim()
        if (normalized.isEmpty()) return null
        val port = normalized.toIntOrNull() ?: return null
        return if (port in 1..65535) port else null
    }

    private fun requestNotificationsPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) return
        if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) {
            return
        }
        notificationsPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
    }

    private fun applySavedTheme() {
        val nightMode = when (prefs.themeMode()) {
            AppPreferences.THEME_DARK -> AppCompatDelegate.MODE_NIGHT_YES
            else -> AppCompatDelegate.MODE_NIGHT_NO
        }
        if (AppCompatDelegate.getDefaultNightMode() != nightMode) {
            AppCompatDelegate.setDefaultNightMode(nightMode)
        }
    }
}
