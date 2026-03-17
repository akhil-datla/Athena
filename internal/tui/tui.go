package tui

import (
	"fmt"
	"strings"

	"github.com/akhil-datla/athena/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	queryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	questionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("039"))

	answerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("035"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	countStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// Model is the bubbletea model for the interactive search TUI.
type Model struct {
	engine   *engine.Engine
	query    string
	results  []engine.SearchResult
	cursor   int
	maxShown int
	quitting bool
}

// New creates a new TUI model.
func New(eng *engine.Engine) Model {
	return Model{
		engine:   eng,
		maxShown: 8,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if len(m.results) > 0 && m.cursor < len(m.results) {
				m.quitting = true
				return m, tea.Quit
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			max := len(m.results) - 1
			if m.maxShown-1 < max {
				max = m.maxShown - 1
			}
			if m.cursor < max {
				m.cursor++
			}

		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.results = m.engine.Search(m.query, 0)
				m.cursor = 0
			}

		default:
			if len(msg.String()) == 1 {
				m.query += msg.String()
				m.results = m.engine.Search(m.query, 0)
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		if len(m.results) > 0 && m.cursor < len(m.results) {
			r := m.results[m.cursor]
			return fmt.Sprintf("\n%s\n%s\n\n",
				questionStyle.Render("Q: "+r.Question),
				answerStyle.Render("A: "+r.Answer),
			)
		}
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Athena Interactive Search"))
	b.WriteString("\n")

	// Search prompt
	cursor := "█"
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(queryStyle.Render(m.query))
	b.WriteString(cursor)
	b.WriteString("\n\n")

	if m.query == "" {
		b.WriteString(answerStyle.Render("  Type to search..."))
		b.WriteString("\n")
	} else if len(m.results) == 0 {
		b.WriteString(answerStyle.Render("  No results found."))
		b.WriteString("\n")
	} else {
		shown := len(m.results)
		if shown > m.maxShown {
			shown = m.maxShown
		}

		b.WriteString(countStyle.Render(fmt.Sprintf("  %d result(s)", len(m.results))))
		b.WriteString("\n\n")

		for i := 0; i < shown; i++ {
			r := m.results[i]
			relevance := formatRelevance(r.Score, m.query)

			pointer := "  "
			qStyle := questionStyle
			if i == m.cursor {
				pointer = selectedStyle.Render("▸ ")
				qStyle = selectedStyle
			}

			b.WriteString(fmt.Sprintf("%s%s  %s\n",
				pointer,
				qStyle.Render(r.Question),
				scoreStyle.Render(relevance),
			))
			b.WriteString(fmt.Sprintf("    %s\n", answerStyle.Render(truncate(r.Answer, 80))))
		}

		if len(m.results) > m.maxShown {
			b.WriteString(fmt.Sprintf("\n%s", countStyle.Render(
				fmt.Sprintf("  ... and %d more results", len(m.results)-m.maxShown),
			)))
			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("  ↑/↓ navigate • enter select • esc quit"))
	b.WriteString("\n")

	return b.String()
}

// SelectedResult returns the currently selected result, if any.
func (m Model) SelectedResult() (engine.SearchResult, bool) {
	if len(m.results) > 0 && m.cursor < len(m.results) {
		return m.results[m.cursor], true
	}
	return engine.SearchResult{}, false
}

func formatRelevance(score int, query string) string {
	if len(query) == 0 {
		return ""
	}
	// Lower score = better match. Normalize to a percentage.
	maxDist := len(query) * 2
	if maxDist == 0 {
		maxDist = 1
	}
	pct := 100 - (score*100)/maxDist
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%d%%", pct)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
