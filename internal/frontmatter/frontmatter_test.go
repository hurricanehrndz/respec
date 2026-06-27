package frontmatter

import (
	"bytes"
	"strings"
	"testing"
)

const bodyText = "# Plan\n\nSome **prose**.\n\n- a\n- b\n\n```go\nx := 1\n```\n"

func TestSetKeepsBodyByteIdentical(t *testing.T) {
	doc := "---\ntitle: my plan\nstatus: ready\n---\n" + bodyText
	out, err := SetString([]byte(doc), "spec_sha256", "deadbeef")
	if err != nil {
		t.Fatalf("SetString: %v", err)
	}
	_, body, has := split(out)
	if !has {
		t.Fatal("output lost its frontmatter")
	}
	if string(body) != bodyText {
		t.Errorf("body changed:\n got %q\nwant %q", body, bodyText)
	}
}

func TestSetUpdatesExistingKeyInPlace(t *testing.T) {
	doc := "---\nspec_sha256: oldhash\ntitle: t\n---\n" + bodyText
	out, err := SetString([]byte(doc), "spec_sha256", "newhash")
	if err != nil {
		t.Fatal(err)
	}
	v, ok, err := GetString(out, "spec_sha256")
	if err != nil || !ok {
		t.Fatalf("GetString after update: ok=%v err=%v", ok, err)
	}
	if v != "newhash" {
		t.Errorf("spec_sha256 = %q, want newhash", v)
	}
	// Existing sibling key must survive.
	if title, ok, _ := GetString(out, "title"); !ok || title != "t" {
		t.Errorf("title not preserved: %q ok=%v", title, ok)
	}
	_, body, _ := split(out)
	if string(body) != bodyText {
		t.Error("body changed on in-place update")
	}
}

func TestSetCreatesFrontmatterWhenAbsent(t *testing.T) {
	doc := bodyText // no frontmatter
	out, err := SetString([]byte(doc), "spec_sha256", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("---\n")) {
		t.Error("expected created frontmatter delimiter")
	}
	_, body, has := split(out)
	if !has {
		t.Fatal("no frontmatter after create")
	}
	if string(body) != bodyText {
		t.Errorf("original content not preserved as body: %q", body)
	}
}

func TestRoundTripIdempotent(t *testing.T) {
	doc := "---\nspec_sha256: h\n---\n" + bodyText
	once, _ := SetString([]byte(doc), "spec_sha256", "h")
	twice, _ := SetString(once, "spec_sha256", "h")
	if !bytes.Equal(once, twice) {
		t.Error("repeated SetString with same value is not idempotent")
	}
}

func TestGetMissing(t *testing.T) {
	doc := "---\ntitle: t\n---\n" + bodyText
	if _, ok, _ := GetString([]byte(doc), "spec_sha256"); ok {
		t.Error("expected spec_sha256 absent")
	}
	if _, ok, _ := GetString([]byte(strings.TrimPrefix(doc, "---\ntitle: t\n---\n")), "x"); ok {
		t.Error("expected no frontmatter -> not present")
	}
}
