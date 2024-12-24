package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/process"
)

var docStyle = lipgloss.NewStyle().Margin(1, 0)

type tickMsg time.Time

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")) // purple

var graphLineStyle1 = lipgloss.NewStyle().
	Foreground(lipgloss.Color("4")) // blue

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")) // cyan

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list      list.Model
	utilChart streamlinechart.Model
	width     int
	height    int
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tickMsg:
		m.utilChart.Push(rand.Float64() * 100.0)
		m.utilChart.Draw()

		// Schedule the next tick
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listWidth := int(0.4 * float64(m.width))
		// Subtract an extra 2 rows for the list’s title and borders.
		listHeight := m.height - 2

		m.list.SetSize(listWidth, listHeight)
	}

	m.utilChart, _ = m.utilChart.Update(msg)
	m.utilChart.DrawAll()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	listWidth := int(0.4 * float64(m.width))
	chartWidth := m.width - listWidth

	// It’s good practice to account for any border or margin:
	// E.g., docStyle might reduce the actual available width/height.
	// You can get margin/padding with docStyle.GetFrameSize() if needed.

	// Render list and chart with explicit widths.
	// For the chart’s height, let’s just reuse our total terminal height
	// minus any vertical margin from docStyle.
	leftView := lipgloss.NewStyle().
		Width(listWidth).
		Render(m.list.View())
	rightView := lipgloss.NewStyle().
		Width(chartWidth).
		Render(m.utilChart.View())

	// Join horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)

	// Finally, wrap in docStyle if desired
	return docStyle.Render(content)
}

func main() {
	width := 36
	height := 14
	minYValue := 0.0
	maxYValue := 100.0
	process, err := process.Processes()
	if err != nil {
		fmt.Printf("error")
	}
	items := []list.Item{}

	for _, proc := range process {
		process_name, _ := proc.Name()
		process_id := "PID: " + strconv.Itoa(int(proc.Pid))
		items = append(items, item{title: "Process: " + process_name, desc: process_id})
	}

	slc1 := streamlinechart.New(width, height)
	slc1.AxisStyle = axisStyle
	slc1.LabelStyle = labelStyle
	slc1.SetYRange(minYValue, maxYValue)                 // set expected Y values (values can be less or greater than what is displayed)
	slc1.SetViewYRange(minYValue, maxYValue)             // setting display Y values will fail unless set expected Y values first
	slc1.SetStyles(runes.ThinLineStyle, graphLineStyle1) // graphLineStyle1 replaces linechart rune style
	slc1.Focus()

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), utilChart: slc1}
	m.list.Title = "My Fave Things"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
