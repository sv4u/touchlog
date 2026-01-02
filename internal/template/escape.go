package template

import "strings"

const (
	// These are internal placeholders used for escaping template syntax in user input.
	// They are replaced before template processing and restored afterwards.
	templateBraceOpen  = "__TEMPLATE_BRACE_OPEN__"
	templateBraceClose = "__TEMPLATE_BRACE_CLOSE__"

	// Second-level placeholders used to escape any pre-existing placeholder strings
	// in user input. This prevents corruption when user input contains the placeholder strings.
	escapedPlaceholderOpen  = "__ESCAPED_PLACEHOLDER_OPEN__"
	escapedPlaceholderClose = "__ESCAPED_PLACEHOLDER_CLOSE__"
)

// escapePlaceholderRecursively escapes a single placeholder to its next level.
// Returns the escaped version and whether a change was made.
func escapePlaceholderRecursively(input, placeholder string) (string, bool) {
	if !strings.Contains(input, placeholder) {
		return input, false
	}

	// Map of known placeholders to their next level
	nextLevel := map[string]string{
		templateBraceOpen:       escapedPlaceholderOpen,
		templateBraceClose:      escapedPlaceholderClose,
		escapedPlaceholderOpen:  "__ESCAPED_PLACEHOLDER_2_OPEN__",
		escapedPlaceholderClose: "__ESCAPED_PLACEHOLDER_2_CLOSE__",
	}

	// Check if we know the next level for this placeholder
	if next, ok := nextLevel[placeholder]; ok {
		return strings.ReplaceAll(input, placeholder, next), true
	}

	// For numbered placeholders (level 2+), increment the number
	// Pattern: __ESCAPED_PLACEHOLDER_N_OPEN__ -> __ESCAPED_PLACEHOLDER_N+1_OPEN__
	// Only process if it's a numbered placeholder (has a digit after the prefix)
	if strings.HasPrefix(placeholder, "__ESCAPED_PLACEHOLDER_") {
		openSuffix := "_OPEN__"
		closeSuffix := "_CLOSE__"

		var suffix string
		var numStr string

		if strings.HasSuffix(placeholder, openSuffix) {
			suffix = openSuffix
			prefix := "__ESCAPED_PLACEHOLDER_"
			numStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, openSuffix), prefix)
		} else if strings.HasSuffix(placeholder, closeSuffix) {
			suffix = closeSuffix
			prefix := "__ESCAPED_PLACEHOLDER_"
			numStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, closeSuffix), prefix)
		} else {
			return input, false
		}

		// Only process if numStr is not empty and is a valid number
		// Empty numStr means it's a base placeholder (already handled by nextLevel map)
		if numStr == "" {
			return input, false
		}

		// Parse the number and increment it
		// For level 2 -> 3, level 3 -> 4, etc.
		if numStr == "2" {
			next := "__ESCAPED_PLACEHOLDER_3" + suffix
			return strings.ReplaceAll(input, placeholder, next), true
		}
		// For level 3+, parse the number properly
		// We'll use a simple approach: if it's a single digit, increment it
		if len(numStr) == 1 && numStr[0] >= '3' && numStr[0] <= '9' {
			nextNum := string(numStr[0] + 1)
			next := "__ESCAPED_PLACEHOLDER_" + nextNum + suffix
			return strings.ReplaceAll(input, placeholder, next), true
		}
		// For higher numbers or unknown formats, use a fallback pattern
		// This shouldn't happen in practice, but provides a safety net
		next := "__ESCAPED_PLACEHOLDER_" + numStr + "_NEXT" + suffix
		return strings.ReplaceAll(input, placeholder, next), true
	}

	return input, false
}

