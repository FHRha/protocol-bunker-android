import java.io.FileInputStream
import java.util.Properties
import org.gradle.api.tasks.Sync

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

val releaseProps = Properties().apply {
    val file = rootProject.file("keystore.properties")
    if (file.exists()) {
        FileInputStream(file).use(::load)
    }
}

fun readConfig(key: String, envKey: String): String? {
    val fromFile = releaseProps.getProperty(key)?.trim().orEmpty()
    if (fromFile.isNotEmpty()) return fromFile
    val fromGradle = (project.findProperty(key) as? String)?.trim().orEmpty()
    if (fromGradle.isNotEmpty()) return fromGradle
    val fromEnv = System.getenv(envKey)?.trim().orEmpty()
    return fromEnv.ifEmpty { null }
}

fun readIntConfig(key: String, envKey: String, fallback: Int): Int =
    readConfig(key, envKey)?.toIntOrNull() ?: fallback

val releaseAppId = readConfig("APP_ID", "PB_APP_ID") ?: "com.protocolbunker.host"
val releaseVersionCode = readIntConfig("VERSION_CODE", "PB_VERSION_CODE", 10000)
val releaseVersionName = readConfig("VERSION_NAME", "PB_VERSION_NAME") ?: "1.0.0"

val releaseStoreFilePath = readConfig("RELEASE_STORE_FILE", "PB_RELEASE_STORE_FILE")
    ?: readConfig("storeFile", "PB_RELEASE_STORE_FILE")
val releaseStorePassword = readConfig("RELEASE_STORE_PASSWORD", "PB_RELEASE_STORE_PASSWORD")
    ?: readConfig("storePassword", "PB_RELEASE_STORE_PASSWORD")
val releaseKeyAlias = readConfig("RELEASE_KEY_ALIAS", "PB_RELEASE_KEY_ALIAS")
    ?: readConfig("keyAlias", "PB_RELEASE_KEY_ALIAS")
val releaseKeyPassword = readConfig("RELEASE_KEY_PASSWORD", "PB_RELEASE_KEY_PASSWORD")
    ?: readConfig("keyPassword", "PB_RELEASE_KEY_PASSWORD")
val enableReleaseLint = readConfig("ENABLE_RELEASE_LINT", "PB_ENABLE_RELEASE_LINT")?.toBooleanStrictOrNull() ?: false
val workspaceRoot = rootProject.projectDir.resolve("..")
val bundledBinaryRoot = project.layout.projectDirectory.dir("src/main/assets/server-binaries")
val runtimeAssetsCommonDir = layout.buildDirectory.dir("generated/runtime-assets/common")
val runtimeAssetsArm64Dir = layout.buildDirectory.dir("generated/runtime-assets/arm64")
val runtimeAssetsArmv7Dir = layout.buildDirectory.dir("generated/runtime-assets/armv7")
val runtimeAssetsX86Dir = layout.buildDirectory.dir("generated/runtime-assets/x86")
val runtimeAssetsX8664Dir = layout.buildDirectory.dir("generated/runtime-assets/x8664")
val runtimeAssetsUniversalDir = layout.buildDirectory.dir("generated/runtime-assets/universal")
val supportedAbis = listOf("arm64-v8a", "armeabi-v7a", "x86", "x86_64")

val syncRuntimeAssetsCommon by tasks.registering(Sync::class) {
    val decksSource = workspaceRoot.resolve("assets/decks")
    val clientDistSource = workspaceRoot.resolve("client/dist")
    val specialsSource = workspaceRoot.resolve("scenarios/classic/SPECIAL_CONDITIONS.json")

    doFirst {
        check(decksSource.exists()) { "Missing assets source directory: ${decksSource.absolutePath}" }
        check(clientDistSource.resolve("index.html").exists()) {
            "Missing client dist index: ${clientDistSource.resolve("index.html").absolutePath}"
        }
        check(specialsSource.exists()) { "Missing specials source file: ${specialsSource.absolutePath}" }
    }

    from(decksSource) { into("server-runtime/assets/decks") }
    from(clientDistSource) { into("server-runtime/client/dist") }
    from(specialsSource) { into("server-runtime/scenarios/classic") }
    into(runtimeAssetsCommonDir)
}

fun registerBinarySyncTask(
    taskName: String,
    outputDir: org.gradle.api.provider.Provider<org.gradle.api.file.Directory>,
    abiDirs: List<String>
) = tasks.register<Sync>(taskName) {
    doFirst {
        abiDirs.forEach { abi ->
            check(bundledBinaryRoot.file("$abi/server-go").asFile.exists()) {
                "Missing server binary for ABI $abi: ${bundledBinaryRoot.file("$abi/server-go").asFile.absolutePath}"
            }
        }
    }
    if (bundledBinaryRoot.file("README.md").asFile.exists()) {
        from(bundledBinaryRoot.file("README.md")) { into("server-binaries") }
    }
    abiDirs.forEach { abi ->
        from(bundledBinaryRoot.dir(abi)) { into("server-binaries/$abi") }
    }
    into(outputDir)
}

