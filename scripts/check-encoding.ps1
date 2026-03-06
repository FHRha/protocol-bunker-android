$ErrorActionPreference = "Stop"

$utf8Strict = New-Object System.Text.UTF8Encoding($false, $true)
$enc1251 = [System.Text.Encoding]::GetEncoding(1251)
$utf8 = [System.Text.Encoding]::UTF8

$skipExtensions = @(
  ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".ico", ".icns",
  ".apk", ".aab", ".jks", ".keystore", ".p12", ".jar", ".zip", ".7z",
  ".pdf", ".mp3", ".mp4", ".avi", ".mov", ".wav", ".ogg", ".ttf", ".otf",
  ".woff", ".woff2", ".so", ".dll", ".exe", ".bin", ".class", ".wasm"
)

function Normalize-Mojibake {
  param([string]$Value)

  $current = $Value
  for ($i = 0; $i -lt 3; $i++) {
    try {
      $decoded = $utf8.GetString($enc1251.GetBytes($current))
      $roundtrip = $enc1251.GetString($utf8.GetBytes($decoded))
    } catch {
      break
    }

    if ($roundtrip -ne $current) { break }
    if ($decoded.Contains([char]0xFFFD)) { break }
    if (-not [regex]::IsMatch($decoded, "[\u0400-\u04FF]")) { break }
    if ($decoded -eq $current) { break }

    $current = $decoded
  }

  return $current
}

function Find-SuspiciousChunks {
  param([string]$Text)

  $suspects = New-Object System.Collections.Generic.List[string]

  # Quoted strings first (most user-facing text lives there)
  $quoted = [regex]::Matches($Text, '"([^"\\]*(?:\\.[^"\\]*)*)"')
  foreach ($m in $quoted) {
    $raw = $m.Groups[1].Value
    $fixed = Normalize-Mojibake -Value $raw
    if ($fixed -ne $raw) {
      $suspects.Add($raw)
    }
  }

  # Then generic tokens for comments/plain text files.
  $tokens = [regex]::Matches($Text, "[\p{L}\p{M}\p{N}_\-/\.]{4,}")
  foreach ($m in $tokens) {
    $raw = $m.Value
    if ($raw -notmatch "[\u0420\u0421]") { continue }
    $fixed = Normalize-Mojibake -Value $raw
    if ($fixed -ne $raw) {
      $suspects.Add($raw)
    }
  }

  return $suspects
}

$issues = New-Object System.Collections.Generic.List[string]
$files = git -c core.quotepath=false ls-files

foreach ($file in $files) {
  if (-not (Test-Path $file)) { continue }
  if ($file -match "^(?:.+/)?(?:node_modules|dist|build|out)/") { continue }
  if ($file -match "^android-app/app/src/main/assets/server-binaries/.+/server-go$") { continue }

  $ext = [System.IO.Path]::GetExtension($file).ToLowerInvariant()
  if ($skipExtensions -contains $ext) { continue }

  $bytes = [System.IO.File]::ReadAllBytes($file)
  if ($bytes.Length -eq 0) { continue }
  if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
    $issues.Add("${file}: UTF-8 BOM detected (use UTF-8 without BOM)")
    continue
  }

  $text = $null
  try {
    $text = $utf8Strict.GetString($bytes)
  } catch {
    $issues.Add("${file}: invalid UTF-8")
    continue
  }

  $suspects = Find-SuspiciousChunks -Text $text
  if ($suspects.Count -gt 0) {
    $preview = ($suspects | Select-Object -First 2) -join " | "
    $issues.Add("${file}: mojibake-like text detected: $preview")
  }
}

if ($issues.Count -gt 0) {
  Write-Host "Encoding check failed:"
  $issues | ForEach-Object { Write-Host " - $_" }
  exit 1
}

Write-Host "Encoding check passed: UTF-8 without BOM, no invalid UTF-8 / mojibake-like chunks detected."
