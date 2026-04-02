## ADDED Requirements

### Requirement: GameWorldPack carries XP formula weights
`GameWorldPack` SHALL include a `Weights` field of type `PackWeights`. `PackWeights` SHALL define: `BaseMultiplier` (float64, scales base XP), `CacheBonusRate` (float64, multiplier on cache_read_tokens), `SpeedBonusCap` (int64, reference milliseconds for speed bonus), `SpeedBonusScale` (float64, XP per ms under cap), `RetryPenalty` (int64, XP lost per retry), `StreakMultiplier` (float64, applied to final XP when streak_days > 1), and `ProviderWeights` (map[string]float64, keyed by provider string e.g. "providers.claude").

#### Scenario: Pack YAML includes weights section
- **WHEN** a pack YAML contains a `weights:` section
- **THEN** it unmarshals into GameWorldPack.Weights with all fields populated

#### Scenario: Missing weights section uses defaults
- **WHEN** a pack YAML has no `weights:` section
- **THEN** GameWorldPack.Weights is populated with values from DefaultPackWeights()

### Requirement: DefaultPackWeights reproduces current formula exactly
`DefaultPackWeights()` SHALL return a `PackWeights` that produces identical XP output to the pre-weights `ComputeXP()` implementation. BaseMultiplier=10.0, CacheBonusRate=0.5, SpeedBonusCap=1000, SpeedBonusScale=0.01, RetryPenalty=50, StreakMultiplier=1.0, ProviderWeights defaults to 1.0 for all providers.

#### Scenario: Default weights preserve existing XP values
- **WHEN** ComputeXP is called with DefaultPackWeights() and the same TokenUsage as before
- **THEN** the resulting XPResult.Final is identical to the pre-weights implementation

### Requirement: ComputeXP accepts PackWeights parameter
`ComputeXP(usage TokenUsage, retryCount int, weights PackWeights)` SHALL use `weights` for all formula coefficients. The provider multiplier for `usage.Provider` SHALL be applied to `XPResult.Final` after all other components are summed. If the provider key is not in `ProviderWeights`, a multiplier of 1.0 SHALL be used.

#### Scenario: Provider multiplier applied
- **WHEN** ComputeXP is called with a usage.Provider of "providers.claude" and weights.ProviderWeights["providers.claude"] = 1.2
- **THEN** XPResult.Final is 1.2× what it would be with multiplier 1.0

#### Scenario: Unknown provider uses 1.0 multiplier
- **WHEN** ComputeXP is called with a provider not in ProviderWeights
- **THEN** no multiplier is applied (effective multiplier 1.0)

#### Scenario: Streak multiplier applied when streak > 1
- **WHEN** the user has streak_days > 1 and weights.StreakMultiplier = 1.1
- **THEN** XPResult.Final is multiplied by 1.1

### Requirement: Pack YAML weights section has sensible defaults in cyberspace pack
The embedded `packs/cyberspace/pack.yaml` SHALL include a `weights:` section with values matching `DefaultPackWeights()`. This ensures the embedded pack is self-documenting and the tuner can read and evolve it from a known baseline.

#### Scenario: Embedded pack weights parse correctly
- **WHEN** the cyberspace pack is loaded via DefaultWorldPackLoader
- **THEN** Weights.BaseMultiplier == 10.0, Weights.CacheBonusRate == 0.5, Weights.ProviderWeights is non-nil
