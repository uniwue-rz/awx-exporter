package main

import "fmt"

type AlertManagerEmailNotifier struct {
	Name         string `json:"name"`
	Group        string `json:"group"`
	Email        string `json:"email"`
	RequireTLS   bool   `json:"require_tls"`
	SendResolved bool   `json:"send_resolved"`
}

/// getReceiverName Returns the receiver name
func (alertManagerEmailNotifier *AlertManagerEmailNotifier) getReceiverName() string {
	return fmt.Sprintf("dynamic-%s-%s-receiver", alertManagerEmailNotifier.Group, alertManagerEmailNotifier.Name)
}
