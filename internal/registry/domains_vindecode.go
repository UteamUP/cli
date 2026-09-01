package registry

func init() {
	// Routes mirror VinDecodeController (/api/vindecode/{vin}). Gated server-side on
	// Fleet.Read plus the Fleet VIN Decode feature; the decode provider base URL is
	// deployment configuration (Fleet:VinDecode:BaseUrl).
	Register(&Domain{Name: "vindecode", Aliases: []string{"vin"}, Description: "Decode vehicle identification numbers", APIPath: "/api/vindecode", Actions: []Action{
		{Name: "decode", HTTPMethod: "GET", Description: "Decode a 17-character VIN", ToolName: "UteamupFleetVinDecode", Args: []ArgDef{
			{Name: "vin", Description: "17-character VIN (no I, O or Q)", Required: true, Type: "string"},
		}, RESTPath: "{vin}"},
	}})
}
