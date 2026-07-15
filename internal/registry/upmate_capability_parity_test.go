package registry

import "testing"

func TestEnabledUpmateCapabilitiesHaveCLIRegistrations(t *testing.T) {
	expected := []string{
		"UteamupAssetList",
		"UteamupAssetGetByGuid",
		"UteamupIoTMonitoringDashboard",
		"UteamupIoTTelemetryPoints",
		"UteamupMarketplaceBrowse",
		"UteamupMarketplaceRequirementsList",
		"UteamupMarketplaceRequirementCreateDraft",
		"UteamupMarketplaceRequirementPublish",
		"UteamupMarketplaceRequirementOffersCompare",
		"UteamupMarketplaceRequirementOfferAccept",
		"UteamupProjectList",
		"UteamupProjectGet",
		"UteamupScheduleAssignmentGetWorkorderOptions",
		"UteamupScheduleAssignmentCreateByGuid",
		"UteamupShiftList",
		"UteamupShiftGetByGuid",
		"UteamupStockCreateMarketplacePurchaseOrderDraft",
		"UteamupStockGetTenantAlerts",
		"UteamupStockSearchItems",
		"UteamupStockListPurchaseOrders",
		"UteamupStockGetPurchaseOrder",
		"UteamupStockGetAtp",
		"UteamupStockSubmitPurchaseOrder",
		"UteamupStockApprovePurchaseOrder",
		"UteamupStockCreateReservation",
		"UteamupTutorialList",
		"UteamupTutorialGet",
		"UteamupWorkorderPrepareCloseoutByGuid",
		"UteamupWorkorderCompleteCloseoutByGuid",
		"UteamupWorkorderCreateByGuid",
		"UteamupWorkorderTemplateCreateFromTemplateByGuid",
	}

	registered := make(map[string]bool)
	for _, domain := range DefaultRegistry.Domains() {
		for _, action := range domain.Actions {
			registered[action.ToolName] = true
		}
	}

	for _, toolName := range expected {
		if !registered[toolName] {
			t.Errorf("enabled UPMate capability %q has no CLI registration", toolName)
		}
	}
}
