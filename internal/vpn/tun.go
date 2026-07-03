package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/songgao/water"
)

type TUNInterface struct {
	Iface *water.Interface
}

func CreateTUN(name string) (*TUNInterface, error) {
	cfg := water.Config{DeviceType: water.TUN}
	cfg.Name = name
	ifce, err := water.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 TUN 设备失败: %w", err)
	}
	return &TUNInterface{Iface: ifce}, nil
}

func (t *TUNInterface) Name() string {
	return t.Iface.Name()
}

func (t *TUNInterface) Close() error {
	return t.Iface.Close()
}

func execCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %s: %w", name, strings.Join(arg, " "), err)
	}
	return nil
}
