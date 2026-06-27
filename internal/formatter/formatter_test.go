package formatter

import (
	"strings"
	"testing"
)

// A fixture exercising: heading, wide GFM table, fenced code, bare URL, and a
// long prose paragraph. Only the prose paragraphs should be reflowed.
const fixture = `# A heading that is intentionally quite long but headings are never reflowed

This is a long paragraph of prose that definitely exceeds the eighty column reflow width so it must wrap onto multiple lines when respec formats it.

| column one heading | column two heading | column three heading goes here |
|--------------------|--------------------|--------------------------------|
| a                  | b                  | c                              |

` + "```go\nfunc main() { fmt.Println(\"this code line is intentionally far longer than eighty columns and must not wrap\") }\n```" + `

See https://example.com/some/very/long/path/that/exceeds/the/reflow/width/by/a/lot for details.
`

func TestReflowsOnlyParagraphsLeavesNonProseByteIdentical(t *testing.T) {
	out := string(Format([]byte(fixture), 80))

	// Heading line preserved verbatim.
	if !strings.Contains(out, "# A heading that is intentionally quite long but headings are never reflowed\n") {
		t.Error("heading was altered")
	}
	// Table rows preserved verbatim (byte-identical).
	for _, row := range []string{
		"| column one heading | column two heading | column three heading goes here |",
		"|--------------------|--------------------|--------------------------------|",
		"| a                  | b                  | c                              |",
	} {
		if !strings.Contains(out, row) {
			t.Errorf("table row altered or missing: %q", row)
		}
	}
	// Fenced code preserved verbatim, including the over-long line.
	if !strings.Contains(out, "func main() { fmt.Println(\"this code line is intentionally far longer than eighty columns and must not wrap\") }") {
		t.Error("fenced code block was altered")
	}
	// The long prose paragraph wrapped (now contains a newline within it).
	para := firstParagraph(out)
	if !strings.Contains(para, "\n") {
		t.Error("long prose paragraph was not wrapped")
	}
	for _, line := range strings.Split(para, "\n") {
		if rlen(line) > 80 {
			t.Errorf("wrapped prose line exceeds 80 cols (%d): %q", rlen(line), line)
		}
	}
}

func TestAtomicTokensNotSplit(t *testing.T) {
	src := "Here is `inline code with spaces` and a tag <span class=\"x y\"> plus a url https://example.com/a/b/c that should all stay intact even though this paragraph is wide enough to wrap several times over the limit.\n"
	out := string(Format([]byte(src), 40))
	mustWhole(t, out, "`inline code with spaces`")
	mustWhole(t, out, "<span class=\"x y\">")
	mustWhole(t, out, "https://example.com/a/b/c")
}

func TestIdempotent(t *testing.T) {
	once := Format([]byte(fixture), 80)
	twice := Format(once, 80)
	if string(once) != string(twice) {
		t.Error("formatting is not idempotent")
	}
}

func TestNoParagraphsIsNoOp(t *testing.T) {
	src := []byte("# only a heading\n")
	if string(Format(src, 80)) != string(src) {
		t.Error("file with no paragraphs should be unchanged")
	}
}

// mustWhole asserts substr appears with no newline inserted inside it.
func mustWhole(t *testing.T, out, substr string) {
	t.Helper()
	if !strings.Contains(out, substr) {
		t.Errorf("atomic token was split or altered: %q not found whole in:\n%s", substr, out)
	}
}

func rlen(s string) int { return len([]rune(s)) }

// firstParagraph returns the first blank-line-delimited block that is plain
// prose (used to inspect the reflowed paragraph).
func firstParagraph(doc string) string {
	for _, block := range strings.Split(doc, "\n\n") {
		b := strings.TrimSpace(block)
		if strings.HasPrefix(b, "This is a long paragraph") {
			return block
		}
	}
	return ""
}
