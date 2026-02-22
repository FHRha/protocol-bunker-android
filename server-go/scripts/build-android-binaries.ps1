$ErrorActionPreference = "Stop"

$RootDir = Resolve-Path (Join-Path $PSScriptRoot "..\\..")
$ServerDir = Join-Path $RootDir "server-go"
$AssetsDir = Join-Path $RootDir "android-app\\app\\src\\main\\assets\\server-binaries"
$AndroidAppDir = Join-Path $RootDir "android-app"

function Resolve-SdkRoot {
    if ($env:ANDROID_SDK_ROOT) { return $env:ANDROID_SDK_ROOT }
    if ($env:ANDROID_HOME) { return $env:ANDROID_HOME }

    $localProps = Join-Path $AndroidAppDir "local.properties"
    if (Test-Path $localProps) {
        foreach ($line in Get-Content $localProps) {
            if ($line -match "^sdk\.dir=(.+)$") {
                $raw = $Matches[1].Trim()
                if (-not $raw) { continue }
                $normalized = $raw -replace "\\\\", "\" -replace "\\:", ":"
                return $normalized
            }
        }
    }

    $repoSdk = Join-Path $AndroidAppDir ".android-sdk"
    if (Test-Path $repoSdk) { return $repoSdk }
    throw "Android SDK path not found. Configure ANDROID_SDK_ROOT/ANDROID_HOME or android-app/local.properties."
}

function Resolve-NdkLlvmBin {
    param([string]$SdkRoot)
    $ndkRoot = Join-Path $SdkRoot "ndk"
    if (-not (Test-Path $ndkRoot)) {
        throw "Android NDK not found in $ndkRoot. Install package 'ndk;27.2.12479018' via sdkmanager."
    }
    $ndkVersionDir = Get-ChildItem -Path $ndkRoot -Directory | Sort-Object Name -Descending | Select-Object -First 1
    if (-not $ndkVersionDir) {
        throw "No NDK versions found in $ndkRoot."
    }
    $llvmBin = Join-Path $ndkVersionDir.FullName "toolchains\\llvm\\prebuilt\\windows-x86_64\\bin"
    if (-not (Test-Path $llvmBin)) {
        throw "NDK LLVM toolchain not found: $llvmBin"
    }
    return $llvmBin
}

function Build-One {
    param(
        [string]$GoArch,
        [string]$Abi,
        [string]$ClangTarget,
        [string]$LlvmBinDir,
        [string]$GoArm = ""
    )

    $OutputPath = Join-Path $AssetsDir "$Abi\\server-go"
    New-Item -ItemType Directory -Force (Split-Path $OutputPath -Parent) | Out-Null
    Write-Host "Building $Abi..."

    $cc = Join-Path $LlvmBinDir "$ClangTarget.cmd"
    if (-not (Test-Path $cc)) {
        throw "Clang wrapper not found for ${Abi}: $cc"
    }

    $env:GOOS = "android"
    $env:GOARCH = $GoArch
    $env:CGO_ENABLED = "1"
    $env:CC = $cc
    if ($GoArm -ne "") {
        $env:GOARM = $GoArm
    } else {
        Remove-Item Env:GOARM -ErrorAction SilentlyContinue
    }

    Push-Location $ServerDir
    try {
        go build -trimpath -ldflags "-s -w" -o $OutputPath .
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed for ABI '${Abi}'"
        }
    } finally {
        Pop-Location
    }
}

$sdkRoot = Resolve-SdkRoot
$llvmBinDir = Resolve-NdkLlvmBin -SdkRoot $sdkRoot

Build-One -GoArch "arm64" -Abi "arm64-v8a" -ClangTarget "aarch64-linux-android26-clang" -LlvmBinDir $llvmBinDir
Build-One -GoArch "arm" -Abi "armeabi-v7a" -GoArm "7" -ClangTarget "armv7a-linux-androideabi26-clang" -LlvmBinDir $llvmBinDir
Build-One -GoArch "amd64" -Abi "x86_64" -ClangTarget "x86_64-linux-android26-clang" -LlvmBinDir $llvmBinDir
Build-One -GoArch "386" -Abi "x86" -ClangTarget "i686-linux-android26-clang" -LlvmBinDir $llvmBinDir

Write-Host "Done. Binaries copied to $AssetsDir"
