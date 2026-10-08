package cmd

import "testing"

func TestTenantSamlCommandsAreReachableAlongsideMembershipCommands(t *testing.T) {
	for _, name := range []string{"get-saml", "update-saml", "test-saml", "show", "list", "select", "industry-profiles-get"} {
		command, remaining, err := rootCmd.Find([]string{"tenant", name})
		if err != nil || len(remaining) != 0 || command.Parent() != tenantCmd {
			t.Fatalf("tenant %s was shadowed by the generated tenant domain: %v", name, err)
		}
		if name == "test-saml" && command != tenantTestSamlCmd {
			t.Fatal("setup test must use the interactive protected loopback")
		}
	}
}

func TestTenantSamlSetupTestRejectsNonGuidAndExtraArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"iteggs"}, {"../other"}, {"00000000-0000-0000-0000-000000000000"}, {"11111111-1111-4111-8111-111111111111", "extra"}} {
		if err := tenantTestSamlCmd.Args(tenantTestSamlCmd, args); err == nil {
			t.Errorf("invalid setup target was accepted: %v", args)
		}
	}
}
