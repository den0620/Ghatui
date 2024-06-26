package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"sync"

    "github.com/gorilla/websocket"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type state int
const (
    stateLogin state = iota
    stateInvite
    stateInviteReceived
    stateChat
)
var (
    conn          *websocket.Conn
    mate          string
    message       = make(chan Message, 1)
    chatting      = make(chan bool, 1)
)
type Message struct {
	Type string
	Data interface{}
}

type model struct {
    usernameInput textinput.Model
    passwordInput textinput.Model
    chatInput     textinput.Model
    viewArea      viewport.Model
    messages      []string
    onlineUsers   string
    state         state
    focusIndex    int
    quitting      bool
    err           string
    inviteFrom    string
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
			    send("Register", []string{m.usernameInput.Value(), m.passwordInput.Value()})
			    mess := <-message
			    if mess.Type == "UserExist" {
			    	send("Login", []string{m.usernameInput.Value(), m.passwordInput.Value()})
			    	mess = <-message
			    }
			    switch mess.Type {
			    case "AlreadyOnline":
					m.err = "User already online"
					return m, nil
				case "WrongPassword":
					m.err = "Wrong password"
					return m, nil
				case "Logged", "Registered":
				    m.state = stateInvite
				    m.chatInput.Focus()
				}
		    } else if m.state == stateInvite {
				send("Invite", m.chatInput.Value())
				m.chatInput.SetValue("")
			} else if m.state == stateInviteReceived {
				if m.chatInput.Value() == "y" || m.chatInput.Value() == "Y" {
					send("InviteCrush", m.inviteFrom)
					mate = m.inviteFrom
					m.state = stateChat
				} else {
					send("InvitePass", m.inviteFrom)
					m.state = stateInvite
				}
				m.chatInput.SetValue("")
		    } else if m.state == stateChat {
				m.messages = append(m.messages, m.chatInput.Value())
				send("Chat", m.chatInput.Value())
				m.chatInput.SetValue("")
		    }
		case tea.KeyCtrlC, tea.KeyEsc:
		        return m, tea.Quit
		}
	case Message:
	    switch msg.Type {
		case "InviteRequest":
	    	m.state = stateInviteReceived
	    	m.inviteFrom = msg.Data.(string)
	    	m.err = fmt.Sprintf("Invite from: %s (y/n)", msg.Data)
		case "InviteAccept":
	    	mate = msg.Data.(string)
	    	m.state = stateChat
		case "InviteRefuse":
	    	m.err = fmt.Sprintf("Invite refused by: %s", msg.Data)
		case "Chat":
	    	m.messages = append(m.messages, fmt.Sprintf("%s: %s", mate, msg.Data))
		case "MateClosed":
	    	m.err = "Your mate has left the chat."	
	       	m.state = stateInvite
		case "OnlineUsers":
		   	m.onlineUsers = msg.Data.(string)
		}
    }
    switch m.state {
    case stateLogin:
		m.usernameInput, _ = m.usernameInput.Update(msg)
		m.passwordInput, _ = m.passwordInput.Update(msg)
    case stateInvite, stateChat, stateInviteReceived:
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
    case stateInvite:
		b.WriteString("Online Users:\n\n")
		b.WriteString(m.onlineUsers + "\n")
		b.WriteString(fmt.Sprintf("Invite user to chat: %s\n", m.chatInput.View()))
		if m.err != "" {
			b.WriteString(fmt.Sprintf("\nNotice: %s\n", m.err))
		}
	case stateInviteReceived:
		b.WriteString(fmt.Sprintf("Invite from %s. Accept? (y/n): %s\n", m.inviteFrom, m.chatInput.View()))
		if m.err != "" {
			b.WriteString(fmt.Sprintf("\nNotice: %s\n", m.err))
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


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./gochat [ip:port]")
		return
	}
	var wg sync.WaitGroup

	var err error
	var addr = "ws://" + os.Args[1] + "/ws"
	conn, _, err = websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}

	go read(conn)
	go ping(conn)

    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if err := p.Start(); err != nil {
    	fmt.Printf("Err: %v", err)
    	os.Exit(1)
    }

    go func() { // read
		for msg := range message {
			p.Send(msg)
			fmt.Print(msg)
		}
	}()

    wg.Wait()
}

func read(conn *websocket.Conn) {
	var msg Message
	for {
		if conn.ReadJSON(&msg) != nil {
			fmt.Print("Error reading from server: ", msg, "\n")
			conn.Close()
		}
		message <- msg
	}
}

func send(Type string, Data ...interface{}) {
	var err error
	if len(Data) == 0 {
		err = conn.WriteJSON(Message{Type: Type})
	} else {
		err = conn.WriteJSON(Message{Type: Type, Data: Data[0]})
	}
	if err != nil {
		fmt.Println("Error sending to server:", err)
	}
}

func ping(conn *websocket.Conn) {
	for {
		if conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)) != nil {
			fmt.Print("Connection timeout\n")
			os.Exit(1)
		}
		time.Sleep(3 * time.Minute)
	}
}