package state

import (
	"path/filepath"
	"testing"
)

func TestProgressAndBookmarks(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.SaveProgress(Progress{SutraID: "T0251", Juan: "001", Title: "心經", Char: 23}); err != nil {
		t.Fatal(err)
	}
	p, err := s.Progress("T0251", "001")
	if err != nil || p.Char != 23 {
		t.Fatalf("progress = %#v, %v", p, err)
	}
	latest, err := s.LatestProgress("T0251")
	if err != nil || latest.Juan != "001" || latest.Char != 23 {
		t.Fatalf("latest progress = %#v, %v", latest, err)
	}
	if err := s.AddBookmark(Bookmark{SutraID: "T0251", Juan: "001", Title: "心經", Char: 23, Snippet: "照見五蘊"}); err != nil {
		t.Fatal(err)
	}
	b, err := s.Bookmarks(10)
	if err != nil || len(b) != 1 || b[0].Snippet != "照見五蘊" {
		t.Fatalf("bookmarks = %#v, %v", b, err)
	}
}
