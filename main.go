package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

var docStyle = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#98971a"))
var docStyle2 = lipgloss.NewStyle().Margin(2, 0, 0, 0).Border(lipgloss.NormalBorder()).Width(20)

type tickMsg time.Time

var list_item_style = lipgloss.NewStyle().
	Foreground(lipgloss.Color("32")).
	Height(1).
	Margin(0, 0)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")). // purple
	Margin(2)

var utilHeaderStyle = lipgloss.NewStyle().
	MarginTop(0).MarginLeft(13).
	Background(lipgloss.Color("63"))

var memoryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("1")).
	PaddingTop(2).
	MarginTop(2).
	PaddingBottom(2).
	BorderStyle(lipgloss.NormalBorder()).
	Width(46).
	Height(18).Align(lipgloss.Center)

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
	list           list.Model
	utilChart      streamlinechart.Model
	cpuUtilzations float64
	memory         *mem.VirtualMemoryStat
	utilChart2     streamlinechart.Model
	list_cpus      list.Model
	selected_list  int32
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

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		if zone.Get("CPU").InBounds(msg) {
			m.selected_list = 1
			docStyle2 = docStyle2.BorderForeground(lipgloss.Color("#98971a"))
			docStyle = docStyle.BorderForeground(lipgloss.Color("#ffffff"))
		} else if zone.Get("Other List").InBounds(msg) {
			m.selected_list = 0
			docStyle2 = docStyle2.BorderForeground(lipgloss.Color("#ffffff"))
			docStyle = docStyle.BorderForeground(lipgloss.Color("#98971a"))
		}

		return m, nil

	case tickMsg:
		m.utilChart.Push(rand.Float64() * 100.0)
		m.utilChart.Draw()

		m.cpuUtilzations = retrieveCurrentCPUUtilization(12)
		m.utilChart2.Push(m.cpuUtilzations)
		m.utilChart2.Draw()
		// send next tick
		return m, tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		//_, v2 := list_item_style.GetFrameSize()
		m.list.SetSize((msg.Width-h)/2, (msg.Height-v)/2)
		m.utilChart.Resize((msg.Width-h)/2, (msg.Height-v)/3)
		m.utilChart2.Resize((msg.Width-h)/3, (msg.Height-v)/3)
		m.list_cpus.SetSize((msg.Width-h)/3, (msg.Height-v)/3)
		h3, _ := list_item_style.GetFrameSize()

		list_item_style = list_item_style.Width((msg.Width - h3) / 2)

		memoryStyle = memoryStyle.
			Height((msg.Height - v) / 3).
			Width((msg.Width - h) / 5)
	}

	m.utilChart, _ = m.utilChart.Update(msg)
	m.utilChart.DrawAll()

	m.utilChart2, _ = m.utilChart2.Update(msg)
	m.utilChart2.DrawAll()
	var cmd tea.Cmd
	if m.selected_list == 0 {
		m.list, cmd = m.list.Update(msg)
	} else if m.selected_list == 1 {
		m.list_cpus, cmd = m.list_cpus.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	s := docStyle.Render(m.list.View())
	w := utilHeaderStyle.Render("Utilization")
	t := docStyle2.Render(m.list_cpus.View())

	w += lipgloss.JoinVertical(lipgloss.Bottom, defaultStyle.Render(m.utilChart.View()))
	total_mem := memoryStyle.Render("Total Memory: " + strconv.Itoa(int(m.memory.Total)/int(math.Pow(1024, 3))) + " GB\n" +
		"Used Memory:  " + strconv.Itoa(int(m.memory.Available)/int(math.Pow(1024, 3))) + " GB")

	temp_view := lipgloss.JoinHorizontal(lipgloss.Top, total_mem, defaultStyle.Render(m.utilChart2.View()), zone.Mark("CPU", t))
	new_view := lipgloss.JoinHorizontal(lipgloss.Top, zone.Mark("Other List", s), w)
	final_view := lipgloss.JoinVertical(lipgloss.Left, new_view, temp_view)
	return zone.Scan(final_view)
}

func retrieveCurrentCPUUtilization(i int32) float64 {
	cpu, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Println("Error retrieving CPU utilization", err)
	}

	return cpu[i]
}

func main() {
	zone.NewGlobal()

	width := 60
	height := 26
	minYValue := 0.0
	maxYValue := 112.0

	cpuUtilizations := retrieveCurrentCPUUtilization(12)

	process, _ := process.Processes()
	mem, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("error")
	}
	items := []list.Item{}

	cpu_items := []list.Item{}

	cpu_iterable, err := cpu.Counts(true)
	if err != nil {
		fmt.Println("Error retrieving CPU info", err)
	}
	for i := range cpu_iterable {
		cpu_title := "Core: " + strconv.Itoa(i+1)
		cpu_items = append(cpu_items, item{title: cpu_title})
	}

	for _, proc := range process {
		process_name, _ := proc.Name()
		process_id := strconv.Itoa(int(proc.Pid))
		test := fmt.Sprintf("Process: %-35s  PID: %5s", process_name, process_id)
		items = append(items, item{title: test})
	}

	slc1 := streamlinechart.New(width, height)
	slc1.AxisStyle = axisStyle
	slc1.LabelStyle = labelStyle
	slc1.SetYRange(minYValue, maxYValue)                // set expected Y values (values can be less or greater than what is displayed)
	slc1.SetViewYRange(minYValue, maxYValue)            // setting display Y values will fail unless set expected Y values first
	slc1.SetStyles(runes.ArcLineStyle, graphLineStyle1) // graphLineStyle1 replaces linechart rune style
	slc1.Focus()

	util_chart := streamlinechart.New(width, height)
	util_chart.AxisStyle = axisStyle
	util_chart.LabelStyle = labelStyle
	util_chart.SetYRange(minYValue, maxYValue)                // set expected Y values (values can be less or greater than what is displayed)
	util_chart.SetViewYRange(minYValue, maxYValue)            // setting display Y values will fail unless set expected Y values first
	util_chart.SetStyles(runes.ArcLineStyle, graphLineStyle1) // graphLineStyle1 replaces linechart rune style
	util_chart.Focus()

	m := model{
		list: list.New(items, list.DefaultDelegate{Styles: list.DefaultItemStyles{NormalTitle: list_item_style}}, 0, 0), utilChart: slc1,
		memory: mem, cpuUtilzations: cpuUtilizations, utilChart2: util_chart,
		list_cpus:     list.New(cpu_items, list.DefaultDelegate{Styles: list.DefaultItemStyles{NormalTitle: list_item_style}}, 0, 0),
		selected_list: 0,
	}
	m.list.Title = "Active Processes"
	m.list_cpus.SetShowTitle(false)
	m.list_cpus.SetShowHelp(false)
	m.list_cpus.SetShowFilter(false)
	m.list_cpus.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.SetShowFilter(false)
	m.list.SetShowStatusBar(false)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
