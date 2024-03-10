package common

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"time"
)

const (
	secondsPerYear     = 31557600
	secondsPerDay      = 86400
	secondsPerHour     = 3600
	secondsPerMinute   = 60
	timeout            = 365 * 24 * time.Hour
	defaultListWidth   = 28
	defaultListHeight  = 40
	defaultDetailWidth = 45
	defaultInputWidth  = 22
	defaultHelpHeight  = 4
	eventsFile         = "events.json"
	inputTimeFormShort = "2006-01-02"
	inputTimeFormLong  = "2006-01-02 15:04:05"
	cError             = "#CF002E"
	cItemTitleDark     = "#F5EB6D"
	cItemTitleLight    = "#F3B512"
	cItemDescDark      = "#9E9742"
	cItemDescLight     = "#FFD975"
	cTitle             = "#2389D3"
	cDetailTitle       = "#3d719c"
	cPromptBorder      = "#569cd6"
	cDimmedTitleDark   = "#DDDDDD"
	cDimmedTitleLight  = "#222222"
	cDimmedDescDark    = "#999999"
	cDimmedDescLight   = "#555555"
	cTextLightGray     = "#FFFDF5"
)

var (
	AppStyle  = lipgloss.NewStyle().Margin(0, 1)
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
	TableMainStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true)
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
	HelpStyle     = list.DefaultStyles().HelpStyle.Width(defaultListWidth).Height(5)
	SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{Light: cItemTitleLight, Dark: cItemTitleDark}).
			Foreground(lipgloss.AdaptiveColor{Light: cItemTitleLight, Dark: cItemTitleDark}).
			Padding(0, 0, 0, 1)
	SelectedDesc = SelectedTitle.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: cItemDescLight, Dark: cItemDescDark})
	DimmedTitle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: cDimmedTitleLight, Dark: cDimmedTitleDark}).
			Padding(0, 0, 0, 2)
	DimmedDesc = DimmedTitle.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: cDimmedDescDark, Dark: cDimmedDescLight})
	TitleBackgroundColor   lipgloss.AdaptiveColor
	TitleForegroundColor   lipgloss.AdaptiveColor
	ActiveBoxBorderColor   lipgloss.AdaptiveColor
	InactiveBoxBorderColor lipgloss.AdaptiveColor
)
