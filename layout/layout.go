package layout

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/guyiicn/tripitaka-cli/model"
)

// Cell is one full-width terminal character cell.
type Cell struct {
	Rune        rune
	SourceIndex int
	Mark        model.Mark
	Kind        string
	Annotation  string
}

// Column is read from top to bottom. Columns are ordered from right to left.
type Column struct {
	Cells       []Cell
	Kind        string
	SourceStart int
}

// Book is a terminal-independent vertical layout.
type Book struct {
	Columns []Column
}

// Options controls logical typesetting. Rows is the fixed character count per column.
type Options struct {
	Rows      int
	ShowNotes bool
}

// Build converts CBETA source indices into right-to-left vertical columns.
func Build(d *model.Document, opt Options) Book {
	if opt.Rows < 4 {
		opt.Rows = 17
	}
	ju := intSet(d.Ju)
	dou := intSet(d.Dou)
	br := intSet(d.Breaks)
	headings := make(map[int][]model.Heading)
	for _, h := range d.Headings {
		headings[h.At] = append(headings[h.At], h)
	}
	notes := make(map[int][]string)
	if opt.ShowNotes {
		for _, n := range d.Notes {
			notes[n.At] = append(notes[n.At], n.Text)
		}
	}
	gaiji := make(map[int]string)
	for _, g := range d.Gaiji {
		gaiji[g.At] = g.Formula
	}

	var out Book
	var col *Column
	flush := func() { col = nil }
	add := func(c Cell) {
		if col == nil || len(col.Cells) >= opt.Rows {
			out.Columns = append(out.Columns, Column{SourceStart: c.SourceIndex})
			col = &out.Columns[len(out.Columns)-1]
		}
		col.Cells = append(col.Cells, c)
	}
	addHeading := func(h model.Heading) {
		flush()
		r := []rune(stripBrackets(h.Text))
		for len(r) > 0 {
			n := min(opt.Rows, len(r))
			cells := make([]Cell, n)
			for i := range cells {
				cells[i] = Cell{Rune: r[i], SourceIndex: h.At, Kind: h.Kind}
			}
			out.Columns = append(out.Columns, Column{Cells: cells, Kind: h.Kind, SourceStart: h.At})
			r = r[n:]
		}
		flush()
	}

	runes := []rune(d.Text)
	for i, r := range runes {
		for _, h := range headings[i] {
			addHeading(h)
		}
		if br[i] {
			flush()
		}
		mark := model.NoMark
		if ju[i] {
			mark = model.FullStop
		} else if dou[i] {
			mark = model.Pause
		}
		annotation := gaiji[i]
		add(Cell{Rune: r, SourceIndex: i, Mark: mark, Annotation: annotation})
		for _, note := range notes[i] {
			for _, nr := range []rune("〔" + note + "〕") {
				add(Cell{Rune: nr, SourceIndex: i, Kind: "note"})
			}
		}
	}
	return out
}

// Page returns up to count logical columns beginning at start.
func (b Book) Page(start, count int) []Column {
	if start < 0 {
		start = 0
	}
	if count < 1 || start >= len(b.Columns) {
		return nil
	}
	end := min(len(b.Columns), start+count)
	return b.Columns[start:end]
}

// ColumnAtSource locates the stable reading position after a resize/reflow.
func (b Book) ColumnAtSource(index int) int {
	// Several structural columns can share the same source position as the
	// following body text. In particular, volume headings commonly live at
	// position zero. A request for the beginning must include those columns.
	if len(b.Columns) == 0 || index <= b.Columns[0].SourceStart {
		return 0
	}
	i := sort.Search(len(b.Columns), func(i int) bool { return b.Columns[i].SourceStart > index })
	return i - 1
}

func intSet(values []int) map[int]bool {
	m := make(map[int]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}

func stripBrackets(s string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune("（）()〔〕【】[]「」『』《》〈〉", r) {
			return -1
		}
		return r
	}, s)
}

// RuneCount exists to make the source-index contract explicit: indices are Unicode code points.
func RuneCount(s string) int { return utf8.RuneCountInString(s) }
