package applets

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	dhcpServerPort = 67
	dhcpClientPort = 68

	dhcpDiscover = 1
	dhcpOffer    = 2
	dhcpRequest  = 3
	dhcpAck      = 5
	dhcpNak      = 6

	optSubnetMask = 1
	optRouter     = 3
	optDNS        = 6
	optReqIP      = 50
	optLeaseTime  = 51
	optMsgType    = 53
	optServerID   = 54
	optParamReq   = 55
	optEnd        = 255
)

type dhcpLease struct {
	IP       net.IP
	Mask     net.IP
	Router   net.IP
	DNS      []net.IP
	ServerID net.IP
}

func init() {
	Register("dhcp", "simple DHCP client", DHCP)
}

func DHCP(args []string) error {
	ifaceName := "eth0"

	if len(args) > 0 {
		ifaceName = args[0]
	}

	if ifaceName == "auto" {
		name, err := firstNonLoopbackInterface()
		if err != nil {
			return err
		}
		ifaceName = name
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return err
	}

	if err := linkUp(ifaceName); err != nil {
		return err
	}

	fmt.Printf("dhcp: using interface %s\n", ifaceName)

	xid := randomXID()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: dhcpClientPort,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := setupSocket(conn, ifaceName); err != nil {
		return err
	}

	if err := conn.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
		return err
	}

	discover := buildDHCPPacket(dhcpDiscover, xid, iface.HardwareAddr, nil, nil)

	fmt.Println("dhcp: sending discover")

	if _, err := conn.WriteToUDP(discover, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: dhcpServerPort,
	}); err != nil {
		return err
	}

	offer, err := readDHCPMessage(conn, xid, dhcpOffer)
	if err != nil {
		return err
	}

	fmt.Printf("dhcp: offered %s\n", offer.IP)

	request := buildDHCPPacket(dhcpRequest, xid, iface.HardwareAddr, offer.IP, offer.ServerID)

	fmt.Println("dhcp: sending request")

	if _, err := conn.WriteToUDP(request, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: dhcpServerPort,
	}); err != nil {
		return err
	}

	ack, err := readDHCPMessage(conn, xid, dhcpAck)
	if err != nil {
		return err
	}

	fmt.Printf("dhcp: lease acquired: %s\n", ack.IP)

	if err := applyLease(ifaceName, ack); err != nil {
		return err
	}

	fmt.Println("dhcp: configured network")

	return nil
}

func firstNonLoopbackInterface() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if len(iface.HardwareAddr) == 0 {
			continue
		}

		return iface.Name, nil
	}

	return "", fmt.Errorf("no usable network interface found")
}

func linkUp(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	return netlink.LinkSetUp(link)
}

func setupSocket(conn *net.UDPConn, iface string) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	var sockErr error

	err = raw.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
		if sockErr != nil {
			return
		}

		sockErr = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, iface)
	})

	if err != nil {
		return err
	}

	return sockErr
}

func randomXID() uint32 {
	var b [4]byte

	if _, err := rand.Read(b[:]); err != nil {
		return uint32(time.Now().UnixNano())
	}

	return binary.BigEndian.Uint32(b[:])
}

func buildDHCPPacket(msgType byte, xid uint32, mac net.HardwareAddr, requestedIP net.IP, serverID net.IP) []byte {
	packet := make([]byte, 240)

	packet[0] = 1 // BOOTREQUEST
	packet[1] = 1 // Ethernet
	packet[2] = 6 // MAC length

	binary.BigEndian.PutUint32(packet[4:8], xid)

	// Broadcast flag
	packet[10] = 0x80
	packet[11] = 0x00

	copy(packet[28:44], mac)

	// Magic cookie
	packet[236] = 99
	packet[237] = 130
	packet[238] = 83
	packet[239] = 99

	opts := []byte{}

	opts = appendOption(opts, optMsgType, []byte{msgType})

	if requestedIP != nil {
		opts = appendOption(opts, optReqIP, requestedIP.To4())
	}

	if serverID != nil {
		opts = appendOption(opts, optServerID, serverID.To4())
	}

	opts = appendOption(opts, optParamReq, []byte{
		optSubnetMask,
		optRouter,
		optDNS,
		optLeaseTime,
	})

	opts = append(opts, optEnd)

	return append(packet, opts...)
}

func appendOption(opts []byte, code byte, data []byte) []byte {
	if data == nil {
		return opts
	}

	opts = append(opts, code)
	opts = append(opts, byte(len(data)))
	opts = append(opts, data...)

	return opts
}

func readDHCPMessage(conn *net.UDPConn, xid uint32, wantedType byte) (dhcpLease, error) {
	buffer := make([]byte, 1500)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return dhcpLease{}, err
		}

		msg := buffer[:n]

		if len(msg) < 240 {
			continue
		}

		if binary.BigEndian.Uint32(msg[4:8]) != xid {
			continue
		}

		msgType, lease := parseDHCPMessage(msg)

		if msgType == dhcpNak {
			return dhcpLease{}, fmt.Errorf("dhcp nak received")
		}

		if msgType != wantedType {
			continue
		}

		return lease, nil
	}
}

func parseDHCPMessage(packet []byte) (byte, dhcpLease) {
	lease := dhcpLease{
		IP: net.IPv4(packet[16], packet[17], packet[18], packet[19]),
	}

	var msgType byte

	options := packet[240:]

	for i := 0; i < len(options); {
		code := options[i]
		i++

		if code == 0 {
			continue
		}

		if code == optEnd {
			break
		}

		if i >= len(options) {
			break
		}

		length := int(options[i])
		i++

		if i+length > len(options) {
			break
		}

		data := options[i : i+length]
		i += length

		switch code {
		case optMsgType:
			if len(data) >= 1 {
				msgType = data[0]
			}

		case optSubnetMask:
			if len(data) >= 4 {
				lease.Mask = net.IPv4(data[0], data[1], data[2], data[3])
			}

		case optRouter:
			if len(data) >= 4 {
				lease.Router = net.IPv4(data[0], data[1], data[2], data[3])
			}

		case optDNS:
			for j := 0; j+3 < len(data); j += 4 {
				lease.DNS = append(lease.DNS, net.IPv4(data[j], data[j+1], data[j+2], data[j+3]))
			}

		case optServerID:
			if len(data) >= 4 {
				lease.ServerID = net.IPv4(data[0], data[1], data[2], data[3])
			}
		}
	}

	return msgType, lease
}

func applyLease(ifaceName string, lease dhcpLease) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return err
	}

	mask := net.IPMask(lease.Mask.To4())
	ones, _ := mask.Size()

	if ones == 0 {
		ones = 24
	}

	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", lease.IP.String(), ones))
	if err != nil {
		return err
	}

	if err := netlink.AddrReplace(link, addr); err != nil {
		return err
	}

	if lease.Router != nil {
		route := netlink.Route{
			LinkIndex: link.Attrs().Index,
			Gw:        lease.Router,
		}

		_ = netlink.RouteDel(&route)

		if err := netlink.RouteReplace(&route); err != nil {
			return err
		}
	}

	if len(lease.DNS) > 0 {
		if err := writeResolvConf(lease.DNS); err != nil {
			return err
		}
	}

	return nil
}

func writeResolvConf(dns []net.IP) error {
	var b strings.Builder

	for _, server := range dns {
		b.WriteString("nameserver ")
		b.WriteString(server.String())
		b.WriteString("\n")
	}

	return os.WriteFile("/etc/resolv.conf", []byte(b.String()), 0644)
}
