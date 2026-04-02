## MODIFIED Requirements

### Requirement: SidecarSchema gains Mode and Signals fields
`SidecarSchema` in `internal/executor/cli_adapter.go` SHALL gain two new optional top-level fields:
- `Mode ModeBlock \`yaml:"mode,omitempty"\``
- `Signals []SignalDeclaration \`yaml:"signals,omitempty"\``

These fields SHALL be defined in the same file or a new `internal/executor/sidecar_schema.go` file. All existing fields and behaviour are unchanged.

#### Scenario: Existing sidecar unmarshals without error
- **WHEN** a sidecar YAML with no `mode:` or `signals:` keys is loaded
- **THEN** `SidecarSchema.Mode` is a zero-value `ModeBlock` and `SidecarSchema.Signals` is nil

#### Scenario: Full sidecar unmarshals all fields
- **WHEN** a sidecar YAML declares `kind`, `mode:`, and `signals:`
- **THEN** all three are populated in the returned `SidecarSchema`

### Requirement: ModeBlock struct
```go
type ModeBlock struct {
    Trigger     string `yaml:"trigger"`
    Logo        string `yaml:"logo"`
    Speaker     string `yaml:"speaker"`
    ExitCommand string `yaml:"exit_command"`
    OnActivate  string `yaml:"on_activate,omitempty"`
}
```

`ModeBlock.IsZero()` SHALL return true when `Trigger` is empty (used to detect absence).

### Requirement: SignalDeclaration struct
```go
type SignalDeclaration struct {
    Topic   string `yaml:"topic"`
    Handler string `yaml:"handler"`
}
```

### Requirement: No behaviour change to CliAdapter execution
`NewCliAdapterFromSidecar` and the `Execute` method are unchanged. The new fields are purely data; the executor pipeline is unaffected.

#### Scenario: CliAdapter executes normally when mode and signals are set
- **WHEN** a sidecar with `mode:` and `signals:` blocks is loaded as a pipeline step
- **THEN** it executes identically to a sidecar without those blocks
