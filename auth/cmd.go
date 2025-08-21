package auth

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/utils"
	"github.com/spf13/cobra"
)

func SetupAuthCommand() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication helpers for GitHub",
	}

	authCmd.AddCommand(setupAuthCheckCommand())

	return authCmd
}

func setupAuthCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Verify GitHub authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := utils.GetGhUsernameGraphQL()
			if err != nil {
				fmt.Println("❌ Failed to fetch user info.")
				fmt.Println(err)

				return err
			}

			fmt.Printf("✅ Authenticated as GitHub user: %s\n", username)

			return nil
		},
	}
}
