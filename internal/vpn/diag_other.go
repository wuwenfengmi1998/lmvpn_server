//go:build !linux

package vpn

func fillPlatformDiag(r *DiagResult) {
	r.HasCapNetAdmin = nil
	r.CapNetAdminNote = "仅 Linux 适用"
	r.IPForward = nil
	r.IPForwardNote = "仅 Linux 适用"
	r.Masquerade = nil
	r.MasqueradeNote = "仅 Linux 适用"
	r.IP6Forward = nil
	r.IP6ForwardNote = "仅 Linux 适用"
	r.Masquerade6 = nil
	r.Masquerade6Note = "仅 Linux 适用"
}
