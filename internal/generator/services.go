package generator

type ServiceSpec struct {
	ID           string
	Name         string
	Image        string
	Ports        []string
	Env          []string
	VolumeMounts []string
	NamedVolumes []string
	DependsOn    []string
	Command      []string
}

var serviceCatalog = map[string]ServiceSpec{
	"mysql": {
		ID:           "mysql",
		Name:         "mysql",
		Image:        "mysql:8.0",
		Ports:        []string{"3306:3306"},
		Env:          []string{"MYSQL_ROOT_PASSWORD=example"},
		VolumeMounts: []string{"mysql-data:/var/lib/mysql"},
		NamedVolumes: []string{"mysql-data"},
	},
	"postgres": {
		ID:           "postgres",
		Name:         "postgres",
		Image:        "postgres:16",
		Ports:        []string{"5432:5432"},
		Env:          []string{"POSTGRES_PASSWORD=example"},
		VolumeMounts: []string{"postgres-data:/var/lib/postgresql/data"},
		NamedVolumes: []string{"postgres-data"},
	},
	"redis": {
		ID:    "redis",
		Name:  "redis",
		Image: "redis:7-alpine",
		Ports: []string{"6379:6379"},
	},
	"analytics": {
		ID:    "analytics",
		Name:  "analytics",
		Image: "metabase/metabase:latest",
		Ports: []string{"3000:3000"},
	},
	"nginx": {
		ID:    "nginx",
		Name:  "nginx",
		Image: "nginx:alpine",
		Ports: []string{"80:80"},
	},
	"traefik": {
		ID:      "traefik",
		Name:    "traefik",
		Image:   "traefik:v2.11",
		Ports:   []string{"80:80", "8080:8080"},
		Command: []string{"--api.insecure=true", "--providers.docker=true"},
	},
	"caddy": {
		ID:           "caddy",
		Name:         "caddy",
		Image:        "caddy:2",
		Ports:        []string{"80:80", "443:443"},
		VolumeMounts: []string{"caddy-data:/data"},
		NamedVolumes: []string{"caddy-data"},
	},
	"rabbitmq": {
		ID:    "rabbitmq",
		Name:  "rabbitmq",
		Image: "rabbitmq:3-management",
		Ports: []string{"5672:5672", "15672:15672"},
	},
	"zookeeper": {
		ID:    "zookeeper",
		Name:  "zookeeper",
		Image: "bitnami/zookeeper:3.9",
		Ports: []string{"2181:2181"},
		Env:   []string{"ALLOW_ANONYMOUS_LOGIN=yes"},
	},
	"kafka": {
		ID:        "kafka",
		Name:      "kafka",
		Image:     "bitnami/kafka:3.7",
		Ports:     []string{"9092:9092"},
		DependsOn: []string{"zookeeper"},
		Env: []string{
			"KAFKA_BROKER_ID=1",
			"KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181",
			"KAFKA_LISTENERS=PLAINTEXT://:9092",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092",
			"ALLOW_PLAINTEXT_LISTENER=yes",
		},
	},
}

var serviceOrder = []string{
	"mysql",
	"postgres",
	"redis",
	"analytics",
	"nginx",
	"traefik",
	"caddy",
	"rabbitmq",
	"kafka",
}
