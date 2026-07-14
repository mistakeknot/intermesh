package profile

const (
	ModeManaged     = "managed"
	ModeRoutingOnly = "routing_only"
	RouterEntryName = "intermesh-router"
)

type Entry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

type Plan struct {
	Version  int      `json:"version"`
	Host     string   `json:"host"`
	Catalog  string   `json:"catalog"`
	Router   string   `json:"router"`
	Mode     string   `json:"mode"`
	Entries  []Entry  `json:"entries"`
	AlwaysOn []string `json:"always_on"`
	Blockers []string `json:"blockers"`
}

type Snapshot struct {
	Version     int     `json:"version"`
	Path        string  `json:"-"`
	Fingerprint string  `json:"fingerprint"`
	Host        string  `json:"host"`
	Catalog     string  `json:"catalog"`
	Router      string  `json:"router"`
	Entries     []Entry `json:"entries"`
}
