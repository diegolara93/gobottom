package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

var docStyle = lipgloss.NewStyle().Margin(0, 0).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#98971a"))
var docStyle2 = lipgloss.NewStyle().Margin(2, 0, 0, 0).Border(lipgloss.NormalBorder()).Width(20).BorderForeground(lipgloss.Color("#ebdbb2"))

var colorMargin = lipgloss.NewStyle().Margin(1)

var textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c6f64"))

type tickMsg time.Time

var list_item_style = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7c6f64")).
	Height(1).
	Margin(0, 0)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#ebdbb2")).
	Margin(0)

var utilHeaderStyle = lipgloss.NewStyle().
	MarginTop(0).MarginLeft(0).
	Background(lipgloss.Color("#cc241d"))

var memoryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("ebdbb2")).
	PaddingTop(0).
	MarginTop(0).
	PaddingBottom(0).
	BorderStyle(lipgloss.NormalBorder()).
	Width(46).
	Height(18).Align(lipgloss.Center)

var graphLineStyle1 = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#cc241d")) // red

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#98971a")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#458588")) // cyan

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
	selected_cpu   int32
	disks          viewport.Model
	hostInfo       viewport.Model
	networkChart   streamlinechart.Model
	networkStats   []net.IOCountersStat
	prevNetStats   []net.IOCountersStat
	lastNetTime    time.Time
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
			docStyle = docStyle.BorderForeground(lipgloss.Color("#ebdbb2"))
		} else if zone.Get("Other List").InBounds(msg) {
			m.selected_list = 0
			docStyle2 = docStyle2.BorderForeground(lipgloss.Color("#ebdbb2"))
			docStyle = docStyle.BorderForeground(lipgloss.Color("#98971a"))
		}

		return m, nil

	case tickMsg:
		mem, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("error getting ram", err)
		}
		m.utilChart.Push(mem.UsedPercent)
		m.utilChart.Draw()

		if m.selected_cpu != int32(m.list_cpus.Index()) {
			m.selected_cpu = int32(m.list_cpus.Index())
			m.utilChart2.ClearAllData()
			m.cpuUtilzations = retrieveCurrentCPUUtilization(m.selected_cpu)
			m.utilChart2.Push(m.cpuUtilzations)
			m.utilChart2.Draw()
		} else {
			m.cpuUtilzations = retrieveCurrentCPUUtilization(m.selected_cpu)
			m.utilChart2.Push(m.cpuUtilzations)
			m.utilChart2.Draw()
		}

		now := time.Now()
		timeDelta := now.Sub(m.lastNetTime).Seconds()

		netStats, err := net.IOCounters(true)
		if err != nil {
			fmt.Println("error getting network stats", err)
		}

		if len(netStats) > 0 && len(m.prevNetStats) > 0 && timeDelta > 0 {
			var totalBandwidth float64
			for i, stat := range netStats {
				if i < len(m.prevNetStats) {
					sent := float64(stat.BytesSent-m.prevNetStats[i].BytesSent) / timeDelta / 1024
					recv := float64(stat.BytesRecv-m.prevNetStats[i].BytesRecv) / timeDelta / 1024
					totalBandwidth += sent + recv
				}
			}
			m.networkChart.Push(totalBandwidth)
			m.networkChart.Draw()
		}

		m.prevNetStats = netStats
		m.lastNetTime = now

		// send next tick
		return m, tea.Tick(time.Second/6, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		//_, v2 := list_item_style.GetFrameSize()
		m.list.SetSize((msg.Width-h)/2, (msg.Height-v)/3+v)
		m.utilChart.Resize((msg.Width-h)/3, (msg.Height-v)/3+v/2)
		m.utilChart2.Resize((msg.Width-h)/3, (msg.Height-v)/3+v/2)
		m.list_cpus.SetSize((msg.Width-h)/3, (msg.Height-v)/3+v)
		m.disks.Height = (msg.Height-v)/3 + v
		m.disks.Width = (msg.Width - h) / 7
		m.hostInfo.Height = (msg.Height-v)/3 + v
		m.hostInfo.Width = (msg.Width - h) / 4
		m.networkChart.Resize((msg.Width-h)/3, (msg.Height-v)/3+v/2)
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

	m.networkChart, _ = m.networkChart.Update(msg)
	m.networkChart.DrawAll()

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
	t := docStyle2.Render(m.list_cpus.View())

	ramChartHeader := utilHeaderStyle.Render("Ram Usage: ")
	cpuUtilHeader := utilHeaderStyle.Render("Core Utilization: ")
	networkChartHeader := utilHeaderStyle.Render("Network Traffic (KB/s): ")

	w := defaultStyle.Render("  " + ramChartHeader + "\n" + m.utilChart.View())
	//total_mem := memoryStyle.Render("Total Memory: " + strconv.Itoa(int(m.memory.Total)/int(math.Pow(1024, 3))) + " GB\n" +
	//	"Used Memory:  " + strconv.Itoa(int(m.memory.Available)/int(math.Pow(1024, 3))) + " GB")

	colors := func() string {
		colors := colorGrid((m.disks.Width/2)+7, m.list.Height())

		b := strings.Builder{}
		for _, x := range colors {
			for _, y := range x {
				s := lipgloss.NewStyle().SetString("  ").Background(lipgloss.Color(y))
				b.WriteString(s.String())
			}
			b.WriteRune('\n')
		}

		return b.String()
	}()

	d := defaultStyle.Render(m.disks.View())
	j := defaultStyle.Render(m.hostInfo.View())
	nc := defaultStyle.Render("  " + networkChartHeader + "\n" + m.networkChart.View())

	temp_view := lipgloss.JoinHorizontal(lipgloss.Bottom,
		defaultStyle.Render("  "+cpuUtilHeader+"\n"+m.utilChart2.View()),
		zone.Mark("CPU", t),
		j,
		nc)
	new_view := lipgloss.JoinHorizontal(lipgloss.Top, w, zone.Mark("Other List", s), d, colorMargin.Render(colors))
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
	height := 30
	minYValue := 0.0
	maxYValue := 112.0

	cpuUtilizations := retrieveCurrentCPUUtilization(0)

	disks, err := disk.Partitions(false)
	if err != nil {
		fmt.Println("error getting disks", err)
	}

	process, err := process.Processes()
	if err != nil {
		fmt.Println("error getting processes", err)
	}

	mem, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("error getting ram", err)
	}

	host, err := host.Info()
	if err != nil {
		fmt.Println("error getting host info", err)
	}

	networkStats, err := net.IOCounters(true)
	if err != nil {
		fmt.Println("error getting network info", err)
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
	slc1.SetYRange(minYValue, maxYValue)
	slc1.SetViewYRange(minYValue, maxYValue)
	slc1.SetStyles(runes.ArcLineStyle, graphLineStyle1)
	slc1.Focus()

	util_chart := streamlinechart.New(width, height)
	util_chart.AxisStyle = axisStyle
	util_chart.LabelStyle = labelStyle
	util_chart.SetYRange(minYValue, maxYValue)
	util_chart.SetViewYRange(minYValue, maxYValue)
	util_chart.SetStyles(runes.ArcLineStyle, graphLineStyle1)
	util_chart.Focus()

	network_chart := streamlinechart.New(width, height)
	network_chart.AxisStyle = axisStyle
	network_chart.LabelStyle = labelStyle
	network_chart.SetYRange(0, 1000)
	network_chart.SetViewYRange(0, 1000)
	network_chart.SetStyles(runes.ArcLineStyle, graphLineStyle1)
	network_chart.Focus()

	vp := viewport.New(30, 30)
	vp2 := viewport.New(30, 30)

	m := model{
		list:           list.New(items, list.DefaultDelegate{Styles: list.DefaultItemStyles{NormalTitle: list_item_style}}, 0, 0),
		utilChart:      slc1,
		memory:         mem,
		cpuUtilzations: cpuUtilizations,
		utilChart2:     util_chart,
		list_cpus:      list.New(cpu_items, list.DefaultDelegate{Styles: list.DefaultItemStyles{NormalTitle: list_item_style}}, 0, 0),
		selected_list:  0,
		selected_cpu:   0,
		disks:          vp,
		hostInfo:       vp2,
		networkStats:   networkStats,
		prevNetStats:   networkStats,
		lastNetTime:    time.Now(),
		networkChart:   network_chart,
	}
	m.list.Title = "Active Processes"
	m.list.Styles.Title = utilHeaderStyle
	m.list_cpus.Title = "CPU Cores"
	m.list_cpus.Styles.Title = utilHeaderStyle
	m.list_cpus.SetShowTitle(true)
	m.list_cpus.SetShowHelp(false)
	m.list_cpus.SetShowFilter(false)
	m.list_cpus.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.SetShowFilter(false)
	m.list.SetShowStatusBar(false)

	m.disks.SetContent(textStyle.Render("Memory Drives: "+strconv.Itoa(len(disks))) +
		"\n" +
		printDisks(disks))

	m.hostInfo.SetContent(textStyle.Render(printHostInfo(host)))

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func printHostInfo(host *host.InfoStat) string {
	str := ""
	str += "Operating System: " + host.Platform + "\nCPU Instruction Set: " + host.KernelArch + "\nUptime: " + strconv.FormatFloat(float64(host.Uptime)/3600.0, 'f', 2, 64) + " Hours"
	return textStyle.Render(str)
}

func printDisks(disks []disk.PartitionStat) string {
	str := ""
	for i := range disks {
		currDisk := disks[i].Mountpoint
		storage, err := disk.Usage(currDisk)
		if err != nil {
			fmt.Println("error retrieving storage", err)
		}
		freeStorage := BytesToTB(storage.Total)
		usedStorage := BytesToTB(storage.Free)
		str += "Drive " + currDisk + " " + strconv.FormatFloat(usedStorage, 'f', 2, 64) + "/" + strconv.FormatFloat(freeStorage, 'f', 2, 64) + "TB\n"
	}
	return textStyle.Render(str)
}

func BytesToTB(bytes uint64) float64 {
	const tb = 1024 * 1024 * 1024 * 1024 // 1 TB = 2^40 bytes
	return float64(bytes) / float64(tb)
}

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#d79921")
	x1y0, _ := colorful.Hex("#cc241d")
	x0y1, _ := colorful.Hex("#458588")
	x1y1, _ := colorful.Hex("#b16286")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}
