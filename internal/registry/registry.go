package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/output"
)

// Peer Sync is device-only (Nearby on the phone). There is no CLI domain and no MCP tool.

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
	Type        string // "string", "int", "uuid", "non-empty-uuid"
	// AllowedValues constrains a string positional argument to exact, stable
	// route-segment values. It prevents free-form values from changing the
	// intended REST path shape.
	AllowedValues []string
	// BodyName overrides the MCP/JSON argument name while preserving a
	// user-friendly positional name in CLI help and validation messages.
	BodyName string
	// QueryName routes a positional argument directly to the named query-string
	// field. This is useful when a GET endpoint's parameter name differs from
	// the CLI client's legacy search aliases.
	QueryName string
}

// FlagDef defines a named flag.
type FlagDef struct {
	Name        string
	Short       string
	Description string
	Default     any
	Type        string // "string", "int", "bool", "float", "stringSlice", "uuid", "non-empty-uuid"
	Required    bool
	// Sensitive keeps the value available to the outgoing request while
	// replacing it with [REDACTED] in diagnostic argument/header logs.
	Sensitive bool
	// BodyName overrides the JSON body field name when the flag value is sent
	// in a POST/PUT/PATCH body. Default is camelCase(Name). Use this when the
	// CLI flag name and the backend DTO field name diverge — e.g. CLI flag
	// `--text` mapping to backend field `bodyHtml` on a comment-create request.
	BodyName string
	// HeaderName routes the flag value to an HTTP request header instead of
	// the JSON body or query string. Used for cross-cutting headers such as
	// `Idempotency-Key` that the backend reads via `[FromHeader]`. When set,
	// the flag is excluded from the body and applied as a header on the
	// outgoing request.
	HeaderName string
	// QueryName routes the flag to an explicit query-string field even for
	// POST/PUT/PATCH requests. Use it for concurrency tokens bound with
	// [FromQuery], while keeping the reviewed mutation model in the JSON body.
	QueryName string
	// MirrorHeaderInBody also writes the same header value into BodyName
	// (default camelCase(Name)). This is limited to compatibility contracts
	// where the backend requires an Idempotency-Key header and temporarily
	// accepts a legacy body key only when both values are equal.
	MirrorHeaderInBody bool
	// JSONFile marks a string flag whose value is a local path to a JSON file.
	// The file is read and parsed at execution time and the parsed value is
	// sent under BodyName (default camelCase(Name)) in the request body — the
	// only way to express array/object payloads that flat flags cannot carry
	// (e.g. bulk stock operations, purchase-order receive lines).
	JSONFile bool
	// RootJSONObjectFile marks a string flag whose local JSON object is merged
	// directly into the request-body root. It never creates a wrapper field,
	// rejects arrays/scalars/duplicate keys, and fails on field collisions.
	RootJSONObjectFile bool
	// UploadFile marks a string flag whose value is a local path to a file
	// sent as a multipart/form-data part named by BodyName (default
	// camelCase(Name)). Remaining flags travel on the query string because the
	// multipart payload owns the body. Used by endpoints binding IFormFile
	// (e.g. stock CSV import).
	UploadFile bool
	// MustBeTrue requires an explicitly enabled boolean safety confirmation.
	// It is stronger than Required, which only proves that the flag was passed.
	MustBeTrue bool
	// LocalOnly keeps a parsed flag entirely inside the CLI. It is used for
	// explicit safety confirmations that must never become API request fields.
	LocalOnly bool
	// StrongETag accepts the opaque token returned by an API response and sends
	// it as one validated, quoted strong ETag header value.
	StrongETag bool
}

var uuidValuePattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

const emptyUUIDValue = "00000000-0000-0000-0000-000000000000"

