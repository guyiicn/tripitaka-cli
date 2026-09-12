package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/guyiicn/tripitaka-cli/model"
	"github.com/guyiicn/tripitaka-cli/tui"
)

func main() {
	check := flag.Bool("check", false, "validate the volume and print a summary without opening the TUI")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法: tripitaka-cli [选项] <prepared-juan.json>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	doc, err := model.Load(flag.Arg(0))
	if err == nil && *check {
		fmt.Printf("%s %s 卷%s：%d 字，%d 分题，%d 夹注，%d 缺字\n",
			doc.ID, doc.Title, doc.Juan, len([]rune(doc.Text)), len(doc.Headings), len(doc.Notes), len(doc.Gaiji))
		return
	}
	if err == nil {
		err = tui.New(doc).Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tripitaka-cli:", err)
		os.Exit(1)
	}
}
