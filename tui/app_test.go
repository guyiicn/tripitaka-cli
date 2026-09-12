package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/guyiicn/tripitaka-cli/catalog"
	"github.com/guyiicn/tripitaka-cli/model"
)

type multiVolumeLibrary struct{}

func (multiVolumeLibrary) Catalog() []catalog.Entry { return nil }
func (multiVolumeLibrary) FirstJuan(string) string  { return "001" }
func (multiVolumeLibrary) Juans(string) []string    { return []string{"001", "002"} }
func (multiVolumeLibrary) Load(id, juan string) (*model.Document, error) {
	return &model.Document{ID: id, Juan: juan, Title: "測試經", Text: "如是我聞一時佛在"}, nil
}

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

func TestCeremonyAndVolumeSequence(t *testing.T) {
	a := New(multiVolumeLibrary{}, nil)
	a.openItem(navItem{id: "T0001", juan: "001"})
	if a.ceremony != ceremonyLotus {
		t.Fatalf("first open ceremony = %v, want lotus", a.ceremony)
	}
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(80, 24)
	a.forward(s)
	if a.ceremony != ceremonyOpening {
		t.Fatalf("after lotus = %v, want opening verse", a.ceremony)
	}
	a.forward(s)
	if a.ceremony != ceremonyText {
		t.Fatalf("after opening = %v, want text", a.ceremony)
	}
	a.switchVolume(1)
	if a.doc.Juan != "002" || a.juanIndex != 1 {
		t.Fatalf("switched to juan %q index %d", a.doc.Juan, a.juanIndex)
	}
	a.start = max(0, len(a.book.Columns)-a.visibleColumns(s))
	a.forward(s)
	if a.ceremony != ceremonyDedication {
		t.Fatalf("after final volume = %v, want dedication", a.ceremony)
	}
}

func TestJuanLabel(t *testing.T) {
	for input, want := range map[string]string{
		"001": "一", "006": "六", "010": "十", "011": "十一",
		"020": "二十", "101": "一百零一", "1001": "一千零一",
	} {
		if got := juanLabel(input); got != want {
			t.Errorf("juanLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
