//go:build !linux

package vpn

func fillPlatformDiag(r *DiagResult) {
	r.HasCapNetAdmin = nil
	r.CapNetAdminNote = "仅 Linux 适用"
	r.IPForward = nil
	r.IPForwardNote = "仅 Linux 适用"
	r.Masquerade = nil
	r.MasqueradeNote = "仅 Linux 适用"
}
