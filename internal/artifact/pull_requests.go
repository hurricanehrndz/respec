package artifact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var pullRequestColumns = []string{"Repository", "PR", "Phase", "Base", "Depends on", "Merge order", "Reason", "State"}
var pullRequestStates = []string{"open", "draft", "merged", "closed", "superseded"}

// The section is opt-in so older plans remain valid. Lint checks the record's
// structure, not remote state or the dependency reasoning behind merge order.
func appendPullRequestProblems(p []string, pl Plan) []string {
	doc := goldmark.New(goldmark.WithExtensions(extension.Table)).Parser().Parse(text.NewReader(pl.body))
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		h, ok := n.(*ast.Heading)
		if !ok || h.Level != 2 || !strings.EqualFold(strings.TrimSpace(nodeText(h, pl.body)), "Pull requests") {
			continue
		}
		found := false
		for c := n.NextSibling(); c != nil; c = c.NextSibling() {
			if next, ok := c.(*ast.Heading); ok && next.Level <= h.Level {
				break
			}
			if table, ok := c.(*extast.Table); ok {
				found = true
				p = validatePullRequestTable(p, table, pl.body)
			}
		}
		if !found {
			p = append(p, "Pull requests: missing table")
		}
	}
	return p
}

func validatePullRequestTable(p []string, table *extast.Table, body []byte) []string {
	columns := map[string]int{}
	header := table.FirstChild()
	for c, i := header.FirstChild(), 0; c != nil; c, i = c.NextSibling(), i+1 {
		name := strings.ToLower(strings.TrimSpace(nodeText(c, body)))
		if _, exists := columns[name]; exists {
			p = append(p, "Pull requests: duplicate column: "+name)
		}
		columns[name] = i
	}
	missing := false
	for _, name := range pullRequestColumns {
		if _, ok := columns[strings.ToLower(name)]; !ok {
			p = append(p, "Pull requests: missing column: "+name)
			missing = true
		}
	}
	if missing {
		return p
	}
	for row, i := header.NextSibling(), 1; row != nil; row, i = row.NextSibling(), i+1 {
		var cells []string
		for c := row.FirstChild(); c != nil; c = c.NextSibling() {
			cells = append(cells, strings.TrimSpace(nodeText(c, body)))
		}
		for _, name := range pullRequestColumns {
			col := columns[strings.ToLower(name)]
			value := ""
			if col < len(cells) {
				value = cells[col]
			}
			prefix := fmt.Sprintf("Pull requests row %d: ", i)
			if value == "" {
				p = append(p, prefix+"missing "+name)
				continue
			}
			switch name {
			case "State":
				if !oneOf(value, pullRequestStates) {
					p = append(p, fmt.Sprintf("%sinvalid State %q (allowed: %s)", prefix, value, strings.Join(pullRequestStates, "|")))
				}
			case "Merge order":
				if order, err := strconv.Atoi(value); err != nil || order < 1 {
					p = append(p, prefix+"Merge order must be a positive integer")
				}
			}
		}
	}
	return p
}
