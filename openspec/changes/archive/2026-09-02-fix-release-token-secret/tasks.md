Retroactive record of already-completed work (issue #8). Shipped
before this change was documented in OpenSpec.

## 1. Release token

- [x] 1.1 Mint a fine-grained PAT scoped to this repo (Contents:
      read/write) — verified by the token existing in the account's
      developer settings
- [x] 1.2 Set it as the `RELEASE_TOKEN` repo secret via
      `gh secret set RELEASE_TOKEN` — verified by
      `gh secret list` showing it present
- [x] 1.3 Confirm a real Release workflow run completes green —
      verified by `gh run list --branch main --json conclusion` showing
      the most recent `Release` run as `success`
