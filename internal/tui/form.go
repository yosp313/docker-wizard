package tui

import (
	"strings"

	"github.com/charmbracelet/huh"
)

type Selection struct {
	Services  []string
	Cancelled bool
}

type serviceOption struct {
	Label string
	ID    string
}

var serviceOptions = []serviceOption{
	{Label: "MySQL", ID: "mysql"},
	{Label: "PostgreSQL", ID: "postgres"},
	{Label: "Redis", ID: "redis"},
	{Label: "Analytics", ID: "analytics"},
	{Label: "Nginx", ID: "nginx"},
	{Label: "Traefik", ID: "traefik"},
	{Label: "Caddy", ID: "caddy"},
	{Label: "RabbitMQ", ID: "rabbitmq"},
	{Label: "Kafka", ID: "kafka"},
}

func Run() (Selection, error) {
	var services []string
	confirm := true

	options := make([]huh.Option[string], 0, len(serviceOptions))
	for _, opt := range serviceOptions {
		options = append(options, huh.NewOption(opt.Label, opt.ID))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select services for docker-compose").
				Description("Choose one or more services to add").
				Options(options...).
				Value(&services),
			huh.NewConfirm().
				Title("Generate files in the current directory").
				Description("This will create docker-compose.yml and Dockerfile").
				Affirmative("Generate").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return Selection{}, err
	}

	if !confirm {
		return Selection{Cancelled: true}, nil
	}

	for i := range services {
		services[i] = strings.TrimSpace(services[i])
	}

	return Selection{Services: services}, nil
}
