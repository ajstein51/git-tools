package main

import (
	"fmt"
	"os"

	"github.com/astein-peddi/git-tooling/prs"
	"github.com/spf13/cobra"
)
	
func main() {
       rootCmd := &cobra.Command{
	       Use:   "peddi-tooling",
	       Short: "Peddi Tooling CLI",
	       Long:  "Tooling for Peddi Git Tasks",
       }

       rootCmd.AddCommand(prs.SetupPrsCommand())

       if err := rootCmd.Execute(); err != nil {
	       fmt.Println(err)
	       os.Exit(1)
       }
}
