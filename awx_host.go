package main

import "time"

type Host struct {
	ID                   int               `json:"id"`
	Type                 string            `json:"type"`
	Url                  string            `json:"url"`
	Related              HostRelated       `json:"related"`
	SummaryFields        HostSummaryFields `json:"summary_fields"`
	Created              time.Time         `json:"created,string"`
	Modified             time.Time         `json:"modified,string"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Inventory            int               `json:"inventory"`
	Enabled              bool              `json:"enabled"`
	InstanceId           string            `json:"instance_id"`
	Variables            string            `json:"variables"`
	HasActiveFailures    bool              `json:"has_active_failures"`
	HasInventorySources  bool              `json:"has_inventory_sources"`
	LastJob              int               `json:"last_job"`
	LastJobHostSummary   int               `json:"last_job_host_summary"`
	InsightSystemId      string            `json:"insight_system_id"`
	AnsibleFactsModified string            `json:"ansible_facts_modified"`
}

type HostRelated struct {
	VariableData        string `json:"variable_data"`
	Groups              string `json:"groups"`
	AllGroups           string `json:"all_groups"`
	JobEvents           string `json:"job_events"`
	JobHostSummaries    string `json:"job_host_summaries"`
	JobActivityStream   string `json:"job_activity_stream"`
	InventorySources    string `json:"inventory_sources"`
	SmartInventories    string `json:"smart_inventories"`
	AdHocCommands       string `json:"ad_hoc_commands"`
	AdHocCommandsEvents string `json:"ad_hoc_commands_events"`
	Insights            string `json:"insights"`
	AnsibleFacts        string `json:"ansible_facts"`
	Inventory           string `json:"inventory"`
}

type HostSummaryFields struct {
	Inventory          InventorySummary     `json:"inventory"`
	LastJob            JobSummary           `json:"last_job"`
	LastJobHostSummary JobSummary           `json:"last_job_host_summary"`
	ModifiedBy         PersonSummary        `json:"modified_by"`
	UserCapabilities   HostUserCapabilities `json:"user_capabilities"`
	Groups             GroupsSummary        `json:"groups"`
	RecentJobs         []JobSummary         `json:"recent_jobs"`
}

type GroupsSummary struct {
	Count  int            `json:"count"`
	Results []GroupSummary `json:"results"`
}

type HostUserCapabilities struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

type JobSummary struct {
	ID              int       `json:"id,omitifempty"`
	Name            string    `json:"name,omitifempty"`
	Status          string    `json:"status,omitifempty"`
	Finished        time.Time `json:"finished,string,omitifempty"`
	JobTemplateId   int       `json:"job_template_id,omitifempty"`
	JobTemplateName string    `json:"job_template_name,omitifempty"`
	Failed          bool      `json:"failed,omitifempty"`
}

type HostResults struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous string  `json:"previous"`
	Results  [] Host `json:"results"`
}
