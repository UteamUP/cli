package models

import (
	"fmt"
	"strings"
)

// ExtractedData is a wrapper holding at most one entity-specific extraction result.
type ExtractedData struct {
	Asset    *ExtractedAssetData    `json:"asset,omitempty"`
	Tool     *ExtractedToolData     `json:"tool,omitempty"`
	Part     *ExtractedPartData     `json:"part,omitempty"`
	Chemical *ExtractedChemicalData `json:"chemical,omitempty"`
}

// Type returns which entity type is populated, or EntityTypeUnclassified if none.
func (e *ExtractedData) Type() EntityType {
	switch {
	case e.Asset != nil:
		return EntityTypeAsset
	case e.Tool != nil:
		return EntityTypeTool
	case e.Part != nil:
		return EntityTypePart
	case e.Chemical != nil:
		return EntityTypeChemical
	default:
		return EntityTypeUnclassified
	}
}

// GetName returns the name from whichever entity is populated.
func (e *ExtractedData) GetName() string {
	switch {
	case e.Asset != nil:
		return e.Asset.Name
	case e.Tool != nil:
		return e.Tool.Name
	case e.Part != nil:
		return e.Part.Name
	case e.Chemical != nil:
		return e.Chemical.Name
	default:
		return ""
	}
}

// GetSerialNumber returns the serial number from whichever entity is populated.
func (e *ExtractedData) GetSerialNumber() string {
	switch {
	case e.Asset != nil:
		return ptrStr(e.Asset.SerialNumber)
	case e.Tool != nil:
		return ptrStr(e.Tool.SerialNumber)
	case e.Part != nil:
		return ptrStr(e.Part.SerialNumber)
	default:
		return ""
	}
}

// GetModelNumber returns the model number from whichever entity is populated.
func (e *ExtractedData) GetModelNumber() string {
	switch {
	case e.Asset != nil:
		return ptrStr(e.Asset.ModelNumber)
	case e.Tool != nil:
		return ptrStr(e.Tool.ModelNumber)
	case e.Part != nil:
		return ptrStr(e.Part.ModelNumber)
	default:
		return ""
	}
}

// GetDescription returns the description from whichever entity is populated.
func (e *ExtractedData) GetDescription() string {
	switch {
	case e.Asset != nil:
		return ptrStr(e.Asset.Description)
	case e.Tool != nil:
		return ptrStr(e.Tool.Description)
	case e.Part != nil:
		return ptrStr(e.Part.Description)
	case e.Chemical != nil:
		return ptrStr(e.Chemical.Description)
	default:
		return ""
	}
}

// GetVendorName returns the suggested vendor name from whichever entity is populated.
// Falls back to manufacturer/brand if no vendor is specified.
func (e *ExtractedData) GetVendorName() string {
	switch {
	case e.Asset != nil:
		if v := ptrStr(e.Asset.SuggestedVendor); v != "" {
			return v
		}
		return ptrStr(e.Asset.ManufacturerBrand)
	case e.Tool != nil:
		if v := ptrStr(e.Tool.SuggestedVendor); v != "" {
			return v
		}
		return ptrStr(e.Tool.ManufacturerBrand)
	case e.Part != nil:
		if v := ptrStr(e.Part.SuggestedVendor); v != "" {
			return v
		}
		return ptrStr(e.Part.ManufacturerBrand)
	case e.Chemical != nil:
		if v := ptrStr(e.Chemical.SuggestedVendor); v != "" {
			return v
		}
		return ptrStr(e.Chemical.ManufacturerName)
	default:
		return ""
	}
}

// GetLocationName returns the suggested location from whichever entity is populated.
func (e *ExtractedData) GetLocationName() string {
	switch {
	case e.Asset != nil:
		return ptrStr(e.Asset.SuggestedLocation)
	default:
		return ""
	}
}

