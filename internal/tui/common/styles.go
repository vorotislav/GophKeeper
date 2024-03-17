package common

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultInputWidth = 22
	cError            = "#CF002E"
	cTitle            = "#2389D3"
	cDetailTitle      = "#3d719c"
	cPromptBorder     = "#569cd6"
	cTextLightGray    = "#FFFDF5"
)

var (
	ListStyle = lipgloss.NewStyle().
			Width(35).
			MarginTop(1).
			PaddingRight(3).
			MarginRight(3).
			Border(lipgloss.RoundedBorder())
	ListColorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3d719c"))
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)
	ListSelectedListItemStyle = lipgloss.NewStyle().
					PaddingLeft(2).
					Foreground(lipgloss.Color("#569cd6"))
	DetailStyle = lipgloss.NewStyle().
			PaddingTop(2)
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).
			PaddingTop(1).
			PaddingBottom(1)
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#569cd6")).
			PaddingBottom(1).
			Bold(true).
			Underline(true).
			Inline(true)
	InputTitleStyle = lipgloss.NewStyle().
			Width(defaultInputWidth).
			Foreground(lipgloss.Color(cTextLightGray)).
			Background(lipgloss.Color(cDetailTitle)).
			Padding(0, 1).
			Align(lipgloss.Center)
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cPromptBorder))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	NoStyle      = lipgloss.NewStyle()
	ErrStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(cError)).Render
	InputStyle   = lipgloss.NewStyle().
			Margin(1, 1).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color(cPromptBorder)).
			Render
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(cTextLightGray)).
			Background(lipgloss.Color(cTitle)).
			Padding(0, 1)
	InactiveBoxBorderColor lipgloss.AdaptiveColor
)
