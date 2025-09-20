package concourse

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Pipeline represents a Concourse pipeline
type Pipeline struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Paused   bool   `json:"paused"`
	Public   bool   `json:"public"`
	Archived bool   `json:"archived"`
	TeamName string `json:"team_name"`
	LastUpdatedUnix int64 `json:"last_updated"`
}

// GetLastUpdated returns the last updated time as a proper time.Time
func (p Pipeline) GetLastUpdated() time.Time {
	if p.LastUpdatedUnix == 0 {
		return time.Time{}
	}
	return time.Unix(p.LastUpdatedUnix, 0)
}

// Job represents a pipeline job
type Job struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PipelineName string `json:"pipeline_name"`
	PipelineID   int    `json:"pipeline_id"`
	TeamName     string `json:"team_name"`
	NextBuild    Build  `json:"next_build,omitempty"`
	FinishedBuild Build `json:"finished_build,omitempty"`
}

// Build represents a job build
type Build struct {
	ID         int    `json:"id"`
	TeamName   string `json:"team_name"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	JobName    string `json:"job_name"`
	APIURL     string `json:"api_url"`
	StartTimeUnix int64 `json:"start_time,omitempty"`
	EndTimeUnix   int64 `json:"end_time,omitempty"`
	PipelineID    int   `json:"pipeline_id"`
	PipelineName  string `json:"pipeline_name"`
}

// GetStartTime returns the start time as a proper time.Time
func (b Build) GetStartTime() time.Time {
	if b.StartTimeUnix == 0 {
		return time.Time{}
	}
	return time.Unix(b.StartTimeUnix, 0)
}

// GetEndTime returns the end time as a proper time.Time
func (b Build) GetEndTime() time.Time {
	if b.EndTimeUnix == 0 {
		return time.Time{}
	}
	return time.Unix(b.EndTimeUnix, 0)
}

// Resource represents a pipeline resource
type Resource struct {
	Name         string                 `json:"name"`
	PipelineName string                 `json:"pipeline_name"`
	TeamName     string                 `json:"team_name"`
	Type         string                 `json:"type"`
	LastCheckedUnix int64               `json:"last_checked,omitempty"`
	Version      map[string]interface{} `json:"version,omitempty"`
	Metadata     []Metadata             `json:"metadata,omitempty"`
}

// GetLastChecked returns the last checked time as a proper time.Time
func (r Resource) GetLastChecked() time.Time {
	if r.LastCheckedUnix == 0 {
		return time.Time{}
	}
	return time.Unix(r.LastCheckedUnix, 0)
}

// Metadata represents resource metadata
type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Team represents a Concourse team
type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Client wraps fly CLI operations
type Client struct {
	target string
}

// NewClient creates a new Concourse client for a specific target
func NewClient(target string) *Client {
	return &Client{target: target}
}

// GetTarget returns the target name
func (c *Client) GetTarget() string {
	return c.target
}

// execFly executes a fly command and returns the output
func (c *Client) execFly(args ...string) ([]byte, error) {
	if c.target != "" {
		args = append([]string{"-t", c.target}, args...)
	}
	
	cmd := exec.Command("fly", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fly command failed: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute fly command: %w", err)
	}
	
	return output, nil
}

// Login authenticates with the target
func (c *Client) Login(teamName, username, password string) error {
	args := []string{"login"}
	if teamName != "" {
		args = append(args, "-n", teamName)
	}
	if username != "" {
		args = append(args, "-u", username)
	}
	if password != "" {
		args = append(args, "-p", password)
	}
	
	_, err := c.execFly(args...)
	return err
}

// LoginInteractive performs interactive login (opens browser)
func (c *Client) LoginInteractive(apiURL, teamName string) error {
	args := []string{"login", "-c", apiURL}
	if teamName != "" {
		args = append(args, "-n", teamName)
	}
	
	if c.target != "" {
		args = append([]string{"-t", c.target}, args...)
	}
	
	// Execute interactively (this will open browser)
	cmd := exec.Command("fly", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Status checks if we're logged in to the target
func (c *Client) Status() (bool, error) {
	_, err := c.execFly("status")
	if err != nil {
		if strings.Contains(err.Error(), "not logged in") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetPipelines retrieves all pipelines
func (c *Client) GetPipelines() ([]Pipeline, error) {
	output, err := c.execFly("pipelines", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get pipelines: %w", err)
	}
	
	var pipelines []Pipeline
	if err := json.Unmarshal(output, &pipelines); err != nil {
		return nil, fmt.Errorf("failed to parse pipelines JSON: %w", err)
	}
	
	return pipelines, nil
}

// GetJobs retrieves jobs for a specific pipeline
func (c *Client) GetJobs(pipeline string) ([]Job, error) {
	output, err := c.execFly("jobs", "-p", pipeline, "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for pipeline %s: %w", pipeline, err)
	}
	
	var jobs []Job
	if err := json.Unmarshal(output, &jobs); err != nil {
		return nil, fmt.Errorf("failed to parse jobs JSON: %w", err)
	}
	
	return jobs, nil
}

// GetResources retrieves resources for a specific pipeline
func (c *Client) GetResources(pipeline string) ([]Resource, error) {
	output, err := c.execFly("resources", "-p", pipeline, "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for pipeline %s: %w", pipeline, err)
	}
	
	var resources []Resource
	if err := json.Unmarshal(output, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resources JSON: %w", err)
	}
	
	return resources, nil
}

// TriggerJob triggers a specific job
func (c *Client) TriggerJob(pipeline, job string) error {
	_, err := c.execFly("trigger-job", "-j", fmt.Sprintf("%s/%s", pipeline, job))
	if err != nil {
		return fmt.Errorf("failed to trigger job %s/%s: %w", pipeline, job, err)
	}
	return nil
}

// TriggerJobWithOutput triggers a job and returns success status and output
func (c *Client) TriggerJobWithOutput(pipeline, job string) (bool, string, error) {
	jobName := fmt.Sprintf("%s/%s", pipeline, job)
	
	// Use exec.Command directly to capture both success/failure cases
	cmd := exec.Command("fly", "-t", c.target, "trigger-job", "-j", jobName)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	
	if err != nil {
		// Check if it's a command execution error or job trigger failure
		if _, ok := err.(*exec.ExitError); ok {
			// fly command ran but returned non-zero exit code (e.g., job not found)
			return false, outputStr, nil
		}
		// Actual command execution error (e.g., fly not found)
		return false, outputStr, err
	}
	
	// Command succeeded - check if output indicates successful job trigger
	success := strings.Contains(strings.ToLower(outputStr), "started")
	return success, outputStr, nil
}

// RerunBuildWithOutput reruns a specific build and returns success status and output
func (c *Client) RerunBuildWithOutput(pipeline, job string, buildNumber int) (bool, string, error) {
	jobName := fmt.Sprintf("%s/%s", pipeline, job)
	buildStr := fmt.Sprintf("%d", buildNumber)
	
	// Use exec.Command directly to capture both success/failure cases
	cmd := exec.Command("fly", "-t", c.target, "rerun-build", "--job", jobName, "--build", buildStr)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	
	if err != nil {
		// Check if it's a command execution error or build rerun failure
		if _, ok := err.(*exec.ExitError); ok {
			// fly command ran but returned non-zero exit code (e.g., build not found)
			return false, outputStr, nil
		}
		// Actual command execution error (e.g., fly not found)
		return false, outputStr, err
	}
	
	// Command succeeded - check if output indicates successful build rerun
	success := strings.Contains(strings.ToLower(outputStr), "started")
	return success, outputStr, nil
}

// CheckResource triggers a check for a specific resource
func (c *Client) CheckResource(pipeline, resource string) error {
	_, err := c.execFly("check-resource", "-r", fmt.Sprintf("%s/%s", pipeline, resource))
	if err != nil {
		return fmt.Errorf("failed to check resource %s/%s: %w", pipeline, resource, err)
	}
	return nil
}

// CheckResourceWithOutput triggers a check for a specific resource and returns success status and output
func (c *Client) CheckResourceWithOutput(pipeline, resource string) (bool, string, error) {
	resourceName := fmt.Sprintf("%s/%s", pipeline, resource)
	
	// Use exec.Command directly to capture both success/failure cases
	cmd := exec.Command("fly", "-t", c.target, "check-resource", "-r", resourceName)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	
	if err != nil {
		// Check if it's a command execution error or resource check failure
		if _, ok := err.(*exec.ExitError); ok {
			// fly command ran but returned non-zero exit code (e.g., resource not found)
			return false, outputStr, nil
		}
		// Actual command execution error (e.g., fly not found)
		return false, outputStr, err
	}
	
	// Command succeeded - check if output indicates successful resource check
	success := strings.Contains(strings.ToLower(outputStr), "succeeded")
	return success, outputStr, nil
}

// UnpausePipeline unpauses a pipeline
func (c *Client) UnpausePipeline(pipeline string) error {
	_, err := c.execFly("unpause-pipeline", "-p", pipeline)
	if err != nil {
		return fmt.Errorf("failed to unpause pipeline %s: %w", pipeline, err)
	}
	return nil
}

// PausePipeline pauses a pipeline
func (c *Client) PausePipeline(pipeline string) error {
	_, err := c.execFly("pause-pipeline", "-p", pipeline)
	if err != nil {
		return fmt.Errorf("failed to pause pipeline %s: %w", pipeline, err)
	}
	return nil
}

// GetBuilds retrieves builds for a specific job
func (c *Client) GetBuilds(pipeline, job string, limit int) ([]Build, error) {
	args := []string{"builds", "-j", fmt.Sprintf("%s/%s", pipeline, job), "--json"}
	if limit > 0 {
		args = append(args, "--count", fmt.Sprintf("%d", limit))
	}
	
	output, err := c.execFly(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds for job %s/%s: %w", pipeline, job, err)
	}
	
	var builds []Build
	if err := json.Unmarshal(output, &builds); err != nil {
		return nil, fmt.Errorf("failed to parse builds JSON: %w", err)
	}
	
	return builds, nil
}

// GetTeams retrieves all teams
func (c *Client) GetTeams() ([]Team, error) {
	output, err := c.execFly("teams", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	
	var teams []Team
	if err := json.Unmarshal(output, &teams); err != nil {
		return nil, fmt.Errorf("failed to parse teams JSON: %w", err)
	}
	
	return teams, nil
}

// Sync syncs with the target (equivalent to fly sync)
func (c *Client) Sync() error {
	_, err := c.execFly("sync")
	return err
}

func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "not authorized") ||
		   strings.Contains(errorStr, "not logged in") ||
		   strings.Contains(errorStr, "unauthorized") ||
		   strings.Contains(errorStr, "authentication")
}