// HTTPMethod maps action names to HTTP methods for REST calls.
//
// Convention: any verb prefixed with `update-` (e.g. update-status, update-notes,
// update-priority) is treated as a sub-route PATCH targeting `{basePath}/{id}/{suffix}`.
// The lookup in runCommand falls back to that rule when the action name isn't
// explicitly listed below; buildRESTPath does the matching path construction.
var HTTPMethod = map[string]string{
	"list":          "GET",
	"get":           "GET",
	"create":        "POST",
	"update":        "PUT",
	"update-status": "PATCH",
	"update-notes":  "PATCH",
	"delete":        "DELETE",
	"search":        "GET",
}

// Action represents a single CLI action (list, get, create, etc.).
type Action struct {
	Name        string
	Description string
	ToolName    string // MCP tool name, e.g. "UteamupAssetList"
	// MCPOnly routes the action through the authenticated JSON-RPC tools/call
	// transport. Use it only when the backend intentionally exposes no REST
	// adapter for the governed tool.
	MCPOnly bool
	// RESTBasePath overrides the domain API path for a single action. Use it
	// when one domain action is intentionally served by a cross-domain adapter.
	RESTBasePath string
	// UseDomainBasePath routes this action to the exact domain API path without
	// the conventional action-derived suffix (for example, a collection search
	// served at the collection root). It cannot be combined with path overrides.
	UseDomainBasePath bool
	// RESTPath is the path suffix appended to the domain's basePath. It supports
	// `{argName}` placeholders that are substituted from the action's positional
	// args. Examples:
	//   "all"                                  → static suffix (legacy use, e.g. list/search)
	//   "{bugExternalGuid}/comments"           → sub-resource list / create
	//   "{bugExternalGuid}/comments/{commentExternalGuid}" → sub-sub-resource get/edit/delete
	// Args consumed by placeholder substitution are removed from the JSON body
	// before the request is sent.
	RESTPath string
	// HTTPMethod overrides the action-name-based HTTP method default. Use this
	// for sub-resource verbs whose name doesn't fit the list/get/create/update/
	// delete convention — e.g. `comments-add` (POST), `attachments-delete`
	// (DELETE). Empty = derived from Action.Name via the HTTPMethod map.
	HTTPMethod string
	Args       []ArgDef
	Flags      []FlagDef
	// DownloadURLField turns the REST response into a streamed local download.
	// DownloadOutputFlag names the local-only output flag, while
	// DownloadDefaultArg supplies the default filename stem.
	DownloadURLField   string
	DownloadOutputFlag string
	DownloadDefaultArg string
	// DisableResponseExport prevents the profile-level JSON export feature
	// from persisting secret-bearing responses such as transfer challenges.
	DisableResponseExport bool
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

// APIClientFactory creates the API client when an action executes. Cobra parses
// persistent flags after commands are registered, so constructing the client
// during package initialization would ignore runtime flags such as --insecure
// and --profile.
type APIClientFactory func() (*client.APIClient, error)

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
func (r *Registry) BuildCommands(apiClientFactory APIClientFactory, logger *logging.Logger, outputFormat *string, export *ExportConfig) []*cobra.Command {
	var commands []*cobra.Command
	for _, domain := range r.domains {
		commands = append(commands, buildDomainCommand(domain, apiClientFactory, logger, outputFormat, export))
	}
	return commands
}

func buildDomainCommand(domain *Domain, apiClientFactory APIClientFactory, logger *logging.Logger, outputFormat *string, export *ExportConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     domain.Name,
		Aliases: domain.Aliases,
		Short:   domain.Description,
	}

	for _, action := range domain.Actions {
		cmd.AddCommand(buildActionCommand(domain, action, apiClientFactory, logger, outputFormat, export))
	}

	return cmd
}

