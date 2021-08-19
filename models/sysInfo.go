package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AvgCPUUsage struct {
	Avg       string `json:"avg"`
	Timestamp string `json:"timestamp"`
}

func CalcAvgCPUFromTop(topOutput string) (AvgCPUUsage, error) {
	var cpuUse float32
	var cpuInfo AvgCPUUsage
	lines := strings.Split(topOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "%Cpu(s):") {
			atributes := strings.Split(line, ",")
			idle, err := strconv.ParseFloat(strings.Trim((atributes[3][:5]), " "), 32)
			if err != nil {
				return cpuInfo, err
			}
			cpuUse = 100 - float32(idle)
		}
	}
	return AvgCPUUsage{fmt.Sprintf("%.1f", cpuUse), time.Now().Format("01-02-2006 15:04:05")}, nil
}
