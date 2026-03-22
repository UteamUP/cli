package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for uteamup (and ut).

Bash:
  source <(uteamup completion bash)
  # or add to ~/.bashrc:
  echo 'source <(uteamup completion bash)' >> ~/.bashrc

Zsh:
  source <(uteamup completion zsh)
  # or add to ~/.zshrc:
  echo 'source <(uteamup completion zsh)' >> ~/.zshrc

Fish:
  uteamup completion fish | source
  # or save permanently:
  uteamup completion fish > ~/.config/fish/completions/uteamup.fish

PowerShell:
  uteamup completion powershell | Out-String | Invoke-Expression`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return cmd.Help()
		}
	},
}
