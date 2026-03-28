## 1. Fix field preservation in confirmEdit

- [x] 1.1 In `confirmEdit()` (`internal/crontui/update.go:261-267`), add `Args: ov.original.Args` and `WorkingDir: ov.original.WorkingDir` to the `updated` entry literal
- [x] 1.2 Verify the rename path (line 286-288) still correctly removes the old entry by original name

## 2. Tests

- [x] 2.1 Add a table-driven test for `confirmEdit` that confirms `Args` and `WorkingDir` are preserved after a rename
- [x] 2.2 Add a test case confirming new entries (no original) still get zero values for both fields
