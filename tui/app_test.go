package tui

import (
	"testing"

	"github.com/guyiicn/tripitaka-cli/model"
)

func TestPunctuationMarksAreSingleCellASCII(t *testing.T) {
	for _, mark := range []model.Mark{model.FullStop, model.Pause} {
		r := markRune(mark)
		if r > 127 {
			t.Fatalf("mark %v = %q; terminal side marks must be unambiguously one cell wide", mark, r)
		}
	}
}
