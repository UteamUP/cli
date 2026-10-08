package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSamlLoginFlagsRequireCompanyAndExcludeApiKeyInputs(t *testing.T) {
	oldSaml, oldCompany, oldKeyAuth, oldEmail := loginSAML, loginCompany, loginKeyAuth, loginEmail
	t.Cleanup(func() { loginSAML, loginCompany, loginKeyAuth, loginEmail = oldSaml, oldCompany, oldKeyAuth, oldEmail })
	loginEmail = ""
	for _, test := range []struct {
		saml, key bool
		company   string
		message   string
	}{
		{true, false, "", "--saml requires --company or --email"},
		{false, false, "iteggs", "--company requires --saml"},
		{true, true, "iteggs", "choose SAML login or API key login"},
	} {
		loginSAML, loginCompany, loginKeyAuth = test.saml, test.company, test.key
		if err := runLogin(&cobra.Command{}, nil); err == nil || !strings.Contains(err.Error(), test.message) {
			t.Errorf("invalid login flags accepted or wrong message: %v", err)
		}
	}
}

func TestSamlEmailDiscoveryRequiresExplicitSamlAndExcludesCompanyOverride(t *testing.T) {
	oldSaml, oldCompany, oldEmail, oldKeyAuth := loginSAML, loginCompany, loginEmail, loginKeyAuth
	t.Cleanup(func() { loginSAML, loginCompany, loginEmail, loginKeyAuth = oldSaml, oldCompany, oldEmail, oldKeyAuth })
	loginKeyAuth, loginEmail = false, "gisli@iteggs.com"
	for _, scenario := range []struct {
		saml             bool
		company, message string
	}{
		{false, "", "--email requires --saml"},
		{true, "iteggs", "choose --company or --email"},
	} {
		loginSAML, loginCompany = scenario.saml, scenario.company
		if err := runLogin(&cobra.Command{}, nil); err == nil || !strings.Contains(err.Error(), scenario.message) {
			t.Fatalf("unsafe discovery flags accepted: %v", err)
		}
	}
}
