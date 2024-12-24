package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/process"
)

type model struct {
	processes []int32
}

func main() {
	procs, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}
	test := procs[0].Pid
	fmt.Printf("%d", test)
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		processes: retrievePID(),

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
	}
}

func retrievePID() []int32 {
	p := []int32{}
	procs, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}
	for _, process := range procs {
		p = append(p, process.Pid)
		fmt.Printf("%d\n", process.Pid)
	}
	return p
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		// Return the updated model to the Bubble Tea runtime for processing.
		// Note that we're not returning a command.
	}
	return m, nil
}

func (m model) View() string {
	// The header
	s := "What should we buy at the market?\n\n"

	// The footer
	s += "\nPress q to quit.\n"

	for _, pids := range m.processes {
		s += strconv.Itoa(int(pids))
		s += "\n"
	}
	// Send the UI for rendering
	return s
}
