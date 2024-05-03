package bonsai

// Release is a placeholder for now.
type Release struct {
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug"`
	ServiceType string `json:"service_type,omitempty"`
	Version     string `json:"version,omitempty"`
	MultiTenant bool   `json:"multi_tenant,omitempty"`
}
