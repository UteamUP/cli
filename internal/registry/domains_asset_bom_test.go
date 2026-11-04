package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssetBomRoutesKeepAssetSessionAndEvidenceGuids(t *testing.T) {
	domain := findDomain("asset-bom")
	const asset = "11111111-1111-4111-8111-111111111111"
	const session = "22222222-2222-4222-8222-222222222222"
	const evidence = "33333333-3333-4333-8333-333333333333"
	prefix := "/api/asset/" + asset + "/bom/"
	tests := []struct{ action, method, path, tool string }{
		{"sources", "GET", "sources", "UteamupAssetBomSources"},
		{"review-create", "POST", "review-sessions", "UteamupAssetBomReviewCreate"},
		{"review-get", "GET", "review-sessions/" + session, "UteamupAssetBomReviewGet"},
		{"analyze", "POST", "review-sessions/" + session + "/analyze", "UteamupAssetBomReviewAnalyze"},
		{"patch-row", "PATCH", "review-sessions/" + session + "/candidates/" + evidence, "UteamupAssetBomReviewPatchRow"},
		{"patch-page", "PATCH", "review-sessions/" + session + "/pages/" + evidence, "UteamupAssetBomReviewPatchPage"},
		{"apply", "POST", "review-sessions/" + session + "/commit", "UteamupAssetBomReviewApply"},
		{"delete", "DELETE", "review-sessions/" + session, "UteamupAssetBomReviewDelete"},
	}
	for _, item := range tests {
		action := findAction(domain, item.action)
		if action == nil || action.MCPOnly || action.HTTPMethod != item.method || action.ToolName != item.tool || !action.RejectExtraArgs {
			t.Fatalf("invalid BOM action %+v", action)
		}
		args := map[string]any{"assetGuid": asset, "sessionGuid": session, "candidateGuid": evidence, "pageGuid": evidence}
		path, consumed := buildRESTPath(domain, *action, args)
		if path != prefix+item.path || len(consumed) != len(action.Args) {
			t.Fatalf("BOM route %s = %s, consumed %v", item.action, path, consumed)
		}
		for _, arg := range action.Args {
			if !arg.Required || arg.Type != "non-empty-uuid" {
				t.Fatalf("BOM identity is not a required GUID: %+v", arg)
			}
		}
		for _, flag := range action.Flags {
			if flag.Name == "from-json" && (!flag.Required || !flag.RootJSONObjectFile) {
				t.Fatalf("reviewed mutation must use a required root JSON object: %+v", flag)
			}
		}
	}
}

func TestAssetBomReviewedBodiesTravelUnchangedWithoutWrappingOrAutomaticConfirmation(t *testing.T) {
	const asset = "11111111-1111-4111-8111-111111111111"
	const session = "22222222-2222-4222-8222-222222222222"
	const evidence = "33333333-3333-4333-8333-333333333333"
	cases := []struct {
		action, json string
		extra        []string
	}{
		{"analyze", `{"pageGuids":["33333333-3333-4333-8333-333333333333"],"confirmedCreditsPerPage":5,"confirmedEstimatedCredits":5,"analysisMode":"ai"}`, nil},
		{"patch-row", `{"revision":7,"row":{"itemNumber":"A-01","partNumber":"P-02","description":"Seal","quantity":0.125,"matchedPartGuid":"33333333-3333-4333-8333-333333333333","createNewPart":false,"uncertaintyAcknowledged":true},"decision":"approved"}`, []string{evidence}},
		{"patch-page", `{"revision":7,"isSelected":false,"exclusionReason":"Reviewed page has no parts table"}`, []string{evidence}},
		{"apply", `{"revision":7,"acknowledged":false}`, nil},
	}
	for _, item := range cases {
		file := filepath.Join(t.TempDir(), "review.json")
		if err := os.WriteFile(file, []byte(item.json), 0600); err != nil {
			t.Fatal(err)
		}
		args := append([]string{item.action, asset, session}, item.extra...)
		args = append(args, "--from-json", file)
		requests, err := runRegisteredDomainCommand(t, "asset-bom", args...)
		if err != nil || len(requests) != 1 {
			t.Fatalf("request failed: %v, %+v", err, requests)
		}
		body := requests[0].body
		if _, wrapped := body["request"]; wrapped {
			t.Fatal("canonical REST body was wrapped")
		}
		if _, bypass := body["globalAdminBypass"]; bypass {
			t.Fatal("client supplied bypass authority")
		}
		if item.action == "apply" && body["acknowledged"] != false {
			t.Fatal("CLI added automatic confirmation")
		}
		if item.action == "analyze" && (body["confirmedCreditsPerPage"] != float64(5) || body["confirmedEstimatedCredits"] != float64(5)) {
			t.Fatalf("confirmed cost changed: %+v", body)
		}
		if item.action == "patch-row" && body["row"].(map[string]any)["quantity"] != 0.125 {
			t.Fatalf("reviewed fractional quantity changed: %+v", body)
		}
	}
}

func TestAssetBomRejectsIntegerAndEmptyIdentitiesBeforeSending(t *testing.T) {
	for _, asset := range []string{"12", "00000000-0000-0000-0000-000000000000"} {
		requests, err := runRegisteredDomainCommand(t, "asset-bom", "sources", asset)
		if err == nil || len(requests) != 0 {
			t.Fatalf("invalid identity reached backend: %v, %+v", err, requests)
		}
	}
}
