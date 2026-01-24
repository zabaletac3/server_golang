package permissions

type Resource string

const (
	ResourceDashboard    Resource = "dashboard"
	ResourceUsers        Resource = "users"
	ResourceAppointments Resource = "appointments"
	ResourcePatients     Resource = "patients"
	ResourceOwners       Resource = "owners"
	ResourceInventory    Resource = "inventory"
	ResourceBilling      Resource = "billing"
	ResourceReports      Resource = "reports"
	ResourceSettings     Resource = "settings"
	ResourceRoles        Resource = "roles"
)

func (r Resource) IsValid() bool {
	switch r {
	case ResourceDashboard, ResourceUsers, ResourceAppointments,
		ResourcePatients, ResourceOwners, ResourceInventory,
		ResourceBilling, ResourceReports, ResourceSettings, ResourceRoles:
		return true
	}
	return false
}

// AllResources retorna todos los recursos disponibles
func AllResources() []Resource {
	return []Resource{
		ResourceDashboard, ResourceUsers, ResourceAppointments,
		ResourcePatients, ResourceOwners, ResourceInventory,
		ResourceBilling, ResourceReports, ResourceSettings, ResourceRoles,
	}
}

type Action string

const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionApprove Action = "approve"
	ActionExport  Action = "export"
	ActionManage  Action = "manage"
)

func (a Action) IsValid() bool {
	switch a {
	case ActionCreate, ActionRead, ActionUpdate,
		ActionDelete, ActionApprove, ActionExport, ActionManage:
		return true
	}
	return false
}

// AllActions retorna todas las acciones disponibles
func AllActions() []Action {
	return []Action{
		ActionCreate, ActionRead, ActionUpdate,
		ActionDelete, ActionApprove, ActionExport, ActionManage,
	}
}
