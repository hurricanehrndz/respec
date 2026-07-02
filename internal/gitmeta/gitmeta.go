// Package gitmeta reads deterministic provenance from a git working tree:
// the origin remote URL, the HEAD commit, and the repository top-level path.
// respec stamp uses these to fill artifact frontmatter so the agent never has
// to shell out to git itself.
package gitmeta

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Meta holds provenance read from a git repository.
type Meta struct {
	Remote   string // origin remote URL; empty when no origin is configured
	Commit   string // HEAD commit sha
	Toplevel string // absolute path of the repository root
}

// Read returns provenance for the git repository containing dir. It returns an
// error when dir is not inside a git work tree or git is unavailable. A missing
// origin remote is not an error (Remote is left empty).
func Read(dir string) (Meta, error) {
	top, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Meta{}, err
	}
	commit, err := run(dir, "rev-parse", "HEAD")
	if err != nil {
		return Meta{}, err
	}
	remote, _ := run(dir, "remote", "get-url", "origin") // optional
	return Meta{Remote: remote, Commit: commit, Toplevel: top}, nil
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}
