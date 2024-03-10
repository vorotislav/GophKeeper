package notes

import (
	"fmt"
	"io"
	"strings"

	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/common"
	"github.com/charmbracelet/bubbles/key"
	"github.com/mistakenelf/teacup/help"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type notesClient interface {
	CreateNote(n models.Note) error
	UpdateNote(n models.Note) error
	Notes() ([]models.Note, error)
	DeleteNote(id int) error
}

type item models.Note

func (i item) GetTitle() string    { return i.Title }
func (i item) FilterValue() string { return i.Title }
func (i item) GetNote() models.Note {
	return models.Note{
		ID:        i.ID,
		Title:     i.Title,
		Text:      i.Text,
		ExpiredAt: i.ExpiredAt,
	}
}

type notesState int

const (
	showNotes notesState = iota
	showInput
)

type NoteModel struct {
	list        list.Model
	viewport    viewport.Model
	nc          notesClient
	lastMessage string
	state       notesState
	im          NoteInputModel
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
	Change: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "change"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "back"),
	),
}

func InitialModel(nc notesClient) NoteModel {
	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "Notes"
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
			{Key: "+", Description: "Add note"},
			{Key: "-", Description: "Delete note"},
			{Key: "r", Description: "Change note"},
		},
	)

	return NoteModel{
		list:  l,
		nc:    nc,
		im:    InitialInputModel(),
		state: showNotes,
		help:  helpModel,
	}
}

func (nm *NoteModel) IsInput() bool {
	if nm.state == showInput {
		return true
	}

	return false
}

func (nm *NoteModel) Init() tea.Cmd { return nil }

func (nm *NoteModel) View() string {
	switch nm.state {
	case showInput:
		return nm.im.View()
	default:
		nm.viewport.SetContent(nm.detailView())

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			nm.listView(),
			nm.viewport.View(),
			nm.help.View())
	}
}

func (nm *NoteModel) Update(msg tea.Msg) (NoteModel, tea.Cmd) {
	var cmd tea.Cmd

	switch nm.state {
	case showNotes:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			horizontal, vertical := common.ListStyle.GetFrameSize()
			paginatorHeight := lipgloss.Height(nm.list.Paginator.View())

			nm.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
			nm.viewport = viewport.New(msg.Width/2-10, msg.Height)
			nm.viewport.SetContent(nm.detailView())
			nm.help.SetSize(msg.Width/2, msg.Height)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Add):
				nm.state = showInput
				nm.im.SetNote(models.Note{})
			case key.Matches(msg, Keymap.Change):
				nm.state = showInput
				nm.im.SetNote(nm.list.SelectedItem().(item).GetNote())
			case key.Matches(msg, Keymap.Remove):
				nm.deleteNote(nm.list.SelectedItem().(item).ID)
			}
		}
		nm.list, cmd = nm.list.Update(msg)
	case showInput:
		nm.im, cmd = nm.im.Update(msg)
		if cmd != nil {
			is, ok := cmd().(InputState)
			if ok {
				switch is.is {
				case CancelState:
					nm.state = showNotes
				case SubmitState:
					nm.sendNote(nm.im.note)
					nm.state = showNotes
				}
			}
		}
	}

	return *nm, cmd
}

func (nm *NoteModel) LoadData() {
	nm.lastMessage = ""
	notes, err := nm.nc.Notes()
	if err != nil {
		nm.lastMessage = err.Error()

		return
	}

	nm.lastMessage = "successful load data"

	items := make([]list.Item, 0, len(notes))
	for _, n := range notes {
		i := item{
			ID:        n.ID,
			Title:     n.Title,
			Text:      n.Text,
			ExpiredAt: n.ExpiredAt,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		}

		items = append(items, i)
	}

	nm.list.SetItems(items)
}

func (nm *NoteModel) listView() string {
	nm.list.Styles.Title = common.ListColorStyle
	nm.list.Styles.FilterPrompt.Foreground(common.ListColorStyle.GetBackground())
	nm.list.Styles.FilterCursor.Foreground(common.ListColorStyle.GetBackground())

	return common.ListStyle.Render(nm.list.View())
}

func (nm *NoteModel) detailView() string {
	builder := &strings.Builder{}
	divider := common.DividerStyle.Render(strings.Repeat("-", nm.viewport.Width)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")

	if it := nm.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderNote(it.(item)))
		builder.WriteString(divider)
	}

	builder.WriteString(nm.lastMessage)
	details := wordwrap.String(builder.String(), nm.viewport.Width)

	return common.DetailStyle.Render(details)
}

func (nm *NoteModel) deleteNote(id int) {
	err := nm.nc.DeleteNote(id)
	if err != nil {
		nm.lastMessage = err.Error()

		return
	}

	nm.lastMessage = fmt.Sprintf("note (id:%d) successful deleted", id)

	nm.LoadData()
}

func (nm *NoteModel) sendNote(note models.Note) {
	var err error
	if note.ID == 0 {
		err = nm.nc.CreateNote(note)
	} else {
		err = nm.nc.UpdateNote(note)
	}

	if err != nil {
		nm.lastMessage = err.Error()

		return
	}

	nm.LoadData()
}

func renderNote(i item) string {
	title := fmt.Sprintf("\n\nTitle: %s\n", i.Title)
	text := fmt.Sprintf("\n\nLogin: %s\n", i.Text)

	createdAt := fmt.Sprintf("\n\nCreated: %s\n", i.CreatedAt.Format(inputTimeFormLong))
	updatedAt := fmt.Sprintf("\n\nUpdated: %s\n", i.UpdatedAt.Format(inputTimeFormLong))
	expiredAt := fmt.Sprintf("\n\nExpiration: %s\n", i.ExpiredAt.Format(inputTimeFormLong))

	return title + text + createdAt + updatedAt + expiredAt
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	note, ok := listItem.(item)
	if !ok {
		return
	}

	line := note.Title

	if index == m.Index() {
		line = common.ListSelectedListItemStyle.Render("> " + line)
	} else {
		line = common.ListItemStyle.Render(line)
	}

	fmt.Fprint(w, line)
}
