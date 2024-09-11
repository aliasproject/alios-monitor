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

// Metric represents the structure of the metric to send
type Metric struct {
	CPU       servermetrics.CPUStats    `json:"cpu"`
	Memory    servermetrics.MemoryStats `json:"memory"`
	Disk      servermetrics.DiskStats   `json:"disk"`
	Timestamp int64                     `json:"timestamp"`
}

func main() {
	// Confirm Arguments
	if len(os.Args) < 2 {
		fmt.Println("Error: Usage: alios-monitor <report-url>")
		os.Exit(1)
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

	for {
		select {
		case <-ticker.C:
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

			// Calculate CPU usage based on the difference between current and previous stats
			cpuUsage := servermetrics.CalculateCPUUsage(prevCPUStats, currentCPUStats)

			// Update previous stats
			prevCPUStats = currentCPUStats

			// Create a metric struct
			metric := Metric{
				CPU:       cpuUsage,
				Memory:    memoryStats,
				Disk:      diskStats,
				Timestamp: time.Now().Unix(),
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
