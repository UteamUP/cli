package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Display who is currently authenticated, the auth method used,
token expiry, and the associated config profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := auth.LoadToken()
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}

		if token == nil {
			fmt.Println("Not authenticated.")
			fmt.Println("Run \"uteamup login\" or \"ut login\" to authenticate.")
			return nil
		}

		fmt.Println("Authentication Status")
		fmt.Println("---------------------")

		if token.Email != "" {
			fmt.Printf("  User:        %s\n", token.Email)
		}
		fmt.Printf("  Method:      %s\n", token.AuthMethod)
		fmt.Printf("  Profile:     %s\n", token.Profile)
		if token.TenantName != "" {
			fmt.Printf("  Tenant:      %s\n", token.TenantName)
		}
		if token.TenantGuid != "" {
			fmt.Printf("  Tenant GUID: %s\n", token.TenantGuid)
		}
		fmt.Printf("  Expires:     %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))

		if token.IsValid() {
			remaining := time.Until(token.ExpiresAt).Round(time.Minute)
			fmt.Printf("  Status:      Valid (%s remaining)\n", remaining)
		} else {
			fmt.Printf("  Status:      Expired\n")
			fmt.Println("\nRun \"uteamup login\" or \"ut login\" to re-authenticate.")
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
