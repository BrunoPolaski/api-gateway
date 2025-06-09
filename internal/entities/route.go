package entities

type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Target string `json:"target"`
}
