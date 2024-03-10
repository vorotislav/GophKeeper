package passwords

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/common"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"time"
)

type inputFields int

const (
	inputTitleField inputFields = iota
	inputLoginField
	inputPasswordField
	inputURLField
	inputNoteField
	inputExpDateField
	inputCancelButton
	inputSubmitButton
)

const (
	inputTimeFormShort = "2006-01-02"
	inputTimeFormLong  = "2006-01-02 15:04:05"
)

type inputState int

const (
	InvisibleState inputState = iota
	ViewState
	CancelState
	SubmitState
)

type InputState struct {
	is inputState
}

type inputKeymap struct {
	Next  key.Binding
	Prev  key.Binding
	Enter key.Binding
	Back  key.Binding
}

var InputKeymap = inputKeymap{
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

type PasswordInputModel struct {
	focus         int
	inputs        []textinput.Model
	inputStatus   string
	viewportLeft  viewport.Model
	viewportRight viewport.Model
	pass          models.Password
}

func (pim *PasswordInputModel) SetPass(p models.Password) {
	pim.pass = p
	if pim.pass.ID > 0 {
		pim.inputs[inputTitleField].SetValue(pim.pass.Title)
		pim.inputs[inputLoginField].SetValue(pim.pass.Login)
		pim.inputs[inputPasswordField].SetValue(pim.pass.Password)
		pim.inputs[inputURLField].SetValue(pim.pass.URL)
		pim.inputs[inputNoteField].SetValue(pim.pass.Note)
		pim.inputs[inputExpDateField].SetValue(pim.pass.ExpirationDate.Format(inputTimeFormLong))
	}
}

func (pim *PasswordInputModel) Pass() models.Password {
	return pim.pass
}

func InitialInputModel() PasswordInputModel {
	pim := PasswordInputModel{}
	const inputSize = 6

	pim.inputs = make([]textinput.Model, inputSize)
	var t textinput.Model

	for i := range pim.inputs {
		t = textinput.New()
		t.CharLimit = 50
		switch i {
		case int(inputTitleField):
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputLoginField):
			t.Placeholder = "Login"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputPasswordField):
			t.Placeholder = "Password"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputURLField):
			t.Placeholder = "URL"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputNoteField):
			t.Placeholder = "Note"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputExpDateField):
			t.Placeholder = "Expiration date"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		}

		pim.inputs[i] = t
	}

	return pim
}

func (pim *PasswordInputModel) Init() tea.Cmd { return nil }

