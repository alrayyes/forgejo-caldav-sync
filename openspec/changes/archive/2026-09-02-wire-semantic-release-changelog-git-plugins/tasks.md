Retroactive record of already-completed work (issue #15). Shipped on
`main` before this change was documented in OpenSpec.

## 1. Plugin wiring

- [x] 1.1 Add `.releaserc.json` listing `@semantic-release/changelog`
      and `@semantic-release/git` in `plugins`, in the right order —
      verified by the Release workflow's log showing
      `Loaded plugin ... from "@semantic-release/changelog"` and
      `... from "@semantic-release/git"`
- [x] 1.2 Confirm a real release generates `CHANGELOG.md` — verified
      by `CHANGELOG.md` existing at the repo root with its top entry
      matching the released version
- [x] 1.3 Confirm the release commit lands as expected — verified by
      `git log -1 --format=%s origin/main` showing the
      `chore(release): <version> [skip ci]` commit
