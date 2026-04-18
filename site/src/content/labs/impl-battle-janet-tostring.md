---
title: "Implementation Battle: Tiered Routing vs Copilot-Only"
slug: "impl-battle-janet-tostring"
description: "Two workflows implement janet-lang/janet#1543 — one uses tier routing (free→paid→copilot), the other goes copilot-only. A judge picks the winner."
date: "2026-04-17"
---

## The Challenge

[janet-lang/janet#1543](https://github.com/janet-lang/janet/issues/1543) proposes a `:_tostring` protocol — a way for tables and structs with prototypes to define their own string representation. Currently, `(string my-table)` always produces `<table 0x…>`. The issue asks Janet to follow its existing `:_name` convention: if a prototype defines `:_tostring`, the runtime should call it instead of the default printer.

This is interesting because it touches three layers at once — the C runtime (`janet_to_string_b`), the pretty-printer (`janet_pretty_`), and the cooperative fiber scheduler (cycle detection must be fiber-local, not global). Getting any one of those wrong produces either a silent correctness failure or undefined behavior.

**Acceptance criteria (from spec):**

```janet
(defn Point [x y] (table/setproto @{:x x :y y} Point))
(put Point :_tostring (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))

(string (Point 3 4))  # => "Point(3, 4)"
(pp (Point 3 4))      # prints Point(3, 4)
# Existing tables without :_tostring print exactly as before
```

Six explicit requirements:

1. Add `:_tostring` as a protocol key on table/struct prototypes
2. `janet_to_string_b` must consult it before the default printer
3. The function receives `self` and must return a string
4. `describe` (and `pp`) must also respect the protocol
5. Must be safe against infinite recursion
6. Must not break existing behavior when `:_tostring` is absent

---

## The Rules

- **Workflow A — Tiered routing:** free model designs the approach, paid model implements, copilot sonnet reviews and refines
- **Workflow B — Copilot-only:** single shot, best model, no chain
- **Judge:** copilot sonnet scores both blind on correctness, safety, completeness, code quality, and test quality
- Neither workflow pushes code — the judge reviews output only

---

## Workflow A: Tiered Routing

### Step 1: Design (gemma-3-27b, free)

```
### Design for Custom :_tostring Protocol in Janet

## 1. Files to Modify
- src/core/pp.c — implements janet_to_string_b
- src/include/janet.h — optional new function signature
- src/core/table.c — prototype traversal

## 2. Prototype Chain Lookup

// In janet_to_string_b() for tables:
Janet func = janet_table_get(t, janet_wrap_symbol(":_tostring"));
if (func != janet_wrap_nil() && janet_type(func) == JANET_FUNCTION) {
    Janet args[1] = {x};
    Janet result = janet_call(func, 1, args);
    if (result != janet_wrap_nil()) {
        janet_to_string_b(buffer, result);
        return;
    }
}

## 3. Recursion Safety
Use a visited set tracking tables currently being processed.
static JanetSet *visited_set = NULL;

## 4. Test Cases
(defn Point [x y] (table/setproto @{:x x :y y} Point))
(put Point :_tostring (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))
(string (Point 3 4))  # => "Point(3, 4)"
```

The free-tier design correctly identifies the files, the prototype-chain lookup, and the need for recursion safety. It gets the API wrong in several places (`JanetSet` doesn't exist, keyword lookup uses the wrong function) — expected at this stage. The design is a scaffold for the implementation step.

### Step 2: Implementation (qwen3-32b, paid)

```diff
--- a/src/core/pp.c
+++ b/src/core/pp.c
@@ -152,11 +152,55 @@
  #include <float.h>
  
+/* Global visited set — later identified as non-fiber-safe */
+static JanetSet *visited_set = NULL;
+
+static void janet_tostring_init_visited(void) {
+    if (!visited_set) visited_set = janet_table_set();
+}
+
  void janet_to_string_b(JanetBuffer *buffer, Janet x) {
      switch (janet_type(x)) {
          case JANET_TABLE: {
              JanetTable *t = janet_unwrap_table(x);
+            if (!visited_set) janet_tostring_init_visited();
+            if (janet_set_has(visited_set, t)) {
+                string_description_b(buffer, "table", (void *)t);
+                break;
+            }
+            janet_set_add(visited_set, t);
+            /* Wrong: colon in keyword string, wrong type (symbol not keyword) */
+            Janet func = janet_table_get(t, janet_wrap_symbol(":_tostring"));
+            if (func != janet_wrap_nil() && janet_type(func) == JANET_FUNCTION) {
+                Janet args[1] = {x};
+                /* Bare janet_call — no error handling */
+                Janet result = janet_call(func, 1, args);
+                janet_set_remove(visited_set, t);
+                if (result != janet_wrap_nil()) {
+                    janet_to_string_b(buffer, result);
+                    break;
+                }
+            }
+            janet_set_remove(visited_set, t);
             string_description_b(buffer, "table", (void *)t);
              break;
          }
```

The paid-tier implementation faithfully transcribes the design's pseudocode — including its errors. `JanetSet`, global mutable state, `janet_wrap_symbol(":_tostring")`, and bare `janet_call` all land verbatim. Eight more defects (no JANET_STRUCT case, no pp update, wrong NaN-box comparisons, license deletion) are added on top.

### Step 3: Review & Refine (copilot sonnet)

The reviewer caught 11 critical defects and rewrote the patch entirely.

**Defects found:**

| # | Location | Problem |
|---|---|---|
| 1 | `JanetSet *visited_set` + all `janet_set_*` | `JanetSet` does not exist in Janet; APIs fabricated |
| 2 | `static JanetSet *visited_set` | Global mutable state is not fiber-safe; cooperative scheduler can re-enter mid-call |
| 3 | `janet_wrap_symbol(":_tostring")` | Colon is syntactic sugar; intern without it; wrong type (symbol vs keyword). Correct: `janet_ckeywordv("_tostring")` |
| 4 | `janet_call(func, 1, args)` | No error handling; panics propagate unchecked. Use `janet_pcall` |
| 5 | `func != janet_wrap_nil()` | Undefined behavior on NaN-boxed values. Use `janet_checktype` |
| 6 | `janet_set_remove` before result printed | Defeats the cycle guard if `:_tostring` returns `self` |
| 7 | First diff hunk | MIT license body partially deleted |
| 8 | `#include "table.h"` | Spurious; all table APIs already in `janet.h` |
| 9 | No `JANET_STRUCT` case | Spec requirement 2 unmet for structs |
| 10 | `janet_pretty_` not updated | `(pp ...)` acceptance criterion cannot pass |
| 11 | `janet_tostring_cleanup()` | Declared but never registered with a deinit hook |

**Corrected C patch:**

```diff
--- a/src/core/pp.c
+++ b/src/core/pp.c
@@ -148,6 +148,53 @@
  #include <float.h>
  
+/*
+ * janet_tostring_protocol — invoke :_tostring if defined on the prototype chain.
+ *
+ * Cycle safety: per-fiber dynamic variable "*tostring-visited*" (a JanetTable)
+ * records every value currently being stringified. If :_tostring calls
+ * (string self), the re-entrant check fires and falls back to the default
+ * representation instead of looping infinitely.
+ *
+ * Returns 1 and writes to buffer on success; 0 to use default printing.
+ */
+static int janet_tostring_protocol(JanetBuffer *buffer, Janet x, JanetTable *chain) {
+    if (!chain) return 0;
+
+    Janet fn_val = janet_table_get(chain, janet_ckeywordv("_tostring"));
+    if (!janet_checktype(fn_val, JANET_FUNCTION)) return 0;
+
+    Janet visited_dyn = janet_dyn("*tostring-visited*");
+    JanetTable *visited;
+    if (janet_checktype(visited_dyn, JANET_TABLE)) {
+        visited = janet_unwrap_table(visited_dyn);
+    } else {
+        visited = janet_table(4);
+        janet_setdyn("*tostring-visited*", janet_wrap_table(visited));
+    }
+
+    if (!janet_checktype(janet_table_rawget(visited, x), JANET_NIL)) return 0;
+    janet_table_put(visited, x, janet_wrap_true());
+
+    Janet result = janet_wrap_nil();
+    JanetFiber *fiber = NULL;
+    JanetSignal sig = janet_pcall(janet_unwrap_function(fn_val), 1, &x, &result, &fiber);
+
+    janet_table_remove(visited, x);
+
+    if (sig != JANET_SIGNAL_OK || !janet_checktype(result, JANET_STRING)) return 0;
+
+    janet_buffer_push_bytes(buffer,
+                            janet_unwrap_string(result),
+                            janet_string_length(janet_unwrap_string(result)));
+    return 1;
+}
+
  void janet_to_string_b(JanetBuffer *buffer, Janet x) {
      switch (janet_type(x)) {
  
@@ -274,6 +317,18 @@
      }
+    case JANET_TABLE: {
+        JanetTable *t = janet_unwrap_table(x);
+        if (!janet_tostring_protocol(buffer, x, t))
+            string_description_b(buffer, "table", (void *)t);
+        break;
+    }
+    case JANET_STRUCT: {
+        JanetStruct st = janet_unwrap_struct(x);
+        JanetTable *proto = janet_struct_proto(st);
+        if (!janet_tostring_protocol(buffer, x, proto))
+            string_description_b(buffer, "struct", (void *)st);
+        break;
+    }
  fallthrough:
  
@@ -705,6 +750,14 @@ static void janet_pretty_one(struct pretty *S, Janet x) {
      case JANET_STRUCT:
      case JANET_TABLE: {
          int istable = janet_checktype(x, JANET_TABLE);
+        {
+            JanetTable *chain = istable
+                ? janet_unwrap_table(x)
+                : janet_struct_proto(janet_unwrap_struct(x));
+            int32_t start = S->buffer->count;
+            if (janet_tostring_protocol(S->buffer, x, chain)) {
+                S->align += S->buffer->count - start;
+                break;
+            }
+        }
```

**Corrected tests:**

```janet
# test/core/tostring-protocol.janet

(def Point @{})
(defn make-point [x y] (table/setproto @{:x x :y y} Point))
(put Point :_tostring
  (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))

(assert (= "Point(3, 4)" (string (make-point 3 4)))   ":_tostring basic")
(assert (= "Point(0, 0)" (string (make-point 0 0)))   ":_tostring zero values")

(def buf @"")
(with-dyns [:out buf] (pp (make-point 3 4)))
(assert (= "Point(3, 4)\n" (string buf)) ":_tostring via pp")

(def plain @{:key "value"})
(assert (string/has-prefix? "<table " (string plain))
        "table without :_tostring uses default")

(def CycleProto @{})
(put CycleProto :_tostring (fn [self] (string self)))
(assert (string/has-prefix? "<table " (string (table/setproto @{} CycleProto)))
        "cyclic :_tostring falls back")

(def ErrProto @{})
(put ErrProto :_tostring (fn [self] (error "oops")))
(assert (string/has-prefix? "<table " (string (table/setproto @{} ErrProto)))
        ":_tostring error falls back")

(def BaseProto @{})
(def ChildProto (table/setproto @{} BaseProto))
(put BaseProto  :_tostring (fn [self] "base"))
(put ChildProto :_tostring (fn [self] "child"))
(assert (= "base"  (string (table/setproto @{} BaseProto)))  "inherited")
(assert (= "child" (string (table/setproto @{} ChildProto))) "overridden")

(def StructProto @{})
(put StructProto :_tostring (fn [self] (string/format "S(%s)" (self :name))))
(assert (= "S(test)" (string (struct/with-proto StructProto :name "test")))
        ":_tostring on struct prototype")
```

**Corrected documentation:**

```
## :_tostring — Custom string representation for tables and structs

Tables and structs with a prototype can define :_tostring on that prototype
to control how they appear in string, describe, and pp.

  (def Point @{})
  (put Point :_tostring
       (fn [self]
         (string/format "Point(%d, %d)" (self :x) (self :y))))
  (defn make-point [x y] (table/setproto @{:x x :y y} Point))
  (string (make-point 3 4))   # => "Point(3, 4)"
  (pp (make-point 3 4))       # prints Point(3, 4)

Contract:
- Defined on the prototype (class-level, not instance-level)
- Receives the table/struct as its sole argument
- Must return a string; any other return type gives silent fallback
- Errors in :_tostring are not propagated; fallback to default
- Recursive (string self) inside :_tostring caught by fiber-local
  visited table; falls back to default instead of looping

Tables without :_tostring are completely unaffected.
```

---

## Workflow B: Copilot-Only (sonnet, single shot)

```diff
--- a/src/core/pp.c
+++ b/src/core/pp.c
@@ -218,6 +218,47 @@ static void janet_escape_buffer_b(JanetBuffer *buffer, JanetBuffer *bx) {
      janet_escape_string_impl(buffer, bx->data, bx->count);
  }
  
+/* Look up :_tostring in the prototype chain of a table or struct.
+ * Key lives on the prototype (class-level), not the object itself.
+ * Returns 1 if handled, 0 to fall back to default representation. */
+static int proto_tostring_b(JanetBuffer *buffer, Janet x) {
+    if (!janet_vm.fiber) return 0;
+    Janet tostring = janet_wrap_nil();
+    if (janet_checktype(x, JANET_TABLE)) {
+        JanetTable *t = janet_unwrap_table(x);
+        if (t->proto) {
+            tostring = janet_table_get(t->proto, janet_ckeywordv("_tostring"));
+        }
+    } else if (janet_checktype(x, JANET_STRUCT)) {
+        /* BUG: janet_struct_proto returns JanetTable*, not const JanetKV* */
+        const JanetKV *proto = janet_struct_proto(janet_unwrap_struct(x));
+        if (proto) {
+            tostring = janet_struct_get(proto, janet_ckeywordv("_tostring"));
+        }
+    } else {
+        return 0;
+    }
+    if (janet_checktype(tostring, JANET_NIL)) return 0;
+    /* BREAKS existing behavior: panics instead of falling back */
+    if (!janet_checktype(tostring, JANET_FUNCTION) &&
+            !janet_checktype(tostring, JANET_CFUNCTION)) {
+        janet_panicf(":_tostring must be a function, got %v", tostring);
+    }
+    Janet argv_buf[1];
+    argv_buf[0] = x;
+    Janet result;
+    if (janet_checktype(tostring, JANET_FUNCTION)) {
+        /* No pcall — errors propagate unchecked */
+        result = janet_call(janet_unwrap_function(tostring), 1, argv_buf);
+    } else {
+        result = janet_unwrap_cfunction(tostring)(1, argv_buf);
+    }
+    const uint8_t *bytes;
+    int32_t len;
+    /* BREAKS existing behavior: panics instead of falling back */
+    if (!janet_bytes_view(result, &bytes, &len)) {
+        janet_panicf(":_tostring must return a string or buffer, got %v", result);
+    }
+    janet_buffer_push_bytes(buffer, bytes, len);
+    return 1;
+}
+
  void janet_to_string_b(JanetBuffer *buffer, Janet x) {
      switch (janet_type(x)) {
  
@@ -285,6 +326,10 @@
+        case JANET_TABLE:
+        case JANET_STRUCT:
+            if (proto_tostring_b(buffer, x)) return;
+            goto fallthrough;
      fallthrough:
          default:
              string_description_b(buffer, janet_type_names[janet_type(x)],
@@ -660,6 +705,14 @@ static void janet_pretty_one(struct pretty *S, Janet x) {
      case JANET_STRUCT:
      case JANET_TABLE: {
          int istable = janet_checktype(x, JANET_TABLE);
+        {
+            int32_t start = S->buffer->count;
+            if (proto_tostring_b(S->buffer, x)) {
+                S->align += S->buffer->count - start;
+                break;
+            }
+        }
```

```janet
# test/tostring-protocol.janet
(import spork/test :as t)

(def Point @{})
(put Point :_tostring
     (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))
(defn make-point [x y] (table/setproto @{:x x :y y} Point))
(def p (make-point 3 4))

(t/assert (= (string p) "Point(3, 4)")   "string uses :_tostring")
(t/assert (= (describe p) "Point(3, 4)") "describe uses :_tostring")

(def VecProto {:_tostring (fn [self]
                            (string/format "Vec(%d %d %d)"
                                           (self :x) (self :y) (self :z)))})
(def v (struct/with-proto VecProto :x 1 :y 2 :z 3))
(t/assert (= (string v) "Vec(1 2 3)") "struct :_tostring")

(def plain @{:a 1 :b 2})
(t/assert (string/find "table" (describe plain)) "plain table uses default")

(def Named @{:_name "Named"
             :_tostring (fn [self] (string/format "Named<%d>" (self :v)))})
(t/assert (= (string (table/setproto @{:v 42} Named)) "Named<42>")
          ":_tostring over :_name")

# Tests written for B's panic contract (not graceful fallback)
(def BadProto @{:_tostring "not-a-function"})
(t/assert (try (string (table/setproto @{} BadProto))
               ([e] (string/find ":_tostring must be a function" e)))
          "non-function :_tostring raises error")

(def RecProto @{})
(put RecProto :_tostring (fn [self] (string self)))
(t/assert (try (string (table/setproto @{} RecProto)) ([_] true))
          "infinite recursion caught by stack guard")
```

```
## :_tostring — Custom string representation

  (def Point @{})
  (put Point :_tostring
       (fn [self] (string/format "Point(%d, %d)" (self :x) (self :y))))
  (defn point [x y] (table/setproto @{:x x :y y} Point))

  (string (point 3 4))    # => "Point(3, 4)"
  (describe (point 3 4))  # => "Point(3, 4)"
  (pp (point 3 4))        # prints Point(3, 4)

Rules:
- Defined on the prototype, not the object itself (class-level, like :_name)
- Receives the table or struct as its sole argument
- Must return a string, symbol, keyword, or buffer
- Resolved through the full prototype chain
- Non-function :_tostring raises a runtime error
- Non-string return raises a runtime error
- Infinite recursion caught by Janet's C-stack recursion guard (panics)
```

---

## The Verdict

| Criterion | Workflow A (tiered, final) | Workflow B (copilot-only) | Weight |
|---|---|---|---|
| **Correctness** | 8 / 10 | 5 / 10 | High |
| **Safety** | 9 / 10 | 3 / 10 | Critical |
| **Completeness** | 8 / 10 | 6 / 10 | High |
| **Code quality** | 8 / 10 | 5 / 10 | Medium |
| **Test quality** | 9 / 10 | 5 / 10 | Medium |
| **Total** | **42 / 50** | **24 / 50** | |

**Winner: Workflow A**

> Implementation A wins decisively. Its use of `janet_pcall` with graceful silent fallback is the only approach that satisfies spec requirements 5 and 6 without turning `(string x)` into a foot-gun. Implementation B's triple-critical combination — bare `janet_call`, panic-on-bad-value, and a struct prototype type mismatch (`JanetTable*` assigned to `const JanetKV*`) — produces undefined behavior and a broken safety contract. A's fiber-local visited-table for cycle detection is also architecturally correct in a cooperative VM; B has no equivalent.

**What happened at each step:**

- **gemma-3-27b (design):** Correctly identified the files and prototype-chain lookup strategy. Got API details wrong — expected for a free-tier design pass.
- **qwen3-32b (implementation):** Faithfully transcribed the design's pseudocode including its errors. Did not independently fix anything.
- **copilot sonnet (review):** Caught all 11 defects, rewrote with correct APIs (`janet_ckeywordv`, `janet_pcall`, `janet_dyn`/`janet_setdyn`), and produced valid Janet tests. The review pass is where the tiered approach earned its score.
- **copilot sonnet (single shot):** Produced cleaner C structure but introduced three critical safety violations and a struct type bug causing undefined behavior in production.

---

## Cost Comparison

| | Workflow A (tiered) | Workflow B (copilot) |
|---|---|---|
| **Steps** | 3 | 1 |
| **Models used** | gemma-3-27b, qwen3-32b, sonnet | sonnet |
| **Approach** | design → implement → review | single shot |
| **Token cost (est.)** | Low + Medium + Medium | High |
| **Final score** | 42 / 50 | 24 / 50 |
| **Safety violations** | 0 | 3 critical |

---

## Takeaway

The tiered approach didn't win because the free and paid models were good — they weren't; both made critical API errors. It won because the review step had a complete, incorrect implementation to critique rather than a blank page, and that critique is where the hardest correctness work happened. The single-shot model, given no prior work to react to, optimized for structural elegance over correctness and produced three safety violations the review pass would have caught. For C runtime patches in a cooperative VM, a design-implement-review chain is cheaper and safer than asking one model to hold the whole spec in context at once.