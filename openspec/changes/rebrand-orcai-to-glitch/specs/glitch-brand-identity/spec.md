## ADDED Requirements

### Requirement: Product is named GLITCH in all user-facing surfaces
All user-facing text SHALL refer to the product as GLITCH. References to "orcai", "ORCAI", "ABBS", or "Agentic Bulletin Board System" SHALL NOT appear in site copy, README, CLI help text, or UI labels.

#### Scenario: CLI help text shows GLITCH
- **WHEN** user runs `glitch --help`
- **THEN** output displays "glitch" as the command name and describes the product as GLITCH

#### Scenario: README opens with GLITCH branding
- **WHEN** user reads README.md
- **THEN** the product is introduced as GLITCH with no reference to ORCAI or ABBS

#### Scenario: Site title and meta reference GLITCH
- **WHEN** user visits powerglove.dev
- **THEN** page title, meta description, and visible copy use GLITCH as the product name

### Requirement: Binary is named glitch
The installed CLI binary SHALL be named `glitch`. The previous binary name `orcai` SHALL be replaced in all build targets, install scripts, and documentation.

#### Scenario: Build produces glitch binary
- **WHEN** user runs the build (make / task)
- **THEN** the output binary is named `glitch`, not `orcai`

#### Scenario: Install script installs glitch
- **WHEN** user runs install.sh
- **THEN** the binary installed to PATH is named `glitch`

### Requirement: GLITCH identity is open and user-owned
Product copy SHALL frame GLITCH as belonging to the user — their own AI hero — not as a branded corporate assistant. System prompts and UI copy SHALL reinforce that each user's GLITCH is their own.

#### Scenario: System prompt introduces GLITCH as the user's own
- **WHEN** GLITCH responds in the assistant panel
- **THEN** the persona feels personal and user-owned, not corporate

#### Scenario: Site copy uses "your GLITCH" framing
- **WHEN** user reads site marketing copy
- **THEN** copy uses "your GLITCH" or equivalent user-ownership language, not "our AI"

### Requirement: Domain is powerglove.dev
All user-facing URL references SHALL use powerglove.dev. References to orcai.* domains SHALL be replaced.

#### Scenario: README links point to powerglove.dev
- **WHEN** user follows links in README.md
- **THEN** links resolve to powerglove.dev

#### Scenario: Site canonical URL is powerglove.dev
- **WHEN** site is built
- **THEN** canonical meta tags and sitemap reference powerglove.dev
