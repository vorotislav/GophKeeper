package passwords

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/common"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/mistakenelf/teacup/help"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type passwordClient interface {
	CreatePassword(pass models.Password) error
	UpdatePassword(pass models.Password) error
	Passwords() ([]models.Password, error)
	DeletePassword(id int) error
}

type item models.Password

func (i item) GetTitle() string    { return i.Title }
func (i item) Description() string { return i.Note }
func (i item) FilterValue() string { return i.Title }
func (i item) GetPass() models.Password {
	return models.Password{
		ID:        i.ID,
		Title:     i.Title,
		Login:     i.Login,
		Password:  i.Password,
		URL:       i.URL,
		Note:      i.Note,
		ExpiredAt: i.ExpiredAt,
	}
}

type echoPasswordMode int

const (
	EchoNormal echoPasswordMode = iota
	EchoPassword
)

type passwordState int

const (
	showPassword passwordState = iota
	showInput
)

type keymap struct {
	Add      key.Binding
	Remove   key.Binding
	Change   key.Binding
	ShowPass key.Binding
	Quit     key.Binding
	Back     key.Binding
}

// Keymap reusable key mappings shared across models
var Keymap = keymap{
	Add: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "add"),
	),
	Remove: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "remove"),
	),
	Change: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "change"),
	),
	ShowPass: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "show pass"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "back"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

type PasswordModel struct {
	list         list.Model
	viewport     viewport.Model
	pc           passwordClient
	lastMessage  string
	im           PasswordInputModel
	echoPassword echoPasswordMode
	state        passwordState
	help         help.Model
}

func (pm *PasswordModel) Init() tea.Cmd { return nil }

func InitialModel(pc passwordClient) PasswordModel {
	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "Passwords"
	l.Styles.Title = common.TitleStyle
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	helpModel := help.New(
		false,
		true,
		"Help",
		help.TitleColor{},
		common.InactiveBoxBorderColor,
		[]help.Entry{
			{Key: "ctrl+c", Description: "Exit GophKeeper"},
			{Key: "j/up", Description: "Move up"},
			{Key: "k/down", Description: "Move down"},
			{Key: "1", Description: "Passwords view"},
			{Key: "2", Description: "Cards view"},
			{Key: "3", Description: "Notes view"},
			{Key: "4", Description: "Media view"},
			{Key: "+", Description: "Add password"},
			{Key: "-", Description: "Delete password"},
			{Key: "r", Description: "Change password"},
			{Key: "s", Description: "Show password in passwords view"},
		},
	)

	return PasswordModel{
		list:         l,
		pc:           pc,
		im:           InitialInputModel(),
		echoPassword: EchoPassword,
		state:        showPassword,
		help:         helpModel,
	}
}

func (pm *PasswordModel) View() string {
	switch pm.state {
	case showInput:
		return pm.im.View()
	default:
		pm.viewport.SetContent(pm.detailView())

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			pm.listView(),
			pm.viewport.View(),
			pm.help.View())
	}
}

func (pm *PasswordModel) Update(msg tea.Msg) (PasswordModel, tea.Cmd) {
	var cmd tea.Cmd

	switch pm.state {
	case showPassword:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			horizontal, vertical := common.ListStyle.GetFrameSize()
			paginatorHeight := lipgloss.Height(pm.list.Paginator.View())

			pm.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
			pm.viewport = viewport.New(msg.Width/2-10, msg.Height)
			pm.viewport.SetContent(pm.detailView())
			pm.help.SetSize(msg.Width/2, msg.Height)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Back):
				pm.state = showPassword
			case key.Matches(msg, Keymap.ShowPass):
				if pm.echoPassword == EchoPassword {
					pm.echoPassword = EchoNormal
				} else {
					pm.echoPassword = EchoPassword
				}
				return *pm, nil
			case key.Matches(msg, Keymap.Add):
				pm.state = showInput
				pm.im.SetPass(models.Password{})
			case key.Matches(msg, Keymap.Remove):
				pm.deletePassword(pm.list.SelectedItem().(item).GetPass().ID)
			case key.Matches(msg, Keymap.Change):
				pm.state = showInput
				currentItem := pm.list.SelectedItem()
				pm.im.SetPass(currentItem.(item).GetPass())
			}
		}

		pm.list, _ = pm.list.Update(msg)
	case showInput:
		pm.im, cmd = pm.im.Update(msg)
		if cmd != nil {
			is, ok := cmd().(InputState)
			if ok {
				switch is.is {
				case CancelState:
					pm.state = showPassword
				case SubmitState:
					pm.sendPassword(pm.im.pass)
					pm.state = showPassword
				}
			}
		}

		return *pm, cmd
	}

	return *pm, cmd
}

