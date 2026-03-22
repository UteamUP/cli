package models

// AssetCSVColumns defines the CSV column order for asset exports.
var AssetCSVColumns = []string{
	"name", "description", "serial_number", "reference_number", "model_number",
	"upc_number", "additional_info", "notes", "check_in_procedure", "check_out_procedure",
	"icon_name", "suggested_vendor", "suggested_category", "suggested_location",
	"manufacturer_brand", "visible_condition", "is_vehicle", "vehicle_type", "license_plate",
	"related_to",
	"image_paths", "original_filenames", "confidence_score", "flagged_for_review", "review_reason",
}

// ToolCSVColumns defines the CSV column order for tool exports.
var ToolCSVColumns = []string{
	"name", "description", "width", "height", "length", "depth", "weight", "value",
	"barcode_number", "serial_number", "reference_number", "model_number", "tool_number",
	"additional_info", "notes", "suggested_vendor", "suggested_category",
	"manufacturer_brand", "related_to",
	"image_paths", "original_filenames", "confidence_score", "flagged_for_review", "review_reason",
}

// PartCSVColumns defines the CSV column order for part exports.
var PartCSVColumns = []string{
	"name", "description", "serial_number", "reference_number", "model_number",
	"part_number", "additional_info", "notes", "value",
	"suggested_vendor", "suggested_category", "manufacturer_brand", "related_to",
	"image_paths", "original_filenames", "confidence_score", "flagged_for_review", "review_reason",
}

// ChemicalCSVColumns defines the CSV column order for chemical exports.
var ChemicalCSVColumns = []string{
	"name", "description", "chemical_formula", "cas_number", "ec_number", "un_number",
	"ghs_hazard_class", "signal_word", "physical_state", "color", "odor",
	"ph", "melting_point", "boiling_point", "flash_point", "solubility",
	"storage_class", "storage_requirements",
	"respiratory_protection", "hand_protection", "eye_protection", "skin_protection",
	"first_aid_measures", "firefighting_measures", "spill_leak_procedures", "disposal_considerations",
	"unit_of_measure", "hazard_statements", "precautionary_statements",
	"manufacturer_name", "suggested_vendor", "suggested_category", "related_to",
	"image_paths", "original_filenames", "confidence_score", "flagged_for_review", "review_reason",
}

// UnclassifiedCSVColumns defines the CSV column order for unclassified exports.
var UnclassifiedCSVColumns = []string{
	"original_filename", "image_path", "confidence_score",
	"flagged_for_review", "review_reason", "classification_reasoning", "related_to",
}

// CSVColumnsByType maps each entity type to its CSV column list.
var CSVColumnsByType = map[EntityType][]string{
	EntityTypeAsset:        AssetCSVColumns,
	EntityTypeTool:         ToolCSVColumns,
	EntityTypePart:         PartCSVColumns,
	EntityTypeChemical:     ChemicalCSVColumns,
	EntityTypeUnclassified: UnclassifiedCSVColumns,
}
