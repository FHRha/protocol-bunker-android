param(
    [string]$Package = "com.protocolbunker.host",
    [int]$Port = 8080,
    [switch]$SetPort,
    [switch]$ResetAppData = $true,
    [switch]$UseDevMode = $true
)

$ErrorActionPreference = "Stop"

$MainActivity = "$Package/.MainActivity"

function Resolve-SdkRoot {
    if ($env:ANDROID_SDK_ROOT) { return $env:ANDROID_SDK_ROOT }
    if ($env:ANDROID_HOME) { return $env:ANDROID_HOME }

    $root = Resolve-Path (Join-Path $PSScriptRoot "..")
    $localProps = Join-Path $root "local.properties"
    if (Test-Path $localProps) {
        foreach ($line in Get-Content $localProps) {
            if ($line -match "^sdk\.dir=(.+)$") {
                return (($Matches[1].Trim()) -replace "\\\\", "\" -replace "\\:", ":")
            }
        }
    }
    $repoSdk = Join-Path $root ".android-sdk"
    if (Test-Path $repoSdk) { return $repoSdk }
    return $null
}

function Resolve-AdbPath {
    $adbCmd = Get-Command adb -ErrorAction SilentlyContinue
    if ($adbCmd) { return $adbCmd.Source }

    $sdkRoot = Resolve-SdkRoot
    if ($sdkRoot) {
        $candidate = Join-Path $sdkRoot "platform-tools\\adb.exe"
        if (Test-Path $candidate) { return $candidate }
    }
    throw "adb not found. Install platform-tools and configure ANDROID_SDK_ROOT/ANDROID_HOME (or local.properties)."
}

function Get-HealthResponse {
    param(
        [string]$AdbPath,
        [int]$HealthPort
    )
    & $AdbPath shell "toybox wget -qO- http://127.0.0.1:$HealthPort/health 2>/dev/null || curl -s http://127.0.0.1:$HealthPort/health 2>/dev/null"
}

function Parse-BoundsCenter {
    param([string]$Bounds)
    if ($Bounds -notmatch "^\[(\d+),(\d+)\]\[(\d+),(\d+)\]$") {
        throw "Unexpected bounds format: $Bounds"
    }
    $x1 = [int]$Matches[1]
    $y1 = [int]$Matches[2]
    $x2 = [int]$Matches[3]
    $y2 = [int]$Matches[4]
    $x = [int](($x1 + $x2) / 2)
    $y = [int](($y1 + $y2) / 2)
    return @($x, $y)
}

function Find-NodeByResourceId {
    param(
        [string]$XmlText,
        [string]$ResourceId,
        [switch]$MustBeEnabled
    )
    if ([string]::IsNullOrWhiteSpace($XmlText)) {
        return $null
    }
    [xml]$xml = $XmlText
    $nodes = $xml.SelectNodes("//node[@resource-id='$ResourceId']")
    if ($null -eq $nodes) {
        return $null
    }
    foreach ($candidate in $nodes) {
        if (-not $MustBeEnabled -or $candidate.GetAttribute("enabled") -eq "true") {
            return $candidate
        }
    }
    return $null
}

function Wait-NodeByResourceId {
    param(
        [string]$AdbPath,
        [string]$ResourceId,
        [int]$TimeoutSeconds = 15,
        [switch]$MustBeEnabled
    )
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        & $AdbPath shell uiautomator dump /sdcard/pb_ui.xml | Out-Null
        $xmlText = & $AdbPath shell cat /sdcard/pb_ui.xml
        $node = Find-NodeByResourceId -XmlText $xmlText -ResourceId $ResourceId -MustBeEnabled:$MustBeEnabled
        if ($null -ne $node) {
            return $node
        }
        Start-Sleep -Milliseconds 400
    }
    throw "Resource not found in UI dump: $ResourceId"
}

function Tap-ResourceId {
    param(
        [string]$AdbPath,
        [string]$ResourceId,
        [int]$TimeoutSeconds = 15
    )
    $node = Wait-NodeByResourceId -AdbPath $AdbPath -ResourceId $ResourceId -TimeoutSeconds $TimeoutSeconds -MustBeEnabled
    $bounds = $node.GetAttribute("bounds")
    $point = Parse-BoundsCenter -Bounds $bounds
    & $AdbPath shell input tap $point[0] $point[1] | Out-Null
}

