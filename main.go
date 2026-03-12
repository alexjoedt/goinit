package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

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

// runInteractiveMode prompts the user for project configuration options.
func runInteractiveMode(config *Config, r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)

	_, _ = fmt.Fprintln(w, "Interactive Go Project Setup")
	_, _ = fmt.Fprintln(w, "============================")

	readLine := func(prompt string) (string, error) {
		_, _ = fmt.Fprint(w, prompt)
		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return strings.TrimSpace(input), nil
	}

	name, err := readLine("Project name: ")
	if err != nil {
		return err
	}
	config.ProjectName = name
	if config.ProjectName == "" {
		return errors.New("project name is required")
	}

	module, err := readLine(fmt.Sprintf("Go module name (default: %s): ", config.ProjectName))
	if err != nil {
		return err
	}
	if module != "" {
		config.ModuleName = module
	}

	taskfile, err := readLine("Add Taskfile.yml? (y/N): ")
	if err != nil {
		return err
	}
	config.WithTaskfile = strings.ToLower(taskfile) == "y"

	makefile, err := readLine("Add Makefile? (y/N): ")
	if err != nil {
		return err
	}
	config.WithMakefile = strings.ToLower(makefile) == "y"

	dockerfile, err := readLine("Add Dockerfile? (y/N): ")
	if err != nil {
		return err
	}
	config.WithDockerfile = strings.ToLower(dockerfile) == "y"

	verbose, err := readLine("Verbose output? (y/N): ")
	if err != nil {
		return err
	}
	config.Verbose = strings.ToLower(verbose) == "y"

	_, _ = fmt.Fprintln(w)
	return nil
}

