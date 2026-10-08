package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var samlCompanyCode = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// DiscoverSAMLCompany uses the selected backend and submits no account address.
// The explicit --saml invocation chooses company SSO; no match leaves normal login available.
func (a *Client) DiscoverSAMLCompany(ctx context.Context, email string) (string, error) {
	domain, err := samlEmailDomain(email)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("company SSO discovery cancelled: %w", err)
	}
	var response map[string]json.RawMessage
	err = a.samlPost(ctx, "/api/auth/saml/discover", map[string]string{"domain": domain}, &response, "")
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", fmt.Errorf("company SSO discovery cancelled: %w", ctxErr)
	}
	if err != nil {
		return "", fmt.Errorf("company SSO discovery unavailable; retry with --company or use normal login")
	}
	raw, present := response["companyCode"]
	var company *string
	if !present || json.Unmarshal(raw, &company) != nil || (company != nil && !samlCompanyCode.MatchString(*company)) {
		return "", fmt.Errorf("company SSO discovery returned an invalid response; use --company or normal login")
	}
	if company == nil {
		return "", fmt.Errorf("no verified company SSO found; use --company or normal login")
	}
	return *company, nil
}

func samlEmailDomain(email string) (string, error) {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 320 || strings.Count(email, "@") != 1 {
		return "", fmt.Errorf("enter an email address for --saml --email, or use --company")
	}
	_, domain, _ := strings.Cut(email, "@")
	domain = strings.ToLower(domain)
	labels := strings.Split(domain, ".")
	valid := len(domain) <= 253 && len(labels) >= 2 && strings.ContainsFunc(labels[len(labels)-1], unicode.IsLetter)
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") ||
			strings.ContainsFunc(label, func(r rune) bool { return r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
			valid = false
		}
	}
	if !valid {
		return "", fmt.Errorf("enter an email with a complete company domain, or use --company")
	}
	return domain, nil
}