// EscapeUserInput escapes template syntax in user-provided input to prevent template injection.
// It escapes {{ and }} so they are treated as literal text rather than template variables.
//
// Strategy: Uses recursive/iterative escaping to handle any depth of placeholder strings:
// 1. Iteratively escape all placeholder strings until no changes are made (handles any depth)
// 2. Then escape the actual template braces ({{ and }})
//
// This prevents corruption when user input contains placeholder strings at any level.
// Reserved system variables (date, time, datetime) should not be escaped as they are trusted.
func EscapeUserInput(input string) string {
	if input == "" {
		return input
	}

	// Stage 1: Iteratively escape all placeholder strings that might exist in user input
	// Keep escaping until no more changes are made (handles any depth)
	escaped := input

	// Track which placeholders were in the original input
	// This allows us to only escape placeholders that were originally present,
	// not ones that were created during the escaping process
	originalPlaceholders := make(map[string]bool)
	allKnownPlaceholders := []string{
		templateBraceOpen,
		templateBraceClose,
		escapedPlaceholderOpen,
		escapedPlaceholderClose,
	}
	// Also check for numbered variants
	for level := 2; level <= 9; level++ {
		levelStr := string(rune('0' + level))
		allKnownPlaceholders = append(allKnownPlaceholders,
			"__ESCAPED_PLACEHOLDER_"+levelStr+"_OPEN__",
			"__ESCAPED_PLACEHOLDER_"+levelStr+"_CLOSE__")
	}
	for _, ph := range allKnownPlaceholders {
		if strings.Contains(input, ph) {
			originalPlaceholders[ph] = true
		}
	}

	maxIterations := 100 // Safety limit
	for iteration := 0; iteration < maxIterations; iteration++ {
		changed := false

		// Define the mapping of placeholders to their next level
		escapeMap := map[string]string{
			templateBraceOpen:       escapedPlaceholderOpen,
			templateBraceClose:      escapedPlaceholderClose,
			escapedPlaceholderOpen:  "__ESCAPED_PLACEHOLDER_2_OPEN__",
			escapedPlaceholderClose: "__ESCAPED_PLACEHOLDER_2_CLOSE__",
		}

		// Also add numbered levels (2 -> 3, 3 -> 4, etc.)
		// Limit to level 8 -> 9 to avoid creating level 10 (which would be ':')
		for level := 2; level <= 8; level++ {
			levelStr := string(rune('0' + level))
			nextLevelStr := string(rune('0' + level + 1))
			escapeMap["__ESCAPED_PLACEHOLDER_"+levelStr+"_OPEN__"] = "__ESCAPED_PLACEHOLDER_" + nextLevelStr + "_OPEN__"
			escapeMap["__ESCAPED_PLACEHOLDER_"+levelStr+"_CLOSE__"] = "__ESCAPED_PLACEHOLDER_" + nextLevelStr + "_CLOSE__"
		}

		// Check placeholders in order from lowest level to highest
		// This ensures base placeholders are escaped before their results are checked
		placeholdersToCheck := []string{
			templateBraceOpen,
			templateBraceClose,
			escapedPlaceholderOpen,
			escapedPlaceholderClose,
			"__ESCAPED_PLACEHOLDER_2_OPEN__", "__ESCAPED_PLACEHOLDER_2_CLOSE__",
			"__ESCAPED_PLACEHOLDER_3_OPEN__", "__ESCAPED_PLACEHOLDER_3_CLOSE__",
			"__ESCAPED_PLACEHOLDER_4_OPEN__", "__ESCAPED_PLACEHOLDER_4_CLOSE__",
			"__ESCAPED_PLACEHOLDER_5_OPEN__", "__ESCAPED_PLACEHOLDER_5_CLOSE__",
			"__ESCAPED_PLACEHOLDER_6_OPEN__", "__ESCAPED_PLACEHOLDER_6_CLOSE__",
			"__ESCAPED_PLACEHOLDER_7_OPEN__", "__ESCAPED_PLACEHOLDER_7_CLOSE__",
			"__ESCAPED_PLACEHOLDER_8_OPEN__", "__ESCAPED_PLACEHOLDER_8_CLOSE__",
			"__ESCAPED_PLACEHOLDER_9_OPEN__", "__ESCAPED_PLACEHOLDER_9_CLOSE__",
		}

		for _, placeholder := range placeholdersToCheck {
			// Only escape if:
			// 1. The placeholder exists in the escaped string, AND
			// 2. It was in the original input (not created during escaping), AND
			// 3. Its next level doesn't already exist (meaning we haven't escaped it yet)
			if strings.Contains(escaped, placeholder) && originalPlaceholders[placeholder] {
				nextLevel, hasNext := escapeMap[placeholder]
				if hasNext && !strings.Contains(escaped, nextLevel) {
					// Use the escapeMap directly
					escaped = strings.ReplaceAll(escaped, placeholder, nextLevel)
					changed = true
					// Mark the newly created placeholder as not original (so it won't be escaped again)
					originalPlaceholders[nextLevel] = false
					break // Only escape one per iteration
				}
			}
		}

		// If no changes were made, we've escaped all placeholder strings
		if !changed {
			break
		}
	}

	// Stage 2: Escape the actual template braces
	// We use a unique string that contains non-word characters so it won't match \{\{(\w+)\}\}
	escaped = strings.ReplaceAll(escaped, "{{", templateBraceOpen)
	escaped = strings.ReplaceAll(escaped, "}}", templateBraceClose)

	return escaped
}

