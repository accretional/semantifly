package subcommands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

	db "accretional.com/semantifly/database"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

var defaultIndexPath = path.Join(os.ExpandEnv("$HOME/.semantifly"), "default")

type SubcommandInfo struct {
	Description string
	Execute     func(context.Context, *db.PgxIface, []string)
}

var subcommandDict = map[string]SubcommandInfo{
	"add": {
		Description: "Add new data to the index",
		Execute:     executeAdd,
	},
	"delete": {
		Description: "Delete data from the index",
		Execute:     executeDelete,
	},
	"get": {
		Description: "Retrieve data from the index",
		Execute:     executeGet,
	},
	"update": {
		Description: "Update existing data in the index",
		Execute:     executeUpdate,
	},
	"search": {
		Description: "Search (lexically) for a term in the index",
		Execute:     executeSearch,
	},
	"start-server": {
		Description: "Start GRPC server",
		Execute:     executeStartServer,
	},
}

func printCmdErr(e string) {
	fmt.Printf("%s\n Try --help to list subcommands and options.\n", e)
}

func CommandReadRun() {
	if len(os.Args) < 2 {
		printCmdErr("expected subcommand like 'add' or 'describe'")
		os.Exit(1)
	}

	ctx, conn, err := setupDBConn()
	if err != nil {
		fmt.Printf("Failed to establish connection to the database: %v", err)
		return
	}

	var dbConn db.PgxIface = conn

	cmdName := os.Args[1]
	args := os.Args[2:]

	if cmdName == "--help" || cmdName == "-h" {
		baseHelp()
		return
	}

	if subcommand, exists := subcommandDict[cmdName]; exists {
		subcommand.Execute(ctx, &dbConn, args)
	} else {
		printCmdErr("No valid subcommand provided.")
		os.Exit(1)
	}
}

func baseHelp() {
	fmt.Println("Semantifly Help")
	fmt.Println("Available subcommands:")

	for cmd, info := range subcommandDict {
		fmt.Printf("  %-10s %s\n", cmd, info.Description)
	}

	fmt.Println("\nUse 'semantifly <subcommand> --help' for more information about a specific subcommand.")
}

func setupDBConn() (context.Context, db.PgxIface, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to PostgreSQL database: %v", err)
	}

	var dbConn db.PgxIface = conn

	err = db.InitializeDatabaseSchema(ctx, &dbConn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise the database schema: %v", err)
	}

	return ctx, conn, nil
}

func isBoolFlag(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	if f == nil {
		return false
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
				flags = append(flags, arg)
			} else {
				if strings.HasPrefix(arg, "--") {
					flagName = arg[2:]
				} else {
					flagName = arg[1:]
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

func inferSourceType(uris []string) (string, error) {
	sourceTypeStr := "local_file"

	for _, u := range uris {
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			sourceTypeStr = "webpage"
		} else if sourceTypeStr == "webpage" {
			return "", fmt.Errorf("inconsistent URI source types")
		}
	}

	return sourceTypeStr, nil
}

func executeAdd(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("add", flag.ExitOnError)
	dataType := cmd.String("type", "text", "The type of the input data")
	sourceType := cmd.String("source-type", "", "How to access the content")
	makeLocalCopy := cmd.Bool("copy", false, "Whether to copy and use the file as it is now, or dynamically access it")
	indexPath := cmd.String("index-path", defaultIndexPath, "Path to the index file")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	if len(nonFlags) < 1 {
		printCmdErr("Add subcommand requires at least one input arg append to index log")
		return
	}

	if *sourceType == "" {
		sourceTypeStr, err := inferSourceType(cmd.Args())
		if err != nil {
			printCmdErr(fmt.Sprintf("Failed to infer source type from URIs: %v\n", err))
		}
		*sourceType = sourceTypeStr
	}

	dataTypeEnum, err := parseDataType(*dataType)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error in parsing DataType: %v\n", err))
	}

	sourceTypeEnum, err := parseSourceType(*sourceType)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error in parsing SourceType: %v\n", err))
	}

	dataUri := cmd.Args()[0]

	addArgs := &pb.AddRequest{
		AddedMetadata: &pb.ContentMetadata{
			URI:        dataUri,
			DataType:   dataTypeEnum,
			SourceType: sourceTypeEnum,
		},
		MakeCopy: *makeLocalCopy,
	}

	err = SubcommandAdd(ctx, conn, addArgs, *indexPath, os.Stdout)
	if err != nil {
		fmt.Printf("Error occurred during add subcommand: %v", err)
		return
	}
}

