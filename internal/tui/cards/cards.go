package cards

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

type cardsClient interface {
	CreateCard(card models.Card) error
	UpdateCard(card models.Card) error
	Cards() ([]models.Card, error)
	DeleteCard(id int) error
}

type item models.Card

func (i item) GetTitle() string    { return i.Name }
func (i item) FilterValue() string { return i.Name }
func (i item) GetCard() models.Card {
	return models.Card(i)
}

type echoCVCMode int

const (
	EchoNormal echoCVCMode = iota
	EchoCVC
)

type state int

const (
	showCards state = iota
	showInput
)

type keymap struct {
	Add      key.Binding
	Remove   key.Binding
	Change   key.Binding
	ShowPass key.Binding
	Quit     key.Binding
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
		key.WithHelp("s", "show CVC"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "back"),
	),
}

type CardModel struct {
	list        list.Model
	viewport    viewport.Model
	cc          cardsClient
	lastMessage string
	im          CardInputModel
	echoCVC     echoCVCMode
	state       state
	help        help.Model
}

func (cm *CardModel) Init() tea.Cmd { return nil }

func InitialModel(cc cardsClient) CardModel {
	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "Cards"
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
			{Key: "+", Description: "Add card"},
			{Key: "-", Description: "Delete card"},
			{Key: "r", Description: "Change card"},
			{Key: "s", Description: "Show CVC in cards view"},
		},
	)

	return CardModel{
		list:    l,
		cc:      cc,
		echoCVC: EchoCVC,
		state:   showCards,
		im:      InitialInputModel(),
		help:    helpModel,
	}
}

func (cm *CardModel) View() string {
	switch cm.state {
	case showInput:
		return cm.im.View()
	default:
		cm.viewport.SetContent(cm.detailView())

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			cm.listView(),
			cm.viewport.View(),
			cm.help.View())
	}
}

func (cm *CardModel) Update(msg tea.Msg) (CardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch cm.state {
	case showCards:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			horizontal, vertical := common.ListStyle.GetFrameSize()
			paginatorHeight := lipgloss.Height(cm.list.Paginator.View())

			cm.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
			cm.viewport = viewport.New(msg.Width/2-10, msg.Height)
			cm.viewport.SetContent(cm.detailView())
			cm.help.SetSize(msg.Width/2, msg.Height)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.ShowPass):
				if cm.echoCVC == EchoCVC {
					cm.echoCVC = EchoNormal
				} else {
					cm.echoCVC = EchoCVC
				}
				return *cm, nil
			case key.Matches(msg, Keymap.Add):
				cm.state = showInput
				cm.im.SetCard(models.Card{})
			case key.Matches(msg, Keymap.Remove):
				cm.deleteCard(cm.list.SelectedItem().(item).GetCard().ID)
			case key.Matches(msg, Keymap.Change):
				cm.state = showInput
				cm.im.SetCard(cm.list.SelectedItem().(item).GetCard())
			}
		}

		cm.list, _ = cm.list.Update(msg)
	case showInput:
		cm.im, cmd = cm.im.Update(msg)
		if cmd != nil {
			is, ok := cmd().(InputState)
			if ok {
				switch is.is {
				case CancelState:
					cm.state = showCards
				case SubmitState:
					cm.sendCard(cm.im.card)
					cm.state = showCards
				}
			}
		}

		return *cm, cmd
	}

	return *cm, cmd
}

func (cm *CardModel) IsInput() bool {
	if cm.state == showInput {
		return true
	}

	return false
}

func (cm *CardModel) LoadData() {
	cm.lastMessage = ""
	cards, err := cm.cc.Cards()
	if err != nil {
		cm.lastMessage = err.Error()

		return
	}

	cm.lastMessage = "successful load data"

	items := make([]list.Item, 0, len(cards))
	for _, c := range cards {
		i := item{
			ID:        c.ID,
			Name:      c.Name,
			Number:    c.Number,
			CVC:       c.CVC,
			ExpMonth:  c.ExpMonth,
			ExpYear:   c.ExpYear,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}

		items = append(items, i)
	}

	cm.list.SetItems(items)
}

func (cm *CardModel) listView() string {
	cm.list.Styles.Title = common.ListColorStyle
	cm.list.Styles.FilterPrompt.Foreground(common.ListColorStyle.GetBackground())
	cm.list.Styles.FilterCursor.Foreground(common.ListColorStyle.GetBackground())

	return common.ListStyle.Render(cm.list.View())
}

func (cm *CardModel) detailView() string {
	builder := &strings.Builder{}
	divider := common.DividerStyle.Render(strings.Repeat("-", cm.viewport.Width)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")

	if it := cm.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderCard(it.(item), cm.echoCVC))
		builder.WriteString(divider)
	}
	builder.WriteString(cm.lastMessage)
	details := wordwrap.String(builder.String(), cm.viewport.Width)

	return common.DetailStyle.Render(details)
}

func (cm *CardModel) sendCard(card models.Card) {
	var err error
	if card.ID == 0 {
		err = cm.cc.CreateCard(card)
	} else {
		err = cm.cc.UpdateCard(card)
	}

	if err != nil {
		cm.lastMessage = err.Error()

		return
	}

	cm.LoadData()
}

func (cm *CardModel) deleteCard(id int) {
	err := cm.cc.DeleteCard(id)
	if err != nil {
		cm.lastMessage = err.Error()

		return
	}

	cm.LoadData()
}

func renderCard(i item, echo echoCVCMode) string {
	name := fmt.Sprintf("\n\nName: %s\n", i.Name)
	number := fmt.Sprintf("\n\nNumber: %s\n", i.Number)
	var cvc string
	if echo == EchoCVC {
		cvc = fmt.Sprintf("\n\nCVC: %s\n", "***")
	} else {
		cvc = fmt.Sprintf("\n\nCVC: %s\n", i.CVC)
	}

	exp := fmt.Sprintf("\n\nExp: %d/%d\n", i.ExpMonth, i.ExpYear)
	createdAt := fmt.Sprintf("\n\nCreated: %s\n", i.CreatedAt.String())
	updatedAt := fmt.Sprintf("\n\nUpdated: %s\n", i.UpdatedAt.String())

	return name + number + cvc + exp + createdAt + updatedAt
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

	line := pass.Name

	if index == m.Index() {
		line = common.ListSelectedListItemStyle.Render("> " + line)
	} else {
		line = common.ListItemStyle.Render(line)
	}

	fmt.Fprint(w, line)
}
