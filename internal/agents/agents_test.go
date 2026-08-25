package agents

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseModelTable(t *testing.T) {
	// The table both pi and prime-agent emit, plus the Node warning line
	// prime-agent prints to stderr when FORCE_COLOR is set. The warning must
	// not leak into the model list as a bogus "provider/model" pair.
	in := strings.Join([]string{
		"provider           model                    context  max-out  thinking  images",
		"litellm-anthropic  anthropic.claude-opus-5  1M       128K     yes       yes",
		"(node:86430) Warning: The 'NO_COLOR' env is ignored due to the 'FORCE_COLOR' env being set.",
		"litellm-anthropic  anthropic.claude-opus-5  1M       128K     yes       yes",
		"amazon-bedrock     amazon.nova-2-lite-v1:0  128K     4.1K     yes       yes",
	}, "\n") + "\n"

	got, err := parseModelTable([]byte(in))
	if err != nil {
		t.Fatalf("parseModelTable: %v", err)
	}
	want := []string{
		"amazon-bedrock/amazon.nova-2-lite-v1:0",
		"litellm-anthropic/anthropic.claude-opus-5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseModelTable:\n got %q\nwant %q", got, want)
	}
}

func TestWithoutEnv(t *testing.T) {
	env := []string{"AWS_PROFILE=cpe", "HOME=/home/u", "PATH=/bin"}
	got := withoutEnv(env, "AWS_PROFILE")
	want := []string{"HOME=/home/u", "PATH=/bin"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("withoutEnv:\n got %q\nwant %q", got, want)
	}
	// The input slice must not be mutated.
	if len(env) != 3 {
		t.Errorf("withoutEnv mutated its input: %q", env)
	}
}
