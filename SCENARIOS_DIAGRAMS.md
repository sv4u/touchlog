# touchlog Scenario Diagrams

This document contains Mermaid diagrams visualizing the scenarios defined in `SCENARIOS.md`. The diagrams are organized by feature file, with overview diagrams for simpler features and detailed diagrams for complex scenarios.

---

## features/platform.feature

### Platform Support Flow

```mermaid
flowchart TD
    Start([touchlog starts]) --> CheckOS{Check Operating System}
    CheckOS -->|darwin| MacOS[macOS detected]
    CheckOS -->|linux| CheckWSL{Check WSL environment}
    CheckOS -->|windows| WindowsNative[Windows Native]
    
    MacOS --> RunVersion[Run --version command]
    CheckWSL -->|WSL detected| WSL[WSL environment]
    CheckWSL -->|Not WSL| Linux[Linux detected]
    WSL --> RunVersion
    Linux --> RunVersion
    
    WindowsNative --> Reject[Exit with error<br/>unsupported platform]
    
    RunVersion --> Success[Exit code 0]
    WSL --> WSLOutput[stdout contains WSL]
    
    Success --> End([End])
    WSLOutput --> End
    Reject --> End
```

---

## features/cli.feature

### CLI Command Routing

```mermaid
flowchart TD
    Start([touchlog command]) --> ParseArgs[Parse arguments]
    ParseArgs --> HasFlag{Has flag?}
    
    HasFlag -->|--help| ShowHelp[Print help text<br/>Exit code 0]
    HasFlag -->|--version| ShowVersion[Print version<br/>Exit code 0]
    HasFlag -->|--unknown| UnknownFlag[Print error<br/>unknown flag<br/>Exit code != 0]
    HasFlag -->|No flag| CheckSubcommand{Has subcommand?}
    
    CheckSubcommand -->|Yes| ValidateSubcommand{Valid subcommand?}
    CheckSubcommand -->|No| DefaultAction[Default action<br/>Launch REPL wizard]
    
    ValidateSubcommand -->|Yes| RouteSubcommand[Route to subcommand handler]
    ValidateSubcommand -->|No| UnknownCommand[Print error<br/>unknown command<br/>Exit code != 0]
    
    ShowHelp --> End([End])
    ShowVersion --> End
    UnknownFlag --> ShowHelp
    UnknownCommand --> ShowHelp
    RouteSubcommand --> End
    DefaultAction --> End
```

---

## features/config.feature

### Configuration Loading and Precedence

```mermaid
flowchart TD
    Start([touchlog command]) --> HasConfigFlag{--config flag provided?}
    
    HasConfigFlag -->|Yes| LoadConfigFile[Load config file]
    HasConfigFlag -->|No| UseDefaults[Use default config]
    
    LoadConfigFile --> CheckFormat{Config format?}
    CheckFormat -->|YAML| ParseYAML[Parse YAML]
    CheckFormat -->|TOML| ParseTOML[Parse TOML]
    
    ParseYAML --> ValidateConfig{Validate config}
    ParseTOML --> ValidateConfig
    
    ValidateConfig -->|Strict mode?| StrictCheck{Unknown keys?}
    ValidateConfig -->|Normal mode| MergeConfig[Merge with defaults]
    
    StrictCheck -->|Yes| RejectConfig[Exit with error<br/>unknown_key]
    StrictCheck -->|No| MergeConfig
    
    MergeConfig --> ApplyCLIOverrides{CLI flags override?}
    UseDefaults --> ApplyCLIOverrides
    
    ApplyCLIOverrides -->|Yes| OverrideValues[CLI flags take precedence]
    ApplyCLIOverrides -->|No| UseConfig[Use config values]
    
    OverrideValues --> FinalConfig[Final configuration]
    UseConfig --> FinalConfig
    RejectConfig --> End([End])
    FinalConfig --> End
```

---

## features/templates.feature

### Template Selection and Application

