package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/output"
)

// ExportConfig holds JSON export settings from the active profile.
type ExportConfig struct {
	Enabled bool
	Dir     string // defaults to ~/.uteamup/exports
}

// ArgDef defines a positional argument.
type ArgDef struct {
	Name        string
	Description string
	Required    bool
	Type        string // "string", "int", "uuid"
}

// FlagDef defines a named flag.
type FlagDef struct {
	Name        string
	Short       string
	Description string
	Default     any
	Type        string // "string", "int", "bool", "float"
	Required    bool
}

// HTTPMethod maps action names to HTTP methods for REST calls.
var HTTPMethod = map[string]string{
	"list":   "GET",
	"get":    "GET",
	"create": "POST",
	"update": "PUT",
	"delete": "DELETE",
	"search": "GET",
}

// Action represents a single CLI action (list, get, create, etc.).
type Action struct {
	Name        string
	Description string
	ToolName    string // MCP tool name, e.g. "UteamupAssetList"
	RESTPath    string // Optional REST override, e.g. "all" or "search"
	Args        []ArgDef
	Flags       []FlagDef
}

// Domain represents an entity domain with its available actions.
type Domain struct {
	Name        string
	Aliases     []string
	Description string
	APIPath     string // REST base path, e.g. "/api/vendor". Auto-derived if empty.
	Actions     []Action
}

// Registry holds all registered domains.
type Registry struct {
	domains []*Domain
}

// DefaultRegistry is the global domain registry.
var DefaultRegistry = &Registry{}

// Register adds a domain to the registry.
func (r *Registry) Register(d *Domain) {
	r.domains = append(r.domains, d)
}

// Register adds a domain to the default registry (package-level convenience).
func Register(d *Domain) {
	DefaultRegistry.Register(d)
}

// BuildCommands generates Cobra commands for all registered domains.
func (r *Registry) BuildCommands(apiClient *client.APIClient, logger *logging.Logger, outputFormat *string, export *ExportConfig) []*cobra.Command {
	var commands []*cobra.Command
	for _, domain := range r.domains {
		commands = append(commands, buildDomainCommand(domain, apiClient, logger, outputFormat, export))
	}
	return commands
}

func buildDomainCommand(domain *Domain, apiClient *client.APIClient, logger *logging.Logger, outputFormat *string, export *ExportConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     domain.Name,
		Aliases: domain.Aliases,
		Short:   domain.Description,
	}

	for _, action := range domain.Actions {
		cmd.AddCommand(buildActionCommand(domain, action, apiClient, logger, outputFormat, export))
	}

	return cmd
}

func buildActionCommand(domain *Domain, action Action, apiClient *client.APIClient, logger *logging.Logger, outputFormat *string, export *ExportConfig) *cobra.Command {
	// Build usage string with positional args
	use := action.Name
	for _, arg := range action.Args {
		if arg.Required {
			use += fmt.Sprintf(" <%s>", arg.Name)
		} else {
			use += fmt.Sprintf(" [%s]", arg.Name)
		}
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: action.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeAction(cmd, args, domain, action, apiClient, logger, outputFormat, export)
		},
	}

	// Register flags
	for _, flag := range action.Flags {
		switch flag.Type {
		case "int":
			def := 0
			if flag.Default != nil {
				def = flag.Default.(int)
			}
			if flag.Short != "" {
				cmd.Flags().IntP(flag.Name, flag.Short, def, flag.Description)
			} else {
				cmd.Flags().Int(flag.Name, def, flag.Description)
			}
		case "bool":
			def := false
			if flag.Default != nil {
				def = flag.Default.(bool)
			}
			if flag.Short != "" {
				cmd.Flags().BoolP(flag.Name, flag.Short, def, flag.Description)
			} else {
				cmd.Flags().Bool(flag.Name, def, flag.Description)
			}
		case "float":
			def := 0.0
			if flag.Default != nil {
				switch v := flag.Default.(type) {
				case float64:
					def = v
				case int:
					def = float64(v)
				}
			}
			cmd.Flags().Float64(flag.Name, def, flag.Description)
		default: // string
			def := ""
			if flag.Default != nil {
				def = flag.Default.(string)
			}
			if flag.Short != "" {
				cmd.Flags().StringP(flag.Name, flag.Short, def, flag.Description)
			} else {
				cmd.Flags().String(flag.Name, def, flag.Description)
			}
		}

		if flag.Required {
			_ = cmd.MarkFlagRequired(flag.Name)
		}
	}

	return cmd
}

