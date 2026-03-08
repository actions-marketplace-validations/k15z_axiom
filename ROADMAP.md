# Roadmap

Planned improvements for axiom, roughly in priority order. Revised 2026-03-08 by the full team.

## Done

Shipped features — kept here for reference.

- **Go unit tests for core packages** — table-driven tests for safePath, glob.Match, cache.ShouldSkip, discovery ordering, verdict parsing
- **Per-tool timeouts** — individual timeout for each agent tool call so one slow operation doesn't eat the entire test budget
- **Flaky test detection & retries** — `--retries N` re-runs failures; if a retry passes, marks the test as flaky
- **Cache invalidation on config change** — changing model, max_iterations, or max_tokens invalidates cached results
- **Duplicate test name detection** — warn or error during discovery if two tests share the same name
- **Watch mode** (`axiom run --watch`) — re-run affected tests when source files change via fsnotify
- **`axiom validate`** — lint test YAML for invalid globs, missing `on` patterns, vague conditions
- **Progress indicator** — "3/10 tests complete" during runs; periodic lines in CI/non-TTY
- **Dry-run mode** (`axiom run --dry-run`) — preview which tests would run, cached status, estimated token cost
- **Per-test config overrides** — `timeout`, `model`, and `max_iterations` in test YAML
- **Test tags/filtering** — `tags: [security, auth]` in test YAML, `axiom run --tag security`
- **Token budget hints** — system message injected when agent approaches token limit
- **Exit code semantics** — 0 = all pass, 1 = failures, 2 = config/setup error
- **PR comment summaries** — `axiom run --format github` outputs GitHub-flavored Markdown for PR comments
- **Reusable GitHub Action** — `uses: k15z/axiom@main` for one-step CI with cache and PR comments
- **Multi-provider support** — Anthropic, OpenAI, Gemini, and OpenAI-compatible endpoints via `provider` and `base_url`
- **Documentation site** — mdBook-based docs with GitHub Pages deployment
- **Enhanced examples** — 11 example tests across architecture, security, error handling, and code quality
- **Agent memory / codebase notes** — agent persists investigation notes across runs with file-hash staleness tracking
- **Agent reasoning diff** (`axiom show --diff`) — compare current cached reasoning against the previous run's reasoning with +/- line markers
- **Infrastructure error distinction** — `Errored` tests (API failures, timeouts) show `!` icon, exit code 2, skip retries and cache updates
- **Verdict parsing tightening** — regex-based with whitespace tolerance and last-match-wins semantics
- **`axiom version`** — version command with ldflags support
- **`--strict` mode** — `axiom run --strict` treats flaky tests (passed on retry) as failures
- **`axiom add` UX improvements** — file selection prompt when multiple YAML files exist, `--run` flag to run the test immediately after adding, YAML validation before writing
- **`axiom init` multi-provider** — respects configured provider instead of hardcoding Anthropic
- **Shared retry logic with jitter** — extracted retry utility across all providers with exponential backoff and jitter to prevent thundering herd
- **Runner-level rate-limit backoff** — automatic exponential backoff with jitter when 429 errors are detected, with cooldown auto-reset
- **Expanded unit test coverage** — 67+ new tests across agent (14 mock-provider tests), runner (22 integration tests), config (edge cases), and provider (Gemini, retry) packages
- **SetupError wrapping** — `add`, `init`, `validate` commands wrap setup errors for correct exit code 2
- **CI improvements** — behavioral tests non-blocking (`continue-on-error`), rate-limit resilience
- **`--quiet` / `-q` flag** — suppress full terminal output, emit only the CI summary line to stderr
- **`axiom list --json`** — JSON output for the list command, consistent with `run` and `show`
- **`axiom cache info`** — show cache entry count, total size, oldest entry, and per-test cache status
- **Cache corruption warning** — `cache.Load()` warns on stderr when JSON is corrupt instead of failing silently
- **`cache clear` respects config** — loads `axiom.yml` so custom `cache.dir` is respected
- **`init` and `add` respect `test_dir` config** — both load config and use `cfg.TestDir`
- **Cache tracks provider and base_url changes** — `HashConfig` includes provider and base_url so switching providers invalidates cached results
- **Context-aware tools (goroutine leak fix)** — threaded `context.Context` into `toolGrep` and directory walk functions so tool timeouts cancel work instead of leaking goroutines
- **Retry loop context cancellation fix** — `break` inside `select` only broke the select, not the `for` loop; fixed to check `ctx.Err()`
- **stripGooglePrefix Gemini bug fix** — dead code path caused 404 errors on certain Gemini model name formats
- **math/rand/v2 standardization** — replaced deprecated `math/rand` with `math/rand/v2` in throttle.go
- **Exit code fix for list/show commands** — setup errors in `list` and `show` now correctly exit with code 2
- **Hardcoded `.axiom/` path fix in output** — output now shows the actual configured `test_dir` instead of hardcoded `.axiom/`
- **Quiet mode flaky count** — CI summary line now includes flaky test count
- **Double space fix in dry-run header** — removed extra space in dry-run output formatting
- **Binary releases via goreleaser** — pre-built binaries for macOS (arm64/amd64), Linux (arm64/amd64), and Windows, published to GitHub Releases
- **`--dir` flag for cache commands** — `cache clear` and `cache info` accept `--dir` for consistency with other commands
- **Getting-started restructure** — leads with `axiom add` as the entry point instead of `axiom init`
- **README Quick Start update** — `axiom add` is now the first command shown
- **Docs cleanup** — removed AI-isms, deduplicated agent memory explanations across pages, added `axiom validate` to CI tips
- **CI docs PR comment example** — already correct (captures output and exit code in one run)

