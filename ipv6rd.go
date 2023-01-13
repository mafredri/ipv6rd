package ipv6rd

import (
	"fmt"
	"net"

	"github.com/bonan/dhcp6rd"
	"github.com/pkg/errors"
)

// Tunnel represents the tunnel configuration for IPv6 6RD.
type Tunnel struct {
	Prefix          *net.IPNet // 6RD prefix.
	RelayPrefix     *net.IPNet // 6RD relay prefix.
	BorderRelay     string     // 6RD border relay (IPv4).
	DelegatedPrefix *net.IPNet // Delegated prefix.
	Address         string     // Tunnel address (anycast).
}

// ParseDHCP takes the current IP and DHCP OPTION_6RD and returns the
// tunnel configuraiton.
func ParseDHCP(ip string, dhcpOption string) (t Tunnel, err error) {
	opt, err := dhcp6rd.UnmarshalDhclient(dhcpOption)
	if err != nil {
		return t, errors.Wrap(err, "parse option-6rd failed")
	}

	if len(opt.Relay) == 0 {
		return t, errors.New("relay missing from option-6rd")
	}

	delegatedPrefix, err := opt.IPNet(net.ParseIP(ip))
	if err != nil {
		return t, errors.Wrap(err, "calculate delegated prefix failed")
	}

	_, relayPrefix, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip, opt.MaskLen))
	if err != nil {
		return t, errors.Wrap(err, "parse relay prefix failed")
	}

	return Tunnel{
		Prefix:          &net.IPNet{IP: opt.Prefix, Mask: net.CIDRMask(opt.PrefixLen, 128)},
		RelayPrefix:     relayPrefix,
		BorderRelay:     "::" + opt.Relay[0].String(),
		DelegatedPrefix: delegatedPrefix,
		Address:         delegatedPrefix.IP.String() + "/128", // Anycast.
	}, nil
}

// Network configuration for a (sub)net within the delegated prefix.
type Network struct {
	IPNet   *net.IPNet
	Address string
	IP      net.IP
}

// NetCalc calculates the (sub)net configuration based on the
// delegated prefix and a partial CIDR for the (sub)net, e.g.
// 0:0:0:10::1/64.
func NetCalc(prefix, netCIDR string) (n Network, err error) {
	_, dpNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return n, err
	}
	nIP, nNet, err := net.ParseCIDR(netCIDR)
	if err != nil {
		return n, err
	}

	netmask := make(net.IPMask, 16)
	ipmask := make(net.IPMask, 16)
	for i := 0; i < 16; i++ {
		netmask[i] = nNet.Mask[i] - dpNet.Mask[i]
		ipmask[i] = 0xFF &^ netmask[i]
	}

	for i := range nNet.Mask {
		nNet.IP[i] = (dpNet.IP[i] & dpNet.Mask[i]) + (nNet.IP[i] & netmask[i])
		nIP[i] = (dpNet.IP[i] & dpNet.Mask[i]) + (nNet.IP[i] & netmask[i]) + (nIP[i] & ipmask[i])
	}

	if !dpNet.Contains(nIP) {
		return n, errors.New(fmt.Sprintf("network IP %s escapes the delegated prefix %s", nIP.String(), dpNet.String()))
	}
	if !nNet.Contains(nIP) {
		return n, errors.New(fmt.Sprintf("network IP %s escapes the network prefix %s", nIP.String(), nNet.String()))
	}

	ones, _ := nNet.Mask.Size()
	return Network{
		IPNet:   nNet,
		Address: fmt.Sprintf("%s/%d", nIP.String(), ones),
		IP:      nIP,
	}, nil
}

// CIDRHasIP returns true if the IP belongs to the provided CIDR.
func CIDRHasIP(cidr, ip string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}
	ipp, _, err := net.ParseCIDR(ip)
	if err != nil {
		ipp = net.ParseIP(ip)
		if ipp == nil {
			return false, errors.Wrap(err, "invalid IP")
		}
	}
	return ipnet.Contains(ipp), nil
}
