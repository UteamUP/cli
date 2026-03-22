package grouper

import (
	"sort"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// selectRepresentative picks the result with the highest classification
// confidence from the group. Panics if the group is empty.
func selectRepresentative(group []models.ImageAnalysisResult) models.ImageAnalysisResult {
	best := 0
	for i := 1; i < len(group); i++ {
		if group[i].Classification.Confidence > group[best].Classification.Confidence {
			best = i
		}
	}
	return group[best]
}

// mergeExtractedData fills nil/empty fields on the representative's extracted
// data from members, preferring higher-confidence members first.
func mergeExtractedData(rep *models.ImageAnalysisResult, members []models.ImageAnalysisResult) {
	if len(members) == 0 {
		return
	}

	// Sort members by descending confidence.
	sorted := make([]models.ImageAnalysisResult, len(members))
	copy(sorted, members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Classification.Confidence > sorted[j].Classification.Confidence
	})

	for _, member := range sorted {
		mergeAsset(rep.ExtractedData.Asset, member.ExtractedData.Asset)
		mergeTool(rep.ExtractedData.Tool, member.ExtractedData.Tool)
		mergePart(rep.ExtractedData.Part, member.ExtractedData.Part)
		mergeChemical(rep.ExtractedData.Chemical, member.ExtractedData.Chemical)
	}
}

// fillStr copies src into dst if dst is nil and src is not nil.
func fillStr(dst **string, src *string) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

// fillFloat copies src into dst if dst is nil and src is not nil.
func fillFloat(dst **float64, src *float64) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

// fillBool copies src into dst if dst is nil and src is not nil.
func fillBool(dst **bool, src *bool) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

func mergeAsset(dst, src *models.ExtractedAssetData) {
	if dst == nil || src == nil {
		return
	}
	if dst.Name == "" && src.Name != "" {
		dst.Name = src.Name
	}
	fillStr(&dst.Description, src.Description)
	fillStr(&dst.SerialNumber, src.SerialNumber)
	fillStr(&dst.ReferenceNumber, src.ReferenceNumber)
	fillStr(&dst.ModelNumber, src.ModelNumber)
	fillStr(&dst.UPCNumber, src.UPCNumber)
	fillStr(&dst.AdditionalInfo, src.AdditionalInfo)
	fillStr(&dst.Notes, src.Notes)
	fillStr(&dst.CheckInProcedure, src.CheckInProcedure)
	fillStr(&dst.CheckOutProcedure, src.CheckOutProcedure)
	fillStr(&dst.IconName, src.IconName)
	fillStr(&dst.SuggestedVendor, src.SuggestedVendor)
	fillStr(&dst.SuggestedCategory, src.SuggestedCategory)
	fillStr(&dst.SuggestedLocation, src.SuggestedLocation)
	fillStr(&dst.ManufacturerBrand, src.ManufacturerBrand)
	fillStr(&dst.VisibleCondition, src.VisibleCondition)
	fillBool(&dst.IsVehicle, src.IsVehicle)
	fillStr(&dst.VehicleType, src.VehicleType)
	fillStr(&dst.LicensePlate, src.LicensePlate)
	fillStr(&dst.AssetCategoryGroup, src.AssetCategoryGroup)
}

func mergeTool(dst, src *models.ExtractedToolData) {
	if dst == nil || src == nil {
		return
	}
	if dst.Name == "" && src.Name != "" {
		dst.Name = src.Name
	}
	fillStr(&dst.Description, src.Description)
	fillStr(&dst.SerialNumber, src.SerialNumber)
	fillStr(&dst.ReferenceNumber, src.ReferenceNumber)
	fillStr(&dst.ModelNumber, src.ModelNumber)
	fillStr(&dst.BarcodeNumber, src.BarcodeNumber)
	fillStr(&dst.ToolNumber, src.ToolNumber)
	fillStr(&dst.AdditionalInfo, src.AdditionalInfo)
	fillStr(&dst.Notes, src.Notes)
	fillStr(&dst.SuggestedVendor, src.SuggestedVendor)
	fillStr(&dst.SuggestedCategory, src.SuggestedCategory)
	fillStr(&dst.ManufacturerBrand, src.ManufacturerBrand)
	fillFloat(&dst.Width, src.Width)
	fillFloat(&dst.Height, src.Height)
	fillFloat(&dst.Length, src.Length)
	fillFloat(&dst.Depth, src.Depth)
	fillFloat(&dst.Weight, src.Weight)
	fillFloat(&dst.Value, src.Value)
}

func mergePart(dst, src *models.ExtractedPartData) {
	if dst == nil || src == nil {
		return
	}
	if dst.Name == "" && src.Name != "" {
		dst.Name = src.Name
	}
	fillStr(&dst.Description, src.Description)
	fillStr(&dst.SerialNumber, src.SerialNumber)
	fillStr(&dst.ReferenceNumber, src.ReferenceNumber)
	fillStr(&dst.ModelNumber, src.ModelNumber)
	fillStr(&dst.PartNumber, src.PartNumber)
	fillStr(&dst.AdditionalInfo, src.AdditionalInfo)
	fillStr(&dst.Notes, src.Notes)
	fillFloat(&dst.Value, src.Value)
	fillStr(&dst.SuggestedVendor, src.SuggestedVendor)
	fillStr(&dst.SuggestedCategory, src.SuggestedCategory)
	fillStr(&dst.ManufacturerBrand, src.ManufacturerBrand)
}

func mergeChemical(dst, src *models.ExtractedChemicalData) {
	if dst == nil || src == nil {
		return
	}
	if dst.Name == "" && src.Name != "" {
		dst.Name = src.Name
	}
	fillStr(&dst.Description, src.Description)
	fillStr(&dst.ChemicalFormula, src.ChemicalFormula)
	fillStr(&dst.CASNumber, src.CASNumber)
	fillStr(&dst.ECNumber, src.ECNumber)
	fillStr(&dst.UNNumber, src.UNNumber)
	fillStr(&dst.GHSHazardClass, src.GHSHazardClass)
	fillStr(&dst.SignalWord, src.SignalWord)
	fillStr(&dst.PhysicalState, src.PhysicalState)
	fillStr(&dst.Color, src.Color)
	fillStr(&dst.Odor, src.Odor)
	fillFloat(&dst.PH, src.PH)
	fillFloat(&dst.MeltingPoint, src.MeltingPoint)
	fillFloat(&dst.BoilingPoint, src.BoilingPoint)
	fillFloat(&dst.FlashPoint, src.FlashPoint)
	fillStr(&dst.Solubility, src.Solubility)
	fillStr(&dst.StorageClass, src.StorageClass)
	fillStr(&dst.StorageRequirements, src.StorageRequirements)
	fillStr(&dst.RespiratoryProtection, src.RespiratoryProtection)
	fillStr(&dst.HandProtection, src.HandProtection)
	fillStr(&dst.EyeProtection, src.EyeProtection)
	fillStr(&dst.SkinProtection, src.SkinProtection)
	fillStr(&dst.FirstAidMeasures, src.FirstAidMeasures)
	fillStr(&dst.FirefightingMeasures, src.FirefightingMeasures)
	fillStr(&dst.SpillLeakProcedures, src.SpillLeakProcedures)
	fillStr(&dst.DisposalConsiderations, src.DisposalConsiderations)
	fillStr(&dst.ManufacturerName, src.ManufacturerName)
	fillStr(&dst.SuggestedVendor, src.SuggestedVendor)
	fillStr(&dst.SuggestedCategory, src.SuggestedCategory)
}
