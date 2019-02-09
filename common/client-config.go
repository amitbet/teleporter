package common

type ClientConfig struct {
	ClientId string `json:"clientId"`
	// NetworkExports []string          `json:"networkExports"` // "192.168.1*" or "*mydomain.*" or "172.54.23.111" or just "*"
	Mapping map[string]string `json:"networkMapping"` // "<ip or domain>" : "<clientId>"
}