func buildActionCommand(domain *Domain, action Action, apiClientFactory APIClientFactory, logger *logging.Logger, outputFormat *string, export *ExportConfig) *cobra.Command {
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
			if err := validateActionInput(cmd, args, action); err != nil {
				return err
			}
			apiClient, err := apiClientFactory()
			if err != nil {
				return err
			}
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
		case "stringSlice":
			var def []string
			if flag.Default != nil {
				if v, ok := flag.Default.([]string); ok {
					def = v
				}
			}
			if flag.Short != "" {
				cmd.Flags().StringSliceP(flag.Name, flag.Short, def, flag.Description)
			} else {
				cmd.Flags().StringSlice(flag.Name, def, flag.Description)
			}
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

func validateActionInput(cmd *cobra.Command, args []string, action Action) error {
	if err := validateActionDefinition(action); err != nil {
		return err
	}

	for index, argument := range action.Args {
		if index >= len(args) {
			continue
		}
		value := strings.TrimSpace(args[index])
		if argument.Type == "uuid" || argument.Type == "non-empty-uuid" {
			if !uuidValuePattern.MatchString(value) {
				return fmt.Errorf("invalid %s: must be a GUID", argument.Name)
			}
			if argument.Type == "non-empty-uuid" && strings.EqualFold(value, emptyUUIDValue) {
				return fmt.Errorf("invalid %s: must be a non-empty GUID", argument.Name)
			}
		}
		if len(argument.AllowedValues) > 0 && !containsExact(argument.AllowedValues, value) {
			return fmt.Errorf(
				"invalid %s: must be one of %s",
				argument.Name,
				strings.Join(argument.AllowedValues, ", "),
			)
		}
	}

	for _, flag := range action.Flags {
		if !cmd.Flags().Changed(flag.Name) {
			continue
		}
		if flag.Type == "uuid" || flag.Type == "non-empty-uuid" {
			value, _ := cmd.Flags().GetString(flag.Name)
			value = strings.TrimSpace(value)
			if !uuidValuePattern.MatchString(value) {
				return fmt.Errorf("invalid --%s: must be a GUID", flag.Name)
			}
			if flag.Type == "non-empty-uuid" && strings.EqualFold(value, emptyUUIDValue) {
				return fmt.Errorf("invalid --%s: must be a non-empty GUID", flag.Name)
			}
		}
		if flag.Required && (flag.Type == "string" || flag.Type == "uuid" || flag.Type == "non-empty-uuid") {
			value, _ := cmd.Flags().GetString(flag.Name)
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("--%s cannot be empty", flag.Name)
			}
		}
		if flag.MustBeTrue {
			value, _ := cmd.Flags().GetBool(flag.Name)
			if !value {
				return fmt.Errorf("--%s must be explicitly enabled", flag.Name)
			}
		}
		if flag.StrongETag {
			value, _ := cmd.Flags().GetString(flag.Name)
			if _, err := formatStrongETag(value); err != nil {
				return fmt.Errorf("invalid --%s: %w", flag.Name, err)
			}
		}
	}

	return nil
}

