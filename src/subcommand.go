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

func isFlag(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	return f != nil
}

func isBoolFlag(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	if f == nil {
		return false
	}

	if getter, ok := f.Value.(flag.Getter); ok {
		if _, ok := getter.Get().(bool); ok {
			return true
		}
	}

	if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok {
		return bf.IsBoolFlag()
	}

	return false
}

func parseArgs(args []string, cmd *flag.FlagSet) ([]string, []string, error) {
	var flags []string
	var nonFlags []string
	var flagName string
	i := 0

	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				if strings.HasPrefix(parts[0], "--") {
					flagName = parts[0][2:]
				} else {
					flagName = parts[0][1:]
				}

				if !isFlag(cmd, flagName) {
					return nil, nil, fmt.Errorf("unrecognized flag: %s", arg)
				}

				flags = append(flags, arg)
			} else {
				if strings.HasPrefix(arg, "--") {
					flagName = arg[2:]
				} else {
					flagName = arg[1:]
				}

				if !isFlag(cmd, flagName) {
					return nil, nil, fmt.Errorf("unrecognized flag: %s", arg)
				}

				flags = append(flags, arg)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !isBoolFlag(cmd, flagName) {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
		i++
	}

	return flags, nonFlags, nil
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

	cmdName := os.Args[1]
	args := os.Args[2:]

	switch cmdName {
	case "add":
		cmd := flag.NewFlagSet("add", flag.ExitOnError)
		dataTypeStr := cmd.String("type", "text", "The type of the input data")
		sourceTypeStr := cmd.String("source-type", "local_file", "How to access the content")
		makeLocalCopy := cmd.Bool("copy", false, "Whether to copy and use the file as it is now, or dynamically access it")
		indexPath := cmd.String("index-path", "", "Path to the index file")

		flags, nonFlags, err := parseArgs(args, cmd)
		if err != nil {
			printCmdErr(fmt.Sprintf("Error: %v", err))
			return
		}

		reorderedArgs := append(flags, nonFlags...)

		if len(nonFlags) < 1 {
			printCmdErr("Add subcommand requires at least one input arg append to index log")
			return
		}

		cmd.Parse(reorderedArgs)

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

	case "delete":
		cmd := flag.NewFlagSet("delete", flag.ExitOnError)
		deleteLocalCopy := cmd.Bool("copy", false, "Whether to delete the copy made")
		indexPath := cmd.String("index-path", "", "Path to the index file")

		flags, nonFlags, err := parseArgs(args, cmd)
		if err != nil {
			printCmdErr(fmt.Sprintf("Error: %v", err))
			return
		}

		reorderedArgs := append(flags, nonFlags...)

		if len(nonFlags) < 1 {
			printCmdErr("Delete subcommand requires at least one input arg.")
			return
		}

		cmd.Parse(reorderedArgs)

		args := subcommands.DeleteArgs{
			IndexPath:  *indexPath,
			DeleteCopy: *deleteLocalCopy,
			DataURIs:   cmd.Args(),
		}

		subcommands.Delete(args)

	case "get":
		cmd := flag.NewFlagSet("get", flag.ExitOnError)
		indexPath := cmd.String("index-path", "", "Path to the index file")

		flags, nonFlags, err := parseArgs(args, cmd)
		if err != nil {
			printCmdErr(fmt.Sprintf("Error: %v", err))
			return
		}

		reorderedArgs := append(flags, nonFlags...)

		if len(nonFlags) != 1 {
			printCmdErr("Get subcommand requires exactly one arg.")
			return
		}

		cmd.Parse(reorderedArgs)

		name := cmd.Args()[0]
		args := subcommands.GetArgs{
			IndexPath: *indexPath,
			Name:      name,
		}

		subcommands.Get(args)

	default:
		printCmdErr("No valid subcommand provided.")
		os.Exit(1)
	}
}

func loadIndex(indexDir string) {
	fmt.Println("loadIndex in commands.go is unimplemented")
}

func parseDataType(str string) (pb.DataType, error) {
	val, ok := pb.DataType_value[strings.ToUpper(str)]
	if !ok {
		return pb.DataType_TEXT, fmt.Errorf("unknown data type: %s", str)
	}
	return pb.DataType(val), nil
}

func parseSourceType(str string) (pb.SourceType, error) {
	val, ok := pb.SourceType_value[strings.ToUpper(str)]
	if !ok {
		return pb.SourceType_LOCAL_FILE, fmt.Errorf("unknown source type: %s", str)
	}
	return pb.SourceType(val), nil
}
