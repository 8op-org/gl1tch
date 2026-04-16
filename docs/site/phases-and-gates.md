# Phases & Gates

- Structured execution stages with pass/fail verification checkpoints
- `(phase ...)` groups steps; `(gate ...)` assertions must pass before completion
- Gates are shell commands that must exit 0 — they run after all steps in the phase
- If a gate fails, the whole phase retries up to `:retries` times
- Real-world usage: site verification gates in `site-create-page.glitch`
- Phases compose with retry, timeout, catch, and other control flow
- Use phases when you need verification checkpoints; plain steps for simple linear workflows
