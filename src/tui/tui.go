package tuipkg

import (
	"context"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	buspkg "main/src/bus"
	commandpkg "main/src/command"
	configpkg "main/src/config"
	eventpkg "main/src/event"
	messagepkg "main/src/message"
	modelpkg "main/src/model"
	toolspkg "main/src/tools"
)

type MessageList = messagepkg.MessageList
type MessageModel = messagepkg.MessageModel

type OptimizedBus = buspkg.OptimizedBus

type Event = eventpkg.Event

type Command = commandpkg.Command

const (
	LEFT_WIDTH_PERCENTAGE = 0.6 // 60% del ancho
	inputHeight           = 3   // borde + entrada + borde
	headerHeight          = 1
	footerHeight          = 1
)

type TUI struct {
	width       int
	height      int
	viewport    viewport.Model
	input       textinput.Model
	spinner     spinner.Model
	program     *tea.Program
	logger      *slog.Logger
	messages    *MessageList
	bus         *OptimizedBus
	command     *Command
	mdEnabled   bool
	mdRendererA *glamour.TermRenderer // with Margin
	mdRendererB *glamour.TermRenderer // with out Margin
	mdWrapWidth int                   // ancho con el que se construyó el renderer
	history     []string
	histIndex   int
	showAlert   bool
	textAlert   string

	listItem *ListItem

	styles struct {
		header         lipgloss.Style
		labelSystem    lipgloss.Style
		labelHuman     lipgloss.Style
		labelAssistant lipgloss.Style
		body           lipgloss.Style
		dots           lipgloss.Style
		help           lipgloss.Style
		inputBox       lipgloss.Style
		alert          lipgloss.Style
	}
}

func NewTUI(
	conf *configpkg.Config,
	bus *OptimizedBus,
	messages *MessageList,
	command *Command,
	logger *slog.Logger,
) *TUI {
	toolspkg.LoadSuggestions(conf)

	vp := viewport.New(80, 20)
	vp.SetContent("")
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8A7DFC")).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		Margin(0, 0, 0, 2)

	ti := textinput.New()
	ti.Placeholder = "Escribe un mensaje o comando (/help)"
	ti.CharLimit = 2048
	ti.Width = 80
	ti.SetValue("")
	ti.Focus()

	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC400"))
	sp.Spinner = spinner.Dot

	logger.Info("Initializing TUI")

	t := &TUI{
		width:     80,
		height:    20,
		viewport:  vp,
		input:     ti,
		spinner:   sp,
		mdEnabled: true,
		bus:       bus,
		messages:  messages,
		command:   command,
		logger:    logger,
		showAlert: false,
		listItem:  NewListItem(0, 0),
	}

	s := &t.styles

	s.header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	s.labelSystem = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("#E07093"))
	s.labelHuman = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("#38ACEC"))
	s.labelAssistant = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("#29BEB0"))
	s.body = lipgloss.NewStyle().PaddingLeft(1)
	s.dots = lipgloss.NewStyle().Foreground(lipgloss.Color("#444"))
	s.help = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	s.inputBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8A7DFC")).
		Margin(0, 2, 0, 2).
		Padding(0, 1)
	s.alert = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC400"))

	t.program = tea.NewProgram(t, tea.WithAltScreen())

	return t
}

func (t *TUI) Init() tea.Cmd {
	t.MDRenderer(t.viewport.Width)

	return tea.Batch(
		tea.EnterAltScreen,
		t.spinner.Tick,
	)
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	listItemFocus := len(t.listItem.items) > 0

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			cmds = append(cmds, tea.Quit)

		case tea.KeyEnter:
			choise := t.listItem.SelectedItem()
			if len(choise) > 0 {
				// Selected item //
				t.input.SetValue(choise + " ")
				t.listItem.Clear()
				t.input.CursorEnd()
			} else {
				rawText := t.input.Value()
				text := strings.TrimSpace(rawText)
				if text != "" {
					msg := MessageModel{
						Type:   modelpkg.TyText,
						Source: modelpkg.ScHuman,
						Text:   text,
					}
					t.bus.Publish(eventpkg.EvtMessage, msg)
					t.input.Reset()
					t.input.SetValue("")
					if len(t.history) == 0 || t.history[len(t.history)-1] != text {
						t.history = append(t.history, text)
					}
					t.histIndex = len(t.history)
				}
			}

		case tea.KeyUp:
			if !listItemFocus {
				// Navega por el historial hacia arriba
				if t.histIndex > 0 {
					t.histIndex--
					t.input.SetValue(t.history[t.histIndex])
					t.input.CursorEnd()
				}
			}

		case tea.KeyDown:
			if !listItemFocus {
				// Navega por el historial hacia abajo
				if t.histIndex < len(t.history) {
					t.histIndex++
					if t.histIndex == len(t.history) {
						t.input.SetValue("")
					} else {
						t.input.SetValue(t.history[t.histIndex])
					}
					t.input.CursorEnd()
				}
			}
		}

		/**
		 * COMMAND SUGGESTIONS
		 */
		t.listItem.GetSuggestions(msg.Type, t.input.Value(), msg.Runes)

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

		viewportHeight := t.height - headerHeight - inputHeight - footerHeight
		t.viewport.Width = int(float64(t.width) * LEFT_WIDTH_PERCENTAGE)
		t.viewport.Height = viewportHeight
		t.input.Width = int(float64(t.width)*LEFT_WIDTH_PERCENTAGE) - 11 // -5 ajuste para igualar al viewport
		t.listItem.SetWidth(t.input.Width)

		t.RenderBody()

	case Event:
		evt := msg.Evt
		switch evt.Type {
		case eventpkg.EvtSystem:
			icmd, _ := evt.Data.(string)
			switch icmd {
			case "quit":
				cmds = append(cmds, tea.Quit)

			case "loading":
				t.textAlert = t.styles.alert.
					Align(lipgloss.Right).
					Render("loading")
				t.showAlert = !t.showAlert

			default:
				panic("Command Unknown")
			}

		case eventpkg.EvtMessage:
			if msgData, ok := evt.Data.(MessageModel); ok {
				switch msgData.Source {
				case modelpkg.ScSystem:
					t.messages.AddMessage(msgData)
				case modelpkg.ScHuman:
					isCmd, showMsg := t.command.IsCommandThenRun(msgData.Text)
					if isCmd {
						msgData.Type = modelpkg.TyCommand
					}
					if showMsg {
						t.messages.AddMessage(msgData)
					}
				case modelpkg.ScAssistant:
					/**
					 * TODO: review assistant executor comand
					 */
					// isCmd := t.command.IsCommandThenRun(msgData.Text)
					// if isCmd {
					// 	msgData.Type = modelpkg.TyCommand
					// }
					t.messages.AddMessage(msgData)
				}
			}
		}
		t.RenderBody()

	default:
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	}

	t.input, cmd = t.input.Update(msg)
	cmds = append(cmds, cmd)

	t.viewport, cmd = t.viewport.Update(msg)
	cmds = append(cmds, cmd)

	t.listItem.list, cmd = t.listItem.list.Update(msg)
	cmds = append(cmds, cmd)

	return t, tea.Batch(cmds...)
}

