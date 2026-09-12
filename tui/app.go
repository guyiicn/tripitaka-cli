package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/guyiicn/tripitaka-cli/catalog"
	"github.com/guyiicn/tripitaka-cli/layout"
	"github.com/guyiicn/tripitaka-cli/library"
	"github.com/guyiicn/tripitaka-cli/model"
	"github.com/guyiicn/tripitaka-cli/state"
)

type mode int

const (
	modeHome mode = iota
	modeSearch
	modeReader
	modeBookmarks
	modeRecent
)

type navItem struct {
	id, juan, title, detail string
	char                    int
	bookmarkID              int64
}

type App struct {
	library              library.Store
	state                *state.Store
	mode                 mode
	doc                  *model.Document
	book                 layout.Book
	rows, start          int
	showNotes            bool
	homeItems, listItems []navItem
	selected             int
	query                []rune
	results              []catalog.Entry
	message              string
}

func New(lib library.Store, st *state.Store) *App {
	return &App{library: lib, state: st, rows: 17, mode: modeHome}
}

// Start opens a volume before Run. It preserves the direct-JSON debugging
// workflow while the no-argument command starts at the home screen.
func (a *App) Start(id, juan string) {
	a.openItem(navItem{id: id, juan: juan})
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
	a.refreshHome()
	for {
		a.draw(s)
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			if a.mode == modeReader {
				a.reflow()
			}
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' {
				a.saveProgress()
				return nil
			}
			switch a.mode {
			case modeHome:
				a.handleHome(ev)
			case modeSearch:
				a.handleSearch(ev)
			case modeReader:
				a.handleReader(s, ev)
			case modeBookmarks, modeRecent:
				a.handleList(ev)
			}
		}
	}
}

func (a *App) handleHome(ev *tcell.EventKey) {
	switch {
	case ev.Rune() == '/':
		a.mode, a.query, a.selected = modeSearch, nil, 0
		a.updateSearch()
	case ev.Rune() == 'b':
		a.openBookmarks()
	case ev.Rune() == 'r':
		a.openRecent()
	case ev.Key() == tcell.KeyUp || ev.Rune() == 'k':
		a.selected = max(0, a.selected-1)
	case ev.Key() == tcell.KeyDown || ev.Rune() == 'j':
		a.selected = min(len(a.homeItems)-1, a.selected+1)
	case ev.Key() == tcell.KeyEnter && len(a.homeItems) > 0:
		a.openItem(a.homeItems[a.selected])
	}
}

func (a *App) handleSearch(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		a.mode, a.selected = modeHome, 0
		return
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(a.query) > 0 {
			a.query = a.query[:len(a.query)-1]
			a.selected = 0
			a.updateSearch()
		}
		return
	case tcell.KeyUp:
		a.selected = max(0, a.selected-1)
		return
	case tcell.KeyDown:
		a.selected = min(len(a.results)-1, a.selected+1)
		return
	case tcell.KeyEnter:
		if len(a.results) > 0 {
			e := a.results[a.selected]
			a.openItem(navItem{id: e.ID, juan: a.library.FirstJuan(e.ID), title: e.Title})
		}
		return
	}
	if ev.Rune() != 0 && ev.Modifiers()&tcell.ModCtrl == 0 {
		a.query = append(a.query, ev.Rune())
		a.selected = 0
		a.updateSearch()
	}
}

func (a *App) handleList(ev *tcell.EventKey) {
	switch {
	case ev.Key() == tcell.KeyEscape:
		a.mode, a.selected = modeHome, 0
	case ev.Key() == tcell.KeyUp || ev.Rune() == 'k':
		a.selected = max(0, a.selected-1)
	case ev.Key() == tcell.KeyDown || ev.Rune() == 'j':
		a.selected = min(len(a.listItems)-1, a.selected+1)
	case ev.Key() == tcell.KeyEnter && len(a.listItems) > 0:
		a.openItem(a.listItems[a.selected])
	case ev.Rune() == 'd' && a.mode == modeBookmarks && len(a.listItems) > 0:
		if err := a.state.DeleteBookmark(a.listItems[a.selected].bookmarkID); err != nil {
			a.message = err.Error()
		} else {
			a.openBookmarks()
			a.message = "书签已删除"
		}
	}
}

func (a *App) handleReader(s tcell.Screen, ev *tcell.EventKey) {
	switch {
	case ev.Key() == tcell.KeyEscape:
		a.saveProgress()
		a.refreshHome()
		a.mode, a.selected = modeHome, 0
	case ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyPgDn || ev.Rune() == ' ' || ev.Rune() == 'l':
		a.start = min(max(0, len(a.book.Columns)-1), a.start+a.visibleColumns(s))
		a.saveProgress()
	case ev.Key() == tcell.KeyRight || ev.Key() == tcell.KeyPgUp || ev.Rune() == 'h':
		a.start = max(0, a.start-a.visibleColumns(s))
		a.saveProgress()
	case ev.Key() == tcell.KeyHome || ev.Rune() == 'g':
		a.start = 0
		a.saveProgress()
	case ev.Key() == tcell.KeyEnd || ev.Rune() == 'G':
		a.start = max(0, len(a.book.Columns)-a.visibleColumns(s))
		a.saveProgress()
	case ev.Rune() == ']':
		a.setRows(a.rows + 1)
	case ev.Rune() == '[':
		a.setRows(max(8, a.rows-1))
	case ev.Rune() == 'n':
		a.showNotes = !a.showNotes
		a.reflow()
	case ev.Rune() == 'b':
		a.addBookmark()
	}
}

