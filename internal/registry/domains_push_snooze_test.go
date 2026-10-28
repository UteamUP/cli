package registry

import "testing"

func TestPushSnoozeUsesOwnPreferenceAndSendsExplicitZero(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "user-ui-state", "set-push-snooze", "--minutes", "0")
	if err != nil || len(requests) != 1 {
		t.Fatalf("request failed: %v, %#v", err, requests)
	}
	request := requests[0]
	if request.method != "PUT" || request.path != "/api/userpreferences/push-snooze" || request.body["durationMinutes"] != float64(0) || len(request.body) != 1 {
		t.Fatalf("unexpected own-user resume: %#v", request)
	}
}

func TestPushSnoozeRequiresExplicitDuration(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "user-ui-state", "set-push-snooze")
	if err == nil || len(requests) != 0 {
		t.Fatalf("missing duration must not silently resume: %v, %#v", err, requests)
	}
}
