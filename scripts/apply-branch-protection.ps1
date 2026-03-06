param(
    [string]$Branch = "main",
    [string]$Repository = "",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "gh CLI is required (https://cli.github.com/)."
}

if ([string]::IsNullOrWhiteSpace($Repository)) {
    $Repository = (gh repo view --json nameWithOwner -q .nameWithOwner).Trim()
}
if ([string]::IsNullOrWhiteSpace($Repository)) {
    throw "Cannot resolve repository. Pass owner/repo via -Repository."
}

$runId = (gh api "repos/$Repository/actions/workflows/ci.yml/runs?branch=$Branch&status=completed&per_page=20" --jq '.workflow_runs[] | select(.conclusion=="success") | .id' |
    Select-Object -First 1).Trim()
if ([string]::IsNullOrWhiteSpace($runId)) {
    throw "No successful ci.yml run found on branch '$Branch'. Run CI once first."
}

$jobNames = gh api "repos/$Repository/actions/runs/$runId/jobs?per_page=100" --jq '.jobs[].name'
$requiredChecks = @("server-ws-integration")
$missing = @()
foreach ($check in $requiredChecks) {
    if (-not ($jobNames -contains $check)) {
        $missing += $check
    }
}
if ($missing.Count -gt 0) {
    throw "Required CI checks not found in latest successful run: $($missing -join ', ')"
}

$payload = @{
    required_status_checks = @{
        strict = $true
        contexts = $requiredChecks
    }
    enforce_admins = $true
    required_pull_request_reviews = @{
        dismiss_stale_reviews = $true
        require_code_owner_reviews = $false
        required_approving_review_count = 1
    }
    restrictions = $null
}
$payloadJson = $payload | ConvertTo-Json -Depth 10

if ($DryRun) {
    Write-Host "Dry run. Would apply branch protection to $Repository:$Branch with payload:"
    Write-Host $payloadJson
    exit 0
}

$tmp = New-TemporaryFile
Set-Content -Path $tmp -Value $payloadJson -Encoding UTF8
gh api `
    --method PUT `
    -H "Accept: application/vnd.github+json" `
    "repos/$Repository/branches/$Branch/protection" `
    --input $tmp | Out-Null
Remove-Item $tmp -Force

Write-Host "Branch protection applied for $Repository:$Branch"
