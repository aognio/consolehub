# Coding & Development Conventions

## Deployment Boundary (Strict Directive)
- **NO DEPLOYMENTS**: Coding agents MUST NOT perform, automate, or execute any application deployment operations. Deployment management is handled exclusively out-of-band by a separate dedicated agent.

## Go Guidelines
- **Formatting**: Standard `gofmt` / `goimports`.
- **Package Names**: Short, lower-case, single word package names (`config`, `auth`, `models`, `services`, `storage`).
- **Error Handling**: Always return wrapped errors (`fmt.Errorf("failed to register run: %w", err)`). Never swallow errors silently.
- **Thin Handlers**: HTTP and WebSocket handlers unpack request objects, perform basic parameter validation, call domain services, and render output. Business logic lives in `internal/services`.

## Documentation & Comments
- All exported functions, structs, interfaces, and methods must have godoc comments.
- Markdown links in docs should reference exact files using GitHub/Markdown relative or file URI paths.

## Testing & TDD
- Write unit tests alongside implementation (`*_test.go`).
- Use table-driven test patterns.
- Integration tests must verify API endpoints and PocketBase repository layer.

## Frontend & Templating
- Server templates located in `internal/templates`.
- HTMX handles HTML fragment swaps (`hx-get`, `hx-post`, `hx-target`, `hx-swap`).
- Alpine.js (`x-data`, `x-show`, `x-on`) handles purely client-side UI interactions (e.g. modals, dropdowns, tab toggles).
