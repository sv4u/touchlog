package errors

import (
	"errors"
	"fmt"
)

// Platform errors
var (
	// ErrPlatformUnsupported is returned when the platform is not supported
	ErrPlatformUnsupported = errors.New("unsupported platform: touchlog only supports macOS, Linux, and WSL")
)

// Configuration errors
var (
	// ErrConfigNotFound is returned when a config file cannot be found
	ErrConfigNotFound = errors.New("config file not found")
	
	// ErrConfigInvalid is returned when a config file cannot be parsed
	ErrConfigInvalid = errors.New("failed to parse config")
	
	// ErrConfigReadFailed is returned when a config file cannot be read
	ErrConfigReadFailed = errors.New("failed to read config file")
	
	// ErrConfigWriteFailed is returned when a config file cannot be written
	ErrConfigWriteFailed = errors.New("failed to write config file")
	
	// ErrConfigUnknownKeys is returned when strict mode validation finds unknown keys
	ErrConfigUnknownKeys = errors.New("unknown config keys found")
	
	// ErrConfigInvalidVariableName is returned when a variable name is invalid or reserved
	ErrConfigInvalidVariableName = errors.New("invalid variable name")
)

// Template errors
var (
	// ErrTemplateNotFound is returned when a template cannot be found
	ErrTemplateNotFound = errors.New("template not found")
	
	// ErrTemplateReadFailed is returned when a template file cannot be read
	ErrTemplateReadFailed = errors.New("failed to read template file")
	
	// ErrTemplateInvalidSyntax is returned when a template has invalid syntax
	ErrTemplateInvalidSyntax = errors.New("invalid template syntax")
	
	// ErrTemplateNameEmpty is returned when a template name is empty
	ErrTemplateNameEmpty = errors.New("template name cannot be empty")
)

// Entry/File errors
var (
	// ErrEntryCreationFailed is returned when entry creation fails
	ErrEntryCreationFailed = errors.New("failed to create entry")
	
	// ErrFileExists is returned when a file already exists and overwrite is not enabled
	ErrFileExists = errors.New("file already exists")
	
	// ErrFileWriteFailed is returned when writing to a file fails
	ErrFileWriteFailed = errors.New("failed to write file")
	
	// ErrFileReadFailed is returned when reading from a file fails
	ErrFileReadFailed = errors.New("failed to read file")
	
	// ErrFilenameGenerationFailed is returned when filename generation fails
	ErrFilenameGenerationFailed = errors.New("failed to generate filename")
	
	// ErrFilenameCollision is returned when all filename collision attempts are exhausted
	ErrFilenameCollision = errors.New("unable to find available filename after maximum attempts")
)

// Directory/Path errors
var (
	// ErrOutputDirRequired is returned when output directory is not set
	ErrOutputDirRequired = errors.New("output directory is required")
	
	// ErrOutputDirInvalid is returned when output directory path is invalid
	ErrOutputDirInvalid = errors.New("invalid output directory path")
	
	// ErrOutputDirNotWritable is returned when output directory is not writable
	ErrOutputDirNotWritable = errors.New("directory is not writable")
	
	// ErrOutputDirNotDirectory is returned when output directory path exists but is not a directory
	ErrOutputDirNotDirectory = errors.New("path exists but is not a directory")
	
	// ErrPathExpansionFailed is returned when path expansion (e.g., ~) fails
	ErrPathExpansionFailed = errors.New("failed to expand path")
	
	// ErrPathInvalid is returned when a path is invalid
	ErrPathInvalid = errors.New("invalid path")
)

// Editor errors
var (
	// ErrEditorNotFound is returned when an editor cannot be found
	ErrEditorNotFound = errors.New("editor not found")
	
	// ErrEditorLaunchFailed is returned when editor launch fails
	ErrEditorLaunchFailed = errors.New("failed to launch editor")
	
	// ErrEditorExitedWithError is returned when editor exits with an error
	ErrEditorExitedWithError = errors.New("editor exited with error")
	
	// ErrEditorCommandEmpty is returned when editor command is empty
	ErrEditorCommandEmpty = errors.New("editor command cannot be empty")
	
	// ErrEditorNotExecutable is returned when editor is not executable
	ErrEditorNotExecutable = errors.New("editor is not executable")
)

// Validation errors
var (
	// ErrInvalidUTF8 is returned when input contains invalid UTF-8
	ErrInvalidUTF8 = errors.New("invalid UTF-8")
	
	// ErrValidationFailed is returned when validation fails
	ErrValidationFailed = errors.New("validation failed")
)

// Permission errors
var (
	// ErrPermissionDenied is returned when permission is denied
	ErrPermissionDenied = errors.New("permission denied")
)

// Timezone errors
var (
	// ErrTimezoneInvalid is returned when timezone is invalid
	ErrTimezoneInvalid = errors.New("invalid timezone")
)

// Wizard errors
var (
	// ErrWizardStateInvalid is returned when wizard state is invalid
	ErrWizardStateInvalid = errors.New("invalid wizard state")
	
	// ErrWizardTransitionInvalid is returned when wizard state transition is invalid
	ErrWizardTransitionInvalid = errors.New("invalid wizard state transition")
	
	// ErrWizardNavigationNotAllowed is returned when back navigation is not allowed
	ErrWizardNavigationNotAllowed = errors.New("back navigation not allowed")
)

// Metadata errors
var (
	// ErrMetadataCollectionFailed is returned when metadata collection fails
	ErrMetadataCollectionFailed = errors.New("failed to collect metadata")
	
	// ErrGitContextFailed is returned when git context cannot be retrieved
	ErrGitContextFailed = errors.New("failed to get git context")
)

// WrapError wraps an error with additional context
// This is a convenience function for adding context to errors
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted additional context
// This is a convenience function for adding formatted context to errors
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

