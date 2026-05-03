# Big Sur Backport

`BigSurBackport` is the long-lived fork branch for keeping current `gh`
features usable on older macOS systems that are capped at Go 1.24.x.

## Policy

- `upstream/trunk` remains the source of truth for new `gh` features and API changes
- `BigSurBackport` carries only the compatibility delta needed to keep the branch buildable on Go 1.24.11 `darwin/amd64`
- the compatibility delta should stay concentrated in `go.mod`, `go.sum`, and narrowly-scoped follow-up code changes when upstream adopts newer APIs
- machine-local coordination notes, worktree paths, session identifiers, and review artifacts must stay in ignored local files

## Routine update flow

1. Fetch upstream changes from a clean checkout

   ```sh
   git fetch upstream
   ```

2. Merge upstream into a separate reconciliation worktree

   ```sh
   git switch BigSurBackport
   git merge upstream/trunk
   ```

   Keep any installed or actively used `gh` checkout on the last proven
   backport release until the reconciliation worktree passes local validation.

3. Reconcile the compatibility layer

   - keep `go 1.24.11`
   - keep the `BigSurBackport` `replace` block in `go.mod`
   - update pinned dependency versions only when upstream changes force it
   - avoid installing runtime artifacts from the reconciliation worktree

4. Re-verify locally with the Big Sur toolchain

   ```sh
   GOTOOLCHAIN=local go mod tidy
   GOTOOLCHAIN=local go test ./...
   GOTOOLCHAIN=local make
   PATH="$(git rev-parse --git-path local-bin):$PATH" GOTOOLCHAIN=local make lint
   ```

5. Merge the validated result back to `BigSurBackport`

   Only after tests, lint, and local review gates pass, push the reconciled
   branch and update the fork branch or draft PR. Regenerate installed binaries,
   shell completions, and manpages from the validated checkout after the
   reconciliation result lands there.

## Linting

Use a repo-local ignored `golangci-lint` binary. The upstream v2.11.4
`darwin-amd64` release binary targets macOS 12.0+, so build the latest usable
v2 line with the local Go 1.24.11 toolchain:

```sh
lint_bin="$(git rev-parse --git-path local-bin)"
mkdir -p "$lint_bin"
GOTOOLCHAIN=local GOBIN="$lint_bin" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0
PATH="$lint_bin:$PATH" GOTOOLCHAIN=local make lint
```

## Validated install artifacts

Do reconciliation work in a separate worktree branch and treat the validated
`BigSurBackport` checkout as the only source for installed runtime artifacts.

- use the worktree branch for forward-porting and conflict resolution
- only update the installed `gh` binary, shell completions, and manpages after
  the worktree changes have been validated and reconciled back into
  `BigSurBackport`
- generate and install completion and manpage assets from the validated
  checkout, not directly from the worktree
