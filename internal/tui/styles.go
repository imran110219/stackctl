package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#06B6D4")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorError     = lipgloss.Color("#EF4444")
	colorMuted     = lipgloss.Color("#6B7280")
	colorHighlight = lipgloss.Color("#E0E7FF")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	categoryStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			MarginTop(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	// Table and tab styles for dashboard and module manager
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorMuted)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted).
				Padding(0, 2)

	// Status indicator styles
	statusRunning = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	statusStopped = lipgloss.NewStyle().
			Foreground(colorError)

	statusUnknown = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Secret masking style
	secretStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	cursorChar = selectedStyle.Render(">")
	checkOn    = selectedStyle.Render("[x]")
	checkOff   = normalStyle.Render("[ ]")
	radioOn    = selectedStyle.Render("(*)")
	radioOff   = normalStyle.Render("( )")
)
