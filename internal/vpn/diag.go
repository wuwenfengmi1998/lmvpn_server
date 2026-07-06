package vpn

import (
	"os"
	"runtime"
)

type DiagResult struct {
	Platform        string `json:"platform"`
	IsRoot          bool   `json:"is_root"`
	HasCapNetAdmin  *bool  `json:"has_cap_net_admin"`
	CapNetAdminNote string `json:"cap_net_admin_note,omitempty"`
	IPForward       *bool  `json:"ip_forward"`
	IPForwardNote   string `json:"ip_forward_note,omitempty"`
	Masquerade      *bool  `json:"masquerade"`
	MasqueradeNote  string `json:"masquerade_note,omitempty"`
	TUNCreate       string `json:"tun_create"`
	TUNRunning      bool   `json:"tun_running"`
	TUNName         string `json:"tun_name,omitempty"`
}

func Diag(svc *VpnService) DiagResult {
	r := DiagResult{Platform: runtime.GOOS, IsRoot: os.Getuid() == 0}

	if svc != nil {
		svc.mu.RLock()
		r.TUNRunning = svc.running
		if svc.tun != nil {
			r.TUNName = svc.tun.Name()
		}
		svc.mu.RUnlock()
	}

	if r.TUNRunning {
		r.TUNCreate = "ok: " + r.TUNName
	} else {
		r.TUNCreate = testTUNCreate()
	}

	fillPlatformDiag(&r)
	return r
}

func testTUNCreate() string {
	t, err := CreateTUN("")
	if err != nil {
		return "fail: " + err.Error()
	}
	name := t.Name()
	_ = t.Close()
	return "ok: " + name
}
