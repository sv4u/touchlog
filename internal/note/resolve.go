package note

import (
	"fmt"

	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// ResolveLinks resolves raw links to edges, returning resolved edges and diagnostics.
// This implements the link resolution rules:
//   - [[type:key]] -> direct resolution by full key
//   - [[key]] -> first try exact match on full key, then fall back to last-segment matching
//   - exact match takes priority to support path-based keys like [[projects/web/auth]]
//   - ambiguous if multiple matches found at either phase
//   - Unknown targets -> diagnostics + unresolved edge rows
func ResolveLinks(rawLinks []model.RawLink, typeKeyMap map[model.TypeKey]model.NoteID, lastSegmentMap map[string][]model.NoteID, sourceType model.TypeName) ([]model.RawLink, []model.Diagnostic) {
	var resolvedEdges []model.RawLink
	var diags []model.Diagnostic

	for _, link := range rawLinks {
		var targetID *model.NoteID
		var diagnostic *model.Diagnostic

		if link.Target.Type != nil {
			// Qualified link: [[type:key]]
			typeKey := model.TypeKey{
				Type: *link.Target.Type,
				Key:  link.Target.Key,
			}
			if id, ok := typeKeyMap[typeKey]; ok {
				targetID = &id
			} else {
				diagnostic = &model.Diagnostic{
					Level:   model.DiagnosticLevelWarn,
					Code:    "UNRESOLVED_LINK",
					Message: fmt.Sprintf("Link target '%s:%s' not found. The target note may not exist or may not have been indexed yet. Use 'touchlog index rebuild' to update the index.", *link.Target.Type, link.Target.Key),
					Span:    link.Span,
				}
			}
		} else {
			// Unqualified link: [[key]]
			// First try exact match on full key, then fall back to last-segment matching
			searchKey := string(link.Target.Key)

			// Priority 1: Try exact match on full key (across all types)
			var exactMatches []model.NoteID
			for typeKey, id := range typeKeyMap {
				if string(typeKey.Key) == searchKey {
					exactMatches = append(exactMatches, id)
				}
			}

			if len(exactMatches) == 1 {
				targetID = &exactMatches[0]
			} else if len(exactMatches) > 1 {
				diagnostic = &model.Diagnostic{
					Level:   model.DiagnosticLevelError,
					Code:    "AMBIGUOUS_LINK",
					Message: fmt.Sprintf("Link target '%s' is ambiguous - matches %d notes with the same key. Use a qualified link (type:key) to specify the target.", searchKey, len(exactMatches)),
					Span:    link.Span,
				}
				targetID = nil
			} else {
				// Priority 2: Fall back to last-segment matching
				lastSeg := config.LastSegment(searchKey)
				matchingIDs := lastSegmentMap[lastSeg]

				if len(matchingIDs) == 0 {
					diagnostic = &model.Diagnostic{
						Level:   model.DiagnosticLevelWarn,
						Code:    "UNRESOLVED_LINK",
						Message: fmt.Sprintf("Link target '%s' not found. The target note may not exist or may not have been indexed yet. Use a qualified link (type:key) or run 'touchlog index rebuild'.", searchKey),
						Span:    link.Span,
					}
				} else if len(matchingIDs) == 1 {
					targetID = &matchingIDs[0]
				} else {
					diagnostic = &model.Diagnostic{
						Level:   model.DiagnosticLevelError,
						Code:    "AMBIGUOUS_LINK",
						Message: fmt.Sprintf("Link target '%s' is ambiguous - matches %d notes with the same last segment. Use a qualified link (type:full/path/key) to specify the target.", searchKey, len(matchingIDs)),
						Span:    link.Span,
					}
					targetID = nil
				}
			}
		}

		// Create resolved edge with resolved ID
		resolvedLink := link
		resolvedLink.ResolvedToID = targetID
		resolvedEdges = append(resolvedEdges, resolvedLink)

		if diagnostic != nil {
			diags = append(diags, *diagnostic)
		}
	}

	return resolvedEdges, diags
}