// GetBrand returns the manufacturer/brand from whichever entity is populated.
func (e *ExtractedData) GetBrand() string {
	switch {
	case e.Asset != nil:
		return ptrStr(e.Asset.ManufacturerBrand)
	case e.Tool != nil:
		return ptrStr(e.Tool.ManufacturerBrand)
	case e.Part != nil:
		return ptrStr(e.Part.ManufacturerBrand)
	case e.Chemical != nil:
		return ptrStr(e.Chemical.ManufacturerName)
	default:
		return ""
	}
}

// ptrStr safely dereferences a *string, returning "" if nil.
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ptrFloat safely formats a *float64, returning "" if nil.
func ptrFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%g", *f)
}

// ptrBool safely formats a *bool, returning "" if nil.
func ptrBool(b *bool) string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%t", *b)
}

// --- ExtractedAssetData ---

// ExtractedAssetData mirrors GeneratedAssetData from the backend.
type ExtractedAssetData struct {
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	SerialNumber       *string `json:"serial_number,omitempty"`
	ReferenceNumber    *string `json:"reference_number,omitempty"`
	ModelNumber        *string `json:"model_number,omitempty"`
	UPCNumber          *string `json:"upc_number,omitempty"`
	AdditionalInfo     *string `json:"additional_info,omitempty"`
	Notes              *string `json:"notes,omitempty"`
	CheckInProcedure   *string `json:"check_in_procedure,omitempty"`
	CheckOutProcedure  *string `json:"check_out_procedure,omitempty"`
	IconName           *string `json:"icon_name,omitempty"`
	SuggestedVendor    *string `json:"suggested_vendor,omitempty"`
	SuggestedCategory  *string `json:"suggested_category,omitempty"`
	SuggestedLocation  *string `json:"suggested_location,omitempty"`
	ManufacturerBrand  *string `json:"manufacturer_brand,omitempty"`
	VisibleCondition   *string `json:"visible_condition,omitempty"`
	IsVehicle          *bool   `json:"is_vehicle,omitempty"`
	VehicleType        *string `json:"vehicle_type,omitempty"`
	LicensePlate       *string `json:"license_plate,omitempty"`
	AssetCategoryGroup *string `json:"asset_category_group,omitempty"`
}

// ToMap converts all fields to a string map for CSV export.
func (a *ExtractedAssetData) ToMap() map[string]string {
	return map[string]string{
		"name":                a.Name,
		"description":         ptrStr(a.Description),
		"serial_number":       ptrStr(a.SerialNumber),
		"reference_number":    ptrStr(a.ReferenceNumber),
		"model_number":        ptrStr(a.ModelNumber),
		"upc_number":          ptrStr(a.UPCNumber),
		"additional_info":     ptrStr(a.AdditionalInfo),
		"notes":               ptrStr(a.Notes),
		"check_in_procedure":  ptrStr(a.CheckInProcedure),
		"check_out_procedure": ptrStr(a.CheckOutProcedure),
		"icon_name":           ptrStr(a.IconName),
		"suggested_vendor":    ptrStr(a.SuggestedVendor),
		"suggested_category":  ptrStr(a.SuggestedCategory),
		"suggested_location":  ptrStr(a.SuggestedLocation),
		"manufacturer_brand":  ptrStr(a.ManufacturerBrand),
		"visible_condition":   ptrStr(a.VisibleCondition),
		"is_vehicle":          ptrBool(a.IsVehicle),
		"vehicle_type":        ptrStr(a.VehicleType),
		"license_plate":       ptrStr(a.LicensePlate),
	}
}

// --- ExtractedToolData ---

