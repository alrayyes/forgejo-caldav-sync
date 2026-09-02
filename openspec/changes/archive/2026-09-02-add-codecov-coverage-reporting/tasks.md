Retroactive record of already-completed work (issue #6). Shipped on
`main` before this change was documented in OpenSpec.

## 1. Coverage reporting

- [x] 1.1 Add a Codecov upload step to CI's `test` job, gated on
      `fail_ci_if_error: true` — verified by a green `test` job on the
      originating pull request with a Codecov check present
- [x] 1.2 Set the `CODECOV_TOKEN` repo secret — verified by the upload
      step succeeding rather than failing on a missing token
- [x] 1.3 Add a Codecov badge to README.md next to the existing CI
      badge — verified by the badge rendering on the repo's README