func (pim *PasswordInputModel) View() string {
	var b strings.Builder
	if pim.pass.ID == 0 {
		b.WriteString(common.InputTitleStyle.Render("New password") + "\n\n\n")
	} else {
		b.WriteString(common.InputTitleStyle.Render("Edit password") + "\n\n\n")

	}

	b.WriteString(common.FocusedStyle.Render("Title") + "\n")
	b.WriteString(pim.inputs[inputTitleField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Login") + "\n")
	b.WriteString(pim.inputs[inputLoginField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Password") + "\n")
	b.WriteString(pim.inputs[inputPasswordField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("URL") + "\n")
	b.WriteString(pim.inputs[inputURLField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Note") + "\n")
	b.WriteString(pim.inputs[inputNoteField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("ExpiredAt") + "\n")
	b.WriteString(pim.inputs[inputExpDateField].View())
	b.WriteString("\n")
	b.WriteString("\n")

	cancelButton := &common.BlurredStyle
	if pim.focus == len(pim.inputs) {
		cancelButton = &common.FocusedStyle
	}
	submitButton := &common.BlurredStyle
	if pim.focus == len(pim.inputs)+1 {
		submitButton = &common.FocusedStyle
	}

	_, err := fmt.Fprintf(
		&b,
		"\n\n%s  %s\n\n%s",
		cancelButton.Render("[ Cancel ]"),
		submitButton.Render("[ Submit ]"),
		common.ErrStyle(pim.inputStatus),
	)
	if err != nil {
		fmt.Printf("Error formatting input string: %v\n", err)
		os.Exit(1)
	}

	//return common.InputStyle(b.String())
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		pim.viewportLeft.View(),
		common.InputStyle(b.String()),
		pim.viewportRight.View())
}

func (pim *PasswordInputModel) Update(msg tea.Msg) (PasswordInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pim.viewportLeft = viewport.New(msg.Width/3, msg.Height)
		pim.viewportRight = viewport.New(msg.Width/3, msg.Height)
	case tea.KeyMsg:
		switch {
		case msg.String() == "1" || msg.String() == "2" || msg.String() == "3" || msg.String() == "4":

		case key.Matches(msg, InputKeymap.Back):
			pim.inputStatus = ""

			return *pim, func() tea.Msg { return InputState{is: CancelState} }
		case key.Matches(msg, InputKeymap.Next):
			pim.focus++
			if pim.focus > int(inputSubmitButton) {
				pim.focus = int(inputTitleField)
			}
		case key.Matches(msg, InputKeymap.Prev):
			pim.focus--
			if pim.focus < int(inputTitleField) {
				pim.focus = int(inputSubmitButton)
			}
		case key.Matches(msg, InputKeymap.Enter):
			switch inputFields(pim.focus) {
			case inputTitleField, inputLoginField, inputPasswordField, inputURLField, inputNoteField, inputExpDateField:
				pim.focus++
			case inputCancelButton:
				pim.resetInputs()
				return *pim, func() tea.Msg { return InputState{is: CancelState} }
			case inputSubmitButton:
				err := pim.validateInputs()
				if err != nil {
					pim.resetInputs()
					pim.focus = 0
					pim.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}
				pim.resetInputs()
				return *pim, func() tea.Msg { return InputState{is: SubmitState} }
			}
		}
	}

	cmds = append(cmds, pim.updateInputs()...)
	for i := 0; i < len(pim.inputs); i++ {
		newModel, cmd := pim.inputs[i].Update(msg)
		pim.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return *pim, tea.Batch(cmds...)
}

func (pim *PasswordInputModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(pim.inputs))
	for i := 0; i <= len(pim.inputs)-1; i++ {
		if i == pim.focus {
			// Set focused state
			cmds[i] = pim.inputs[i].Focus()
			pim.inputs[i].PromptStyle = common.FocusedStyle
			pim.inputs[i].TextStyle = common.FocusedStyle
			continue
		}
		// Remove focused state
		pim.inputs[i].Blur()
		pim.inputs[i].PromptStyle = common.NoStyle
		pim.inputs[i].TextStyle = common.NoStyle
	}
	return cmds
}

func (pim *PasswordInputModel) resetInputs() {
	for i := 0; i < len(pim.inputs); i++ {
		pim.inputs[i].Reset()
	}
	pim.focus = 0
	pim.inputStatus = ""
}

func (pim *PasswordInputModel) validateInputs() error {
	if pim.pass.ID == 0 {
		title := pim.inputs[inputTitleField].Value()
		login := pim.inputs[inputLoginField].Value()
		password := pim.inputs[inputPasswordField].Value()
		url := pim.inputs[inputURLField].Value()
		note := pim.inputs[inputNoteField].Value()
		expDate := pim.inputs[inputExpDateField].Value()

		if title == "" || login == "" || password == "" {
			return fmt.Errorf("empty fields")
		}
		timeFormat := inputTimeFormLong

		if expDate != "" {
			if len(expDate) < len(inputTimeFormLong) {
				timeFormat = inputTimeFormShort
			}

			ts, err := time.ParseInLocation(timeFormat, expDate, time.UTC)
			if err != nil {
				return fmt.Errorf("parse expiration date: %w", err)
			}
			if ts.Before(time.Now()) {
				return fmt.Errorf("expiration date is in the past")
			}

			pim.pass.ExpirationDate = ts
		}

		pim.pass.Title = title
		pim.pass.Login = login
		pim.pass.Password = password
		pim.pass.URL = url
		pim.pass.Note = note

	} else {
		title := pim.inputs[inputTitleField].Value()
		login := pim.inputs[inputLoginField].Value()
		password := pim.inputs[inputPasswordField].Value()
		url := pim.inputs[inputURLField].Value()
		note := pim.inputs[inputNoteField].Value()
		expDate := pim.inputs[inputExpDateField].Value()

		timeFormat := inputTimeFormLong

		if expDate != "" {
			if len(expDate) < len(inputTimeFormLong) {
				timeFormat = inputTimeFormShort
			}

			ts, err := time.ParseInLocation(timeFormat, expDate, time.UTC)
			if err != nil {
				return fmt.Errorf("parse expiration date: %w", err)
			}
			if ts.Before(time.Now()) {
				return fmt.Errorf("expiration date is in the past")
			}

			pim.pass.ExpirationDate = ts
		}

		if title != "" {
			pim.pass.Title = title
		}

		if login != "" {
			pim.pass.Login = login
		}

		if password != "" {
			pim.pass.Password = password
		}

		if url != "" {
			pim.pass.URL = url
		}

		if note != "" {
			pim.pass.Note = note
		}
	}

	return nil
}
