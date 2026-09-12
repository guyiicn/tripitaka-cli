package catalog

import "testing"

func TestSearchTraditionalSimplifiedPinyinAndID(t *testing.T) {
	entries := []Entry{{
		ID: "T0251", Title: "般若波羅蜜多心經", Simplified: "般若波罗蜜多心经",
		initials: "brblmdxj", pinyin: "banruoboluomiduoxinjing",
	}}
	for _, q := range []string{"心經", "心经", "brblmdxj", "banruo", "T0251", "251"} {
		if got := Search(entries, q, 10); len(got) != 1 || got[0].ID != "T0251" {
			t.Fatalf("query %q returned %#v", q, got)
		}
	}
}
