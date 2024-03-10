package start

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/common"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

type sessionClient interface {
	Login(user models.User) (models.Session, error)
	Register(user models.User) (models.Session, error)
}

type inputFields int

const (
	inputLoginField inputFields = iota
	inputPasswordField
	inputCancelButton
	inputLoginButton
	inputRegisterButton
)

type keymap struct {
	Next  key.Binding
	Prev  key.Binding
	Enter key.Binding
	Back  key.Binding
}

var Keymap = keymap{
	Next: key.NewBinding(
		key.WithKeys("tab"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

type StartModelState int

const (
	UnknownState StartModelState = iota
	BackState
	SessionState
)

type StartModel struct {
	focus         int
	inputs        []textinput.Model
	timer         timer.Model
	inputStatus   string
	State         StartModelState
	client        sessionClient
	viewportLeft  viewport.Model
	viewportRight viewport.Model
	session       models.Session
}

func InitialModel(cl sessionClient) StartModel {
	sm := StartModel{
		State:  UnknownState,
		client: cl,
	}

	sm.inputs = make([]textinput.Model, 2)
	var t textinput.Model
	for i := range sm.inputs {
		t = textinput.New()
		t.CharLimit = 30
		switch i {
		case 0:
			t.Placeholder = "Login"
			t.Focus()
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case 1:
			t.Placeholder = "Password"
			t.CharLimit = 20
			t.EchoMode = textinput.EchoPassword
		}
		sm.inputs[i] = t
	}

	return sm
}

func (sm StartModel) Init() tea.Cmd { return nil }

func (sm StartModel) View() string {
	var b strings.Builder
	b.WriteString(common.InputTitleStyle.Render("Login in GophKeeper") + "\n\n\n")
	for i := range sm.inputs {
		b.WriteString(sm.inputs[i].View() + "\n")
		if i < len(sm.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	cancelButton := &common.BlurredStyle
	if sm.focus == len(sm.inputs) {
		cancelButton = &common.FocusedStyle
	}
	loginButton := &common.BlurredStyle
	if sm.focus == len(sm.inputs)+1 {
		loginButton = &common.FocusedStyle
	}
	registerButton := &common.BlurredStyle
	if sm.focus == len(sm.inputs)+2 {
		registerButton = &common.FocusedStyle
	}
	_, err := fmt.Fprintf(
		&b,
		"\n\n%s  %s  %s\n\n%s",
		cancelButton.Render("[ Cancel ]"),
		loginButton.Render("[ Login ]"),
		registerButton.Render("[ Register ]"),
		common.ErrStyle(sm.inputStatus),
	)
	if err != nil {
		fmt.Printf("Error formatting input string: %v\n", err)
		os.Exit(1)
	}

	//return common.InputStyle(b.String())
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sm.viewportLeft.View(),
		common.InputStyle(b.String()),
		sm.viewportRight.View())
}

func (sm StartModel) Update(msg tea.Msg) (StartModel, tea.Cmd) {
	//var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sm.viewportLeft = viewport.New(msg.Width/3, msg.Height)
		sm.viewportRight = viewport.New(msg.Width/3, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Back):
			if sm.inputStatus != "" {
				sm.inputStatus = ""
				return sm, nil
			}
			sm.State = BackState
			return sm, nil
		case key.Matches(msg, Keymap.Next):
			sm.focus++
			if sm.focus > int(inputRegisterButton) {
				sm.focus = int(inputLoginField)
			}
		case key.Matches(msg, Keymap.Prev):
			sm.focus--
			if sm.focus < int(inputLoginField) {
				sm.focus = int(inputRegisterButton)
			}
		case key.Matches(msg, Keymap.Enter):
			switch inputFields(sm.focus) {
			case inputLoginField, inputPasswordField:
				sm.focus++
			case inputCancelButton:
				sm.State = BackState
				return sm, nil
			case inputLoginButton:
				err := sm.login()
				if err != nil {
					sm.inputs[inputLoginField].Reset()
					sm.inputs[inputPasswordField].Reset()
					sm.focus = 0
					sm.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}
				sm.State = SessionState
				return sm, nil
			case inputRegisterButton:
				err := sm.register()
				if err != nil {
					sm.inputs[inputLoginField].Reset()
					sm.inputs[inputPasswordField].Reset()
					sm.focus = 0
					sm.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}
				sm.State = SessionState
				return sm, nil
			}
		}
	}

	cmds = append(cmds, sm.updateInputs()...)
	for i := 0; i < len(sm.inputs); i++ {
		newModel, cmd := sm.inputs[i].Update(msg)
		sm.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return sm, tea.Batch(cmds...)
}

func (sm StartModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(sm.inputs))
	for i := 0; i <= len(sm.inputs)-1; i++ {
		if i == sm.focus {
			// Set focused state
			cmds[i] = sm.inputs[i].Focus()
			sm.inputs[i].PromptStyle = common.FocusedStyle
			sm.inputs[i].TextStyle = common.FocusedStyle
			continue
		}
		// Remove focused state
		sm.inputs[i].Blur()
		sm.inputs[i].PromptStyle = common.NoStyle
		sm.inputs[i].TextStyle = common.NoStyle
	}
	return cmds
}

func (sm StartModel) resetInputs() {
	sm.inputs[inputLoginField].Reset()
	sm.inputs[inputPasswordField].Reset()
	sm.focus = 0
	sm.inputStatus = ""
}

func (sm StartModel) login() error {
	login := sm.inputs[inputLoginField].Value()
	password := sm.inputs[inputPasswordField].Value()
	if login == "" && password == "" {
		return fmt.Errorf("empty fields")
	}

	session, err := sm.client.Login(models.User{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return fmt.Errorf("cannot login user (%s): %w", login, err)
	}

	sm.session = session
	return nil
}

func (sm StartModel) register() error {
	login := sm.inputs[inputLoginField].Value()
	password := sm.inputs[inputPasswordField].Value()
	if login == "" && password == "" {
		return fmt.Errorf("empty fields")
	}

	session, err := sm.client.Register(models.User{
		ID:       0,
		Login:    login,
		Password: password,
	})

	sm.session = session
	if err != nil {
		return fmt.Errorf("cannot register new user (%s): %w", login, err)
	}

	return nil
}
