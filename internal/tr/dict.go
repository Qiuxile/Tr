package tr

import "strings"

// DictLookup attempts to translate an English word or phrase into Chinese
// using the embedded dictionary. Returns (translation, true) on success.
// Only supports en -> zh; returns ("", false) otherwise.
//
// The lookup is deliberately conservative: it preserves the case and
// punctuation of the input, skips code-like tokens, and only reports success
// when at least one word actually matched. Reporting an unchanged sentence as
// a successful "translation" would hide the fact that offline mode has no
// entry for it.
func DictLookup(text string) (string, bool) {
	original := strings.TrimSpace(text)
	if original == "" {
		return "", false
	}

	// Try an exact phrase match first, ignoring case.
	if result, ok := dictEntries[strings.ToLower(original)]; ok {
		return result, true
	}

	words := strings.Fields(original)
	if len(words) <= 1 {
		return "", false
	}

	out := make([]string, 0, len(words))
	matches := 0
	for _, word := range words {
		translated, ok := translateDictToken(word)
		if ok {
			out = append(out, translated)
			matches++
			continue
		}
		out = append(out, word)
	}
	if matches == 0 {
		return "", false
	}
	return strings.Join(out, " "), true
}

// translateDictToken translates one whitespace-delimited token while keeping
// its surrounding punctuation. Code-like tokens (identifiers, acronyms, paths,
// numbers) are left untouched.
func translateDictToken(word string) (string, bool) {
	core := strings.Trim(word, `.,!?;:"'()[]{}<>`)
	if core == "" || isCodeLikeToken(core) {
		return "", false
	}
	key := strings.ToLower(core)
	translation, ok := dictEntries[key]
	if !ok {
		return "", false
	}
	idx := strings.Index(strings.ToLower(word), key)
	if idx < 0 {
		return translation, true
	}
	return word[:idx] + translation + word[idx+len(key):], true
}

// isCodeLikeToken keeps the offline dictionary from mangling identifiers,
// paths and all-caps acronyms in mixed prose/code text.
func isCodeLikeToken(token string) bool {
	if token == "" {
		return true
	}
	if strings.ContainsAny(token, "_/\\@#$%^&*+=|~`") {
		return true
	}
	for _, r := range token {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	// Acronyms such as CLI/API/JSON are far more likely to be code than prose.
	upper := strings.ToUpper(token)
	if upper == token && len(token) <= 6 {
		for _, r := range token {
			if r >= 'A' && r <= 'Z' {
				return true
			}
		}
	}
	return false
}