// ExtractedToolData mirrors GeneratedToolData from the backend.
type ExtractedToolData struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	SerialNumber      *string  `json:"serial_number,omitempty"`
	ReferenceNumber   *string  `json:"reference_number,omitempty"`
	ModelNumber       *string  `json:"model_number,omitempty"`
	BarcodeNumber     *string  `json:"barcode_number,omitempty"`
	ToolNumber        *string  `json:"tool_number,omitempty"`
	AdditionalInfo    *string  `json:"additional_info,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	SuggestedVendor   *string  `json:"suggested_vendor,omitempty"`
	SuggestedCategory *string  `json:"suggested_category,omitempty"`
	ManufacturerBrand *string  `json:"manufacturer_brand,omitempty"`
	Width             *float64 `json:"width,omitempty"`
	Height            *float64 `json:"height,omitempty"`
	Length            *float64 `json:"length,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Weight            *float64 `json:"weight,omitempty"`
	Value             *float64 `json:"value,omitempty"`
}

// ToMap converts all fields to a string map for CSV export.
func (t *ExtractedToolData) ToMap() map[string]string {
	return map[string]string{
		"name":               t.Name,
		"description":        ptrStr(t.Description),
		"width":              ptrFloat(t.Width),
		"height":             ptrFloat(t.Height),
		"length":             ptrFloat(t.Length),
		"depth":              ptrFloat(t.Depth),
		"weight":             ptrFloat(t.Weight),
		"value":              ptrFloat(t.Value),
		"barcode_number":     ptrStr(t.BarcodeNumber),
		"serial_number":      ptrStr(t.SerialNumber),
		"reference_number":   ptrStr(t.ReferenceNumber),
		"model_number":       ptrStr(t.ModelNumber),
		"tool_number":        ptrStr(t.ToolNumber),
		"additional_info":    ptrStr(t.AdditionalInfo),
		"notes":              ptrStr(t.Notes),
		"suggested_vendor":   ptrStr(t.SuggestedVendor),
		"suggested_category": ptrStr(t.SuggestedCategory),
		"manufacturer_brand": ptrStr(t.ManufacturerBrand),
	}
}

// --- ExtractedPartData ---

