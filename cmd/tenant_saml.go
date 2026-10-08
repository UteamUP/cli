package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/security"
)

var tenantSamlGUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

var tenantTestSamlCmd = &cobra.Command{
	Use: "test-saml <tenantGuid>",
	Short: "Test a disabled tenant SAML connection in the system browser",
	Long: `Validate a saved SAML connection using a restricted setup test.

Requires Tenant.ManageSaml plus tenant ownership or configured platform administration.
The current profile's owner session is retained. A successful test permits activation
through tenant update-saml with enabled=true in the configuration JSON file.

Example:
  uteamup tenant test-saml <tenantGuid> --profile dev`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || !tenantSamlGUID.MatchString(args[0]) || args[0] == "00000000-0000-0000-0000-000000000000" {
			return fmt.Errorf("test-saml requires one non-empty tenant GUID")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil { return fmt.Errorf("loading configuration: %w", err) }
		profile, selectedName, err := selectedProfileConfig(cfg, profileName)
		if err != nil { return err }
		origin, err := security.CanonicalOrigin(profile.BaseURL)
		if err != nil { return err }
		token, err := auth.LoadToken()
		if err != nil { return fmt.Errorf("reading owner session: %w", err) }
		if token == nil || !token.IsValid() { return fmt.Errorf("sign in as an authorized tenant owner before running the setup test") }
		if err := token.ValidateBinding(origin, selectedName); err != nil { return err }
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		fmt.Fprintln(cmd.OutOrStdout(), "Finish the restricted SAML setup test in your browser. Press Ctrl+C to cancel.")
		if err := auth.NewClient(origin, insecure, logging.New(logging.LevelInfo)).TestSAML(ctx, args[0], token.AccessToken); err != nil { return err }
		fmt.Fprintln(cmd.OutOrStdout(), "SAML setup validated. Your current session was retained; activate the connection when ready.")
		return nil
	},
}

func init() { tenantCmd.AddCommand(tenantTestSamlCmd) }
