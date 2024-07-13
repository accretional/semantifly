package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"accretional.com/semantifly/subcommands"
)

func printCmdErr(e string) {
	fmt.Println(fmt.Sprintf("%s\n Try --help to list subcommands and options.", e))
}

func CommandReadRun() {
	if len(os.Args) < 2 {
        printCmdErr("expected subcommand like 'add' or 'describe'")
        os.Exit(1)
    }
	semantifly_dir := flag.String("semantifly_dir", os.ExpandEnv("$HOME/.semantifly"), "Where to read existing semantifly data, and write new semantifly data.")
	_, err := os.ReadDir(*semantifly_dir)
	if err != nil {
        fmt.Println(fmt.Sprintf("No existing semantifly directory detected. Creating new semantifly directory at %s", *semantifly_dir))
		err := os.Mkdir(*semantifly_dir, 0777)
		if err != nil {
			printCmdErr(fmt.Sprintf("Failed to create semantifly directory at %s: %s", *semantifly_dir, err))
		}
    }
	semantifly_index := flag.String("semantifly_index", "default", "By default, semantifly writes data to the 'default' subdir of the configured semantifly_dir. Setting this to a value other than 'default' allows for multiple indices on the same local machine.")
	index_path := path.Join(*semantifly_dir, *semantifly_index)
	indexLog := path.Join(index_path, "index.log")
	_, err = os.ReadDir(index_path)
	if err != nil {
        fmt.Println(fmt.Sprintf("No existing semantifly index detected. Creating new semantifly index at %s", index_path))
		err := os.Mkdir(index_path, 0777)
		if err != nil {
			printCmdErr(fmt.Sprintf("Failed to create semantifly index directory at %s: %s", index_path, err))
		}
		_, err = os.Create(indexLog)
		if err != nil {
			printCmdErr(fmt.Sprintf("Failed to create semantifly index at %s: %s", indexLog, err))
		}
    }
	loadIndex(index_path)
	f, err := os.OpenFile(indexLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		printCmdErr(fmt.Sprintf("Failed to create open index log at %s: %s", indexLog, err))
	}
	_, err = f.WriteString(strings.Join(os.Args, " ")+"\n")
	if err != nil {
		printCmdErr(fmt.Sprintf("Failed to append to index log at %s: %s", indexLog, err))
	}
	f.Close() 

	switch os.Args[1] {
	case "add":
		cmd := flag.NewFlagSet("add", flag.ExitOnError)
		dataType := cmd.String("type", "text", "The type of the input data")
		cmd.Parse(os.Args[2:])
		subcommands.Add(subcommands.AddArgs{DataType: *dataType, DataURIs: cmd.Args()})
	default:
		printCmdErr("No valid subcommand provided.")
		os.Exit(1)
	}
}

func loadIndex(indexDir string) {
	fmt.Println("loadIndex in commands.go is unimplemented")
}