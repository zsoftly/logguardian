# Branch Protection Rules

## Main Branch Protection

To enforce the release branch workflow and prevent direct pushes to main, configure the following branch protection rules in GitHub:

### GitHub Settings → Branches → Add Rule

**Branch name pattern:** `main`

### Protection Settings

✅ **Require a pull request before merging**
- ✅ Require approvals: 1
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from CODEOWNERS (optional)

✅ **Require status checks to pass before merging**
- ✅ Require branches to be up to date before merging
- Required status checks:
  - `CI / Lint and Code Quality`
  - `CI / Security Analysis`
  - `CI / Test`
  - `CI / Build`

✅ **Require conversation resolution before merging**

✅ **Require signed commits** (optional but recommended)

✅ **Include administrators** (enforce for everyone)

✅ **Restrict who can push to matching branches** (optional)
- Add specific users or teams who can merge PRs

### Do NOT Enable
- ❌ Allow force pushes
- ❌ Allow deletions
- ❌ Allow bypass of required pull requests

## Release Branch Workflow

With these protections in place:

1. **No direct pushes to main** - All changes must go through a PR
2. **Release branches** (`release/*`) must be created for version updates
3. **Feature branches** (`feature/*`, `feat/*`, `issues/*`) for development
4. **All PRs must pass CI** - Tests, linting, security checks
5. **PRs require approval** - At least one reviewer must approve

## Exceptions

Only repository administrators can:
- Merge PRs to main (after approval)
- Create tags from main

## Branch Retention Policy

- **Release branches** (`release/*`) are NEVER deleted - kept for historical reference
- **Feature branches** can be deleted after merge
- **Main branch** is protected and cannot be deleted

## Setting Up Protection

Branch protection can be configured via:
- GitHub UI: Settings → Branches → Add rule
- GitHub CLI: Using `gh api` commands
- GitHub API: Direct API calls

## Verifying Protection

Check current protection status:
- GitHub UI: Settings → Branches → View rules
- GitHub CLI: `gh api repos/:owner/:repo/branches/main/protection`