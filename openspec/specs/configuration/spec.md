# configuration Specification

## Purpose

Resolves this tool's settings from flags, environment variables and a
config file, in that precedence order, and helps a first-time user get
a config file started instead of hand-authoring one from documentation
alone.

## Requirements

### Requirement: Layered configuration precedence

The system SHALL resolve every setting from, in order of precedence:
an explicit flag, an environment variable, a config file, then a
built-in default.

#### Scenario: A flag overrides everything else

- **WHEN** a setting is passed as a flag, and also set via environment
  variable and config file
- **THEN** the flag's value is what's used

#### Scenario: An environment variable overrides the config file

- **WHEN** a setting has no flag passed, but is set both via
  environment variable and config file
- **THEN** the environment variable's value is what's used

#### Scenario: A config file value overrides the built-in default

- **WHEN** a setting has no flag or environment variable set, but is
  present in the config file
- **THEN** the config file's value is what's used

### Requirement: Starter config file

The system SHALL provide an `init` command that writes a starter
config file populated with the defaults it would otherwise fall back
to, at the XDG config directory unless a different path is given.

#### Scenario: Writing a starter config

- **WHEN** `init` is run and no config file exists yet at the target
  path
- **THEN** a config file is written there, and the command reports the
  path it wrote

#### Scenario: Refusing to overwrite

- **WHEN** `init` is run and a config file already exists at the
  target path, without `--force`
- **THEN** the command fails without touching the existing file

### Requirement: First-run prompt gated on a real terminal

The system SHALL offer to write a starter config file when it finds
no config file and no required environment variable set, but only
when standard input is an interactive terminal.

#### Scenario: Interactive run with nothing configured

- **WHEN** the process starts with no config file, no required setting
  in the environment, and stdin is a terminal
- **THEN** it prompts to write a starter config file before reporting
  the missing settings

#### Scenario: Non-interactive run with nothing configured

- **WHEN** the process starts with no config file, no required setting
  in the environment, and stdin is not a terminal (CI, a container, a
  piped invocation)
- **THEN** it reports the missing settings and exits, without
  prompting or blocking on input
