# Use Case 1: Wrapping curl for a REST API

## Scenario

You own an internal REST API. You hit it with `curl` all the time. You remember the base URL, the auth header, the right `Content-Type`, and the pagination params. It’s muscle memory.

Your teammates don’t have that muscle memory. Every time someone asks “how do I call this endpoint?”, you paste a `curl` one-liner from Slack history. Then you paste the follow-up for auth. Then you paste the follow-up for pagination. It works, but it’s the same conversation on repeat.

One day you notice you’ve copy-pasted the same three commands six times in a month. That’s the moment you realize it’s a pattern. The agent shouldn’t need you as a human man page.

---

## The manual session (before the skill exists)

This is what it looks like when there’s no wrapper. The agent can run commands, but it has to re-learn your API conventions every time.

```
You: Can you list the last 20 events for user 123?

Agent: What’s the base URL and how do I auth?

You: Base is https://api.internal/v1. Auth is Bearer token in $INTERNAL_API_TOKEN.

Agent: Any default headers? Pagination? JSON?

You: Always JSON. Pagination is page/per_page.

Agent: OK. Running:
  curl -sS -m 30 -X GET \
    "https://api.internal/v1/users/123/events?page=1&per_page=20" \
    -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
    -H "Accept: application/json"
```

Next time, different endpoint. Same briefing.

```
You: Can you POST a new comment on ticket 77?

Agent: Same questions again: base URL, auth header, content type?

You: Yes. Same as before. Also, POST needs Content-Type: application/json.
```

The commands are not the hard part. The boring, repeated context is.

---

## Pattern recognition: why this is a Tool Wrapper

This is a wrapper because it’s not “run curl”. It’s “run curl the way we run curl here”.

The skill encapsulates:

- **Command syntax**: flags you always want (`-m`, `-sS`, `--fail`), and when to use `-X`, `-d`, `-H`.
- **Auth + preconditions**: where the token comes from, and how to check it’s present.
- **Domain conventions**: base URLs, endpoints, pagination params, and what errors usually mean.

A good wrapper stops the agent from asking the same five setup questions, and stops teammates from cargo-culting half-correct one-liners from Slack.

---

## Skill creation: capturing it as a SKILL.md

You do a few manual sessions until you can describe the workflow in plain English:

- “Base URL is X.”
- “Auth is Bearer token from env var Y.”
- “Pagination uses `page`/`per_page`.”
- “POST/PUT must send JSON and set Content-Type.”
- “Timeouts and non-2xx should fail fast.”

Then you run `skill-creator` in that same agent session and ask it to produce a wrapper skill.

The output is a `SKILL.md` you can commit to `.github/skills/`.

---

## The result: a complete curl-wrapper `SKILL.md`

Copy-paste this into a file like `.github/skills/internal-api-curl/SKILL.md` and adapt the URLs, headers, and endpoints.

```markdown
---
name: internal-api-curl
description: Wrapper around curl for calling the internal REST API. Adds auth/header defaults, pagination conventions, and basic troubleshooting.
trigger-phrases:
  - "curl the internal api"
  - "hit api.internal"
  - "call the REST API with curl"
  - "curl POST"
  - "curl GET"
allowed-tools:
  - RunInTerminal
---

# Internal API curl wrapper

Use this skill whenever you need to call the internal REST API with `curl`.

## Defaults

- Base URL (prod): `https://api.internal/v1`
- Base URL (staging): `https://staging.api.internal/v1`
- Default environment: prod (unless the user explicitly says staging)

## Prerequisites (fail early)

Before running any request, verify the token env var is present:

```sh
echo "$INTERNAL_API_TOKEN"
```

If it’s empty, stop and tell the user to export it. Do not ask them to paste tokens into chat.

## Request rules

Always include:

- `-sS` (quiet, but still shows errors)
- `--fail` (treat non-2xx as failure)
- `-m 30` (timeout; don't hang forever)
- `-H "Authorization: Bearer $INTERNAL_API_TOKEN"`
- `-H "Accept: application/json"`

For `POST` and `PUT`, also include:

- `-H "Content-Type: application/json"`

Never use `-k/--insecure`. If TLS fails, stop and tell the user to fix certs/CA.

## Intent → command mapping

### GET (list)

Use pagination params `page` and `per_page`.

```sh
curl -sS --fail -m 30 \
  -X GET "https://api.internal/v1/events?page=1&per_page=50" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json"
```

### GET (single resource)

```sh
curl -sS --fail -m 30 \
  -X GET "https://api.internal/v1/tickets/77" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json"
```

### POST (JSON)

Keep JSON small and inline unless it's big. If it's big, write a temp file and use `-d @file.json`.

```sh
curl -sS --fail -m 30 \
  -X POST "https://api.internal/v1/tickets/77/comments" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"body":"hello from curl"}'
```

### PUT (JSON)

```sh
curl -sS --fail -m 30 \
  -X PUT "https://api.internal/v1/tickets/77" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"status":"closed"}'
```

### DELETE

```sh
curl -sS --fail -m 30 \
  -X DELETE "https://api.internal/v1/tickets/77/comments/12" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json"
```

## Troubleshooting (common failures)

- **401**: token missing/expired. Check `$INTERNAL_API_TOKEN`.
- **403**: token is valid but lacks permission.
- **404**: wrong path or ID.
- **429**: rate limit. Wait, then retry once.
- **5xx**: server problem. Retry once, then stop.

## Output handling

If the user asks to extract fields, pipe to `jq` in the same terminal command.
Example:

```sh
curl -sS --fail -m 30 \
  -X GET "https://api.internal/v1/tickets/77" \
  -H "Authorization: Bearer $INTERNAL_API_TOKEN" \
  -H "Accept: application/json" \
| jq -r '.status'
```
```

---

## What makes this a good wrapper

- It kills repeated setup questions (base URL, auth header, JSON defaults).
- It makes safe defaults the path of least resistance (timeouts, fail-fast, no `-k`).
- It maps intent to concrete commands (GET/POST/PUT/DELETE patterns).
- It encodes your domain conventions (pagination param names, env var names, staging vs prod).
