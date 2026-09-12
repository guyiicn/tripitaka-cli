package layout

import (
	"testing"

	"github.com/guyiicn/tripitaka-cli/model"
)

func TestBuildVerticalColumnsAndMarks(t *testing.T) {
	d := &model.Document{Text: "觀自在菩薩行深", Ju: []int{3}, Breaks: []int{5}}
	b := Build(d, Options{Rows: 4})
	if got := len(b.Columns); got != 3 {
		t.Fatalf("columns = %d, want 3", got)
	}
	if got := string([]rune{b.Columns[0].Cells[0].Rune, b.Columns[0].Cells[1].Rune, b.Columns[0].Cells[2].Rune, b.Columns[0].Cells[3].Rune}); got != "觀自在菩" {
		t.Fatalf("first column = %q", got)
	}
	if b.Columns[0].Cells[3].Mark != model.FullStop {
		t.Fatal("full stop was not attached to its source character")
	}
	if got := b.Columns[2].SourceStart; got != 5 {
		t.Fatalf("paragraph column starts at %d, want 5", got)
	}
}

func TestHeadingGetsDedicatedColumn(t *testing.T) {
	d := &model.Document{Text: "如是我聞一時佛在", Headings: []model.Heading{{At: 4, Text: "序品第一", Kind: "fen"}}}
	b := Build(d, Options{Rows: 8})
	if len(b.Columns) != 3 || b.Columns[1].Kind != "fen" {
		t.Fatalf("unexpected columns: %#v", b.Columns)
	}
}

func TestColumnAtSourceSurvivesReflow(t *testing.T) {
	d := &model.Document{Text: "一二三四五六七八九十甲乙丙丁"}
	b := Build(d, Options{Rows: 5})
	if got := b.ColumnAtSource(11); got != 2 {
		t.Fatalf("column = %d, want 2", got)
	}
}
