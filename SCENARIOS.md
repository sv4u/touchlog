# touchlog Gherkin Scenario Suite

This document is organized as **virtual feature files** that exactly match the intended on-disk layout under `features/`.
Each section corresponds to a single `.feature` file and contains **exactly one** `Feature:` block (single responsibility).

Technical layering order:

1. `features/platform.feature`
2. `features/cli.feature`
3. `features/config.feature`
4. `features/templates.feature`
5. `features/new.feature`
6. `features/editor.feature`
7. `features/repl_wizard.feature`
8. `features/repl_ui.feature`
9. `features/metadata.feature`
10. `features/errors.feature`
11. `features/list_search.feature` (future)

---

## features/platform.feature

```gherkin
Feature: Platform support (macOS / Linux / WSL)

  Scenario: Runs on macOS
    Given the operating system is "darwin"
    When I run "touchlog --version"
    Then the exit code is 0

  Scenario: Runs on Linux
    Given the operating system is "linux"
    When I run "touchlog --version"
    Then the exit code is 0

  Scenario: Runs on WSL
    Given the operating system is "linux"
    And the environment indicates WSL
    When I run "touchlog --version"
    Then the exit code is 0
    And stdout contains "WSL"

  Scenario: Refuses to run on unsupported Windows native
    Given the operating system is "windows"
    When I run "touchlog --version"
    Then the exit code is not 0
    And stderr contains "unsupported platform"
```

---

## features/cli.feature

```gherkin
Feature: touchlog CLI entrypoint and subcommand routing

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"
    And the local timezone is "America/Denver"

  Scenario: Help flag prints help for root command
    When I run "touchlog --help"
    Then the exit code is 0
    And stdout contains "Usage:"
    And stdout contains "--help"
    And stderr is empty

  Scenario: Version flag prints semantic version and exits successfully
    When I run "touchlog --version"
    Then the exit code is 0
    And stdout matches the regex "touchlog v\d+\.\d+\.\d+(-[0-9A-Za-z\.-]+)?"

  Scenario: Unknown flag produces a clear error and non-zero exit
    When I run "touchlog --does-not-exist"
    Then the exit code is not 0
    And stderr contains "unknown flag"
    And stdout contains "Usage:"

  Scenario: Unknown subcommand returns an error and prints help
    When I run "touchlog frobnicate"
    Then the exit code is not 0
    And stderr contains "unknown command"
    And stdout contains "Usage:"
```

---

## features/config.feature

```gherkin
Feature: Configuration loading, validation, and precedence

  Background:
    Given a clean temporary working directory

  Scenario: Load YAML config successfully
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      template: "default"
      '''
    When I run "touchlog config validate --config touchlog.yaml"
    Then the exit code is 0
    And stdout contains "valid"

  Scenario: Load TOML config successfully
    Given a config file at "touchlog.toml" with:
      '''
      output_dir = "./logs"
      template = "default"
      '''
    When I run "touchlog config validate --config touchlog.toml"
    Then the exit code is 0
    And stdout contains "valid"

  Scenario: Reject unknown config keys (strict mode)
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      unknown_key: "surprise"
      '''
    When I run "touchlog config validate --config touchlog.yaml --strict"
    Then the exit code is not 0
    And stderr contains "unknown_key"

  Scenario: CLI flags override config values
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      '''
    When I run "touchlog new --config touchlog.yaml --output ./override --message 'x'"
    Then a file exists under "./override"
```

---

## features/templates.feature

