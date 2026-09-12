package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/guyiicn/tripitaka-cli/library"
	"github.com/guyiicn/tripitaka-cli/model"
	"github.com/guyiicn/tripitaka-cli/state"
	"github.com/guyiicn/tripitaka-cli/tui"
)

func main() {
	check := flag.Bool("check", false, "validate the volume and print a summary without opening the TUI")
	dataDir := flag.String("data", "local-data", "prepared CBETA data directory")
	catalogPath := flag.String("catalog", "", "catalog.json path (default: <data>/catalog.json)")
	statePath := flag.String("state", "", "state database path (default: XDG data directory)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法: tripitaka-cli [选项] [prepared-juan.json]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	var lib library.Store
	var err error
	var startID, startJuan string
	if flag.NArg() == 1 {
		doc, loadErr := model.Load(flag.Arg(0))
		if loadErr != nil {
			err = loadErr
		} else if *check {
			fmt.Printf("%s %s 卷%s：%d 字，%d 分题，%d 夹注，%d 缺字\n",
				doc.ID, doc.Title, doc.Juan, len([]rune(doc.Text)), len(doc.Headings), len(doc.Notes), len(doc.Gaiji))
			return
		} else {
			lib = library.SingleDocument(doc)
			startID, startJuan = doc.ID, doc.Juan
		}
	} else {
		if *catalogPath == "" {
			*catalogPath = filepath.Join(*dataDir, "catalog.json")
		}
		lib, err = library.OpenDirectory(*dataDir, *catalogPath)
	}
	if err != nil {
		fail(err)
	}
	if *statePath == "" {
		*statePath, err = state.DefaultPath()
		if err != nil {
			fail(err)
		}
	}
	st, err := state.Open(*statePath)
	if err != nil {
		fail(err)
	}
	defer st.Close()
	app := tui.New(lib, st)
	if startID != "" {
		app.Start(startID, startJuan)
	}
	if err := app.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "tripitaka-cli:", err)
	os.Exit(1)
}
