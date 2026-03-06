#!/usr/bin/env bash

set -euo pipefail

BRANCH="${1:-main}"
REPO="${2:-}"
DRY_RUN="${DRY_RUN:-0}"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required (https://cli.github.com/)." >&2
  exit 1
fi

if [[ -z "$REPO" ]]; then
  REPO="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
fi

if [[ -z "$REPO" ]]; then
  echo "Cannot resolve repository. Pass owner/repo as second arg." >&2
  exit 1
fi

RUN_ID="$(gh api "repos/$REPO/actions/workflows/ci.yml/runs?branch=$BRANCH&status=completed&per_page=20" \
  --jq '.workflow_runs[] | select(.conclusion=="success") | .id' | head -n 1)"

if [[ -z "$RUN_ID" ]]; then
  echo "No successful ci.yml run found on branch '$BRANCH'." >&2
  echo "Run CI at least once before applying branch protection." >&2
  exit 1
fi

mapfile -t JOB_NAMES < <(gh api "repos/$REPO/actions/runs/$RUN_ID/jobs?per_page=100" --jq '.jobs[].name')

REQUIRED=("server-ws-integration" "android-e2e-emulator")
MISSING=()
for req in "${REQUIRED[@]}"; do
  found=0
  for job in "${JOB_NAMES[@]}"; do
    if [[ "$job" == "$req" ]]; then
      found=1
      break
    fi
  done
  if [[ $found -eq 0 ]]; then
    MISSING+=("$req")
  fi
done

if [[ ${#MISSING[@]} -gt 0 ]]; then
  echo "Required CI checks not found in latest successful run: ${MISSING[*]}" >&2
  echo "Found jobs: ${JOB_NAMES[*]}" >&2
  exit 1
fi

PAYLOAD_FILE="$(mktemp)"
cat > "$PAYLOAD_FILE" <<JSON
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["server-ws-integration", "android-e2e-emulator"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1
  },
  "restrictions": null
}
JSON

if [[ "$DRY_RUN" == "1" ]]; then
  echo "Dry run. Would apply branch protection to $REPO:$BRANCH with payload:"
  cat "$PAYLOAD_FILE"
  rm -f "$PAYLOAD_FILE"
  exit 0
fi

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "repos/$REPO/branches/$BRANCH/protection" \
  --input "$PAYLOAD_FILE" >/dev/null

rm -f "$PAYLOAD_FILE"
echo "Branch protection applied for $REPO:$BRANCH"
