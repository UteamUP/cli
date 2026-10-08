package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSamlLoginFlagsRequireCompanyAndExcludeApiKeyInputs(t *testing.T) {
	oldSaml, oldCompany, oldKeyAuth := loginSAML, loginCompany, loginKeyAuth
	t.Cleanup(func() { loginSAML, loginCompany, loginKeyAuth = oldSaml, oldCompany, oldKeyAuth })
	for _, test := range []struct {
		saml, key bool
		company   string
		message   string
	}{
		{true, false, "", "--saml requires --company"},
		{false, false, "iteggs", "--company requires --saml"},
		{true, true, "iteggs", "choose SAML login or API key login"},
	} {
		loginSAML, loginCompany, loginKeyAuth = test.saml, test.company, test.key
		if err := runLogin(&cobra.Command{}, nil); err == nil || !strings.Contains(err.Error(), test.message) {
			t.Errorf("invalid login flags accepted or wrong message: %v", err)
		}
	}
}
