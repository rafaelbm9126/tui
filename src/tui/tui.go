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
	eventpkg "main/src/event"
	messagepkg "main/src/message"
)

type MessageList = messagepkg.MessageList
type MessageModel = messagepkg.MessageModel

type OptimizedBus = buspkg.OptimizedBus

type Event = eventpkg.Event

type Command = commandpkg.Command

const (
	LEFT_WIDTH_PERCENTAGE = 0.6 // 60% del ancho
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
	mdRenderer  *glamour.TermRenderer
	mdWrapWidth int // ancho con el que se construyó el renderer
	history     []string
	histIndex   int
	showAlert   bool
	textAlert   string
	styles      struct {
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
	bus *OptimizedBus,
	messages *MessageList,
	command *Command,
	logger *slog.Logger,
) *TUI {
	vp := viewport.New(80, 20)
	vp.SetContent("") // No mostrar contenido hasta salir del estado de carga
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8A7DFC")).
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			cmds = append(cmds, tea.Quit)

		case tea.KeyEnter:
			rawText := t.input.Value()
			text := strings.TrimSpace(rawText)
			if text != "" {
				msg := MessageModel{Type: messagepkg.Human, Text: text}
				t.bus.Publish(eventpkg.EvtMessage, msg)
				t.input.Reset()
				t.input.SetValue("")

				if len(t.history) == 0 || t.history[len(t.history)-1] != text {
					t.history = append(t.history, text)
				}
				t.histIndex = len(t.history)
			}

		case tea.KeyUp:
			// Navega por el historial hacia arriba
			if t.histIndex > 0 {
				t.histIndex--
				t.input.SetValue(t.history[t.histIndex])
				t.input.CursorEnd()
			}

		case tea.KeyDown:
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

	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

		inputHeight := 3 // borde + entrada + borde
		headerHeight := 1
		footerHeight := 1

		viewportHeight := t.height - inputHeight - headerHeight - footerHeight
		t.viewport.Width = int(float64(t.width) * LEFT_WIDTH_PERCENTAGE)
		t.viewport.Height = viewportHeight
		t.input.Width = int(float64(t.width)*LEFT_WIDTH_PERCENTAGE) - 11 // -5 ajuste para igualar al viewport

		t.RenderBody()

	case Event:
		evt := msg.Evt
		switch evt.Type {
		case eventpkg.EvtSystem:
			icmd, _ := evt.Data.(string)
			switch icmd {

			case "q", "quit":
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
				switch msgData.Type {
				case messagepkg.System:
					t.messages.AddMessageSystem(msgData.Text)
				case messagepkg.Human:
					t.messages.AddMessageHuman(msgData.Text)
					t.command.IsCommandThenRun(msgData.Text)
				case messagepkg.Assistant:
					t.messages.AddMessageAssistant(msgData.Text, msgData.From)
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
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8A7DFC")).
		Margin(1, 0, 0, 0)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left,
			HeaderViewTui(t), // header
			t.viewport.View(),
			FooterViewTui(t), // footer
			input,
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
		t.mdRenderer = nil
		t.mdWrapWidth = 0
		return
	}

	// Crea el renderer solo si ha cambiado el ancho o no existe
	if t.mdRenderer == nil || t.mdWrapWidth != width {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-14), // -14 por el margen izquierdo
		)
		if err != nil {
			t.mdEnabled = false
			return
		}
		t.mdRenderer = renderer
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
		header := message.Type.String()
		switch message.Type {
		case messagepkg.System:
			label = t.styles.labelSystem
		case messagepkg.Human:
			label = t.styles.labelHuman
		case messagepkg.Assistant:
			label = t.styles.labelAssistant
			header += " [" + message.From + "]"
		}
		if !message.Time.IsZero() {
			header += " - " + message.Time.Format("15:04:05")
		}
		sb.WriteString(label.Render(header) + "\n")
		// [End] Message Header //

		// Message Body //
		body := message.Text
		if t.mdEnabled && t.mdRenderer != nil {
			rendered, err := t.mdRenderer.Render(body)
			if err == nil {
				body = strings.TrimSpace(rendered)
			}
		}
		sb.WriteString(t.styles.body.Render(body) + "\n")
		// [End] Message Body //
	}

	return sb.String()
}
