package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aliasproject/servermetrics"
)

// Version is the agent's own version, e.g. "v1.2.0" -- overridden at build
// time via -ldflags "-X main.Version=vX.Y.Z" by the release workflow.
// Left as "dev" for local/manual builds, which Alios treats as "unknown"
// rather than matching (or claiming to be behind) any real release.
var Version = "dev"

// Metric represents the structure of the metric to send
type Metric struct {
	Version    string                         `json:"version"`
	CPU        servermetrics.CPUStats         `json:"cpu"`
	Memory     servermetrics.MemoryStats      `json:"memory"`
	Disk       servermetrics.DiskStats        `json:"disk"`
	Containers []servermetrics.ContainerStats `json:"containers"`
	Timestamp  int64                          `json:"timestamp"`
}

func main() {
	// Confirm Arguments
	if len(os.Args) < 2 {
		fmt.Println("Error: Usage: alios-monitor <report-url>")
		os.Exit(1)
	}

	// Allow checking the installed version directly over SSH
	// (`/opt/alios-monitor --version`) without waiting on a metrics report.
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Println(Version)
		return
	}

	// Previous stats for comparison
	var prevCPUStats servermetrics.CPUStats
	var err error

	// Initialize previous stats
	prevCPUStats, err = servermetrics.GetCPUStats()
	if err != nil {
		fmt.Println("Error getting initial CPU stats:", err)
		return
	}

	// Create a ticker that ticks every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Collect new CPU stats
		currentCPUStats, err := servermetrics.GetCPUStats()
		if err != nil {
			fmt.Println("Error collecting CPU stats:", err)
			continue
		}

		// Collect memory stats
		memoryStats, err := servermetrics.GetMemoryStats()
		if err != nil {
			log.Println("Error collecting memory stats:", err)
			continue
		}

		// Collect disk stats
		diskStats, err := servermetrics.GetDiskStats("/")
		if err != nil {
			log.Println("Error collecting disk stats:", err)
			continue
		}

		// Collect container stats
		containerStats, err := servermetrics.GetContainerStats()
		if err != nil {
			log.Println("Error collecting container stats:", err)
			continue
		}

		// Calculate CPU usage based on the difference between current and previous stats
		cpuUsage := servermetrics.CalculateCPUUsage(prevCPUStats, currentCPUStats)

		// Update previous stats
		prevCPUStats = currentCPUStats

		// Create a metric struct
		metric := Metric{
			Version:    Version,
			CPU:        cpuUsage,
			Memory:     memoryStats,
			Disk:       diskStats,
			Containers: containerStats,
			Timestamp:  time.Now().Unix(),
		}

		// Send the metric to the remote server
		err = sendMetric(metric)
		if err != nil {
			fmt.Println("Error sending metric:", err)
		} else {
			fmt.Println("Metric sent successfully")
		}
	}
}

// sendMetric sends the metric to the remote server
func sendMetric(metric Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	resp, err := http.Post(os.Args[1], "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	return nil
}
