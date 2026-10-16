package registry

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
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
		Description: "Manage customers by public GUID",
		APIPath:     "/api/customer",
		Actions:     customerActions(),
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

// Customer mutations preserve legacy responses unless original receipt metadata
// is supplied. Never generate replacement operation keys or revisions on retry.
func customerActions() []Action {
	identity := []ArgDef{{Name: "customerGuid", Description: "Customer public GUID", Required: true, Type: "non-empty-uuid"}}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed customer JSON with public GUID references; optional idempotencyKey and mutationOutcomeVersion=1 (plus expectedUpdatedAt on update) negotiate a minimal outcome; retain unchanged on retry"
	confirmation := FlagDef{Name: "confirm", Description: "Confirm this customer change; reconcile uncertain outcomes before retrying", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}
	return []Action{
		{Name: "list", Description: "List visible customers", ToolName: "UteamupCustomerList", HTTPMethod: "GET", Flags: paginationFlags()},
		{Name: "get", Description: "Get a customer by public GUID", ToolName: "UteamupCustomerGet", HTTPMethod: "GET", RESTPath: "{customerGuid}", Args: identity},
		{Name: "create", Description: "Create a customer from reviewed JSON", ToolName: "UteamupCustomerCreate", HTTPMethod: "POST", Flags: []FlagDef{fields, confirmation}},
		{Name: "update", Description: "Update a customer by public GUID", ToolName: "UteamupCustomerUpdate", HTTPMethod: "PUT", RESTPath: "{customerGuid}", Args: identity, Flags: []FlagDef{fields, confirmation}},
		{Name: "delete", Description: "Delete a customer after explicit confirmation", ToolName: "UteamupCustomerDelete", HTTPMethod: "DELETE", RESTPath: "{customerGuid}", Args: identity, Flags: []FlagDef{confirmation,
			{Name: "idempotency-key", Description: "Original nonempty operation GUID; supply all three mutation metadata flags together", Type: "non-empty-uuid", QueryName: "idempotencyKey"},
			{Name: "expected-updated-at", Description: "Exact canonical reviewed UTC revision (seconds or up to seven fractional digits, without trailing zeros); retain unchanged on retry", Type: "string", QueryName: "expectedUpdatedAt"},
			{Name: "mutation-outcome-version", Description: "Minimal outcome version (1); requires original operation and reviewed revision", Type: "int", QueryName: "mutationOutcomeVersion"},
		}},
		{Name: "search", Description: "Search visible customer names", ToolName: "UteamupCustomerSearch", HTTPMethod: "GET", UseDomainBasePath: true, Args: []ArgDef{{Name: "query", QueryName: "nameFilter", Description: "Customer name filter", Required: true, Type: "string"}}, Flags: paginationFlags()},
	}
}

var customerRevisionPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,7})?Z$`)

func validateCustomerDeleteMetadata(cmd *cobra.Command, action Action) error {
	if action.ToolName != "UteamupCustomerDelete" {
		return nil
	}
	flags := []string{"idempotency-key", "expected-updated-at", "mutation-outcome-version"}
	supplied := 0
	for _, name := range flags {
		if cmd.Flags().Changed(name) {
			supplied++
		}
	}
	if supplied == 0 {
		return nil
	}
	if supplied != len(flags) {
		return fmt.Errorf("Customer delete requires all three original mutation metadata flags together")
	}
	key, _ := cmd.Flags().GetString("idempotency-key")
	version, _ := cmd.Flags().GetInt("mutation-outcome-version")
	if !uuidValuePattern.MatchString(key) || strings.EqualFold(key, emptyUUIDValue) || version != 1 {
		return fmt.Errorf("Customer delete requires a nonempty operation GUID and mutation outcome version 1")
	}
	revision, _ := cmd.Flags().GetString("expected-updated-at")
	const layout = "2006-01-02T15:04:05.9999999Z"
	parsed, err := time.Parse(layout, revision)
	if !customerRevisionPattern.MatchString(revision) || err != nil || parsed.Year() < 1 || parsed.Format(layout) != revision || parsed.IsZero() {
		return fmt.Errorf("Customer delete requires the exact canonical original UTC revision, with at most seven fractional digits and no trailing zeros")
	}
	return nil
}
