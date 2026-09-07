package registry

import "testing"

func TestWorkforceSelfServiceRegistrationsAreUniqueAndActorScoped(t *testing.T) {
	for _, group := range []struct {
		domain  string
		actions []string
	}{
		{"shift-assignment", []string{"mine", "confirm-mine", "decline-mine", "self-assign"}},
		{"shift-request", []string{"mine", "incoming", "respond", "withdraw-mine"}},
	} {
		domain := findDomain(group.domain)
		for _, name := range group.actions {
			count := 0
			for _, action := range domain.Actions {
				if action.Name == name {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("%s/%s registered %d times, want exactly once", group.domain, name, count)
			}
			action := findDomainAction(t, group.domain, name)
			for _, arg := range action.Args {
				if arg.Name == "userId" || arg.Name == "userGuid" || arg.Name == "tenantId" || arg.Name == "tenantGuid" || arg.Type != "uuid" {
					t.Fatalf("own routes derive actor and tenant; only record UUID args allowed: %+v", arg)
				}
			}
			for _, flag := range action.Flags {
				if flag.Name == "user-id" || flag.Name == "user-guid" || flag.Name == "tenant-id" || flag.Name == "tenant-guid" {
					t.Fatalf("own routes must not accept identity overrides: %+v", flag)
				}
			}
		}
	}
}

func TestWorkforceSelfServiceDecisionsRequireExplicitInputs(t *testing.T) {
	respond := findDomainAction(t, "shift-request", "respond")
	accept := findFlag(respond, "accept")
	if accept == nil || !accept.Required || accept.Type != "bool" {
		t.Fatal("counterpart response requires explicit accept true or false")
	}
	withdraw := findDomainAction(t, "shift-request", "withdraw-mine")
	token := findFlag(withdraw, "concurrency-token")
	if token == nil || !token.Required || token.QueryName != "concurrencyToken" || token.BodyName != "concurrencyToken" {
		t.Fatal("withdraw must carry the latest token in the query")
	}
}

func TestWorkforceOwnDateFiltersMapForMcpAndRest(t *testing.T) {
	mine := findDomainAction(t, "shift-assignment", "mine")
	for name, wire := range map[string]string{"date-from": "dateFrom", "date-to": "dateTo"} {
		flag := findFlag(mine, name)
		if flag == nil || flag.BodyName != wire || flag.QueryName != wire {
			t.Fatalf("%s must map to %s for both MCP and REST", name, wire)
		}
	}
}
