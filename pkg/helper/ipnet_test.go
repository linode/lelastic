package helper

import (
	"fmt"
	"testing"
)

func TestSubnetRaw(t *testing.T) {
	tests := []struct {
		prefix     string
		length     int
		subnet     uint32
		wantPrefix string
	}{
		{"10.0.0.0/16", 17, 1, "10.0.128.0/17"},
		{"10.0.0.0/8", 24, 1, "10.0.1.0/24"},
		{"10.0.0.0/8", 24, 256, "10.1.0.0/24"},
		{"10.0.0.0/8", 24, 65536, "11.0.0.0/24"},
		{"10.0.0.0/8", 25, 1, "10.0.0.128/25"},
		{"10.0.0.0/8", 25, 2, "10.0.1.0/25"},
		{"10.0.0.0/8", 32, 2, "10.0.0.2/32"},
		{"2000:1234::/32", 48, 8, "2000:1234:8::/48"},
		{"2000:1234::/32", 80, 3615, "2000:1234:0:0:e1f::/80"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Test: %s", tc.prefix), func(t *testing.T) {
			p, err := IPNetParse(tc.prefix)
			if err != nil {
				t.Errorf("Failed ! unable to parse %s", tc.prefix)
				return
			}

			ps, err := IPNetParse(tc.wantPrefix)
			if err != nil {
				t.Errorf("Failed ! unable to parse %s", tc.wantPrefix)
				return
			}

			subnet := p.SubnetRaw(tc.length, tc.subnet)

			if subnet.Plen() != ps.Plen() {
				t.Errorf("Failed ! got %d want %d", subnet.Plen(), ps.Plen())
				return
			}

			if !subnet.IP.Equal(ps.IP) {
				t.Errorf("Failed ! got %s want %s", subnet.String(), ps.String())
				return
			}

			t.Logf("Success !")
		})
	}
}

func TestSubnet(t *testing.T) {
	tests := []struct {
		prefix     string
		length     int
		subnet     uint32
		wantPrefix string
	}{
		{"10.0.0.0/16", 17, 1, "10.0.128.0/17"},
		{"10.0.0.0/8", 24, 1, "10.0.1.0/24"},
		{"10.0.0.0/8", 24, 256, "10.1.0.0/24"},
		{"10.0.0.0/8", 24, 65536, "11.0.0.0/24"},
		{"10.0.0.0/8", 25, 1, "10.0.0.128/25"},
		{"10.0.0.0/8", 25, 2, "10.0.1.0/25"},
		{"10.0.0.0/8", 32, 2, "10.0.0.2/32"},
		{"2000:1234::/32", 48, 17, "2000:1234:17::/48"},
		{"2000:1234::/32", 48, 256, "2000:1234:256::/48"},
		{"2000:1234::/32", 48, 1024, "2000:1234:1024::/48"},
		{"2000:1234::/32", 48, 10001, "2000:1235:1::/48"},
		{"2000:1234::/32", 56, 18, "2000:1234:0:1800::/56"},
		{"2000:1234::/32", 64, 1001, "2000:1234:0:1001::/64"},
		{"2000:1234::/32", 96, 256, "2000:1234::256:0:0/96"},
		{"2000:1234::/32", 128, 1001, "2000:1234::1001/128"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Test: %s", tc.prefix), func(t *testing.T) {
			p, err := IPNetParse(tc.prefix)
			if err != nil {
				t.Errorf("Failed ! unable to parse %s", tc.prefix)
				return
			}

			ps, err := IPNetParse(tc.wantPrefix)
			if err != nil {
				t.Errorf("Failed ! unable to parse %s", tc.wantPrefix)
				return
			}

			subnet := p.Subnet(tc.length, tc.subnet)

			if subnet.Plen() != ps.Plen() {
				t.Errorf("Failed ! got %d want %d", subnet.Plen(), ps.Plen())
				return
			}

			if !subnet.IP.Equal(ps.IP) {
				t.Errorf("Failed ! got %s want %s", subnet.String(), ps.String())
				return
			}

			t.Logf("Success !")
		})
	}
}
