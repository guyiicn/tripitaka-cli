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

func TestPunctuationIsPlacedRightOfHanGlyph(t *testing.T) {
	const glyphX = 12
	if got := markX(glyphX); got != 14 {
		t.Fatalf("mark x = %d, want 14 (after the two-cell Han glyph)", got)
	}
}