```mermaid
flowchart TD
    Start([touchlog new command]) --> LoadConfig[Load configuration]
    LoadConfig --> GetTemplateName{Get template name}
    
    GetTemplateName -->|From CLI --template flag| UseCLITemplate[Use CLI template name]
    GetTemplateName -->|From config| UseConfigTemplate[Use config template name]
    GetTemplateName -->|Neither| UseDefault[Use default template]
    
    UseCLITemplate --> FindTemplate{Template exists?}
    UseConfigTemplate --> FindTemplate
    UseDefault --> FindTemplate
    
    FindTemplate -->|Yes| LoadTemplate[Load template content]
    FindTemplate -->|No| TemplateError[Exit with error<br/>template not found]
    
    LoadTemplate --> CollectVars[Collect template variables<br/>date, title, message, tags]
    CollectVars --> RenderTemplate[Render template with variables]
    
    RenderTemplate --> WriteFile[Write to log file]
    TemplateError --> End([End])
    WriteFile --> End
```

---

## features/new.feature

### Non-Interactive Log Entry Creation Flow

```mermaid
flowchart TD
    Start([touchlog new]) --> ParseArgs[Parse arguments]
    ParseArgs --> ValidateMessage{Message provided?}
    
    ValidateMessage -->|--stdin| ReadStdin[Read from stdin]
    ValidateMessage -->|--message flag| UseMessage[Use message flag]
    ValidateMessage -->|Empty message| MessageError[Exit with error<br/>message must not be empty]
    
    ReadStdin --> ValidateUTF8{Valid UTF-8?}
    ValidateUTF8 -->|No| UTF8Error[Exit with error<br/>invalid UTF-8]
    ValidateUTF8 -->|Yes| ProcessStdin[Process stdin content]
    
    ProcessStdin --> CollectData[Collect entry data]
    UseMessage --> CollectData
    
    CollectData --> DetermineOutputDir{Output directory?}
    DetermineOutputDir -->|--output flag| UseCLIOutput[Use CLI output dir]
    DetermineOutputDir -->|From config| UseConfigOutput[Use config output dir]
    DetermineOutputDir -->|Default| UseDefaultOutput[Use default output dir]
    
    UseCLIOutput --> CreateDir{Directory exists?}
    UseConfigOutput --> CreateDir
    UseDefaultOutput --> CreateDir
    
    CreateDir -->|No| CreateDirectory[Create directory]
    CreateDir -->|Yes| CheckWritable{Directory writable?}
    CreateDirectory --> CheckWritable
    
    CheckWritable -->|No| PermissionError[Exit with error<br/>permission denied]
    CheckWritable -->|Yes| GenerateFilename[Generate filename<br/>date_slug.md]
    
    GenerateFilename --> CheckExists{File exists?}
    CheckExists -->|Yes| OverwriteFlag{--overwrite flag?}
    CheckExists -->|No| ApplyTemplate[Apply template]
    
    OverwriteFlag -->|Yes| ApplyTemplate
    OverwriteFlag -->|No| AddSuffix[Add numeric suffix<br/>_1.md or -1.md]
    AddSuffix --> ApplyTemplate
    
    ApplyTemplate --> WriteFile[Write log file]
    WriteFile --> Success[Exit code 0<br/>stdout: Wrote log]
    
    MessageError --> End([End])
    UTF8Error --> End
    PermissionError --> End
    Success --> End
```

---

## features/editor.feature

