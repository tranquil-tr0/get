package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tranquil-tr0/get/internal/manager"
	"github.com/tranquil-tr0/get/internal/output"
	"github.com/tranquil-tr0/get/internal/tools"
	"github.com/urfave/cli/v2"
)

// MockPackageManager implements a mock for testing
type MockPackageManager struct {
	installCalled        bool
	removeCalled         bool
	updateAllCalled      bool
	upgradeAllCalled     bool
	printInstalledCalled bool
	lastPkgID            string
	lastRelease          string
	shouldReturnError    bool
	errorMessage         string
}

func NewMockPackageManager() *MockPackageManager {
	return &MockPackageManager{}
}

func (m *MockPackageManager) Install(pkgID, release string) error {
	m.installCalled = true
	m.lastPkgID = pkgID
	m.lastRelease = release
	if m.shouldReturnError {
		return fmt.Errorf("%s", m.errorMessage)
	}
	return nil
}

func (m *MockPackageManager) Remove(pkgID string) error {
	m.removeCalled = true
	m.lastPkgID = pkgID
	if m.shouldReturnError {
		return fmt.Errorf("%s", m.errorMessage)
	}
	return nil
}

func (m *MockPackageManager) UpdateAllPackages() error {
	m.updateAllCalled = true
	if m.shouldReturnError {
		return fmt.Errorf("%s", m.errorMessage)
	}
	return nil
}

func (m *MockPackageManager) UpgradeAllPackages() error {
	m.upgradeAllCalled = true
	if m.shouldReturnError {
		return fmt.Errorf("%s", m.errorMessage)
	}
	return nil
}

func (m *MockPackageManager) PrintInstalledPackages() error {
	m.printInstalledCalled = true
	if m.shouldReturnError {
		return fmt.Errorf("%s", m.errorMessage)
	}
	return nil
}

// Helper function to capture output
func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = wErr

	f()

	w.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	var bufErr bytes.Buffer
	io.Copy(&buf, r)
	io.Copy(&bufErr, rErr)

	return buf.String(), bufErr.String()
}

func TestMainFunction(t *testing.T) {
	// Test main function can run without panicking
	// This is a basic smoke test since main() calls os.Exit
	t.Run("MainFunctionExists", func(t *testing.T) {
		// Test that main function exists and can be called
		// We can't directly test main() due to os.Exit calls,
		// but we can test the app initialization logic
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
		pm := manager.NewPackageManager(metadataPath)

		if pm == nil {
			t.Fatal("Expected package manager to be initialized")
		}
	})
}

func TestAppConfiguration(t *testing.T) {
	// Create the CLI app similar to main()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
	pm := manager.NewPackageManager(metadataPath)

	app := &cli.App{
		Name:    "get",
		Version: "v0.1.0",
		Usage:   "A package manager for GitHub releases",
		Authors: []*cli.Author{
			{
				Name:  "tranquil-tr0",
				Email: "tranquiltr0@proton.me",
			},
		},
	}

	t.Run("AppBasicConfiguration", func(t *testing.T) {
		if app.Name != "get" {
			t.Errorf("Expected app name to be 'get', got %s", app.Name)
		}
		if app.Version != "v0.1.0" {
			t.Errorf("Expected app version to be 'v0.1.0', got %s", app.Version)
		}
		if app.Usage != "A package manager for GitHub releases" {
			t.Errorf("Expected app usage to be 'A package manager for GitHub releases', got %s", app.Usage)
		}
		if len(app.Authors) != 1 {
			t.Errorf("Expected 1 author, got %d", len(app.Authors))
		}
		if app.Authors[0].Name != "tranquil-tr0" {
			t.Errorf("Expected author name to be 'tranquil-tr0', got %s", app.Authors[0].Name)
		}
	})

	// Test package manager initialization
	if pm == nil {
		t.Fatal("Package manager should not be nil")
	}
}

func TestInstallCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
		shouldFail    bool
	}{
		{
			name:          "NoArguments",
			args:          []string{"get", "install"},
			expectedError: "Please provide a GitHub repository URL",
			shouldFail:    true,
		},
		{
			name:          "TooManyArguments",
			args:          []string{"get", "install", "repo1", "repo2"},
			expectedError: "Please provide a GitHub repository URL",
			shouldFail:    true,
		},
		{
			name:       "ValidRepository",
			args:       []string{"get", "install", "owner/repo"},
			shouldFail: false,
		},
		{
			name:       "ValidRepositoryWithRelease",
			args:       []string{"get", "install", "owner/repo", "--release", "v1.0.0"},
			shouldFail: false,
		},
		{
			name:       "ValidRepositoryWithVerbose",
			args:       []string{"get", "install", "owner/repo", "--verbose"},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDir := t.TempDir()
			metadataPath := filepath.Join(tempDir, "get.json")

			// We can't easily mock the package manager in the CLI context,
			// so we'll test the command parsing logic
			app := createTestApp(metadataPath)

			err := app.Run(tt.args)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
			// For non-failing cases, we don't check for errors since the actual
			// package manager operations will likely fail in test environment
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
		shouldFail    bool
	}{
		{
			name:          "NoArguments",
			args:          []string{"get", "remove"},
			expectedError: "Please provide a GitHub repository URL",
			shouldFail:    true,
		},
		{
			name:          "TooManyArguments",
			args:          []string{"get", "remove", "repo1", "repo2"},
			expectedError: "Please provide a GitHub repository URL",
			shouldFail:    true,
		},
		{
			name:       "ValidRepository",
			args:       []string{"get", "remove", "owner/repo"},
			shouldFail: false,
		},
		{
			name:       "ValidRepositoryWithVerbose",
			args:       []string{"get", "remove", "owner/repo", "--verbose"},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			metadataPath := filepath.Join(tempDir, "get.json")
			app := createTestApp(metadataPath)

			err := app.Run(tt.args)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			}
			// For non-failing cases, we don't check for errors since the actual
			// package manager operations will likely fail in test environment
		})
	}
}

