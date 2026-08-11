package registry

import "testing"

func paymentProvidersAction(t *testing.T, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("payment-providers")
	if domain == nil {
		t.Fatal("payment-providers domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("payment-providers action %q is not registered", actionName)
	return nil, Action{}
}

func TestPaymentProvidersRoutesMirrorGlobalAdminController(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		method     string
		args       map[string]any
		path       string
	}{
		{"list", "GET", map[string]any{}, "/api/globaladmin/payment-providers"},
		{"health-check", "POST", map[string]any{"providerName": "kling"}, "/api/globaladmin/payment-providers/kling/health-check"},
		{"set-active", "POST", map[string]any{"providerName": "kling"}, "/api/globaladmin/payment-providers/kling/set-active"},
	}
	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := paymentProvidersAction(t, test.actionName)
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q", action.HTTPMethod, test.method)
			}
			path, _ := buildRESTPath(domain, action, test.args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
		})
	}
}

func TestPaymentProvidersDeclaresNoIntIdentifier(t *testing.T) {
	t.Parallel()
	domain := findDomain("payment-providers")
	if domain == nil {
		t.Fatal("payment-providers domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" {
				t.Fatalf("action %s declares int arg %s — provider identity is a name key, never an int id", action.Name, arg.Name)
			}
		}
	}
}
