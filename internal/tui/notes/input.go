package notes

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
	inputTextField
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

type NoteInputModel struct {
	focus         int
	inputs        []textinput.Model
	inputStatus   string
	viewportLeft  viewport.Model
	viewportRight viewport.Model
	note          models.Note
}

func (nim *NoteInputModel) SetNote(n models.Note) {
	nim.resetInputs()
	nim.note = n
	if nim.note.ID > 0 {
		nim.inputs[inputTitleField].SetValue(nim.note.Title)
		nim.inputs[inputTextField].SetValue(nim.note.Text)
		nim.inputs[inputExpDateField].SetValue(nim.note.ExpiredAt.Format(inputTimeFormLong))
	}
}

func (nim *NoteInputModel) Note() models.Note {
	return nim.note
}

func InitialInputModel() NoteInputModel {
	pim := NoteInputModel{}
	const inputSize = 3

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
		case int(inputTextField):
			t.Placeholder = "Text"
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

func (nim *NoteInputModel) Init() tea.Cmd { return nil }

func (nim *NoteInputModel) View() string {
	var b strings.Builder
	if nim.note.ID == 0 {
		b.WriteString(common.InputTitleStyle.Render("New note") + "\n\n\n")
	} else {
		b.WriteString(common.InputTitleStyle.Render("Edit note") + "\n\n\n")

	}

	b.WriteString(common.FocusedStyle.Render("Title") + "\n")
	b.WriteString(nim.inputs[inputTitleField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Text") + "\n")
	b.WriteString(nim.inputs[inputTextField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("ExpiredAt") + "\n")
	b.WriteString(nim.inputs[inputExpDateField].View())
	b.WriteString("\n")
	b.WriteString("\n")

	cancelButton := &common.BlurredStyle
	if nim.focus == len(nim.inputs) {
		cancelButton = &common.FocusedStyle
	}
	submitButton := &common.BlurredStyle
	if nim.focus == len(nim.inputs)+1 {
		submitButton = &common.FocusedStyle
	}

	_, err := fmt.Fprintf(
		&b,
		"\n\n%s  %s\n\n%s",
		cancelButton.Render("[ Cancel ]"),
		submitButton.Render("[ Submit ]"),
		common.ErrStyle(nim.inputStatus),
	)
	if err != nil {
		fmt.Printf("Error formatting input string: %v\n", err)
		os.Exit(1)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		common.InputStyle(b.String()))
}

func (nim *NoteInputModel) Update(msg tea.Msg) (NoteInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		nim.viewportLeft = viewport.New(msg.Width/3, msg.Height)
		nim.viewportRight = viewport.New(msg.Width/3, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, InputKeymap.Back):
			if nim.inputStatus != "" {
				nim.inputStatus = ""
				return *nim, func() tea.Msg { return InputState{is: CancelState} }
			}

			return *nim, func() tea.Msg { return InputState{is: CancelState} }
		case key.Matches(msg, InputKeymap.Next):
			nim.focus++
			if nim.focus > int(inputSubmitButton) {
				nim.focus = int(inputTitleField)
			}
		case key.Matches(msg, InputKeymap.Prev):
			nim.focus--
			if nim.focus < int(inputTitleField) {
				nim.focus = int(inputSubmitButton)
			}
		case key.Matches(msg, InputKeymap.Enter):
			switch inputFields(nim.focus) {
			case inputTitleField, inputTextField, inputExpDateField:
				nim.focus++
			case inputCancelButton:
				nim.resetInputs()
				return *nim, func() tea.Msg { return InputState{is: CancelState} }
			case inputSubmitButton:
				err := nim.validateInputs()
				if err != nil {
					nim.resetInputs()
					nim.focus = 0
					nim.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}
				nim.resetInputs()

				return *nim, func() tea.Msg { return InputState{is: SubmitState} }
			}
		}
	}

	cmds = append(cmds, nim.updateInputs()...)
	for i := 0; i < len(nim.inputs); i++ {
		newModel, cmd := nim.inputs[i].Update(msg)
		nim.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return *nim, tea.Batch(cmds...)
}

func (nim *NoteInputModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(nim.inputs))
	for i := 0; i <= len(nim.inputs)-1; i++ {
		if i == nim.focus {
			// Set focused state
			cmds[i] = nim.inputs[i].Focus()
			nim.inputs[i].PromptStyle = common.FocusedStyle
			nim.inputs[i].TextStyle = common.FocusedStyle
			continue
		}
		// Remove focused state
		nim.inputs[i].Blur()
		nim.inputs[i].PromptStyle = common.NoStyle
		nim.inputs[i].TextStyle = common.NoStyle
	}
	return cmds
}

func (nim *NoteInputModel) resetInputs() {
	for i := 0; i < len(nim.inputs); i++ {
		nim.inputs[i].Reset()
	}

	nim.focus = 0
	nim.inputStatus = ""
}

func (nim *NoteInputModel) validateInputs() error {
	if nim.note.ID == 0 {
		title := nim.inputs[inputTitleField].Value()
		text := nim.inputs[inputTextField].Value()
		expDate := nim.inputs[inputExpDateField].Value()

		if title == "" || text == "" {
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

			nim.note.ExpiredAt = ts
		}

		nim.note.Title = title
		nim.note.Text = text

	} else {
		title := nim.inputs[inputTitleField].Value()
		text := nim.inputs[inputTextField].Value()
		expDate := nim.inputs[inputExpDateField].Value()

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

			nim.note.ExpiredAt = ts
		}

		if title != "" {
			nim.note.Title = title
		}

		if text != "" {
			nim.note.Text = text
		}
	}

	return nil
}
