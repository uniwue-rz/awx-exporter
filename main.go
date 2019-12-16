package main

import (
	altMgrConfig "./alertmanager/config"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

/// AWXConfig Is used for the basic configuration of the AWX connection
type AWXConfig struct {
	Host             string
	UserName         string
	InventorySources []string
	Token            string
	Timeout          time.Duration
}

/// PrometheusConfig is used for the keys of the variables that contain the prometheus config
type PrometheusConfig struct {
	configName         string
	configHostOverride bool
	HostNameVar        string
	IpVar              string
}

/// BlackboxConfig contains the config name for the black box
type BlackboxConfig struct {
	configName    string
	IgnoredGroups []string
	HostNameVar   string
	IpVar         string
}

/// AlertManagerConfig contains the config name for the Alertmanager
type AlertManagerConfig struct {
	configName  string
	sourceFile  string
	sendResolve bool
	requireTls  bool
}

/// Creates the config object that should be used for the application
type Config struct {
	awx          AWXConfig
	prometheus   PrometheusConfig
	blackbox     BlackboxConfig
	alertmanager AlertManagerConfig
}

/// Creates a new AWX request that can be used for the query
func createAuthenticateAWXRequest(config Config, path string, method string, body io.Reader, withoutPrefix bool) *http.Request {
	fullUrl := fmt.Sprintf("%s/api/v2/%s", config.awx.Host, path)
	if withoutPrefix {
		fullUrl = fmt.Sprintf("%s%s", config.awx.Host, path)
	}
	req, err := http.NewRequest(method, fullUrl, body)
	if err != nil {
		log.Fatal("Error creating the request", err)
	}
	bearerToken := fmt.Sprintf("Bearer %s", config.awx.Token)
	req.Header.Set("Authorization", bearerToken)
	return req
}

/// sendRequest Sends the request with the given configuration
func sendRequest(r *http.Request, config Config) *http.Response {
	client := http.Client{Timeout: config.awx.Timeout}
	data, err := client.Do(r)
	if err != nil {
		log.Fatal("Error sending the request", err)
	}
	return data
}

/// getInventories Returns the inventory query results
func getInventories(config Config, inventoryName string) InventoryResult {
	var path string
	if inventoryName != "" {
		path = fmt.Sprintf("inventories?name=%s", inventoryName)
	} else {
		path = "inventories"
	}
	request := createAuthenticateAWXRequest(config, path, "GET", nil, false)
	response := sendRequest(request, config)
	if response.StatusCode != 200 {
		log.Fatalf("Server returns error status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	var results InventoryResult
	err := decoder.Decode(&results)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return results
}

/// Returns the group that match the given search query
func getGroups(config Config, searchQuery string) GroupResults {
	var path string
	if searchQuery != "" {
		path = fmt.Sprintf("groups/?%s", searchQuery)
	} else {
		path = "groups"
	}
	request := createAuthenticateAWXRequest(config, path, "GET", nil, false)
	response := sendRequest(request, config)
	if response.StatusCode != 200 {
		log.Fatalf("Server returns error status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	var results GroupResults
	err := decoder.Decode(&results)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return results
}

/// getHosts Returns the hosts that match the given query string
func getHosts(config Config, searchQuery string) HostResults {
	var path string
	if searchQuery != "" {
		path = fmt.Sprintf("hosts/?%s", searchQuery)
	} else {
		path = "hosts"
	}
	request := createAuthenticateAWXRequest(config, path, "GET", nil, false)
	response := sendRequest(request, config)
	if response.StatusCode != 200 {
		log.Fatalf("Server returns error status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	var results HostResults
	err := decoder.Decode(&results)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return results
}

/// getGroupsHosts Returns the hosts that belong to a given group
func getGroupHost(config Config, group Group) HostResults {
	request := createAuthenticateAWXRequest(config, group.Related.Hosts, "GET", nil, true)
	response := sendRequest(request, config)
	decoder := json.NewDecoder(response.Body)
	var results HostResults
	err := decoder.Decode(&results)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return results
}

///getHostVariables Returns the host data that should be used.
func getHostVariables(config Config, host Host) map[string]interface{} {
	vars := make(map[string]interface{})
	request := createAuthenticateAWXRequest(
		config,
		host.Related.VariableData,
		"GET",
		nil,
		true)
	response := sendRequest(request, config)
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&vars)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return vars
}

/// getGroupVariables Returns the group variables for the
func getGroupVariables(config Config, group Group) map[string]interface{} {
	vars := make(map[string]interface{})
	request := createAuthenticateAWXRequest(
		config,
		group.Related.VariableData,
		"GET",
		nil,
		true)
	response := sendRequest(request, config)
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&vars)
	if err != nil {
		log.Fatal("There was an error decoding the results", err)
	}
	return vars
}

///createPrometheusHosts Creates the host nodes that can be directly extracted as prometheus configurations
func createPrometheusHosts(
	config Config,
	group string,
	hostVariables map[string]interface{},
	prometheusConfig interface{},
	prometheusHosts []PrometheusHost) []PrometheusHost {
	// Set the prometheus config to host one if the host has any setting
	// and host override is set to true
	if hostPrometheusConfig, ok := hostVariables[config.prometheus.configName];
		ok && config.prometheus.configHostOverride {
		prometheusConfig = hostPrometheusConfig
	}
	for _, promSingleNode := range prometheusConfig.([]interface{}) {
		prometheusHost := PrometheusHost{}
		labels := PrometheusHostLabel{}
		var targets []string
		if ipVar, ok := hostVariables[config.prometheus.IpVar]; ok {
			labels.IP = fmt.Sprintf("%v", ipVar)
		}
		if hostNameVar, ok := hostVariables[config.prometheus.HostNameVar]; ok {
			labels.Host = fmt.Sprintf("%v", hostNameVar)
		}
		labels.Group = group
		if prometheusJobName, ok := promSingleNode.(map[string]interface{})["name"]; ok {
			labels.Job = fmt.Sprintf("%v", prometheusJobName)
		}
		if prometheusPort, ok := promSingleNode.(map[string]interface{})["port"]; ok {
			target := fmt.Sprintf("%s:%.0f", labels.Host, prometheusPort)
			targets = append(targets, target)
		}
		prometheusHost.Labels = labels
		prometheusHost.Targets = targets
		prometheusHosts = append(prometheusHosts, prometheusHost)
	}
	return prometheusHosts
}

///createPrometheusConfig Creates the Prometheus config
func createPrometheusConfig(config Config, nextPage string, prometheusHosts []PrometheusHost) []PrometheusHost {
	allGroups := GroupResults{}
	if nextPage == "" {
		allGroups = getGroups(config, "")
	} else {
		allGroups = getGroups(config, nextPage)
	}
	if allGroups.Count > 0 {
		for _, group := range allGroups.Results {
			groupVariables := getGroupVariables(config, group)
			if prometheusConfig, ok := groupVariables[config.prometheus.configName]; ok {
				hosts := getGroupHost(config, group)
				for _, host := range hosts.Results {
					hostVariables := getHostVariables(config, host)
					prometheusHosts = createPrometheusHosts(config, group.Name, hostVariables, prometheusConfig, prometheusHosts)
				}
			}
		}
	}
	Next := allGroups.Next
	if Next != "" {
		parsedUrl, err := url.Parse(Next)
		if err != nil {
			log.Fatal("The given url can not be parsed", err)
		}
		nextPageQuery := parsedUrl.RawQuery
		return createPrometheusConfig(config, nextPageQuery, prometheusHosts)
	}
	return prometheusHosts
}

/// getHostWithBlackBoxConfig Returns the hosts with blackbox configuration
func getHostWithBlackBoxConfig(config Config) HostResults {
	hosts := getHosts(config, "host_filter=variables__icontains=blackbox_config")
	return hosts
}

/// createBlackBoxHosts Creates the blackbox list from the host variables.
func createBlackBoxHosts(
	config Config,
	group string,
	hostVariables map[string]interface{},
	blackboxHosts []BlackboxHost) []BlackboxHost {
	if blackboxConfig, ok := hostVariables[config.blackbox.configName]; ok {
		for _, singleBlackboxConfig := range blackboxConfig.([]interface{}) {
			blackboxHost := BlackboxHost{}
			labels := BlackboxHostLabel{}
			if ipVar, ok := hostVariables[config.blackbox.IpVar]; ok {
				labels.IP = fmt.Sprintf("%v", ipVar)
			}
			if hostNameVar, ok := hostVariables[config.blackbox.HostNameVar]; ok {
				labels.Host = fmt.Sprintf("%v", hostNameVar)
			}
			if module, ok := singleBlackboxConfig.(map[string]interface{})["module"]; ok {
				labels.Module = fmt.Sprintf("%v", module)
			}
			labels.Job = "blackbox"
			labels.Group = group
			targets := singleBlackboxConfig.(map[string]interface{})["targets"].([]interface{})
			for _, target := range targets {
				blackboxHost.Targets = append(blackboxHost.Targets, fmt.Sprintf("%v", target))
			}
			blackboxHost.Labels = labels
			blackboxHosts = append(blackboxHosts, blackboxHost)
		}

	}
	return blackboxHosts
}

/// inSlice Checks if the given key exists in the given slice
func inSlice(key string, dataList []string) bool {
	for _, element := range dataList {
		if element == key {
			return true
		}
	}
	return false
}

/// getBlackboxHostGroup Returns the right Group for the given blackbox
/// It is the first group that is not inside the IgnoredGroups
func getBlackboxHostGroup(config Config, host Host) string {
	groups := host.SummaryFields.Groups.Results
	for _, group := range groups {
		if inSlice(group.Name, config.blackbox.IgnoredGroups) == false {
			return group.Name
		}
	}
	return ""
}

/// createBlackboxConfig Creates the blackbox configuration objects that can be printed as json
func createBlackboxConfig(config Config, nextPage string, blackboxHosts []BlackboxHost) []BlackboxHost {
	hosts := HostResults{}
	if nextPage == "" {
		hosts = getHostWithBlackBoxConfig(config)
	} else {
		hosts = getHosts(config, nextPage)
	}
	if hosts.Count > 0 {
		for _, host := range hosts.Results {
			variables := getHostVariables(config, host)
			group := getBlackboxHostGroup(config, host)
			if group != "" {
				blackboxHosts = createBlackBoxHosts(config, group, variables, blackboxHosts)
			}
		}
	}
	Next := hosts.Next
	if Next != "" {
		parsedUrl, err := url.Parse(Next)
		if err != nil {
			log.Fatal("The given url can not be parsed", err)
		}
		nextPageQuery := parsedUrl.RawQuery
		return createBlackboxConfig(config, nextPageQuery, blackboxHosts)
	}
	return blackboxHosts
}

/// notifierExists checks if the given notifier exists in the given list, returns it when not gives error
func notifierExists(notifierName string, notifiers []AlertManagerEmailNotifier) (AlertManagerEmailNotifier, error) {
	for _, notifier := range notifiers {
		if notifier.getReceiverName() == notifierName {
			return notifier, nil
		}
	}
	var notifier AlertManagerEmailNotifier
	return notifier, errors.New("notifier not found")
}

/// removeNotExistingRoutes Removes the non existing routes form the data config
func removeNotExistingRoutes(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	i := 0
	for _, route := range dataConfig.Route.Routes {
		if dynamicReceiverRegexp.MatchString(route.Receiver) {
			_, err := notifierExists(route.Receiver, notifiers)
			if err == nil {
				dataConfig.Route.Routes[i] = route
				i++
			}
		} else {
			dataConfig.Route.Routes[i] = route
			i++
		}
	}
	dataConfig.Route.Routes = dataConfig.Route.Routes[:i]
}

/// updateExistingRoutes Updates the existing routes
func updateExistingRoutes(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	for _, route := range dataConfig.Route.Routes {
		if dynamicReceiverRegexp.MatchString(route.Receiver) {
			notifier, err := notifierExists(route.Receiver, notifiers)
			if err == nil {
				match := make(map[string]string)
				match["group"] = notifier.Group
				route.Match = match
			}
		}
	}
}

/// routeExistsInDataConfig Checks if the given route exist in the data config
func routeExistsInDataConfig(routeReceiver string, dataConfig *altMgrConfig.Config) bool {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	for _, route := range dataConfig.Route.Routes {
		if dynamicReceiverRegexp.MatchString(route.Receiver) {
			if route.Receiver == routeReceiver {
				return true
			}
		}
	}
	return false
}

/// addNewRoutes Add the new routes to the configurations
func addNewRoutes(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	for _, notifier := range notifiers {
		if routeExistsInDataConfig(notifier.getReceiverName(), dataConfig) == false {
			route := altMgrConfig.Route{}
			route.Receiver = notifier.getReceiverName()
			match := make(map[string]string)
			match["group"] = notifier.Group
			route.Match = match
			dataConfig.Route.Routes = append(dataConfig.Route.Routes, &route)
		}
	}
}

/// removeNotExistingReceivers Removes the non existing notifiers from the list of notifiers
func removeNotExistingReceivers(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	i := 0
	for _, receiver := range dataConfig.Receivers {
		if dynamicReceiverRegexp.MatchString(receiver.Name) {
			_, err := notifierExists(receiver.Name, notifiers)
			if err == nil {
				dataConfig.Receivers[i] = receiver
				i++
			}
		} else {
			dataConfig.Receivers[i] = receiver
			i++
		}
	}
	dataConfig.Receivers = dataConfig.Receivers[:i]
}

/// receiverExistsInDataConfig Check is the receiver exists in the data config list
func receiverExistsInDataConfig(receiverName string, dataConfig *altMgrConfig.Config) bool {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	for _, receiver := range dataConfig.Receivers {
		if dynamicReceiverRegexp.MatchString(receiver.Name) {
			if receiver.Name == receiverName {
				return true
			}
		}
	}
	return false
}

/// Update the existing dynamic receivers with the new values
func updateExistingReceivers(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	dynamicReceiverRegexp := regexp.MustCompile("^dynamic-*")
	for _, receiver := range dataConfig.Receivers {
		if dynamicReceiverRegexp.MatchString(receiver.Name) {
			notifier, err := notifierExists(receiver.Name, notifiers)
			if err == nil {
				for idx := range receiver.EmailConfigs {
					receiver.EmailConfigs[idx].RequireTLS = &notifier.RequireTLS
					receiver.EmailConfigs[idx].To = notifier.Email
					receiver.EmailConfigs[idx].VSendResolved = notifier.SendResolved
				}
			}
		}
	}
}

/// addNewReceivers Adds the new receivers to the list of existing ones
func addNewReceivers(dataConfig *altMgrConfig.Config, notifiers []AlertManagerEmailNotifier) {
	for _, notifier := range notifiers {
		if receiverExistsInDataConfig(notifier.getReceiverName(), dataConfig) == false {
			receiver := altMgrConfig.Receiver{}
			receiver.Name = notifier.getReceiverName()
			emailConfig := altMgrConfig.EmailConfig{
				To:         notifier.Email,
				RequireTLS: &notifier.RequireTLS,
			}
			emailConfig.VSendResolved = notifier.SendResolved
			receiver.EmailConfigs = append(receiver.EmailConfigs, &emailConfig)
			dataConfig.Receivers = append(dataConfig.Receivers, &receiver)
		}
	}
}

/// createAlertManagerNotifiers Creates the Alert Manager configurations from an existing config file.
func createAlertManagerConfig(config Config) *altMgrConfig.Config {
	var notifiers []AlertManagerEmailNotifier
	notifiers = getAlertManagerNotifiers(config, "", notifiers)
	dataConfig, _, err := readAlertManagerConfig(config)
	if err != nil {
		log.Fatal("Can not read the alertmanager config", err)
	}
	// Remove the non existing receivers
	removeNotExistingReceivers(dataConfig, notifiers)
	// Update the existing receivers
	updateExistingReceivers(dataConfig, notifiers)
	// Add new receivers
	addNewReceivers(dataConfig, notifiers)
	/// Remove non existing routes
	removeNotExistingRoutes(dataConfig, notifiers)
	/// Update existing routes
	updateExistingRoutes(dataConfig, notifiers)
	/// Add new Routes
	addNewRoutes(dataConfig, notifiers)
	return dataConfig
}

/// readAlertManagerConfig reads the alertmanager configurations
func readAlertManagerConfig(applicationConfig Config) (*altMgrConfig.Config, []byte, error) {
	dataConfig, content, err := altMgrConfig.LoadFile(applicationConfig.alertmanager.sourceFile)
	if err != nil {
		panic(err)
	}
	return dataConfig, content, err
}

/// createAlertManagerNotifiers  Creates AlertManager notifiers the given configuration of the AWX
func createAlertManagerNotifiers(
	config Config,
	group string,
	alertManagerConfig interface{},
	notifiers []AlertManagerEmailNotifier) []AlertManagerEmailNotifier {
	for _, alertManagerSingleConfig := range alertManagerConfig.([]interface{}) {
		// Find the notifier type and create
		if notifierType, ok := alertManagerSingleConfig.(map[string]interface{})["type"]; ok {
			if notifierType == "email" {
				emailNotifier := AlertManagerEmailNotifier{}
				// Set the name of the given notifier
				if name, ok := alertManagerSingleConfig.(map[string]interface{})["name"]; ok {
					emailNotifier.Name = fmt.Sprintf("%v", name)
				}
				// Set the group of the given notifier
				emailNotifier.Group = group
				if receiverConfig, ok := alertManagerSingleConfig.(map[string]interface{})["receiver-config"]; ok {
					if emailTo, ok := receiverConfig.(map[string]interface{})["to"]; ok {
						emailNotifier.Email = fmt.Sprintf("%v", emailTo)
					}
					if requireTLS, ok := alertManagerSingleConfig.(map[string]interface{})["require-tls"]; ok {
						emailNotifier.RequireTLS = requireTLS.(bool)
					} else {
						emailNotifier.RequireTLS = config.alertmanager.requireTls
					}
					if sendResolve, ok := alertManagerSingleConfig.(map[string]interface{})["send-resolve"]; ok {
						emailNotifier.SendResolved = sendResolve.(bool)
					} else {
						emailNotifier.SendResolved = config.alertmanager.sendResolve
					}
				}
				notifiers = append(notifiers, emailNotifier)
			}
		}
	}
	return notifiers
}

/// getAlertManagerNotifiers Returns the list of groups that have the alertmanager included
func getAlertManagerNotifiers(
	config Config,
	nextPage string,
	alertManagerNotifiers []AlertManagerEmailNotifier) []AlertManagerEmailNotifier {
	var groups GroupResults
	if nextPage == "" {
		groups = getGroups(config, "variables__icontains=alertmanager_config")
	} else {
		groups = getGroups(config, nextPage)
	}
	if groups.Count > 0 {
		for _, group := range groups.Results {
			groupVariables := getGroupVariables(config, group)
			if alertManagerConfig, ok := groupVariables[config.alertmanager.configName]; ok {
				alertManagerNotifiers = createAlertManagerNotifiers(
					config,
					group.Name,
					alertManagerConfig,
					alertManagerNotifiers)
			}
		}
	}
	Next := groups.Next
	if Next != "" {
		parsedUrl, err := url.Parse(Next)
		if err != nil {
			log.Fatal("The given url can not be parsed", err)
		}
		nextPageQuery := parsedUrl.RawQuery
		return getAlertManagerNotifiers(config, nextPageQuery, alertManagerNotifiers)
	}
	return alertManagerNotifiers
}

/// readConfiguration Returns the configurations file for the given path.
func readConfiguration(configPath string) Config {
	cfg, err := ini.Load(configPath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	configHostOverride, err := cfg.Section("PROMETHEUS").Key("ConfigHostOverride").Bool()
	if err != nil {
		fmt.Printf("The Host override in promtheus should be boolean: %v", err)
		os.Exit(1)
	}
	timeout, err := cfg.Section("AWX").Key("TimeOut").Duration()
	if err != nil {
		fmt.Printf("The timeout in AWX should be an integer with unit (s,m,h,...): %v", err)
		os.Exit(1)
	}
	alertManagerRequireTls, err := cfg.Section("ALERTMANAGER").Key("RequireTLSDefault").Bool()
	if err != nil {
		fmt.Printf("The RequireTLSDefault for the Alertmanager should be boolean: %v", err)
		os.Exit(1)
	}
	alertManagerSendResolve, err := cfg.Section("ALERTMANAGER").Key("SendResolveDefault").Bool()
	if err != nil {
		fmt.Printf("The SendResolveDefault for the Alertmanager should be boolean: %v", err)
		os.Exit(1)
	}
	var config = Config{
		awx: AWXConfig{
			Host:             cfg.Section("AWX").Key("HostName").String(),
			UserName:         cfg.Section("AWX").Key("UserName").String(),
			Token:            cfg.Section("AWX").Key("Token").String(),
			Timeout:          timeout,
			InventorySources: strings.Split(cfg.Section("AWX").Key("InventorySources").String(), ","),
		},
		prometheus: PrometheusConfig{
			configName:         cfg.Section("PROMETHEUS").Key("ConfigName").String(),
			configHostOverride: configHostOverride,
			IpVar:              cfg.Section("PROMETHEUS").Key("IpVar").String(),
			HostNameVar:        cfg.Section("PROMETHEUS").Key("HostNameVar").String(),
		},
		blackbox: BlackboxConfig{
			configName:    cfg.Section("BLACKBOX").Key("ConfigName").String(),
			IgnoredGroups: strings.Split(cfg.Section("BLACKBOX").Key("IgnoredGroups").String(), ","),
			IpVar:         cfg.Section("BLACKBOX").Key("IpVar").String(),
			HostNameVar:   cfg.Section("BLACKBOX").Key("HostNameVar").String(),
		},
		alertmanager: AlertManagerConfig{
			configName:  cfg.Section("ALERTMANAGER").Key("ConfigName").String(),
			sourceFile:  cfg.Section("ALERTMANAGER").Key("SourceFile").String(),
			sendResolve: alertManagerSendResolve,
			requireTls:  alertManagerRequireTls,
		},
	}
	return config
}

func main() {
	configPath := flag.String("config-path", "config.ini", "The path to the configuration")
	alertManagerMode := flag.Bool("alertmanager", false, "The Alert Manager mode for the exporter")
	prometheusMode := flag.Bool("prometheus", false, "The Prometheus mode for the exporter")
	blackboxMode := flag.Bool("blackbox", false, "Blackbox mode for the exporter")
	flag.Parse()
	config := readConfiguration(*configPath)
	if *alertManagerMode {
		alertManagerConfig := createAlertManagerConfig(config)
		fmt.Println(alertManagerConfig)
	}
	if *prometheusMode {
		var prometheusHosts []PrometheusHost
		prometheusHosts = createPrometheusConfig(config, "", prometheusHosts)
		printable, err := json.Marshal(prometheusHosts)
		if err != nil {
			log.Fatal("Error marshaling prometheus host", err)
		}
		fmt.Println(string(printable))
	}
	if *blackboxMode {
		var blackboxHosts []BlackboxHost
		blackboxHosts = createBlackboxConfig(config, "", blackboxHosts)
		printable, err := json.Marshal(blackboxHosts)
		if err != nil {
			log.Fatal("Error marshaling blackbox host", err)
		}
		fmt.Println(string(printable))
	}
}
