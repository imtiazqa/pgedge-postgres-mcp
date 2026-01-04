package suite

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pgedge/AItoolsFramework/common/config"
	"github.com/pgedge/AItoolsFramework/common/executor"
	"github.com/stretchr/testify/suite"
)

// TestResult tracks individual test execution
type TestResult struct {
	Name      string
	Status    string // PASS, FAIL, SKIP
	Duration  time.Duration
	Error     error
	StartTime time.Time
}

// BaseSuite provides common test suite functionality
type BaseSuite struct {
	suite.Suite

	// Configuration
	Config *config.TestConfig

	// Execution context
	Ctx context.Context

	// Executor for running commands
	Executor executor.Executor

	// Test tracking
	StartTime time.Time
	Results   []TestResult

	// Current test tracking
	currentTestStart time.Time
}

// SetupSuite runs once before all tests
func (s *BaseSuite) SetupSuite() {
	s.Ctx = context.Background()
	s.StartTime = time.Now()
	s.Results = make([]TestResult, 0)

	// Load configuration
	s.loadConfig()

	// Initialize executor
	s.initExecutor()

	s.T().Logf("Test suite started at %s", s.StartTime.Format(time.RFC3339))
}

// SetupTest runs before each test
func (s *BaseSuite) SetupTest() {
	s.currentTestStart = time.Now()
	result := TestResult{
		Name:      s.T().Name(),
		StartTime: s.currentTestStart,
		Status:    "RUNNING",
	}
	s.Results = append(s.Results, result)

	s.T().Logf("Test started: %s", s.T().Name())
}

// TearDownTest runs after each test
func (s *BaseSuite) TearDownTest() {
	// Update test result
	idx := len(s.Results) - 1
	if idx >= 0 {
		s.Results[idx].Duration = time.Since(s.currentTestStart)
		if s.T().Failed() {
			s.Results[idx].Status = "FAIL"
		} else if s.T().Skipped() {
			s.Results[idx].Status = "SKIP"
		} else {
			s.Results[idx].Status = "PASS"
		}

		s.T().Logf("Test %s: %s (duration: %s)",
			s.Results[idx].Status,
			s.Results[idx].Name,
			s.Results[idx].Duration)
	}
}

// TearDownSuite runs once after all tests
func (s *BaseSuite) TearDownSuite() {
	// Clean up executor
	if s.Executor != nil {
		if err := s.Executor.Cleanup(s.Ctx); err != nil {
			s.T().Logf("Warning: executor cleanup error: %v", err)
		}
	}

	// Print summary
	s.printSummary()
}

// Helper methods

// ExecCommand is a helper for executing commands with default timeout
func (s *BaseSuite) ExecCommand(cmd string) (string, int, error) {
	timeout := s.Config.Timeouts.Command
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	ctx, cancel := context.WithTimeout(s.Ctx, timeout)
	defer cancel()

	return s.Executor.Exec(ctx, cmd)
}

