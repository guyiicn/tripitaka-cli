package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/guyiicn/tripitaka-cli/layout"
	"github.com/guyiicn/tripitaka-cli/model"
)

type App struct {
	doc       *model.Document
	book      layout.Book
	rows      int
	start     int
	showNotes bool
}

func New(doc *model.Document) *App {
	return &App{doc: doc, rows: 17}
}

func (a *App) Run() error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	a.reflow(s)
	for {
		a.draw(s)
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			a.reflow(s)
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyEscape || ev.Rune() == 'q':
				return nil
			case ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyPgDn || ev.Rune() == ' ' || ev.Rune() == 'l':
				a.start = min(len(a.book.Columns)-1, a.start+a.visibleColumns(s))
			case ev.Key() == tcell.KeyRight || ev.Key() == tcell.KeyPgUp || ev.Rune() == 'h':
				a.start = max(0, a.start-a.visibleColumns(s))
			case ev.Key() == tcell.KeyHome || ev.Rune() == 'g':
				a.start = 0
			case ev.Key() == tcell.KeyEnd || ev.Rune() == 'G':
				a.start = max(0, len(a.book.Columns)-a.visibleColumns(s))
			case ev.Rune() == ']':
				a.setRows(s, a.rows+1)
			case ev.Rune() == '[':
				a.setRows(s, max(8, a.rows-1))
			case ev.Rune() == 'n':
				a.showNotes = !a.showNotes
				a.reflow(s)
			}
		}
	}
}

func (a *App) setRows(s tcell.Screen, rows int) {
	if rows > 40 {
		rows = 40
	}
	a.rows = rows
	a.reflow(s)
}

func (a *App) reflow(s tcell.Screen) {
	pos := 0
	if a.start < len(a.book.Columns) {
		pos = a.book.Columns[a.start].SourceStart
	}
	a.book = layout.Build(a.doc, layout.Options{Rows: a.rows, ShowNotes: a.showNotes})
	a.start = a.book.ColumnAtSource(pos)
}

func (a *App) visibleColumns(s tcell.Screen) int {
	w, _ := s.Size()
	return max(1, (w-4)/3)
}

func (a *App) draw(s tcell.Screen) {
	s.Clear()
	w, h := s.Size()
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorDefault)
	muted := style.Foreground(tcell.ColorGray)
	red := style.Foreground(tcell.NewRGBColor(178, 52, 42))
	title := style.Bold(true).Foreground(tcell.NewRGBColor(190, 130, 45))
	note := style.Foreground(tcell.ColorDarkCyan)

	putString(s, 1, 0, fmt.Sprintf("大藏經  %s  卷%s", a.doc.Title, a.doc.Juan), title, w-2)
	availableRows := max(1, h-3)
	rows := min(a.rows, availableRows)
	cols := a.book.Page(a.start, a.visibleColumns(s))
	for ci, col := range cols {
		x := w - 3 - ci*3
		if x < 0 {
			break
		}
		for y, cell := range col.Cells {
			if y >= rows {
				break
			}
			cs := style
			if cell.Kind == "fen" {
				cs = red.Bold(true)
			} else if cell.Kind == "xu" || cell.Kind == "note" {
				cs = note
			}
			r := cell.Rune
			if cell.Annotation != "" {
				r = '□'
			}
			s.SetContent(x, y+1, r, nil, cs)
			if cell.Mark == model.FullStop {
				s.SetContent(x-1, y+1, '。', nil, red)
			} else if cell.Mark == model.Pause {
				s.SetContent(x-1, y+1, '、', nil, red)
			}
		}
	}
	pct := 100
	if len(a.book.Columns) > 1 {
		pct = a.start * 100 / (len(a.book.Columns) - 1)
	}
	status := fmt.Sprintf("←/Space 后翻  → 前翻  [/] 每列%d字  n 夹注:%s  q 退出  %d%%", a.rows, onOff(a.showNotes), pct)
	putString(s, 1, h-1, status, muted, w-2)
	s.Show()
}

func putString(s tcell.Screen, x, y int, value string, style tcell.Style, limit int) {
	used := 0
	for _, r := range value {
		width := 1
		if r >= 0x1100 {
			width = 2
		}
		if used+width > limit {
			break
		}
		s.SetContent(x+used, y, r, nil, style)
		used += width
	}
}

func onOff(v bool) string {
	if v {
		return "开"
	}
	return "关"
}
