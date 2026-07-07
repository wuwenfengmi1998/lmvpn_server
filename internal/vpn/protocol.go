package vpn

type initMessage struct {
	Type      string `json:"type"`
	IP        string `json:"ip"`
	Prefix    int    `json:"prefix"`
	MTU       int    `json:"mtu"`
	ServerIP  string `json:"server_ip"`
	IP6       string `json:"ip6,omitempty"`
	Prefix6   int    `json:"prefix6,omitempty"`
	ServerIP6 string `json:"server_ip6,omitempty"`
}

type controlMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