func executeAction(cmd *cobra.Command, args []string, domain *Domain, action Action, apiClient *client.APIClient, logger *logging.Logger, outputFormat *string, export *ExportConfig) error {
	toolArgs := make(map[string]any)

	// Positional args
	for i, argDef := range action.Args {
		if i >= len(args) {
			if argDef.Required {
				return fmt.Errorf("missing required argument: %s", argDef.Name)
			}
			continue
		}
		toolArgs[argDef.Name] = convertArg(args[i], argDef.Type)
	}

	// Flags
	for _, flag := range action.Flags {
		if !cmd.Flags().Changed(flag.Name) {
			if flag.Default != nil {
				toolArgs[toCamelCase(flag.Name)] = flag.Default
			}
			continue
		}
		switch flag.Type {
		case "int":
			v, _ := cmd.Flags().GetInt(flag.Name)
			toolArgs[toCamelCase(flag.Name)] = v
		case "bool":
			v, _ := cmd.Flags().GetBool(flag.Name)
			toolArgs[toCamelCase(flag.Name)] = v
		case "float":
			v, _ := cmd.Flags().GetFloat64(flag.Name)
			toolArgs[toCamelCase(flag.Name)] = v
		default:
			v, _ := cmd.Flags().GetString(flag.Name)
			toolArgs[toCamelCase(flag.Name)] = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Build REST endpoint path from domain and action
	restPath := buildRESTPath(domain, action, toolArgs)
	httpMethod := HTTPMethod[action.Name]
	if httpMethod == "" {
		httpMethod = "GET"
	}

	logger.Debug("calling %s %s (tool: %s) with args %v", httpMethod, restPath, action.ToolName, toolArgs)

	result, err := apiClient.CallREST(ctx, httpMethod, restPath, toolArgs, action.Name)
	if err != nil {
		return err
	}

	// Export JSON to file if enabled
	if export != nil && export.Enabled && result != nil {
		if exportErr := exportJSON(export, domain.Name, action.Name, result, logger); exportErr != nil {
			logger.Warn("failed to export JSON: %v", exportErr)
		}
	}

	format := output.ParseFormat(*outputFormat)
	return output.Print(format, result)
}

// exportJSON writes the raw JSON response to a file in the export directory.
func exportJSON(export *ExportConfig, domainName, actionName string, data json.RawMessage, logger *logging.Logger) error {
	dir := export.Dir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, ".uteamup", "exports")
	}
	// Expand ~ if present
	if strings.HasPrefix(dir, "~/") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[2:])
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating export dir: %w", err)
	}

	// Pretty-print the JSON
	var pretty json.RawMessage
	indented, err := json.MarshalIndent(json.RawMessage(data), "", "  ")
	if err != nil {
		pretty = data
	} else {
		pretty = indented
	}

	filename := fmt.Sprintf("%s_%s.json", domainName, actionName)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, pretty, 0640); err != nil {
		return fmt.Errorf("writing export: %w", err)
	}

	logger.Info("exported JSON to %s", path)
	return nil
}

// buildRESTPath constructs the REST API path from domain + action.
func buildRESTPath(domain *Domain, action Action, args map[string]any) string {
	basePath := domain.APIPath
	if basePath == "" {
		// Derive from domain name: "vendor" → "/api/vendor", "asset-type" → "/api/assettype"
		basePath = "/api/" + strings.ReplaceAll(domain.Name, "-", "")
	}

	switch action.Name {
	case "get", "update", "delete":
		if id, ok := args["id"]; ok {
			return fmt.Sprintf("%s/%v", basePath, id)
		}
	case "search":
		if action.RESTPath != "" {
			return basePath + "/" + action.RESTPath
		}
		return basePath + "/search"
	case "list":
		if action.RESTPath != "" {
			return basePath + "/" + action.RESTPath
		}
	}

	return basePath
}

func convertArg(value, argType string) any {
	switch argType {
	case "int":
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
		return value
	default:
		return value
	}
}

// toCamelCase converts kebab-case to camelCase (page-size → pageSize).
func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// Domains returns all registered domains (for documentation generation).
func (r *Registry) Domains() []*Domain {
	return r.domains
}
