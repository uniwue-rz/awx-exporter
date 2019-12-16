package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

/// TestReadConfiguration tests the configuration reading functions
func TestReadConfiguration(t *testing.T) {
	configFile := "config_test.ini"
	config := readConfiguration(configFile)
	// Awx Settings test
	if config.awx.Host != "https://awx.rz.uni-wuerzburg.de" {
		t.Errorf("Hosts are not the same")
	}
	if config.awx.Token != "testToken" {
		t.Errorf("Tokens are not the same")
	}
	data := []string{
		"testSource1",
		"testSource2",
	}
	if reflect.DeepEqual(data, config.awx.InventorySources) == false {
		t.Errorf("The sources are not the same")
	}
	if config.awx.UserName != "testUser" {
		t.Errorf("The usernames are not the same")
	}
	duration, _ := time.ParseDuration("10s")
	if config.awx.Timeout != duration {
		t.Errorf("The timeouts are not the same")
	}
	// Prometheus Settings test
	if config.prometheus.configHostOverride != true {
		t.Errorf("The configHostOverrides do not match")
	}
	if config.prometheus.configName != "prometheus_config" {
		t.Error("The Prometheus config names are not the same")
	}
	// AlertManager Settings test
	if config.alertmanager.configName != "alertmanager_config" {
		t.Error("The AlertManager config names are not the same")
	}
	if config.alertmanager.sourceFile != "conf.good.yml" {
		t.Errorf("The Alertmanager source files are not the same")
	}
	// Blackbox Settings test
	if config.blackbox.configName != "blackbox_config" {
		t.Errorf("The blackbox config names are not the same")
	}
}

/// TestCreateAWXRequest Test Request creation method
/// to make it work set the ENV AWX_TOKEN with a working value
func TestCreateAWXRequest(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	config.awx.Token = awxToken
	path := "inventories"
	method := "GET"
	req := createAuthenticateAWXRequest(config, path, method, nil, false)
	path = fmt.Sprintf("/api/v2/%s", path)
	parsedUrl, _ := url.Parse(config.awx.Host)

	if req.Host != parsedUrl.Host {
		t.Errorf("The Request hosts are not the same")
	}
	if req.URL.Path != path {
		t.Errorf("The Request path are not the same")
	}
	if req.Body != nil {
		t.Errorf("The Requests body is not the same")
	}
	if req.Method != method {
		t.Errorf("The Rquests method is not the same")
	}
}

/// TestSendRequest Tests the send request method to AWX
func TestSendRequest(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	config.awx.Token = awxToken
	path := "inventories"
	method := "GET"
	req := createAuthenticateAWXRequest(config, path, method, nil, false)
	response := sendRequest(req, config)
	if response.StatusCode != 200 {
		t.Errorf("The response status was not 200")
	}
	decoder := json.NewDecoder(response.Body)
	var results InventoryResult
	err := decoder.Decode(&results)
	if err != nil {
		t.Errorf("There was an error decoding or retrving the data")
	}
	if results.Count == 0 {
		t.Errorf("There was an error decoding or retrving the data")
	}
}

/// Tests if the results are valid
func testCreatePrometheusConfig(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	config.awx.Token = awxToken
	var groups []PrometheusHost
	groupsRes := createPrometheusConfig(config, "", groups)
	if len(groupsRes) == 0 {
		t.Errorf("The results are not valid")
	}
}

/// Tests if the blackbox configs can be generated
func testCreateBlackboxConfig(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	var hosts []BlackboxHost
	config.awx.Token = awxToken
	blackboxHosts := createBlackboxConfig(config, "", hosts)
	if len(blackboxHosts) == 0 {
		t.Errorf("The results are not valid")
	}
}

func TestGetAlertManagerNotifiers(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	var alertManagerNotifiers []AlertManagerEmailNotifier
	config.awx.Token = awxToken
	alertNotifiers := getAlertManagerNotifiers(config, "", alertManagerNotifiers)
	if len(alertNotifiers) == 0 {
		t.Errorf("The results are not valid")
	}
}

func TestAlertManagerConfigRead(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	config.awx.Token = awxToken
	_, _, err := readAlertManagerConfig(config)
	if err != nil {
		t.Errorf("The results are not valid")
	}
}

func TestCreateAlertManagerConfig(t *testing.T) {
	awxToken := os.Getenv("AWX_TOKEN")
	config := readConfiguration("config_test.ini")
	config.awx.Token = awxToken
	alertManagerConfig := createAlertManagerConfig(config)
	if len(alertManagerConfig.Route.Routes) == 0 {
		t.Errorf("The results are not valid")

	}
}
