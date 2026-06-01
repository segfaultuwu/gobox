package applets

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

func init() {
	Register("ping", "send ICMP echo requests", Ping)
}

func Ping(args []string) error {
	cfg, err := parsePingArgs(args)
	if err != nil {
		return err
	}

	ipaddr, err := net.ResolveIPAddr("ip4", cfg.host)
	if err != nil {
		return fmt.Errorf("cannot resolve %s: %w", cfg.host, err)
	}

	ip := ipaddr.IP.To4()
	if ip == nil {
		return fmt.Errorf("not an IPv4 address: %s", cfg.host)
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return fmt.Errorf("raw socket: %w", err)
	}
	defer syscall.Close(fd)

	tv := syscall.NsecToTimeval(cfg.timeout.Nanoseconds())
	_ = syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

	addr := &syscall.SockaddrInet4{}
	copy(addr.Addr[:], ip)

	fmt.Printf("PING %s (%s): %d data bytes\n", cfg.host, ip.String(), cfg.size)

	identifier := uint16(os.Getpid() & 0xffff)

	sent := 0
	received := 0

	for seq := 1; cfg.count == 0 || seq <= cfg.count; seq++ {
		packet := makeICMPEcho(identifier, uint16(seq), cfg.size)

		start := time.Now()

		if err := syscall.Sendto(fd, packet, 0, addr); err != nil {
			fmt.Printf("sendto: %v\n", err)
		} else {
			sent++
		}

		reply, from, err := recvICMP(fd, identifier, uint16(seq))
		if err != nil {
			fmt.Printf("Request timeout for icmp_seq %d\n", seq)
		} else {
			received++
			elapsed := time.Since(start)

			fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.2f ms\n",
				reply.dataLen,
				from.String(),
				seq,
				reply.ttl,
				float64(elapsed.Microseconds())/1000.0,
			)
		}

		if cfg.count == 0 || seq < cfg.count {
			time.Sleep(cfg.interval)
		}
	}

	loss := 0.0
	if sent > 0 {
		loss = float64(sent-received) / float64(sent) * 100
	}

	fmt.Printf("\n--- %s ping statistics ---\n", cfg.host)
	fmt.Printf("%d packets transmitted, %d received, %.0f%% packet loss\n", sent, received, loss)

	return nil
}

type pingConfig struct {
	host     string
	count    int
	size     int
	timeout  time.Duration
	interval time.Duration
}

type pingReply struct {
	ttl     int
	dataLen int
}

func parsePingArgs(args []string) (pingConfig, error) {
	cfg := pingConfig{
		count:    4,
		size:     56,
		timeout:  2 * time.Second,
		interval: time.Second,
	}

	var rest []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-c":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing count after -c")
			}

			value, err := strconv.Atoi(args[i+1])
			if err != nil || value < 0 {
				return cfg, fmt.Errorf("invalid count: %s", args[i+1])
			}

			cfg.count = value
			i++

		case "-s":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing size after -s")
			}

			value, err := strconv.Atoi(args[i+1])
			if err != nil || value < 0 {
				return cfg, fmt.Errorf("invalid size: %s", args[i+1])
			}

			cfg.size = value
			i++

		case "-W":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing timeout after -W")
			}

			value, err := strconv.Atoi(args[i+1])
			if err != nil || value < 0 {
				return cfg, fmt.Errorf("invalid timeout: %s", args[i+1])
			}

			cfg.timeout = time.Duration(value) * time.Second
			i++

		case "-i":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing interval after -i")
			}

			value, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || value < 0 {
				return cfg, fmt.Errorf("invalid interval: %s", args[i+1])
			}

			cfg.interval = time.Duration(value * float64(time.Second))
			i++

		default:
			rest = append(rest, arg)
		}
	}

	if len(rest) != 1 {
		return cfg, fmt.Errorf("usage: ping [-c count] [-s size] [-W timeout] [-i interval] host")
	}

	cfg.host = rest[0]

	return cfg, nil
}

func makeICMPEcho(identifier uint16, sequence uint16, payloadSize int) []byte {
	packet := make([]byte, 8+payloadSize)

	packet[0] = 8 // ICMP echo request
	packet[1] = 0 // code

	binary.BigEndian.PutUint16(packet[4:6], identifier)
	binary.BigEndian.PutUint16(packet[6:8], sequence)

	for i := 8; i < len(packet); i++ {
		packet[i] = byte(i & 0xff)
	}

	check := checksum(packet)
	binary.BigEndian.PutUint16(packet[2:4], check)

	return packet
}

func recvICMP(fd int, identifier uint16, sequence uint16) (pingReply, net.IP, error) {
	buf := make([]byte, 1500)

	for {
		n, from, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			return pingReply{}, nil, err
		}

		sa, ok := from.(*syscall.SockaddrInet4)
		if !ok {
			continue
		}

		if n < 20+8 {
			continue
		}

		ipHeaderLen := int(buf[0]&0x0f) * 4
		if n < ipHeaderLen+8 {
			continue
		}

		ttl := int(buf[8])

		icmp := buf[ipHeaderLen:n]

		icmpType := icmp[0]
		icmpCode := icmp[1]

		if icmpType != 0 || icmpCode != 0 {
			continue
		}

		replyID := binary.BigEndian.Uint16(icmp[4:6])
		replySeq := binary.BigEndian.Uint16(icmp[6:8])

		if replyID != identifier || replySeq != sequence {
			continue
		}

		ip := net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3])

		return pingReply{
			ttl:     ttl,
			dataLen: len(icmp) - 8,
		}, ip, nil
	}
}

func checksum(data []byte) uint16 {
	var sum uint32

	for len(data) > 1 {
		sum += uint32(binary.BigEndian.Uint16(data[:2]))
		data = data[2:]
	}

	if len(data) == 1 {
		sum += uint32(data[0]) << 8
	}

	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return ^uint16(sum)
}
