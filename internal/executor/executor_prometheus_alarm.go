package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/internal/command"
	"infra/internal/helper"
	v1 "k8s.io/api/networking/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type PrometheusAlarmExecutor struct {
	k8s             client.Client
	commandExecutor *command.Executor
	http            *http.Client
}

func NewPrometheusAlarmExecutor(k8s client.Client, commandExecutor *command.Executor) *PrometheusAlarmExecutor {
	return &PrometheusAlarmExecutor{
		k8s:             k8s,
		commandExecutor: commandExecutor,
		http:            http.DefaultClient,
	}
}

func (p *PrometheusAlarmExecutor) Run(ctx context.Context) error {
	prometheusEndpoint := ""
	if !helper.CommandExists("rke2") {
		var ingresses v1.IngressList
		err := p.k8s.List(ctx, &ingresses, &client.ListOptions{
			Namespace: "monitoring",
		})
		if err != nil {
			return fmt.Errorf("failed to list ingresses: %w", err)
		}

		for _, ing := range ingresses.Items {
			for _, rule := range ing.Spec.Rules {
				if strings.HasPrefix(rule.Host, "prometheus") {
					prometheusEndpoint = fmt.Sprintf("https://%s", rule.Host)
				}
			}
		}
	} else {
		p.commandExecutor.Execute("kubectl", "port-forward", "prometheus-prometheus-monitoring-kube-prometheus-0", "-n", "monitoring")
		prometheusEndpoint = "http://localhost:9090"
	}

	if prometheusEndpoint == "" {
		return fmt.Errorf("failed to find alertmanager ingress")
	}

	res, err := p.GetAlerts(prometheusEndpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch alerts: %w", err)
	}
	var alerts []string
	for _, alert := range res.Data.Groups {
		for _, rule := range alert.Rules {
			ruleName := rule.Name
			ruleAlerts := make([]string, 0)
			for _, alert := range rule.Alerts {
				if alert.State != "firing" || alert.Labels["alertname"] == "Watchdog" || alert.Labels["alertname"] == "InfoInhibitor" {
					continue
				}
				ruleAlerts = append(ruleAlerts, alert.Annotations["description"])
			}
			if len(ruleAlerts) > 0 {
				alerts = append(alerts, fmt.Sprintf("%s: \n\t%v", ruleName, strings.Join(ruleAlerts, "\n\t")))
			}
		}
	}
	if len(alerts) > 0 {
		return fmt.Errorf("alerts detected: \n%v", strings.Join(alerts, "\n"))
	}
	return nil
}

func (p *PrometheusAlarmExecutor) Name() string {
	return "prometheus_alarm"
}

func (p *PrometheusAlarmExecutor) GetAlerts(baseUrl string) (*AlertsResponse, error) {
	resp, err := p.http.Get(fmt.Sprintf("%s/api/v1/rules?type=alert", baseUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var alertsResponse AlertsResponse
	err = json.NewDecoder(resp.Body).Decode(&alertsResponse)
	if err != nil {
		return nil, err
	}
	return &alertsResponse, nil
}

type AlertsResponse struct {
	Status string `json:"status"`
	Data   Data
}

type Data struct {
	Groups []Group
}

type Group struct {
	Rules []Rule
}

type Rule struct {
	Name   string
	State  string
	Alerts []Alert
}

type Alert struct {
	Labels      map[string]string
	Annotations map[string]string
	State       string
}
