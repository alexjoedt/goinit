package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"crypto/rand"
	"encoding/hex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func setupTestDir(t *testing.T, projectName string) (string, func()) {
	t.Helper()

	dir := filepath.Join("./test", randomName(), projectName)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.FailNow()
	}
	
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestInitGitRepo(t *testing.T) {
	dir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   dir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = initGitRepo(config)
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(config.TargetDir, ".git"))
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

			// Create target directory first
			err := os.MkdirAll(config.TargetDir, 0755)
			require.NoError(t, err)

			err = initGoMod(config)
			assert.NoError(t, err)
			
			goModPath := filepath.Join(config.TargetDir, "go.mod")
			assert.FileExists(t, goModPath)
			
			// Verify go.mod content
			content, err := os.ReadFile(goModPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), "module "+tt.wantModule)
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

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createMainDotGo(config)
	assert.NoError(t, err)
	
	mainGoPath := filepath.Join(config.TargetDir, "main.go")
	assert.FileExists(t, mainGoPath)
	
	// Verify content
	content, err := os.ReadFile(mainGoPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `fmt.Println("Hello test_project")`)
}

func TestCreateGitIgnore(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createGitignore(config)
	assert.NoError(t, err)
	
	gitignorePath := filepath.Join(config.TargetDir, ".gitignore")
	assert.FileExists(t, gitignorePath)
	
	// Verify content includes essential entries
	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "*.exe")
	assert.Contains(t, contentStr, "*.test")
	assert.Contains(t, contentStr, "bin/")
}

func TestCreateReadme(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createReadme(config)
	assert.NoError(t, err)
	
	readmePath := filepath.Join(config.TargetDir, "README.md")
	assert.FileExists(t, readmePath)
	
	// Verify content
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# test_project")
}

func TestCreateTaskfile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createTaskfile(config)
	assert.NoError(t, err)
	
	taskfilePath := filepath.Join(config.TargetDir, "Taskfile.yml")
	assert.FileExists(t, taskfilePath)
	
	// Verify content
	content, err := os.ReadFile(taskfilePath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "BINARY: test_project")
	assert.Contains(t, contentStr, "version: '3'")
}

func TestCreateMakefile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createMakefile(config)
	assert.NoError(t, err)
	
	makefilePath := filepath.Join(config.TargetDir, "Makefile")
	assert.FileExists(t, makefilePath)
	
	// Verify content
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "test_project")
	assert.Contains(t, contentStr, ".PHONY:")
}

func TestCreateDockerfile(t *testing.T) {
	testDir, cleanup := setupTestDir(t, "test_project")
	t.Cleanup(cleanup)

	config := &Config{
		ProjectName: "test_project",
		TargetDir:   testDir,
	}

	// Create target directory first
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	err = createDockerfile(config)
	assert.NoError(t, err)
	
	dockerfilePath := filepath.Join(config.TargetDir, "Dockerfile")
	assert.FileExists(t, dockerfilePath)
	
	// Verify content
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "test_project")
	assert.Contains(t, contentStr, "FROM golang:")
}

func TestRun_BasicProject(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "basic_project")
	t.Cleanup(cleanup)

	// Change to test root directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { 
		_ = os.Chdir(originalWd) // Ignore error in cleanup
	})
	err = os.Chdir(testRootDir)
	require.NoError(t, err)

	// Set up test config
	config := &Config{
		ProjectName:    "basic_project",
		ModuleName:     "",
		WithMakefile:   false,
		WithTaskfile:   false,
		WithDockerfile: false,
	}

	err = run(config)
	assert.NoError(t, err)

	// Verify all basic files are created
	assert.FileExists(t, filepath.Join(config.TargetDir, "README.md"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "go.mod"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "main.go"))
	assert.FileExists(t, filepath.Join(config.TargetDir, ".gitignore"))
	assert.DirExists(t, filepath.Join(config.TargetDir, ".git"))
	
	// Verify optional files are NOT created
	assert.NoFileExists(t, filepath.Join(config.TargetDir, "Makefile"))
	assert.NoFileExists(t, filepath.Join(config.TargetDir, "Taskfile.yml"))
	assert.NoFileExists(t, filepath.Join(config.TargetDir, "Dockerfile"))
}

func TestRun_WithAllOptions(t *testing.T) {
	testRootDir, cleanup := setupTestDir(t, "full_project") 
	t.Cleanup(cleanup)

	// Change to test root directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { 
		_ = os.Chdir(originalWd) // Ignore error in cleanup
	})
	err = os.Chdir(testRootDir)
	require.NoError(t, err)

	// Set up test config
	config := &Config{
		ProjectName:    "full_project",
		ModuleName:     "github.com/test/full_project",
		WithMakefile:   true,
		WithTaskfile:   true,
		WithDockerfile: true,
	}

	err = run(config)
	assert.NoError(t, err)

	// Verify all files are created
	assert.FileExists(t, filepath.Join(config.TargetDir, "README.md"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "go.mod"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "main.go"))
	assert.FileExists(t, filepath.Join(config.TargetDir, ".gitignore"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "Makefile"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "Taskfile.yml"))
	assert.FileExists(t, filepath.Join(config.TargetDir, "Dockerfile"))
	assert.DirExists(t, filepath.Join(config.TargetDir, ".git"))
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
				require.NoError(t, err)
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
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = os.Chdir(originalWd) // Ignore error in cleanup
			})
			err = os.Chdir(testRootDir)
			require.NoError(t, err)

			if tt.setupFunc != nil {
				tt.setupFunc(t, testRootDir)
			}

			config := &Config{
				ProjectName: tt.projectName,
			}

			err = run(config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
	assert.NoError(t, err)

	// Verify file exists and has correct content
	assert.FileExists(t, testFile)
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, testString, string(data))

	// Test file permissions
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	
	// Check that file is readable and writable by owner
	mode := info.Mode()
	if runtime.GOOS != "windows" {
		assert.True(t, mode&0600 != 0, "File should be readable and writable by owner")
	}
}

func TestWriteStringToFile_ErrorCases(t *testing.T) {
	// Test writing to invalid path (read-only parent directory)
	if runtime.GOOS != "windows" {
		testRootDir, cleanup := setupTestDir(t, "readonly_test")
		t.Cleanup(cleanup)

		readOnlyDir := filepath.Join(testRootDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0555) // Read and execute only
		require.NoError(t, err)

		testFile := filepath.Join(readOnlyDir, "testfile.txt")
		err = writeStringToFile(testFile, "test")
		assert.Error(t, err)
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
			name:     "existing command - ls/dir",
			binName:  func() string { if runtime.GOOS == "windows" { return "dir" } else { return "ls" } }(),
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
			assert.Equal(t, tt.expected, result)
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

	// Create target directory for command execution
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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

	// Create target directory
	err := os.MkdirAll(config.TargetDir, 0755)
	require.NoError(t, err)

	// Pre-create a README.md file
	readmePath := filepath.Join(config.TargetDir, "README.md")
	err = os.WriteFile(readmePath, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Attempt to create README - should fail because file exists
	err = createReadme(config)
	assert.Error(t, err)

	// Verify original content is preserved
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Equal(t, "existing content", string(content))
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