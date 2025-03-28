package entity

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

type RequestError map[string][]string

func (e RequestError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e RequestError) Add(field, message string) RequestError {
	if e == nil {
		e = RequestError{}
	}
	if _, ok := e[field]; !ok {
		e[field] = []string{}
	}
	e[field] = append(e[field], message)
	return e
}

func (e RequestError) HasError() bool {
	return len(e) > 0
}

type Rowscan interface {
	// Scan *sql.Row|Rows.Scan
	Scan(dest ...any) error
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = removeDiacritics(slug)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func removeDiacritics(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, c := range s {
		switch c {
		case 'á', 'à', 'ã', 'â', 'ä': b.WriteRune('a')
		case 'é', 'è', 'ê', 'ë': b.WriteRune('e')
		case 'í', 'ì', 'î', 'ï': b.WriteRune('i')
		case 'ó', 'ò', 'õ', 'ô', 'ö': b.WriteRune('o')
		case 'ú', 'ù', 'û', 'ü': b.WriteRune('u')
		case 'ç': b.WriteRune('c')
		default:
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				b.WriteRune(c)
			} else {
				b.WriteRune('-')
			}
		}
	}
	return b.String()
}
