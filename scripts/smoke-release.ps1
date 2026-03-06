$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $true

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Invoke-Step {
  param([scriptblock]$Command)
  & $Command
  if ($LASTEXITCODE -ne 0) {
    throw "Command failed with exit code $LASTEXITCODE"
  }
}

Write-Host "[smoke] Validate encoding"
Invoke-Step { ./scripts/check-encoding.ps1 }

Write-Host "[smoke] Install JS deps (shared/client/scenarios)"
Invoke-Step { npm --prefix shared ci --no-audit --no-fund }
Invoke-Step { npm --prefix client ci --no-audit --no-fund }
Invoke-Step { npm --prefix scenarios install --no-audit --no-fund --package-lock=false }

Write-Host "[smoke] Verify generated special-effect contract"
Invoke-Step { node ./scripts/generate-special-effect-contract.mjs --check }

Write-Host "[smoke] Go tests"
Push-Location server-go
Invoke-Step { go test ./... }
Pop-Location

Write-Host "[smoke] Shared build + tests"
Invoke-Step { npm --prefix shared run build }
Invoke-Step { npm --prefix shared run test }

Write-Host "[smoke] Scenarios tests + build"
Invoke-Step { npm --prefix scenarios run test }
Invoke-Step { npm --prefix scenarios run build }

Write-Host "[smoke] Client typecheck + build"
Invoke-Step { npm --prefix client run typecheck }
Invoke-Step { npm --prefix client run build }

Write-Host "[smoke] OK"
