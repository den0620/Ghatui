package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/gorilla/websocket"

    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
)

type state int
const (
    stateLogin state = iota
    stateChat
)

var CONNECTION *websocket.Conn
type User struct {
    Name     string
    Password string
    Conn     string //*websocket.Conn
    Chatting bool
}

type model struct {
    usernameInput textinput.Model
    passwordInput textinput.Model
    chatInput     textinput.Model
    viewArea      viewport.Model
    messages      []string
    state         state
    focusIndex    int
    quitting      bool
    err           string
}
func (m model) Init() tea.Cmd {
    return textarea.Blink
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
	case tea.KeyTab:
	    if m.state == stateLogin {
	        m.focusIndex++
	        if m.focusIndex > 1 {
		    m.focusIndex = 0
	        }
	        if m.focusIndex == 0 {
	            m.usernameInput.Focus()
		    m.passwordInput.Blur()
	        } else {
		    m.usernameInput.Blur()
		    m.passwordInput.Focus()
	        }
	    }
	case tea.KeyEnter:
	    if m.state == stateLogin {
		if m.usernameInput.Value() == validUsername && m.passwordInput.Value() == validPassword {
		    m.state = stateChat
		    m.chatInput.Focus()
		} else {
		    m.err = "Invalid username or password"
		    return m, nil
		}
	    } else if m.state == stateChat {
		m.messages = append(m.messages, m.chatInput.Value())
		m.chatInput.SetValue("")
	    }
	case tea.KeyCtrlC, tea.KeyEsc:
	        return m, tea.Quit
	}
    }

    switch m.state {
    case stateLogin:
	m.usernameInput, _ = m.usernameInput.Update(msg)
	m.passwordInput, _ = m.passwordInput.Update(msg)
    case stateChat:
	m.chatInput, _ = m.chatInput.Update(msg)
    }

    return m, nil
}
func (m model) View() string {
    var b strings.Builder

    switch m.state {
    case stateLogin:
        b.WriteString("Enter your credentials\n\n")
	b.WriteString(fmt.Sprintf("Username: %s\n", m.usernameInput.View()))
	b.WriteString(fmt.Sprintf("Password: %s\n", m.passwordInput.View()))
        if m.err != "" {
	    b.WriteString(fmt.Sprintf("\nError: %s\n", m.err))
        }
    case stateChat:
	b.WriteString("Chat Room\n\n")
	for _, msg := range m.messages {
	    b.WriteString(fmt.Sprintf("%s\n", msg))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Message: %s\n", m.chatInput.View()))
	b.WriteString("\nPress Ctrl+C to exit.\n")
    }

    return b.String() 
}
func initialModel() model {
    ui := textinput.New()
    ui.Placeholder = "Username"
    ui.Focus()	
    ui.CharLimit = 32
    ui.Width = 32
    ui.Focus()

    pi := textinput.New()
    pi.Placeholder = "Password"
    pi.EchoMode = textinput.EchoPassword
    pi.EchoCharacter = '*'
    pi.CharLimit = 32
    pi.Width = 32

    ci := textinput.New()
    ci.Placeholder = "Type a message"
    ci.Width = 50

    return model{
    	usernameInput: ui,
	passwordInput: pi,
	chatInput:     ci,
	state:         stateLogin,
    	quitting:      false,
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
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if err := p.Start(); err != nil {
    	fmt.Printf("Err: %v", err)
    	os.Exit(1)
    }
}

