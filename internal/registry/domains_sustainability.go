package registry

// Sustainability readings, waste, carbon emissions, emission factors, energy
// benchmarks, goals and energy anomalies. The backend MCP tools apply the same
// permissions as the REST controllers and the Sustainability feature gate.
func init() {
	page := []FlagDef{
		{Name: "page", Description: "Page number, starting at 1", Type: "int", Default: 1},
		{Name: "page-size", Description: "Records per page (1-100)", Type: "int", Default: 25},
	}
	guidArg := func(name, body, what string) []ArgDef {
		return []ArgDef{{Name: name, BodyName: body, Description: what + " public GUID", Required: true, Type: "non-empty-uuid"}}
	}
	withPage := func(flags ...FlagDef) []FlagDef { return append(flags, page...) }

	energyFlags := func(create bool) []FlagDef {
		return []FlagDef{
			{Name: "utility-meter-guid", Description: "Utility meter public GUID", Required: create, Type: "non-empty-uuid"},
			{Name: "reading-value", Description: "Meter value read", Required: create, Type: "decimal"},
			{Name: "reading-date", Description: "Reading date and time (ISO 8601, UTC)", Required: create, Type: "string"},
			{Name: "consumption", Description: "Consumption since the previous reading", Type: "decimal"},
			{Name: "reading-type", Description: "Manual, Automated or Estimated", Type: "string"},
			{Name: "utility-type", Description: "Electricity, NaturalGas, Water, Steam, CompressedAir, Diesel, Propane or FuelOil", Type: "string"},
			{Name: "unit", Description: "Unit, for example kWh", Type: "string"},
			{Name: "cost-per-unit", Description: "Cost per unit", Type: "decimal"},
			{Name: "total-cost", Description: "Total cost", Type: "decimal"},
			{Name: "notes", Description: "Notes", Type: "string"},
			{Name: "source-system", Description: "Source system name", Type: "string"},
			{Name: "external-reference-id", Description: "Reference in the source system", Type: "string"},
		}
	}
	waterFlags := func(create bool) []FlagDef {
		return []FlagDef{
			{Name: "location-guid", Description: "Location public GUID", Required: create, Type: "non-empty-uuid"},
			{Name: "reading-date", Description: "Reading date (ISO 8601, UTC)", Required: create, Type: "string"},
			{Name: "consumption", Description: "Water consumed", Required: create, Type: "decimal"},
			{Name: "utility-meter-guid", Description: "Utility meter public GUID", Type: "non-empty-uuid"},
			{Name: "unit", Description: "Unit, for example m3", Type: "string"},
			{Name: "water-source", Description: "Municipal, Well, Recycled, Rainwater or Surface", Type: "string"},
			{Name: "cost-per-unit", Description: "Cost per unit", Type: "decimal"},
			{Name: "total-cost", Description: "Total cost", Type: "decimal"},
			{Name: "notes", Description: "Notes", Type: "string"},
		}
	}
	wasteFlags := func(create bool) []FlagDef {
		return []FlagDef{
			{Name: "location-guid", Description: "Location public GUID", Required: create, Type: "non-empty-uuid"},
			{Name: "record-date", Description: "Record date (ISO 8601, UTC)", Required: create, Type: "string"},
			{Name: "waste-category", Description: "Waste category, for example Used oil", Required: create, Type: "string"},
			{Name: "quantity", Description: "Quantity", Required: create, Type: "decimal"},
			{Name: "waste-type", Description: "General, Hazardous, Recyclable, Electronic, Organic or Construction", Type: "string"},
			{Name: "disposal-method", Description: "Landfill, Recycled, Incinerated, Composted, Reused or HazardousDisposal", Type: "string"},
			{Name: "unit", Description: "Unit, for example kg", Type: "string"},
			{Name: "asset-guid", Description: "Asset public GUID", Type: "non-empty-uuid"},
			{Name: "workorder-guid", Description: "Work order public GUID", Type: "non-empty-uuid"},
			{Name: "manifest-number", Description: "Waste manifest number", Type: "string"},
			{Name: "regulatory-reference", Description: "Regulatory reference", Type: "string"},
			{Name: "cost", Description: "Disposal cost", Type: "decimal"},
			{Name: "notes", Description: "Notes", Type: "string"},
		}
	}
	carbonFlags := func(create bool) []FlagDef {
		return []FlagDef{
			{Name: "scope", Description: "Scope1, Scope2 or Scope3", Required: create, Type: "string"},
			{Name: "category", Description: "Category, for example Stationary combustion", Required: create, Type: "string"},
			{Name: "source-description", Description: "What produced the emission", Required: create, Type: "string"},
			{Name: "period-start", Description: "Period start (ISO 8601, UTC)", Required: create, Type: "string"},
			{Name: "period-end", Description: "Period end (ISO 8601, UTC)", Required: create, Type: "string"},
			{Name: "co2e-kg", Description: "CO2-equivalent in kg", Required: create, Type: "decimal"},
			{Name: "activity-data", Description: "Activity data, for example litres of fuel", Required: create, Type: "decimal"},
			{Name: "emission-factor-value", Description: "Emission factor value used", Required: create, Type: "decimal"},
			{Name: "activity-unit", Description: "Unit of the activity data", Type: "string"},
			{Name: "co2e-tonnes", Description: "CO2-equivalent in tonnes (derived from kg when left out)", Type: "decimal"},
			{Name: "co2-kg", Description: "CO2 in kg", Type: "decimal"},
			{Name: "ch4-kg", Description: "CH4 in kg", Type: "decimal"},
			{Name: "n2o-kg", Description: "N2O in kg", Type: "decimal"},
			{Name: "calculation-method", Description: "LocationBased, MarketBased or Direct", Type: "string"},
			{Name: "data-quality", Description: "Measured, Calculated, Estimated or Default", Type: "string"},
			{Name: "asset-guid", Description: "Asset public GUID", Type: "non-empty-uuid"},
			{Name: "notes", Description: "Notes", Type: "string"},
		}
	}

	Register(&Domain{
		Name:        "sustainability",
		Aliases:     []string{"esg"},
		Description: "Energy and water readings, waste, carbon emissions, goals and energy anomalies",
		Actions: []Action{
			{Name: "energy-reading-list", Description: "List energy readings, newest first", ToolName: "UteamupSustainabilityEnergyReadingList", MCPOnly: true, RejectExtraArgs: true,
				Flags: withPage(FlagDef{Name: "utility-meter-guid", Description: "Only this utility meter's readings", Type: "non-empty-uuid"})},
			{Name: "energy-reading-get", Description: "Get an energy reading", ToolName: "UteamupSustainabilityEnergyReadingGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("reading-guid", "readingGuid", "Energy reading")},
			{Name: "energy-reading-create", Description: "Record an energy reading", ToolName: "UteamupSustainabilityEnergyReadingCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: energyFlags(true)},
			{Name: "energy-reading-update", Description: "Correct an energy reading (only sent values change)", ToolName: "UteamupSustainabilityEnergyReadingUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("reading-guid", "readingGuid", "Energy reading"), Flags: energyFlags(false)},
			{Name: "water-reading-list", Description: "List water readings, newest first", ToolName: "UteamupSustainabilityWaterReadingList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "water-reading-get", Description: "Get a water reading", ToolName: "UteamupSustainabilityWaterReadingGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("reading-guid", "readingGuid", "Water reading")},
			{Name: "water-reading-create", Description: "Record a water reading", ToolName: "UteamupSustainabilityWaterReadingCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: waterFlags(true)},
			{Name: "water-reading-update", Description: "Correct a water reading (only sent values change)", ToolName: "UteamupSustainabilityWaterReadingUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("reading-guid", "readingGuid", "Water reading"), Flags: waterFlags(false)},
			{Name: "waste-record-list", Description: "List waste records, newest first", ToolName: "UteamupSustainabilityWasteRecordList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "waste-record-get", Description: "Get a waste record", ToolName: "UteamupSustainabilityWasteRecordGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("record-guid", "recordGuid", "Waste record")},
			{Name: "waste-record-create", Description: "Record disposed waste", ToolName: "UteamupSustainabilityWasteRecordCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: wasteFlags(true)},
			{Name: "waste-record-update", Description: "Correct a waste record (only sent values change)", ToolName: "UteamupSustainabilityWasteRecordUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("record-guid", "recordGuid", "Waste record"), Flags: wasteFlags(false)},
			{Name: "carbon-emission-list", Description: "List carbon emissions, newest period first", ToolName: "UteamupSustainabilityCarbonEmissionList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "carbon-emission-get", Description: "Get a carbon emission", ToolName: "UteamupSustainabilityCarbonEmissionGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("emission-guid", "emissionGuid", "Carbon emission")},
			{Name: "carbon-emission-create", Description: "Record a carbon emission", ToolName: "UteamupSustainabilityCarbonEmissionCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: carbonFlags(true)},
			{Name: "carbon-emission-update", Description: "Correct a carbon emission (only sent values change)", ToolName: "UteamupSustainabilityCarbonEmissionUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("emission-guid", "emissionGuid", "Carbon emission"), Flags: carbonFlags(false)},
			{Name: "emission-factor-list", Description: "List emission factors", ToolName: "UteamupSustainabilityEmissionFactorList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "emission-factor-get", Description: "Get an emission factor", ToolName: "UteamupSustainabilityEmissionFactorGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("emission-factor-guid", "emissionFactorGuid", "Emission factor")},
			{Name: "benchmark-list", Description: "List energy benchmarks", ToolName: "UteamupSustainabilityEnergyBenchmarkList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "benchmark-get", Description: "Get an energy benchmark", ToolName: "UteamupSustainabilityEnergyBenchmarkGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("benchmark-guid", "benchmarkGuid", "Energy benchmark")},
			{Name: "goal-list", Description: "List sustainability goals with progress", ToolName: "UteamupSustainabilityGoalList", MCPOnly: true, RejectExtraArgs: true,
				Flags: withPage(FlagDef{Name: "include-inactive", Description: "Include deactivated goals", Type: "bool"})},
			{Name: "goal-get", Description: "Get a sustainability goal with progress", ToolName: "UteamupSustainabilityGoalGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("goal-guid", "goalGuid", "Sustainability goal")},
			{Name: "anomaly-list", Description: "List energy anomalies, newest first", ToolName: "UteamupSustainabilityEnergyAnomalyList", MCPOnly: true, RejectExtraArgs: true,
				Flags: withPage(
					FlagDef{Name: "status", Description: "New, Acknowledged, Investigating, Resolved or FalsePositive", Type: "string"},
					FlagDef{Name: "severity", Description: "Low, Medium, High or Critical", Type: "string"},
					FlagDef{Name: "type", Description: "ReadingOutOfRange, UnusualPattern, FrequencyDeviation, MissingReading or SuddenChange", Type: "string"},
				)},
			{Name: "anomaly-get", Description: "Get an energy anomaly", ToolName: "UteamupSustainabilityEnergyAnomalyGet", MCPOnly: true, RejectExtraArgs: true,
				Args: guidArg("anomaly-guid", "anomalyGuid", "Energy anomaly")},
		},
	})
}
