package registry

const iotCommandAPIPath = "/api/iot/commands"

func init() {
	Register(&Domain{
		Name:        "iot",
		Aliases:     []string{"internet-of-things"},
		Description: "Inspect the selected tenant's dedicated IoT Preview environment",
		Actions: []Action{
			{
				Name:        "status",
				Description: "Show environment, pricing, usage and lifecycle status",
				ToolName:    "UteamupIoTEnvironmentStatus",
			},
			{
				Name:        "monitoring",
				Description: "Show device health, data freshness, queue backlogs, credentials and alerts",
				ToolName:    "UteamupIoTMonitoringDashboard",
			},
			{
				Name:        "telemetry",
				Description: "Query normalized telemetry with bounded cursor pagination",
				ToolName:    "UteamupIoTTelemetryPoints",
				Flags: []FlagDef{
					{Name: "from", Description: "UTC range start (ISO-8601)", Type: "string"},
					{Name: "to", Description: "UTC range end (ISO-8601)", Type: "string"},
					{Name: "device-guid", Description: "Device GUID filter", Type: "string"},
					{Name: "asset-guid", Description: "Asset GUID filter", Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition GUID filter", Type: "string"},
					{Name: "limit", Description: "Page size (1-500)", Type: "int", Default: 100},
					{Name: "before-received-at", Description: "UTC cursor timestamp", Type: "string"},
					{Name: "before-point-guid", Description: "Point GUID paired with cursor timestamp", Type: "string"},
				},
			},
			{
				Name:        "rules",
				Description: "List baseline, threshold and heartbeat automation rules",
				ToolName:    "UteamupIoTRulesList",
			},
			{
				Name:         "command-definitions",
				Description:  "List versioned command definitions and their risk controls",
				ToolName:     "UteamupIoTCommandDefinitionsList",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "definitions",
				Flags: []FlagDef{
					{Name: "device-type-guid", BodyName: "deviceTypeGuid", Description: "Optional IoT device type GUID", Type: "uuid"},
					{Name: "include-inactive", BodyName: "includeInactive", Description: "Include inactive definition versions", Type: "bool", Default: false},
				},
			},
			{
				Name:         "command-definition-create",
				Description:  "Create an allowlisted, versioned IoT command definition",
				ToolName:     "UteamupIoTCommandDefinitionCreate",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "POST",
				RESTPath:     "definitions",
				Flags:        iotCommandDefinitionFlags(false),
			},
			{
				Name:         "command-definition-update",
				Description:  "Create a concurrency-bound revision of a command definition",
				ToolName:     "UteamupIoTCommandDefinitionUpdate",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "PUT",
				RESTPath:     "definitions/{definitionGuid}",
				Args: []ArgDef{
					{Name: "definitionGuid", Description: "Current command-definition GUID", Required: true, Type: "uuid"},
				},
				Flags: iotCommandDefinitionFlags(true),
			},
			{
				Name:         "command-control",
				Description:  "Read global, tenant, and effective command kill switches",
				ToolName:     "UteamupIoTCommandControlGet",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "control",
			},
			{
				Name:         "command-control-update",
				Description:  "Update the tenant command kill switch with a reviewed reason",
				ToolName:     "UteamupIoTCommandControlUpdate",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "PUT",
				RESTPath:     "control",
				Flags: []FlagDef{
					iotCommandIdempotencyFlag(),
					{Name: "commands-enabled", BodyName: "commandsEnabled", Description: "Enable or pause tenant command dispatch", Required: true, Type: "bool"},
					{Name: "reason", Description: "Reviewed kill-switch reason", Required: true, Type: "string"},
					{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Current control UpdatedAt value; omit only when creating it", Type: "string"},
				},
			},
			{
				Name:         "command-preview",
				Description:  "Validate and persist an evidence-bound command preview without dispatch",
				ToolName:     "UteamupIoTCommandRequestPreview",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "POST",
				RESTPath:     "requests/preview",
				Flags: []FlagDef{
					iotCommandIdempotencyFlag(),
					{Name: "definition-guid", BodyName: "definitionGuid", Description: "Active command-definition GUID", Required: true, Type: "uuid"},
					{Name: "device-guid", BodyName: "deviceGuid", Description: "Active IoT device GUID", Required: true, Type: "uuid"},
					{Name: "parameters-file", BodyName: "parameters", Description: "JSON file containing reviewed command parameters", Required: true, Type: "string", JSONFile: true},
					{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Current IoT device UpdatedAt value", Required: true, Type: "string"},
				},
			},
			{
				Name:         "command-requests",
				Description:  "List bounded governed command requests and current status",
				ToolName:     "UteamupIoTCommandRequestsList",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "requests",
				Flags: []FlagDef{
					{Name: "page", Description: "Page number", Type: "int", Default: 1},
					{Name: "page-size", BodyName: "pageSize", Description: "Page size, 1-100", Type: "int", Default: 25},
					{Name: "device-guid", BodyName: "deviceGuid", Description: "Optional IoT device GUID", Type: "uuid"},
					{Name: "status", Description: "Optional request status", Type: "string"},
					{Name: "risk-class", BodyName: "riskClass", Description: "Optional Low or High risk class", Type: "string"},
					{Name: "command-key", BodyName: "commandKey", Description: "Optional exact command key", Type: "string"},
				},
			},
			{
				Name:         "command-request",
				Description:  "Read one command request with approvals, receipts, and outcomes",
				ToolName:     "UteamupIoTCommandRequestGet",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "requests/{requestGuid}",
				Args: []ArgDef{
					{Name: "requestGuid", Description: "IoT command request GUID", Required: true, Type: "uuid"},
				},
			},
			iotCommandDecisionAction(
				"command-confirm",
				"Confirm the requester's exact preview",
				"UteamupIoTCommandRequestConfirm",
				"confirm",
				false,
			),
			iotCommandDecisionAction(
				"command-approve",
				"Approve a confirmed high-risk command as the second actor",
				"UteamupIoTCommandRequestApprove",
				"approve",
				false,
			),
			iotCommandDecisionAction(
				"command-reject",
				"Reject a confirmed high-risk command",
				"UteamupIoTCommandRequestReject",
				"reject",
				false,
			),
			iotCommandDecisionAction(
				"command-cancel",
				"Cancel an eligible command request with an explicit reason",
				"UteamupIoTCommandRequestCancel",
				"cancel",
				true,
			),
			{
				Name:         "command-monitoring",
				Description:  "Review command queue, approval, failure, and receipt health",
				ToolName:     "UteamupIoTCommandMonitoringGet",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "monitoring",
			},
		},
	})
}

func iotCommandIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:               "idempotency-key",
		BodyName:           "idempotencyKey",
		HeaderName:         "Idempotency-Key",
		MirrorHeaderInBody: true,
		Description:        "Retry-stable idempotency GUID",
		Required:           true,
		Type:               "uuid",
	}
}

func iotCommandDefinitionFlags(includeExpectedVersion bool) []FlagDef {
	flags := []FlagDef{
		iotCommandIdempotencyFlag(),
		{Name: "device-type-guid", BodyName: "deviceTypeGuid", Description: "IoT device type GUID", Required: true, Type: "uuid"},
		{Name: "command-key", BodyName: "commandKey", Description: "Stable allowlisted command key", Required: true, Type: "string"},
		{Name: "display-name", BodyName: "displayName", Description: "Human-readable command name", Required: true, Type: "string"},
		{Name: "description", Description: "Optional reviewed description", Type: "string"},
		{Name: "parameters-schema-json", BodyName: "parametersSchemaJson", Description: "Closed JSON Schema for command parameters", Required: true, Type: "string"},
		{Name: "risk-class", BodyName: "riskClass", Description: "Low or High", Type: "string", Default: "Low"},
		{Name: "timeout-seconds", BodyName: "timeoutSeconds", Description: "Provider timeout, 1-300 seconds", Type: "int", Default: 30},
		{Name: "is-reversible", BodyName: "isReversible", Description: "Whether the device command is reversible", Type: "bool", Default: false},
		{Name: "required-permission", BodyName: "requiredPermission", Description: "Permission checked again before dispatch", Type: "string", Default: "IOT.Control"},
		{Name: "minimum-firmware-version", BodyName: "minimumFirmwareVersion", Description: "Optional minimum firmware version", Type: "string"},
		{Name: "maximum-firmware-version", BodyName: "maximumFirmwareVersion", Description: "Optional maximum firmware version", Type: "string"},
	}
	if includeExpectedVersion {
		flags = append(flags, FlagDef{
			Name:        "expected-updated-at",
			BodyName:    "expectedUpdatedAt",
			Description: "Current definition UpdatedAt value",
			Required:    true,
			Type:        "string",
		})
	}
	return flags
}

func iotCommandDecisionAction(
	name string,
	description string,
	toolName string,
	route string,
	requireReason bool,
) Action {
	flags := []FlagDef{
		iotCommandIdempotencyFlag(),
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Current request UpdatedAt value", Required: true, Type: "string"},
		{Name: "comment", Description: "Optional review comment", Type: "string"},
	}
	if requireReason {
		flags = append(flags, FlagDef{
			Name:        "reason",
			Description: "Explicit cancellation reason",
			Required:    true,
			Type:        "string",
		})
	}
	return Action{
		Name:         name,
		Description:  description,
		ToolName:     toolName,
		RESTBasePath: iotCommandAPIPath,
		HTTPMethod:   "POST",
		RESTPath:     "requests/{requestGuid}/" + route,
		Args: []ArgDef{
			{Name: "requestGuid", Description: "IoT command request GUID", Required: true, Type: "uuid"},
		},
		Flags: flags,
	}
}