func (a *App) updateSearch() { a.results = catalog.Search(a.library.Catalog(), string(a.query), 20) }

func (a *App) openItem(item navItem) {
	if item.bookmarkID == 0 && item.char == 0 && a.state != nil {
		if p, err := a.state.LatestProgress(item.id); err == nil {
			item.juan, item.char = p.Juan, p.Char
		}
	}
	doc, err := a.library.Load(item.id, item.juan)
	if err != nil {
		a.message = "打开失败：" + err.Error()
		return
	}
	a.doc = doc
	a.book = layout.Build(doc, layout.Options{Rows: a.rows, ShowNotes: a.showNotes})
	a.start = a.book.ColumnAtSource(item.char)
	a.mode, a.message = modeReader, ""
	a.saveProgress()
}

func (a *App) setRows(rows int) {
	if rows > 40 {
		rows = 40
	}
	a.rows = rows
	a.reflow()
	a.saveProgress()
}
func (a *App) reflow() {
	if a.doc == nil {
		return
	}
	pos := a.currentChar()
	a.book = layout.Build(a.doc, layout.Options{Rows: a.rows, ShowNotes: a.showNotes})
	a.start = a.book.ColumnAtSource(pos)
}
func (a *App) currentChar() int {
	if a.start >= 0 && a.start < len(a.book.Columns) {
		return a.book.Columns[a.start].SourceStart
	}
	return 0
}
func (a *App) saveProgress() {
	if a.state == nil || a.doc == nil {
		return
	}
	pos := a.currentChar()
	_ = a.state.SaveProgress(state.Progress{SutraID: a.doc.ID, Juan: a.doc.Juan, Title: a.doc.Title, Char: pos, Context: snippet(a.doc.Text, pos, 12)})
}
func (a *App) addBookmark() {
	pos := a.currentChar()
	err := a.state.AddBookmark(state.Bookmark{SutraID: a.doc.ID, Juan: a.doc.Juan, Title: a.doc.Title, Char: pos, Snippet: snippet(a.doc.Text, pos, 18)})
	if err != nil {
		a.message = "书签失败：" + err.Error()
	} else {
		a.message = "书签已保存"
	}
}

func (a *App) refreshHome() {
	if a.state == nil {
		return
	}
	recent, _ := a.state.Recent(6)
	marks, _ := a.state.Bookmarks(6)
	a.homeItems = nil
	for _, p := range recent {
		a.homeItems = append(a.homeItems, navItem{id: p.SutraID, juan: p.Juan, title: p.Title, char: p.Char, detail: fmt.Sprintf("卷%s · 字 %d", p.Juan, p.Char+1)})
	}
	for _, b := range marks {
		a.homeItems = append(a.homeItems, navItem{id: b.SutraID, juan: b.Juan, title: b.Title, char: b.Char, detail: "书签 · " + b.Snippet, bookmarkID: b.ID})
	}
}
func (a *App) openBookmarks() {
	marks, _ := a.state.Bookmarks(100)
	a.listItems = nil
	for _, b := range marks {
		a.listItems = append(a.listItems, navItem{id: b.SutraID, juan: b.Juan, title: b.Title, char: b.Char, detail: b.Snippet, bookmarkID: b.ID})
	}
	a.mode, a.selected = modeBookmarks, 0
}
func (a *App) openRecent() {
	recent, _ := a.state.Recent(100)
	a.listItems = nil
	for _, p := range recent {
		a.listItems = append(a.listItems, navItem{id: p.SutraID, juan: p.Juan, title: p.Title, char: p.Char, detail: fmt.Sprintf("卷%s · 字 %d", p.Juan, p.Char+1)})
	}
	a.mode, a.selected = modeRecent, 0
}
func (a *App) visibleColumns(s tcell.Screen) int { w, _ := s.Size(); return max(1, (w-4)/3) }

func (a *App) draw(s tcell.Screen) {
	s.Clear()
	switch a.mode {
	case modeHome:
		a.drawHome(s)
	case modeSearch:
		a.drawSearch(s)
	case modeReader:
		a.drawReader(s)
	case modeBookmarks:
		a.drawList(s, "书签", "Enter 打开  d 删除  Esc 返回")
	case modeRecent:
		a.drawList(s, "阅读进度", "Enter 继续  Esc 返回")
	}
	s.Show()
}

