package route

import (
	"strings"
	"unicode"
)

var stopwords = map[string]struct{}{
	"a": {}, "after": {}, "an": {}, "and": {}, "did": {}, "do": {}, "does": {}, "for": {}, "from": {},
	"give": {}, "good": {}, "in": {}, "is": {}, "it": {}, "me": {}, "of": {}, "on": {}, "or": {},
	"please": {}, "plus": {}, "should": {}, "than": {}, "the": {}, "this": {}, "to": {}, "use": {},
	"what": {}, "when": {}, "with": {}, "you": {}, "your": {},
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), " ")
}

func tokens(value string) []string {
	parts := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
	})
	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if len([]rune(part)) < 2 {
			continue
		}
		if _, skip := stopwords[part]; skip {
			continue
		}
		if _, exists := seen[part]; exists {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return result
}
