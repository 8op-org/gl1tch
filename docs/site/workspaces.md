# Workspaces

- A workspace is project-scoped configuration: model, provider, ES URL, and default params
- Lives in `workspace.glitch` at your project root
- Uses the same s-expression syntax as workflows
- `--workspace` flag on any command loads the workspace config
- Workspace name resolves automatically by walking up from your current directory
- Elasticsearch URL flows into workflow runs automatically
- Workflows and results resolve relative to the workspace directory
