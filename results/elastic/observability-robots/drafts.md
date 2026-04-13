## TBC Placeholders in elastic/observability-robots docs

Filtering to source docs only (excluding `.artifacts` build outputs which mirror them):

| File | Line | Context |
|------|------|---------|
| `docs/machine-user.md` | 161 | `- **Further details**: See TBC` |
| `docs/machine-user.md` | 180 | `> TBC` (Rotation section) |
| `docs/machine-user.md` | 189 | `- Buildkite (TBC)` |
| `docs/teams/ci/macos/index.md` | 10 | `> TBC` (entire section placeholder) |
| `docs/teams/ci/keyless/gh-ephemeral-tokens-gcp.md` | 169 | `TBC` |
| `docs/teams/ci/dependencies/updatecli.md` | 8 | `> TBC` (Structure section) |
| `docs/teams/ci/dependencies/updatecli.md` | 152 | `> TBC` (Local development section) |
| `docs/teams/ci/secrets-security-incident-runbook.md` | 172 | `**Rotation Procedure:** \`TBC\`` |
| `docs/teams/ci/secrets-security-incident-runbook.md` | 195 | `**AWS:** \`TBC\`` |
| `docs/teams/ci/secrets-security-incident-runbook.md` | 197 | `**GCP:** \`TBC\`` |

**10 TBC placeholders** across 5 source doc files. The heaviest concentration is in `secrets-security-incident-runbook.md` (3 placeholders for rotation/cloud procedures) and `machine-user.md` (3 placeholders for details, rotation, and Buildkite).

The `.artifacts/docs/html/` files are build outputs that mirror these same placeholders — fixing the source files would resolve those too.

---