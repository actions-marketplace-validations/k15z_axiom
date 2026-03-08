# Getting Started

## Prerequisites

- Go 1.25 or later
- An API key from [Anthropic](https://console.anthropic.com/) (default), [OpenAI](https://platform.openai.com/api-keys), or [Google Gemini](https://aistudio.google.com/apikey)

## Install

```bash
go install github.com/k15z/axiom/cmd/axiom@latest
```

## Set Your API Key

Create a `.env` file in your project root:

```bash
echo "ANTHROPIC_API_KEY=sk-ant-..." > .env
```

Or export it directly:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Axiom loads `.env` automatically. Existing environment variables take precedence.

## Your First Test

From your project root:

```bash
axiom add "all API routes require authentication"
```

Axiom's agent explores your codebase, finds the relevant files, and generates a test:

```yaml
test_api_routes_require_authentication:
  on:
    - src/routes/**/*.py
  condition: >
    All route handlers that access user data must require authentication.
    Public endpoints (health checks, login, registration) are exempt.
```

It shows you the generated YAML and asks for confirmation before writing it to disk. If multiple test files exist in `.axiom/`, it prompts you to pick one. After writing, it offers to run the test immediately.

To skip the prompts and run right away:

```bash
axiom add "all API routes require authentication" --run
```

A typical test costs $0.01--0.05 with Haiku, more with Sonnet or Opus. Use `axiom run --dry-run` to estimate costs before running.

## Generate Tests Automatically

To generate a batch of tests based on your codebase:

```bash
axiom init
```

This creates `.axiom/tests.yml` with generated test definitions and `axiom.yml` with default configuration.

## Validate Your Tests

Before running tests, check them for problems:

```bash
axiom validate
```

This catches invalid glob syntax, missing `on` patterns (tests that can never be cached), and vague conditions -- before you spend API calls.

## Run Tests

```bash
axiom run
```

Axiom discovers all YAML files in `.axiom/`, checks the cache, and runs any tests whose trigger files have changed. On the first run, all tests execute.

```
  axiom

  .axiom/tests.yml
    ✓ test_no_sql_injection (8.4s)
      All database queries use parameterized statements via the ORM
    ✗ test_auth_middleware (9.2s)
      Route GET /admin/users bypasses auth -- accesses request.user without verify_token()
    ○ test_rate_limiting (cached)

  2 passed · 1 failed · 1 cached
```

Run a single test:

```bash
axiom run test_auth_middleware
```

Ignore the cache and run everything:

```bash
axiom run --all
```

Preview what would run and the estimated cost without calling the API:

```bash
axiom run --dry-run
```

## Write Tests by Hand

Create `.axiom/my_tests.yml`:

```yaml
test_readme_exists:
  condition: >
    The project must have a README.md file in the root directory
    that includes installation instructions and a usage section.
```

See [Writing Tests](./writing-tests.md) for the full YAML format and tips on writing effective conditions.

## Watch Mode

Re-run affected tests automatically when source files change:

```bash
axiom run --watch
```

Axiom watches files matching your tests' `on` globs and re-runs only the affected tests when changes are detected. Press Ctrl+C to stop.

## Next Steps

- [Writing Tests](./writing-tests.md) -- learn the YAML format and how to write effective conditions
- [Examples](./examples.md) -- curated test examples across security, architecture, and code quality
- [Configuration](./configuration.md) -- customize model, providers, timeouts, and caching
- [CI Integration](./ci-integration.md) -- set up axiom in GitHub Actions
