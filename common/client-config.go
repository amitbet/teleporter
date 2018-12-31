package common

type ClientConfig struct {
	ClientId       string            `json:"clientId"`
	NetworkExports []string          `json:"networkExports"`
	Mapping        map[string]string `json:"networkMapping"`
}
