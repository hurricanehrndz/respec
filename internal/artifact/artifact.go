// Package artifact defines the typed frontmatter schema for the three respec
// store artifacts (research.md, spec.md, plan.md) and validates them: required
// frontmatter fields, allowed status values, and required body sections.
package artifact

import (
	"fmt"
	"strings"

	"github.com/hurricanehrndz/respec/internal/frontmatter"
	"gopkg.in/yaml.v3"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Kind identifies one of the three store artifacts.
type Kind string

const (
	KindResearch Kind = "research"
	KindSpec     Kind = "spec"
	KindPlan     Kind = "plan"
)

// StatusDone is the terminal plan status; a depends_on reference is satisfied
// only when the dependency's plan reaches it.
const StatusDone = "done"

// Allowed status values per artifact.
var (
	researchStatuses = []string{"draft", "complete"}
	specStatuses     = []string{"draft", "approved"}
	planStatuses     = []string{"draft", "approved", "in-progress", StatusDone}
	executionModes   = []string{"manual", "auto"}
)

// Research is the typed frontmatter of research.md. The agent writes topic,
// tags, and status; respec stamp fills date, repo, repo_path, and git_commit.
type Research struct {
	Topic     string   `yaml:"topic"`
	Tags      []string `yaml:"tags"`
	Status    string   `yaml:"status"`
	Date      string   `yaml:"date"`
	Repo      string   `yaml:"repo"`
	RepoPath  string   `yaml:"repo_path"`
	GitCommit string   `yaml:"git_commit"`
	body      []byte
}

// Spec is the typed frontmatter of spec.md. The agent writes title, status, and
// tags; respec stamp fills date.
type Spec struct {
	Title  string   `yaml:"title"`
	Status string   `yaml:"status"`
	Tags   []string `yaml:"tags"`
	Date   string   `yaml:"date"`
	body   []byte
}

// Plan is the typed frontmatter of plan.md. The agent writes title, status,
// optional execution_mode, and optional depends_on (slugs of efforts that must
// finish first); respec stamp fills spec_sha256.
type Plan struct {
	Title         string   `yaml:"title"`
	Status        string   `yaml:"status"`
	ExecutionMode string   `yaml:"execution_mode"`
	SpecSHA       string   `yaml:"spec_sha256"`
	DependsOn     []string `yaml:"depends_on"`
	body          []byte
}

// IsDone reports whether the plan has reached its terminal status.
func (pl Plan) IsDone() bool { return pl.Status == StatusDone }

// Artifact is the common contract: report validation problems (empty = valid).
type Artifact interface {
	Validate() []string
}

// StatusOf returns the frontmatter status of any artifact kind.
func StatusOf(a Artifact) string {
	switch v := a.(type) {
	case Research:
		return v.Status
	case Spec:
		return v.Status
	case Plan:
		return v.Status
	default:
		return ""
	}
}

// Parse decodes data into the typed artifact for kind.
func Parse(kind Kind, data []byte) (Artifact, error) {
	fm, _ := frontmatter.Raw(data)
	body := frontmatter.Body(data)
	switch kind {
	case KindResearch:
		var r Research
		if err := unmarshal(fm, &r); err != nil {
			return nil, err
		}
		r.body = body
		return r, nil
	case KindSpec:
		var s Spec
		if err := unmarshal(fm, &s); err != nil {
			return nil, err
		}
		s.body = body
		return s, nil
	case KindPlan:
		var p Plan
		if err := unmarshal(fm, &p); err != nil {
			return nil, err
		}
		p.body = body
		return p, nil
	default:
		return nil, fmt.Errorf("unknown artifact kind %q", kind)
	}
}

// Validate reports problems with research.md.
func (r Research) Validate() []string {
	var p []string
	p = appendMissing(p, "topic", r.Topic)
	p = appendMissing(p, "status", r.Status)
	p = appendMissing(p, "date", r.Date)
	p = appendMissing(p, "repo", r.Repo)
	p = appendMissing(p, "repo_path", r.RepoPath)
	p = appendMissing(p, "git_commit", r.GitCommit)
	p = appendBadStatus(p, r.Status, researchStatuses)
	hs := headings(r.body)
	p = appendMissingSections(p, hs, "Research Question", "Summary", "Findings", "Open Questions")
	return p
}

// Validate reports problems with spec.md.
func (s Spec) Validate() []string {
	var p []string
	p = appendMissing(p, "title", s.Title)
	p = appendMissing(p, "status", s.Status)
	p = appendMissing(p, "date", s.Date)
	p = appendBadStatus(p, s.Status, specStatuses)
	hs := headings(s.body)
	p = appendMissingSections(p, hs, "Requirements")
	return p
}

// Validate reports problems with plan.md.
func (pl Plan) Validate() []string {
	var p []string
	p = appendMissing(p, "title", pl.Title)
	p = appendMissing(p, "status", pl.Status)
	p = appendMissing(p, "spec_sha256", pl.SpecSHA)
	p = appendBadStatus(p, pl.Status, planStatuses)
	if pl.ExecutionMode != "" && !oneOf(pl.ExecutionMode, executionModes) {
		p = append(p, fmt.Sprintf("invalid execution_mode %q (allowed: %s)", pl.ExecutionMode, strings.Join(executionModes, "|")))
	}
	hs := headings(pl.body)
	p = appendMissingSections(p, hs, "Overview", "Automated Verification", "Manual Verification")
	if !hasHeadingPrefix(hs, "phase") {
		p = append(p, "missing section: at least one Phase")
	}
	return p
}

func appendMissing(p []string, field, value string) []string {
	if strings.TrimSpace(value) == "" {
		return append(p, fmt.Sprintf("missing frontmatter: %s", field))
	}
	return p
}

func appendBadStatus(p []string, got string, allowed []string) []string {
	if got == "" || oneOf(got, allowed) {
		return p
	}
	return append(p, fmt.Sprintf("invalid status %q (allowed: %s)", got, strings.Join(allowed, "|")))
}

func appendMissingSections(p []string, hs []string, required ...string) []string {
	for _, s := range required {
		if !hasHeading(hs, s) {
			p = append(p, "missing section: "+s)
		}
	}
	return p
}

func oneOf(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

func hasHeading(hs []string, name string) bool {
	return oneOf(strings.ToLower(name), hs)
}

func hasHeadingPrefix(hs []string, prefix string) bool {
	prefix = strings.ToLower(prefix)
	for _, h := range hs {
		if strings.HasPrefix(h, prefix) {
			return true
		}
	}
	return false
}

// headings returns the lowercased, trimmed text of every heading in body.
func headings(body []byte) []string {
	doc := goldmark.New().Parser().Parse(text.NewReader(body))
	var hs []string
	// The walker never returns an error, so ast.Walk cannot fail.
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			hs = append(hs, strings.ToLower(strings.TrimSpace(nodeText(h, body))))
		}
		return ast.WalkContinue, nil
	})
	return hs
}

// nodeText concatenates the raw text of a node's descendant text segments,
// avoiding the deprecated ast.Node.Text method.
func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
		default:
			b.WriteString(nodeText(c, src))
		}
	}
	return b.String()
}

func unmarshal(fm []byte, into any) error {
	if len(strings.TrimSpace(string(fm))) == 0 {
		return nil
	}
	if err := yaml.Unmarshal(fm, into); err != nil {
		return fmt.Errorf("parsing frontmatter: %w", err)
	}
	return nil
}