```gherkin
Feature: Log content templating

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"
    And the local timezone is "America/Denver"

  Scenario: Apply a named template from config
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      template:
        name: "daily"
      templates:
        daily: |
          # {{date}}
          ## Title
          {{title}}
          ## Notes
          {{message}}
      '''
    When I run "touchlog new --config touchlog.yaml --title 'Standup' --message 'Did A, B, C'"
    Then the exit code is 0
    And the log file content contains "# 2025-12-31"
    And the log file content contains "## Title"
    And the log file content contains "Standup"
    And the log file content contains "## Notes"
    And the log file content contains "Did A, B, C"

  Scenario: Template can be overridden by CLI flag
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      template:
        name: "daily"
      templates:
        daily: "DAILY {{message}}"
        minimal: "MIN {{message}}"
      '''
    When I run "touchlog new --config touchlog.yaml --template minimal --message 'x'"
    Then the log file content contains "MIN x"
    And the log file content does not contain "DAILY"

  Scenario: Missing template name fails clearly
    Given a config file at "touchlog.yaml" with:
      '''
      template:
        name: "missing"
      templates:
        daily: "..."
      '''
    When I run "touchlog new --config touchlog.yaml --message 'x'"
    Then the exit code is not 0
    And stderr contains "template 'missing' not found"
    And no log files exist
```

---

## features/new.feature

```gherkin
Feature: Creating log entries non-interactively (touchlog new)

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"
    And the local timezone is "America/Denver"

  Scenario: Create a basic entry with message
    When I run "touchlog new --message 'Did code review for PR-123'"
    Then the exit code is 0
    And stdout contains "Wrote log"
    And exactly 1 log file exists in the output directory
    And the log file content contains "Did code review for PR-123"
    And the log file content contains "2025-12-31"
    And stderr is empty

  Scenario: Create an entry with a title and message
    When I run "touchlog new --title 'Code Review' --message 'Reviewed PR-123'"
    Then the exit code is 0
    And the log file content contains "Code Review"
    And the log file content contains "Reviewed PR-123"

  Scenario: Create an entry with tags
    When I run "touchlog new --message 'Worked on touchlog' --tag 'work' --tag 'touchlog'"
    Then the exit code is 0
    And the log file content contains "work"
    And the log file content contains "touchlog"

  Scenario: Reject empty message when message is required
    When I run "touchlog new --message ''"
    Then the exit code is not 0
    And stderr contains "message must not be empty"
    And no log files exist

  Scenario: Support reading message from stdin
    Given stdin is:
      '''
      Investigated flaky test in CI.
      Root cause: timing issue.
      '''
    When I run "touchlog new --stdin"
    Then the exit code is 0
    And the log file content contains "Investigated flaky test in CI."
    And the log file content contains "Root cause: timing issue."

  Scenario: Use configured output directory by default
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      '''
    When I run "touchlog new --config touchlog.yaml --message 'hello'"
    Then the exit code is 0
    And a file exists under "./logs"

  Scenario: CLI flag overrides configured output directory
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: "./logs"
      '''
    When I run "touchlog new --config touchlog.yaml --output ./override --message 'hello'"
    Then the exit code is 0
    And a file exists under "./override"
    And no file exists under "./logs"

  Scenario: Generated filename is stable and includes date and slug
    When I run "touchlog new --message 'Investigate SCC finding'"
    Then the exit code is 0
    And the created filename matches the regex "2025-12-31_.*\.md"

  Scenario: Avoid overwriting by adding numeric suffix
    Given I run "touchlog new --message 'Same timestamp collision'"
    And I run "touchlog new --message 'Same timestamp collision'"
    Then exactly 2 log files exist in the output directory
    And the filenames are distinct
    And one filename ends with "_1.md" or "-1.md"

  Scenario: Overwrite explicitly when requested
    Given a pre-existing log file at the expected generated path with content "old"
    When I run "touchlog new --message 'new' --overwrite"
    Then the exit code is 0
    And the log file content contains "new"
    And the log file content does not contain "old"
```

---

## features/editor.feature

