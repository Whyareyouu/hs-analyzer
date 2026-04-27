package textutil

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	htmlTagPattern      = regexp.MustCompile(`<[^>]+>`)
	blankLinePattern    = regexp.MustCompile(`\n{3,}`)
	spaceBeforeNL       = regexp.MustCompile(`[ \t]+\n`)
	EffectChangePattern = regexp.MustCompile(`^(ATK|HEALTH)_CHANGE:(-?\d+)$`)
)

// CleanCardText normalizes Hearthstone card text into a plain-text payload.
func CleanCardText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"[x]", "",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"\r\n", "\n",
		"\r", "\n",
	)

	text = replacer.Replace(text)
	text = html.UnescapeString(text)
	text = htmlTagPattern.ReplaceAllString(text, "")
	text = spaceBeforeNL.ReplaceAllString(text, "\n")
	text = blankLinePattern.ReplaceAllString(text, "\n\n")

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.Join(strings.Fields(line), " ")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// ParseEffectChange extracts ATK/HEALTH delta values from formatted effect codes.
func ParseEffectChange(effectName string) (field string, value int, ok bool) {
	matches := EffectChangePattern.FindStringSubmatch(effectName)
	if matches == nil {
		return "", 0, false
	}

	value, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, false
	}

	return matches[1], value, true
}
