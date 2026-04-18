---
title: "Implementation Battle: Tiered Routing vs Copilot-Only"
slug: "impl-battle-janet-tostring"
description: "Two workflows implement janet-lang/janet#1543 — one uses tier routing (free→paid→copilot), the other goes copilot-only. A judge picks the winner."
date: "2026-04-17"
---

## The Challenge

[janet-lang/janet#1543](https://github.com/janet-lang/janet/issues/1543) — add a `:_tostring` protocol so tables and structs can define their own string representation.

```janet
(def Point @{})
(put Point :_tostring
  (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))

(string (table/setproto @{:x 3 :y 4} Point))  # => "Point(3, 4)"
```

Touches three layers: the C runtime (`janet_to_string_b`), the pretty-printer, and the cooperative fiber scheduler. Getting any one wrong produces silent correctness failures or undefined behavior.

---

## The Rules

| | Workflow A | Workflow B |
|---|---|---|
| **Strategy** | Tiered routing | Single shot |
| **Models** | qwen3-8b → qwen3.5-flash → copilot sonnet | copilot sonnet |
| **Steps** | design → implement → review | one pass |

Judge: copilot sonnet scores both on correctness, safety, completeness, code quality, test quality.

---

## Workflow A: Tiered Routing

**Step 1 — Design (qwen3-8b, local free).** Identified the right files and prototype-chain lookup. Got API details wrong (`JanetSet` doesn't exist, wrong keyword syntax) — expected for a free model. Produced a scaffold, not a patch.

**Step 2 — Implementation (qwen3.5-flash, paid).** Faithfully transcribed the design including its errors. `JanetSet`, global mutable state, `janet_wrap_symbol(":_tostring")`, bare `janet_call` — all landed verbatim. Did not independently fix anything.

**Step 3 — Review (copilot sonnet).** Caught 11 defects. Rewrote the entire patch.

Key fixes:

| Defect | Fix |
|---|---|
| `JanetSet` (fabricated API) | Fiber-local `janet_dyn` visited table |
| Global mutable state | Per-fiber dynamic variable |
| `janet_wrap_symbol(":_tostring")` | `janet_ckeywordv("_tostring")` |
| Bare `janet_call` | `janet_pcall` with graceful fallback |
| No struct support | Added `JANET_STRUCT` case |
| No pretty-printer update | Updated `janet_pretty_one` |

The review step is where the tiered approach earned its score. Final implementation uses `janet_pcall` for error safety, fiber-local cycle detection, and handles both tables and structs:

```c
static int janet_tostring_protocol(JanetBuffer *buffer, Janet x,
                                    JanetTable *chain) {
    if (!chain) return 0;
    Janet fn_val = janet_table_get(chain, janet_ckeywordv("_tostring"));
    if (!janet_checktype(fn_val, JANET_FUNCTION)) return 0;

    /* Fiber-local visited set for cycle detection */
    Janet visited_dyn = janet_dyn("*tostring-visited*");
    JanetTable *visited;
    if (janet_checktype(visited_dyn, JANET_TABLE)) {
        visited = janet_unwrap_table(visited_dyn);
    } else {
        visited = janet_table(4);
        janet_setdyn("*tostring-visited*", janet_wrap_table(visited));
    }
    if (!janet_checktype(janet_table_rawget(visited, x), JANET_NIL))
        return 0;
    janet_table_put(visited, x, janet_wrap_true());

    Janet result = janet_wrap_nil();
    JanetFiber *fiber = NULL;
    JanetSignal sig = janet_pcall(
        janet_unwrap_function(fn_val), 1, &x, &result, &fiber);
    janet_table_remove(visited, x);

    if (sig != JANET_SIGNAL_OK ||
        !janet_checktype(result, JANET_STRING)) return 0;

    janet_buffer_push_bytes(buffer,
        janet_unwrap_string(result),
        janet_string_length(janet_unwrap_string(result)));
    return 1;
}
```

---

## Workflow B: Copilot-Only

Single shot produced cleaner C structure but introduced three critical safety violations:

| Problem | Impact |
|---|---|
| Bare `janet_call` instead of `janet_pcall` | Errors propagate unchecked — crashes the VM |
| `janet_panicf` on bad `:_tostring` value | `(string x)` becomes a foot-gun |
| Wrong struct proto type (`JanetTable*` → `const JanetKV*`) | Undefined behavior |
| No cycle detection | Infinite recursion on `(string self)` |

The implementation panics on errors instead of falling back gracefully — violating spec requirement 6 ("must not break existing behavior"). Tests were written for the panic contract, so they pass against B but would fail against the spec.

---

## The Verdict

| Criterion | A (tiered) | B (copilot) |
|---|---|---|
| Correctness | 8 | 5 |
| Safety | **9** | **3** |
| Completeness | 8 | 6 |
| Code quality | 8 | 5 |
| Test quality | 9 | 5 |
| **Total** | **42 / 50** | **24 / 50** |

**Winner: Workflow A.**

> A's use of `janet_pcall` with graceful silent fallback is the only approach that satisfies the spec without turning `(string x)` into a foot-gun. B's triple-critical combination — bare `janet_call`, panic-on-bad-value, and a struct prototype type mismatch — produces undefined behavior and a broken safety contract.

---

## Takeaway

The tiered approach didn't win because the free and paid models were good — they weren't. Both made critical API errors. It won because the review step had a complete, incorrect implementation to critique, and that critique is where the hardest correctness work happened. The single-shot model optimized for structural elegance over correctness and produced three safety violations the review pass would have caught.

For C runtime patches in a cooperative VM, a design-implement-review chain is cheaper and safer than asking one model to hold the whole spec in context at once.
