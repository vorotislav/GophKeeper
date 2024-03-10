package media

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
	inputFilePathField
	inputNoteField
	inputExpDateField
	inputCancelButton
	inputSubmitButton
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

type MediaInputModel struct {
	focus         int
	inputs        []textinput.Model
	inputStatus   string
	viewportLeft  viewport.Model
	viewportRight viewport.Model
	media         models.Media
}

func InitialInputModel() MediaInputModel {
	mim := MediaInputModel{}
	const inputSize = 4

	mim.inputs = make([]textinput.Model, inputSize)

	var t textinput.Model

	for i := range mim.inputs {
		t = textinput.New()
		t.CharLimit = 50
		switch i {
		case int(inputTitleField):
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputFilePathField):
			t.Placeholder = "File path"
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

		mim.inputs[i] = t
	}

	return mim
}

func (mim *MediaInputModel) Media() models.Media {
	return mim.media
}

func (mim *MediaInputModel) SetMedia(m models.Media) {
	mim.resetInputs()
	mim.media = m
}

func (mim *MediaInputModel) Init() tea.Cmd { return nil }

func (mim *MediaInputModel) View() string {
	var b strings.Builder
	if mim.media.ID == 0 {
		b.WriteString(common.InputTitleStyle.Render("New media") + "\n\n\n")
	} else {
		b.WriteString(common.InputTitleStyle.Render("Edit media") + "\n\n\n")

	}

	b.WriteString(common.FocusedStyle.Render("Title") + "\n")
	b.WriteString(mim.inputs[inputTitleField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("File path") + "\n")
	b.WriteString(mim.inputs[inputFilePathField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Note") + "\n")
	b.WriteString(mim.inputs[inputNoteField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("ExpiredAt") + "\n")
	b.WriteString(mim.inputs[inputExpDateField].View())
	b.WriteString("\n")
	b.WriteString("\n")

	cancelButton := &common.BlurredStyle
	if mim.focus == len(mim.inputs) {
		cancelButton = &common.FocusedStyle
	}
	submitButton := &common.BlurredStyle
	if mim.focus == len(mim.inputs)+1 {
		submitButton = &common.FocusedStyle
	}

	_, err := fmt.Fprintf(
		&b,
		"\n\n%s  %s\n\n%s",
		cancelButton.Render("[ Cancel ]"),
		submitButton.Render("[ Submit ]"),
		common.ErrStyle(mim.inputStatus),
	)
	if err != nil {
		fmt.Printf("Error formatting input string: %v\n", err)
		os.Exit(1)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		common.InputStyle(b.String()))
}

func (mim *MediaInputModel) Update(msg tea.Msg) (MediaInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mim.viewportLeft = viewport.New(msg.Width/3, msg.Height)
		mim.viewportRight = viewport.New(msg.Width/3, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, InputKeymap.Back):
			mim.inputStatus = ""

			return *mim, func() tea.Msg { return InputState{is: CancelState} }
		case key.Matches(msg, InputKeymap.Next):
			mim.focus++
			if mim.focus > int(inputSubmitButton) {
				mim.focus = int(inputTitleField)
			}
		case key.Matches(msg, InputKeymap.Prev):
			mim.focus--
			if mim.focus < int(inputTitleField) {
				mim.focus = int(inputSubmitButton)
			}
		case key.Matches(msg, InputKeymap.Enter):
			switch inputFields(mim.focus) {
			case inputTitleField, inputFilePathField, inputNoteField, inputExpDateField:
				mim.focus++
			case inputCancelButton:
				mim.resetInputs()
				return *mim, func() tea.Msg { return InputState{is: CancelState} }
			case inputSubmitButton:
				err := mim.validateInputs()
				if err != nil {
					mim.resetInputs()
					mim.focus = 0
					mim.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}

				mim.resetInputs()
				return *mim, func() tea.Msg { return InputState{is: SubmitState} }
			}
		}
	}

	cmds = append(cmds, mim.updateInputs()...)
	for i := 0; i < len(mim.inputs); i++ {
		newModel, cmd := mim.inputs[i].Update(msg)
		mim.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return *mim, tea.Batch(cmds...)
}

func (mim *MediaInputModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(mim.inputs))
	for i := 0; i <= len(mim.inputs)-1; i++ {
		if i == mim.focus {
			// Set focused state
			cmds[i] = mim.inputs[i].Focus()
			mim.inputs[i].PromptStyle = common.FocusedStyle
			mim.inputs[i].TextStyle = common.FocusedStyle
			continue
		}
		// Remove focused state
		mim.inputs[i].Blur()
		mim.inputs[i].PromptStyle = common.NoStyle
		mim.inputs[i].TextStyle = common.NoStyle
	}
	return cmds
}

func (mim *MediaInputModel) resetInputs() {
	for i := 0; i < len(mim.inputs); i++ {
		mim.inputs[i].Reset()
	}
	mim.focus = 0
	mim.inputStatus = ""
}

func (mim *MediaInputModel) validateInputs() error {
	title := mim.inputs[inputTitleField].Value()
	filepath := mim.inputs[inputFilePathField].Value()
	note := mim.inputs[inputNoteField].Value()
	expDate := mim.inputs[inputExpDateField].Value()

	if filepath == "" {
		return fmt.Errorf("empty file path")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer file.Close()
	body, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("open read: %w", err)
	}

	if mim.media.ID == 0 {
		if title == "" || filepath == "" {
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

			mim.media.ExpiredAt = ts
		}

		mim.media.Title = title
		mim.media.MediaType = file.Name()
		mim.media.Body = body
		mim.media.Note = note

	} else {
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

			mim.media.ExpiredAt = ts
		}

		if title != "" {
			mim.media.Title = title
		}

		mim.media.MediaType = file.Name()
		mim.media.Body = body

		if note != "" {
			mim.media.Note = note
		}
	}

	return nil
}

/*

type Media struct {
	ID        int
	Title     string
	Body      []byte
	MediaType string
	Note      string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
*/
