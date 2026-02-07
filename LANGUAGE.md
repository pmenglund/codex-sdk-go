# Language Guidelines – Go

This project is primarily written in **Go**. All code contributions should adhere to Go’s idioms and best practices, as well as project-specific conventions outlined here.

## Formatting and Style

- **gofmt and linting:** Always format code with `gofmt` (or `go fmt`) before committing. The CI will reject code not properly formatted. Also run `golint` or `golangci-lint` as configured.
- **Idiomatic Go:** Follow guidance from *Effective Go* – e.g., use camelCase for variable and function names, CapitalizedNames for exported symbols, and avoid overly long or obscure names.
- **Error handling:** Check and handle errors from every call that can fail. Use Go’s multi-value returns for errors. Do not swallow errors; if ignoring is absolutely necessary, add a comment explaining why.
- **Logging:** Use the project’s logging library (if any) or the standard library `log` responsibly. No excessive logging in hot code paths. Remove or lower the level of debug logs before final commit unless they are permanent.
- **Comments:** For any exported function or public package API, include a clear comment (will be used by GoDoc). Keep comments up-to-date if code changes. Internally, comment complex logic or non-obvious decisions.

## Project Structure and Conventions

- **Modules and Packages:** Organize code into packages by functionality. Avoid circular dependencies. Place code in the appropriate folder as per the repository layout (see APP.md for architecture).
- **Naming:** Package names should be short and lowercase (no underscores or caps). Use clear names for packages and files (e.g., `httphandler.go` for HTTP handlers).
- **File structure:** One type per file if possible. Keep file lengths reasonable; if a file grows too large (> ~300–400 lines), consider refactoring.
- **Third-party libraries:** Prefer standard library solutions when available (e.g., `net/http` for HTTP, `database/sql` for DB access). If a third-party package is needed, ensure it’s well-maintained. Stick to the libraries already in use in the project unless there’s a strong reason to introduce a new one (and discuss in an ExecPlan or issue).

## Testing

- **Test framework:** Use Go’s built-in `testing` package for unit tests. Place test files alongside code (`foo_test.go` for `foo.go`).
- **Test coverage:** Every new feature or bug fix should come with adequate tests. Aim to cover core logic and edge cases. We strive for high coverage, but more importantly, meaningful tests.
- **Test style:** Follow table-driven test patterns for readability when testing multiple scenarios. Use descriptive test case names. Keep tests deterministic; avoid external network calls or time-dependent logic in tests (use mocking or interfaces for external interactions).
- **Running tests:** Use `go test ./...` to run all tests. Ensure tests pass locally before committing. If adding a new package, add it to CI scripts if needed so it’s tested.
- **Benchmarking (if applicable):** If performance is a concern in a piece of code, include benchmarks (in `_test.go` files) and note any performance expectations.

## Common Patterns and Anti-Patterns

- **Concurrency:** Use goroutines and channels appropriately. Avoid data races – use mutexes or channel synchronization for shared data. If using concurrency, also provide tests that cover concurrent scenarios if possible (e.g. use `-race` detector).
- **Resource management:** Use `context.Context` for cancellation/timeouts on IO or long-running operations. Respect context cancellation. Close any opened resources (files, DB connections) promptly, usually via `defer`.
- **Error messages:** When returning errors, provide context. E.g., `fmt.Errorf("failed to load config: %w", err)`. This aids debugging while preserving original error via wrapping.
- **Panics:** Do not use panics for normal error handling. Panics are only acceptable for truly unrecoverable situations (and even then, prefer returning errors). If a panic or runtime exception is encountered (like from a library), consider recovering at boundary of a goroutine to log and continue if it’s safe.
- **Performance:** Write clear code first; optimize only if profiling shows a need. However, do consider the complexity of algorithms (don’t accidentally use O(n^2) where n could be large, etc.). Document any performance-related decisions in code comments or planning docs.

## Go Tools and Workflow

- **Building:** Use the provided Makefile or commands from your selected workflow file in `workflows/` (for example, `workflows/LINEAR.md`) to build the project (`go build ...`). Ensure no build warnings.
- **Dependency management:** We use Go modules. Import statements should be versioned via `go.mod`. Run `go mod tidy` after adding new imports to keep the module file tidy.
- **Debugging:** It’s fine to use `println` or `log.Printf` for local debugging, but remove those before committing. Instead, write tests or use the debugger for ongoing verification.
- **CLI and scripts:** Use the standard Go CLI flags and `cobra` (if this project uses Cobra/Viper for CLI) according to project norms. Document any new command-line flags you add in the README or help text.

## Example and References

- **Examples in Repo:** Look at existing code for style guidance. E.g., follow patterns from `pkg/handlers/example.go` or similar modules to see how things are implemented in this project.
- **Reference docs:** Refer to [Effective Go](https://go.dev/doc/effective_go) and the Go standard library documentation for best practices. Adhere to any additional style guidelines this team has communicated (if in doubt, ask or check prior PRs).

By following these language-specific guidelines, we maintain a consistent and idiomatic codebase in Go. When the AI agent (or any contributor) writes code respecting these conventions, the result is more readable, maintainable, and blends seamlessly with the existing code.
