package main

import (
	"fmt"
	"os"

	"github.com/astein-peddi/git-tooling/completion"
	"github.com/astein-peddi/git-tooling/prs"
	"github.com/spf13/cobra"
)
	
func main() {
	completion.FetchAllBranches()

	rootCmd := &cobra.Command{
		Use:   "peddi-tooling",
		Short: "Peddi Tooling CLI",
		Long:  "Tooling for Peddi Git Tasks",
	}

	rootCmd.AddCommand(prs.SetupPrsCommand())
	rootCmd.AddCommand(completion.SetupAutoCompleteCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// go build -o peddi-tooling.exe ./main.go