package catalog

type ServiceSpec struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Image        string   `json:"image"`
	Ports        []string `json:"ports"`
	Expose       []string `json:"expose"`
	Env          []string `json:"env"`
	VolumeMounts []string `json:"volumeMounts"`
	NamedVolumes []string `json:"namedVolumes"`
	DependsOn    []string `json:"dependsOn"`
	Command      []string `json:"command"`
	Public       bool     `json:"public"`
	Selectable   bool     `json:"selectable"`
	Order        int      `json:"order"`
	Requires     []string `json:"requires"`
}
