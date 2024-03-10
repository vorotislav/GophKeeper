package tui

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/tui/cards"
	"GophKeeper/internal/tui/media"
	"GophKeeper/internal/tui/notes"
	"GophKeeper/internal/tui/passwords"
	"GophKeeper/internal/tui/start"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

type sessionClient interface {
	Login(user models.User) (models.Session, error)
	Register(user models.User) (models.Session, error)
}

type passwordClient interface {
	CreatePassword(pass models.Password) error
	UpdatePassword(pass models.Password) error
	Passwords() ([]models.Password, error)
	DeletePassword(id int) error
}

type cardsClient interface {
	CreateCard(c models.Card) error
	UpdateCard(c models.Card) error
	Cards() ([]models.Card, error)
	DeleteCard(id int) error
}

type notesClient interface {
	CreateNote(n models.Note) error
	UpdateNote(n models.Note) error
	Notes() ([]models.Note, error)
	DeleteNote(id int) error
}

type mediaClient interface {
	CreateMedia(m models.Media) error
	UpdateMedia(m models.Media) error
	Medias() ([]models.Media, error)
	DeleteMedia(id int) error
}

var (
	color = termenv.EnvColorProfile().Color
	help  = termenv.Style{}.Foreground(color("241")).Styled
)

type state int

const (
	startView state = iota
	programView
)

type programState int

const (
	programNoState programState = iota
	programPassword
	programCard
	programNote
	programMedia
)

type model struct {
	state        state
	programState programState

	sm start.StartModel
	pm passwords.PasswordModel
	cm cards.CardModel
	nm notes.NoteModel
	mm media.MediaModel

	width, height int
}

type keymap struct {
	Add    key.Binding
	Remove key.Binding
	Next   key.Binding
	Prev   key.Binding
	Enter  key.Binding
	Back   key.Binding
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
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "back"),
	),
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("GophKeeper")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	windowSizeMsg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "1" {
			if !m.checkInput() {
				m.programState = programPassword
				m.pm, cmd = m.pm.Update(windowSizeMsg)
			}

		}
		if msg.String() == "2" {
			if !m.checkInput() {
				m.programState = programCard
				m.cm, cmd = m.cm.Update(windowSizeMsg)
			}
		}
		if msg.String() == "3" {
			if !m.checkInput() {
				m.programState = programNote
				m.nm, cmd = m.nm.Update(windowSizeMsg)
			}

		}
		if msg.String() == "4" {
			if !m.checkInput() {
				m.programState = programMedia
				m.mm, cmd = m.mm.Update(windowSizeMsg)
			}
		}
		if msg.String() == "esc" {
			if m.state == programView {
				return m, nil
			} else {
				return m, tea.Quit
			}
		}
	}

	if m.state == startView {
		m.sm, cmd = m.sm.Update(msg)

		switch m.sm.State {
		case start.BackState:
			return m, tea.Quit
		case start.SessionState:
			m.state = programView

			m.pm.LoadData()
			m.nm.LoadData()
			m.cm.LoadData()
			m.mm.LoadData()
			return updateByState(m)
		}
		return m, cmd
	} else {
		switch m.programState {
		case programPassword:
			m.pm, cmd = m.pm.Update(msg)

			return m, cmd
		case programCard:
			m.cm, cmd = m.cm.Update(msg)

			return m, cmd
		case programNote:
			m.nm, cmd = m.nm.Update(msg)

			return m, cmd
		case programMedia:
			m.mm, cmd = m.mm.Update(msg)

			return m, cmd
		default:
			return m, nil
		}
	}
}

func updateByState(m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	windowSizeMsg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	}

	switch m.programState {
	case programPassword:
		m.pm, cmd = m.pm.Update(windowSizeMsg)
	case programCard:
		m.cm, cmd = m.cm.Update(windowSizeMsg)
	case programNote:
		m.nm, cmd = m.nm.Update(windowSizeMsg)
	case programMedia:
		m.mm, cmd = m.mm.Update(windowSizeMsg)
	}

	return m, cmd
}

func (m model) checkInput() bool {
	if m.pm.IsInput() {
		return true
	}

	if m.cm.IsInput() {
		return true
	}

	if m.nm.IsInput() {
		return true
	}

	if m.mm.IsInput() {
		return true
	}

	return false
}

func (m model) View() string {
	if m.state == startView {
		return m.sm.View()
	}

	switch m.programState {
	case programPassword:
		return m.pm.View()
	case programCard:
		return m.cm.View()
	case programNote:
		return m.nm.View()
	case programMedia:
		return m.mm.View()
	}

	return m.pm.View()
}

func InitModel(
	cl sessionClient,
	pc passwordClient,
	cc cardsClient,
	nc notesClient,
	mc mediaClient,
) model {
	return model{
		state:        startView,
		programState: programPassword,
		sm:           start.InitialModel(cl),
		pm:           passwords.InitialModel(pc),
		cm:           cards.InitialModel(cc),
		nm:           notes.InitialModel(nc),
		mm:           media.InitialModel(mc),
	}
}
