package main

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	procs, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			// Maybe the process ended or we lack permissions
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		// Always check if memInfo is nil
		if memInfo == nil {
			continue
		}

		fmt.Printf("PID: %d | Name: %s | CPU: %.2f%% | RSS: %v bytes\n",
			p.Pid, name, cpuPercent, memInfo.RSS)
	}
}
