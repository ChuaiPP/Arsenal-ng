// Package loader handles loading and searching cheat files.
//
// This file contains search functionality that filters cheats based on query
// strings. It supports multi-word queries where all terms must match, and
// searches across tool names, tags, titles, commands, and descriptions.
package loader

import (
	"strings"

	"github.com/halilkirazkaya/arsenal-ng/internal/model"
)

// =============================================================================
// Search
// =============================================================================

// scoreResult holds a cheat and its fuzzy match score.
type scoreResult struct {
	cheat *model.Cheat
	score int
}

// Search filters and ranks cheats by fuzzy match quality.
// Results are sorted descending by score: best matches first.
// Supports multi-word queries, ranks by exact match > substring > fuzzy char sequence.
func Search(cheats []*model.Cheat, query string) []*model.Cheat {
	if query == "" {
		return cheats
	}

	query = strings.ToLower(query)
	terms := strings.Fields(query)

	var scored []scoreResult
	for _, cheat := range cheats {
		searchText := buildSearchText(cheat)
		score := 0
		matched := true
		for _, term := range terms {
			termScore := scoreTerm(searchText, term, cheat)
			if termScore == 0 {
				matched = false
				break
			}
			score += termScore
		}
		if matched {
			scored = append(scored, scoreResult{cheat: cheat, score: score})
		}
	}

	// Sort by score descending
	sortResults(scored)

	results := make([]*model.Cheat, len(scored))
	for i, sr := range scored {
		results[i] = sr.cheat
	}
	return results
}

// scoreTerm scores a single term against a cheat's search text.
// Higher scores = better match. Zero = no match.
func scoreTerm(searchText, term string, cheat *model.Cheat) int {
	// 1. Exact word match in tool name (highest priority)
	if strings.Contains(strings.ToLower(cheat.Tool), term) {
		return 150
	}
	// 2. Starts-with match in tool name
	if strings.HasPrefix(strings.ToLower(cheat.Tool), term) {
		return 130
	}
	// 3. Exact word match in title
	if strings.Contains(strings.ToLower(cheat.Title), term) {
		return 100
	}
	// 4. Tag match
	for _, tag := range cheat.Tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return 80
		}
	}
	// 5. Stage / tactic match
	if strings.Contains(strings.ToLower(cheat.Stage), term) ||
		strings.Contains(strings.ToLower(cheat.Tactic), term) {
		return 70
	}
	// 6. Substring match in search text
	if strings.Contains(searchText, term) {
		// Longer substring matches rank higher (less likely to be coincidence)
		return 50 + len(term)
	}
	// 7. Fuzzy char sequence match
	if score := fuzzyScore(searchText, term); score > 0 {
		return score
	}
	return 0
}

// fuzzyScore checks if all characters of 'term' appear in order in 'text'.
// Returns a score based on how close together the matching characters are.
// Returns 0 if not all characters match.
func fuzzyScore(text, term string) int {
	ti := 0
	lastMatch := -1
	score := 0
	consecutiveCount := 0

	for i, c := range text {
		if ti < len(term) && byte(c) == term[ti] {
			score += 10
			if lastMatch >= 0 && i-lastMatch <= 2 {
				consecutiveCount++
				score += consecutiveCount * 5 // Bonus for nearby matches
			} else {
				consecutiveCount = 0
			}
			lastMatch = i
			ti++
		}
	}

	if ti < len(term) {
		return 0 // Not all characters matched
	}
	return score
}

// sortResults sorts scored results in descending order (highest first).
func sortResults(results []scoreResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// buildSearchText creates a searchable string from all cheat fields.
func buildSearchText(cheat *model.Cheat) string {
	return strings.ToLower(
		cheat.Tool + " " +
			strings.Join(cheat.Tags, " ") + " " +
			cheat.Stage + " " +
			cheat.Tactic + " " +
			cheat.Title + " " +
			cheat.Command + " " +
			cheat.Desc,
	)
}

