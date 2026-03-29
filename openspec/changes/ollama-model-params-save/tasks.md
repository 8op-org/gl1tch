## 1. orcai-ollama: Inference Options Support

- [x] 1.1 Add `stringSlice` type (or `flag.Func`) to `orcai-ollama` to collect repeated `--option key=value` flags
- [x] 1.2 Update `generateRequest` struct to include `Options map[string]interface{} \`json:"options,omitempty"\``
- [x] 1.3 Parse collected `--option` flags into the options map, coercing integer-parseable strings to `int`
- [x] 1.4 Pass the options map into `callOllama` and include it in the JSON body
- [x] 1.5 Write tests: options absent when no flag given, single option forwarded as int, multiple options forwarded, string option preserved as string

## 2. orcai-ollama: Create Model Command

- [x] 2.1 Add `--create-model <name>` flag to the `run` function
- [x] 2.2 Implement `createModel(base, name string, options map[string]interface{}, stderr io.Writer) error` that builds a Modelfile string and pipes it to `ollama create <name> -f -`
- [x] 2.3 When `--create-model` is set, call `createModel` and print `Created model '<name>'` to stdout, then return (skip inference)
- [x] 2.4 Return an error if `--create-model` is set but no model is resolved
- [x] 2.5 Write tests: creates correct Modelfile content, exits without inference, error on missing base model

## 3. Create Custom Ollama Model Aliases Locally

- [x] 3.1 Run `orcai-ollama --model llama3.2 --option num_ctx=16384 --create-model llama3.2-16k` to create the llama alias
- [x] 3.2 Run `orcai-ollama --model qwen2.5 --option num_ctx=16384 --create-model qwen2.5-16k` to create the qwen2.5 alias
- [x] 3.3 Run `orcai-ollama --model qwen3:8b --option num_ctx=16384 --create-model qwen3:8b-16k` to create the qwen3:8b alias
- [x] 3.4 Verify all three aliases appear in `ollama list`

## 4. Update Local Wrapper Configs

- [x] 4.1 Add `llama3.2-16k` and `qwen2.5-16k` model entries to `~/.config/orcai/wrappers/ollama.yaml`
- [x] 4.2 Add `ollama/llama3.2-16k`, `ollama/qwen2.5-16k`, and `ollama/qwen3:8b-16k` entries to `~/.config/orcai/wrappers/opencode.yaml`

## 5. Update Plugin Repo YAML Defaults

- [x] 5.1 Add `llama3.2-16k` and `qwen2.5-16k` to `orcai-plugins/plugins/ollama/ollama.yaml` models list with descriptive labels
- [x] 5.2 Add `ollama/llama3.2-16k`, `ollama/qwen2.5-16k`, and `ollama/qwen3:8b-16k` to `orcai-plugins/plugins/opencode/opencode.yaml` models list

## 6. Documentation

- [x] 6.1 Write `orcai-plugins/plugins/ollama/README.md` documenting `--option`, `--create-model`, and the full workflow to create high-context models
- [x] 6.2 Write `orcai-plugins/plugins/opencode/README.md` documenting custom Ollama model usage (`ollama/<name>` format) and referencing the ollama plugin README for model creation
