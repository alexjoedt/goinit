package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func init() {
	if !binExists("git") {
		log.Fatal("git command not found in PATH. Please install Git: https://git-scm.com/downloads")
	}

	if !binExists("go") {
		log.Fatal("go command not found in PATH. Please install Go: https://golang.org/doc/install")
	}
}

var (
	version = "unknown"
)

// Config holds all CLI configuration options
type Config struct {
	ProjectName    string
	ModuleName     string
	TargetDir      string
	Verbose        bool
	WithTaskfile   bool
	WithMakefile   bool
	WithDockerfile bool
	ShowVersion    bool
	ShowHelp       bool
	Interactive    bool
}

const usage = `goinit - A simple Go project initializer

USAGE:
    goinit [OPTIONS] <project-name>
    goinit --interactive

EXAMPLES:
    goinit my-app                          # Create a simple Go project
    goinit -t -d my-webapp                 # With Taskfile and Dockerfile
    goinit -m my-tool                      # With Makefile
    goinit --interactive                   # Interactive mode

OPTIONS:
    -t, --taskfile     Add Taskfile.yml for task automation
    -m, --makefile     Add Makefile for build automation
    -d, --dockerfile   Add Dockerfile for containerization
        --module <name>    Custom Go module name (default: project-name)
    -i, --interactive  Interactive mode - prompts for all options
    -v, --verbose      Show detailed output
    -h, --help         Show this help message
        --version      Show version information
`

// runInteractiveMode prompts the user for project configuration options
func runInteractiveMode(config *Config) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("Interactive Go Project Setup")
	fmt.Println("============================")
	
	// Project name
	fmt.Print("Project name: ")
	if input, err := reader.ReadString('\n'); err == nil {
		config.ProjectName = strings.TrimSpace(input)
	}
	
	if config.ProjectName == "" {
		return errors.New("project name is required")
	}
	
	// Module name
	fmt.Printf("Go module name (default: %s): ", config.ProjectName)
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.TrimSpace(input)
		if input != "" {
			config.ModuleName = input
		}
	}
	
	// Ask about individual files
	fmt.Print("Add Taskfile.yml? (y/N): ")
	if input, err := reader.ReadString('\n'); err == nil {
		config.WithTaskfile = strings.ToLower(strings.TrimSpace(input)) == "y"
	}
	
	fmt.Print("Add Makefile? (y/N): ")
	if input, err := reader.ReadString('\n'); err == nil {
		config.WithMakefile = strings.ToLower(strings.TrimSpace(input)) == "y"
	}
	
	fmt.Print("Add Dockerfile? (y/N): ")
	if input, err := reader.ReadString('\n'); err == nil {
		config.WithDockerfile = strings.ToLower(strings.TrimSpace(input)) == "y"
	}
	
	fmt.Print("Verbose output? (y/N): ")
	if input, err := reader.ReadString('\n'); err == nil {
		config.Verbose = strings.ToLower(strings.TrimSpace(input)) == "y"
	}
	
	fmt.Println()
	return nil
}

func main() {
	config := &Config{}

	// Help and version flags
	flag.BoolVar(&config.ShowHelp, "help", false, "Show help message")
	flag.BoolVar(&config.ShowHelp, "h", false, "Show help message")
	flag.BoolVar(&config.ShowVersion, "version", false, "Show version information")

	// General options
	flag.BoolVar(&config.Verbose, "verbose", false, "Show detailed output")
	flag.BoolVar(&config.Verbose, "v", false, "Show detailed output")
	flag.BoolVar(&config.Interactive, "interactive", false, "Interactive mode")
	flag.BoolVar(&config.Interactive, "i", false, "Interactive mode")
	flag.StringVar(&config.ModuleName, "module", "", "Custom Go module name")

	// File options
	flag.BoolVar(&config.WithTaskfile, "taskfile", false, "Add Taskfile.yml")
	flag.BoolVar(&config.WithTaskfile, "t", false, "Add Taskfile.yml")
	flag.BoolVar(&config.WithMakefile, "makefile", false, "Add Makefile")
	flag.BoolVar(&config.WithMakefile, "m", false, "Add Makefile")
	flag.BoolVar(&config.WithDockerfile, "dockerfile", false, "Add Dockerfile")
	flag.BoolVar(&config.WithDockerfile, "d", false, "Add Dockerfile")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if config.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	if config.ShowVersion {
		fmt.Printf("goinit version %s\n", version)
		os.Exit(0)
	}

	config.ProjectName = flag.Arg(0)

	if err := run(config); err != nil {
		fatal("%v", err)
	}
}

// Templates
var (
	//go:embed templates
	tplFiles embed.FS
)

