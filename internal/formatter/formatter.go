// Package formatter reflows Markdown prose paragraphs to a target width while
// leaving every non-paragraph construct (tables, fenced code, headings, HTML
// blocks, etc.) byte-identical (R-11).
//
// Approach: parse with goldmark (GFM enabled so pipe tables are recognized as
// tables, not paragraphs), find the source byte-span of each top-level
// paragraph, and rewrite only those spans. Everything outside a reflowed
// paragraph span is copied through verbatim.
package formatter

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

type span struct {
	start, stop int
	reflowed    string
}

// Format reflows top-level prose paragraphs in src to width columns and returns
// the new document bytes. Non-paragraph bytes are preserved exactly.
func Format(src []byte, width int) []byte {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	doc := md.Parser().Parse(text.NewReader(src))

	var spans []span
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		p, ok := n.(*ast.Paragraph)
		if !ok {
			return ast.WalkContinue, nil
		}
		// shortcut: v1 reflows only top-level paragraphs. Paragraphs inside
		// list items or block quotes are left untouched to avoid disturbing
		// their indentation; revisit if indentation-aware reflow is needed.
		if p.Parent() == nil || p.Parent().Kind() != ast.KindDocument {
			return ast.WalkSkipChildren, nil
		}
		lines := p.Lines()
		if lines.Len() == 0 {
			return ast.WalkContinue, nil
		}
		start := lines.At(0).Start
		stop := lines.At(lines.Len() - 1).Stop
		tokens := tokenize(string(src[start:stop]))
		spans = append(spans, span{start: start, stop: stop, reflowed: wrap(tokens, width)})
		return ast.WalkContinue, nil
	})

	if len(spans) == 0 {
		return src
	}

	var out bytes.Buffer
	prev := 0
	for _, sp := range spans {
		out.Write(src[prev:sp.start])
		out.WriteString(sp.reflowed)
		prev = sp.stop
	}
	out.Write(src[prev:])
	return out.Bytes()
}

// tokenize splits paragraph source into wrap tokens, treating inline code
// spans and HTML tags as atomic (never split, even when they contain spaces).
// Bare URLs (autolinks) contain no spaces and are naturally atomic.
func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			flush()
			i++
		case c == '`':
			run := backtickRun(s, i)
			if end := closingBacktick(s, i+run, run); end != -1 {
				cur.WriteString(s[i:end]) // whole `code span` atomic
				i = end
			} else {
				cur.WriteByte(c)
				i++
			}
		case c == '<' && i+1 < len(s) && isTagStart(s[i+1]):
			if gt := strings.IndexByte(s[i:], '>'); gt != -1 {
				cur.WriteString(s[i : i+gt+1]) // whole <tag ...> atomic
				i += gt + 1
			} else {
				cur.WriteByte(c)
				i++
			}
		default:
			cur.WriteByte(c)
			i++
		}
	}
	flush()
	return tokens
}

func backtickRun(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

// closingBacktick finds the end index (exclusive) of a closing backtick run of
// exactly length run, searching from index from. Returns -1 if none.
func closingBacktick(s string, from, run int) int {
	i := from
	for i < len(s) {
		if s[i] == '`' {
			j := i
			for j < len(s) && s[j] == '`' {
				j++
			}
			if j-i == run {
				return j
			}
			i = j
		} else {
			i++
		}
	}
	return -1
}

func isTagStart(c byte) bool {
	return c == '/' || c == '!' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// wrap greedily packs tokens into lines no wider than width. A single token
// longer than width is placed on its own line rather than split.
func wrap(tokens []string, width int) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(tokens[0])
	lineLen := utf8.RuneCountInString(tokens[0])
	for _, tok := range tokens[1:] {
		tl := utf8.RuneCountInString(tok)
		if lineLen+1+tl > width {
			b.WriteByte('\n')
			b.WriteString(tok)
			lineLen = tl
		} else {
			b.WriteByte(' ')
			b.WriteString(tok)
			lineLen += 1 + tl
		}
	}
	return b.String()
}