```gherkin
Feature: Editor integration and handoff semantics

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"

  Scenario: --editor flag overrides EDITOR environment variable
    Given the environment variable "EDITOR" is set to "vim"
    When I run "touchlog new --output ./logs --message 'x' --edit --editor nvim"
    Then a process is launched with command "nvim"

  Scenario: Configured editor overrides EDITOR environment variable
    Given the environment variable "EDITOR" is set to "vim"
    And a config file at "touchlog.yaml" with:
      '''
      editor: "nvim"
      '''
    When I run "touchlog new --config touchlog.yaml --output ./logs --message 'x' --edit"
    Then a process is launched with command "nvim"

  Scenario: EDITOR environment variable is used when no explicit editor is set
    Given the environment variable "EDITOR" is set to "vim"
    When I run "touchlog new --output ./logs --message 'x' --edit"
    Then a process is launched with command "vim"

  Scenario: If EDITOR is unset, touchlog falls back to a safe default
    Given the environment variable "EDITOR" is not set
    And the executable "vi" exists on PATH
    When I run "touchlog new --output ./logs --message 'x' --edit"
    Then the exit code is 0
    And a process is launched with command "vi" and argument equal to the created log file path

  Scenario: If no editor can be resolved, touchlog writes the file but warns
    Given the environment variable "EDITOR" is not set
    And the executable "vi" does not exist on PATH
    And the executable "nano" does not exist on PATH
    When I run "touchlog new --output ./logs --message 'x' --edit"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"
    And stderr contains "no editor found; skipping edit"

  Scenario: Non-interactive new creates final file and exits successfully
    When I run "touchlog new --output ./logs --title 'X' --message 'Hello'"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"
    And stdout contains "Created"
    And stdout contains the created log file path

  Scenario: With --edit, touchlog launches editor and then exits without further lifecycle management
    Given the environment variable "EDITOR" is set to "vim"
    When I run "touchlog new --output ./logs --message 'Hello' --edit"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"
    And a process is launched with command "vim" and argument equal to the created log file path
    And touchlog performs no deletion of the created file after launching the editor

  Scenario: Editor launch fails and the created file still remains
    Given the environment variable "EDITOR" is set to "nonexistent-editor"
    When I run "touchlog new --output ./logs --message 'Hello' --edit"
    Then the exit code is not 0
    And stderr contains "failed to launch editor"
    And exactly 1 log file exists in "./logs"
    And touchlog does not delete the created file
```

---

## features/repl_wizard.feature

