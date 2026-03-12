package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T, projectName string) (string, func()) {
	t.Helper()

	dir := filepath.Join("./test", randomName(), projectName)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) || (err == nil && info.IsDir()) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertNoFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("expected file to not exist: %s", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) || (err == nil && !info.IsDir()) {
		t.Errorf("expected directory to exist: %s", path)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected string to contain %q\ngot: %q", substr, s)
	}
}

func assertEqual(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestInitGitRepo(t *testing.T) {
	dir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   dir,
	}

	err := initGitRepo(config)
	assertNoError(t, err)
	assertDirExists(t, filepath.Join(config.TargetDir, ".git"))
}

func TestInitGoMod(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		moduleName  string
		wantModule  string
	}{
		{
			name:        "project name only",
			projectName: "test-project",
			moduleName:  "",
			wantModule:  "test-project",
		},
		{
			name:        "custom module name",
			projectName: "test-project",
			moduleName:  "github.com/alexjoedt/go-test",
			wantModule:  "github.com/alexjoedt/go-test",
		},
		{
			name:        "module with dots",
			projectName: "my-project",
			moduleName:  "example.com/user/repo",
			wantModule:  "example.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := setupTestDir(t, tt.projectName)
			defer cleanup()

			config := &Config{
				ProjectName: tt.projectName,
				ModuleName:  tt.moduleName,
				TargetDir:   dir,
			}

			err := initGoMod(config)
			assertNoError(t, err)

			goModPath := filepath.Join(config.TargetDir, "go.mod")
			assertFileExists(t, goModPath)

			content, err := os.ReadFile(goModPath)
			requireNoError(t, err)
			assertContains(t, string(content), "module "+tt.wantModule)
		})
	}
}

func TestCreateMainDotGo(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createMainDotGo(config)
	assertNoError(t, err)

	mainGoPath := filepath.Join(config.TargetDir, "main.go")
	assertFileExists(t, mainGoPath)

	content, err := os.ReadFile(mainGoPath)
	requireNoError(t, err)
	assertContains(t, string(content), `fmt.Println("Hello test_project")`)
}

func TestCreateGitIgnore(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createGitignore(config)
	assertNoError(t, err)

	gitignorePath := filepath.Join(config.TargetDir, ".gitignore")
	assertFileExists(t, gitignorePath)

	content, err := os.ReadFile(gitignorePath)
	requireNoError(t, err)
	contentStr := string(content)
	assertContains(t, contentStr, "*.exe")
	assertContains(t, contentStr, "*.test")
	assertContains(t, contentStr, "bin/")
}

func TestCreateReadme(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createReadme(config)
	assertNoError(t, err)

	readmePath := filepath.Join(config.TargetDir, "README.md")
	assertFileExists(t, readmePath)

	content, err := os.ReadFile(readmePath)
	requireNoError(t, err)
	assertContains(t, string(content), "# test_project")
}

func TestCreateTaskfile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createTaskfile(config)
	assertNoError(t, err)

	taskfilePath := filepath.Join(config.TargetDir, "Taskfile.yml")
	assertFileExists(t, taskfilePath)

	content, err := os.ReadFile(taskfilePath)
	requireNoError(t, err)
	contentStr := string(content)
	assertContains(t, contentStr, "BINARY: test_project")
	assertContains(t, contentStr, "version: '3'")
}

func TestCreateMakefile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createMakefile(config)
	assertNoError(t, err)

	makefilePath := filepath.Join(config.TargetDir, "Makefile")
	assertFileExists(t, makefilePath)

	content, err := os.ReadFile(makefilePath)
	requireNoError(t, err)
	contentStr := string(content)
	assertContains(t, contentStr, "test_project")
	assertContains(t, contentStr, ".PHONY:")
}

func TestCreateDockerfile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	err := createDockerfile(config)
	assertNoError(t, err)

	dockerfilePath := filepath.Join(config.TargetDir, "Dockerfile")
	assertFileExists(t, dockerfilePath)

	content, err := os.ReadFile(dockerfilePath)
	requireNoError(t, err)
	contentStr := string(content)
	assertContains(t, contentStr, "test_project")
	assertContains(t, contentStr, "FROM golang:")
}