// Eventually retries a condition until it succeeds or times out
func (s *BaseSuite) Eventually(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// WaitFor waits for a condition to be true
func (s *BaseSuite) WaitFor(condition func() bool, timeout time.Duration) error {
	if s.Eventually(condition, timeout, 1*time.Second) {
		return nil
	}
	return fmt.Errorf("condition not met within %s", timeout)
}

// Private methods

// loadConfig loads the test configuration
func (s *BaseSuite) loadConfig() {
	configPath := config.GetConfigPath()

	loader := config.NewLoader()
	cfg, err := loader.Load(s.Ctx, configPath)
	if err != nil {
		s.T().Fatalf("Failed to load config from %s: %v", configPath, err)
	}

	// Set defaults
	validator := config.NewValidator()
	validator.SetDefaults(cfg)

	s.Config = cfg

	s.T().Logf("Loaded configuration from: %s (environment: %s)", configPath, s.Config.Environment)
}

// initExecutor initializes the executor based on configuration
func (s *BaseSuite) initExecutor() {
	mode, err := executor.ParseMode(s.Config.Execution.Mode)
	if err != nil {
		s.T().Fatalf("Invalid execution mode: %v", err)
	}

	// Determine OS image - prefer Container.OSImage, fallback to legacy OSImage
	osImage := s.Config.Execution.Container.OSImage
	if osImage == "" {
		osImage = s.Config.Execution.OSImage
	}

	// Determine skip sudo check - container mode takes precedence
	skipSudoCheck := s.Config.Execution.SkipSudoCheck
	if mode == executor.ModeContainerSystemd {
		skipSudoCheck = s.Config.Execution.Container.SkipSudoCheck
	}

	opts := &executor.ExecutorOptions{
		Timeout:       s.Config.Timeouts.Command,
		LogCommands:   s.Config.Reporting.LogLevel == "detailed",
		LogOutput:     s.Config.Reporting.LogLevel == "detailed",
		SkipSudoCheck: skipSudoCheck,
	}

	exec, err := executor.NewExecutor(mode, osImage, opts)
	if err != nil {
		s.T().Fatalf("Failed to create executor: %v", err)
	}

	// Start the executor
	if err := exec.Start(s.Ctx); err != nil {
		s.T().Fatalf("Failed to start executor: %v", err)
	}

	s.Executor = exec

	s.T().Logf("Executor initialized: %s", mode.String())
	if mode == executor.ModeContainerSystemd {
		s.T().Logf("Container image: %s (systemd: %v)", osImage, s.Config.Execution.Container.UseSystemd)
	}
}

// printSummary prints a test execution summary with beautiful tables
func (s *BaseSuite) printSummary() {
	totalDuration := time.Since(s.StartTime)

	// Count passes, failures, and skips
	passed := 0
	failed := 0
	skipped := 0

	for _, result := range s.Results {
		switch result.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "SKIP":
			skipped++
		}
	}

	// Create the summary table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Use ColoredBright style
	t.SetStyle(table.StyleColoredBright)

	// Fix footer visibility by customizing colors
	style := t.Style()
	style.Color.Footer = text.Colors{text.BgHiCyan, text.FgBlack}
	t.SetStyle(*style)

	// Configure title
	t.SetTitle("🧪 Test Suite Summary")

	// Add headers
	t.AppendHeader(table.Row{"#", "Test Name", "Status", "Duration"})

	// Configure column alignments
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight}, // # column
		{Number: 2, Align: text.AlignLeft},  // Test Name
		{Number: 3, Align: text.AlignLeft},  // Status
		{Number: 4, Align: text.AlignRight}, // Duration
	})

	// Add test results
	for i, result := range s.Results {
		// Clean up test name - remove suite prefix
		testName := result.Name
		// Try to strip common test suite prefixes
		if idx := strings.LastIndex(testName, "/"); idx != -1 {
			testName = testName[idx+1:]
		}

		var status string
		// Use simpler status format in CI to avoid rendering issues
		if os.Getenv("CI") != "" {
			switch result.Status {
			case "PASS":
				status = "✓ PASS"
			case "FAIL":
				status = "✗ FAIL"
			case "SKIP":
				status = "○ SKIP"
			default:
				status = fmt.Sprintf("⚠ %s", result.Status)
			}
		} else {
			switch result.Status {
			case "PASS":
				status = text.FgGreen.Sprintf("✓ PASS")
			case "FAIL":
				status = text.FgRed.Sprintf("✗ FAIL")
			case "SKIP":
				status = text.FgYellow.Sprintf("○ SKIP")
			default:
				status = text.FgYellow.Sprintf("⚠ %s", result.Status)
			}
		}

		// Format duration consistently
		durationStr := formatDuration(result.Duration)
		t.AppendRow(table.Row{i + 1, testName, status, durationStr})
	}

	// Add separator before footer
	t.AppendSeparator()

	// Add footer with totals
	totalTests := len(s.Results)
	var statusSummary string
	if failed > 0 {
		statusSummary = fmt.Sprintf("%d PASSED, %d FAILED, %d SKIPPED", passed, failed, skipped)
	} else if skipped > 0 {
		statusSummary = fmt.Sprintf("%d PASSED, %d SKIPPED ✨", passed, skipped)
	} else {
		statusSummary = fmt.Sprintf("%d/%d PASSED ✨", passed, totalTests)
	}

	totalDurationStr := formatDuration(totalDuration)
	t.AppendFooter(table.Row{"", fmt.Sprintf("TOTAL: %d tests", totalTests), statusSummary, totalDurationStr})

	// Print execution context before table
	fmt.Println()
	fmt.Printf("📋 Execution Mode: %s\n", text.FgCyan.Sprint(s.Config.Execution.Mode))
	if s.Config.Execution.Mode == "container" || s.Config.Execution.Mode == "container-systemd" {
		osImage := s.Config.Execution.Container.OSImage
		if osImage == "" {
			osImage = s.Config.Execution.OSImage
		}
		fmt.Printf("🐳 OS Image: %s\n", text.FgCyan.Sprint(osImage))
	}
	fmt.Printf("🌍 Environment: %s\n", text.FgCyan.Sprint(s.Config.Environment))
	fmt.Printf("⏱️  Total Duration: %s\n", text.FgCyan.Sprint(totalDuration.Round(time.Millisecond)))

	// Print table after context
	fmt.Println()
	t.Render()
	fmt.Println()
}

// formatDuration formats a duration with consistent width for table alignment
func formatDuration(d time.Duration) string {
	// Always show as seconds with 3 decimal places for consistency
	seconds := float64(d) / float64(time.Second)
	return fmt.Sprintf("%.3fs", seconds)
}