// run creates the go project directory
func run(config *Config) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Handle interactive mode
	if config.Interactive {
		if err := runInteractiveMode(config); err != nil {
			return fmt.Errorf("interactive mode failed: %w", err)
		}
	}

	// Better validation with helpful suggestions
	if config.ProjectName == "" {
		return errors.New(`Project name is required.

Examples:
  goinit my-app                    # Simple project
  goinit --interactive             # Interactive setup
  goinit -t -d my-webapp           # With Taskfile and Dockerfile

Use 'goinit --help' for more information.`)
	}

	// Validate project name
	if strings.TrimSpace(config.ProjectName) == "" {
		return errors.New("Project name cannot be empty or just whitespace")
	}

	config.ProjectName = filepath.Base(config.ProjectName)
	
	// Check for invalid characters in project name
	if strings.ContainsAny(config.ProjectName, `<>:"/\|?*`) {
		return fmt.Errorf("Project name '%s' contains invalid characters. Use only letters, numbers, hyphens, and underscores", config.ProjectName)
	}

	config.TargetDir = filepath.Join(wd, config.ProjectName)
	if _, err := os.Stat(config.TargetDir); !os.IsNotExist(err) {
		return fmt.Errorf(`Directory '%s' already exists.

Solutions:
  • Choose a different project name: goinit my-other-project
  • Remove the existing directory: rm -rf %s
  • Use a parent directory: mkdir parent && cd parent && goinit %s`, config.TargetDir, config.ProjectName, config.ProjectName)
	}

	// Initialize progress tracking
	steps := []string{
		"Creating project directory",
		"Initializing Go module",
		"Initializing Git repository", 
		"Creating .gitignore",
		"Creating README.md",
		"Creating main.go",
	}
	
	if config.WithTaskfile {
		steps = append(steps, "Creating Taskfile.yml")
	}
	if config.WithMakefile {
		steps = append(steps, "Creating Makefile")
	}
	if config.WithDockerfile {
		steps = append(steps, "Creating Dockerfile")
	}

	progress := NewProgressTracker(steps)
	progress.Start(config)

	info(config, "Creating project in %s", config.TargetDir)

	if err := os.MkdirAll(config.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory '%s': %w", config.TargetDir, err)
	}
	progress.NextStep(config)

	// Init Go Module
	info(config, "Initializing Go module")
	if err := initGoMod(config); err != nil {
		return fmt.Errorf(`failed to initialize Go module: %w

Make sure Go is properly installed:
  • Download from: https://golang.org/dl/
  • Verify with: go version`, err)
	}
	progress.NextStep(config)

	// Init Git Repo
	info(config, "Initializing Git repository")
	if err := initGitRepo(config); err != nil {
		return fmt.Errorf(`failed to initialize Git repository: %w

Make sure Git is properly installed:
  • Download from: https://git-scm.com/downloads
  • Verify with: git --version`, err)
	}
	progress.NextStep(config)

	// Create .gitignore
	info(config, "Creating .gitignore")
	if err := createGitignore(config); err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}
	progress.NextStep(config)

	// Create README.md
	info(config, "Creating README.md")
	if err := createReadme(config); err != nil {
		return fmt.Errorf("failed to create README.md file: %w", err)
	}
	progress.NextStep(config)

	// Create main.go
	info(config, "Creating main.go")
	if err := createMainDotGo(config); err != nil {
		return fmt.Errorf("failed to create main.go file: %w", err)
	}
	progress.NextStep(config)

	if config.WithTaskfile {
		if !binExists("task") {
			warn("Task binary not found. Install from: https://taskfile.dev/installation/")
		}

		info(config, "Creating Taskfile.yml")
		if err := createTaskfile(config); err != nil {
			warn("Failed to create Taskfile.yml: %v", err)
		} else {
			progress.NextStep(config)
		}
	}

	if config.WithMakefile {
		if !binExists("make") {
			warn("Make binary not found. Install build tools for your system.")
		}

		info(config, "Creating Makefile")
		if err := createMakefile(config); err != nil {
			warn("Failed to create Makefile: %v", err)
		} else {
			progress.NextStep(config)
		}
	}

	if config.WithDockerfile {
		if !binExists("docker") {
			warn("Docker binary not found. Install from: https://docs.docker.com/get-docker/")
		}

		info(config, "Creating Dockerfile")
		if err := createDockerfile(config); err != nil {
			warn("Failed to create Dockerfile: %v", err)
		} else {
			progress.NextStep(config)
		}
	}

	progress.Complete(config)
	return nil
}

func initGitRepo(config *Config) error {
	return execCommand(config, "git", "init", "--initial-branch=main")
}

func initGoMod(config *Config) error {
	module := config.ProjectName
	if config.ModuleName != "" {
		module = config.ModuleName
	}
	return execCommand(config, "go", "mod", "init", module)
}