val syncRuntimeAssetsArm64 by registerBinarySyncTask(
    taskName = "syncRuntimeAssetsArm64",
    outputDir = runtimeAssetsArm64Dir,
    abiDirs = listOf("arm64-v8a")
)
val syncRuntimeAssetsArmv7 by registerBinarySyncTask(
    taskName = "syncRuntimeAssetsArmv7",
    outputDir = runtimeAssetsArmv7Dir,
    abiDirs = listOf("armeabi-v7a")
)
val syncRuntimeAssetsX86 by registerBinarySyncTask(
    taskName = "syncRuntimeAssetsX86",
    outputDir = runtimeAssetsX86Dir,
    abiDirs = listOf("x86")
)
val syncRuntimeAssetsX8664 by registerBinarySyncTask(
    taskName = "syncRuntimeAssetsX8664",
    outputDir = runtimeAssetsX8664Dir,
    abiDirs = listOf("x86_64")
)
val syncRuntimeAssetsUniversal by registerBinarySyncTask(
    taskName = "syncRuntimeAssetsUniversal",
    outputDir = runtimeAssetsUniversalDir,
    abiDirs = supportedAbis
)
val resolvedReleaseStoreFile = releaseStoreFilePath?.let { rawPath ->
    val moduleRelative = file(rawPath)
    if (moduleRelative.exists()) moduleRelative else rootProject.file(rawPath)
}
val hasReleaseSigning = !releaseStoreFilePath.isNullOrBlank() &&
    !releaseStorePassword.isNullOrBlank() &&
    !releaseKeyAlias.isNullOrBlank() &&
    !releaseKeyPassword.isNullOrBlank() &&
    (resolvedReleaseStoreFile?.exists() == true)

android {
    namespace = "com.protocolbunker.host"
    compileSdk = 35
    sourceSets.getByName("main").assets.setSrcDirs(emptyList<String>())
    sourceSets.getByName("main").assets.srcDir(runtimeAssetsCommonDir)

    defaultConfig {
        applicationId = releaseAppId
        minSdk = 26
        targetSdk = 35
        versionCode = releaseVersionCode
        versionName = releaseVersionName
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    flavorDimensions += "abi"
    productFlavors {
        create("arm64") { dimension = "abi" }
        create("armv7") { dimension = "abi" }
        create("x86") { dimension = "abi" }
        create("x8664") { dimension = "abi" }
        create("universal") { dimension = "abi" }
    }
    sourceSets.getByName("arm64").assets.srcDir(runtimeAssetsArm64Dir)
    sourceSets.getByName("armv7").assets.srcDir(runtimeAssetsArmv7Dir)
    sourceSets.getByName("x86").assets.srcDir(runtimeAssetsX86Dir)
    sourceSets.getByName("x8664").assets.srcDir(runtimeAssetsX8664Dir)
    sourceSets.getByName("universal").assets.srcDir(runtimeAssetsUniversalDir)

    signingConfigs {
        create("release") {
            if (hasReleaseSigning) {
                storeFile = requireNotNull(resolvedReleaseStoreFile)
                storePassword = releaseStorePassword
                keyAlias = releaseKeyAlias
                keyPassword = releaseKeyPassword
                enableV1Signing = true
                enableV2Signing = true
            }
        }
    }

    buildTypes {
        debug {
            isMinifyEnabled = false
        }
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            isDebuggable = false
            signingConfig =
                if (hasReleaseSigning) signingConfigs.getByName("release")
                else signingConfigs.getByName("debug")
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    lint {
        // In restricted networks (where dl.google.com is blocked), lint artifacts may be unreachable.
        // Keep release assembly available by default; enable lint explicitly in full-network CI.
        checkReleaseBuilds = enableReleaseLint
        abortOnError = enableReleaseLint
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }
}

tasks.named("preBuild").configure {
    dependsOn(syncRuntimeAssetsCommon)
    dependsOn(syncRuntimeAssetsArm64)
    dependsOn(syncRuntimeAssetsArmv7)
    dependsOn(syncRuntimeAssetsX86)
    dependsOn(syncRuntimeAssetsX8664)
    dependsOn(syncRuntimeAssetsUniversal)
}

if (!hasReleaseSigning) {
    logger.warn(
        "Release keystore is not configured. " +
            "Release build will use debug signing. " +
            "Configure android-app/keystore.properties for production signing."
    )
}

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.appcompat:appcompat:1.7.0")
    implementation("com.google.android.material:material:1.12.0")
    implementation("androidx.activity:activity-ktx:1.9.3")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.8.6")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1")

    testImplementation("junit:junit:4.13.2")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
    androidTestImplementation("androidx.test:runner:1.6.2")
    androidTestImplementation("androidx.test:core-ktx:1.6.1")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.6.1")
    androidTestImplementation("androidx.test.uiautomator:uiautomator:2.3.0")
}
