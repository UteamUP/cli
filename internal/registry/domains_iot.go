package registry

const iotCommandAPIPath = "/api/iot/commands"

const iotDeviceAPIPath = "/api/iot"

func init() {
	Register(&Domain{
		Name:        "iot",
		Aliases:     []string{"internet-of-things"},
		Description: "Inspect the selected tenant's dedicated IoT Preview environment",
		Actions: []Action{
			{
				Name:         "status",
				Description:  "Show environment, pricing, usage and lifecycle status",
				ToolName:     "UteamupIoTEnvironmentStatus",
				RESTBasePath: "/api/iot/environment",
				HTTPMethod:   "GET",
			},
			{
				Name:         "monitoring",
				Description:  "Show device health, data freshness, queue backlogs, credentials and alerts",
				ToolName:     "UteamupIoTMonitoringDashboard",
				RESTBasePath: "/api/iot/monitoring",
				HTTPMethod:   "GET",
			},
			{
				Name:         "overview",
				Description:  "Summarize electrical measurements, coverage and energy over a complete window",
				ToolName:     "UteamupIoTTelemetryOverview",
				RESTBasePath: "/api/iot/monitoring",
				HTTPMethod:   "GET",
				RESTPath:     "telemetry/overview",
				Flags: []FlagDef{
					{Name: "from", Description: "UTC observation start (ISO-8601); default last 24h, at most 31 days", Type: "string"},
					{Name: "to", Description: "UTC observation end (ISO-8601); default now", Type: "string"},
				},
			},
			{
				Name:         "telemetry",
				Description:  "Query normalized telemetry with bounded cursor pagination",
				ToolName:     "UteamupIoTTelemetryPoints",
				RESTBasePath: "/api/iot/monitoring",
				HTTPMethod:   "GET",
				RESTPath:     "telemetry/points",
				Flags: []FlagDef{
					{Name: "from", Description: "UTC range start (ISO-8601); default 24h before --to, range at most 90 days", Type: "string"},
					{Name: "to", Description: "UTC range end (ISO-8601); default now", Type: "string"},
					{Name: "device-guid", Description: "Device GUID filter", Type: "string"},
					{Name: "asset-guid", Description: "Asset GUID filter", Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition GUID filter", Type: "string"},
					{Name: "limit", Description: "Page size (1-500)", Type: "int", Default: 100},
					{Name: "before-received-at", Description: "UTC cursor timestamp", Type: "string"},
					{Name: "before-point-guid", Description: "Point GUID paired with cursor timestamp", Type: "string"},
				},
			},
			{
				Name:         "rules",
				Description:  "List baseline, threshold and heartbeat automation rules",
				ToolName:     "UteamupIoTRulesList",
				RESTBasePath: "/api/iot/rules",
				HTTPMethod:   "GET",
			},
			{
				Name:         "device-types",
				Description:  "List IoT device types; a type GUID is needed to create a device",
				ToolName:     "UteamupIoTDeviceTypesList",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "device-types",
			},
			{
				Name:         "devices",
				Description:  "List IoT devices with status and connectionState (noActiveKey, neverConnected or connected)",
				ToolName:     "UteamupIoTDevicesList",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "devices",
				Flags: []FlagDef{
					{Name: "search", Description: "Text in the device name, device id or serial number", Type: "string"},
					{Name: "page", Description: "Page number, from 1", Type: "int", Default: 1},
					{Name: "page-size", BodyName: "pageSize", Description: "Devices per page (1-100)", Type: "int", Default: 25},
				},
			},
			{
				Name:         "device",
				Description:  "Get one IoT device by GUID",
				ToolName:     "UteamupIoTDeviceGet",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "devices/{deviceGuid}",
				Args:         []ArgDef{{Name: "deviceGuid", Description: "Device GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:         "device-create",
				Description:  "Create an IoT device; then issue it a key with device-key-issue",
				ToolName:     "UteamupIoTDeviceCreate",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "POST",
				RESTPath:     "devices",
				Flags: append([]FlagDef{
					{Name: "name", Description: "Device name", Required: true, Type: "string"},
					{Name: "device-id", BodyName: "deviceId", Description: "Identifier the device reports with, e.g. a Home Assistant entity id", Required: true, Type: "string"},
					{Name: "device-type-guid", BodyName: "deviceTypeGuid", Description: "Device type GUID from 'ut iot device-types'", Required: true, Type: "uuid"},
				}, iotDeviceDetailFlags()...),
			},
			{
				Name:         "device-update",
				Description:  "Update an IoT device; omitted optional fields keep their value",
				ToolName:     "UteamupIoTDeviceUpdate",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "PUT",
				RESTPath:     "devices/{deviceGuid}",
				Args:         []ArgDef{{Name: "deviceGuid", Description: "Device GUID", Required: true, Type: "uuid"}},
				// The REST model always applies isActive, so the flag is required rather than defaulting to false.
				Flags: append([]FlagDef{
					{Name: "name", Description: "Device name", Required: true, Type: "string"},
					{Name: "is-active", BodyName: "isActive", Description: "Active (true) or inactive (false)", Required: true, Type: "bool"},
				}, iotDeviceDetailFlags()...),
			},
			{
				Name:         "device-keys",
				Description:  "List a device's keys (prefix, expiry, last use, active); secrets are never shown again",
				ToolName:     "UteamupIoTDeviceKeysList",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "devices/{deviceGuid}/api-keys",
				Args:         []ArgDef{{Name: "deviceGuid", Description: "Device GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:         "device-key-issue",
				Description:  "Issue a device key; the secret is shown once (MQTT username {tenantGuid}:{deviceGuid})",
				ToolName:     "UteamupIoTDeviceKeyIssue",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "POST",
				RESTPath:     "devices/{deviceGuid}/api-keys",
				Args:         []ArgDef{{Name: "deviceGuid", Description: "Device GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "name", Description: "Key name (1-100 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Key description (up to 500 characters)", Type: "string"},
					{Name: "expires-at", BodyName: "expiresAt", Description: "UTC expiry (ISO-8601), in the future; omit for none", Type: "string"},
				},
			},
			{
				Name:         "device-key-revoke",
				Description:  "Revoke a device key; the device can no longer authenticate with it",
				ToolName:     "UteamupIoTDeviceKeyRevoke",
				RESTBasePath: iotDeviceAPIPath,
				HTTPMethod:   "DELETE",
				RESTPath:     "api-keys/{apiKeyGuid}",
				Args:         []ArgDef{{Name: "apiKeyGuid", Description: "Device key GUID from 'ut iot device-keys'", Required: true, Type: "uuid"}},
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
				},
			},
			{
				Name:         "command-definition-history",
				Description:  "List active and inactive command-definition revisions with lifecycle evidence",
				ToolName:     "UteamupIoTCommandDefinitionHistory",
				RESTBasePath: iotCommandAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "definitions/history",
				Flags: []FlagDef{
					{Name: "device-type-guid", BodyName: "deviceTypeGuid", Description: "Optional IoT device type GUID", Type: "uuid"},
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
		flags = append(flags,
			FlagDef{
				Name:        "expected-updated-at",
				BodyName:    "expectedUpdatedAt",
				Description: "Current definition UpdatedAt value",
				Required:    true,
				Type:        "string",
			},
			FlagDef{
				Name:        "is-active",
				BodyName:    "isActive",
				Description: "Whether the new immutable revision can be previewed",
				Type:        "bool",
				Default:     true,
			},
			FlagDef{
				Name:        "lifecycle-reason",
				BodyName:    "lifecycleReason",
				Description: "Reason required when active availability changes",
				Type:        "string",
			},
		)
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

func iotDeviceDetailFlags() []FlagDef {
	return []FlagDef{
		{Name: "serial-number", BodyName: "serialNumber", Description: "Serial number", Type: "string"},
		{Name: "firmware-version", BodyName: "firmwareVersion", Description: "Firmware version", Type: "string"},
		{Name: "description", Description: "Description", Type: "string"},
		{Name: "location-guid", BodyName: "locationGuid", Description: "Location GUID", Type: "uuid"},
		{Name: "location-floor-guid", BodyName: "locationFloorGuid", Description: "Room GUID (a location node from 'ut location node-tree'); implies its location and must lie inside --location-guid when both are given", Type: "uuid"},
		{Name: "asset-guid", BodyName: "assetGuid", Description: "GUID of the asset the device measures", Type: "uuid"},
		{Name: "ip-address", BodyName: "ipAddress", Description: "IP address", Type: "string"},
		{Name: "mac-address", BodyName: "macAddress", Description: "MAC address", Type: "string"},
		{Name: "tag-guids", BodyName: "tagGuids", Description: "Tag GUIDs, repeatable or comma-separated (complete set)", Type: "stringSlice"},
	}
}
