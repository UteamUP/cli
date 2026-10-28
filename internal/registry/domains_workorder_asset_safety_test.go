package registry

import (
	"testing"
)

func TestWorkorderAssetSafetyKeepsReadCaptureAndAcknowledgementSeparate(t *testing.T) {
	const wo = "11111111-1111-4111-8111-111111111111"
	const asset = "22222222-2222-4222-8222-222222222222"
	const snapshot = "33333333-3333-4333-8333-333333333333"
	for _, tc := range []struct {
		action, method, suffix string
		flags                  []string
	}{
		{"asset-safety", "GET", "", nil},
		{"asset-safety-capture", "POST", "/capture", nil},
		{"asset-safety-acknowledge", "POST", "/acknowledge", []string{"--asset-guid", asset, "--snapshot-guid", snapshot}},
	} {
		t.Run(tc.action, func(t *testing.T) {
			args := append([]string{tc.action, wo}, tc.flags...)
			requests, err := runRegisteredDomainCommand(t, "workorder", args...)
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) != 1 {
				t.Fatalf("requests = %d", len(requests))
			}
			r := requests[0]
			if r.method != tc.method || r.path != "/api/workorder/by-guid/"+wo+"/asset-safety"+tc.suffix {
				t.Fatalf("route = %s %s", r.method, r.path)
			}
			if r.header.Get("Authorization") != "Bearer registered-domain-token" {
				t.Fatal("named session missing")
			}
			if tc.suffix == "/acknowledge" {
				if len(r.body) != 2 || r.body["assetGuid"] != asset || r.body["snapshotGuid"] != snapshot {
					t.Fatalf("acknowledgement body = %v", r.body)
				}
			} else if len(r.body) != 0 {
				t.Fatalf("unexpected implicit acknowledgement = %v", r.body)
			}
		})
	}
}

func TestWorkorderAssetSafetyRequiresAnExplicitSnapshotAndPublicGuids(t *testing.T) {
	for _, args := range [][]string{
		{"asset-safety", "42"},
		{"asset-safety-capture", "00000000-0000-0000-0000-000000000000"},
		{"asset-safety-acknowledge", "11111111-1111-4111-8111-111111111111", "--asset-guid", "22222222-2222-4222-8222-222222222222"},
	} {
		requests, err := runRegisteredDomainCommand(t, "workorder", args...)
		if err == nil || len(requests) != 0 {
			t.Fatalf("invalid review must not send: %v %v", args, requests)
		}
	}
}
