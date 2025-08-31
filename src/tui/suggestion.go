package tuipkg

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	configpkg "main/src/config"
)

type Config = configpkg.Config

type gotSuggestionsList []string
type keyMap struct{}

type SuggestionsType struct {
	Config      *Config
	Suggestions gotSuggestionsList
	Query       string
}

func NewSuggestions(config *Config) *SuggestionsType {
	return &SuggestionsType{
		Config:      config,
		Suggestions: gotSuggestionsList{},
		Query:       "",
	}
}

func (s *SuggestionsType) gotSuggestions() tea.Msg {
	list := s.Config.Config.Messages.Commands.Collection
	var suggestions gotSuggestionsList
	for _, item := range list {
		suggestions = append(suggestions, item.Command)
		for _, variant := range item.Variants {
			suggestions = append(suggestions, variant.Command)
		}
	}
	s.Suggestions = suggestions
	return suggestions
}

/**
 * TODO: Do not apply yet
 */
func (s *SuggestionsType) SearchSuggestions(query string) gotSuggestionsList {
	var results gotSuggestionsList
	query = strings.TrimPrefix(strings.ToLower(query), "/")

	for _, item := range s.Suggestions {
		if strings.Contains(strings.ToLower(item), query) {
			results = append(results, item)
			if len(results) >= 5 {
				break
			}
		}
	}
	return results
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next")),
		key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "prev")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