func (pm *PasswordModel) IsInput() bool {
	if pm.state == showInput {
		return true
	}

	return false
}

func (pm *PasswordModel) listView() string {
	pm.list.Styles.Title = common.ListColorStyle
	pm.list.Styles.FilterPrompt.Foreground(common.ListColorStyle.GetBackground())
	pm.list.Styles.FilterCursor.Foreground(common.ListColorStyle.GetBackground())

	return common.ListStyle.Render(pm.list.View())
}

func (pm *PasswordModel) detailView() string {
	builder := &strings.Builder{}
	divider := common.DividerStyle.Render(strings.Repeat("-", pm.viewport.Width)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")

	if it := pm.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderPassword(it.(item), pm.echoPassword))
		builder.WriteString(divider)
	}

	builder.WriteString(pm.lastMessage)
	details := wordwrap.String(builder.String(), pm.viewport.Width)

	return common.DetailStyle.Render(details)
}

func (pm *PasswordModel) LoadData() {
	pm.list.SetItems(nil)
	pm.lastMessage = ""
	passwords, err := pm.pc.Passwords()
	if err != nil {
		pm.lastMessage = err.Error()

		return
	}

	pm.lastMessage = "successful load data"

	items := make([]list.Item, 0, len(passwords))
	for _, p := range passwords {
		i := item{
			ID:        p.ID,
			Title:     p.Title,
			Login:     p.Login,
			Password:  p.Password,
			URL:       p.URL,
			Note:      p.Note,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			ExpiredAt: p.ExpiredAt,
		}

		items = append(items, i)
	}

	pm.list.SetItems(items)
}

func (pm *PasswordModel) sendPassword(password models.Password) {
	var err error
	if password.ID == 0 {
		err = pm.pc.CreatePassword(password)
	} else {
		err = pm.pc.UpdatePassword(password)
	}

	if err != nil {
		pm.lastMessage = err.Error()
		return
	}

	pm.LoadData()
}

func (pm *PasswordModel) deletePassword(id int) {
	err := pm.pc.DeletePassword(id)
	if err != nil {
		pm.lastMessage = err.Error()
		return
	}

	pm.lastMessage = fmt.Sprintf("password (id: %d) successful deleted", id)

	pm.LoadData()
}

func renderPassword(i item, echo echoPasswordMode) string {
	title := fmt.Sprintf("\n\nTitle: %s\n", i.Title)
	login := fmt.Sprintf("\n\nLogin: %s\n", i.Login)
	var pass string
	if echo == EchoPassword {
		pass = fmt.Sprintf("\n\nPassword: %s\n", "************")
	} else {
		pass = fmt.Sprintf("\n\nPassword: %s\n", i.Password)
	}
	url := fmt.Sprintf("\n\nURL: %s\n", i.URL)
	note := fmt.Sprintf("\n\nNote: %s\n", i.Note)
	createdAt := fmt.Sprintf("\n\nCreated: %s\n", i.CreatedAt.Format(common.InputTimeFormLong))
	updatedAt := fmt.Sprintf("\n\nUpdated: %s\n", i.UpdatedAt.Format(common.InputTimeFormLong))
	expirationAt := fmt.Sprintf("\n\nExpiration: %s\n", i.ExpiredAt.Format(common.InputTimeFormLong))

	return title + login + pass + url + note + createdAt + updatedAt + expirationAt
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	pass, ok := listItem.(item)
	if !ok {
		return
	}

	line := pass.Title

	if index == m.Index() {
		line = common.ListSelectedListItemStyle.Render("> " + line)
	} else {
		line = common.ListItemStyle.Render(line)
	}

	fmt.Fprint(w, line)
}