func validateActionDefinition(action Action) error {
	if action.UseDomainBasePath && (action.RESTBasePath != "" || action.RESTPath != "") {
		return fmt.Errorf(
			"invalid action %s: domain-base routing cannot use REST path overrides",
			action.Name,
		)
	}

	rootObjectFiles := 0
	for _, flag := range action.Flags {
		if flag.RootJSONObjectFile {
			rootObjectFiles++
			if flag.Type != "string" || flag.BodyName != "" || flag.HeaderName != "" ||
				flag.QueryName != "" || flag.JSONFile || flag.UploadFile || flag.LocalOnly ||
				flag.MirrorHeaderInBody || flag.MustBeTrue || flag.StrongETag {
				return fmt.Errorf(
					"invalid action %s: --%s must be an unwrapped root-object JSON file",
					action.Name,
					flag.Name,
				)
			}
		}
		if flag.LocalOnly && (flag.Type != "bool" || flag.BodyName != "" ||
			flag.HeaderName != "" || flag.QueryName != "" || flag.JSONFile ||
			flag.RootJSONObjectFile || flag.UploadFile || flag.MirrorHeaderInBody ||
			flag.StrongETag) {
			return fmt.Errorf(
				"invalid action %s: --%s must be a local-only boolean",
				action.Name,
				flag.Name,
			)
		}
	}
	if rootObjectFiles > 1 {
		return fmt.Errorf("invalid action %s: only one root-object JSON file is allowed", action.Name)
	}

	for _, argument := range action.Args {
		if len(argument.AllowedValues) == 0 {
			continue
		}
		if argument.Type != "string" {
			return fmt.Errorf(
				"invalid action %s: %s allowed values require a string argument",
				action.Name,
				argument.Name,
			)
		}
		seen := make(map[string]struct{}, len(argument.AllowedValues))
		for _, value := range argument.AllowedValues {
			if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "/\\?#") {
				return fmt.Errorf(
					"invalid action %s: %s contains an unsafe allowed route value",
					action.Name,
					argument.Name,
				)
			}
			if _, exists := seen[value]; exists {
				return fmt.Errorf(
					"invalid action %s: %s contains duplicate allowed values",
					action.Name,
					argument.Name,
				)
			}
			seen[value] = struct{}{}
		}
	}

	return nil
}

