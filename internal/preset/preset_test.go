package preset

import (
	"strings"
	"testing"
)

func TestSamples(t *testing.T) {
	for _, p := range All {
		s, err := p.Sample()
		if err != nil {
			t.Errorf("%s: %v", p.Key, err)
			continue
		}
		if strings.TrimSpace(s) == "" {
			t.Errorf("%s: empty sample", p.Key)
		}
	}
	if _, ok := Get("nope"); ok {
		t.Error("Get(nope)")
	}
	if p, _ := Get("terminal-session"); !p.IsTerminal() || p.SupportsLineNumbers() {
		t.Errorf("terminal-session flags wrong: %+v", p)
	}
	if p, _ := Get("error-log"); !p.SupportsLineNumbers() {
		t.Error("error-log should support line numbers")
	}
}
