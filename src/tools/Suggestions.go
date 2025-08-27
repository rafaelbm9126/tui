package toolspkg

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	configpkg "main/src/config"
)

type SuggestionType struct {
	id          int
	name        string
	description string
}

var suggestions = []SuggestionType{}

func LoadSuggestions(config *configpkg.Config) {
	list := config.Config.Messages.Commands.List

	for idx, command := range list {
		suggestions = append(suggestions, SuggestionType{
			id:          idx + 1,
			name:        command[0],
			description: command[1],
		})
	}
}

func Suggestions(key tea.KeyType, text string, value []rune, search byte) []string {
	if len(text) <= 2 && key == tea.KeyBackspace {
		return []string{}
	}

	if len(text) > 0 {
		if text[0] == search {
			if key == tea.KeyBackspace {
				text = text[:len(text)-1]
			}
			if key == tea.KeyRunes {
				text = text + string(value)
			}

			preFormat := [][]string{}
			collection := []string{}
			for _, row := range SearchSuggestions(text) {
				preFormat = append(preFormat, []string{row.name, row.description})
			}
			for _, row := range FormatTableRows(preFormat, "- ", "") {
				collection = append(
					collection,
					row,
				)
			}
			return collection
		}
	}

	return []string{}
}

func SearchSuggestions(query string) []SuggestionType {
	var results []SuggestionType
	query = strings.TrimPrefix(strings.ToLower(query), "/")

	for _, s := range suggestions {
		if strings.Contains(strings.ToLower(s.name), query) {
			results = append(results, s)
			if len(results) >= 4 {
				break
			}
		}
	}
	return results
}