func containsExact(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func executeAction(cmd *cobra.Command, args []string, domain *Domain, action Action, apiClient *client.APIClient, logger *logging.Logger, outputFormat *string, export *ExportConfig) error {
	toolArgs := make(map[string]any)
	queryParams := make(map[string]any)
	downloadOutputPath := ""
	downloadDefaultBase := ""

	// Positional args
	for i, argDef := range action.Args {
		if i >= len(args) {
			if argDef.Required {
				return fmt.Errorf("missing required argument: %s", argDef.Name)
			}
			continue
		}
		value := convertArg(args[i], argDef.Type)
		if argDef.Name == action.DownloadDefaultArg {
			downloadDefaultBase = fmt.Sprint(value)
		}
		if argDef.QueryName != "" {
			queryParams[argDef.QueryName] = value
			continue
		}
		bodyName := argDef.Name
		if argDef.BodyName != "" {
			bodyName = argDef.BodyName
		}
		toolArgs[bodyName] = value
	}

	// Flags. HeaderName routes a flag to an HTTP header instead of the body /
	// query — kept separate so it never leaks back into toolArgs (which would
	// re-introduce the `Idempotency-Key`-as-body-field bug fixed by adding
	// HeaderName in the first place).
	headers := make(map[string]string)
	var uploadField, uploadPath string
	for _, flag := range action.Flags {
		if flag.Name == action.DownloadOutputFlag {
			if cmd.Flags().Changed(flag.Name) {
				downloadOutputPath, _ = cmd.Flags().GetString(flag.Name)
			}
			continue
		}
		if flag.LocalOnly {
			continue
		}
		if flag.QueryName != "" {
			if !cmd.Flags().Changed(flag.Name) {
				if flag.Default != nil {
					queryParams[flag.QueryName] = flag.Default
				}
				continue
			}
			switch flag.Type {
			case "int":
				v, _ := cmd.Flags().GetInt(flag.Name)
				queryParams[flag.QueryName] = v
			case "bool":
				v, _ := cmd.Flags().GetBool(flag.Name)
				queryParams[flag.QueryName] = v
			case "float":
				v, _ := cmd.Flags().GetFloat64(flag.Name)
				queryParams[flag.QueryName] = v
			case "stringSlice":
				v, _ := cmd.Flags().GetStringSlice(flag.Name)
				queryParams[flag.QueryName] = v
			default:
				v, _ := cmd.Flags().GetString(flag.Name)
				queryParams[flag.QueryName] = v
			}
			continue
		}
		if flag.HeaderName != "" {
			var headerValue string
			if cmd.Flags().Changed(flag.Name) {
				v, _ := cmd.Flags().GetString(flag.Name)
				if v != "" {
					if flag.StrongETag {
						formatted, formatErr := formatStrongETag(v)
						if formatErr != nil {
							return fmt.Errorf("invalid --%s: %w", flag.Name, formatErr)
						}
						v = formatted
					}
					headers[flag.HeaderName] = v
					headerValue = v
				}
			} else if flag.Default != nil {
				if dv, ok := flag.Default.(string); ok && dv != "" {
					if flag.StrongETag {
						formatted, formatErr := formatStrongETag(dv)
						if formatErr != nil {
							return fmt.Errorf("invalid default --%s: %w", flag.Name, formatErr)
						}
						dv = formatted
					}
					headers[flag.HeaderName] = dv
					headerValue = dv
				}
			}
			if flag.MirrorHeaderInBody && headerValue != "" {
				fieldName := flag.BodyName
				if fieldName == "" {
					fieldName = toCamelCase(flag.Name)
				}
				toolArgs[fieldName] = headerValue
			}
			continue
		}
		fieldName := flag.BodyName
		if fieldName == "" {
			fieldName = toCamelCase(flag.Name)
		}
		if flag.RootJSONObjectFile {
			if cmd.Flags().Changed(flag.Name) {
				path, _ := cmd.Flags().GetString(flag.Name)
				parsed, err := readRootJSONObjectFile(path)
				if err != nil {
					return fmt.Errorf("reading --%s: %w", flag.Name, err)
				}
				if err := mergeRootJSONObject(toolArgs, parsed); err != nil {
					return fmt.Errorf("reading --%s: %w", flag.Name, err)
				}
			}
			continue
		}
		if flag.JSONFile {
			if cmd.Flags().Changed(flag.Name) {
				v, _ := cmd.Flags().GetString(flag.Name)
				parsed, err := readJSONFileFlag(v)
				if err != nil {
					return fmt.Errorf("reading --%s: %w", flag.Name, err)
				}
				toolArgs[fieldName] = parsed
			}
			continue
		}
		if flag.UploadFile {
			if cmd.Flags().Changed(flag.Name) {
				uploadField = fieldName
				uploadPath, _ = cmd.Flags().GetString(flag.Name)
			}
			continue
		}
		if !cmd.Flags().Changed(flag.Name) {
			if flag.Default != nil {
				toolArgs[fieldName] = flag.Default
			}
			continue
		}
		switch flag.Type {
		case "int":
			v, _ := cmd.Flags().GetInt(flag.Name)
			toolArgs[fieldName] = v
		case "bool":
			v, _ := cmd.Flags().GetBool(flag.Name)
			toolArgs[fieldName] = v
		case "float":
			v, _ := cmd.Flags().GetFloat64(flag.Name)
			toolArgs[fieldName] = v
		case "stringSlice":
			v, _ := cmd.Flags().GetStringSlice(flag.Name)
			toolArgs[fieldName] = v
		default:
			v, _ := cmd.Flags().GetString(flag.Name)
			toolArgs[fieldName] = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var result json.RawMessage
	var err error
	if action.MCPOnly {
		if len(headers) > 0 || len(queryParams) > 0 || uploadPath != "" {
			return fmt.Errorf("MCP-only action %s cannot use REST headers, query flags, or uploads", action.Name)
		}
		logger.Debug("calling MCP tool %s with args %v", action.ToolName, loggedArgsForAction(action, toolArgs))
		result, err = apiClient.CallTool(ctx, action.ToolName, toolArgs)
		if err != nil {
			return err
		}
	} else {
		// Build REST endpoint path from domain and action. Path-template placeholders
		// consumed during substitution are stripped from the body so they don't double-
		// leak as JSON fields on POST/PUT/PATCH.
		restPath, consumed := buildRESTPath(domain, action, toolArgs)
		restPath = appendQueryParameters(restPath, queryParams)
		for _, name := range consumed {
			delete(toolArgs, name)
		}

		// Action.HTTPMethod wins over the action-name-based default. The static map
		// covers the standard CRUD verbs; the `update-<sub>` rule is the fallback.
		httpMethod := action.HTTPMethod
		if httpMethod == "" {
			httpMethod = HTTPMethod[action.Name]
		}
		if httpMethod == "" {
			if strings.HasPrefix(action.Name, "update-") {
				httpMethod = "PATCH"
			} else {
				httpMethod = "GET"
			}
		}

		loggedArgs, loggedHeaders := redactSensitiveActionValues(action, toolArgs, headers)
		logger.Debug("calling %s %s (tool: %s) with args %v headers %v", httpMethod, restPath, action.ToolName, loggedArgs, loggedHeaders)

		if uploadPath != "" {
			result, err = apiClient.CallRESTUpload(ctx, httpMethod, restPath, uploadField, uploadPath, toolArgs, headers, action.Name)
		} else {
			result, err = apiClient.CallREST(ctx, httpMethod, restPath, toolArgs, headers, action.Name)
		}
		if err != nil {
			return err
		}
	}

	if action.DownloadURLField != "" {
		var payload map[string]any
		if err := json.Unmarshal(result, &payload); err != nil {
			return fmt.Errorf("reading download response: %w", err)
		}
		downloadURL, ok := payload[action.DownloadURLField].(string)
		if !ok || strings.TrimSpace(downloadURL) == "" {
			return fmt.Errorf("download response is missing %s", action.DownloadURLField)
		}
		if downloadOutputPath == "" {
			downloadOutputPath = defaultDownloadPath(downloadDefaultBase, downloadURL)
		}
		written, err := apiClient.DownloadFile(ctx, downloadURL, downloadOutputPath)
		if err != nil {
			return err
		}
		result, err = json.Marshal(map[string]any{"path": downloadOutputPath, "bytes": written})
		if err != nil {
			return fmt.Errorf("formatting download result: %w", err)
		}
	}

	// Export JSON to file if enabled
	if export != nil && export.Enabled && result != nil && !action.DisableResponseExport {
		if exportErr := exportJSON(export, domain.Name, action.Name, result, logger); exportErr != nil {
			logger.Warn("failed to export JSON: %v", exportErr)
		}
	}

	format := output.ParseFormat(*outputFormat)
	return output.Print(format, result)
}

func formatStrongETag(value string) (string, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return "", fmt.Errorf("must be one non-empty opaque token without surrounding whitespace")
	}
	if value == "*" || strings.HasPrefix(strings.ToUpper(value), "W/") ||
		strings.Contains(value, ",") {
		return "", fmt.Errorf("must not be weak, wildcard, or contain multiple ETags")
	}

	opaque := value
	if strings.HasPrefix(value, "\"") || strings.HasSuffix(value, "\"") {
		if len(value) < 3 || !strings.HasPrefix(value, "\"") ||
			!strings.HasSuffix(value, "\"") {
			return "", fmt.Errorf("has mismatched ETag quotes")
		}
		opaque = value[1 : len(value)-1]
	}
	if opaque == "" || len(opaque) > 128 || strings.ContainsAny(opaque, "\",") ||
		strings.IndexFunc(opaque, func(character rune) bool {
			return character < 0x20 || character == 0x7f
		}) >= 0 {
		return "", fmt.Errorf("contains an invalid opaque ETag token")
	}
	return "\"" + opaque + "\"", nil
}

func defaultDownloadPath(base, rawURL string) string {
	extension := ".bin"
	if parsed, err := url.Parse(rawURL); err == nil {
		if candidate := filepath.Ext(parsed.Path); candidate != "" && len(candidate) <= 10 {
			extension = candidate
		}
	}
	if strings.TrimSpace(base) == "" {
		base = "uteamup-download"
	}
	return base + extension
}

func loggedArgsForAction(action Action, args map[string]any) map[string]any {
	logged, _ := redactSensitiveActionValues(action, args, nil)
	return logged
}

func appendQueryParameters(path string, params map[string]any) string {
	if len(params) == 0 {
		return path
	}
	values := url.Values{}
	for key, value := range params {
		switch typed := value.(type) {
		case []string:
			for _, item := range typed {
				values.Add(key, item)
			}
		default:
			values.Set(key, fmt.Sprint(value))
		}
	}
	encoded := values.Encode()
	if encoded == "" {
		return path
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + encoded
}

// redactSensitiveActionValues returns diagnostic-only copies with values from
// Sensitive flags removed. The original request maps are never mutated.
func redactSensitiveActionValues(action Action, args map[string]any, headers map[string]string) (map[string]any, map[string]string) {
	redactedArgs := make(map[string]any, len(args))
	for key, value := range args {
		redactedArgs[key] = value
	}
	redactedHeaders := make(map[string]string, len(headers))
	for key, value := range headers {
		redactedHeaders[key] = value
	}

	for _, flag := range action.Flags {
		if !flag.Sensitive {
			continue
		}
		if flag.HeaderName != "" {
			if _, exists := redactedHeaders[flag.HeaderName]; exists {
				redactedHeaders[flag.HeaderName] = "[REDACTED]"
			}
			continue
		}

		fieldName := flag.BodyName
		if fieldName == "" {
			fieldName = toCamelCase(flag.Name)
		}
		if _, exists := redactedArgs[fieldName]; exists {
			redactedArgs[fieldName] = "[REDACTED]"
		}
	}

	return redactedArgs, redactedHeaders
}

// readJSONFileFlag loads a JSONFile-marked flag: the file at path is read and
// parsed as JSON; the parsed value (array or object) becomes the body field.
func readJSONFileFlag(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing %s as JSON: %w", filepath.Base(path), err)
	}
	return parsed, nil
}

var rejectedRootObjectWrappers = map[string]struct{}{
	"args":      {},
	"arguments": {},
	"body":      {},
	"data":      {},
	"input":     {},
	"model":     {},
	"payload":   {},
	"request":   {},
}

// readRootJSONObjectFile loads one strict JSON object for direct request-root
// merging. It rejects wrapper-shaped payloads because the caller cannot know
// or choose an API binder wrapper for this primitive.
func readRootJSONObjectFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	parsed, err := decodeStrictJSONValue(decoder)
	if err != nil {
		return nil, fmt.Errorf("parsing %s as JSON: %w", filepath.Base(path), err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("parsing %s as JSON: multiple root values", filepath.Base(path))
		}
		return nil, fmt.Errorf("parsing %s as JSON: %w", filepath.Base(path), err)
	}

	object, ok := parsed.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must contain one JSON object at the root", filepath.Base(path))
	}
	if len(object) == 1 {
		for key := range object {
			if _, rejected := rejectedRootObjectWrappers[strings.ToLower(key)]; rejected {
				return nil, fmt.Errorf(
					"%s must contain request fields directly, without a %q wrapper",
					filepath.Base(path),
					key,
				)
			}
		}
	}
	return object, nil
}

func decodeStrictJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return token, nil
	}

	switch delimiter {
	case '{':
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("object field name is not a string")
			}
			if _, exists := object[key]; exists {
				return nil, fmt.Errorf("duplicate JSON object field %q", key)
			}
			value, err := decodeStrictJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		closing, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if closing != json.Delim('}') {
			return nil, fmt.Errorf("JSON object is not closed")
		}
		return object, nil
	case '[':
		array := make([]any, 0)
		for decoder.More() {
			value, err := decodeStrictJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		closing, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if closing != json.Delim(']') {
			return nil, fmt.Errorf("JSON array is not closed")
		}
		return array, nil
	default:
		return nil, fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}

func mergeRootJSONObject(destination, source map[string]any) error {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := destination[key]; exists {
			return fmt.Errorf("request field %q is supplied more than once", key)
		}
	}
	for _, key := range keys {
		destination[key] = source[key]
	}
	return nil
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
// Returns the path and the list of arg names consumed by path-template
// substitution (those must be removed from the JSON body before the request
// is sent so the same arg doesn't appear in both URL and body).
func buildRESTPath(domain *Domain, action Action, args map[string]any) (string, []string) {
	domainBasePath := domain.APIPath
	if domainBasePath == "" {
		domainBasePath = "/api/" + strings.ReplaceAll(domain.Name, "-", "")
	}
	if action.UseDomainBasePath {
		return domainBasePath, nil
	}

	basePath := action.RESTBasePath
	if basePath == "" {
		basePath = domainBasePath
	}

	// If the action declares an explicit RESTPath, expand `{argName}` placeholders
	// against the args map. A path with no placeholders is treated as a literal
	// suffix (preserves the legacy `RESTPath: "all"` / `"search"` usage).
	if action.RESTPath != "" {
		expanded, consumed := expandPathTemplate(action.RESTPath, args)
		return basePath + "/" + expanded, consumed
	}

	// Resolve the positional identifier to use in the URL path. Most legacy
	// domains pass an integer `id`; GUID-first domains (bugsandfeatures, etc.)
	// declare their required arg as `externalGuid` per the GUID-first rule.
	// Accept both so every domain routes correctly without needing RESTPath
	// overrides or API-key handlers that special-case each verb.
	idValue, hasID := args["id"]
	idArgName := "id"
	if !hasID {
		idValue, hasID = args["externalGuid"]
		idArgName = "externalGuid"
	}

	switch action.Name {
	case "get", "update", "delete":
		if hasID {
			return fmt.Sprintf("%s/%v", basePath, idValue), []string{idArgName}
		}
	case "update-status":
		// PATCH /api/<domain>/{id}/status is the convention established by
		// BugsAndFeaturesController.UpdateStatus — status-only transitions
		// get their own sub-route so they can't be conflated with a full
		// update (PUT /<id>). Domains that reuse this verb must match.
		if hasID {
			return fmt.Sprintf("%s/%v/status", basePath, idValue), []string{idArgName}
		}
	case "search":
		return basePath + "/search", nil
	case "list":
		return basePath, nil
	default:
		// Generic `update-<sub>` sub-route: e.g. `update-notes` →
		// PATCH `{basePath}/{id}/notes`. Mirrors the `update-status` pattern
		// so new sub-route PATCH endpoints route correctly without per-verb
		// case statements here.
		if hasID && strings.HasPrefix(action.Name, "update-") {
			suffix := strings.TrimPrefix(action.Name, "update-")
			if suffix != "" {
				return fmt.Sprintf("%s/%v/%s", basePath, idValue, suffix), []string{idArgName}
			}
		}
	}

	return basePath, nil
}

// expandPathTemplate replaces every `{argName}` token in tmpl with the matching
// value from args. Returns the expanded path and the list of arg names that
// were consumed. Unknown placeholders are left intact so callers can spot a
// malformed template at request time rather than via a silently-wrong URL.
func expandPathTemplate(tmpl string, args map[string]any) (string, []string) {
	var consumed []string
	out := tmpl
	for {
		start := strings.Index(out, "{")
		if start < 0 {
			break
		}
		end := strings.Index(out[start:], "}")
		if end < 0 {
			break
		}
		end += start
		name := out[start+1 : end]
		value, ok := args[name]
		if !ok {
			// Stop expanding at the first unknown placeholder so the caller
			// sees the raw token and can diagnose the registry typo.
			break
		}
		out = out[:start] + fmt.Sprintf("%v", value) + out[end+1:]
		consumed = append(consumed, name)
	}
	return out, consumed
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