func TestRun_BasicProject(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "basic_project")
	t.Cleanup(cleanup)

	// Change to test root directory
	originalWd, err := os.Getwd()
	requireNoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})
	err = os.Chdir(testRootDir)
	requireNoError(t, err)

	config := &Config{
		ProjectName:    "basic_project",
		ModuleName:     "",
		WithMakefile:   false,
		WithTaskfile:   false,
		WithDockerfile: false,
	}

	err = run(config, io.Discard, io.Discard)
	assertNoError(t, err)

	assertFileExists(t, filepath.Join(config.TargetDir, "README.md"))
	assertFileExists(t, filepath.Join(config.TargetDir, "go.mod"))
	assertFileExists(t, filepath.Join(config.TargetDir, "main.go"))
	assertFileExists(t, filepath.Join(config.TargetDir, ".gitignore"))
	assertDirExists(t, filepath.Join(config.TargetDir, ".git"))

	assertNoFileExists(t, filepath.Join(config.TargetDir, "Makefile"))
	assertNoFileExists(t, filepath.Join(config.TargetDir, "Taskfile.yml"))
	assertNoFileExists(t, filepath.Join(config.TargetDir, "Dockerfile"))
}

func TestRun_WithAllOptions(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "full_project")
	t.Cleanup(cleanup)

	// Change to test root directory
	originalWd, err := os.Getwd()
	requireNoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})
	err = os.Chdir(testRootDir)
	requireNoError(t, err)

	config := &Config{
		ProjectName:    "full_project",
		ModuleName:     "github.com/test/full_project",
		WithMakefile:   true,
		WithTaskfile:   true,
		WithDockerfile: true,
	}

	err = run(config, io.Discard, io.Discard)
	assertNoError(t, err)

	assertFileExists(t, filepath.Join(config.TargetDir, "README.md"))
	assertFileExists(t, filepath.Join(config.TargetDir, "go.mod"))
	assertFileExists(t, filepath.Join(config.TargetDir, "main.go"))
	assertFileExists(t, filepath.Join(config.TargetDir, ".gitignore"))
	assertFileExists(t, filepath.Join(config.TargetDir, "Makefile"))
	assertFileExists(t, filepath.Join(config.TargetDir, "Taskfile.yml"))
	assertFileExists(t, filepath.Join(config.TargetDir, "Dockerfile"))
	assertDirExists(t, filepath.Join(config.TargetDir, ".git"))
}

func TestRun_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		setupFunc   func(t *testing.T, testDir string) // Additional setup
		wantError   bool
	}{
		{
			name:        "empty project name",
			projectName: "",
			wantError:   true,
		},
		{
			name:        "existing directory",
			projectName: "existing_project",
			setupFunc: func(t *testing.T, testDir string) {
				// Pre-create the target directory with same path that run() will use
				wd, _ := os.Getwd()
				targetPath := filepath.Join(wd, "existing_project")
				err := os.MkdirAll(targetPath, 0755)
				if err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRootDir, cleanup := setupTestDir(t, tt.projectName)
			t.Cleanup(cleanup)

			// Change to test root directory
			originalWd, err := os.Getwd()
			requireNoError(t, err)
			t.Cleanup(func() {
				_ = os.Chdir(originalWd)
			})
			err = os.Chdir(testRootDir)
			requireNoError(t, err)

			if tt.setupFunc != nil {
				tt.setupFunc(t, testRootDir)
			}

			config := &Config{
				ProjectName: tt.projectName,
			}

			err = run(config, io.Discard, io.Discard)
			if tt.wantError {
				assertError(t, err)
			} else {
				assertNoError(t, err)
			}
		})
	}
}

func TestWriteStringToFile(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "write_test")
	t.Cleanup(cleanup)

	testFile := filepath.Join(testRootDir, "testfile.txt")
	testString := "test-string-content\nwith multiple lines"

	err := writeStringToFile(testFile, testString)
	assertNoError(t, err)

	assertFileExists(t, testFile)
	data, err := os.ReadFile(testFile)
	requireNoError(t, err)
	assertEqual(t, testString, string(data))

	info, err := os.Stat(testFile)
	requireNoError(t, err)

	mode := info.Mode()
	if runtime.GOOS != "windows" {
		if mode&0600 == 0 {
			t.Error("file should be readable and writable by owner")
		}
	}
}

func TestWriteStringToFile_ErrorCases(t *testing.T) {
	// Test writing to invalid path (read-only parent directory)
	if runtime.GOOS != "windows" {
		testRootDir, cleanup := setupTestDir(t, "readonly_test")
		t.Cleanup(cleanup)

		readOnlyDir := filepath.Join(testRootDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0555)
		requireNoError(t, err)

		testFile := filepath.Join(readOnlyDir, "testfile.txt")
		err = writeStringToFile(testFile, "test")
		assertError(t, err)
	}
}

func TestBinExists(t *testing.T) {
	tests := []struct {
		name     string
		binName  string
		expected bool
	}{
		{
			name:     "existing command - echo",
			binName:  "echo",
			expected: true,
		},
		{
			name: "existing command - ls/dir",
			binName: func() string {
				if runtime.GOOS == "windows" {
					return "dir"
				} else {
					return "ls"
				}
			}(),
			expected: true,
		},
		{
			name:     "non-existing command",
			binName:  "definitely-not-existing-command-12345",
			expected: false,
		},
		{
			name:     "empty string",
			binName:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := binExists(tt.binName)
			if result != tt.expected {
				t.Errorf("binExists(%q) = %v, want %v", tt.binName, result, tt.expected)
			}
		})
	}
}

