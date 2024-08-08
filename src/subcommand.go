package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
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
			printCmdErr(fmt.Sprintf("Failed to create semantifly index log at %s: %s", indexLog, err))
		}
	}
	f, err := os.OpenFile(indexLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		printCmdErr(fmt.Sprintf("Failed to open index log at %s: %s", indexLog, err))
	}
	_, err = f.WriteString(strings.Join(os.Args, " ") + "\n")
	if err != nil {
		printCmdErr(fmt.Sprintf("Failed to append to index log at %s: %s", indexLog, err))
	}
	f.Close()

	switch os.Args[1] {
	case "add":
		cmd := flag.NewFlagSet("add", flag.ExitOnError)
		dataTypeStr := cmd.String("type", "text", "The type of the input data")
		sourceTypeStr := cmd.String("source-type", "file", "How to access the content")
		makeLocalCopy := cmd.Bool("copy", false, "Whether to copy and use the file as it is now, or dynamically access it")
		indexPath := cmd.String("index-path", "", "Path to the index file")
		cmd.Parse(os.Args[2:])

		if len(cmd.Args()) < 1 {
			printCmdErr("Add subcommand requires at least one input arg append to index log")
			return
		}

		dataType, err := parseDataType(*dataTypeStr)
		if err != nil {
			printCmdErr(fmt.Sprintf("Invalid data type: %v", err))
			return
		}

		sourceType, err := parseSourceType(*sourceTypeStr)
		if err != nil {
			printCmdErr(fmt.Sprintf("Invalid source type: %v", err))
			return
		}

		args := subcommands.AddArgs{
			IndexPath:  *indexPath,
			DataType:   dataType,
			SourceType: sourceType,
			MakeCopy:   *makeLocalCopy,
			DataURIs:   cmd.Args(),
		}

		subcommands.Add(args)

	default:
		printCmdErr("No valid subcommand provided.")
		os.Exit(1)
	}
}

func loadIndex(indexDir string) {
	fmt.Println("loadIndex in commands.go is unimplemented")
}

func parseDataType(str string) (pb.DataType, error) {
	switch str {
	case "text":
		return pb.DataType_TEXT, nil
	default:
		return pb.DataType_TEXT, fmt.Errorf("unknown data type: %s", str)
	}
}

func parseSourceType(str string) (pb.SourceType, error) {
	switch str {
	case "file":
		return pb.SourceType_LOCAL_FILE, nil
	default:
		return pb.SourceType_LOCAL_FILE, fmt.Errorf("unknown source type: %s", str)
	}
}
