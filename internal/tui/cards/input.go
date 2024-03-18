package cards

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
	"strconv"
	"strings"
)

type inputFields int

const (
	inputTitleField inputFields = iota
	inputNumberField
	inputCVCField
	inputExpField
	inputCancelButton
	inputSubmitButton
)

type inputState int

const (
	InvisibleState inputState = iota
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

type CardInputModel struct {
	focus         int
	inputs        []textinput.Model
	inputStatus   string
	viewportLeft  viewport.Model
	viewportRight viewport.Model
	card          models.Card
}

func InitialInputModel() CardInputModel {
	cim := CardInputModel{}
	const inputSize = 4

	cim.inputs = make([]textinput.Model, inputSize)
	var t textinput.Model

	for i := range cim.inputs {
		t = textinput.New()
		t.CharLimit = 50
		switch i {
		case int(inputTitleField):
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
		case int(inputNumberField):
			t.Placeholder = "4505 **** **** 1234"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
			t.CharLimit = 20
			t.Width = 30
			t.Validate = ccnValidator
		case int(inputCVCField):
			t.Placeholder = "XXX"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
			t.Validate = cvvValidator
			t.CharLimit = 3
		case int(inputExpField):
			t.Placeholder = "MM/YY"
			t.PromptStyle = common.FocusedStyle
			t.TextStyle = common.FocusedStyle
			t.Validate = expValidator
			t.CharLimit = 5
		}

		cim.inputs[i] = t
	}

	return cim
}

func (cim *CardInputModel) Update(msg tea.Msg) (CardInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cim.viewportLeft = viewport.New(msg.Width/3, msg.Height)
		cim.viewportRight = viewport.New(msg.Width/3, msg.Height)
	case tea.KeyMsg:
		switch {
		case msg.String() == "1" || msg.String() == "2" || msg.String() == "3" || msg.String() == "4":
		case key.Matches(msg, InputKeymap.Back):
			cim.inputStatus = ""
			return *cim, func() tea.Msg { return InputState{is: CancelState} }
		case key.Matches(msg, InputKeymap.Next):
			cim.focus++
			if cim.focus > int(inputSubmitButton) {
				cim.focus = int(inputTitleField)
			}
		case key.Matches(msg, InputKeymap.Prev):
			cim.focus--
			if cim.focus < int(inputTitleField) {
				cim.focus = int(inputSubmitButton)
			}
		case key.Matches(msg, InputKeymap.Enter):
			switch inputFields(cim.focus) {
			case inputTitleField, inputNumberField, inputCVCField, inputExpField:
				cim.focus++
			case inputCancelButton:
				cim.resetInputs()
				return *cim, func() tea.Msg { return InputState{is: CancelState} }
			case inputSubmitButton:
				err := cim.validateInputs()
				if err != nil {
					cim.resetInputs()
					cim.focus = 0
					cim.inputStatus = fmt.Sprintf("Error: %v", err)
					break
				}

				cim.resetInputs()

				return *cim, func() tea.Msg { return InputState{is: SubmitState} }
			}
		}
	}

	cmds = append(cmds, cim.updateInputs()...)
	for i := 0; i < len(cim.inputs); i++ {
		newModel, cmd := cim.inputs[i].Update(msg)
		cim.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return *cim, tea.Batch(cmds...)
}

func (cim *CardInputModel) View() string {
	var b strings.Builder
	if cim.card.ID == 0 {
		b.WriteString(common.InputTitleStyle.Render("New card") + "\n\n\n")
	} else {
		b.WriteString(common.InputTitleStyle.Render("Edit card") + "\n\n\n")
	}

	b.WriteString(common.FocusedStyle.Render("Title") + "\n")
	b.WriteString(cim.inputs[inputTitleField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("Card Number") + "\n")
	b.WriteString(cim.inputs[inputNumberField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("CVC") + "\n")
	b.WriteString(cim.inputs[inputCVCField].View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(common.FocusedStyle.Render("EXP") + "\n")
	b.WriteString(cim.inputs[inputExpField].View())

	cancelButton := &common.BlurredStyle
	if cim.focus == len(cim.inputs) {
		cancelButton = &common.FocusedStyle
	}
	submitButton := &common.BlurredStyle
	if cim.focus == len(cim.inputs)+1 {
		submitButton = &common.FocusedStyle
	}

	_, err := fmt.Fprintf(
		&b,
		"\n\n%s  %s\n\n%s",
		cancelButton.Render("[ Cancel ]"),
		submitButton.Render("[ Submit ]"),
		common.ErrStyle(cim.inputStatus),
	)
	if err != nil {
		fmt.Printf("Error formatting input string: %v\n", err)
		os.Exit(1)
	}

	//return common.InputStyle(b.String())
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		cim.viewportLeft.View(),
		common.InputStyle(b.String()),
		cim.viewportRight.View())
}

func (cim *CardInputModel) SetCard(c models.Card) {
	cim.card = c
	if cim.card.ID > 0 {
		cim.inputs[inputTitleField].SetValue(cim.card.Name)
		cim.inputs[inputNumberField].SetValue(cim.card.Number)
		cim.inputs[inputCVCField].SetValue(cim.card.CVC)
		cim.inputs[inputExpField].SetValue(fmt.Sprintf("%d/%d", cim.card.ExpMonth, cim.card.ExpYear))
	}
}

func (cim *CardInputModel) Card() models.Card {
	return cim.card
}

func (cim *CardInputModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(cim.inputs))
	for i := 0; i <= len(cim.inputs)-1; i++ {
		if i == cim.focus {
			// Set focused state
			cmds[i] = cim.inputs[i].Focus()
			cim.inputs[i].PromptStyle = common.FocusedStyle
			cim.inputs[i].TextStyle = common.FocusedStyle
			continue
		}
		// Remove focused state
		cim.inputs[i].Blur()
		cim.inputs[i].PromptStyle = common.NoStyle
		cim.inputs[i].TextStyle = common.NoStyle
	}
	return cmds
}

func (cim *CardInputModel) resetInputs() {
	for i := 0; i < len(cim.inputs); i++ {
		cim.inputs[i].Reset()
	}
	cim.focus = 0
	cim.inputStatus = ""
}

func (cim *CardInputModel) validateInputs() error {
	if cim.card.ID == 0 {
		title := cim.inputs[inputTitleField].Value()
		number := cim.inputs[inputNumberField].Value()
		cvc := cim.inputs[inputCVCField].Value()
		expDate := cim.inputs[inputExpField].Value()

		if title == "" || number == "" || cvc == "" || expDate == "" {
			return fmt.Errorf("empty fields")
		}

		exps := strings.Split(expDate, "/")
		if len(exps) != 2 {
			return fmt.Errorf("exp date size")
		}

		expMonth, err := strconv.Atoi(exps[0])
		if err != nil {
			return fmt.Errorf("exp month invalid type")
		}

		expYear, err := strconv.Atoi(exps[1])
		if err != nil {
			return fmt.Errorf("exp year invalid type")
		}

		cim.card.Name = title
		cim.card.Number = number
		cim.card.CVC = cvc
		cim.card.ExpMonth = expMonth
		cim.card.ExpYear = expYear
	} else {
		title := cim.inputs[inputTitleField].Value()
		number := cim.inputs[inputNumberField].Value()
		cvc := cim.inputs[inputCVCField].Value()
		expDate := cim.inputs[inputExpField].Value()

		exps := strings.Split(expDate, "/")
		if len(exps) != 2 {
			return fmt.Errorf("exp date size")
		}

		if title != "" {
			cim.card.Name = title
		}

		if number != "" {
			cim.card.Number = number
		}

		if cvc != "" {
			cim.card.CVC = cvc
		}

		if expDate != "" {
			expMonth, err := strconv.Atoi(exps[0])
			if err != nil {
				return fmt.Errorf("exp month invalid type")
			}

			expYear, err := strconv.Atoi(exps[1])
			if err != nil {
				return fmt.Errorf("exp year invalid type")
			}
			cim.card.ExpMonth = expMonth
			cim.card.ExpYear = expYear
		}
	}

	return nil
}

// Validator functions to ensure valid input
func ccnValidator(s string) error {
	// Credit Card Number should a string less than 20 digits
	// It should include 16 integers and 3 spaces
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}

	if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	// The last digit should be a number unless it is a multiple of 4 in which
	// case it should be a space
	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}

	// The remaining digits should be integers
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func expValidator(s string) error {
	// The 3 character should be a slash (/)
	// The rest should be numbers
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	// There should be only one slash and it should be in the 2nd index (3rd character)
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}

	return nil
}

func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	// Since the input will already ensure that the CVV is a string of length 3,
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}

/*
type Card struct {
	ID        int
	Name      string
	Number    string
	CVC       string
	ExpMonth  int
	ExpYear   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
*/