// ExtractedPartData mirrors GeneratedPartData from the backend.
type ExtractedPartData struct {
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	SerialNumber      *string  `json:"serial_number,omitempty"`
	ReferenceNumber   *string  `json:"reference_number,omitempty"`
	ModelNumber       *string  `json:"model_number,omitempty"`
	PartNumber        *string  `json:"part_number,omitempty"`
	AdditionalInfo    *string  `json:"additional_info,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	Value             *float64 `json:"value,omitempty"`
	SuggestedVendor   *string  `json:"suggested_vendor,omitempty"`
	SuggestedCategory *string  `json:"suggested_category,omitempty"`
	ManufacturerBrand *string  `json:"manufacturer_brand,omitempty"`
}

// ToMap converts all fields to a string map for CSV export.
func (p *ExtractedPartData) ToMap() map[string]string {
	return map[string]string{
		"name":               p.Name,
		"description":        ptrStr(p.Description),
		"serial_number":      ptrStr(p.SerialNumber),
		"reference_number":   ptrStr(p.ReferenceNumber),
		"model_number":       ptrStr(p.ModelNumber),
		"part_number":        ptrStr(p.PartNumber),
		"additional_info":    ptrStr(p.AdditionalInfo),
		"notes":              ptrStr(p.Notes),
		"value":              ptrFloat(p.Value),
		"suggested_vendor":   ptrStr(p.SuggestedVendor),
		"suggested_category": ptrStr(p.SuggestedCategory),
		"manufacturer_brand": ptrStr(p.ManufacturerBrand),
	}
}

// --- ExtractedChemicalData ---

// ExtractedChemicalData mirrors GeneratedChemicalData from the backend.
type ExtractedChemicalData struct {
	Name                    string   `json:"name"`
	Description             *string  `json:"description,omitempty"`
	ChemicalFormula         *string  `json:"chemical_formula,omitempty"`
	CASNumber               *string  `json:"cas_number,omitempty"`
	ECNumber                *string  `json:"ec_number,omitempty"`
	UNNumber                *string  `json:"un_number,omitempty"`
	GHSHazardClass          *string  `json:"ghs_hazard_class,omitempty"`
	SignalWord              *string  `json:"signal_word,omitempty"`
	PhysicalState           *string  `json:"physical_state,omitempty"`
	Color                   *string  `json:"color,omitempty"`
	Odor                    *string  `json:"odor,omitempty"`
	PH                      *float64 `json:"ph,omitempty"`
	MeltingPoint            *float64 `json:"melting_point,omitempty"`
	BoilingPoint            *float64 `json:"boiling_point,omitempty"`
	FlashPoint              *float64 `json:"flash_point,omitempty"`
	Solubility              *string  `json:"solubility,omitempty"`
	StorageClass            *string  `json:"storage_class,omitempty"`
	StorageRequirements     *string  `json:"storage_requirements,omitempty"`
	RespiratoryProtection   *string  `json:"respiratory_protection,omitempty"`
	HandProtection          *string  `json:"hand_protection,omitempty"`
	EyeProtection           *string  `json:"eye_protection,omitempty"`
	SkinProtection          *string  `json:"skin_protection,omitempty"`
	FirstAidMeasures        *string  `json:"first_aid_measures,omitempty"`
	FirefightingMeasures    *string  `json:"firefighting_measures,omitempty"`
	SpillLeakProcedures     *string  `json:"spill_leak_procedures,omitempty"`
	DisposalConsiderations  *string  `json:"disposal_considerations,omitempty"`
	UnitOfMeasure           string   `json:"unit_of_measure"`
	HazardStatements        []string `json:"hazard_statements"`
	PrecautionaryStatements []string `json:"precautionary_statements"`
	ManufacturerName        *string  `json:"manufacturer_name,omitempty"`
	SuggestedVendor         *string  `json:"suggested_vendor,omitempty"`
	SuggestedCategory       *string  `json:"suggested_category,omitempty"`
}

// ToMap converts all fields to a string map for CSV export.
func (c *ExtractedChemicalData) ToMap() map[string]string {
	return map[string]string{
		"name":                     c.Name,
		"description":              ptrStr(c.Description),
		"chemical_formula":         ptrStr(c.ChemicalFormula),
		"cas_number":               ptrStr(c.CASNumber),
		"ec_number":                ptrStr(c.ECNumber),
		"un_number":                ptrStr(c.UNNumber),
		"ghs_hazard_class":         ptrStr(c.GHSHazardClass),
		"signal_word":              ptrStr(c.SignalWord),
		"physical_state":           ptrStr(c.PhysicalState),
		"color":                    ptrStr(c.Color),
		"odor":                     ptrStr(c.Odor),
		"ph":                       ptrFloat(c.PH),
		"melting_point":            ptrFloat(c.MeltingPoint),
		"boiling_point":            ptrFloat(c.BoilingPoint),
		"flash_point":              ptrFloat(c.FlashPoint),
		"solubility":               ptrStr(c.Solubility),
		"storage_class":            ptrStr(c.StorageClass),
		"storage_requirements":     ptrStr(c.StorageRequirements),
		"respiratory_protection":   ptrStr(c.RespiratoryProtection),
		"hand_protection":          ptrStr(c.HandProtection),
		"eye_protection":           ptrStr(c.EyeProtection),
		"skin_protection":          ptrStr(c.SkinProtection),
		"first_aid_measures":       ptrStr(c.FirstAidMeasures),
		"firefighting_measures":    ptrStr(c.FirefightingMeasures),
		"spill_leak_procedures":    ptrStr(c.SpillLeakProcedures),
		"disposal_considerations":  ptrStr(c.DisposalConsiderations),
		"unit_of_measure":          c.UnitOfMeasure,
		"hazard_statements":        strings.Join(c.HazardStatements, "; "),
		"precautionary_statements": strings.Join(c.PrecautionaryStatements, "; "),
		"manufacturer_name":        ptrStr(c.ManufacturerName),
		"suggested_vendor":         ptrStr(c.SuggestedVendor),
		"suggested_category":       ptrStr(c.SuggestedCategory),
	}
}
