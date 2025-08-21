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

// debug build: go build -o peddi-tooling.exe ./main.go
// release build: go build -ldflags="-s -w" -o peddi-tooling.exe ./main.go
