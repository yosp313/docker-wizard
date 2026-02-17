package generator

type ServiceSpec struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	Image        string   `json:"image"`
	Ports        []string `json:"ports"`
	Env          []string `json:"env"`
	VolumeMounts []string `json:"volumeMounts"`
	NamedVolumes []string `json:"namedVolumes"`
	DependsOn    []string `json:"dependsOn"`
	Command      []string `json:"command"`
	Selectable   bool     `json:"selectable"`
	Order        int      `json:"order"`
	Requires     []string `json:"requires"`
}