// unescapePlaceholderRecursively unescapes a placeholder from its current level to the previous level.
// Returns the unescaped version and whether a change was made.
func unescapePlaceholderRecursively(input, placeholder string) (string, bool) {
	if !strings.Contains(input, placeholder) {
		return input, false
	}

	// Map of placeholders to their previous level (reverse of escape map)
	prevLevel := map[string]string{
		escapedPlaceholderOpen:            templateBraceOpen,
		escapedPlaceholderClose:           templateBraceClose,
		"__ESCAPED_PLACEHOLDER_2_OPEN__":  escapedPlaceholderOpen,
		"__ESCAPED_PLACEHOLDER_2_CLOSE__": escapedPlaceholderClose,
	}

	// Check if we know the previous level for this placeholder
	if prev, ok := prevLevel[placeholder]; ok {
		return strings.ReplaceAll(input, placeholder, prev), true
	}

	// For numbered placeholders (level 3+), decrement the number
	// Pattern: __ESCAPED_PLACEHOLDER_N_OPEN__ -> __ESCAPED_PLACEHOLDER_N-1_OPEN__
	if strings.HasPrefix(placeholder, "__ESCAPED_PLACEHOLDER_") {
		openSuffix := "_OPEN__"
		closeSuffix := "_CLOSE__"

		var suffix string
		var numStr string

		if strings.HasSuffix(placeholder, openSuffix) {
			suffix = openSuffix
			prefix := "__ESCAPED_PLACEHOLDER_"
			numStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, openSuffix), prefix)
		} else if strings.HasSuffix(placeholder, closeSuffix) {
			suffix = closeSuffix
			prefix := "__ESCAPED_PLACEHOLDER_"
			numStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, closeSuffix), prefix)
		} else {
			return input, false
		}

		// Handle level 3 -> 2, level 4 -> 3, etc.
		if numStr == "3" {
			prev := "__ESCAPED_PLACEHOLDER_2" + suffix
			return strings.ReplaceAll(input, placeholder, prev), true
		}
		// For level 4+, decrement the number
		if len(numStr) == 1 && numStr[0] >= '4' && numStr[0] <= '9' {
			prevNum := string(numStr[0] - 1)
			prev := "__ESCAPED_PLACEHOLDER_" + prevNum + suffix
			return strings.ReplaceAll(input, placeholder, prev), true
		}
		// For placeholders with "_NEXT" suffix (fallback pattern), remove it
		if strings.HasSuffix(numStr, "_NEXT") {
			prevNum := strings.TrimSuffix(numStr, "_NEXT")
			prev := "__ESCAPED_PLACEHOLDER_" + prevNum + suffix
			return strings.ReplaceAll(input, placeholder, prev), true
		}
	}

	return input, false
}

