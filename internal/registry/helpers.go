package registry

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// newRequestGUID returns a random version-4 GUID string.
func newRequestGUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}

// paginationFlags returns standard pagination flags used across most list actions.
func paginationFlags() []FlagDef {
	return []FlagDef{
		{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
		{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
	}
}

// idArg returns a standard required integer ID positional argument.
//
// Legacy. New GUID-first domains should declare an `externalGuid` positional
// arg inline — per CLIGuidelines.md ("GUID-first domains should declare their
// positional arg as externalGuid"). Existing callers below remain on int ID
// for now because the corresponding backend routes still accept {id:int};
// migration is tracked separately.
func idArg() []ArgDef {
	return []ArgDef{{Name: "id", Description: "Record ID", Required: true, Type: "int"}}
}

// queryArg returns a required search query positional argument.
func queryArg() []ArgDef {
	return []ArgDef{{Name: "query", Description: "Search term", Required: true, Type: "string"}}
}

// externalGuidArg returns a required Guid positional argument named `externalGuid`.
// GUID-first domains (per CLIGuidelines.md) use this in place of idArg().
func externalGUIDArg() []ArgDef {
	return []ArgDef{{Name: "externalGuid", Description: "Record GUID", Required: true, Type: "uuid"}}
}

// jsonFlag returns the --from-json flag for JSON file input.
func jsonFlag() FlagDef {
	return FlagDef{
		Name:               "from-json",
		Description:        "JSON file with request data",
		Type:               "string",
		RootJSONObjectFile: true,
	}
}

// crudActions returns standard CRUD actions for a domain using the legacy
// integer `id` positional arg. New GUID-first domains should declare their
// own actions with an `externalGuid` positional arg so the CLI surface
// advertises `externalGuid` per CLIGuidelines.md.
func crudActions(entityPrefix string) []Action {
	return []Action{
		{Name: "list", Description: "List records", ToolName: "Uteamup" + entityPrefix + "List", Flags: paginationFlags()},
		{Name: "get", Description: "Get by ID", ToolName: "Uteamup" + entityPrefix + "Get", Args: idArg()},
		{Name: "create", Description: "Create a record", ToolName: "Uteamup" + entityPrefix + "Create", Flags: []FlagDef{jsonFlag()}},
		{Name: "update", Description: "Update a record", ToolName: "Uteamup" + entityPrefix + "Update", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
		{Name: "delete", Description: "Delete a record", ToolName: "Uteamup" + entityPrefix + "Delete", Args: idArg()},
	}
}

// listGetActions returns list + get actions for read-only or simple domains.
// Uses the legacy integer `id` positional arg; new GUID-first domains should
// declare their own actions with an `externalGuid` positional arg per CLIGuidelines.md.
func listGetActions(entityPrefix string) []Action {
	return []Action{
		{Name: "list", Description: "List records", ToolName: "Uteamup" + entityPrefix + "List", Flags: paginationFlags()},
		{Name: "get", Description: "Get by ID", ToolName: "Uteamup" + entityPrefix + "Get", Args: idArg()},
	}
}