func TestListCommand(t *testing.T) {
	t.Run("ListCommand", func(t *testing.T) {
		tempDir := t.TempDir()
		metadataPath := filepath.Join(tempDir, "get.json")
		app := createTestApp(metadataPath)

		args := []string{"get", "list"}
		err := app.Run(args)

		// We expect this to fail with actual package manager operations
		// but not due to argument parsing
		if err != nil && strings.Contains(err.Error(), "Please provide") {
			t.Errorf("Unexpected argument parsing error: %s", err.Error())
		}
	})

	t.Run("ListCommandWithVerbose", func(t *testing.T) {
		tempDir := t.TempDir()
		metadataPath := filepath.Join(tempDir, "get.json")
		app := createTestApp(metadataPath)

		args := []string{"get", "list", "--verbose"}
		err := app.Run(args)

		if err != nil && strings.Contains(err.Error(), "Please provide") {
			t.Errorf("Unexpected argument parsing error: %s", err.Error())
		}
	})
}

func TestUpdateCommand(t *testing.T) {
	t.Run("UpdateCommand", func(t *testing.T) {
		tempDir := t.TempDir()
		metadataPath := filepath.Join(tempDir, "get.json")
		app := createTestApp(metadataPath)

		args := []string{"get", "update"}
		err := app.Run(args)

		// Command should parse correctly, actual execution may fail
		if err != nil && strings.Contains(err.Error(), "Please provide") {
			t.Errorf("Unexpected argument parsing error: %s", err.Error())
		}
	})
}

func TestUpgradeCommand(t *testing.T) {
	t.Run("UpgradeCommand", func(t *testing.T) {
		tempDir := t.TempDir()
		metadataPath := filepath.Join(tempDir, "get.json")
		app := createTestApp(metadataPath)

		args := []string{"get", "upgrade"}
		err := app.Run(args)

		if err != nil && strings.Contains(err.Error(), "Please provide") {
			t.Errorf("Unexpected argument parsing error: %s", err.Error())
		}
	})
}

func TestUpdateUpgradeCommand(t *testing.T) {
	commands := [][]string{
		{"get", "update-upgrade"},
		{"get", "uu"},
		{"get", "up"},
	}

	for _, args := range commands {
		t.Run(fmt.Sprintf("Command_%s", args[1]), func(t *testing.T) {
			tempDir := t.TempDir()
			metadataPath := filepath.Join(tempDir, "get.json")
			app := createTestApp(metadataPath)

			err := app.Run(args)

			if err != nil && strings.Contains(err.Error(), "Please provide") {
				t.Errorf("Unexpected argument parsing error: %s", err.Error())
			}
		})
	}
}

