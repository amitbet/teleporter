package agent

type ListenerConfig struct {
	Port      int    `json:"port"`
	Type      string `json:"type"`
	LocalOnly bool   `json:"acceptLocalOnly"`
}
type TetherConfig struct {
	TargetPort     int        `json:"port"`
	TargetHost     string     `json:"host"`
	ConnectionType string     `json:"connectionType"`
	ConnectionName string     `json:"connectionName"`
	Credentials    string     `json:"credentials"`
	Proxy          *ProxyInfo `json:"proxy"`
}

type AgentConfig struct {
	Servers              []ListenerConfig `json:"servers"`
	Connections          []TetherConfig   `json:"tethers"`
	NetworkConfiguration ClientConfig     `json:"netConf"`
	AuthenticateSocks5   bool             `json:"authenticateSocks5"`
	Proxy                *ProxyInfo       `json:"proxy"`
	NumConnsPerTether    int              `json:"numConnsPerTether"`
}
type ClientConfig struct {
	ClientId string            `json:"clientId"`
	Mapping  map[string]string `json:"networkMapping"` // "<ip or domain>" : "<clientId>"

}