### Editor Resolution and Launch Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Config
    participant Env
    participant Editor
    participant Filesystem
    
    User->>CLI: touchlog new --edit
    CLI->>Config: Check config.editor
    Config-->>CLI: editor value or null
    
    alt Config has editor
        CLI->>CLI: Use config.editor
    else CLI has --editor flag
        CLI->>CLI: Use --editor flag
    else Check EDITOR env var
        CLI->>Env: Get EDITOR environment variable
        Env-->>CLI: EDITOR value or null
        
        alt EDITOR is set
            CLI->>CLI: Use EDITOR value
        else Fallback to defaults
            CLI->>CLI: Check for vi on PATH
            alt vi exists
                CLI->>CLI: Use vi
            else Check nano
                alt nano exists
                    CLI->>CLI: Use nano
                else No editor found
                    CLI->>Filesystem: Write file anyway
                    CLI->>User: Warn: no editor found - skipping edit
                end
            end
        end
    end
    
    CLI->>Filesystem: Create log file
    Filesystem-->>CLI: File created
    
    alt Editor resolved
        CLI->>Editor: Launch editor with file path
        Editor-->>CLI: Editor process started
        CLI->>CLI: Exit (no lifecycle management)
    else Editor not found
        CLI->>User: File created, warning shown
        CLI->>CLI: Exit successfully
    end
```

### Editor Precedence Decision Tree

```mermaid
flowchart TD
    Start([--edit flag provided]) --> CheckCLIFlag{--editor flag?}
    
    CheckCLIFlag -->|Yes| UseCLIEditor[Use CLI --editor flag]
    CheckCLIFlag -->|No| CheckConfig{Config has editor?}
    
    CheckConfig -->|Yes| UseConfigEditor[Use config editor]
    CheckConfig -->|No| CheckEnv{EDITOR env var?}
    
    CheckEnv -->|Yes| UseEnvEditor[Use EDITOR env var]
    CheckEnv -->|No| CheckVi{vi on PATH?}
    
    CheckVi -->|Yes| UseVi[Use vi]
    CheckVi -->|No| CheckNano{nano on PATH?}
    
    CheckNano -->|Yes| UseNano[Use nano]
    CheckNano -->|No| NoEditor[No editor found<br/>Write file + warn]
    
    UseCLIEditor --> LaunchEditor[Launch editor with file path]
    UseConfigEditor --> LaunchEditor
    UseEnvEditor --> LaunchEditor
    UseVi --> LaunchEditor
    UseNano --> LaunchEditor
    
    LaunchEditor --> EditorResult{Editor launch success?}
    EditorResult -->|Success| Exit[Exit code 0<br/>File remains]
    EditorResult -->|Failure| EditorError[Exit code != 0<br/>stderr: failed to launch editor<br/>File remains]
    
    NoEditor --> Exit
    EditorError --> End([End])
    Exit --> End
```

---

## features/repl_wizard.feature

### REPL Wizard State Machine

```mermaid
stateDiagram-v2
    [*] --> MainMenu: touchlog (no args)
    
    MainMenu --> SelectAction: Display menu
    SelectAction --> OutputDirPrompt: User selects "Create new entry"
    SelectAction --> [*]: User selects "Quit"
    
    OutputDirPrompt --> TitlePrompt: User enters output dir<br/>(Back available)
    OutputDirPrompt --> MainMenu: User selects "Back"
    
    TitlePrompt --> TagsPrompt: User enters title<br/>(Back available)
    TitlePrompt --> OutputDirPrompt: User selects "Back"
    
    TagsPrompt --> MessagePrompt: User enters tags<br/>(Back available)
    TagsPrompt --> TitlePrompt: User selects "Back"
    
    MessagePrompt --> CreateFile: User enters message<br/>(Back available)
    MessagePrompt --> TagsPrompt: User selects "Back"
    
    CreateFile --> FileCreated: Create log file<br/>(Back NOT available)
    
    FileCreated --> EditorPrompt: Show "Open editor?" prompt
    EditorPrompt --> ReviewScreen: User selects "No" or skips
    EditorPrompt --> LaunchEditor: User selects "Yes"
    
    LaunchEditor --> ReviewScreen: Editor launched<br/>Return to review
    
    ReviewScreen --> LaunchEditor: User selects Open editor again
    ReviewScreen --> Confirm: User selects Confirm or wq or wq!
    ReviewScreen --> Cancel: User selects Cancel or q
    ReviewScreen --> QuitKeepFile: User selects Quit and keep file or q!
    
    Confirm --> [*]: Exit code 0<br/>File saved
    Cancel --> [*]: Exit code 0<br/>File deleted
    QuitKeepFile --> [*]: Exit code 0<br/>File remains