func (a *App) drawHome(s tcell.Screen) {
	w, h := s.Size()
	base, muted, accent, selected := styles()
	center(s, 1, "大 藏 經", accent.Bold(true), w)
	putString(s, 2, 3, "/ 搜索经名、经号或拼音", muted, w-4)
	y := 5
	recentCount := 0
	for _, it := range a.homeItems {
		if !strings.HasPrefix(it.detail, "书签") {
			recentCount++
		}
	}
	putString(s, 2, y, "最近阅读", accent, w-4)
	y++
	if recentCount == 0 {
		putString(s, 4, y, "尚无阅读进度，按 / 搜索", muted, w-6)
		y += 2
	} else {
		for i := 0; i < recentCount && y < h-5; i++ {
			drawItem(s, 3, y, a.homeItems[i], i == a.selected, base, selected, w-5)
			y++
		}
		y++
	}
	putString(s, 2, y, "书签", accent, w-4)
	y++
	if len(a.homeItems) == recentCount {
		putString(s, 4, y, "尚无书签", muted, w-6)
	} else {
		for i := recentCount; i < len(a.homeItems) && y < h-2; i++ {
			drawItem(s, 3, y, a.homeItems[i], i == a.selected, base, selected, w-5)
			y++
		}
	}
	putString(s, 1, h-1, "/ 搜索  r 全部进度  b 全部书签  ↑↓ 选择  Enter 打开  q 退出", muted, w-2)
	if a.message != "" {
		putString(s, 2, h-2, a.message, accent, w-4)
	}
}

func (a *App) drawSearch(s tcell.Screen) {
	w, h := s.Size()
	base, muted, accent, selected := styles()
	putString(s, 2, 1, "搜索："+string(a.query)+"_", accent.Bold(true), w-4)
	if len(a.query) == 0 {
		putString(s, 2, 3, "支持繁体、简体、拼音/首字母和经号", muted, w-4)
	}
	for i, e := range a.results {
		y := i + 3
		if y >= h-1 {
			break
		}
		drawItem(s, 2, y, navItem{title: e.Title, detail: e.ID + "  " + e.By}, i == a.selected, base, selected, w-4)
	}
	putString(s, 1, h-1, "输入搜索  ↑↓ 选择  Enter 打开  Esc 返回", muted, w-2)
}
func (a *App) drawList(s tcell.Screen, title, footer string) {
	w, h := s.Size()
	base, muted, accent, selected := styles()
	center(s, 1, title, accent.Bold(true), w)
	if len(a.listItems) == 0 {
		center(s, 4, "暂无记录", muted, w)
	}
	for i, it := range a.listItems {
		y := i + 3
		if y >= h-1 {
			break
		}
		drawItem(s, 2, y, it, i == a.selected, base, selected, w-4)
	}
	putString(s, 1, h-1, footer, muted, w-2)
}

func (a *App) drawReader(s tcell.Screen) {
	w, h := s.Size()
	base, muted, accent, _ := styles()
	red := base.Foreground(tcell.NewRGBColor(178, 52, 42))
	note := base.Foreground(tcell.ColorDarkCyan)
	putString(s, 1, 0, fmt.Sprintf("大藏經  %s  卷%s", a.doc.Title, a.doc.Juan), accent.Bold(true), w-2)
	rows := min(a.rows, max(1, h-3))
	for ci, col := range a.book.Page(a.start, a.visibleColumns(s)) {
		x := w - 3 - ci*3
		if x < 0 {
			break
		}
		for y, cell := range col.Cells {
			if y >= rows {
				break
			}
			cs := base
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
			if cell.Mark != model.NoMark {
				s.SetContent(markX(x), y+1, markRune(cell.Mark), nil, red)
			}
		}
	}
	pct := 100
	if len(a.book.Columns) > 1 {
		pct = a.start * 100 / (len(a.book.Columns) - 1)
	}
	putString(s, 1, h-1, fmt.Sprintf("←/Space 后翻  → 前翻  b 书签  [/] %d字  n 注:%s  Esc 首页  q 退出  %d%%", a.rows, onOff(a.showNotes), pct), muted, w-2)
	if a.message != "" {
		putString(s, 1, h-2, a.message, accent, w-2)
	}
}

func styles() (tcell.Style, tcell.Style, tcell.Style, tcell.Style) {
	base := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorDefault)
	return base, base.Foreground(tcell.ColorGray), base.Foreground(tcell.NewRGBColor(190, 130, 45)), base.Reverse(true)
}
func drawItem(s tcell.Screen, x, y int, item navItem, active bool, base, selected tcell.Style, width int) {
	st, prefix := base, "  "
	if active {
		st, prefix = selected, "› "
	}
	putString(s, x, y, prefix+item.title+"  "+item.detail, st, width)
}
func center(s tcell.Screen, y int, text string, style tcell.Style, width int) {
	x := max(0, (width-displayWidth(text))/2)
	putString(s, x, y, text, style, width-x)
}
func displayWidth(value string) int {
	n := 0
	for _, r := range value {
		if r >= 0x1100 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
func snippet(text string, at, radius int) string {
	r := []rune(text)
	at = max(0, min(len(r), at))
	start, end := max(0, at-radius), min(len(r), at+radius)
	out := string(r[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(r) {
		out += "…"
	}
	return out
}
func markX(glyphX int) int { return glyphX + 2 }
func markRune(mark model.Mark) rune {
	if mark == model.FullStop {
		return '.'
	}
	return ','
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
