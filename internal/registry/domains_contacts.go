package registry

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	Register(&Domain{
		Name:        "contact",
		Aliases:     []string{"contacts"},
		Description: "Manage contacts",
		Actions: append(contactActions(),
			Action{Name: "search", Description: "Search contacts", ToolName: "UteamupContactSearch", Args: queryArg(), Flags: paginationFlags()},
			Action{
				Name:         "ice-assignments",
				Description:  "List users whose ICE record points at a contact. Reads are privacy-audited.",
				ToolName:     "UteamupEmergencycontactAssignmentsForContact",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
			},
			Action{
				Name:         "ice-assignments-sync",
				Description:  "Replace the employees whose ICE record points at a contact",
				ToolName:     "UteamupEmergencycontactAssignmentsSyncForContact",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}/assignments",
				HTTPMethod:   "PUT",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{{
					Name:               "from-json",
					Description:        "JSON object with reviewed userGuids array; use [] to clear all assignments",
					Type:               "string",
					Required:           true,
					RootJSONObjectFile: true,
				}},
			},
			Action{
				Name:         "ice-assignments-review",
				Description:  "Read a coherent ICE assignment baseline; this does not grant offline authority",
				ToolName:     "UteamupEmergencycontactAssignmentsReview",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
				Flags:        []FlagDef{{Name: "reviewed-version", Type: "int", Default: 1, QueryName: "reviewedVersion", Description: "Explicit reviewed ICE snapshot version (1)"}},
			},
			Action{
				Name:         "ice-assignments-replace-reviewed",
				Description:  "Apply a confirmed original ICE replacement; retain the same body and operation key after an uncertain response",
				ToolName:     "UteamupEmergencycontactAssignmentsReplaceReviewed",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}/assignments",
				HTTPMethod:   "PUT",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from-json", Description: "Exact original four-key JSON: userGuids, idempotencyKey, expectedAssignmentsHash, mutationOutcomeVersion=1; never refresh an unresolved retry", Type: "string", Required: true, RootJSONObjectFile: true},
					{Name: "idempotency-key", Description: "Same original operation GUID as the retained JSON body", Type: "uuid", Required: true, HeaderName: "Idempotency-Key"},
					{Name: "confirm", Description: "Explicitly confirm the reviewed ICE assignment replacement", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true},
				},
			},
		),
	})

	Register(&Domain{
		Name:        "customer",
		Aliases:     []string{"customers"},
		Description: "Manage customers",
		Actions: append(crudActions("Customer"),
			Action{Name: "search", Description: "Search customers", ToolName: "UteamupCustomerSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

}

// contactActions keeps the original normal REST contracts but opts mutations into
// immutable actor-bound outcomes. Never generate a replacement key or refresh a
// reviewed revision automatically after an uncertain response.
func contactActions() []Action {
	intent := FlagDef{Name: "idempotency-key", Description: "One confirmed operation GUID; retain the same complete request after an uncertain result", Required: true, Type: "uuid"}
	version := FlagDef{Name: "mutation-outcome-version", Description: "Contact request version: 1 direct references, 2 original Type-create dependencies", Type: "int", Default: 1}
	revision := FlagDef{Name: "expected-updated-at", Description: "Exact original Contact updatedAt UTC wire value, including all seven fractional digits; never refresh an unresolved retry", Required: true, Type: "string"}
	confirmation := FlagDef{Name: "confirm", Description: "Confirm the reviewed Contact mutation", Required: true, Type: "bool", MustBeTrue: true, LocalOnly: true}
	parents := FlagDef{Name: "contact-type-create-operation-guids", Description: "1–128 sorted unique canonical original Type-create operation GUIDs; requires request version 2", Type: "stringSlice"}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed Contact operational fields JSON; keep the original file unchanged across uncertain retries and supply receipt metadata separately"
	return []Action{
		{Name: "list", Description: "List visible contacts", ToolName: "UteamupContactList", RESTPath: "paginated", Flags: paginationFlags()},
		{Name: "get", Description: "Get a visible contact by public GUID", ToolName: "UteamupContactGet", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}},
		{Name: "create", Description: "Create a contact with an original operation key and minimal acknowledgment", ToolName: "UteamupContactCreate", Flags: []FlagDef{intent, version, parents, confirmation, fields}},
		{Name: "update", Description: "Update the reviewed Contact GUID and exact original revision", ToolName: "UteamupContactUpdate", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{intent, version, parents, revision, confirmation, fields}},
		{Name: "delete", Description: "Delete the reviewed Contact GUID and retain the original key and revision for retries", ToolName: "UteamupContactDelete", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{intent, version, revision, confirmation}},
	}
}

func validateContactLogicalParents(cmd *cobra.Command, action Action) error {
	if action.ToolName != "UteamupContactCreate" && action.ToolName != "UteamupContactUpdate" {
		return nil
	}
	version, _ := cmd.Flags().GetInt("mutation-outcome-version")
	supplied := cmd.Flags().Changed("contact-type-create-operation-guids")
	if version == 1 && !supplied {
		return nil
	}
	if version != 2 || !supplied {
		return fmt.Errorf("Contact logical references require version 2 and original parent GUIDs")
	}
	parents, _ := cmd.Flags().GetStringSlice("contact-type-create-operation-guids")
	if len(parents) < 1 || len(parents) > 128 {
		return fmt.Errorf("Contact logical references must contain 1–128 parent GUIDs")
	}
	for index, parent := range parents {
		if !uuidValuePattern.MatchString(parent) || parent == emptyUUIDValue || strings.ToLower(parent) != parent ||
			(index > 0 && parents[index-1] >= parent) {
			return fmt.Errorf("Contact parent GUIDs must be canonical, sorted and unique")
		}
	}
	return nil
}
