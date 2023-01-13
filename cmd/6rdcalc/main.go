package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mafredri/ipv6rd"

	"github.com/alecthomas/kingpin"
)

func main() {
	var (
		tunnel = kingpin.Command("tunnel", "Calculate 6RD tunnel configuration")
		ip     = tunnel.Arg("IP", "IPv4 Address").Required().IP()
		option = tunnel.Arg("OPTION 6RD", "Value of DHCP OPTION_6RD (212)").Required().String()

		network         = kingpin.Command("network", "Calculate network configuration based on the delegated 6RD prefix")
		delegatedPrefix = network.Arg("DELEGATED PREFIX", "Delegated 6RD prefix").Required().String()
		networkCIDR     = network.Arg("IP", "Network IP in CIDR notation (without prefix), e.g. 0:0:0:10::1/64").Required().String()

		contains       = kingpin.Command("contains", "Check if the network prefix contains the IP")
		containsPrefix = contains.Arg("PREFIX", "The network prefix that defines the bounds").Required().String()
		containsCIDR   = contains.Arg("CIDR", "The CIDR (or IP) that will be checked against prefix").Required().String()
		containsQuiet  = contains.Flag("quiet", "Do not print anything to stdout, exit with 0 if the prefix contains the IP, otherwise 1").Short('q').Bool()
	)
	kingpin.CommandLine.Help = "This tool can be used to calculate the 6RD tunnel configuration from IPv4 DHCP servers that offer OPTION_6RD (212) and new network configurations within the 6RD delegated prefix."
	kingpin.CommandLine.HelpFlag.Short('h')

	switch kingpin.Parse() {
	case tunnel.FullCommand():
		c, err := ipv6rd.ParseDHCP(strings.TrimSpace(ip.String()), *option)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Printf("PREFIX=%q\n", c.Prefix.String())
		fmt.Printf("RELAY_PREFIX=%q\n", c.RelayPrefix)
		fmt.Printf("BORDER_RELAY=%q\n", c.BorderRelay)
		fmt.Printf("DELEGATED_PREFIX=%q\n", c.DelegatedPrefix)
		fmt.Printf("TUNNEL_ADDRESS=%q\n", c.Address)
	case network.FullCommand():
		c, err := ipv6rd.NetCalc(strings.TrimSpace(*delegatedPrefix), strings.TrimSpace(*networkCIDR))
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Printf("NETWORK=%q\n", c.IPNet.String())
		fmt.Printf("ADDRESS=%q\n", c.Address)
		fmt.Printf("IP=%q\n", c.IP.String())
	case contains.FullCommand():
		c, err := ipv6rd.CIDRHasIP(strings.TrimSpace(*containsPrefix), strings.TrimSpace(*containsCIDR))
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if *containsQuiet {
			if c {
				os.Exit(0)
			}
			os.Exit(1)
		}
		fmt.Printf("%t\n", c) // Print true or false.
	}
}