func TestExecCommand(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "exec_test")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "exec_test",
		TargetDir:   testDir,
	}

	tests := []struct {
		name      string
		cmd       string
		args      []string
		wantError bool
	}{
		{
			name:      "simple echo command",
			cmd:       "echo",
			args:      []string{"hello"},
			wantError: false,
		},
		{
			name:      "echo without args",
			cmd:       "echo",
			args:      nil,
			wantError: false,
		},
		{
			name:      "non-existing command",
			cmd:       "definitely-not-existing-command-12345",
			args:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := execCommand(config, tt.cmd, tt.args...)
			if tt.wantError {
				assertError(t, err)
			} else {
				assertNoError(t, err)
			}
		})
	}
}

// Test file creation with existing files (should fail)
func TestFileCreation_ExistingFiles(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "existing_files_test")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "existing_files_test",
		TargetDir:   testDir,
	}

	readmePath := filepath.Join(config.TargetDir, "README.md")
	err := os.WriteFile(readmePath, []byte("existing content"), 0644)
	requireNoError(t, err)

	err = createReadme(config)
	assertError(t, err)

	content, err := os.ReadFile(readmePath)
	requireNoError(t, err)
	assertEqual(t, "existing content", string(content))
}

func TestRandomName(t *testing.T) {
	var name string
	for range 100 {
		newName := randomName()
		if newName == name {
			t.Fatal("names should be unequal")
		}
	}
}

func randomName() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "randfolder"
	}
	return hex.EncodeToString(b)
}

func TestRunInteractiveMode_HappyPath(t *testing.T) {
	t.Parallel()

	input := "my-project\ngithub.com/user/my-project\ny\ny\ny\nn\n"
	var out bytes.Buffer
	config := &Config{}

	err := runInteractiveMode(config, strings.NewReader(input), &out)
	assertNoError(t, err)

	assertEqual(t, "my-project", config.ProjectName)
	assertEqual(t, "github.com/user/my-project", config.ModuleName)
	if !config.WithTaskfile {
		t.Error("expected WithTaskfile to be true")
	}
	if !config.WithMakefile {
		t.Error("expected WithMakefile to be true")
	}
	if !config.WithDockerfile {
		t.Error("expected WithDockerfile to be true")
	}
	if config.Verbose {
		t.Error("expected Verbose to be false")
	}

	assertContains(t, out.String(), "Interactive Go Project Setup")
}

func TestRunInteractiveMode_EmptyName(t *testing.T) {
	t.Parallel()

	input := "\n"
	var out bytes.Buffer
	config := &Config{}

	err := runInteractiveMode(config, strings.NewReader(input), &out)
	assertError(t, err)
	assertContains(t, err.Error(), "project name is required")
}

func TestRunInteractiveMode_DefaultModule(t *testing.T) {
	t.Parallel()

	input := "my-tool\n\nn\nn\nn\nn\n"
	var out bytes.Buffer
	config := &Config{}

	err := runInteractiveMode(config, strings.NewReader(input), &out)
	assertNoError(t, err)

	assertEqual(t, "my-tool", config.ProjectName)
	assertEqual(t, "", config.ModuleName)
}

func TestRun_InvalidChars(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
	}{
		{"angle brackets", "my<project>"},
		{"colon", "my:project"},
		{"quotes", `my"project"`},
		{"pipe", "my|project"},
		{"question mark", "my?project"},
		{"asterisk", "my*project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRootDir, cleanup := setupTestDir(t, "invalid_chars")
			t.Cleanup(cleanup)

			originalWd, err := os.Getwd()
			requireNoError(t, err)
			t.Cleanup(func() { _ = os.Chdir(originalWd) })
			err = os.Chdir(testRootDir)
			requireNoError(t, err)

			config := &Config{ProjectName: tt.projectName}
			err = run(config, io.Discard, io.Discard)
			assertError(t, err)
			assertContains(t, err.Error(), "invalid characters")
		})
	}
}

func TestRun_VerboseOutput(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "verbose_project")
	t.Cleanup(cleanup)

	originalWd, err := os.Getwd()
	requireNoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWd) })
	err = os.Chdir(testRootDir)
	requireNoError(t, err)

	var stdout, stderr bytes.Buffer
	config := &Config{
		ProjectName: "verbose_project",
		Verbose:     true,
	}

	err = run(config, &stdout, &stderr)
	assertNoError(t, err)

	// Progress output goes to stdout
	assertContains(t, stdout.String(), "verbose_project")

	// Info messages go to stderr
	assertContains(t, stderr.String(), "Creating project in")
}
