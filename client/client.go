package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var CONNECTION *websocket.Conn

type User struct {
    Name     string
    Password string
    Conn     string //*websocket.Conn
    Chatting bool
}

type Keymap struct {
    help key.Binding
    login key.Binding
    register key.Binding
    invite key.Binding
    quit key.Binding
}
func (k Keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}
func (k Keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.login, k.register, k.invite}, // first column
		{k.help, k.quit},                // second column
	}
}

type model struct {
    keymap      Keymap
    help        help.Model
    user        User
    viewArea    viewport.Model
    messages    []string
    senderStyle lipgloss.Style
    inputArea   textarea.Model
    quitting    bool
}
func (m model) Init() tea.Cmd {
    return textarea.Blink
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var (
		iaCmd tea.Cmd
		vaCmd tea.Cmd
	)

	m.inputArea, iaCmd = m.inputArea.Update(msg)
	m.viewArea, vaCmd = m.viewArea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.inputArea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.inputArea.Value())
			m.viewArea.SetContent(strings.Join(m.messages, "\n"))
			m.inputArea.Reset()
			m.viewArea.GotoBottom()
		}
	}

	/* We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}*/

	return m, tea.Batch(iaCmd, vaCmd)
}
func (m model) View() string {
    return fmt.Sprintf(
		"%s\n\n%s",
		m.viewArea.View(),
		m.inputArea.View(),
	) + "\n\n" + m.help.ShortHelpView([]key.Binding{
        m.keymap.login,
        m.keymap.register,
        m.keymap.invite,
    })
}
func initialModel(init_user User) model {
	ia := textarea.New()
	ia.Placeholder = "your input here"
	ia.Prompt = "> "
	ia.CharLimit = 128
	ia.SetHeight(1)
	ia.Focus()

	ia.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ia.ShowLineNumbers = false
	ia.KeyMap.InsertNewline.SetEnabled(false)

	va := viewport.New(60,5)
	va.SetContent("still empty")

	return model{
		keymap: Keymap{
			help: key.NewBinding(
				key.WithKeys("h"),
				key.WithHelp("h", "help"),
			),
			login: key.NewBinding(
				key.WithKeys("l"),
				key.WithHelp("l", "login"),
			),
			register: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "register"),
			),
			invite: key.NewBinding(
				key.WithKeys("i"),
				key.WithHelp("i", "invite"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help:      help.New(),
    	user:      init_user,
    	viewArea:  va,
    	messages:  []string{},
    	senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
    	inputArea: ia,
    	quitting:  false,
	}
}
func initialUser() User {
	return User{
    	Name:     "Test",
    	Password: "123456",
    	Conn:     "penis",
    	Chatting: false,
	}
}


func main() {
    p := tea.NewProgram(initialModel(initialUser()),tea.WithAltScreen())
    if err := p.Start(); err != nil {
    	fmt.Printf("Err: %v", err)
    	os.Exit(1)
    }
}

