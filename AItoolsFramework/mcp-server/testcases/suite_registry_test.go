package testcases

// ============================================================================
// Test Suite Registry - MCP Server Test Cases
//
// This file serves as the registry and orchestrator for all test cases in the
// MCP server test suite. It provides 11 main test cases that call granular
// test methods defined in separate test files.
//
// Purpose:
//   - Central registry of all test cases in the suite
//   - Maps high-level test names to granular test implementations
//   - Provides entry points for selective or full test execution
//   - Maintains logical test organization and execution order
//
// Test Case Listing:
//   Test01_RepositoryInstallation     → Repository setup and verification
//   Test02_PostgreSQLSetup            → PostgreSQL installation and config
//   Test03_MCPServerInstallation      → MCP package installation
//   Test04_InstallationValidation     → Binary and config validation
//   Test05_TokenManagement            → Token creation and management
//   Test06_UserManagement             → User creation and management
//   Test07_PackageFilesVerification   → File system verification
//   Test08_ServiceManagement          → Service lifecycle management
//   Test09_KnowledgeBuilder           → KB builder functionality
//   Test10_MCPServerWithKB            → MCP+KB integration testing
//   Test11_StdioMode                  → Stdio mode with JSON-RPC
//
// Usage:
//   go test -v ./testcases                    # Run all tests
//   go test -v -run Test08 ./testcases        # Run specific test
//   go test -v -run "Test0[1-5]" ./testcases  # Run tests 1-5
// ============================================================================

// Test01_RepositoryInstallation verifies repository installation and availability
func (s *MCPServerTestSuite) Test01_RepositoryInstallation() {
	s.T().Log("=== Test 01: Repository Installation ===")

	// Call granular repository tests
	s.Run("Repository_Installation", s.testRepository_Installation)
	s.Run("Repository_PackageAvailability", s.testRepository_PackageAvailability)
}

// Test02_PostgreSQLSetup verifies PostgreSQL installation and configuration
func (s *MCPServerTestSuite) Test02_PostgreSQLSetup() {
	s.T().Log("=== Test 02: PostgreSQL Setup ===")

	// Call granular PostgreSQL tests
	s.Run("PostgreSQL_Installation", s.testPostgreSQL_Installation)
	s.Run("PostgreSQL_ServiceStatus", s.testPostgreSQL_ServiceStatus)
	s.Run("PostgreSQL_DatabaseConnection", s.testPostgreSQL_DatabaseConnection)
	s.Run("PostgreSQL_MCPDatabase", s.testPostgreSQL_MCPDatabase)
}

// Test03_MCPServerInstallation verifies MCP server package installation
func (s *MCPServerTestSuite) Test03_MCPServerInstallation() {
	s.T().Log("=== Test 03: MCP Server Installation ===")

	// Call granular MCP server installation tests
	s.Run("MCPServer_PackagesInstalled", s.testMCPServer_PackagesInstalled)
	s.Run("Installation_MCPPackages", s.testInstallation_MCPPackages)
	s.Run("Installation_Repository", s.testInstallation_Repository)
}

// Test04_InstallationValidation verifies all installed binaries and configurations
func (s *MCPServerTestSuite) Test04_InstallationValidation() {
	s.T().Log("=== Test 04: Installation Validation ===")

	// Call granular installation validation tests
	s.Run("Installation_PackageFiles", s.testInstallation_PackageFiles)
	s.Run("MCPServer_BinaryFunctional", s.testMCPServer_BinaryFunctional)
	s.Run("MCPServer_ConfigValid", s.testMCPServer_ConfigValid)
	s.Run("MCPServer_EnvironmentFile", s.testMCPServer_EnvironmentFile)
}

// Test05_TokenManagement verifies token creation and management
func (s *MCPServerTestSuite) Test05_TokenManagement() {
	s.T().Log("=== Test 05: Token Management ===")

	// Call granular token management tests
	s.Run("Token_FileExists", s.testToken_FileExists)
	s.Run("Token_CreateToken", s.testToken_CreateToken)
	s.Run("Token_ListTokens", s.testToken_ListTokens)
}

// Test06_UserManagement verifies user creation and management
func (s *MCPServerTestSuite) Test06_UserManagement() {
	s.T().Log("=== Test 06: User Management ===")

	// Call granular user management tests
	s.Run("User_CreateUser", s.testUser_CreateUser)
	s.Run("User_ListUsers", s.testUser_ListUsers)
	s.Run("User_FilePermissions", s.testUser_FilePermissions)
}

// Test07_PackageFilesVerification verifies all package files and directories
func (s *MCPServerTestSuite) Test07_PackageFilesVerification() {
	s.T().Log("=== Test 07: Package Files Verification ===")

	// Call granular file verification tests
	s.Run("Files_BinariesExist", s.testFiles_BinariesExist)
	s.Run("Files_ConfigFiles", s.testFiles_ConfigFiles)
	s.Run("Files_DataDirectory", s.testFiles_DataDirectory)
	s.Run("Files_LogDirectories", s.testFiles_LogDirectories)
	s.Run("Files_SystemdService", s.testFiles_SystemdService)
}

// Test08_ServiceManagement verifies service management functionality
func (s *MCPServerTestSuite) Test08_ServiceManagement() {
	s.T().Log("=== Test 08: Service Management ===")

	// Call granular service management tests
	s.Run("Service_SystemdManagement", s.testService_SystemdManagement)
}

// Test09_KnowledgeBuilder verifies knowledge base builder functionality
func (s *MCPServerTestSuite) Test09_KnowledgeBuilder() {
	s.T().Log("=== Test 09: Knowledge Builder ===")

	// Call granular KB builder tests
	s.Run("KB_BuilderInstallation", s.testKB_BuilderInstallation)
}

// Test10_MCPServerWithKB verifies MCP server integration with knowledge base
func (s *MCPServerTestSuite) Test10_MCPServerWithKB() {
	s.T().Log("=== Test 10: MCP Server with Knowledge Base ===")

	// Call granular MCP+KB integration tests
	s.Run("MCPKB_Integration", s.testMCPKB_Integration)
}

// Test11_StdioMode verifies stdio mode functionality
func (s *MCPServerTestSuite) Test11_StdioMode() {
	s.T().Log("=== Test 11: Stdio Mode ===")

	// Call granular stdio mode tests
	s.Run("Stdio_ModeWithDatabaseConnectivity", s.testStdio_ModeWithDatabaseConnectivity)
}
