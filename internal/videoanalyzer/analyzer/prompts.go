// Package analyzer provides Gemini Vision AI analysis for video-based inventory classification.
package analyzer

// VideoAnalysisPrompt is the prompt sent to Gemini for video-based CMMS entity extraction.
// It instructs the model to watch the entire video, identify all CMMS entities visible at
// any point, and report the timestamp when each entity first appears.
const VideoAnalysisPrompt = `You are an expert inventory analyst for a CMMS (Computerized Maintenance Management System). Watch this entire video carefully and identify ALL distinct physical entities visible at any point.

## YOUR TASK

1. Watch the full video from start to finish.
2. Identify every distinct physical entity (equipment, tools, parts, chemicals) visible.
3. For each entity, record the TIMESTAMP (MM:SS) when it FIRST becomes clearly visible.
4. Classify each entity and extract detailed data.
5. If the same physical item appears at multiple points in the video, report it ONLY ONCE with its earliest timestamp.
6. Look for nameplates, labels, serial numbers, barcodes, brand logos, and any visible text.
7. Assess the visible condition of each entity (Good, Fair, Poor).

## ENTITY TYPES

Classify each entity as exactly one of:
- "asset" — Equipment, machinery, vehicles, fixed installations, large systems
- "tool" — Handheld tools, power tools, specialized equipment, measuring instruments
- "part" — Components, spare parts, consumables, replacement items
- "chemical" — Chemicals, lubricants, cleaning agents, hazardous materials, liquids with labels
- "unclassified" — Cannot determine type with reasonable confidence

## RELATIONSHIPS

If an entity is clearly PART OF or ATTACHED TO another entity (e.g., a filter attached to a compressor), set the ` + "`related_to`" + ` field to the name of the parent entity. This helps build the asset hierarchy.

## CRITICAL RULES

- Report ONLY entities you can CLEARLY see in the video. Do not guess or infer hidden items.
- Each physical item should appear ONLY ONCE even if visible at multiple timestamps.
- Use the EARLIEST timestamp when the entity is first clearly identifiable.
- If you cannot read a serial number fully, include what you CAN read and flag for review.
- If confidence is below 0.5, set ` + "`flagged_for_review`" + ` to true.
- Timestamps MUST be in MM:SS format (e.g., "00:15", "01:30", "10:05").
- If a location or address is visible in the video (signs, labels, GPS overlays), include it in ` + "`suggested_location`" + `.

## RESPONSE FORMAT

Return a JSON object with an ` + "`entities`" + ` array. Each element must have:

` + "```" + `json
{
  "entities": [
    {
      "type": "asset|tool|part|chemical|unclassified",
      "timestamp": "MM:SS",
      "confidence": 0.0 to 1.0,
      "reasoning": "brief explanation of classification",
      "flagged_for_review": false,
      "review_reason": "reason if flagged, null otherwise",
      "related_to": "parent entity name or null",
      "extracted_data": { ... }
    }
  ]
}
` + "```" + `

## ENTITY-SPECIFIC FIELD SCHEMAS

### For type = "asset":
` + "```" + `
{
  "name": "<descriptive name of the asset>",
  "description": "<what the asset is and its visible characteristics>",
  "serial_number": "<from nameplate/label or null>",
  "reference_number": "<reference/asset tag number or null>",
  "model_number": "<from nameplate/label or null>",
  "upc_number": "<UPC/barcode number if visible or null>",
  "additional_info": "<any extra visible info not fitting other fields>",
  "notes": "<observations about condition, environment, installation>",
  "check_in_procedure": null,
  "check_out_procedure": null,
  "icon_name": "<suggest a Material Design icon name>",
  "suggested_vendor": "<manufacturer/brand name if visible>",
  "suggested_category": "<one of: General, VehicleAndFleet, EnergyAndPower, HVAC, Plumbing, Electrical, Safety, Manufacturing, IT, Facilities, Medical, Laboratory, Agriculture, Construction, Warehouse>",
  "suggested_location": "<location hint if visible on labels or in video>",
  "manufacturer_brand": "<brand/manufacturer from logo or nameplate>",
  "visible_condition": "<Good, Fair, Poor, or null if unclear>",
  "is_vehicle": false,
  "vehicle_type": "<Car, Truck, Van, Forklift, Trailer, Bus, Motorcycle, Boat, or null>",
  "license_plate": "<license plate text or null>",
  "asset_category_group": "<same categories as suggested_category>"
}
` + "```" + `

### For type = "tool":
` + "```" + `
{
  "name": "<descriptive name of the tool>",
  "description": "<what the tool is and its visible characteristics>",
  "width": null,
  "height": null,
  "length": null,
  "depth": null,
  "weight": null,
  "value": null,
  "barcode_number": "<barcode if visible or null>",
  "serial_number": "<from label or null>",
  "reference_number": "<reference number or null>",
  "model_number": "<from label or null>",
  "tool_number": "<tool number/ID if visible or null>",
  "additional_info": "<extra visible info>",
  "notes": "<observations about condition, wear>",
  "suggested_vendor": "<manufacturer/brand if visible>",
  "suggested_category": "<Hand Tools, Power Tools, Measuring, Cutting, Welding, Pneumatic, Hydraulic, Electrical, Safety, Cleaning, Automotive, Garden>",
  "manufacturer_brand": "<brand from markings>"
}
` + "```" + `

### For type = "part":
` + "```" + `
{
  "name": "<descriptive name of the part>",
  "description": "<what the part is and its visible characteristics>",
  "serial_number": "<from label or null>",
  "reference_number": "<reference number or null>",
  "model_number": "<from label or null>",
  "part_number": "<part number from packaging/label or null>",
  "additional_info": "<extra visible info>",
  "notes": "<observations about condition, packaging>",
  "value": null,
  "suggested_vendor": "<manufacturer/brand if visible>",
  "suggested_category": "<Filters, Belts, Bearings, Gaskets, Fasteners, Electrical, Hydraulic, Pneumatic, Seals, Valves, Gears>",
  "manufacturer_brand": "<brand from packaging/markings>"
}
` + "```" + `

### For type = "chemical":
` + "```" + `
{
  "name": "<product/chemical name>",
  "description": "<what the chemical is>",
  "chemical_formula": "<if visible on label or null>",
  "cas_number": "<CAS number if visible or null>",
  "ec_number": "<EC number if visible or null>",
  "un_number": "<UN number if visible or null>",
  "ghs_hazard_class": "<GHS class if visible or null>",
  "signal_word": "<Danger or Warning if visible or null>",
  "physical_state": "<Solid, Liquid, Gas, Aerosol, or null>",
  "color": "<visible color or null>",
  "odor": null,
  "ph": null,
  "melting_point": null,
  "boiling_point": null,
  "flash_point": null,
  "solubility": null,
  "storage_class": "<storage class if visible or null>",
  "storage_requirements": "<from label or null>",
  "respiratory_protection": "<from SDS/label or null>",
  "hand_protection": "<from SDS/label or null>",
  "eye_protection": "<from SDS/label or null>",
  "skin_protection": "<from SDS/label or null>",
  "first_aid_measures": "<from label or null>",
  "firefighting_measures": "<from label or null>",
  "spill_leak_procedures": "<from label or null>",
  "disposal_considerations": "<from label or null>",
  "unit_of_measure": "<L, mL, kg, g, oz, gal, or null>",
  "hazard_statements": "<H-statements from label or null>",
  "precautionary_statements": "<P-statements from label or null>",
  "manufacturer_name": "<manufacturer if visible>",
  "suggested_vendor": "<vendor/distributor if visible>",
  "suggested_category": "<Lubricants, Cleaning, Adhesives, Sealants, Paints, Solvents, Acids, Bases, Fuels, Coolants, Refrigerants, Gases>"
}
` + "```" + `

### For type = "unclassified":
Set ` + "`extracted_data`" + ` to ` + "`null`" + `.

Watch the entire video now and return ONLY the JSON object with the ` + "`entities`" + ` array. No markdown, no fences, no extra text.`

// VendorEnrichmentPrompt is the prompt sent to Gemini to enrich a detected vendor name
// with additional information. Use fmt.Sprintf(VendorEnrichmentPrompt, vendorName).
const VendorEnrichmentPrompt = `Given this manufacturer/vendor name: "%s"

Provide the following information about this company. If you are not confident about any field, set it to null.

Return ONLY a JSON object:
{
  "full_name": "<official full legal company name>",
  "website": "<official website URL>",
  "business_category": "<primary business category: Manufacturing, Industrial Supply, Tools, Electronics, Automotive, Chemical, Safety Equipment, HVAC, Plumbing, Electrical, General Supply, Other>",
  "country": "<country of headquarters>",
  "confidence": 0.0 to 1.0
}

No markdown, no fences, no extra text. Only the JSON object.`
