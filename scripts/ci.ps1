param(
    [string]$AppId = $(if ($env:PB_APP_ID) { $env:PB_APP_ID } else { "com.protocolbunker.host" }),
    [switch]$RunDeviceTests
)

$ErrorActionPreference = "Stop"

$RootDir = Resolve-Path (Join-Path $PSScriptRoot "..")

function Resolve-SdkRoot {
    param([string]$AndroidAppDir)
    if ($env:ANDROID_SDK_ROOT) { return $env:ANDROID_SDK_ROOT }
    if ($env:ANDROID_HOME) { return $env:ANDROID_HOME }

    $localProps = Join-Path $AndroidAppDir "local.properties"
    if (Test-Path $localProps) {
        foreach ($line in Get-Content $localProps) {
            if ($line -match "^sdk\.dir=(.+)$") {
                $raw = $Matches[1].Trim()
                if (-not $raw) { continue }
                return ($raw -replace "\\\\", "\" -replace "\\:", ":")
            }
        }
    }

    $repoSdk = Join-Path $AndroidAppDir ".android-sdk"
    if (Test-Path $repoSdk) { return $repoSdk }
    return $null
}

function Resolve-Aapt2OverrideArg {
    param([string]$AndroidAppDir)
    $sdkRoot = Resolve-SdkRoot -AndroidAppDir $AndroidAppDir
    if (-not $sdkRoot) { return $null }

    $candidates = @(
        (Join-Path $sdkRoot "build-tools\\35.0.0\\aapt2.exe"),
        (Join-Path $sdkRoot "build-tools\\34.0.0\\aapt2.exe")
    )
    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            $resolved = (Get-Item $candidate).FullName -replace '\\', '/'
            return "-Pandroid.aapt2FromMavenOverride=$resolved"
        }
    }
    return $null
}

Write-Host "==> go test"
Push-Location (Join-Path $RootDir "server-go")
go test ./...
.\scripts\build-android-binaries.ps1
Pop-Location

Write-Host "==> android assembleDebug/assembleRelease"
Push-Location (Join-Path $RootDir "android-app")
$aapt2Arg = Resolve-Aapt2OverrideArg -AndroidAppDir (Get-Location).Path
$gradleBuildArgs = @("assembleDebug", "assembleRelease")
if ($aapt2Arg) {
    Write-Host "Using local aapt2 override: $aapt2Arg"
    $gradleBuildArgs += $aapt2Arg
}
.\gradlew.bat @gradleBuildArgs

if ($RunDeviceTests.IsPresent) {
    Write-Host "==> connectedDebugAndroidTest"
    $gradleTestArgs = @("connectedDebugAndroidTest")
    if ($aapt2Arg) {
        $gradleTestArgs += $aapt2Arg
    }
    .\gradlew.bat @gradleTestArgs
    Write-Host "==> smoke-e2e"
    .\scripts\smoke-e2e.ps1 -Package $AppId
} else {
    Write-Host "==> device tests skipped (use -RunDeviceTests to enable)"
}

Pop-Location

Write-Host "CI pipeline commands finished"
