package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PrometheusTarget represents a Prometheus scrape target
type PrometheusTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// TargetManager manages Prometheus file-based service discovery targets
type TargetManager struct {
	targetsDir string
}

// NewTargetManager creates a new Prometheus target manager
func NewTargetManager(targetsDir string) *TargetManager {
	return &TargetManager{
		targetsDir: targetsDir,
	}
}

// RegisterInstance registers an instance as a Prometheus scrape target
func (tm *TargetManager) RegisterInstance(instanceIP, instanceID, serviceID, projectID, serviceName string) error {
	// Ensure targets directory exists
	if err := os.MkdirAll(tm.targetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create targets directory: %w", err)
	}

	// Create target file (one file per instance for easy management)
	targetFile := filepath.Join(tm.targetsDir, fmt.Sprintf("%s.json", instanceID))

	target := PrometheusTarget{
		Targets: []string{
			fmt.Sprintf("%s:9100", instanceIP), // Node Exporter
			fmt.Sprintf("%s:8080", instanceIP), // cAdvisor (optional)
		},
		Labels: map[string]string{
			"instance_id": instanceID,
			"service_id":   serviceID,
			"project_id":   projectID,
			"service_name": serviceName,
			"job":          "click-deploy-instances",
		},
	}

	data, err := json.MarshalIndent([]PrometheusTarget{target}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal target: %w", err)
	}

	if err := os.WriteFile(targetFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	return nil
}

// UnregisterInstance removes an instance from Prometheus targets
func (tm *TargetManager) UnregisterInstance(instanceID string) error {
	targetFile := filepath.Join(tm.targetsDir, fmt.Sprintf("%s.json", instanceID))
	
	if err := os.Remove(targetFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove target file: %w", err)
	}
	
	return nil
}

// RegisterDatabase registers a database instance as a Prometheus scrape target
func (tm *TargetManager) RegisterDatabase(instanceIP, instanceID, databaseID, projectID, databaseName, engine string) error {
	if err := os.MkdirAll(tm.targetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create targets directory: %w", err)
	}

	targetFile := filepath.Join(tm.targetsDir, fmt.Sprintf("db-%s.json", databaseID))

	target := PrometheusTarget{
		Targets: []string{
			fmt.Sprintf("%s:9100", instanceIP), // Node Exporter
		},
		Labels: map[string]string{
			"instance_id":  instanceID,
			"database_id":  databaseID,
			"project_id":   projectID,
			"database_name": databaseName,
			"engine":       engine,
			"job":          "click-deploy-databases",
		},
	}

	data, err := json.MarshalIndent([]PrometheusTarget{target}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal target: %w", err)
	}

	if err := os.WriteFile(targetFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	return nil
}

// UnregisterDatabase removes a database from Prometheus targets
func (tm *TargetManager) UnregisterDatabase(databaseID string) error {
	targetFile := filepath.Join(tm.targetsDir, fmt.Sprintf("db-%s.json", databaseID))
	
	if err := os.Remove(targetFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove target file: %w", err)
	}
	
	return nil
}