```

### REPL Wizard Detailed Flow

```mermaid
flowchart TD
    Start([touchlog]) --> MainMenu[Main Menu<br/>Select an action]
    
    MainMenu --> UserInput1{User input}
    UserInput1 -->|1 - Create new entry| OutputDirStep[Output Directory Prompt]
    UserInput1 -->|Quit| Exit([Exit])
    
    OutputDirStep --> UserInput2{User input}
    UserInput2 -->|Enter path| ValidatePath{Path valid?}
    UserInput2 -->|Back| MainMenu
    UserInput2 -->|Use default| TitleStep[Title Prompt]
    
    ValidatePath -->|No| ShowError1[Show error<br/>Stay on step]
    ValidatePath -->|Yes| TitleStep
    ShowError1 --> UserInput2
    
    TitleStep --> UserInput3{User input}
    UserInput3 -->|Enter title| TagsStep[Tags Prompt]
    UserInput3 -->|Back| OutputDirStep
    UserInput3 -->|Skip| TagsStep
    
    TagsStep --> UserInput4{User input}
    UserInput4 -->|Enter tags| MessageStep[Message Prompt]
    UserInput4 -->|Back| TitleStep
    UserInput4 -->|Skip| MessageStep
    
    MessageStep --> UserInput5{User input}
    UserInput5 -->|Enter message| CreateFile[Create Log File]
    UserInput5 -->|Back| TagsStep
    UserInput5 -->|Skip| CreateFile
    
    CreateFile --> FileExists{File created?}
    FileExists -->|Yes| ShowCreated[Show created file path]
    FileExists -->|No| FileError[File creation error]
    
    ShowCreated --> EditorPrompt[Editor Prompt<br/>Open editor now?]
    EditorPrompt --> UserInput6{User input}
    
    UserInput6 -->|1 - Yes| LaunchEditor[Launch Editor]
    UserInput6 -->|2 - No| ReviewScreen[Review Summary Screen]
    
    LaunchEditor --> EditorLaunched{Editor launched?}
    EditorLaunched -->|Success| ReviewScreen
    EditorLaunched -->|Failure| ReviewScreen
    
    ReviewScreen --> ShowOptions["Show Options<br/>1) Open editor again<br/>2) Confirm<br/>3) Cancel<br/>4) Quit and keep file<br/>Also wq, wq!, q, q!"]
    
    ShowOptions --> UserInput7{User input}
    
    UserInput7 -->|1 or Open editor| LaunchEditor
    UserInput7 -->|2 or wq or wq!| Confirm[Confirm - Save & Exit]
    UserInput7 -->|3 or q| Cancel[Cancel - Delete & Exit]
    UserInput7 -->|4 or q!| QuitKeep[Quit - Keep File]
    
    Confirm --> Success[Exit code 0<br/>File saved]
    Cancel --> DeleteFile[Delete file]
    QuitKeep --> Success2[Exit code 0<br/>File remains]
    
    DeleteFile --> Success3[Exit code 0<br/>File deleted]
    
    FileError --> End([End])
    Success --> End
    Success2 --> End
    Success3 --> End
    Exit --> End
