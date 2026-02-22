package com.protocolbunker.host.server

import android.content.Context
import android.os.Build
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File

internal object ServerBinaryInstaller {
    private const val BIN_ASSET_ROOT = "server-binaries"
    private const val BIN_RELATIVE_TARGET = "server-go/server-go"
    private const val BIN_META_FILE = "server-go/.installed_abi"

    private const val RUNTIME_ASSET_ROOT = "server-runtime"
    private const val RUNTIME_RELATIVE_ROOT = "server-runtime"
    private const val RUNTIME_META_FILE = "server-runtime/.installed_stamp"

    suspend fun ensureInstalled(context: Context, logger: (String) -> Unit): Boolean =
        withContext(Dispatchers.IO) {
            val supportedAbis = Build.SUPPORTED_ABIS?.toList().orEmpty()
            if (supportedAbis.isEmpty()) {
                logger("ABI не определён: список SUPPORTED_ABIS пуст")
                return@withContext false
            }

            val selectedAbi = supportedAbis.firstOrNull { hasBundledBinary(context, it) }
            if (selectedAbi == null) {
                logger("Не найден встроенный Go-бинарь под ABI: ${supportedAbis.joinToString()}")
                return@withContext false
            }

            val binaryReady = installBinaryIfNeeded(context, selectedAbi, logger)
            if (!binaryReady) return@withContext false

            val runtimeReady = installRuntimeIfNeeded(context, logger)
            if (!runtimeReady) return@withContext false

            true
        }

    fun runtimeAssetsRoot(context: Context): File =
        File(context.filesDir, "$RUNTIME_RELATIVE_ROOT/assets")

    fun runtimeClientDistRoot(context: Context): File =
        File(context.filesDir, "$RUNTIME_RELATIVE_ROOT/client/dist")

    fun runtimeScenariosRoot(context: Context): File =
        File(context.filesDir, "$RUNTIME_RELATIVE_ROOT/scenarios")

    private fun installBinaryIfNeeded(context: Context, abi: String, logger: (String) -> Unit): Boolean {
        val target = File(context.filesDir, BIN_RELATIVE_TARGET)
        val meta = File(context.filesDir, BIN_META_FILE)
        val installStamp = appInstallStamp(context)
        val expectedMeta = "$abi|$installStamp"
        val alreadyInstalled = target.exists() && meta.exists() && meta.readText().trim() == expectedMeta
        if (alreadyInstalled) {
            if (!target.canExecute()) {
                target.setExecutable(true, false)
            }
            return true
        }

        target.parentFile?.mkdirs()
        val temp = File(target.parentFile, "${target.name}.tmp")
        context.assets.open("$BIN_ASSET_ROOT/$abi/server-go").use { input ->
            temp.outputStream().use { output -> input.copyTo(output) }
        }
        if (target.exists()) {
            target.delete()
        }
        check(temp.renameTo(target)) { "Не удалось установить бинарь в ${target.absolutePath}" }
        target.setExecutable(true, false)

        meta.parentFile?.mkdirs()
        meta.writeText(expectedMeta)
        logger("Go-бинарь установлен для ABI: $abi")
        return true
    }

    private fun installRuntimeIfNeeded(context: Context, logger: (String) -> Unit): Boolean {
        val hasDecksInApk = hasAssetDir(context, "$RUNTIME_ASSET_ROOT/assets/decks")
        val hasClientInApk = hasAssetFile(context, "$RUNTIME_ASSET_ROOT/client/dist/index.html")
        val hasSpecialsInApk = hasAssetFile(context, "$RUNTIME_ASSET_ROOT/scenarios/classic/SPECIAL_CONDITIONS.json")
        if (!hasDecksInApk || !hasClientInApk || !hasSpecialsInApk) {
            logger("В APK не найдены игровые ресурсы (decks/client-dist/scenarios)")
            return false
        }

        val runtimeRoot = File(context.filesDir, RUNTIME_RELATIVE_ROOT)
        val runtimeMeta = File(context.filesDir, RUNTIME_META_FILE)
        val installStamp = appInstallStamp(context)
        val targetDecks = File(runtimeRoot, "assets/decks")
        val targetClient = File(runtimeRoot, "client/dist/index.html")
        val targetSpecials = File(runtimeRoot, "scenarios/classic/SPECIAL_CONDITIONS.json")
        val runtimeReady = targetDecks.isDirectory && targetClient.isFile && targetSpecials.isFile
        val runtimeUpToDate = runtimeMeta.exists() && runtimeMeta.readText().trim() == installStamp
        if (runtimeReady && runtimeUpToDate) {
            return true
        }

        if (runtimeRoot.exists()) {
            runtimeRoot.deleteRecursively()
        }
        runtimeRoot.mkdirs()

        copyAssetTree(context, RUNTIME_ASSET_ROOT, runtimeRoot)

        val installedDecks = File(runtimeRoot, "assets/decks")
        val installedClient = File(runtimeRoot, "client/dist/index.html")
        val installedSpecials = File(runtimeRoot, "scenarios/classic/SPECIAL_CONDITIONS.json")
        if (!installedDecks.isDirectory || !installedClient.isFile || !installedSpecials.isFile) {
            logger("Не удалось установить игровые ресурсы в ${runtimeRoot.absolutePath}")
            return false
        }

        runtimeMeta.parentFile?.mkdirs()
        runtimeMeta.writeText(installStamp)
        logger("Игровые ресурсы установлены: ${runtimeRoot.absolutePath}")
        return true
    }

    private fun copyAssetTree(context: Context, assetPath: String, destination: File) {
        val children = context.assets.list(assetPath).orEmpty()
        if (children.isEmpty()) {
            destination.parentFile?.mkdirs()
            context.assets.open(assetPath).use { input ->
                destination.outputStream().use { output -> input.copyTo(output) }
            }
            return
        }

        if (!destination.exists()) {
            destination.mkdirs()
        }
        for (child in children) {
            copyAssetTree(context, "$assetPath/$child", File(destination, child))
        }
    }

    private fun hasBundledBinary(context: Context, abi: String): Boolean {
        val files = runCatching { context.assets.list("$BIN_ASSET_ROOT/$abi") }.getOrNull() ?: return false
        return files.contains("server-go")
    }

    private fun hasAssetDir(context: Context, path: String): Boolean {
        val list = runCatching { context.assets.list(path) }.getOrNull() ?: return false
        return list.isNotEmpty()
    }

    private fun hasAssetFile(context: Context, path: String): Boolean =
        runCatching {
            context.assets.open(path).use { }
            true
        }.getOrDefault(false)

    @Suppress("DEPRECATION")
    private fun appInstallStamp(context: Context): String {
        val fallback = "vc=unknown|lu=unknown"
        return runCatching {
            @Suppress("DEPRECATION")
            val packageInfo =
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                    context.packageManager.getPackageInfo(
                        context.packageName,
                        android.content.pm.PackageManager.PackageInfoFlags.of(0)
                    )
                } else {
                    context.packageManager.getPackageInfo(context.packageName, 0)
                }
            val versionCode =
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) packageInfo.longVersionCode
                else packageInfo.versionCode.toLong()
            "vc=$versionCode|lu=${packageInfo.lastUpdateTime}"
        }.getOrDefault(fallback)
    }
}
