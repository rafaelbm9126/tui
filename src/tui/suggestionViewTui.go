package tuipkg

import (
	toolspkg "main/src/tools"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SuggestionType = toolspkg.SuggestionType

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type ListItem struct {
	list        list.Model
	items       []list.Item
	suggestions []SuggestionType
	width       int
	height      int
}

func NewListItem(width, height int) *ListItem {
	li := ListItem{
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), width, height),
		items:       []list.Item{},
		suggestions: []SuggestionType{},
		width:       width,
		height:      height,
	}
	li.list.SetShowTitle(false)
	li.list.SetShowStatusBar(false)
	li.list.SetFilteringEnabled(false)
	li.list.SetShowPagination(false)
	li.list.SetShowFilter(false)
	li.list.SetShowHelp(false)
	li.list.DisableQuitKeybindings()

	return &li
}

func (li *ListItem) SetWidth(width int) {
	li.width = width
	li.list.SetWidth(width)
}

func (li *ListItem) SetHeight(height int) {
	li.height = height
	li.list.SetHeight(height)
}

func (li *ListItem) ShowListItems() string {
	wrapper := viewport.New(li.width, li.height)
	wrapper.SetContent("")
	wrapper.Style = lipgloss.NewStyle().
		// Background(lipgloss.Color("#8A7DFC")).
		Margin(0, 0, 0, 2).
		Padding(0, 1, 0, 1)
	wrapper.SetContent(li.list.View())

	if len(li.items) > 0 {
		return wrapper.View()
	}
	return ""
}

func (li *ListItem) GetSuggestions(key tea.KeyType, text string, value []rune) {
	li.suggestions = toolspkg.Suggestions(key, text, value, '/')
}

func (li *ListItem) ShowItems(mdRender func(string) (string, error)) {
	li.items = []list.Item{}
	/**
	 * TODO: optimize
	 */
	for _, sugg := range li.suggestions {
		// title, err := mdRender(sugg.Name)
		// if err != nil {
		// 	continue
		// }
		// desc, err := mdRender(sugg.Description)
		// if err != nil {
		// 	continue
		// }
		li.items = append(
			li.items,
			item{
				title: sugg.Name,
				desc:  sugg.Description,
			},
		)
	}
	li.list.SetItems(li.items)
}

func (li *ListItem) SelectedItem() string {
	choice, ok := li.list.SelectedItem().(item)
	if !ok {
		return ""
	}
	return choice.title
}

func (li *ListItem) Clear() {
	li.items = []list.Item{}
	li.list.SetItems(li.items)
}
