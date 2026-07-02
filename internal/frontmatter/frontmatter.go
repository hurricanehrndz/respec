// Package frontmatter reads and modifies the leading YAML frontmatter block of
// a Markdown document while leaving the document body byte-identical.
package frontmatter

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

var delim = []byte("---")

// split separates a leading `---\n…\n---\n` YAML block from the body. It
// returns the frontmatter content (between the delimiter lines), the body
// (everything after the closing delimiter line), and whether a frontmatter
// block was present.
func split(data []byte) (fm, body []byte, has bool) {
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, data, false
	}
	rest := data[4:]
	off := 0
	for {
		nl := bytes.IndexByte(rest[off:], '\n')
		var line []byte
		if nl == -1 {
			line = rest[off:]
		} else {
			line = rest[off : off+nl]
		}
		if bytes.Equal(bytes.TrimRight(line, "\r"), delim) {
			fm = rest[:off]
			if nl == -1 {
				body = nil
			} else {
				body = rest[off+nl+1:]
			}
			return fm, body, true
		}
		if nl == -1 {
			return nil, data, false // no closing delimiter
		}
		off += nl + 1
	}
}

// Body returns the document body with any leading frontmatter block removed.
// When no frontmatter is present the whole input is the body.
func Body(data []byte) []byte {
	_, body, _ := split(data)
	return body
}

// Raw returns the inner YAML frontmatter bytes (without the --- delimiter
// lines) and whether a frontmatter block was present.
func Raw(data []byte) ([]byte, bool) {
	fm, _, has := split(data)
	return fm, has
}

// GetString returns the string value of key from the frontmatter and whether it
// was present.
func GetString(data []byte, key string) (string, bool, error) {
	fm, _, has := split(data)
	if !has {
		return "", false, nil
	}
	m := map[string]interface{}{}
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return "", false, fmt.Errorf("parsing frontmatter: %w", err)
	}
	v, ok := m[key]
	if !ok || v == nil {
		// A key with a null value ("date:") is present but empty — report it
		// as such rather than stringifying nil to "<nil>".
		return "", ok, nil
	}
	return fmt.Sprintf("%v", v), true, nil
}

// KV is a frontmatter key/value pair for SetMany.
type KV struct {
	Key   string
	Value string
}

// SetString sets key to value in the document's frontmatter and returns the new
// document bytes. The body is preserved byte-for-byte. When no frontmatter
// exists, one is created and the original content becomes the body.
func SetString(data []byte, key, value string) ([]byte, error) {
	return SetMany(data, []KV{{Key: key, Value: value}})
}

// SetMany sets each pair's key to its value in a single order-preserving pass,
// returning the new document bytes. Existing keys are updated in place; new keys
// are appended in the order given. The body is preserved byte-for-byte, and
// frontmatter is created when absent (the original content becomes the body).
func SetMany(data []byte, kvs []KV) ([]byte, error) {
	fm, body, has := split(data)

	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if has && len(bytes.TrimSpace(fm)) > 0 {
		var doc yaml.Node
		if err := yaml.Unmarshal(fm, &doc); err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
		if len(doc.Content) > 0 && doc.Content[0].Kind == yaml.MappingNode {
			mapping = doc.Content[0]
		}
	}

	for _, kv := range kvs {
		setMapKey(mapping, kv.Key, kv.Value)
	}

	marshaled, err := yaml.Marshal(mapping)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(marshaled)
	out.WriteString("---\n")
	out.Write(body)
	return out.Bytes(), nil
}

// setMapKey updates an existing key's scalar value or appends a new key/value
// pair, preserving the order of existing keys.
func setMapKey(m *yaml.Node, key, value string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1].Kind = yaml.ScalarNode
			m.Content[i+1].Tag = "!!str"
			m.Content[i+1].Value = value
			m.Content[i+1].Style = 0
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
}
