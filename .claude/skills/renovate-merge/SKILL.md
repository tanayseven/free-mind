---
name: renovate-merge
description: "Use this skill when asked to process, fix, and merge a batch of GitHub PRs on a repository. Covers: pulling the repo, checking out each PR branch via gh CLI, running npm install and npm run scripts to surface errors, fixing those errors, pushing back to the PR branch, checking CI status, and rebasing + merging each PR into main in a loop."
compatibility: "Any environment with git, gh CLI, and Node.js/npm installed"
requires: "gh CLI authenticated (gh auth login), git configured with push access to the repo"
---

# GitHub PR Batch Review & Merge Loop For Renovate PRs

## Why this skill exists

When multiple PRs are open on a repo and each may have broken dependencies or
failing checks, you need a repeatable loop that:

1. Fetches the next open PR
2. Checks it out locally
3. Validates it (`go mod tidy`, `go build`, `go vet`, `go test`)
4. Fixes any errors
5. Pushes the fix back to the PR branch
6. Waits for CI and then merges into `main`

This skill documents the exact commands and error-handling logic for that loop.

---

## Prerequisites

```bash
# Verify tools are available
git --version
gh --version
go version

# Ensure gh is authenticated
gh auth status

# Ensure you are inside the target repo directory before starting
pwd
git remote -v
```

---

## Step 0 — Pull and rebase main

Always start from a clean, up-to-date `main`:

```bash
git checkout main
git fetch origin
git pull --rebase origin main
```

If there are uncommitted local changes blocking the pull:

```bash
git stash
git pull --rebase origin main
git stash pop
```

---

## Step 1 — List all open PRs

```bash
gh pr list --state open --json number,title,headRefName,author \
  --template '{{range .}}#{{.number}} {{.title}} (branch: {{.headRefName}}) by {{.author.login}}{{"\n"}}{{end}}'
```

Capture the numbers into a variable for the loop, filtering to only Renovate PRs (branches prefixed `renovate/`):

```bash
PR_NUMBERS=$(gh pr list --state open --json number,headRefName \
  --jq '[.[] | select(.headRefName | startswith("renovate/")) | .number] | sort | .[]')
echo "$PR_NUMBERS"
```

---

## Step 2 — Loop over each PR

```bash
for PR in $PR_NUMBERS; do
  echo "=========================================="
  echo "Processing PR #$PR"
  echo "=========================================="

  # --- 2a. Skip non-Renovate branches (safety check) ---
  BRANCH=$(gh pr view "$PR" --json headRefName --jq '.headRefName')
  if [[ "$BRANCH" != renovate/* ]]; then
    echo "PR #$PR branch '$BRANCH' is not a renovate/* branch — skipping"
    continue
  fi

  # --- 2b. Skip already-closed PRs ---
  STATUS=$(gh pr view "$PR" --json state --jq '.state')
  if [ "$STATUS" != "OPEN" ]; then
    echo "PR #$PR is $STATUS — skipping"
    continue
  fi

  # --- 2c. Check out the PR branch locally ---
  gh pr checkout "$PR"

  # Confirm which branch you are on
  git branch --show-current

  # --- 2d. Rebase the PR branch on top of latest main ---
  git fetch origin main
  git rebase origin/main
  # If the rebase has conflicts, resolve them (see Conflict Resolution section)
  # then: git rebase --continue

  # --- 2e. Tidy dependencies ---
  go mod tidy 2>&1 | tee /tmp/pr_modtidy.log

  # --- 2f. Build all packages ---
  go build ./... 2>&1 | tee /tmp/pr_build.log \
    || echo "BUILD FAILED — fix required"

  # --- 2g. Vet all packages ---
  go vet ./... 2>&1 | tee /tmp/pr_vet.log \
    || echo "VET FAILED — fix required"

  # --- 2h. Run tests ---
  go test ./... 2>&1 | tee /tmp/pr_test.log \
    || echo "TESTS FAILED — fix required"

  # --- 2i. Fix errors (see Fixing Errors section below) ---
  # ... make fixes ...

  # --- 2j. Commit and push fixes ---
  git add go.mod go.sum
  git add -u   # stage any other modified tracked files
  git diff --cached --quiet || git commit -m "fix: resolve errors for PR #$PR"
  git push origin HEAD

  # --- 2k. Check CI status ---
  echo "Waiting for CI checks on PR #$PR ..."
  gh pr checks "$PR" --watch   # streams until all checks complete

  # Verify final check status
  gh pr checks "$PR"

  # --- 2l. Rebase-merge into main ---
  gh pr merge "$PR" --rebase --delete-branch

  echo "PR #$PR merged and branch deleted."

  # Return to main and pull the merged changes before the next PR
  git checkout main
  git pull --rebase origin main

done

echo "All PRs processed."
```

---

## Fixing Errors

### go mod tidy errors

| Symptom | Fix |
|---|---|
| `no required module provides package X` | The package was removed upstream; check if an import needs removing or a new module path is needed |
| `go.sum: missing entry` | Run `go mod download` then `go mod tidy` again |
| `ambiguous import` | Two modules provide the same package — pin the correct one in `go.mod` with `replace` or remove the duplicate |
| `build constraints exclude all Go files` | Wrong GOOS/GOARCH in the environment; check with `go env` |

