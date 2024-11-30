package plugin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"os"
)

func loadGeoipFromFile(fp string, ccm map[string]struct{}) (map[string][]*ipCIDR, error) {
	defer quickGC()()
	file, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	geoip, err := loadGeoip(file, ccm)
	if err != nil {
		return nil, err
	}
	return geoip, nil
}

func loadGeoip(b []byte, ccm map[string]struct{}) (map[string][]*ipCIDR, error) {
	var list GeoIPList
	err := proto.Unmarshal(b, &list)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]*ipCIDR, len(ccm))
	for cc := range ccm {
		m[cc] = []*ipCIDR{}
	}
	for _, unit := range list.Entry {
		_, ok := m[unit.CountryCode]
		if ok {
			for _, cidr := range unit.Cidr {
				ipC, err := newIpCIDR(cidr)
				if err != nil {
					continue
				}
				m[unit.CountryCode] = append(m[unit.CountryCode], ipC)
			}
		}
	}
	return m, nil
}

func newIpCIDR(cidr *CIDR) (*ipCIDR, error) {
	if cidr.Prefix > 255 {
		return nil, errors.New("invalid prefix")
	}
	ip := net.IP(cidr.Ip)
	m := ipCIDR{
		prefix: uint8(cidr.Prefix),
	}

	if ip4 := ip.To4(); ip4 != nil {
		// IPv4
		m.a = uint64(binary.BigEndian.Uint32(ip4))
	} else {
		// IPv6
		m.v6 = true
		ip6 := ip.To16()
		m.a = binary.BigEndian.Uint64(ip6[:8])
		m.b = binary.BigEndian.Uint64(ip6[8:])
	}
	return &m, nil
}

// if ipv4 a[32bit-nil,32bit-nil] b [32bit-nil,32bit-not_nil]
// if ipv6 ...
type ipCIDR struct {
	a      uint64
	b      uint64
	prefix uint8
	v6     bool
}

func (cidr *ipCIDR) String() string {
	var ipStr string
	if !cidr.v6 {
		// IPv4
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, uint32(cidr.a))
		ipStr = net.IP(ip).String()
	} else {
		// IPv6
		ip := make([]byte, 16)
		binary.BigEndian.PutUint64(ip[:8], cidr.a)
		binary.BigEndian.PutUint64(ip[8:], cidr.b)
		ipStr = net.IP(ip).String()
	}
	return fmt.Sprintf("%s/%d", ipStr, cidr.prefix)
}

func (cidr *ipCIDR) Contains(ip net.IP) bool {
	if !cidr.v6 {
		ip4 := ip.To4()
		if ip4 == nil {
			return false
		}
		// IPv4
		ipVal := uint64(binary.BigEndian.Uint32(ip4))
		mask := ^uint64(0) << (32 - cidr.prefix)
		return (ipVal & mask) == (cidr.a & mask)
	} else {
		// IPv6
		ip6 := ip.To16()
		if ip6 == nil {
			return false
		}
		a := binary.BigEndian.Uint64(ip6[:8])
		b := binary.BigEndian.Uint64(ip6[8:])

		if cidr.prefix > 64 {
			// Prefix spans both parts of the IPv6 address
			maskA := ^uint64(0)
			maskB := ^uint64(0) << (128 - cidr.prefix)
			return (a&maskA) == (cidr.a&maskA) && ((b & maskB) == (cidr.b & maskB))
		} else {
			// Prefix only affects the first 64 bits
			maskA := ^uint64(0) << (64 - cidr.prefix)
			return (a & maskA) == (cidr.a & maskA)
		}
	}
}
