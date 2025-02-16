package main

import (
	"html/template"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/influxdata/tdigest"
)

func TestTDigestCompressionHTMLReport(t *testing.T) {
	type ReportRow struct {
		Compression  int
		TotalAllocMB float64
		PeakAllocMB  float64
		ExecTimeMS   float64
	}

	var reportData []ReportRow

	for compression := 100; compression <= 1000; compression += 100 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		allocBefore := m.TotalAlloc
		heapAllocBefore := m.HeapAlloc

		startTime := time.Now()

		td := tdigest.NewWithCompression(float64(compression))
		for j := 0; j < 1_000_000; j++ {
			value := rand.Float64() * 1000
			td.Add(value, 1) // Adding random values between 0 and 1000
		}

		execTime := time.Since(startTime).Milliseconds()

		runtime.ReadMemStats(&m)
		allocAfter := m.TotalAlloc
		heapAllocAfter := m.HeapAlloc
		totalAlloc := allocAfter - allocBefore
		peakAlloc := heapAllocAfter - heapAllocBefore

		totalAllocMB := float64(totalAlloc) / (1024 * 1024)
		peakAllocMB := float64(peakAlloc) / (1024 * 1024)

		reportData = append(reportData, ReportRow{
			Compression:  compression,
			TotalAllocMB: totalAllocMB,
			PeakAllocMB:  peakAllocMB,
			ExecTimeMS:   float64(execTime),
		})
	}

	tmpl := template.Must(template.New("report").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>TDigest Compression Report</title>
		</head>
		<body>
			<h1>TDigest Compression Report</h1>
			<table border="1">
				<tr>
					<th>Compression</th>
					<th>Total Memory Allocated (MB)</th>
					<th>Peak Memory Allocated (MB)</th>
					<th>Execution Time (ms)</th>
				</tr>
				{{range .}}
				<tr>
					<td>{{.Compression}}</td>
					<td>{{.TotalAllocMB}}</td>
					<td>{{.PeakAllocMB}}</td>
					<td>{{.ExecTimeMS}}</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
	`))

	file, err := os.Create("report.html")
	if err != nil {
		t.Fatalf("Failed to create report file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, reportData)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}
}
