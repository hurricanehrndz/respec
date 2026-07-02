// Package store resolves paths within the central respec store: change
// directories (<store>/<repo>/<slug>/) and the sibling artifact files within
// them. A change is grouped under its primary repo (an owner-repo slug) and
// identified by its slug; the effort date lives in frontmatter, not the path.
package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	researchFile = "research.md"
	specFile     = "spec.md"
	planFile     = "plan.md"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Change identifies one effort in the store. Slug is the effort's identity
// within its repo (and doubles as a same-repo depends_on reference; cross-repo
// references are "repo/slug").
type Change struct {
	Repo string
	Slug string
	Dir  string
}

// Changes walks the store rooted at root and returns every change directory
// (a <repo>/<slug>/ dir containing a plan.md). A missing store yields no
// changes (not an error).
func Changes(root string) ([]Change, error) {
	repos, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Change
	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}
		repoDir := filepath.Join(root, repo.Name())
		entries, err := os.ReadDir(repoDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(repoDir, e.Name())
			if _, err := os.Stat(PlanPath(dir)); err != nil {
				continue // only directories that are real changes
			}
			out = append(out, Change{Repo: repo.Name(), Slug: e.Name(), Dir: dir})
		}
	}
	return out, nil
}

// Store is rooted at an (already ~-expanded) absolute store path.
type Store struct {
	Root string
}

// New returns a Store rooted at root.
func New(root string) Store {
	return Store{Root: root}
}

// Slugify lowercases s and replaces runs of non-alphanumerics with hyphens,
// trimming leading/trailing hyphens.
func Slugify(s string) string {
	s = slugRe.ReplaceAllString(toLower(s), "-")
	return trimHyphen(s)
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func trimHyphen(s string) string {
	for len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == '-' {
		s = s[:len(s)-1]
	}
	return s
}

// ChangeDir builds a change-dir path under the store for the given repo (an
// already-canonical owner-repo slug) and effort slug: <store>/<repo>/<slug>/.
func (s Store) ChangeDir(repo, slug string) string {
	return filepath.Join(s.Root, repo, Slugify(slug))
}

// SpecPath returns the spec.md path inside a change dir.
func SpecPath(changeDir string) string { return filepath.Join(changeDir, specFile) }

// PlanPath returns the plan.md path inside a change dir.
func PlanPath(changeDir string) string { return filepath.Join(changeDir, planFile) }

// ResearchPath returns the research.md path inside a change dir.
func ResearchPath(changeDir string) string { return filepath.Join(changeDir, researchFile) }

// ValidateChangeDir verifies that changeDir exists and contains a plan.md (the
// minimum needed for stamp/status). It returns an error describing what is
// missing otherwise.
func ValidateChangeDir(changeDir string) error {
	info, err := os.Stat(changeDir)
	if err != nil {
		return fmt.Errorf("change dir %q: %w", changeDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("change dir %q is not a directory", changeDir)
	}
	if _, err := os.Stat(PlanPath(changeDir)); err != nil {
		return fmt.Errorf("missing %s in %q: %w", planFile, changeDir, err)
	}
	return nil
}

// SHA256File returns the lowercase hex sha256 of the file at path.
func SHA256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
