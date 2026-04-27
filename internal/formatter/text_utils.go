package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"hsanalyzer/internal/textutil"
)

var (
	effectChangePattern = textutil.EffectChangePattern
)

type effectDetails struct {
	Label       string
	Description string
}

func marshalIndentedJSON(value interface{}) (string, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return "", err
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

func cleanCardText(text string) string {
	return textutil.CleanCardText(text)
}

// CleanCardText normalizes card text for external consumers using the same
// policy as the existing JSON formatter.
func CleanCardText(text string) string {
	return textutil.CleanCardText(text)
}

func describeEffect(effectName, sourceName, sourceText, language string) effectDetails {
	if matches := effectChangePattern.FindStringSubmatch(effectName); matches != nil {
		value, _ := strconv.Atoi(matches[2])
		switch matches[1] {
		case "ATK":
			if language == "en" {
				return effectDetails{
					Label:       "Attack changed",
					Description: fmt.Sprintf("Attack is now %d.", value),
				}
			}
			return effectDetails{
				Label:       "Изменение атаки",
				Description: fmt.Sprintf("Атака теперь равна %d.", value),
			}
		case "HEALTH":
			if language == "en" {
				return effectDetails{
					Label:       "Health changed",
					Description: fmt.Sprintf("Health is now %d.", value),
				}
			}
			return effectDetails{
				Label:       "Изменение здоровья",
				Description: fmt.Sprintf("Здоровье теперь равно %d.", value),
			}
		}
	}

	if label, ok := effectKeywordLabel(effectName, language); ok {
		description := sourceText
		if description == "" {
			if language == "en" {
				if sourceName != "" {
					description = fmt.Sprintf("Triggered by %s.", sourceName)
				} else {
					description = "The effect has triggered."
				}
			} else {
				if sourceName != "" {
					description = fmt.Sprintf("Сработал эффект карты %s.", sourceName)
				} else {
					description = "Эффект сработал."
				}
			}
		}

		return effectDetails{
			Label:       label,
			Description: description,
		}
	}

	if language == "en" {
		return effectDetails{
			Label:       prettifyEffectCode(effectName),
			Description: sourceText,
		}
	}

	return effectDetails{
		Label:       prettifyEffectCode(effectName),
		Description: sourceText,
	}
}

func effectKeywordLabel(effectName, language string) (string, bool) {
	labelsRU := map[string]string{
		"TRIGGER_VISUAL": "Срабатывание эффекта",
		"BATTLECRY":      "Боевой клич",
		"DEATHRATTLE":    "Предсмертный хрип",
		"COMBO":          "Серия приемов",
		"SECRET":         "Секрет",
		"DISCOVER":       "Раскопка",
	}
	labelsEN := map[string]string{
		"TRIGGER_VISUAL": "Effect triggered",
		"BATTLECRY":      "Battlecry",
		"DEATHRATTLE":    "Deathrattle",
		"COMBO":          "Combo",
		"SECRET":         "Secret",
		"DISCOVER":       "Discover",
	}

	if language == "en" {
		label, ok := labelsEN[effectName]
		return label, ok
	}

	label, ok := labelsRU[effectName]
	return label, ok
}

func prettifyEffectCode(effectName string) string {
	if strings.TrimSpace(effectName) == "" {
		return ""
	}

	parts := strings.Split(strings.ToLower(effectName), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}
