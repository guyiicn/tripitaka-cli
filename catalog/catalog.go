package catalog

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

type Entry struct {
	ID         string `json:"id"`
	Canon      string `json:"canon"`
	Title      string `json:"title"`
	Simplified string `json:"s"`
	By         string `json:"by"`
	Juans      int    `json:"juans"`

	initials string
	pinyin   string
}

func Load(path string) ([]Entry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, err
	}
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	args.Heteronym = false
	for i := range entries {
		parts := pinyin.LazyPinyin(entries[i].Simplified, args)
		entries[i].pinyin = strings.Join(parts, "")
		var initials strings.Builder
		for _, part := range parts {
			if part != "" {
				initials.WriteByte(part[0])
			}
		}
		entries[i].initials = initials.String()
	}
	return entries, nil
}

type scored struct {
	entry Entry
	score int
}

func Search(entries []Entry, query string, limit int) []Entry {
	q := normalize(query)
	if q == "" || limit <= 0 {
		return nil
	}
	var found []scored
	for _, e := range entries {
		id := normalize(e.ID)
		idNumber := strings.TrimLeft(strings.TrimLeft(id, "abcdefghijklmnopqrstuvwxyz"), "0")
		title := normalize(e.Title)
		simple := normalize(e.Simplified)
		by := normalize(e.By)
		score := 0
		switch {
		case q == id || q == idNumber:
			score = 1000
		case q == title || q == simple:
			score = 900
		case strings.HasPrefix(title, q) || strings.HasPrefix(simple, q):
			score = 800
		case strings.Contains(title, q) || strings.Contains(simple, q):
			score = 700
		case q == e.initials:
			score = 650
		case strings.HasPrefix(e.initials, q):
			score = 600
		case strings.HasPrefix(e.pinyin, q):
			score = 550
		case strings.Contains(by, q):
			score = 400
		default:
			continue
		}
		found = append(found, scored{entry: e, score: score})
	}
	sort.SliceStable(found, func(i, j int) bool {
		if found[i].score != found[j].score {
			return found[i].score > found[j].score
		}
		if len([]rune(found[i].entry.Title)) != len([]rune(found[j].entry.Title)) {
			return len([]rune(found[i].entry.Title)) < len([]rune(found[j].entry.Title))
		}
		return found[i].entry.ID < found[j].entry.ID
	})
	if len(found) > limit {
		found = found[:limit]
	}
	out := make([]Entry, len(found))
	for i := range found {
		out[i] = found[i].entry
	}
	return out
}

func normalize(s string) string {
	return strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			return -1
		}
		return r
	}, s))
}
