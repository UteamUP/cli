package registry

import (
	"reflect"
	"testing"
)

const qualityPolicyAPIPath = "/api/quality/policies"

func TestQualityPolicyDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11)
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search":            {method: "GET", path: qualityPolicyAPIPath, tool: "UteamupQualityPolicySearch"},
		"get":               {method: "GET", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID, tool: "UteamupQualityPolicyGet"},
		"defaults":          {method: "GET", path: qualityPolicyAPIPath + "/defaults/NonConformance", tool: "UteamupQualityPolicyDefaultsGet"},
		"bootstrap-drafts":  {method: "POST", path: qualityPolicyAPIPath + "/bootstrap-drafts", tool: "UteamupQualityPolicyBootstrapDrafts"},
		"draft-create":      {method: "POST", path: qualityPolicyAPIPath, tool: "UteamupQualityPolicyDraftCreate"},
		"draft-update":      {method: "PUT", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID, tool: "UteamupQualityPolicyDraftUpdate"},
		"submit-for-review": {method: "POST", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID + "/submit-for-review", tool: "UteamupQualityPolicySubmitForReview"},
		"return-to-draft":   {method: "POST", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID + "/return-to-draft", tool: "UteamupQualityPolicyReturnToDraft"},
		"publish":           {method: "POST", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID + "/publish", tool: "UteamupQualityPolicyPublish"},
		"supersede":         {method: "POST", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID + "/supersede", tool: "UteamupQualityPolicySupersede"},
		"withdraw":          {method: "POST", path: qualityPolicyAPIPath + "/" + qmsAuditValidGUID + "/withdraw", tool: "UteamupQualityPolicyWithdraw"},
	})
}

func TestQualityPolicySearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(t, qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11), "search")
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"record-kind":        {QueryName: "recordKind", Type: "string"},
		"status":             {QueryName: "status", Type: "string"},
		"scope":              {QueryName: "scope", Type: "string"},
		"site-location-guid": {QueryName: "siteLocationGuid", Type: "non-empty-uuid"},
		"page":               {QueryName: "page", Type: "int", Default: 1},
		"page-size":          {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestQualityPolicyRequestBodyMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"draft-create": {},
		"draft-update": {concurrency: true},
	})
}

func TestQualityPolicyLifecycleActionsRequireHeadersAndConfirmation(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11)
	for _, name := range []string{"submit-for-review", "return-to-draft", "publish", "supersede", "withdraw"} {
		action := qmsAuditAction(t, domain, name)
		if _, hasRequestFile := qmsAuditFindFlag(action, "request-file"); hasRequestFile {
			t.Errorf("%s must not require a request file: its body is empty or a reason object", name)
		}
		idempotency := qmsAuditFlag(t, action, "idempotency-key")
		if !idempotency.Required || idempotency.HeaderName != "Idempotency-Key" {
			t.Errorf("%s idempotency contract = %+v", name, idempotency)
		}
		concurrency := qmsAuditFlag(t, action, "concurrency-token")
		if !concurrency.Required || !concurrency.Sensitive || concurrency.HeaderName != "If-Match" || !concurrency.StrongETag {
			t.Errorf("%s concurrency contract = %+v", name, concurrency)
		}
		confirm := qmsAuditFlag(t, action, "confirm")
		if !confirm.Required || !confirm.MustBeTrue || !confirm.LocalOnly {
			t.Errorf("%s confirmation contract = %+v", name, confirm)
		}
	}
	for _, name := range []string{"return-to-draft", "withdraw", "supersede"} {
		reason := qmsAuditFlag(t, qmsAuditAction(t, domain, name), "reason")
		if !reason.Required || reason.BodyName != "reason" || reason.Type != "string" {
			t.Errorf("%s reason contract = %+v", name, reason)
		}
	}
}

func TestQualityPolicySupersedeCarriesBothValidatorsAndTheSuccessor(t *testing.T) {
	t.Parallel()
	supersede := qmsAuditAction(t, qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11), "supersede")
	predecessor := qmsAuditFlag(t, supersede, "predecessor-concurrency-token")
	if !predecessor.Required || !predecessor.Sensitive || !predecessor.StrongETag ||
		predecessor.HeaderName != "X-Predecessor-If-Match" {
		t.Fatalf("predecessor validator contract = %+v", predecessor)
	}
	successor := qmsAuditFlag(t, supersede, "successor-policy-guid")
	if !successor.Required || successor.BodyName != "successorPolicyGuid" || successor.Type != "non-empty-uuid" {
		t.Fatalf("successor contract = %+v", successor)
	}
}

func TestQualityPolicyBootstrapIsConfirmedAndBodyless(t *testing.T) {
	t.Parallel()
	bootstrap := qmsAuditAction(t, qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11), "bootstrap-drafts")
	if len(bootstrap.Args) != 0 {
		t.Fatalf("bootstrap must take no path arguments, got %v", bootstrap.Args)
	}
	idempotency := qmsAuditFlag(t, bootstrap, "idempotency-key")
	if !idempotency.Required || idempotency.HeaderName != "Idempotency-Key" {
		t.Fatalf("bootstrap idempotency contract = %+v", idempotency)
	}
	if _, has := qmsAuditFindFlag(bootstrap, "concurrency-token"); has {
		t.Fatalf("bootstrap creates new drafts and must not demand a concurrency token")
	}
	confirm := qmsAuditFlag(t, bootstrap, "confirm")
	if !confirm.MustBeTrue || !confirm.LocalOnly {
		t.Fatalf("bootstrap confirmation contract = %+v", confirm)
	}
}

func TestQualityPolicyDefaultsRecordKindsAreClosed(t *testing.T) {
	t.Parallel()
	want := []string{
		"NonConformance",
		"CorrectivePreventiveAction",
		"QualityAuditProgram",
		"QualityAudit",
		"QualityAuditFinding",
		"InspectionTestPlan",
		"InspectionTestPlanExecution",
		"SupplierCorrectiveActionRequest",
	}
	defaults := qmsAuditAction(t, qmsAuditDomain(t, "quality-policy", qualityPolicyAPIPath, 11), "defaults")
	if len(defaults.Args) != 1 || !reflect.DeepEqual(defaults.Args[0].AllowedValues, want) {
		t.Fatalf("defaults record kinds = %v, want %v", defaults.Args, want)
	}
}