// UnescapeUserInput restores escaped template syntax back to literal braces.
// This should be called after template processing to restore the original braces.
//
// Strategy: Uses recursive/iterative unescaping (reverse of EscapeUserInput):
// 1. First, restore the actual template braces ({{ and }})
// 2. Then iteratively restore placeholder strings (from highest level to lowest) until no changes
//
// This ensures all placeholder strings are restored correctly regardless of depth.
//
// Known Limitation: The recursive approach cannot perfectly distinguish between placeholders
// that were original input vs. those created during escaping. This means that in some edge cases
// (e.g., when the original input is already at level 2 or higher), the round-trip might not
// perfectly preserve the original input. However, this works correctly for the common cases
// where the original input is at level 1 (templateBraceOpen) or contains actual template braces ({{).
func UnescapeUserInput(input string) string {
	if input == "" {
		return input
	}

	// Track which placeholders were in the original input (before escaping)
	// Key insight: We need to determine if escapedPlaceholderOpen was original or created.
	// - If escapedPlaceholderOpen exists in input AND there are no numbered placeholders, it was
	//   created from templateBraceOpen during escaping (should unescape to templateBraceOpen).
	// - If escapedPlaceholderOpen exists in input AND there are numbered placeholders, it was
	//   created from templateBraceOpen (should unescape to templateBraceOpen).
	// - If escapedPlaceholderOpen does NOT exist in input but numbered placeholders do, then
	//   escapedPlaceholderOpen was original and was escaped to numbered placeholders. After
	//   unescaping numbered placeholders, escapedPlaceholderOpen will be created. At that point,
	//   we should NOT unescape it further (it was original).
	hasEscapedPlaceholderInInput := strings.Contains(input, escapedPlaceholderOpen) || strings.Contains(input, escapedPlaceholderClose)

	// Find the highest numbered level in the input (if any)
	// This tells us what the original level was (original level = highest level - 1)
	highestNumberedLevel := 0
	for level := 2; level <= 9; level++ {
		levelStr := string(rune('0' + level))
		if strings.Contains(input, "__ESCAPED_PLACEHOLDER_"+levelStr+"_") {
			highestNumberedLevel = level
		}
	}

	// escapedPlaceholderWasOriginal is true only if:
	// - escapedPlaceholderOpen does NOT exist in input (it was escaped to numbered placeholders)
	// - AND numbered placeholders exist (confirming it was escaped)
	// This means: if escapedPlaceholderOpen exists in input, it was created (not original)
	escapedPlaceholderWasOriginal := !hasEscapedPlaceholderInInput && highestNumberedLevel > 0

	// Track the original level of escapedPlaceholderOpen (if it was original)
	// If the highest numbered level is N, then the original input had level N-1
	// For example, if highest level is 3, then the original input had level 2
	// But we need to find the lowest numbered level that was in the original input
	// If the original had level 2, after escaping it became level 3
	// So if highest level is 3, the original was at level 2
	// But we want to know what level escapedPlaceholderOpen was at originally
	// If the original input had level 2, then escapedPlaceholderOpen was NOT in the original
	// (it was at level 2, not level 1)
	// So we need to find the lowest numbered level in the input
	lowestNumberedLevel := 0
	for level := 2; level <= 9; level++ {
		levelStr := string(rune('0' + level))
		if strings.Contains(input, "__ESCAPED_PLACEHOLDER_"+levelStr+"_") {
			lowestNumberedLevel = level
			break
		}
	}

	// If escapedPlaceholderOpen was original, it was at level 1
	// But if the original input had numbered placeholders (level 2+), then escapedPlaceholderOpen
	// was NOT in the original input - the original was at a higher level
	// So originalEscapedPlaceholderLevel should be 1 only if escapedPlaceholderOpen was in the original
	// Otherwise, it should be 0 (meaning escapedPlaceholderOpen was not in the original)
	originalEscapedPlaceholderLevel := 0
	if escapedPlaceholderWasOriginal {
		// If the lowest numbered level is N, then the original input had level N-1
		// But escapedPlaceholderOpen was NOT in the original (it was at level N-1, not level 1)
		// So we should track that the original was at level N-1, not level 1
		// Actually, we want to know: when we unescape to level 1, should we stop?
		// If the original input had level 2, we should stop at level 2, not unescape to level 1
		// So originalEscapedPlaceholderLevel should be the level that was in the original input
		if lowestNumberedLevel > 0 {
			// The original input had level lowestNumberedLevel-1 (because it was escaped to lowestNumberedLevel)
			// But wait, if the original had level 2, after escaping it became level 3
			// So if lowestNumberedLevel is 3, the original was at level 2
			// But we want to know: should we stop at level 2? Yes.
			// So originalEscapedPlaceholderLevel should be lowestNumberedLevel - 1
			originalEscapedPlaceholderLevel = lowestNumberedLevel - 1
		} else {
			// Edge case: escapedPlaceholderWasOriginal is true but no numbered placeholders
			// This shouldn't happen in practice, but set to 1 (escapedPlaceholderOpen level) as fallback
			originalEscapedPlaceholderLevel = 1
		}
	}

	// Stage 1: Restore the escaped braces back to literal {{ and }}
	unescaped := strings.ReplaceAll(input, templateBraceOpen, "{{")
	unescaped = strings.ReplaceAll(unescaped, templateBraceClose, "}}")

	// Stage 2: Iteratively restore placeholder strings (from highest to lowest level)
	// Keep unescaping until no more changes are made (handles any depth)
	// Build the reverse mapping (next level -> previous level)
	unescapeMap := map[string]string{
		"__ESCAPED_PLACEHOLDER_2_OPEN__":  escapedPlaceholderOpen,
		"__ESCAPED_PLACEHOLDER_2_CLOSE__": escapedPlaceholderClose,
		escapedPlaceholderOpen:            templateBraceOpen,
		escapedPlaceholderClose:           templateBraceClose,
	}

	// Add numbered levels (3 -> 2, 4 -> 3, etc.)
	for level := 3; level <= 9; level++ {
		levelStr := string(rune('0' + level))
		prevLevelStr := string(rune('0' + level - 1))
		unescapeMap["__ESCAPED_PLACEHOLDER_"+levelStr+"_OPEN__"] = "__ESCAPED_PLACEHOLDER_" + prevLevelStr + "_OPEN__"
		unescapeMap["__ESCAPED_PLACEHOLDER_"+levelStr+"_CLOSE__"] = "__ESCAPED_PLACEHOLDER_" + prevLevelStr + "_CLOSE__"
	}

	// Check placeholders in order from highest level to lowest
	placeholdersToCheck := []string{
		"__ESCAPED_PLACEHOLDER_9_OPEN__", "__ESCAPED_PLACEHOLDER_9_CLOSE__",
		"__ESCAPED_PLACEHOLDER_8_OPEN__", "__ESCAPED_PLACEHOLDER_8_CLOSE__",
		"__ESCAPED_PLACEHOLDER_7_OPEN__", "__ESCAPED_PLACEHOLDER_7_CLOSE__",
		"__ESCAPED_PLACEHOLDER_6_OPEN__", "__ESCAPED_PLACEHOLDER_6_CLOSE__",
		"__ESCAPED_PLACEHOLDER_5_OPEN__", "__ESCAPED_PLACEHOLDER_5_CLOSE__",
		"__ESCAPED_PLACEHOLDER_4_OPEN__", "__ESCAPED_PLACEHOLDER_4_CLOSE__",
		"__ESCAPED_PLACEHOLDER_3_OPEN__", "__ESCAPED_PLACEHOLDER_3_CLOSE__",
		"__ESCAPED_PLACEHOLDER_2_OPEN__", "__ESCAPED_PLACEHOLDER_2_CLOSE__",
		escapedPlaceholderOpen,
		escapedPlaceholderClose,
	}

	maxIterations := 100 // Safety limit
	for iteration := 0; iteration < maxIterations; iteration++ {
		changed := false
		for _, placeholder := range placeholdersToCheck {
			// Unescape if the placeholder exists
			if strings.Contains(unescaped, placeholder) {
				prevLevel, hasPrev := unescapeMap[placeholder]
				if hasPrev {
					// Determine the current level of the placeholder
					currentLevel := 0
					if placeholder == escapedPlaceholderOpen || placeholder == escapedPlaceholderClose {
						currentLevel = 1
					} else if strings.HasPrefix(placeholder, "__ESCAPED_PLACEHOLDER_") {
						// Extract the level number
						levelStr := ""
						if strings.HasSuffix(placeholder, "_OPEN__") {
							levelStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, "_OPEN__"), "__ESCAPED_PLACEHOLDER_")
						} else if strings.HasSuffix(placeholder, "_CLOSE__") {
							levelStr = strings.TrimPrefix(strings.TrimSuffix(placeholder, "_CLOSE__"), "__ESCAPED_PLACEHOLDER_")
						}
						if levelStr != "" {
							// Parse the level number
							if len(levelStr) == 1 && levelStr[0] >= '2' && levelStr[0] <= '9' {
								currentLevel = int(levelStr[0] - '0')
							}
						}
					}

					// Special case for escapedPlaceholderOpen/Close -> templateBraceOpen/Close:
					// Only unescape if escapedPlaceholderOpen was NOT original (it was created from templateBraceOpen)
					if prevLevel == templateBraceOpen || prevLevel == templateBraceClose {
						// Only skip if escapedPlaceholderOpen was original AND all numbered placeholders have been unescaped
						if escapedPlaceholderWasOriginal {
							// Check if there are still numbered placeholders remaining
							hasNumberedRemaining := false
							for level := 2; level <= 9; level++ {
								levelStr := string(rune('0' + level))
								if strings.Contains(unescaped, "__ESCAPED_PLACEHOLDER_"+levelStr+"_") {
									hasNumberedRemaining = true
									break
								}
							}
							// Only skip if no numbered placeholders remain
							if !hasNumberedRemaining {
								continue
							}
						}
						// Otherwise, always unescape (it was created from templateBraceOpen or numbered placeholders)
					} else {
						// For numbered placeholders (level 2+), check if we're at the original level
						// If the original was at level N, and we're currently at level N, we should stop
						// (we've restored it to the original level)
						if escapedPlaceholderWasOriginal && currentLevel > 0 && originalEscapedPlaceholderLevel > 0 {
							// If we're at the original level, stop unescaping
							if currentLevel == originalEscapedPlaceholderLevel {
								continue
							}
						}
						// Also check if previous level already exists (heuristic)
						if strings.Contains(unescaped, prevLevel) {
							// Previous level exists, so it might have been original - skip unescaping
							continue
						}
					}
					unescaped = strings.ReplaceAll(unescaped, placeholder, prevLevel)
					changed = true
					break // Only unescape one per iteration
				}
			}
		}
		// If no changes were made, we've restored all placeholder strings
		if !changed {
			break
		}
	}

	return unescaped
}

// ShouldEscapeVariable determines if a variable should be escaped based on its name.
// System variables (date, time, datetime) are trusted and should not be escaped.
// User-provided variables (title, message, tags, and custom variables) should be escaped.
func ShouldEscapeVariable(varName string) bool {
	// System variables that are trusted (generated by touchlog)
	trustedVars := map[string]bool{
		"date":     true,
		"time":     true,
		"datetime": true,
		// Metadata variables (Phase 7) will also be trusted
		"user":   true,
		"host":   true,
		"branch": true,
		"commit": true,
	}

	// If it's a trusted variable, don't escape
	if trustedVars[varName] {
		return false
	}

	// All other variables (title, message, tags, custom vars) should be escaped
	return true
}
