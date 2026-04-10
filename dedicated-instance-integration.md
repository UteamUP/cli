# CLI: Dedicated Instance Integration

## Changes Required

### 1. New Domain Registry: `internal/registry/domains_dedicated_instance.go`

```go
package registry

func init() {
    RegisterDomain(Domain{
        Name:        "dedicated",
        Description: "Manage dedicated API/frontend instances",
        Commands: []Command{
            {Name: "status", Description: "Get dedicated instance status", Tool: "get_dedicated_instance_status"},
            {Name: "billing", Description: "Get billing breakdown", Tool: "get_dedicated_instance_billing"},
            {Name: "provision", Description: "Provision dedicated instances", Tool: "provision_dedicated_instance"},
            {Name: "scale", Description: "Scale dedicated instances", Tool: "scale_dedicated_instance"},
            {Name: "decommission", Description: "Decommission dedicated instances", Tool: "decommission_dedicated_instance"},
            {Name: "waf", Description: "Toggle WAF Premium", Tool: "toggle_waf_premium"},
            {Name: "migrate", Description: "Schedule data migration", Tool: "schedule_migration"},
        },
    })
}
```

### 2. Tenant Profile: `~/.uteamup/config.json`

Add `dedicated_api_url` to profile:
```json
{
  "profiles": {
    "default": {
      "api_base_url": "https://api.uteamup.com",
      "dedicated_api_url": "https://acme.api.uteamup.com/api"
    }
  }
}
```

### 3. Dynamic URL in Client

In `internal/client/client.go`, resolve URL per request:
```go
func (c *Client) getBaseURL() string {
    if c.config.DedicatedApiUrl != "" {
        return c.config.DedicatedApiUrl
    }
    return c.config.ApiBaseUrl
}
```
