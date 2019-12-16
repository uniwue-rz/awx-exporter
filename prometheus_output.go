package main

type PrometheusHostLabel struct {
	Group string `json:"group"`
	Host  string `json:"host"`
	IP    string `json:"ip"`
	Job   string `json:"job"`
}

type PrometheusHost struct {
	Labels  PrometheusHostLabel `json:"labels"`
	Targets []string            `json:"targets"`
}

type BlackboxHostLabel struct {
	Group  string `json:"group"`
	Host   string `json:"host"`
	IP     string `json:"ip"`
	Job    string `json:"job"`
	Module string `json:"module"`
}

type BlackboxHost struct {
	Labels  BlackboxHostLabel `json:"labels"`
	Targets []string          `json:"targets"`
}