```gherkin
Feature: Interactive REPL wizard control flow and lifecycle

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"
    And the local timezone is "America/Denver"

  Scenario: Running with no arguments launches the REPL wizard
    When I run "touchlog"
    Then the exit code is 0
    And stdout contains "touchlog interactive"
    And stdout contains "Select an action"
    And stdout contains "Create new entry"
    And stdout contains "Quit"

  Scenario: Back is available on all steps prior to file creation
    Given I start "touchlog"
    When I input "1"  # Create new entry
    Then stdout contains "Select output directory"
    And stdout contains "Options:"
    And stdout contains "Back"

    When I input "Use default"
    Then stdout contains "Enter title (optional)"
    And stdout contains "Options:"
    And stdout contains "Back"

    When I input "Title"
    Then stdout contains "Enter tags (comma-separated, optional)"
    And stdout contains "Options:"
    And stdout contains "Back"

    When I input "work"
    Then stdout contains "Enter message (optional)"
    And stdout contains "Options:"
    And stdout contains "Back"

  Scenario: After file creation, Back is not offered and is rejected if typed
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    And stdout contains "Options:"
    And stdout does not contain "Back"
    When I input "Back"
    Then stdout contains "Unknown selection" or stdout contains "Invalid option"
    And stdout contains "Options:"
    And exactly 1 log file exists in "./logs"

  Scenario: Wizard creates the final filename before offering editor
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Daily Notes"
    And I input "work,touchlog"
    And I input "Draft message"
    Then exactly 1 log file exists in "./logs"
    And the created filename matches the regex "2025-12-31_.*\.md"
    And stdout contains "Created"
    And stdout contains the created log file path
    And stdout contains "Open editor to edit the entry now?"

  Scenario: Wizard offers editor option even when message is provided inline
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Daily Notes"
    And I input "work,touchlog"
    And I input "Initial message typed inline"
    Then stdout contains "Open editor to edit the entry now?"
    When I input "1"  # Yes
    Then a process is launched with the resolved editor and argument equal to the created log file path
    Then stdout contains "Review summary"
    When I input "confirm"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"

  Scenario: Wizard allows skipping the editor after providing an inline message
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Daily Notes"
    And I input "work,touchlog"
    And I input "Inline message"
    Then stdout contains "Open editor to edit the entry now?"
    When I input "2"  # No
    Then stdout contains "Review summary"
    When I input "confirm"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"
    And the log file content contains "Inline message"

  Scenario: Wizard can open editor multiple times before confirmation
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Daily Notes"
    And I input "work,touchlog"
    And I input "Draft"
    Then stdout contains "Open editor to edit the entry now?"
    When I input "1"  # Yes
    Then a process is launched with the resolved editor and argument equal to the created log file path
    Then stdout contains "Review summary"
    And stdout contains "Options:"
    And stdout contains "1) Open editor again"
    And stdout contains "2) Confirm"
    And stdout contains "3) Cancel"
    When I input "1"
    Then a process is launched with the resolved editor and argument equal to the created log file path
    When I input "2"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"

  Scenario: Cancel deletes the final file immediately
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Daily Notes"
    And I input "work"
    And I input "Draft message"
    Then exactly 1 log file exists in "./logs"
    And stdout contains "Review summary"
    When I input "cancel"
    Then the exit code is 0
    And no log files exist in "./logs"
    And stdout contains "Deleted"
    And stdout contains the deleted file path

  Scenario: Review screen shows all actions and aliases on-screen
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    And stdout contains "Options:"
    And stdout contains "1) Open editor"
    And stdout contains "3) Confirm (save & exit)"
    And stdout contains "4) Cancel (delete & exit)"
    And stdout contains "5) Quit and keep file"
    And stdout contains "You can also type: :wq, :wq!, :q, :q!"

  Scenario: :q deletes the created file at the review screen
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    And exactly 1 log file exists in "./logs"
    When I input ":q"
    Then the exit code is 0
    And no log files exist in "./logs"

  Scenario: :q! keeps the created file at the review screen
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    And exactly 1 log file exists in "./logs"
    When I input ":q!"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"

  Scenario: :wq confirms (save and exit) at the review screen
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    When I input ":wq"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"

  Scenario: :wq! confirms (save and exit) at the review screen
    Given I start "touchlog"
    When I input "1"
    And I input "./logs"
    And I input "Title"
    And I input "work"
    And I input "Draft"
    Then stdout contains "Review summary"
    When I input ":wq!"
    Then the exit code is 0
    And exactly 1 log file exists in "./logs"
```

---

## features/repl_ui.feature

```gherkin
Feature: REPL TUI implementation (Bubble Tea + Bubbles + Lip Gloss)

  Background:
    Given a clean temporary working directory

  Scenario: REPL renders menus using a selectable list component
    Given I start "touchlog"
    Then the UI renders a selectable list for "Select an action"
    And the UI highlights the current selection
    And the UI includes an on-screen help footer listing keybindings and commands

  Scenario: REPL supports keyboard navigation and shows it on-screen
    Given I start "touchlog"
    Then the UI help footer contains "↑/↓ to move"
    And the UI help footer contains "enter to select"
    And the UI help footer contains ":q to quit"

  Scenario: REPL shows validation errors inline without crashing the UI
    Given I start "touchlog"
    When I select "Create new entry"
    And I select "Enter a path"
    And I enter "/root/forbidden"
    Then the UI shows an inline error message containing "not writable"
    And the UI remains interactive on the same step

  Scenario: REPL shows a spinner while creating the file when filesystem is slow
    Given I start "touchlog"
    When I complete pre-creation fields
    And file creation takes longer than 200ms
    Then the UI shows a spinner labeled "Creating entry"
    And after creation completes the UI transitions to "Review summary"

  Scenario: REPL prints a stable final path in a copy-friendly way
    Given I start "touchlog"
    When I complete the wizard up to file creation
    Then the UI displays the created file path as a single uninterrupted line
```

