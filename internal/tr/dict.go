package tr

import "strings"

// DictLookup attempts to translate an English word or phrase into Chinese
// using the embedded dictionary. Returns (translation, true) on success.
// Only supports en -> zh; returns ("", false) otherwise.
func DictLookup(text string) (string, bool) {
	key := strings.TrimSpace(strings.ToLower(text))
	if key == "" {
		return "", false
	}

	// Try exact phrase match first
	if result, ok := dictEntries[key]; ok {
		return result, true
	}

	// Try word-by-word translation for multi-word inputs
	words := strings.Fields(key)
	if len(words) <= 1 {
		return "", false
	}

	translated := make([]string, 0, len(words))
	for _, w := range words {
		if t, ok := dictEntries[w]; ok {
			translated = append(translated, t)
		} else {
			translated = append(translated, w)
		}
	}
	// Return word-by-word result even if some words weren't found
	if len(translated) > 0 {
		return strings.Join(translated, " "), true
	}
	return "", false
}