func createTaskfile(config *Config) error {
	tf, err := tplFiles.ReadFile("templates/Taskfile.yml")
	if err != nil {
		return fmt.Errorf("failed to read Taskfile.yml template: %w", err)
	}

	tpl, err := template.New("taskfile").Delims("[[", "]]").Parse(string(tf))
	if err != nil {
		return fmt.Errorf("failed to parse Taskfile.yml template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{config.ProjectName})
	if err != nil {
		return fmt.Errorf("failed to execute Taskfile.yml template: %w", err)
	}

	taskfilePath := filepath.Join(config.TargetDir, "Taskfile.yml")
	if _, err = os.Stat(taskfilePath); !os.IsNotExist(err) {
		return fmt.Errorf("taskfile.yml already exists at '%s'", taskfilePath)
	}

	if err := writeStringToFile(taskfilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Taskfile.yml: %w", err)
	}

	return nil
}

func createDockerfile(config *Config) error {
	df, err := tplFiles.ReadFile("templates/Dockerfile")
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile template: %w", err)
	}

	tpl, err := template.New("dockerfile").Parse(string(df))
	if err != nil {
		return fmt.Errorf("failed to parse Dockerfile template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{config.ProjectName})
	if err != nil {
		return fmt.Errorf("failed to execute Dockerfile template: %w", err)
	}

	dockerfilePath := filepath.Join(config.TargetDir, "Dockerfile")
	if _, err = os.Stat(dockerfilePath); !os.IsNotExist(err) {
		return fmt.Errorf("dockerfile already exists at '%s'", dockerfilePath)
	}

	if err := writeStringToFile(dockerfilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

func createMakefile(config *Config) error {
	mf, err := tplFiles.ReadFile("templates/Makefile")
	if err != nil {
		return fmt.Errorf("failed to read Makefile template: %w", err)
	}

	t, err := template.New("makefile").Parse(string(mf))
	if err != nil {
		return fmt.Errorf("failed to parse Makefile template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, struct{ ProjectName string }{config.ProjectName})
	if err != nil {
		return fmt.Errorf("failed to execute Makefile template: %w", err)
	}

	makefilePath := filepath.Join(config.TargetDir, "Makefile")
	if _, err = os.Stat(makefilePath); !os.IsNotExist(err) {
		return fmt.Errorf("makefile already exists at '%s'", makefilePath)
	}

	if err := writeStringToFile(makefilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

func createReadme(config *Config) error {
	rd, err := tplFiles.ReadFile("templates/README.md")
	if err != nil {
		return fmt.Errorf("failed to read README.md template: %w", err)
	}

	t, err := template.New("readme").Parse(string(rd))
	if err != nil {
		return fmt.Errorf("failed to parse README.md template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, struct{ ProjectName string }{config.ProjectName})
	if err != nil {
		return fmt.Errorf("failed to execute README.md template: %w", err)
	}

	readmePath := filepath.Join(config.TargetDir, "README.md")
	if _, err = os.Stat(readmePath); !os.IsNotExist(err) {
		return fmt.Errorf("README.md already exists at '%s'", readmePath)
	}

	if err := writeStringToFile(readmePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

func createGitignore(config *Config) error {
	gi, err := tplFiles.ReadFile("templates/gitignore")
	if err != nil {
		return fmt.Errorf("failed to read gitignore template: %w", err)
	}

	gitignorePath := filepath.Join(config.TargetDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); !os.IsNotExist(err) {
		return fmt.Errorf(".gitignore already exists at '%s'", gitignorePath)
	}

	if err := writeStringToFile(gitignorePath, string(gi)); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

func createMainDotGo(config *Config) error {
	mg, err := tplFiles.ReadFile("templates/main.go")
	if err != nil {
		return fmt.Errorf("failed to read main.go template: %w", err)
	}

	t, err := template.New("main.go").Parse(string(mg))
	if err != nil {
		return fmt.Errorf("failed to parse main.go template: %w", err)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, struct{ ProjectName string }{config.ProjectName})
	if err != nil {
		return fmt.Errorf("failed to execute main.go template: %w", err)
	}

	mainPath := filepath.Join(config.TargetDir, "main.go")
	if _, err = os.Stat(mainPath); !os.IsNotExist(err) {
		return fmt.Errorf("main.go already exists at '%s'", mainPath)
	}

	if err := writeStringToFile(mainPath, buf.String()); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

func execCommand(config *Config, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = config.TargetDir
	
	if output, err := c.CombinedOutput(); err != nil {
		if len(output) > 0 {
			return fmt.Errorf("command '%s %v' failed: %s", cmd, args, string(output))
		}
		return fmt.Errorf("command '%s %v' failed: %w", cmd, args, err)
	}
	
	return nil
}

func binExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func writeStringToFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", path, err)
	}
	
	defer func() {
		_ = f.Close()
	}()

	if _, err = f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file '%s': %w", path, err)
	}
	return nil
}
