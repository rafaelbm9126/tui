package toolspkg

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	configpkg "main/src/config"
)

type SuggestionType struct {
	Id          int
	Name        string
	Description string
}

var suggestions = []SuggestionType{}

func LoadSuggestions(config *configpkg.Config) {
	list := config.Config.Messages.Commands.List

	for idx, command := range list {
		suggestions = append(suggestions, SuggestionType{
			Id:          idx + 1,
			Name:        strings.ReplaceAll(command[0], "`", ""),
			Description: command[1],
		})
	}
}

func Suggestions(key tea.KeyType, text string, value []rune, search byte) []SuggestionType {
	if len(text) <= 2 && key == tea.KeyBackspace {
		return []SuggestionType{}
	}

	if len(text) > 0 {
		if text[0] == search {
			if key == tea.KeyBackspace {
				text = text[:len(text)-1]
			}
			if key == tea.KeyRunes {
				text = text + string(value)
			}
			return SearchSuggestions(text)
		}
	}

	return []SuggestionType{}
}

func SearchSuggestions(query string) []SuggestionType {
	var results []SuggestionType
	query = strings.TrimPrefix(strings.ToLower(query), "/")

	for _, s := range suggestions {
		if strings.Contains(strings.ToLower(s.Name), query) {
			results = append(results, s)
			if len(results) >= 3 {
				break
			}
		}
	}
	return results
}
