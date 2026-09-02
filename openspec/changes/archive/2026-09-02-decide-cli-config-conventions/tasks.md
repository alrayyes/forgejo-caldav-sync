## 1. Config struct

- [x] 1.1 Add `mapstructure` tags to `Config` and a `Validate() error`
      method, dropping the old `Load(getenv)` function — verified by
      `go test ./internal/config/...`
- [x] 1.2 Rewrite `internal/config`'s tests against `Validate()`
      directly — verified the same way

## 2. Cobra command tree

- [x] 2.1 Add `spf13/cobra`, `spf13/viper`, `adrg/xdg`,
      `golang.org/x/term` as pinned dependencies — verified by `go
build ./...`
- [x] 2.2 Build the root command with a `configFlag` table driving both
      flag registration and viper binding (flag, env, mapstructure key
      in one place) — verified by `forgejo-caldav-sync --help` listing
      every setting
- [x] 2.3 Implement `resolveConfig`: viper defaults, `BindPFlag`,
      `BindEnv`, config file read, `Unmarshal`, `Validate` — verified
      by `TestResolveConfig*` in `cmd/forgejo-caldav-sync/root_test.go`
      covering flag-beats-env, env-beats-file, file-beats-default, and
      the all-defaults case

## 3. init command

- [x] 3.1 Implement `init` writing a starter config file to the XDG
      config path, refusing to overwrite without `--force` — verified
      by `TestRunInit*` in `cmd/forgejo-caldav-sync/init_test.go`
- [x] 3.2 Implement the first-run prompt (`offerInit`), gated on
      `isTerminal(os.Stdin)` — verified manually: piped/no-TTY run
      skips the prompt and reports missing settings directly

## 4. healthcheck command

- [x] 4.1 Convert the `os.Args[1] == "healthcheck"` check into a
      hidden cobra subcommand reusing `resolveConfig` — verified by
      the existing `TestCheckHealth*` tests plus a real `docker build`
      and container run, confirming `docker inspect --format=
'{{.State.Health.Status}}'` reports `healthy`

## 5. Backward compatibility

- [x] 5.1 Confirm every existing `.env` variable name still works
      unchanged — verified manually: full env-var-only invocation
      (`FORGEJO_BASE_URL=... forgejo-caldav-sync`) resolves identically
      to before this change
- [x] 5.2 Confirm flag > env > file precedence end-to-end — verified
      manually with `--addr` overriding `ADDR` overriding a config
      file's `addr:`

## 6. Documentation

- [x] 6.1 README's Configuration section documents flags, env vars and
      the config file/`init` command together
- [x] 6.2 CONTRIBUTING.md's "How it fits together" section reflects
      the new `cmd/` file layout

## 7. Verification

- [x] 7.1 `go test -race ./...` green, `golangci-lint run ./...`
      clean, `govulncheck` clean (including a proactive bump of a
      transitive `golang.org/x/text` dependency past a
      not-actually-reachable advisory)
- [x] 7.2 `docker build .` succeeds and the built image runs correctly
      end-to-end (verified with a mock CalDAV server: starts, reports
      `/healthz` 200, shuts down cleanly on `SIGTERM`)
