package agent

type ListenerConfig struct {
	Port              int               `json:"port"`
	Type              string            `json:"type"`
	LocalOnly         bool              `json:"acceptLocalOnly"`
	UseAuthentication bool              `json:"useAuthentication"`
	AuthorizedClients map[string]string `json:"authClients"`
}
// type AuthClient struct {
// 	ClientId string `json:"clientId"`
// 	Secret   string `json:"secret"`
// }

type TetherConfig struct {
	TargetPort     int        `json:"port"`
	TargetHost     string     `json:"host"`
	ConnectionType string     `json:"connectionType"`
	ConnectionName string     `json:"connectionName"`
	Proxy          *ProxyInfo `json:"proxy"`
	ClientPassword string     `json:"password"`
}

type AgentConfig struct {
	Servers              []ListenerConfig `json:"servers"`
	Connections          []TetherConfig   `json:"tethers"`
	NetworkConfiguration ClientConfig     `json:"netConf"`
	Proxy                *ProxyInfo       `json:"proxy"`
	NumConnsPerTether    int              `json:"numConnsPerTether"`
}
type ClientConfig struct {
	Secret   string            `json:"secret"`
	ClientId string            `json:"clientId"`
	Mapping  map[string]string `json:"networkMapping"` // "<ip or domain>" : "<clientId>"
}