```

---

## features/repl_ui.feature

### REPL TUI Rendering and Interaction

```mermaid
flowchart TD
    Start([touchlog starts]) --> InitTUI[Initialize TUI<br/>Bubble Tea]
    
    InitTUI --> RenderMenu[Render Selectable List<br/>Menu Component]
    RenderMenu --> ShowHelp[Display Help Footer<br/>Keybindings: ↑/↓, enter, :q]
    
    ShowHelp --> UserInteraction{User interaction}
    
    UserInteraction -->|Keyboard navigation| HighlightSelection[Highlight current selection]
    UserInteraction -->|Enter| SelectOption[Select option]
    UserInteraction -->|:q| Quit[Quit application]
    
    HighlightSelection --> RenderMenu
    
    SelectOption --> ValidateInput{Input validation}
    ValidateInput -->|Valid| ProcessInput[Process input]
    ValidateInput -->|Invalid| ShowInlineError[Show inline error message<br/>Stay on same step]
    
    ShowInlineError --> UserInteraction
    
    ProcessInput --> CheckFileCreation{File creation step?}
    CheckFileCreation -->|Yes| CheckSlow{File creation > 200ms?}
    CheckFileCreation -->|No| NextStep[Move to next step]
    
    CheckSlow -->|Yes| ShowSpinner[Show spinner<br/>Creating entry]
    CheckSlow -->|No| NextStep
    
    ShowSpinner --> FileCreated{File created?}
    FileCreated -->|Yes| HideSpinner[Hide spinner]
    HideSpinner --> TransitionReview[Transition to Review summary]
    
    NextStep --> RenderMenu
    TransitionReview --> ShowFilePath[Display file path<br/>Single uninterrupted line<br/>Copy-friendly format]
    
    ShowFilePath --> UserInteraction
    Quit --> End([End])
```

---

## features/metadata.feature

### Metadata Capture and Enrichment Flow

```mermaid
flowchart TD
    Start([touchlog new]) --> CheckMetadataConfig{Metadata config enabled?}
    
    CheckMetadataConfig -->|include_user: true| GetUser[Get username]
    CheckMetadataConfig -->|include_user: false| SkipUser[Skip user metadata]
    
    CheckMetadataConfig -->|include_host: true| GetHost[Get hostname]
    CheckMetadataConfig -->|include_host: false| SkipHost[Skip host metadata]
    
    GetUser --> AddUserField[Add user field to log]
    GetHost --> AddHostField[Add host field to log]
    
    SkipUser --> CheckGitFlag{--include-git flag?}
    SkipHost --> CheckGitFlag
    AddUserField --> CheckGitFlag
    AddHostField --> CheckGitFlag
    
    CheckGitFlag -->|Yes| CheckGitRepo{Inside git repo?}
    CheckGitFlag -->|No| ApplyTemplate[Apply template]
    
    CheckGitRepo -->|Yes| GetGitInfo[Get git context:<br/>- Current branch<br/>- Commit hash]
    CheckGitRepo -->|No| ApplyTemplate
    
    GetGitInfo --> AddGitFields[Add git fields to log:<br/>branch: main<br/>commit: abc1234]
    AddGitFields --> ApplyTemplate
    
    ApplyTemplate --> WriteFile[Write log file with metadata]
    WriteFile --> End([End])
```

---

## features/errors.feature

### Error Handling and Safe Behavior

```mermaid
flowchart TD
    Start([touchlog command]) --> Operation[Execute operation]
    
    Operation --> CheckOutputDir{Output directory operation}
    CheckOutputDir -->|Create directory| CheckWritable{Directory writable?}
    CheckWritable -->|No| PermissionError[Exit code != 0<br/>stderr: permission denied<br/>No files created]
    CheckWritable -->|Yes| Continue1[Continue]
    
    Operation --> CheckStdin{Reading from stdin?}
    CheckStdin -->|Yes| ValidateUTF8{Valid UTF-8?}
    ValidateUTF8 -->|No| UTF8Error[Exit code != 0<br/>stderr: invalid UTF-8]
    ValidateUTF8 -->|Yes| Continue2[Continue]
    
    Operation --> CheckConfig{Using config file?}
    CheckConfig -->|Yes| CheckConfigExists{Config file exists?}
    CheckConfigExists -->|No| ConfigNotFound[Exit code != 0<br/>stderr: config file not found<br/>No files created]
    CheckConfigExists -->|Yes| ParseConfig[Parse config file]
    
    ParseConfig --> ConfigValid{Config valid?}
    ConfigValid -->|No| ConfigParseError[Exit code != 0<br/>stderr: failed to parse config]
    ConfigValid -->|Yes| Continue3[Continue]
    
    Continue1 --> Success([Success])
    Continue2 --> Success
    Continue3 --> Success
    
    PermissionError --> End([End])
    UTF8Error --> End
    ConfigNotFound --> End
    ConfigParseError --> End
    Success --> End
