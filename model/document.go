package model

import (
	"encoding/json"
	"fmt"
	"os"
)

// Mark is a punctuation mark attached to the preceding source character.
type Mark int

const (
	NoMark Mark = iota
	FullStop
	Pause
)

// Heading is a structural title beginning at a source character index.
type Heading struct {
	At   int
	Text string
	Kind string
}

// Note is an interlinear annotation attached after a source character.
type Note struct {
	At   int
	Text string
}

// Gaiji describes a missing glyph and its CBETA composition formula.
type Gaiji struct {
	At      int
	Formula string
}

// Document is the compact JSON v2 format emitted by tripitaka/cbeta_prep.py.
type Document struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	By       string    `json:"by"`
	Juan     string    `json:"juan"`
	N        int       `json:"n"`
	Text     string    `json:"text"`
	Ju       []int     `json:"ju"`
	Dou      []int     `json:"dou"`
	Breaks   []int     `json:"br"`
	Ranges   [][2]int  `json:"xr"`
	Headings []Heading `json:"-"`
	Notes    []Note    `json:"-"`
	Gaiji    []Gaiji   `json:"-"`
	Version  int       `json:"v"`
	Fen      [][]any   `json:"fen"`
	NoteRaw  [][]any   `json:"note"`
	GaijiRaw [][]any   `json:"gx"`
}

// Load reads and validates one prepared CBETA volume.
func Load(path string) (*Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Document
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("decode CBETA JSON: %w", err)
	}
	if d.Text == "" {
		return nil, fmt.Errorf("CBETA JSON has no text")
	}
	if err := d.decodeTuples(); err != nil {
		return nil, err
	}
	return &d, nil
}

func (d *Document) decodeTuples() error {
	for _, row := range d.Fen {
		if len(row) < 2 {
			return fmt.Errorf("invalid fen tuple")
		}
		at, ok := number(row[0])
		text, ok2 := row[1].(string)
		if !ok || !ok2 {
			return fmt.Errorf("invalid fen tuple types")
		}
		kind := "fen"
		if len(row) > 2 {
			if v, ok := row[2].(string); ok {
				kind = v
			}
		}
		d.Headings = append(d.Headings, Heading{At: at, Text: text, Kind: kind})
	}
	for _, row := range d.NoteRaw {
		if len(row) != 2 {
			return fmt.Errorf("invalid note tuple")
		}
		at, ok := number(row[0])
		text, ok2 := row[1].(string)
		if !ok || !ok2 {
			return fmt.Errorf("invalid note tuple types")
		}
		d.Notes = append(d.Notes, Note{At: at, Text: text})
	}
	for _, row := range d.GaijiRaw {
		if len(row) != 2 {
			return fmt.Errorf("invalid gx tuple")
		}
		at, ok := number(row[0])
		formula, ok2 := row[1].(string)
		if !ok || !ok2 {
			return fmt.Errorf("invalid gx tuple types")
		}
		d.Gaiji = append(d.Gaiji, Gaiji{At: at, Formula: formula})
	}
	return nil
}

func number(v any) (int, bool) {
	f, ok := v.(float64)
	return int(f), ok && f >= -1 && f == float64(int(f))
}
