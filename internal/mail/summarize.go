package mail

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	htmlTagRe    = regexp.MustCompile(`<[^>]+>`)
	htmlEntityRe = regexp.MustCompile(`&[a-zA-Z]+;|&#[0-9]+;`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

var htmlEntities = map[string]string{
	"&amp;":   "&",
	"&lt;":    "<",
	"&gt;":    ">",
	"&quot;":  `"`,
	"&apos;":  "'",
	"&nbsp;":  " ",
	"&ndash;": "-",
	"&mdash;": "-",
	"&lsquo;": "'",
	"&rsquo;": "'",
	"&ldquo;": `"`,
	"&rdquo;": `"`,
}

// Summarize extracts a 2-line plain-text summary from a message body.
// Returns (line1, line2); line2 may be empty if the body is short.
func Summarize(body string, lineWidth int) (string, string) {
	text := htmlTagRe.ReplaceAllString(body, " ")
	text = htmlEntityRe.ReplaceAllStringFunc(text, func(e string) string {
		if v, ok := htmlEntities[strings.ToLower(e)]; ok {
			return v
		}
		return " "
	})
	text = whitespaceRe.ReplaceAllString(text, " ")
	text = strings.TrimFunc(text, unicode.IsSpace)

	if text == "" {
		return "", ""
	}

	line1 := wordWrap(text, lineWidth)
	rest := strings.TrimFunc(text[len(line1):], unicode.IsSpace)
	line2 := wordWrap(rest, lineWidth)

	return line1, line2
}

// wordWrap returns up to width characters of s, breaking at the last space.
func wordWrap(s string, width int) string {
	if len(s) <= width {
		return s
	}
	idx := strings.LastIndexFunc(s[:width], unicode.IsSpace)
	if idx <= 0 {
		return s[:width]
	}
	return s[:idx]
}
