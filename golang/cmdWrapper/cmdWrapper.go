// Package cmdWrapper provides a unified interface for executing shell commands
// with consistent error handling and platform-specific optimizations.
package cmdWrapper

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// Options holds configuration for command execution
type Options struct {
	Dir        string        // Working directory
	HideWindow bool          // Hide command window on Windows
	Stdout     *bytes.Buffer // Custom stdout buffer (nil uses os.Stdout)
	Stderr     *bytes.Buffer // Custom stderr buffer (nil uses os.Stderr)
}

// defaultOptions returns default options
func defaultOptions() Options {
	return Options{
		Dir:        "",
		HideWindow: runtime.GOOS == "windows", // Hide window by default on Windows
		Stdout:     nil,
		Stderr:     nil,
	}
}

// Option is a function that modifies Options
type Option func(*Options)

// WithDir sets the working directory
func WithDir(dir string) Option {
	return func(o *Options) {
		o.Dir = dir
	}
}

// WithHideWindow sets whether to hide the command window on Windows
func WithHideWindow(hide bool) Option {
	return func(o *Options) {
		o.HideWindow = hide
	}
}

// WithStdout sets a custom stdout buffer
func WithStdout(buf *bytes.Buffer) Option {
	return func(o *Options) {
		o.Stdout = buf
	}
}

// WithStderr sets a custom stderr buffer
func WithStderr(buf *bytes.Buffer) Option {
	return func(o *Options) {
		o.Stderr = buf
	}
}

// createCommand creates an exec.Cmd with the given options
func createCommand(name string, args []string, opts Options) *exec.Cmd {
	cmd := exec.Command(name, args...)

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	// Set stdout/stderr
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	// Hide window on Windows if requested
	if runtime.GOOS == "windows" && opts.HideWindow {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	return cmd
}

// Run executes a command and returns any error
func Run(name string, args ...string) error {
	return RunWithOptions(name, args)
}

// RunWithOptions executes a command with options and returns any error
func RunWithOptions(name string, args []string, opts ...Option) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	cmd := createCommand(name, args, options)
	return cmd.Run()
}

// RunWithOutput executes a command and returns combined output
func RunWithOutput(name string, args ...string) (string, error) {
	return RunWithOutputAndOptions(name, args)
}

// RunWithOutputAndOptions executes a command with options and returns combined output
func RunWithOutputAndOptions(name string, args []string, opts ...Option) (string, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// For output capture, we need custom buffers
	var stdout, stderr bytes.Buffer
	options.Stdout = &stdout
	options.Stderr = &stderr

	cmd := createCommand(name, args, options)
	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String() + stderr.String()
	return output, err
}

// Start starts a command but does not wait for it to complete
func Start(name string, args ...string) error {
	return StartWithOptions(name, args)
}

// StartWithOptions starts a command with options but does not wait for it to complete
func StartWithOptions(name string, args []string, opts ...Option) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	cmd := createCommand(name, args, options)
	return cmd.Start()
}

// RunWithDir is a convenience function to run a command in a specific directory
func RunWithDir(dir, name string, args ...string) error {
	return RunWithOptions(name, args, WithDir(dir))
}

// RunWithDirAndOutput is a convenience function to run a command in a specific directory and get output
func RunWithDirAndOutput(dir, name string, args ...string) (string, error) {
	return RunWithOutputAndOptions(name, args, WithDir(dir))
}
