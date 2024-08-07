package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "semantifly",
		Short: "Semantifly is a tool for implementing local RAG for GenAI-for-coding",
		Long:  `Semantifly allows users to add data sources, gather and index data, and implement Retrieval Augmented Generation with local data.`,
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add data or data sources",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Add command called")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}