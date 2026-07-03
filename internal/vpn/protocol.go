package vpn

type initMessage struct {
	Type     string `json:"type"`
	IP       string `json:"ip"`
	Prefix   int    `json:"prefix"`
	MTU      int    `json:"mtu"`
	ServerIP string `json:"server_ip"`
}

type controlMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
