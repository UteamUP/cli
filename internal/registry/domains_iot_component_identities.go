package registry

func iotIdentityReviewFlags() []FlagDef {
	return []FlagDef{
		{Name: "operation-id", BodyName: "operationId", Description: "Original operation GUID; preserve exact terms on uncertain retries", Required: true, Type: "non-empty-uuid"},
		{Name: "expected-source-revision", BodyName: "expectedSourceRevision", Description: "Exact current source revision GUID", Required: true, Type: "non-empty-uuid"},
		{Name: "reason", Description: "Reviewed reason, maximum1000 characters", Required: true, Type: "string"},
	}
}

func iotComponentIdentityActions() []Action {
	source := []ArgDef{{Name: "sourceGuid", Description: "Existing collector public GUID", Required: true, Type: "non-empty-uuid"}}
	return []Action{
		{
			Name: "source-generations", Description: "List retained collector generations; legacy credentials are not promoted",
			ToolName: "UteamupIoTSourceGenerationsList", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "GET", RESTPath: "devices/{sourceGuid}/source-generations", Args: source, RejectExtraArgs: true,
		},
		{
			Name: "source-enroll", Description: "Enroll or re-enroll a reviewed collector; re-enrollment retires prior component credentials",
			ToolName: "UteamupIoTSourceEnroll", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "POST", RESTPath: "devices/{sourceGuid}/source-generations", Args: source, RejectExtraArgs: true,
			Flags: append(iotIdentityReviewFlags(), FlagDef{Name: "source-kind", BodyName: "sourceKind", Description: "Explicit collector/source kind, maximum64 characters", Required: true, Type: "string"}),
		},
		{
			Name: "source-retire", Description: "Retire an exact reviewed collector generation while retaining history",
			ToolName: "UteamupIoTSourceRetire", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "POST", RESTPath: "devices/{sourceGuid}/source-generations/retire", Args: source, RejectExtraArgs: true,
			Flags: append(iotIdentityReviewFlags(), FlagDef{Name: "source-generation-guid", BodyName: "sourceGenerationGuid", Description: "Exact enrolled generation GUID", Required: true, Type: "non-empty-uuid"}),
		},
		{
			Name: "source-key-issue", Description: "Issue a generation-bound component source key; protect the one-time returned secret",
			ToolName: "UteamupIoTSourceKeyIssue", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "POST", RESTPath: "devices/{sourceGuid}/api-keys", Args: source, RejectExtraArgs: true,
			Flags: []FlagDef{
				{Name: "name", Description: "Credential display name", Required: true, Type: "string"},
				{Name: "description", Description: "Optional credential description", Type: "string"},
				{Name: "expires-at", BodyName: "expiresAt", Description: "Optional explicit UTC expiry (ISO8601)", Type: "string"},
				{Name: "credential-purpose", BodyName: "credentialPurpose", Description: "Explicitly choose componentSourceV2", Required: true, Type: "string", AllowedValues: []string{"componentSourceV2"}},
				{Name: "source-generation-guid", BodyName: "sourceGenerationGuid", Description: "Exact current enrolled generation GUID", Required: true, Type: "non-empty-uuid"},
			},
		},
		{
			Name: "physical-identities", Description: "Page reviewed physical aliases for one source generation with current IoT and stock visibility",
			ToolName: "UteamupIoTPhysicalIdentitiesList", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "GET", RESTPath: "devices/{sourceGuid}/identities", Args: source, RejectExtraArgs: true,
			Flags: []FlagDef{
				{Name: "generation-guid", QueryName: "generationGuid", Description: "Exact retained source generation GUID", Required: true, Type: "non-empty-uuid"},
				{Name: "page", QueryName: "page", Description: "Page starting at1", Type: "int", Default: 1},
				{Name: "page-size", QueryName: "pageSize", Description: "Bounded aliases per page", Type: "int", Default: 50},
			},
		},
		{
			Name: "physical-identity-review", Description: "Review one exact hardware alias, physical unit and channel semantics; does not install or move stock",
			ToolName: "UteamupIoTPhysicalIdentityReview", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "POST", RESTPath: "devices/{sourceGuid}/identities/review", Args: source, RejectExtraArgs: true,
			Flags: append(iotIdentityReviewFlags(),
				FlagDef{Name: "source-generation-guid", BodyName: "sourceGenerationGuid", Description: "Reviewed collector generation GUID", Required: true, Type: "non-empty-uuid"},
				FlagDef{Name: "identity-file", BodyName: "identity", Description: "JSON object with exact namespace and stableHardwareId; no guessed MAC/name aliases", Required: true, Type: "string", JSONFile: true},
				FlagDef{Name: "expected-identity-revision", BodyName: "expectedIdentityRevision", Description: "Current identity revision; omit only for a new reviewed alias", Type: "non-empty-uuid"},
				FlagDef{Name: "unit-guid", BodyName: "unitGuid", Description: "Reviewed serialized physical unit GUID", Required: true, Type: "non-empty-uuid"},
				FlagDef{Name: "expected-unit-sequence", BodyName: "expectedUnitSequence", Description: "Exact current unit sequence, including explicit0", Required: true, Type: "int"},
				FlagDef{Name: "effective-from", BodyName: "effectiveFrom", Description: "Inclusive reviewed UTC boundary (ISO8601)", Required: true, Type: "string"},
				FlagDef{Name: "effective-to", BodyName: "effectiveTo", Description: "Optional exclusive UTC boundary; omit for an open interval", Type: "string"},
				FlagDef{Name: "channels-file", BodyName: "channels", Description: "Reviewed channel array with exact identifiers, existing attributeDefinitionGuid, unit, valueKind, quality and expectsTelemetry", Required: true, Type: "string", JSONFile: true}),
		},
		{
			Name: "physical-identity-end", Description: "End an exact reviewed binding, optionally block the alias, and retain its historical evidence",
			ToolName: "UteamupIoTPhysicalIdentityEnd", RESTBasePath: iotDeviceAPIPath,
			HTTPMethod: "POST", RESTPath: "devices/{sourceGuid}/identities/{identityGuid}/end", RejectExtraArgs: true,
			Args: append(append([]ArgDef{}, source...), ArgDef{Name: "identityGuid", Description: "Reviewed physical identity public GUID", Required: true, Type: "non-empty-uuid"}),
			Flags: append(iotIdentityReviewFlags(),
				FlagDef{Name: "expected-identity-revision", BodyName: "expectedIdentityRevision", Description: "Exact current identity revision GUID", Required: true, Type: "non-empty-uuid"},
				FlagDef{Name: "expected-unit-sequence", BodyName: "expectedUnitSequence", Description: "Exact current unit sequence, including explicit0", Required: true, Type: "int"},
				FlagDef{Name: "effective-at", BodyName: "effectiveAt", Description: "Exclusive binding end as an explicit UTC instant (ISO8601)", Required: true, Type: "string"},
				FlagDef{Name: "block-identity", BodyName: "blockIdentity", Description: "Explicitly block the alias after ending its binding", Type: "bool"}),
		},
	}
}
