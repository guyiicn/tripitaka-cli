package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/guyiicn/tripitaka-cli/model"
	"github.com/guyiicn/tripitaka-cli/tui"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法: tripitaka-cli <prepared-juan.json>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	doc, err := model.Load(flag.Arg(0))
	if err == nil {
		err = tui.New(doc).Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tripitaka-cli:", err)
		os.Exit(1)
	}
}
