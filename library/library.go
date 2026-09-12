package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/guyiicn/tripitaka-cli/catalog"
	"github.com/guyiicn/tripitaka-cli/model"
)

type Store interface {
	Catalog() []catalog.Entry
	Load(id, juan string) (*model.Document, error)
	FirstJuan(id string) string
	Juans(id string) []string
}

type Directory struct {
	root    string
	entries []catalog.Entry
}

func OpenDirectory(root, catalogPath string) (*Directory, error) {
	entries, err := catalog.Load(catalogPath)
	if err != nil {
		return nil, err
	}
	return &Directory{root: root, entries: entries}, nil
}

func (d *Directory) Catalog() []catalog.Entry { return d.entries }

func (d *Directory) Load(id, juan string) (*model.Document, error) {
	return model.Load(filepath.Join(d.root, id, juan+".json"))
}

func (d *Directory) FirstJuan(id string) string {
	juans := d.Juans(id)
	if len(juans) > 0 {
		return juans[0]
	}
	return "001"
}

func (d *Directory) Juans(id string) []string {
	b, err := os.ReadFile(filepath.Join(d.root, id, "_meta.json"))
	if err == nil {
		var m struct {
			Juans []string `json:"juans"`
		}
		if json.Unmarshal(b, &m) == nil && len(m.Juans) > 0 {
			return m.Juans
		}
	}
	return nil
}

type Single struct {
	doc *model.Document
}

func SingleDocument(doc *model.Document) *Single { return &Single{doc: doc} }
func (s *Single) Catalog() []catalog.Entry {
	return []catalog.Entry{{ID: s.doc.ID, Title: s.doc.Title, By: s.doc.By, Juans: 1}}
}
func (s *Single) Load(id, juan string) (*model.Document, error) {
	if id != s.doc.ID {
		return nil, fmt.Errorf("volume %s is not available", id)
	}
	return s.doc, nil
}
func (s *Single) FirstJuan(id string) string { return s.doc.Juan }
func (s *Single) Juans(id string) []string   { return []string{s.doc.Juan} }
