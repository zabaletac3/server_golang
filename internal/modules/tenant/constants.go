package tenant

type TenantStatus string

const (
	Active    TenantStatus = "active"
	Inactive  TenantStatus = "inactive"
	Trial     TenantStatus = "trial"
	Suspended TenantStatus = "suspended"
)