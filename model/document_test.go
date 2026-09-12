package model

import "testing"

func TestDecodeTuples(t *testing.T) {
	d := Document{
		Text:     "佛",
		Fen:      [][]any{{float64(0), "序品第一", "fen"}},
		NoteRaw:  [][]any{{float64(0), "佛陀"}},
		GaijiRaw: [][]any{{float64(0), "佛*心"}},
	}
	if err := d.decodeTuples(); err != nil {
		t.Fatal(err)
	}
	if d.Headings[0].Text != "序品第一" || d.Notes[0].Text != "佛陀" || d.Gaiji[0].Formula != "佛*心" {
		t.Fatalf("unexpected decoded document: %#v", d)
	}
}
