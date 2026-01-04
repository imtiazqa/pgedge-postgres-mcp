package suite

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// printSummary prints a test execution summary
func (s *BaseSuite) printSummary() {
	totalDuration := time.Since(s.StartTime)

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

	s.T().Logf("\n" + strings.Repeat("=", 70))
	s.T().Logf("TEST SUITE SUMMARY")
	s.T().Logf(strings.Repeat("=", 70))
	s.T().Logf("Total Tests:   %d", len(s.Results))
	s.T().Logf("Passed:        %d", passed)
	s.T().Logf("Failed:        %d", failed)
	s.T().Logf("Skipped:       %d", skipped)
	s.T().Logf("Duration:      %s", totalDuration)
	s.T().Logf(strings.Repeat("=", 70))
}