```

---

## features/list_search.feature (Future)

### List and Search Operations Flow

```mermaid
flowchart TD
    Start([touchlog list/search]) --> ParseCommand{Command type?}
    
    ParseCommand -->|list| ListCommand[List Command]
    ParseCommand -->|search| SearchCommand[Search Command]
    
    ListCommand --> LoadLogs[Load log files from output directory]
    LoadLogs --> HasTagFilter{--tag filter?}
    
    HasTagFilter -->|Yes| FilterByTag[Filter logs by tag]
    HasTagFilter -->|No| SortLogs[Sort logs by date/time]
    
    FilterByTag --> SortLogs
    SortLogs --> SortDescending[Sort descending<br/>Newest first]
    
    SortDescending --> DisplayList[Display list of files]
    
    SearchCommand --> LoadLogs2[Load log files from output directory]
    LoadLogs2 --> GetQuery[Get --query parameter]
    GetQuery --> SearchContent[Search file contents<br/>for query substring]
    SearchContent --> FilterResults[Filter matching files]
    FilterResults --> DisplayResults[Display search results]
    
    DisplayList --> End([End])
    DisplayResults --> End
```

---

## Cross-Feature Integration Overview

### Complete touchlog Command Flow

```mermaid
flowchart TD
    Start([User runs touchlog]) --> PlatformCheck[Platform Check<br/>macOS/Linux/WSL only]
    
    PlatformCheck --> ParseCLI[Parse CLI Arguments]
    ParseCLI --> HasSubcommand{Has subcommand?}
    
    HasSubcommand -->|No| LaunchREPL[Launch REPL Wizard<br/>Interactive TUI]
    HasSubcommand -->|new| NewCommand[new command]
    HasSubcommand -->|config| ConfigCommand[config command]
    HasSubcommand -->|list| ListCommand[List command<br/>Future]
    HasSubcommand -->|search| SearchCommand[Search command<br/>Future]
    
    NewCommand --> LoadConfig[Load Configuration<br/>YAML/TOML]
    LoadConfig --> ValidateConfig[Validate Config]
    ValidateConfig --> ResolveTemplate[Resolve Template]
    ResolveTemplate --> CollectMetadata[Collect Metadata<br/>user, host, git]
    CollectMetadata --> CreateEntry[Create Log Entry]
    CreateEntry --> ResolveEditor{--edit flag?}
    
    ResolveEditor -->|Yes| LaunchEditor[Launch Editor]
    ResolveEditor -->|No| Success[Success]
    
    LaunchEditor --> Success
    
    LaunchREPL --> REPLFlow[REPL Wizard Flow<br/>Interactive steps]
    REPLFlow --> CreateEntry
    
    ConfigCommand --> ValidateConfigOnly[Validate Config Only]
    ValidateConfigOnly --> Success
    
    ListCommand --> ListFlow[List/Filter Logs]
    SearchCommand --> SearchFlow[Search Logs]
    
    ListFlow --> Success
    SearchFlow --> Success
    
    Success --> End([End])
```

---

## Summary

This document provides visual representations of all scenarios defined in `SCENARIOS.md`:

- **Platform**: Simple platform detection flow
- **CLI**: Command routing and flag handling
- **Config**: Configuration loading with precedence rules
- **Templates**: Template selection and rendering
- **New**: Detailed non-interactive entry creation
- **Editor**: Editor resolution with fallback chain
- **REPL Wizard**: Complex state machine and detailed flow
- **REPL UI**: TUI rendering and interaction patterns
- **Metadata**: Metadata capture and enrichment
- **Errors**: Error handling and safe behavior
- **List/Search**: Future list and search operations

Each diagram type (flowchart, sequence, state diagram) is chosen based on what best represents the scenario's behavior and interactions.