## Next Up

High-priority items that directly impact adoption and reliability.

### Distribution & Releases

- **Homebrew formula** — `brew install axiom` for macOS/Linux users.
- **Curl installer** — `curl -fsSL https://axiom.dev/install.sh | sh` for quick setup.
- **Semver tagged releases** — proper `v1.x.x` tags with changelogs. The GitHub Action should reference version tags (`@v1`) instead of `@main`.

## Medium-Term

Important improvements that expand axiom's reach and improve the developer experience. Target: 1-3 months.

### Documentation & Adoption

- **Document streaming behavior per provider** — note that Anthropic shows live streaming progress while OpenAI and Gemini show results after each test completes. Manages expectations, zero code change.

### Architecture & Tech Debt

- **Extract TestExecutor from runner** — the 280-line `Run()` function handles cache checks, provider creation, progress routing, notes, retries, and results all in one closure. Extract per-test execution into a testable struct to make retry logic, timeouts, and provider reuse cleaner.
- **Consolidate config loading** — `config.go` has 5 loading functions (`Load`, `LoadWithoutKey`, `LoadMinimal`, `LoadForDryRun`, `LoadAPIKey`) with duplicated YAML unmarshal + defaults logic. Consolidate into `Load(opts LoadOpts)`.
- **Reuse provider across tests** — currently creates a new provider (and SDK client) per test to route streaming progress. Should share the client and route progress differently.
- **Structured error types** — `isRateLimitErr` in runner uses string matching ("429", "rate limit") while the provider layer has proper `rateLimitError` types. Align on typed errors throughout.

### Test Lifecycle

- **Auto-validate before running** — run fast validation checks (glob syntax, missing conditions, duplicate names) before starting expensive agent calls. Saves API costs on malformed tests.
- **`axiom show` displays model used** — when showing cached reasoning, include which model produced it. Important with per-test model overrides.
- **Discovery line numbers in errors** — when YAML parsing fails, surface the line number (available from `yaml.Node`) instead of just the file path.

## Future Vision

Larger features that expand what axiom can do. These are ideas, not commitments.

- **Custom tool plugins** — let `axiom.yml` define additional tools the agent can use (e.g. run a linter, query a database schema, call an API endpoint)
- **Test dependencies** — allow tests to declare `depends_on: [other_test]` so they only run after prerequisites pass
- **Snapshot testing** — save and diff agent reasoning across runs to detect regressions in test behavior
- **Condition quality scoring** — use a fast model to rate test conditions on specificity, measurability, and relevance before running them
- **Verbose tool tracing** (`axiom run --trace`) — log every tool call, its arguments, output size, and duration to a file for post-mortem debugging
- **Expected failures (`expect_fail`)** — mark tests as expected to fail (like `@pytest.mark.xfail`). Revisit if users request it; current mental model is "conditions should always hold."
- **Cost dashboard** — track cumulative API spend over time, per-test cost trends, and cost-per-run history
- **Config inheritance** — shared base config for monorepos or teams, with per-directory overrides
- **OpenAI/Gemini streaming** — SSE streaming for live progress updates across all providers. Currently Anthropic-only. Significant effort (SSE parsing, error handling, testing) with limited adoption impact since most users start with Anthropic.
