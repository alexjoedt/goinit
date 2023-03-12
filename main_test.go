package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDir = "test"
)

func initTestDir() string {
	targetDir = filepath.Join(testDir, "project")
	os.MkdirAll(targetDir, 0755)
	return targetDir
}

func cleanUp() {
	os.RemoveAll(testDir)
}

func TestInitGitRepo(t *testing.T) {
	initTestDir()
	defer cleanUp()
	err := initGitRepo()
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(targetDir, ".git"))
}

func TestInitGoMod(t *testing.T) {
	testCases := []struct {
		projectName string
		moduleName  string
	}{
		{projectName: "test-project", moduleName: ""},
		{projectName: "test-project", moduleName: "github.com/alexjoedt/go-test"},
	}
	for _, tc := range testCases {
		initTestDir()
		projectName = tc.projectName
		moduleName = tc.moduleName
		assert.NoError(t, initGoMod())
		assert.FileExists(t, filepath.Join(targetDir, "go.mod"))
		cleanUp()
	}
}

func TestCreateMainDotGo(t *testing.T) {
	initTestDir()
	defer cleanUp()

	assert.NoError(t, createMainDotGo())
	assert.FileExists(t, filepath.Join(targetDir, "main.go"))
}

func TestCreateGitIgnore(t *testing.T) {
	initTestDir()
	defer cleanUp()

	assert.NoError(t, createGitignore())
	assert.FileExists(t, filepath.Join(targetDir, ".gitignore"))
}

func TestCreateReadme(t *testing.T) {
	initTestDir()
	defer cleanUp()

	assert.NoError(t, createReadme())
	assert.FileExists(t, filepath.Join(targetDir, "README.md"))
}

func TestCreateTaskfile(t *testing.T) {
	initTestDir()
	defer cleanUp()

	assert.NoError(t, createTaskfile())
	assert.FileExists(t, filepath.Join(targetDir, "Taskfile.yml"))
}

func TestCreateMakefile(t *testing.T) {
	initTestDir()
	defer cleanUp()

	assert.NoError(t, createMakefile())
	assert.FileExists(t, filepath.Join(targetDir, "Makefile"))
}

func TestRun(t *testing.T) {
	testCases := []struct {
		projectName  string
		moduleName   string
		withMakefile bool
		withTaskfile bool
		verbose      bool
	}{
		{projectName: testDir + "/project", moduleName: "", withMakefile: false, withTaskfile: false},
		{projectName: testDir + "/project", moduleName: "github.com/alexjoedt/go-test", withMakefile: true, withTaskfile: true},
		{projectName: testDir, moduleName: "", withMakefile: true, withTaskfile: false},
		{projectName: testDir, moduleName: "github.com/alexjoedt/go-test", withMakefile: false, withTaskfile: true},
	}

	for _, tc := range testCases {
		projectName = tc.projectName
		moduleName = tc.moduleName
		withMakefile = tc.withMakefile
		withTaskfile = tc.withTaskfile
		verbose = tc.verbose

		assert.NoError(t, run())

		if withMakefile {
			assert.FileExists(t, filepath.Join(targetDir, "Makefile"))
		}

		if withTaskfile {
			assert.FileExists(t, filepath.Join(targetDir, "Taskfile.yml"))
		}

		assert.FileExists(t, filepath.Join(targetDir, "README.md"))
		assert.FileExists(t, filepath.Join(targetDir, "go.mod"))
		assert.FileExists(t, filepath.Join(targetDir, "main.go"))
		assert.FileExists(t, filepath.Join(targetDir, ".gitignore"))
		assert.DirExists(t, filepath.Join(targetDir, ".git"))
		cleanUp()
	}
}

func TestWriteStringToFile(t *testing.T) {
	os.MkdirAll(testDir, 0755)

	testFile := filepath.Join(testDir, "testfile")
	testString := "test-string"

	assert.NoError(t, writeStringToFile(testFile, testString))

	data, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testString, string(data))
	assert.FileExists(t, testFile)

	cleanUp()
}

func TestBinExists(t *testing.T) {
	assert.True(t, binExists("echo"))
	assert.False(t, binExists("not-existing-bin"))
}

func TestExecCommand(t *testing.T) {
	initTestDir()
	defer cleanUp()
	assert.NoError(t, execCommand("echo"))
	assert.NoError(t, execCommand("echo"), "test")
	assert.Error(t, execCommand("not-existing-bin"))
}
