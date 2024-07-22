package helper

import (
	"encoding/json"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type IPNet net.IPNet

func IPNetParse(s string) (*IPNet, error) {
	ip, subnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("Parse failed: %v", err)
	}

	i := &IPNet{
		IP:   ip,
		Mask: subnet.Mask,
	}

	return i, nil
}

// IPNetFromAddr reads net.Addr and converts it to IPNet
func IPNetFromAddr(a net.Addr) (*IPNet, error) {
	_, p, err := net.ParseCIDR(a.String())
	if err != nil {
		return nil, err
	}

	return &IPNet{
		IP:   p.IP,
		Mask: p.Mask,
	}, nil
}

// Set parses a string into an IPNet. this is used for the flag interface
func (i *IPNet) Set(v string) error {
	ip, subnet, err := net.ParseCIDR(v)
	if err != nil {
		return fmt.Errorf("Parse failed: %v", err)
	}

	*i = IPNet{
		IP:   ip,
		Mask: subnet.Mask,
	}

	return nil
}

func (i *IPNet) String() string {
	j, _ := i.Mask.Size()
	return fmt.Sprintf("%v/%v", i.IP, j)
}

func (i *IPNet) Plen() uint32 {
	j, _ := i.Mask.Size()
	return uint32(j)
}

func (i *IPNet) Equal(b IPNet) bool {
	bm, _ := b.Mask.Size()
	im, _ := i.Mask.Size()
	if i.IP.Equal(b.IP) && bm == im {
		return true
	}
	return false
}

func (i *IPNet) Find(b []IPNet) int {
	for j, bb := range b {
		if i.Equal(bb) {
			return j
		}
	}
	return -1
}

func (i *IPNet) Copy() *IPNet {
	c := IPNet{
		IP:   append([]byte{}, i.IP...),
		Mask: net.CIDRMask(int(i.Plen()), 128),
	}
	return &c
}

// Subnet is an IP Subnet calculator basically. supporting ipv4/6
func (i *IPNet) Subnet(length int, subnet uint32) *IPNet {
	i4 := i.IP.To4()
	if i4 != nil {
		return i.SubnetRaw(length, subnet)
	}
	return i.SubnetRaw(length, HumanHexConvert(subnet))
}

// SubnetRaw sets the IP Subnet to a specific subnet/prefix length
// i.e. 2600:2000::/32.SubnetRaw(48, 10) -> 2600:2000:10::/48)
func (i *IPNet) SubnetRaw(length int, subnet uint32) *IPNet {
	ip := IPNet{
		IP:   append([]byte{}, i.IP...),
		Mask: net.CIDRMask(length, 128),
	}

	// if IPv4 convert to 32 bit slice
	if i4 := i.IP.To4(); i4 != nil {
		// this a IPv4
		ip.IP = append([]byte{}, i4...)
	}

	// find offest for ipv6. (only dealing with a uint32, so 32 bit run out before ipv6 does
	offset := (length / 8) - 4
	bo := 0

	// shift subnet byte if only a fraction of a byte is split
	if length%8 > 0 {
		bo = 8 - (length % 8)
		offset += 1
	}

	b := ItoB(subnet << bo)

	for i := range b {
		if offset+i < 0 {
			continue
		} else if offset+i >= len(ip.IP) {
			// if this fires, goLinject will panic
			log.WithFields(log.Fields{"Topic": "SubnetRaw"}).Errorf("offset of %d is greater than IP length", offset+i)
		}
		ip.IP[offset+i] += b[i]
	}

	return &ip
}

// UnmarshalYAML interface for IPNet
func (i *IPNet) UnmarshalYAML(n *yaml.Node) error {
	t, err := IPNetParse(n.Value)
	if err != nil {
		return err
	}
	*i = *t
	return nil
}

// UnmarshalJSON interface for an IPNet
func (i *IPNet) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Decode and ensure we maintain the full IP address in the IPNet that we return.
	t, err := IPNetParse(s)
	if err != nil {
		return err
	}
	*i = *t
	return nil
}

// MarshalJSON interface for IPNet
func (i *IPNet) MarshalJSON() ([]byte, error) {
	return []byte(`"` + i.String() + `"`), nil
}
