## Why

A bug report or feature request filed through GitHub's UI had no
structure — no required fields, no prompt for acceptance criteria or
a definition of done. `rules/docs.md` requires YAML issue forms
encoding the four-part shape (description, acceptance criteria,
definition of done) as required fields, and a terse PR template.
Neither existed in this repo.

## What Changes

- `.github/ISSUE_TEMPLATE/bug_report.yml`: `description`,
  `reproduction`, `expected`, all required textareas.
- `.github/ISSUE_TEMPLATE/feature_request.yml`: `description`,
  `acceptance_criteria`, `definition_of_done`, all required textareas.
- `.github/PULL_REQUEST_TEMPLATE.md`: a summary section and a
  test-plan checklist, nothing else.

## Capabilities

### New Capabilities

(none — this shapes how contributors file issues and PRs, it doesn't
change what the service does)

### Modified Capabilities

(none)

This is tooling/docs: `skip_specs: true` is set in `.openspec.yaml`.

## Impact

`.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md`.
