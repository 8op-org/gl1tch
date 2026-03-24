---
name: bubble-scaffold
description: Scaffold a new BubbleTea component following orcai's Model/BubbleModel/Update/View pattern with Dracula palette styling and starter tests. Invoke with component name and optional fields.
disable-model-invocation: true
---

Generate a new BubbleTea component for orcai.

When invoked with a component name and optional field spec (e.g. `/bubble-scaffold mylist items:[]string selected:int`):

1. Read these files to understand patterns:
   - internal/promptbuilder/model.go — Model/BubbleModel struct split
   - internal/promptbuilder/view.go — lipgloss layout patterns
   - internal/promptbuilder/update.go (if exists) or bubble.go — Update() dispatch
   - internal/sidebar/sidebar.go — WindowSizeMsg handling, Dracula colors
   - sdk/styles/styles.go — palette constants (Purple, Pink, Cyan, Green, etc.)
   - internal/promptbuilder/bubble_test.go — test patterns

2. Parse arguments into:
   - Component name (CamelCase for types, snake_case for package)
   - Fields list (default: `width int`, `height int`, `focused bool`)

3. Generate these files in internal/<name>/:

   **model.go** — Data model:
   ```go
   package <name>

   type Model struct {
       width  int
       height int
       // user-specified fields
   }

   func New(<args>) Model { ... }
   ```

   **bubble.go** — BubbleTea integration:
   ```go
   package <name>

   import tea "github.com/charmbracelet/bubbletea"

   type BubbleModel struct {
       m Model
   }

   func (b BubbleModel) Init() tea.Cmd { return nil }

   func (b BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.WindowSizeMsg:
           b.m.width = msg.Width
           b.m.height = msg.Height
       case tea.KeyMsg:
           switch msg.String() {
           case "q", "ctrl+c":
               return b, tea.Quit
           }
       }
       return b, nil
   }

   func (b BubbleModel) View() string {
       // lipgloss layout stub
   }
   ```

   **<name>_test.go** — Starter tests (at least 6):
   - TestNew — constructor creates valid model
   - TestResize — WindowSizeMsg updates width/height
   - TestQuit — 'q' key returns quit cmd
   - TestCtrlC — ctrl+c returns quit cmd
   - TestEmptyState — model with zero values renders without panic
   - TestFocus — if focused field exists, test focus/blur

4. Run `go build ./internal/<name>/...` to verify it compiles
5. Report: files created, build status

Use lipgloss styles from sdk/styles/styles.go. Follow exact import paths from go.mod.
