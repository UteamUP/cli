package registry

import "testing"

func TestDocumentUploadDomainIssuesAndReadsTickets(t *testing.T) {
	t.Parallel()
	domain := findDomain("document-upload")
	if domain == nil {
		t.Fatal("document-upload domain is not registered")
	}
	want := map[string]string{"ticket": "UteamupDocumentCreateUploadTicket", "status": "UteamupDocumentUploadTicketGet"}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		if want[action.Name] != action.ToolName || !action.MCPOnly {
			t.Fatalf("action contract = %+v", action)
		}
	}
}
