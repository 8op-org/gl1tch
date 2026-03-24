---
name: release
description: Bump semantic version, generate CHANGELOG entry from git commits, tag the release, build Docker image, and create GitHub release. Invoke with bump type: major, minor, or patch.
disable-model-invocation: true
---

Create a new orcai release.

When invoked with a bump type as argument (e.g. `/release patch` or `/release minor`):

1. **Determine current version**:
   - Check for internal/version/version.go or cmd/version.go
   - Fall back to latest git tag: `git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"`
   - If no version infrastructure exists, create internal/version/version.go:
     ```go
     package version
     const Version = "v0.1.0"
     ```

2. **Calculate new version**:
   - Parse current semver (major.minor.patch)
   - Apply bump: major resets minor+patch to 0, minor resets patch to 0
   - Format: v<major>.<minor>.<patch>

3. **Collect commits since last tag**:
   - `git log <last-tag>..HEAD --oneline --no-merges`
   - Group by prefix: feat(…) → Features, fix(…) → Bug Fixes, chore(…) → Maintenance, refactor(…) → Refactoring
   - Skip commits with "Co-Authored-By" only lines

4. **Update version file** with new version string

5. **Generate CHANGELOG entry**:
   ```markdown
   ## v<new> — <date>

   ### Features
   - …

   ### Bug Fixes
   - …

   ### Maintenance
   - …
   ```
   Prepend to CHANGELOG.md (create if missing)

6. **Commit and tag**:
   ```bash
   git add internal/version/version.go CHANGELOG.md
   git commit -m "chore: release v<new>"
   git tag v<new>
   ```

7. **Build Docker image** (if Docker is available):
   ```bash
   docker build -t orcai:v<new> -t orcai:latest . 2>&1 | tail -5
   ```

8. **Create GitHub release**:
   ```bash
   gh release create v<new> \
     --title "orcai v<new>" \
     --notes "<changelog entry>" \
     --latest
   ```

9. **Report**: new version, tag created, release URL

Note: Requires `gh auth` and optionally Docker. Skips steps gracefully if tools unavailable.