### Build errors

```bash
# Read the captured log
cat /tmp/pr_build.log

# Common patterns
# 1. "undefined: X" — a symbol was removed or renamed in the upgraded dependency
#    → check the module's changelog / release notes, update call sites

# 2. "cannot use X as type Y" — API signature changed
#    → update the call site to match the new signature

# 3. "import cycle" — rare after a dep bump, but possible
#    → refactor the cyclic import
```

### Vet errors

```bash
cat /tmp/pr_vet.log

# go vet errors are usually real bugs — fix them in source.
# Common: printf format mismatch, unreachable code, suspicious lock copying.
```

### Test failures

```bash
# Run a single package verbosely to narrow down
go test -v ./path/to/failing/pkg/...

# Run a single test by name
go test -v -run TestMyFunc ./...

# Update golden files / snapshots if the change is intentional
# (project-specific — check test helpers for update flags)

# Race detection
go test -race ./...
```

### Rebase conflicts

```bash
# See which files conflict
git status

# Open each conflicted file, resolve <<<<< ===== >>>>> markers, then:
git add <resolved-file>
git rebase --continue

# To abort and start over:
git rebase --abort
```

---

## Checking PR and CI status

```bash
# View PR overview (title, status, checks, reviews)
gh pr view "$PR"

# View only the CI checks (non-interactive)
gh pr chec￼￼Choose filesNo file chosen


ks "$PR"

# Watch checks stream live until they finish
gh pr checks "$PR" --watch

# View the PR diff to understand what changed
gh pr diff "$PR"

# View failed CI run logs
gh run list --branch "$(git branch --show-current)"
gh run view <run-id> --log-failed
```

---

## Merge strategies

```bash
# Rebase merge (keeps a linear history — recommended)
gh pr merge "$PR" --rebase --delete-branch

# Squash merge (collapses all commits into one)
gh pr merge "$PR" --squash --delete-branch

# Regular merge commit
gh pr merge "$PR" --merge --delete-branch
```

Use `--rebase` by default for a clean linear history on `main`.

---

## Handling edge cases

### PR is already merged or closed

```bash
STATUS=$(gh pr view "$PR" --json state --jq '.state')
if [ "$STATUS" != "OPEN" ]; then
  echo "PR #$PR is $STATUS — skipping"
  continue
fi
```

### CI checks are failing after your push

```bash
gh run list --branch "$(git branch --show-current)"
gh run view￼￼Choose filesNo file chosen


 <run-id> --log-failed
```

Fix the issue, commit again, and push. Checks re-trigger automatically.

### PR has requested changes / not approved

```bash
# See review status
gh pr view "$PR" --json reviewDecision,reviews

# If approval is required and not yet granted:
gh pr review "$PR" --approve   # if you have permission
```

### Push rejected (non-fast-forward)

```bash
git pull --rebase origin "$(git branch --show-current)"
# resolve any conflicts
git push origin HEAD
```

---

## Full self-contained example (condensed)

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$1"   # pass repo path as first argument
cd "$REPO_DIR"

git checkout main
git pull --rebase origin main
￼￼Choose filesNo file chosen



PR_NUMBERS=$(gh pr list --state open --json number,headRefName \
  --jq '[.[] | select(.headRefName | startswith("renovate/")) | .number] | sort | .[]')

for PR in $PR_NUMBERS; do
  STATUS=$(gh pr view "$PR" --json state --jq '.state')
  [ "$STATUS" != "OPEN" ] && echo "Skipping #$PR ($STATUS)" && continue

  gh pr checkout "$PR"
  git fetch origin main && git rebase origin/main

  go mod tidy
  go build ./... || echo "BUILD FAILED"
  go vet ./...   || echo "VET FAILED"
  go test ./...  || echo "TESTS FAILED"

  git add go.mod go.sum
  git add -u
  git diff --cached --quiet || git commit -m "fix: errors for PR #$PR"
  git push origin HEAD

  gh pr checks "$PR" --watch

  gh pr merge "$PR" --rebase --delete-branch

  git checkout main
  git pull --rebase origin main
done

echo "All PRs processed."
```

> **Tip:** Remove `|| echo` guards once you know which checks are relevant,
> so failures actually stop the loop and prompt you to fix them.

---

## Quick-reference cheatsheet

| Goal | Command |
|---|---|
| List open PRs | `gh pr list --state open` |
| Check out PR branch | `gh pr checkout <number>` |
| Tidy modules | `go mod tidy` |
| Build all packages | `go build ./...` |
| Vet all packages | `go vet ./...` |
| Run all tests | `go test ./...` |
| Run tests with race detector | `go test -race ./...` |
| View PR checks | `gh pr checks <number>` |
| Watch checks live | `gh pr checks <number> --watch` |
| Rebase-merge and delete | `gh pr merge <number> --rebase --delete-branch` |
| View failed CI logs | `gh run view <run-id> --log-failed` |
| Abort a bad rebase | `git rebase --abort` |