---

## features/metadata.feature

```gherkin
Feature: Metadata capture and enrichment

  Background:
    Given a clean temporary working directory
    And the system time is fixed at "2025-12-31T12:00:00-07:00"

  Scenario: Record username and hostname when enabled
    Given a config file at "touchlog.yaml" with:
      '''
      metadata:
        include_user: true
        include_host: true
      '''
    When I run "touchlog new --config touchlog.yaml --message 'x'"
    Then the exit code is 0
    And the log file contains a "user" field
    And the log file contains a "host" field

  Scenario: Do not record user/host when disabled
    Given a config file at "touchlog.yaml" with:
      '''
      metadata:
        include_user: false
        include_host: false
      '''
    When I run "touchlog new --config touchlog.yaml --message 'x'"
    Then the log file does not contain a "user" field
    And the log file does not contain a "host" field

  Scenario: Include git context when inside a git repository
    Given a git repository initialized in the working directory
    And the current branch is "main"
    And the current commit hash is "abc1234"
    When I run "touchlog new --message 'x' --include-git"
    Then the log file content contains "branch: main"
    And the log file content contains "commit: abc1234"
```

---

## features/errors.feature

```gherkin
Feature: Robust errors and safe behavior

  Background:
    Given a clean temporary working directory

  Scenario: Output directory cannot be created
    Given the output directory path is "/root/forbidden" and is not writable
    When I run "touchlog new --output /root/forbidden --message 'x'"
    Then the exit code is not 0
    And stderr contains "permission denied"
    And no log files exist

  Scenario: Invalid UTF-8 in stdin is handled according to policy
    Given stdin contains invalid UTF-8 bytes
    When I run "touchlog new --stdin"
    Then the exit code is not 0
    And stderr contains "invalid UTF-8"

  Scenario: Config file path does not exist
    When I run "touchlog new --config does-not-exist.yaml --message 'x'"
    Then the exit code is not 0
    And stderr contains "config file not found"
    And no log files exist

  Scenario: Config file cannot be parsed
    Given a config file at "touchlog.yaml" with:
      '''
      output_dir: [this is not valid yaml
      '''
    When I run "touchlog new --config touchlog.yaml --message 'x'"
    Then the exit code is not 0
    And stderr contains "failed to parse config"
```

---

## features/list_search.feature

```gherkin
Feature: Listing and searching logs (future)

  Background:
    Given an output directory "./logs" containing:
      | filename                      | content                           |
      | 2025-12-30_work.md            | "tags: [work]\nMessage A"        |
      | 2025-12-31_touchlog.md        | "tags: [touchlog]\nMessage B"    |
      | 2025-12-31_touchlog_1.md      | "tags: [touchlog]\nMessage C"    |

  Scenario: List recent entries sorted by time descending
    When I run "touchlog list --output ./logs"
    Then the exit code is 0
    And stdout lists files in descending order by date/time
    And stdout contains "2025-12-31_touchlog_1.md" before "2025-12-31_touchlog.md"

  Scenario: Filter list by tag
    When I run "touchlog list --output ./logs --tag touchlog"
    Then stdout contains "2025-12-31_touchlog.md"
    And stdout contains "2025-12-31_touchlog_1.md"
    And stdout does not contain "2025-12-30_work.md"

  Scenario: Search by substring
    When I run "touchlog search --output ./logs --query 'Message C'"
    Then the exit code is 0
    And stdout contains "2025-12-31_touchlog_1.md"
```
