# BigSurBackport

`BigSurBackport` is the long-lived fork branch for keeping current `gh`
features usable on older macOS systems that are capped at Go 1.24.x.

## Policy

- `upstream/trunk` remains the source of truth for new `gh` features and API changes
- `origin/BigSurBackport` carries only the compatibility delta needed to keep the branch buildable on Go 1.24.11 `darwin/amd64`
- the compatibility delta should stay concentrated in `go.mod`, `go.sum`, and narrowly-scoped follow-up code changes when upstream adopts newer APIs

## Routine update flow

1. Fetch upstream changes

   ```sh
   git fetch upstream
   ```

2. Merge upstream into the backport branch

   ```sh
   git switch BigSurBackport
   git merge upstream/trunk
   ```

3. Reconcile the compatibility layer

   - keep `go 1.24.11`
   - keep the `BigSurBackport` `replace` block in `go.mod`
   - update pinned dependency versions only when upstream changes force it

4. Re-verify locally with the Big Sur toolchain

   ```sh
   GOTOOLCHAIN=local go mod tidy
   GOTOOLCHAIN=local go test ./...
   GOTOOLCHAIN=local make
   ```

## Linting

The current upstream lint toolchain has moved past what can be built locally on
Go 1.24.11. Use GitHub Actions on the fork for that signal, or run the lint
workflow from a newer host when needed.
