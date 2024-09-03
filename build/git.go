package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var warnedAboutGit bool

// RunGit runs a git subcommand and returns its output.
// The command must complete successfully.
func RunGit(args ...string) string {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
			if !warnedAboutGit {
				log.Println("Warning: can't find 'git' in PATH")
				warnedAboutGit = true
			}
			return ""
		}
		log.Fatal(strings.Join(cmd.Args, " "), ": ", err, "\n", stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

// getDate returns original date of the commit
func getDate(commit string) string {
	if commit == "" {
		return ""
	}
	out := RunGit("show", "-s", "--format=%ct", commit)
	if out == "" {
		return ""
	}
	date, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse git commit date: %v", err))
	}
	return time.Unix(date, 0).Format("20060102")
}

// readGitFile returns content of file in .git directory.
func readGitFile(file string) string {
	content, err := os.ReadFile(filepath.Join(".git", file))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}