func executeDelete(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteLocalCopy := cmd.Bool("copy", false, "Whether to delete the copy made")
	indexPath := cmd.String("index-path", defaultIndexPath, "Path to the index file")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	if len(nonFlags) < 1 {
		printCmdErr("Delete subcommand requires at least one input arg.")
		return
	}

	deleteArgs := &pb.DeleteRequest{
		DeleteCopy: *deleteLocalCopy,
		Names:      cmd.Args(),
	}

	err = SubcommandDelete(ctx, conn, deleteArgs, *indexPath, os.Stdout)
	if err != nil {
		fmt.Printf("Error occurred during delete subcommand: %v", err)
		return
	}
}

func executeGet(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("get", flag.ExitOnError)
	indexPath := cmd.String("index-path", defaultIndexPath, "Path to the index file")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	if len(nonFlags) != 1 {
		printCmdErr("Get subcommand requires exactly one arg.")
		return
	}

	getArgs := &pb.GetRequest{
		Name: cmd.Args()[0],
	}

	resp, _, err := SubcommandGet(ctx, conn, getArgs, *indexPath, os.Stdout)
	if err != nil {
		fmt.Printf("Error occurred during get subcommand: %v\n", err)
		return
	}

	fmt.Println(resp)
}

func executeUpdate(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("update", flag.ExitOnError)
	dataType := cmd.String("type", "text", "The type of the input data")
	sourceType := cmd.String("source-type", "", "How to access the content")
	makeLocalCopy := cmd.Bool("copy", false, "Whether to copy and use the file as it is now, or dynamically access it")
	indexPath := cmd.String("index-path", defaultIndexPath, "Path to the index file")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	if len(nonFlags) != 2 {
		printCmdErr("Update subcommand requires two input args - index name and updated URI")
		return
	}

	if *sourceType == "" {
		sourceTypeStr, err := inferSourceType(cmd.Args())
		if err != nil {
			printCmdErr(fmt.Sprintf("Failed to infer source type from URIs: %v\n", err))
		}
		*sourceType = sourceTypeStr
	}

	dataTypeEnum, err := parseDataType(*dataType)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error in parsing DataType: %v\n", err))
	}

	sourceTypeEnum, err := parseSourceType(*sourceType)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error in parsing SourceType: %v\n", err))
	}

	updateArgs := &pb.UpdateRequest{
		Name: cmd.Args()[0],
		UpdatedMetadata: &pb.ContentMetadata{
			URI:        cmd.Args()[1],
			DataType:   dataTypeEnum,
			SourceType: sourceTypeEnum,
		},
		UpdateCopy: *makeLocalCopy,
	}

	err = SubcommandUpdate(ctx, conn, updateArgs, *indexPath, os.Stdout)
	if err != nil {
		fmt.Printf("Error occurred during update subcommand: %v", err)
		return
	}
}

func executeSearch(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("search", flag.ExitOnError)
	indexPath := cmd.String("index-path", defaultIndexPath, "Path to the index file")
	topN := cmd.Int("n", 1, "Top n search results")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	if len(nonFlags) != 1 {
		printCmdErr("Search subcommand requires exactly one arg.")
		return
	}

	searchArgs := &pb.LexicalSearchRequest{
		SearchTerm: cmd.Args()[0],
		TopN:       int32(*topN),
	}

	results, err := SubcommandLexicalSearch(searchArgs, *indexPath, os.Stdout)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error during search: %v", err))
		return
	}

	PrintSearchResults(results, os.Stdout)
}

func executeStartServer(ctx context.Context, conn *db.PgxIface, args []string) {
	cmd := flag.NewFlagSet("start-server", flag.ExitOnError)
	serverIndexPath := cmd.String("index-path", defaultIndexPath, "Path to the server index")
	port := cmd.String("port", "50051", "Port for the server")

	flags, nonFlags, err := parseArgs(args, cmd)
	if err != nil {
		printCmdErr(fmt.Sprintf("Error: %v", err))
		return
	}

	reorderedArgs := append(flags, nonFlags...)
	cmd.Parse(reorderedArgs)

	portStr := ":" + *port
	lis, err := net.Listen("tcp", portStr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSemantiflyServer(s, SemantiflyNewServer(ctx, conn, *serverIndexPath))
	log.Printf("server listening at %v", lis.Addr())
	log.Printf("using index path: %v", *serverIndexPath)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
