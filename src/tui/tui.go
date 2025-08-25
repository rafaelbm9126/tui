package tui

import (
	"context"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"main/src/bus"
	"main/src/event"
	"main/src/message"
)

type TUI struct {
	width       int
	height      int
	viewport    viewport.Model
	input       textinput.Model
	program     *tea.Program
	logger      *slog.Logger
	messages    *MessageList
	bus         *OptimizedBus
	mdEnabled   bool
	mdRenderer  *glamour.TermRenderer
	mdWrapWidth int // ancho con el que se construyó el renderer
	history     []string
	histIndex   int
	styles      struct {
		header         lipgloss.Style
		labelSystem    lipgloss.Style
		labelHuman     lipgloss.Style
		labelAssistant lipgloss.Style
		body           lipgloss.Style
		dots           lipgloss.Style
		help           lipgloss.Style
		inputBox       lipgloss.Style
	}
}

func NewTUI(bus *OptimizedBus, messages *MessageList, logger *slog.Logger) *TUI {
	vp := viewport.New(80, 20)
	vp.SetContent("") // No mostrar contenido hasta salir del estado de carga
	vp.Style = lipgloss.NewStyle().
		Margin(0, 5, 0, 5)

	ti := textinput.New()
	ti.Placeholder = "Escribe un mensaje o comando (/help)"
	ti.CharLimit = 2048
	ti.Width = 80
	ti.SetValue("")
	ti.Focus()

	logger.Info("Initializing TUI")

	t := &TUI{
		width:     80,
		height:    20,
		viewport:  vp,
		input:     ti,
		mdEnabled: true,

		logger: logger,

		messages: messages,

		bus: bus,
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
		BorderForeground(lipgloss.Color("#8a7dfc")).
		Padding(0, 1).
		Margin(0, 5, 0, 5)

	t.program = tea.NewProgram(t, tea.WithAltScreen())

	return t
}

func (t *TUI) Init() tea.Cmd {
	t.MDRenderer(t.viewport.Width)

	return tea.Batch(tea.EnterAltScreen)
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
				msg := MessageModel{Type: Human, Text: text}
				t.bus.Publish(EvtMessage, msg)
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
		t.viewport.Width = t.width - 14 // -14 por los bordes
		t.viewport.Height = viewportHeight
		t.input.Width = t.width - 14 // -14 por los bordes

		t.RenderBody()

	case Event:
		evt := msg.evt
		switch evt.Type {
		case EvtSystem:
			icmd, _ := evt.Data.(string)
			switch icmd {
			case "q", "quit":
				cmds = append(cmds, tea.Quit)
			default:
				panic("Command Unknown")
			}
		case EvtMessage:
			if msgData, ok := evt.Data.(MessageModel); ok {
				switch msgData.Type {
				case System:
					t.messages.AddMessageSystem(msgData.Text)
				case Human:
					t.messages.AddMessageHuman(msgData.Text)
				case Assistant:
					t.messages.AddMessageAssistant(msgData.Text, msgData.From)
				}
			}
		}
		t.RenderBody()
	}

	t.input, cmd = t.input.Update(msg)
	cmds = append(cmds, cmd)

	t.viewport, cmd = t.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return t, tea.Batch(cmds...)
}

func (t *TUI) View() string {
	header := t.styles.header.
		Margin(0, 5, 0, 5).
		Render(" Agentes TUI ") + "  " + t.styles.help.
		Margin(0, 5, 0, 5).
		Render("(Enter: enviar · /help · /md on|off · ↑↓ historial · PgUp/PgDn scroll · Esc/Ctrl+C salir)")

	footer := t.styles.help.
		Margin(0, 5, 0, 5).
		Render(" ESC/Ctrl+C: Salir • PgUp/PgDn: Desplazar • ↑/↓: Historial")

	input := t.styles.inputBox.Render(t.input.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		t.viewport.View(),
		footer,
		input,
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
	content := t.messages.PrePrintMessages(t)
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
