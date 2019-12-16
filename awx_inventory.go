package main

import "time"

type InventoryRelated struct {
	CreatedBy              string `json:"created_by"`
	ModifiedBy             string `json:"modified_by"`
	Hosts                  string `json:"hosts"`
	Groups                 string `json:"groups"`
	RootGroup              string `json:"root_group"`
	VariableData           string `json:"variable_data"`
	Script                 string `json:"script"`
	Tree                   string `json:"tree"`
	InventorySources       string `json:"inventory_sources"`
	UpdateInventorySources string `json:"update_inventory_sources"`
	ActivityStream         string `json:"activity_stream"`
	JobTemplates           string `json:"job_templates"`
	AdHocCommands          string `json:"ad_hoc_commands"`
	AccessList             string `json:"access_list"`
	ObjectRoles            string `json:"object_roles"`
	InstanceGroups         string `json:"instance_groups"`
	Copy                   string `json:"copy"`
	Organization           string `json:"organization"`
}

type InventorySummaryFields struct {
	Organization     OrganizationSummary       `json:"organization"`
	CreatedBy        PersonSummary             `json:"created_by"`
	ModifiedBy       PersonSummary             `json:"modified_by"`
	ObjectRoles      ObjectRoles               `json:"object_roles"`
	UserCapabilities InventoryUserCapabilities `json:"user_capabilities"`
}

type ObjectRoles struct {
	AdminRole  RoleSummary `json:"admin_role"`
	UpdateRole RoleSummary `json:"update_role"`
	AdhocRole  RoleSummary `json:"adhoc_role"`
	UseRole    RoleSummary `json:"use_role"`
	ReadRole   RoleSummary `json:"read_role"`
}

type OrganizationSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PersonSummary struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type RoleSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type InventoryUserCapabilities struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
	Copy   bool `json:"copy"`
	Adhoc  bool `json:"adhoc"`
}

type Inventory struct {
	ID                       int                    `json:"id"`
	Type                     string                 `json:"type"`
	Url                      string                 `json:"url"`
	Related                  InventoryRelated       `json:"related"`
	SummaryFields            InventorySummaryFields `json:"summary_fields"`
	Created                  time.Time              `json:"created,string"`
	Modified                 time.Time              `json:"modified,string"`
	Name                     string                 `json:"name"`
	Description              string                 `json:"description"`
	Organization             int                    `json:"organization"`
	Kind                     string                 `json:"kind"`
	HostFilter               string                 `json:"host_filter"`
	Variables                string                 `json:"variables"`
	HasActiveFailures        bool                   `json:"has_active_failures"`
	HostsWithActiveFailures  int                    `json:"hosts_with_active_failures"`
	TotalGroups              int                    `json:"total_groups"`
	GroupsWithActiveFailures int                    `json:"groups_with_active_failures"`
	HasInventorySources      bool                   `json:"has_inventory_sources"`
	InsightsCredential       string                 `json:"insights_credential"`
	PendingDeletion          bool                   `json:"pending_deletion"`
}

type InventoryResult struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Results  []Inventory `json:"results"`
}
