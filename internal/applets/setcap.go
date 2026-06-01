package applets

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	xattrCapability = "security.capability"

	vfsCapRevision2     = 0x02000000
	vfsCapFlagEffective = 0x000001

	capDataSize = 20
)

var capabilityNames = map[string]int{
	"cap_chown":              0,
	"cap_dac_override":       1,
	"cap_dac_read_search":    2,
	"cap_fowner":             3,
	"cap_fsetid":             4,
	"cap_kill":               5,
	"cap_setgid":             6,
	"cap_setuid":             7,
	"cap_setpcap":            8,
	"cap_linux_immutable":    9,
	"cap_net_bind_service":   10,
	"cap_net_broadcast":      11,
	"cap_net_admin":          12,
	"cap_net_raw":            13,
	"cap_ipc_lock":           14,
	"cap_ipc_owner":          15,
	"cap_sys_module":         16,
	"cap_sys_rawio":          17,
	"cap_sys_chroot":         18,
	"cap_sys_ptrace":         19,
	"cap_sys_pacct":          20,
	"cap_sys_admin":          21,
	"cap_sys_boot":           22,
	"cap_sys_nice":           23,
	"cap_sys_resource":       24,
	"cap_sys_time":           25,
	"cap_sys_tty_config":     26,
	"cap_mknod":              27,
	"cap_lease":              28,
	"cap_audit_write":        29,
	"cap_audit_control":      30,
	"cap_setfcap":            31,
	"cap_mac_override":       32,
	"cap_mac_admin":          33,
	"cap_syslog":             34,
	"cap_wake_alarm":         35,
	"cap_block_suspend":      36,
	"cap_audit_read":         37,
	"cap_perfmon":            38,
	"cap_bpf":                39,
	"cap_checkpoint_restore": 40,
}

func init() {
	Register("setcap", "set file capabilities", Setcap)
}

func Setcap(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: setcap [-r] capability file")
	}

	if args[0] == "-r" || args[0] == "--remove" {
		if len(args) != 2 {
			return fmt.Errorf("usage: setcap -r file")
		}

		return removeCapability(args[1])
	}

	if len(args) != 2 {
		return fmt.Errorf("usage: setcap capability file")
	}

	spec := args[0]
	path := args[1]

	return setCapabilitySpec(path, spec)
}

func setCapabilitySpec(path string, spec string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	caps, effective, err := parseCapabilitySpec(spec)
	if err != nil {
		return err
	}

	data := buildCapabilityXattr(caps, effective)

	if err := syscall.Setxattr(path, xattrCapability, data, 0); err != nil {
		return fmt.Errorf("setcap %s %s: %w", spec, path, err)
	}

	return nil
}

func removeCapability(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	if err := syscall.Removexattr(path, xattrCapability); err != nil {
		if err == syscall.ENODATA || err == syscall.ENOTSUP {
			return nil
		}

		return fmt.Errorf("setcap -r %s: %w", path, err)
	}

	return nil
}

func parseCapabilitySpec(spec string) ([]int, bool, error) {
	parts := strings.Split(spec, "+")
	if len(parts) != 2 {
		return nil, false, fmt.Errorf("invalid capability spec: %s", spec)
	}

	capPart := parts[0]
	flagPart := parts[1]

	if flagPart == "" {
		return nil, false, fmt.Errorf("missing capability flags")
	}

	effective := strings.Contains(flagPart, "e")
	permitted := strings.Contains(flagPart, "p")

	if !permitted {
		return nil, false, fmt.Errorf("only permitted capabilities are supported, use +p or +ep")
	}

	var caps []int

	for _, rawName := range strings.Split(capPart, ",") {
		name := strings.ToLower(strings.TrimSpace(rawName))
		if name == "" {
			continue
		}

		capNum, ok := capabilityNames[name]
		if !ok {
			return nil, false, fmt.Errorf("unknown capability: %s", name)
		}

		caps = append(caps, capNum)
	}

	if len(caps) == 0 {
		return nil, false, fmt.Errorf("no capabilities specified")
	}

	return caps, effective, nil
}

func buildCapabilityXattr(caps []int, effective bool) []byte {
	data := make([]byte, capDataSize)

	magic := uint32(vfsCapRevision2)
	if effective {
		magic |= vfsCapFlagEffective
	}

	binary.LittleEndian.PutUint32(data[0:4], magic)

	var permittedLow uint32
	var permittedHigh uint32

	for _, capNum := range caps {
		if capNum < 32 {
			permittedLow |= 1 << uint(capNum)
		} else {
			permittedHigh |= 1 << uint(capNum-32)
		}
	}

	// struct vfs_cap_data:
	// magic_etc
	// data[0].permitted
	// data[0].inheritable
	// data[1].permitted
	// data[1].inheritable
	binary.LittleEndian.PutUint32(data[4:8], permittedLow)
	binary.LittleEndian.PutUint32(data[8:12], 0)
	binary.LittleEndian.PutUint32(data[12:16], permittedHigh)
	binary.LittleEndian.PutUint32(data[16:20], 0)

	return data
}
