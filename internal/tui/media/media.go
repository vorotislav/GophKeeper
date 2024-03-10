package media

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/common"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakenelf/teacup/help"
)

const (
	inputTimeFormShort = "2006-01-02"
	inputTimeFormLong  = "2006-01-02 15:04:05"
)

type mediaClient interface {
	CreateMedia(m models.Media) error
	UpdateMedia(m models.Media) error
	Medias() ([]models.Media, error)
	DeleteMedia(id int) error
}

type item models.Media

func (i item) GetTitle() string       { return i.Title }
func (i item) FilterValue() string    { return i.Title }
func (i item) GetMedia() models.Media { return models.Media(i) }

type state int

const (
	showMedias state = iota
	showInput
)

type MediaModel struct {
	list        list.Model
	viewport    viewport.Model
	mc          mediaClient
	lastMessage string
	state       state
	im          MediaInputModel
	help        help.Model
}

type keymap struct {
	Add    key.Binding
	Remove key.Binding
	Change key.Binding
	Quit   key.Binding
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
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "back"),
	),
}

func InitialModel(mc mediaClient) MediaModel {
	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "Medias"
	l.SetFilteringEnabled(false)

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
			{Key: "+", Description: "Add media"},
			{Key: "-", Description: "Delete media"},
		},
	)

	return MediaModel{
		list:  l,
		mc:    mc,
		state: showMedias,
		help:  helpModel,
		im:    InitialInputModel(),
	}
}

func (mm *MediaModel) IsInput() bool {
	if mm.state == showInput {
		return true
	}

	return false
}

func (mm *MediaModel) Init() tea.Cmd { return nil }

func (mm *MediaModel) View() string {
	switch mm.state {
	case showInput:
		return mm.im.View()
	default:
		mm.viewport.SetContent(mm.detailView())

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			mm.listView(),
			mm.viewport.View(),
			mm.help.View())
	}
}

func (mm *MediaModel) Update(msg tea.Msg) (MediaModel, tea.Cmd) {
	var cmd tea.Cmd

	switch mm.state {
	case showMedias:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			horizontal, vertical := common.ListStyle.GetFrameSize()
			paginatorHeight := lipgloss.Height(mm.list.Paginator.View())

			mm.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
			mm.viewport = viewport.New(msg.Width/2-10, msg.Height)
			mm.viewport.SetContent(mm.detailView())
			mm.help.SetSize(msg.Width/2, msg.Height)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Add):
				mm.state = showInput
				mm.im.SetMedia(models.Media{})
			case key.Matches(msg, Keymap.Change):
				mm.state = showInput
				mm.im.SetMedia(mm.list.SelectedItem().(item).GetMedia())
			case key.Matches(msg, Keymap.Remove):
				mm.deleteMedia(mm.list.SelectedItem().(item).ID)
			}
		}
		mm.list, cmd = mm.list.Update(msg)
	case showInput:
		mm.im, cmd = mm.im.Update(msg)
		if cmd != nil {
			is, ok := cmd().(InputState)
			if ok {
				switch is.is {
				case CancelState:
					mm.state = showMedias
				case SubmitState:
					mm.sendMedia(mm.im.Media())
					mm.state = showMedias
				}
			}
		}
	}

	return *mm, cmd
}

func (mm *MediaModel) LoadData() {
	mm.lastMessage = ""
	medias, err := mm.mc.Medias()
	if err != nil {
		mm.lastMessage = err.Error()

		return
	}

	mm.lastMessage = "successful load data"

	items := make([]list.Item, 0, len(medias))
	for _, m := range medias {
		i := item{
			ID:        m.ID,
			Title:     m.Title,
			Body:      m.Body,
			MediaType: m.MediaType,
			Note:      m.Note,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}

		items = append(items, i)
	}

	mm.list.SetItems(items)
}

func (nm *MediaModel) listView() string {
	nm.list.Styles.Title = common.ListColorStyle
	nm.list.Styles.FilterPrompt.Foreground(common.ListColorStyle.GetBackground())
	nm.list.Styles.FilterCursor.Foreground(common.ListColorStyle.GetBackground())

	return common.ListStyle.Render(nm.list.View())
}

func (nm *MediaModel) detailView() string {
	builder := &strings.Builder{}
	divider := common.DividerStyle.Render(strings.Repeat("-", nm.viewport.Width)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")

	if it := nm.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderMedia(it.(item)))
		builder.WriteString(divider)
	}

	builder.WriteString(nm.lastMessage)
	details := wordwrap.String(builder.String(), nm.viewport.Width)

	return common.DetailStyle.Render(details)
}

func (mm *MediaModel) deleteMedia(id int) {
	err := mm.mc.DeleteMedia(id)
	if err != nil {
		mm.lastMessage = err.Error()

		return
	}

	mm.lastMessage = fmt.Sprintf("media (id:%d) successful deleted", id)

	mm.LoadData()
}

func (mm *MediaModel) sendMedia(media models.Media) {
	var err error
	if media.ID == 0 {
		err = mm.mc.CreateMedia(media)
	} else {
		err = mm.mc.UpdateMedia(media)
	}

	if err != nil {
		mm.lastMessage = err.Error()

		return
	}

	mm.LoadData()
}

func renderMedia(i item) string {
	title := fmt.Sprintf("\n\nTitle: %s\n", i.Title)
	mType := fmt.Sprintf("\n\nMediaType: %s\n", i.MediaType)
	note := fmt.Sprintf("\n\nNote: %s\n", i.Note)

	createdAt := fmt.Sprintf("\n\nCreated: %s\n", i.CreatedAt.Format(inputTimeFormLong))
	updatedAt := fmt.Sprintf("\n\nUpdated: %s\n", i.UpdatedAt.Format(inputTimeFormLong))
	expiredAt := fmt.Sprintf("\n\nExpiration: %s\n", i.ExpiredAt.Format(inputTimeFormLong))

	return title + mType + note + createdAt + updatedAt + expiredAt
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	media, ok := listItem.(item)
	if !ok {
		return
	}

	line := media.Title

	if index == m.Index() {
		line = common.ListSelectedListItemStyle.Render("> " + line)
	} else {
		line = common.ListItemStyle.Render(line)
	}

	fmt.Fprint(w, line)
}