func main() {
	config := &Config{}
	var showHelp, showVersion bool

	// Help and version flags
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

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

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("goinit version %s\n", version)
		os.Exit(0)
	}

	config.ProjectName = flag.Arg(0)

	if err := run(config, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Templates
var (
	//go:embed templates
	tplFiles embed.FS
)

// run creates the go project directory.
func run(config *Config, stdout, stderr io.Writer) error {
	if !binExists("git") {
		return errors.New("git not found in PATH; install from https://git-scm.com/downloads")
	}
	if !binExists("go") {
		return errors.New("go not found in PATH; install from https://golang.org/doc/install")
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Handle interactive mode
	if config.Interactive {
		if err := runInteractiveMode(config, os.Stdin, stdout); err != nil {
			return fmt.Errorf("interactive mode failed: %w", err)
		}
	}

	// Normalize and validate project name.
	config.ProjectName = strings.TrimSpace(config.ProjectName)
	if config.ProjectName == "" {
		return errors.New(`project name is required

Examples:
  goinit my-app                    # Simple project
  goinit --interactive             # Interactive setup
  goinit -t -d my-webapp           # With Taskfile and Dockerfile

Use 'goinit --help' for more information`)
	}

	config.ProjectName = filepath.Base(config.ProjectName)

	// Check for invalid characters in project name
	if strings.ContainsAny(config.ProjectName, `<>:"/\|?*`) {
		return fmt.Errorf("Project name '%s' contains invalid characters. Use only letters, numbers, hyphens, and underscores", config.ProjectName) //nolint: staticcheck
	}

	config.TargetDir = filepath.Join(wd, config.ProjectName)
	if _, err := os.Stat(config.TargetDir); !os.IsNotExist(err) {
		return fmt.Errorf(`directory '%s' already exists.

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

	progress := NewProgressTracker(steps, stdout)
	progress.Start(config)

	info(stderr, config, "Creating project in %s", config.TargetDir)

	if err := os.MkdirAll(config.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory '%s': %w", config.TargetDir, err)
	}
	progress.NextStep(config)

	// Init Go Module
	info(stderr, config, "Initializing Go module")
	if err := initGoMod(config); err != nil {
		return fmt.Errorf(`failed to initialize Go module: %w

Make sure Go is properly installed:
  • Download from: https://golang.org/dl/
  • Verify with: go version`, err)
	}
	progress.NextStep(config)

	// Init Git Repo
	info(stderr, config, "Initializing Git repository")
	if err := initGitRepo(config); err != nil {
		return fmt.Errorf(`failed to initialize Git repository: %w

Make sure Git is properly installed:
  • Download from: https://git-scm.com/downloads
  • Verify with: git --version`, err)
	}
	progress.NextStep(config)

	// Create .gitignore
	info(stderr, config, "Creating .gitignore")
	if err := createGitignore(config); err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}
	progress.NextStep(config)

	// Create README.md
	info(stderr, config, "Creating README.md")
	if err := createReadme(config); err != nil {
		return fmt.Errorf("failed to create README.md file: %w", err)
	}
	progress.NextStep(config)

	// Create main.go
	info(stderr, config, "Creating main.go")
	if err := createMainDotGo(config); err != nil {
		return fmt.Errorf("failed to create main.go file: %w", err)
	}
	progress.NextStep(config)

	if config.WithTaskfile {
		if !binExists("task") {
			warn(stderr, "Task binary not found. Install from: https://taskfile.dev/installation/")
		}

		info(stderr, config, "Creating Taskfile.yml")
		if err := createTaskfile(config); err != nil {
			warn(stderr, "Failed to create Taskfile.yml: %v", err)
		} else {
			progress.NextStep(config)
		}
	}

	if config.WithMakefile {
		if !binExists("make") {
			warn(stderr, "Make binary not found. Install build tools for your system.")
		}

		info(stderr, config, "Creating Makefile")
		if err := createMakefile(config); err != nil {
			warn(stderr, "Failed to create Makefile: %v", err)
		} else {
			progress.NextStep(config)
		}
	}

	if config.WithDockerfile {
		if !binExists("docker") {
			warn(stderr, "Docker binary not found. Install from: https://docs.docker.com/get-docker/")
		}

		info(stderr, config, "Creating Dockerfile")
		if err := createDockerfile(config); err != nil {
			warn(stderr, "Failed to create Dockerfile: %v", err)
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

// createFileFromTemplate reads the named embedded template, renders it with
// config data, and writes the result to destName inside config.TargetDir.
// An optional delimPair (left, right) overrides the default "{{" "}}" delimiters.
func createFileFromTemplate(config *Config, templatePath, destName string, delimPair ...string) error {
	raw, err := tplFiles.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template %q: %w", templatePath, err)
	}

	t := template.New(destName)
	if len(delimPair) == 2 {
		t = t.Delims(delimPair[0], delimPair[1])
	}
	t, err = t.Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template %q: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, struct{ ProjectName string }{config.ProjectName}); err != nil {
		return fmt.Errorf("execute template %q: %w", templatePath, err)
	}

	dest := filepath.Join(config.TargetDir, destName)
	if _, err = os.Stat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("%q already exists", dest)
	}
	return writeStringToFile(dest, buf.String())
}

func createTaskfile(config *Config) error {
	return createFileFromTemplate(config, "templates/Taskfile.yml", "Taskfile.yml", "[[", "]]")
}

func createDockerfile(config *Config) error {
	return createFileFromTemplate(config, "templates/Dockerfile", "Dockerfile")
}

func createMakefile(config *Config) error {
	return createFileFromTemplate(config, "templates/Makefile", "Makefile")
}

func createReadme(config *Config) error {
	return createFileFromTemplate(config, "templates/README.md", "README.md")
}

func createGitignore(config *Config) error {
	return createFileFromTemplate(config, "templates/gitignore", ".gitignore")
}

func createMainDotGo(config *Config) error {
	return createFileFromTemplate(config, "templates/main.go", "main.go")
}

func execCommand(config *Config, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = config.TargetDir
	output, err := c.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("command %q failed: %s: %w", cmd+" "+strings.Join(args, " "), bytes.TrimSpace(output), err)
		}
		return fmt.Errorf("command %q failed: %w", cmd+" "+strings.Join(args, " "), err)
	}
	return nil
}

func binExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func writeStringToFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
