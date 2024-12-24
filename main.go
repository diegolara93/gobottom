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
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

var docStyle = lipgloss.NewStyle().Margin(0, 0).MaxHeight(50)

type tickMsg time.Time

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")). // purple
	Margin(1)

var utilHeaderStyle = lipgloss.NewStyle().
	MarginTop(1).MarginLeft(13).
	Background(lipgloss.Color("63"))

var memoryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("1")).
	PaddingTop(2).
	MarginTop(2).
	PaddingBottom(2).
	BorderStyle(lipgloss.NormalBorder())

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
	memory    int
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
		h, v := docStyle.GetFrameSize()
		m.list.SetSize((msg.Width-h)/2, (msg.Height-v)/2)
	}

	m.utilChart, _ = m.utilChart.Update(msg)
	m.utilChart.DrawAll()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := docStyle.Render(m.list.View())
	w := utilHeaderStyle.Render("Utilization")
	// w += defaultStyle.Render(m.utilChart.View())
	w += lipgloss.JoinVertical(lipgloss.Right, defaultStyle.Render(m.utilChart.View()))
	total_mem := memoryStyle.Render("Total Memory: " + strconv.Itoa(m.memory))
	new_view := lipgloss.JoinHorizontal(lipgloss.Center, s, w)
	new_view += lipgloss.JoinVertical(lipgloss.Center, total_mem)
	return new_view
}

func main() {
	width := 36
	height := 18
	minYValue := 0.0
	maxYValue := 112.0
	process, _ := process.Processes()
	mem, err := mem.VirtualMemory()

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

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), utilChart: slc1, memory: int(mem.Total)}
	m.list.Title = "Active Processes"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
