# Topic-Chain

Topic-Chain is a Go implementation for classifying follow-up questions and routing them to conversation topic chains. The repository includes CLCBench and ELCBench demo datasets, optional model-service clients, and an optional Python LLM-as-judge workflow.

## Requirements

- Go 1.23.3 (the module declares `go 1.22.2` and `toolchain go1.23.3`)
- Python 3.10 or newer for the optional judge workflow
- An OpenAI-compatible chat endpoint for the interactive demo

ChatGLM, Dify, embeddings, and Qdrant are optional. They are only needed for the corresponding workflows or integration tests.

## Clone and enter the runnable project

The commands below assume that the repository has already been cloned:

```bash
cd topic-chain
```

Run all Go commands from this directory. The program uses relative paths such as `./bench/main.json` and `./record/`.

## Configure the API

The repository contains the safe template `config_demo.env`. It does not contain real credentials.

### Option A: local `config.env`

Copy the template and edit the new local file:

```bash
cp config_demo.env config.env
```

On Windows PowerShell:

```powershell
Copy-Item config_demo.env config.env
```

Set at least these OpenAI-compatible settings in `config.env`:

```dotenv
OPENAI_MODEL_NAME=gpt-4o-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=replace-with-your-api-key
```

`config.env` is ignored by Git. Never commit it or paste real credentials into `config_demo.env`, source comments, README files, or issue reports.

### Option B: environment variables

Bash:

```bash
export OPENAI_MODEL_NAME=gpt-4o-mini
export OPENAI_BASE_URL=https://api.openai.com/v1
export OPENAI_API_KEY=replace-with-your-api-key
```

Windows PowerShell:

```powershell
$env:OPENAI_MODEL_NAME = "gpt-4o-mini"
$env:OPENAI_BASE_URL = "https://api.openai.com/v1"
$env:OPENAI_API_KEY = "replace-with-your-api-key"
```

Environment variables are suitable for CI and temporary local tests. The Go program loads `config.env` when it exists, while the process environment remains the source of truth.

## Build and offline tests

Download the pinned modules, run the test suite, and build every package:

```bash
go mod download
go test ./...
go build ./...
```

The default Go tests are safe to run without live model services. Tests that require ChatGLM, Dify, an OpenAI-compatible embedding endpoint, or Qdrant are skipped unless explicitly enabled:

```bash
RUN_INTEGRATION_TESTS=1 go test ./...
```

Only set `RUN_INTEGRATION_TESTS=1` when the required services and credentials are available.

## Start the interactive demo

After configuring the OpenAI-compatible endpoint, start the program from the repository root:

```bash
go run .
```

Windows PowerShell uses the same command:

```powershell
go run .
```

Startup behavior:

1. `main.go` loads case `0` from `bench/main.json`.
2. The historical user/assistant messages are sent to the configured LLM to initialize the topic chains. This happens before the first interactive prompt, so startup may take time and consume API quota.
3. The program prints the generated topic chains and prompts for a new question.
4. Type a question and press Enter to route it through the initialized conversation chains.
5. Type `cmd-quit` and press Enter to exit and save the session context.

The program writes the generated session context to `llama/records_<case>_save_<timestamp>.json`. Token usage is written to `record/usage_<timestamp>.txt`; usage files are ignored by Git.

## Run the CLCBench demo file

The public CLCBench fixture is:

```text
bench/talking_records_demo_CLCBench.json
```

The current `main.go` intentionally reads the stable runtime path `bench/main.json`. To test the named CLCBench fixture, temporarily use it as `main.json`, run the program, and restore the original file afterward. For example, in PowerShell:

```powershell
Copy-Item bench/main.json bench/main.json.bak
Copy-Item bench/talking_records_demo_CLCBench.json bench/main.json -Force
go run .
Move-Item bench/main.json.bak bench/main.json -Force
```

The same workflow in Bash is:

```bash
cp bench/main.json bench/main.json.bak
cp bench/talking_records_demo_CLCBench.json bench/main.json
go run .
mv bench/main.json.bak bench/main.json
```

If the process is interrupted, restore `bench/main.json` manually before the next run.

## Optional LLM-as-judge workflow

Install the Python dependency and configure the same OpenAI-compatible variables:

```bash
python -m pip install -r requirements.txt
cd bench/llm_as_judge
python llm_judge.py
```

The judge reads an answer file from `llm_ans/` and writes generated results under `judge_res/`. Generated judge results are ignored by Git.

## Repository layout

- `main.go`: interactive topic-chain demo entry point
- `config/`: environment loading
- `utils/`, `meta/`, `prompt/`, `types/`: core topic-chain implementation
- `record/`: dataset loading and runtime usage records
- `chatglm_sdk/`, `dify_sdk/`: optional model-service clients
- `bench/`: demo datasets and optional evaluation scripts
- `config_demo.env`: safe configuration template with placeholders only

## Public-repository checklist

Before pushing to GitHub:

- Keep `config.env` local and verify it is ignored by Git.
- Use environment variables or a local secret manager in CI.
- Do not place API keys, JWTs, private endpoints, or credential-bearing comments in source files.
- Run `go test ./...` and `go build ./...` from `topic-chain`.
- Run the interactive demo only with a temporary credential and rotate credentials that were ever exposed in chat, logs, or shell history.
