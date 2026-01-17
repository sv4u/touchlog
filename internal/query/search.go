package query

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sv4u/touchlog/internal/store"
)

// SearchResult represents a single search result
type SearchResult struct {
	ID      string
	Type    string
	Key     string
	Title   string
	State   string
	Created string
	Updated string
	Path    string
	Tags    []string
}

// ExecuteSearch executes a search query and returns results
func ExecuteSearch(vaultRoot string, q *SearchQuery) ([]SearchResult, error) {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Build SQL query
	query, args := buildSearchQuery(q)

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	// Collect results
	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var tagsJSON sql.NullString
		if err := rows.Scan(&result.ID, &result.Type, &result.Key, &result.Title, &result.State, &result.Created, &result.Updated, &result.Path, &tagsJSON); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		// Parse tags JSON
		if tagsJSON.Valid {
			// Tags are stored as JSON array in the query result
			// For now, we'll parse them from the JSON
			// In a full implementation, we'd use a proper JSON parser
			result.Tags = parseTagsFromJSON(tagsJSON.String)
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// buildSearchQuery builds the SQL query from the search query AST
func buildSearchQuery(q *SearchQuery) (string, []interface{}) {
	var args []interface{}

	// Base query with tags aggregation
	query := `
		SELECT 
			n.id,
			n.type,
			n.key,
			n.title,
			n.state,
			n.created,
			n.updated,
			n.path,
			COALESCE(json_group_array(t.tag), '[]') as tags
		FROM nodes n
		LEFT JOIN tags t ON n.id = t.node_id
	`

	// Build WHERE clause
	var whereParts []string

	if len(q.Types) > 0 {
		placeholders := make([]string, len(q.Types))
		for i, typ := range q.Types {
			placeholders[i] = "?"
			args = append(args, typ)
		}
		whereParts = append(whereParts, fmt.Sprintf("n.type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(q.States) > 0 {
		placeholders := make([]string, len(q.States))
		for i, state := range q.States {
			placeholders[i] = "?"
			args = append(args, state)
		}
		whereParts = append(whereParts, fmt.Sprintf("n.state IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(q.Tags) > 0 {
		// Tag filtering requires a subquery or JOIN
		// For "match all tags", we need to ensure all tags are present
		// For "match any tag", we need at least one tag to match
		if q.MatchAnyTag {
			// Match any tag: use EXISTS
			placeholders := make([]string, len(q.Tags))
			for i, tag := range q.Tags {
				placeholders[i] = "?"
				args = append(args, tag)
			}
			whereParts = append(whereParts, fmt.Sprintf(
				"EXISTS (SELECT 1 FROM tags t2 WHERE t2.node_id = n.id AND t2.tag IN (%s))",
				strings.Join(placeholders, ","),
			))
		} else {
			// Match all tags: ensure count matches
			placeholders := make([]string, len(q.Tags))
			for i, tag := range q.Tags {
				placeholders[i] = "?"
				args = append(args, tag)
			}
			whereParts = append(whereParts, fmt.Sprintf(
				"(SELECT COUNT(DISTINCT t3.tag) FROM tags t3 WHERE t3.node_id = n.id AND t3.tag IN (%s)) = ?",
				strings.Join(placeholders, ","),
			))
			args = append(args, len(q.Tags))
		}
	}

	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	// GROUP BY for tag aggregation
	query += " GROUP BY n.id, n.type, n.key, n.title, n.state, n.created, n.updated, n.path"

	// ORDER BY for deterministic output
	query += " ORDER BY n.type, n.key"

	// LIMIT and OFFSET
	if q.Limit != nil && *q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", *q.Limit)
	}
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return query, args
}

// parseTagsFromJSON parses tags from a JSON array string
func parseTagsFromJSON(jsonStr string) []string {
	var tags []string
	if err := json.Unmarshal([]byte(jsonStr), &tags); err != nil {
		// If parsing fails, return empty slice
		return []string{}
	}
	return tags
}
