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
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskUsedPct float64 `json:"disk_used_pct"`
	Timestamp   int64   `json:"timestamp"`
}

func main() {
	// Previous stats for comparison
	var prevStats servermetrics.CPUStats
	var err error

	// Initialize previous stats
	prevStats, err = servermetrics.GetCPUStats()
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
			currentStats, err := servermetrics.GetCPUStats()
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
			cpuUsage := servermetrics.CalculateCPUUsage(prevStats, currentStats)

			// Update previous stats
			prevStats = currentStats

			// Create a metric struct
			metric := Metric{
				CPUUsage:    cpuUsage,
				MemoryTotal: memoryStats.Total,
				MemoryUsed:  memoryStats.Used,
				DiskTotal:   diskStats.Total,
				DiskUsed:    diskStats.Used,
				DiskUsedPct: diskStats.UsedPct,
				Timestamp:   time.Now().Unix(),
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



    fmt.Println(os.Args[1])

	resp, err := http.Post("https://alios.test/metrics", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	return nil
}