func (t *TUI) View() string {
	input := t.styles.inputBox.Render(t.input.View())

	nvp := viewport.New(
		int(float64(t.width)*(1-LEFT_WIDTH_PERCENTAGE)),
		t.viewport.Height+1,
	)
	nvp.SetContent("")
	nvp.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("#8A7DFC")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8A7DFC")).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(true).
		Margin(1, 0, 0, 0)

	// SUGGESTIONS //
	t.listItem.ShowItems(t.mdRendererB.Render)
	adjust := 0
	if len(t.listItem.items) > 0 {
		adjust = 9
	}
	t.viewport.Height = t.height - headerHeight - inputHeight - footerHeight - adjust - 1
	t.listItem.SetHeight(adjust)
	// --- //

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left,
			HeaderViewTui(t),           // header
			t.viewport.View(),          // body
			t.listItem.ShowListItems(), // suggestions
			FooterViewTui(t),           // footer
			input,                      // input
		),
		nvp.View(),
	)
}

func (t *TUI) Run(ctx context.Context, cancel context.CancelFunc) (*tea.Program, error) {
	go func() {
		<-ctx.Done()
		t.program.Quit()
	}()

	_, err := t.program.Run()

	if ctx.Err() != nil {
		return t.program, ctx.Err()
	}

	return t.program, err
}

func (t *TUI) Program() *tea.Program {
	return t.program
}

func (t *TUI) RenderBody() {
	content := t.PrePrintMessages()
	t.viewport.SetContent(content)
	t.viewport.GotoBottom()
}

func (t *TUI) DottedLine(width int) string {
	dots := strings.Repeat("·", width)
	return t.styles.dots.
		Margin(1, 0, 0, 0).
		Render(dots)
}

func (t *TUI) MDRenderer(width int) {
	if !t.mdEnabled {
		t.mdRendererA = nil
		t.mdRendererB = nil
		t.mdWrapWidth = 0
		return
	}

	// Crea el renderer solo si ha cambiado el ancho o no existe
	if t.mdRendererA == nil || t.mdWrapWidth != width {
		rendererA, errA := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			/**
			 * TODO: define size viewport adjustment
			 */
			glamour.WithWordWrap(width-14),
		)
		rendererB, errB := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		)
		if errA != nil || errB != nil {
			t.mdEnabled = false
			return
		}
		t.mdRendererA = rendererA
		t.mdRendererB = rendererB
		t.mdWrapWidth = width
	}
}

func (t *TUI) PrePrintMessages() string {
	var sb strings.Builder

	for index, message := range t.messages.Messages {
		if index > 0 {
			sb.WriteString(t.DottedLine(t.width) + "\n")
		}

		// Message Header //
		var label lipgloss.Style
		header := message.Source.String()
		switch message.Source {
		case modelpkg.ScSystem:
			label = t.styles.labelSystem
		case modelpkg.ScHuman:
			label = t.styles.labelHuman
		case modelpkg.ScAssistant:
			label = t.styles.labelAssistant
			header += " [" + message.WrittenBy + "]"
		}
		if !message.CreatedAt.IsZero() {
			header += " - " + message.CreatedAt.Format("15:04:05")
		}
		sb.WriteString(label.Render(header) + "\n")
		// [End] Message Header //

		// Message Body //
		body := message.Text
		if t.mdEnabled && t.mdRendererA != nil {
			rendered, err := t.mdRendererA.Render(body)
			if err == nil {
				body = strings.TrimSpace(rendered)
			}
		}
		sb.WriteString(t.styles.body.Render(body) + "\n")
		// [End] Message Body //
	}

	return sb.String()
}