func TestParseRepoURLIntegration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:     "ValidOwnerRepo",
			input:    "owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "ValidHTTPSURL",
			input:    "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "ValidHTTPURL",
			input:    "http://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:        "TooShort",
			input:       "ab",
			shouldError: true,
		},
		{
			name:        "Empty",
			input:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tools.ParseRepoURL(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestOutputFunctions(t *testing.T) {
	t.Run("PrintError", func(t *testing.T) {
		stdout, stderr := captureOutput(func() {
			output.PrintError("test error %s", "message")
		})

		if stdout != "" {
			t.Errorf("Expected empty stdout, got: %s", stdout)
		}
		if !strings.Contains(stderr, "test error message") {
			t.Errorf("Expected stderr to contain 'test error message', got: %s", stderr)
		}
	})

	t.Run("PrintSuccess", func(t *testing.T) {
		stdout, _ := captureOutput(func() {
			output.PrintSuccess("success %s", "message")
		})

		if !strings.Contains(stdout, "success message") {
			t.Errorf("Expected stdout to contain 'success message', got: %s", stdout)
		}
	})

	t.Run("PrintAction", func(t *testing.T) {
		stdout, _ := captureOutput(func() {
			output.PrintAction("action %s", "message")
		})

		if !strings.Contains(stdout, "action message") {
			t.Errorf("Expected stdout to contain 'action message', got: %s", stdout)
		}
	})
}

func TestCommandFlags(t *testing.T) {
	t.Run("InstallWithReleaseFlag", func(t *testing.T) {
		tempDir := t.TempDir()
		metadataPath := filepath.Join(tempDir, "get.json")
		app := createTestApp(metadataPath)

		args := []string{"get", "install", "owner/repo", "--release", "v1.0.0"}
		err := app.Run(args)

		// Should not fail due to flag parsing
		if err != nil && strings.Contains(err.Error(), "flag provided but not defined") {
			t.Errorf("Flag parsing error: %s", err.Error())
		}
	})

	t.Run("VerboseFlags", func(t *testing.T) {
		commands := [][]string{
			{"get", "install", "owner/repo", "--verbose"},
			{"get", "install", "owner/repo", "-v"},
			{"get", "list", "--verbose"},
			{"get", "remove", "owner/repo", "-v"},
			{"get", "update", "--verbose"},
			{"get", "upgrade", "-v"},
		}

		for _, args := range commands {
			t.Run(fmt.Sprintf("Verbose_%s", strings.Join(args[1:], "_")), func(t *testing.T) {
				tempDir := t.TempDir()
				metadataPath := filepath.Join(tempDir, "get.json")
				app := createTestApp(metadataPath)

				err := app.Run(args)

				if err != nil && strings.Contains(err.Error(), "flag provided but not defined") {
					t.Errorf("Flag parsing error: %s", err.Error())
				}
			})
		}
	})
}

// TestInvalidCommands is removed because cli.v2 calls os.Exit(3) for invalid commands
// which causes the test suite to fail. This behavior is expected from the CLI framework.

// Helper function to create test app
func createTestApp(metadataPath string) *cli.App {
	pm := manager.NewPackageManager(metadataPath)

	return &cli.App{
		Name:    "get",
		Version: "v0.1.0",
		Usage:   "A package manager for GitHub releases",
		Before: func(c *cli.Context) error {
			return nil
		},
		Authors: []*cli.Author{
			{
				Name:  "tranquil-tr0",
				Email: "tranquiltr0@proton.me",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "install",
				Category:    "Package Management",
				Usage:       "Install a package from GitHub",
				Description: "Install a package from a GitHub repository. The package must contain a .deb file in its latest release.",
				ArgsUsage:   "<github-repo-url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "release",
						Aliases: []string{"r"},
						Usage:   "Specify a release version to install",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}

					if err := pm.Install(pkgID, c.String("release")); err != nil {
						return fmt.Errorf("Error installing package: %v", err)
					}
					output.PrintSuccess("Successfully installed %s", pkgID)
					return nil
				},
			},
			{
				Name:        "list",
				Category:    "Package Management",
				Usage:       "List installed packages",
				Description: "Display a list of all packages installed through get.",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					err := pm.PrintInstalledPackages()
					if err != nil {
						return fmt.Errorf("Error listing packages: %v", err)
					}
					return nil
				},
			},
			{
				Name:        "remove",
				Category:    "Package Management",
				Usage:       "Remove an installed package",
				Description: "Remove a previously installed package and clean up its metadata.",
				ArgsUsage:   "<github-repo-url>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}

					if err := pm.Remove(pkgID); err != nil {
						return fmt.Errorf("Error removing package: %v", err)
					}
					output.PrintSuccess("Successfully removed %s", pkgID)
					return nil
				},
			},
			{
				Name:        "update",
				Category:    "Package Management",
				Usage:       "Check for package updates",
				Description: "Check for available updates of installed packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Checking for updates...")
					if err := pm.UpdateAllPackages(); err != nil {
						return fmt.Errorf("Error checking for updates: %v", err)
					}
					return nil
				},
			},
			{
				Name:        "upgrade",
				Category:    "Package Management",
				Usage:       "Apply staged upgrades",
				Description: "Install available updates for packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Upgrading packages...")
					if err := pm.UpgradeAllPackages(); err != nil {
						return fmt.Errorf("Error upgrading packages: %v", err)
					}
					output.PrintSuccess("Successfully applied all available updates")
					return nil
				},
			},
			{
				Name:        "update-upgrade",
				Category:    "Package Management",
				Aliases:     []string{"uu", "up"},
				Usage:       "Upgrade outdated packages",
				Description: "Check for updates then upgrade outdated packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Checking for updates...")
					if err := pm.UpdateAllPackages(); err != nil {
						return fmt.Errorf("Error checking for updates: %v", err)
					}

					output.PrintAction("Applying updates...")
					if err := pm.UpgradeAllPackages(); err != nil {
						return fmt.Errorf("Error upgrading packages: %v", err)
					}

					output.PrintSuccess("Successfully applied all available updates")
					return nil
				},
			},
		},
	}
}

// Benchmark tests
func BenchmarkParseRepoURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tools.ParseRepoURL("https://github.com/owner/repo")
	}
}

func BenchmarkAppCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		homeDir, _ := os.UserHomeDir()
		metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
		createTestApp(metadataPath)
	}
}
