package library

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDirectoryJuans(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "T2058")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := []byte(`{"id":"T2058","juans":["001","002","003","004","005","006"]}`)
	if err := os.WriteFile(filepath.Join(dir, "_meta.json"), meta, 0o644); err != nil {
		t.Fatal(err)
	}
	d := &Directory{root: root}
	want := []string{"001", "002", "003", "004", "005", "006"}
	if got := d.Juans("T2058"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Juans() = %v, want %v", got, want)
	}
}
