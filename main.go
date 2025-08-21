package main

import (
	"fmt"
	"os"

	"github.com/astein-peddi/git-tooling/auth"
	"github.com/astein-peddi/git-tooling/completion"
	"github.com/astein-peddi/git-tooling/projects"
	"github.com/astein-peddi/git-tooling/prs"
	"github.com/spf13/cobra"
)

func main() {
	// utils.TestAuth()

	// completion.FetchAllBranches()

	rootCmd := &cobra.Command{
		Use:   "peddi-tooling",
		Short: "Peddi Tooling CLI",
		Long:  "Tooling for Peddinghaus Git Tasks",
	}

	rootCmd.AddCommand(completion.SetupAutoCompleteCommand())
	rootCmd.AddCommand(prs.SetupPrsCommand())
	rootCmd.AddCommand(auth.SetupAuthCommand())
	rootCmd.AddCommand(projects.SetupProjectsCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// go build -o peddi-tooling.exe ./main.go
