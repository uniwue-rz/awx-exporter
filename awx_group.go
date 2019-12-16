package main

import "time"

type Group struct {
	ID                      int                `json:"id"`
	Type                    string             `json:"type"`
	Url                     string             `json:"url"`
	Related                 GroupRelated       `json:"related"`
	SummaryFields           GroupSummaryFields `json:"summary_fields"`
	Created                 time.Time          `json:"created,string"`
	Modified                time.Time          `json:"modified,string"`
	Name                    string             `json:"name"`
	Description             string             `json:"description"`
	Inventory               int                `json:"inventory"`
	Variables               string             `json:"variables"`
	HasActiveFailures       bool               `json:"has_active_failures"`
	TotalHosts              int                `json:"total_hosts"`
	HostsWithActiveFailures int                `json:"hosts_with_active_failures"`
	TotalGroups             int                `json:"total_groups"`
	GroupsWithFailures      int                `json:"groups_with_failures"`
	HasInventorySources     bool               `json:"has_inventory_sources"`
}

type GroupRelated struct {
	VariableData      string `json:"variable_data"`
	Hosts             string `json:"hosts"`
	PotentialChildren string `json:"potential_children"`
	Children          string `json:"children"`
	AllHosts          string `json:"all_hosts"`
	JobEvents         string `json:"job_events"`
	JobHostSummaries  string `json:"job_host_summaries"`
	ActivityStream    string `json:"activity_stream"`
	AdHocCommands     string `json:"ad_hoc_commands"`
	Inventory         string `json:"inventory"`
}

type InventorySummary struct {
	ID                           int    `json:"id"`
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	HasActiveFailures            bool   `json:"has_active_failures"`
	TotalHosts                   int    `json:"total_hosts"`
	TotalGroups                  int    `json:"total_groups"`
	GroupsWithActiveFailures     int    `json:"groups_with_active_failures"`
	HasInventorySources          bool   `json:"has_inventory_sources"`
	TotalInventorySources        int    `json:"total_inventory_sources"`
	InventorySourcesWithFailures int    `json:"inventory_sources_with_failures"`
	OrganizationId               int    `json:"organization_id"`
	Kind                         string `json:"kind"`
}

type GroupUserCapabilities struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
	Copy   bool `json:"copy"`
}

type GroupSummaryFields struct {
	Inventory        InventorySummary      `json:"inventory"`
	UserCapabilities GroupUserCapabilities `json:"user_capabilities"`
}

type GroupResults struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous string  `json:"previous"`
	Results  []Group `json:"results"`
}

type GroupSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