function Set-SwitchState {
    param(
        [string]$AdbPath,
        [string]$PackageName,
        [string]$SwitchId,
        [bool]$Enabled,
        [switch]$Optional
    )
    $resourceId = "$PackageName`:id/$SwitchId"
    try {
        $node = Wait-NodeByResourceId -AdbPath $AdbPath -ResourceId $resourceId -TimeoutSeconds 15
    } catch {
        if ($Optional) {
            return
        }
        throw
    }
    $current = ($node.GetAttribute("checked") -eq "true")
    if ($current -ne $Enabled) {
        $bounds = $node.GetAttribute("bounds")
        $point = Parse-BoundsCenter -Bounds $bounds
        & $AdbPath shell input tap $point[0] $point[1] | Out-Null
    }
}

function Set-PortInput {
    param(
        [string]$AdbPath,
        [string]$PackageName,
        [int]$TargetPort
    )
    Tap-ResourceId -AdbPath $AdbPath -ResourceId "$PackageName`:id/portInput"
    Start-Sleep -Milliseconds 150
    & $AdbPath shell input keyevent 123 | Out-Null # KEYCODE_MOVE_END
    for ($i = 0; $i -lt 6; $i++) {
        & $AdbPath shell input keyevent 67 | Out-Null # KEYCODE_DEL
    }
    & $AdbPath shell input text "$TargetPort" | Out-Null
}

$adb = Resolve-AdbPath
$devices = & $adb devices
if (-not ($devices -match "device$")) {
    throw "No connected Android device/emulator found (adb devices)."
}

Write-Host "1) Launching app UI..."
if ($ResetAppData) {
    & $adb shell pm clear $Package | Out-Null
    Start-Sleep -Milliseconds 400
}
& $adb shell am start -n $MainActivity | Out-Null
Start-Sleep -Seconds 2

Write-Host "2) Tapping Start..."
Set-SwitchState -AdbPath $adb -PackageName $Package -SwitchId "devModeSwitch" -Enabled ([bool]$UseDevMode) -Optional
if ($SetPort) {
    Set-PortInput -AdbPath $adb -PackageName $Package -TargetPort $Port
}
Tap-ResourceId -AdbPath $adb -ResourceId "$Package`:id/startButton"
Start-Sleep -Seconds 2

Write-Host "3) Checking health endpoint from device loopback..."
$ok = $false
for ($i = 0; $i -lt 20; $i++) {
    $health = Get-HealthResponse -AdbPath $adb -HealthPort $Port
    if ($health -match '"status":"ok"') {
        $ok = $true
        break
    }
    Start-Sleep -Milliseconds 500
}
if (-not $ok) {
    throw "Health check failed. Response: $health"
}

Write-Host "4) Sending app to background..."
& $adb shell input keyevent 3 | Out-Null
Start-Sleep -Seconds 1

Write-Host "5) Verifying service notification exists..."
$notif = & $adb shell dumpsys notification --noredact
if (-not ($notif -match "Protocol: Bunker")) {
    throw "Foreground notification not found"
}

Write-Host "6) Reopening app and tapping Stop..."
& $adb shell am start -n $MainActivity | Out-Null
Start-Sleep -Seconds 1
try {
    Tap-ResourceId -AdbPath $adb -ResourceId "$Package`:id/stopButton"
} catch {
    Tap-ResourceId -AdbPath $adb -ResourceId "$Package`:id/startButton"
}
Start-Sleep -Seconds 1

Write-Host "7) Verifying endpoint is down..."
$stopped = $false
for ($i = 0; $i -lt 12; $i++) {
    $healthAfterStop = Get-HealthResponse -AdbPath $adb -HealthPort $Port
    if (-not ($healthAfterStop -match '"status":"ok"')) {
        $stopped = $true
        break
    }
    Start-Sleep -Milliseconds 500
}
if (-not $stopped) {
    throw "Health endpoint is still reachable after stop"
}

Write-Host "Smoke e2e passed"
