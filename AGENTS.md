# Repository Guidelines

## Project Structure & Module Organization

This is the Go module `github.com/a-aslani/wotop`. The root package defines the Clean Architecture contracts (`application.go`, `controller.go`, and `usecase.go`). Reusable adapters and helpers live in focused top-level packages such as `pubsub/`, `jwt/`, `mailer/`, `postgres_db/`, `upload_file/`, and `util/`. Shared domain contracts and payloads are under `model/`.

The CLI is in `cmd/`: `cmd/wotop/main.go` is the executable, while `cmd/templates/` contains the files it scaffolds. `examples/monolith_ddd_simple_app/` demonstrates an application layout and configuration. Keep generated output and example-specific code out of reusable library packages.

## Build, Test, and Development Commands

- `go test ./...` — run all package tests.
- `go test ./pubsub -run TestName` — run one focused test while iterating.
- `go vet ./...` — catch common Go correctness issues.
- `gofmt -w path/to/file.go` — format edited Go files before committing.
- `go build ./cmd/wotop` — compile the CLI.
- `go run ./cmd/wotop --help` — run the CLI locally and inspect commands.

Use Go 1.24 as declared in `go.mod`. Run `go mod tidy` only when dependency changes require it, and include resulting `go.mod` and `go.sum` updates.

## Coding Style & Naming Conventions

Follow standard Go formatting: tabs for indentation and `gofmt` as the source of truth. Use short, lowercase package names; exported identifiers use `PascalCase`, and unexported identifiers use `camelCase`. Name Go files with `snake_case.go`; use the same pattern for generated use-case directories (for example, `get_user_info`). Keep interfaces close to the consumer and make exported APIs purposeful and documented when their intent is not obvious.

## Testing Guidelines

Place tests beside the package they cover in `*_test.go` files. Use the standard `testing` package; Testify is available for assertions and requirements. Name tests `Test<Subject>_<Scenario>` (for example, `TestConsumer_RetriesFailedMessage`) and cover both successful behavior and meaningful error paths. Run `go test ./...` before opening a pull request. No repository-wide coverage threshold is currently configured.

## Commit & Pull Request Guidelines

Recent history uses Conventional Commit-style, imperative subjects such as `feat: add support for JSON file type` and `fix: handle upload failure`. Use `feat:`, `fix:`, `refactor:`, `test:`, or `docs:` as appropriate, keeping each commit focused.

Pull requests should explain the behavioral change, identify affected packages, link the relevant issue when one exists, and list tests run. Include CLI output or screenshots only when they clarify a user-visible change. Avoid unrelated formatting or dependency churn.
