package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear authentication token",
	Long: `Remove the cached authentication token.

After logging out, you must run "uteamup login" or "ut login"
again before using any commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.ClearToken(); err != nil {
			return fmt.Errorf("clearing token: %w", err)
		}
		fmt.Println("Logged out successfully.")
		return nil
	},
}
