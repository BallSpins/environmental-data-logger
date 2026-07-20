package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Metadata menyimpan informasi mesin dan environment eksekusi benchmark
type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	GoVersion string    `json:"go_version"`
	NumCPU    int       `json:"num_cpu"`
	RedisVer  string    `json:"redis_version,omitempty"`
	MySQLVer  string    `json:"mysql_version,omitempty"`
}

// ResultPayload merepresentasikan output metrik ringkas terstruktur per eksperimen
type ResultPayload struct {
	Experiment  string                 `json:"experiment"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metrics     map[string]interface{} `json:"metrics"`
	Environment Metadata               `json:"environment"`
}

// SaveMetadata menulis metadata.json ke dalam target output folder
func SaveMetadata(outputDir string, redisVer, mysqlVer string) error {
	meta := Metadata{
		Timestamp: time.Now(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
		RedisVer:  redisVer,
		MySQLVer:  mysqlVer,
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, "metadata.json"), bytes, 0644)
}

// ExportResult menyimpan summary metrics terstruktur ke format JSON, CSV, dan Markdown
type ExportData struct {
	Experiment string
	Params     map[string]interface{}
	Metrics    map[string]float64
}

func SaveResults(outputDir string, expData ExportData, redisVer, mysqlVer string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	meta := Metadata{
		Timestamp: time.Now(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
		RedisVer:  redisVer,
		MySQLVer:  mysqlVer,
	}

	// 1. Save JSON
	resPayload := ResultPayload{
		Experiment:  expData.Experiment,
		Parameters:  expData.Params,
		Metrics:     make(map[string]interface{}),
		Environment: meta,
	}
	for k, v := range expData.Metrics {
		resPayload.Metrics[k] = v
	}

	jsonBytes, err := json.MarshalIndent(resPayload, "", "  ")
	if err != nil {
		return err
	}
	jsonFilename := fmt.Sprintf("result-%s.json", expData.Experiment)
	err = os.WriteFile(filepath.Join(outputDir, jsonFilename), jsonBytes, 0644)
	if err != nil {
		return err
	}

	// 2. Save CSV (Excel Friendly)
	csvFilename := fmt.Sprintf("result-%s.csv", expData.Experiment)
	csvFile, err := os.Create(filepath.Join(outputDir, csvFilename))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Header CSV
	csvFile.WriteString("Timestamp,Experiment,Metric,Value\n")
	timestampStr := meta.Timestamp.Format(time.RFC3339)
	for k, v := range expData.Metrics {
		csvLine := fmt.Sprintf("%s,%s,%s,%.4f\n", timestampStr, expData.Experiment, k, v)
		csvFile.WriteString(csvLine)
	}

	// 3. Save Markdown
	mdFilename := fmt.Sprintf("result-%s.md", expData.Experiment)
	mdFile, err := os.Create(filepath.Join(outputDir, mdFilename))
	if err != nil {
		return err
	}
	defer mdFile.Close()

	mdFile.WriteString(fmt.Sprintf("# Benchmark Result: %s\n\n", expData.Experiment))
	mdFile.WriteString(fmt.Sprintf("* **Executed At**: %s\n", timestampStr))
	mdFile.WriteString(fmt.Sprintf("* **OS**: %s / %s\n", meta.OS, meta.Arch))
	mdFile.WriteString(fmt.Sprintf("* **Go Version**: %s\n\n", meta.GoVersion))

	mdFile.WriteString("## Configured Parameters\n\n")
	mdFile.WriteString("| Parameter | Value |\n")
	mdFile.WriteString("| --- | --- |\n")
	for k, v := range expData.Params {
		mdFile.WriteString(fmt.Sprintf("| %s | %v |\n", k, v))
	}
	mdFile.WriteString("\n")

	mdFile.WriteString("## Measured Metrics\n\n")
	mdFile.WriteString("| Metric Name | Value |\n")
	mdFile.WriteString("| --- | --- |\n")
	for k, v := range expData.Metrics {
		mdFile.WriteString(fmt.Sprintf("| %s | %.4f |\n", k, v))
	}
	mdFile.WriteString("\n")

	return nil
}
