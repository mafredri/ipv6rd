# ipv6rd

Calculate 6RD prefixes from DHCP (IPv4) OPTION_6RD (212).

Either use this library or the included `6rdcalc` command.

## `6rdcalc`

This tool is useful for scripting, e.g. when used in a dhclient hook.

There's `ipv6calc`, why create a new tool? This tool is written in Go, so it's easily portable to multiple architectures. Since it also does `OPTION_6RD` parsing, it saves you from calculating the initial prefix in sh/ash/Bash and lets you easily create new `/64` prefixes for your networks (LAN, device, docker, etc.).

### Installation

```shell
go install github.com/mafredri/ipv6rd/cmd/6rdcalc@latest
```

Other platforms:

```shell
git clone https://github.com/mafredri/ipv6rd.git
cd ipv6rd
GOOS=linux GOARCH=mips go build ./cmd/6rdcalc
scp 6rdcalc unifi-security-gateway.local:
ssh ubnt@unifi-security-gateway.local
ubnt@unifi-security-gateway:~$ ./6rdcalc
```

### Tunnel configuration

Generating the tunnel configuration from the given 6RD configuration:

```shell
6rdcalc tunnel 84.240.100.100 "14 38 8193 8195 62464 0 0 0 0 0 84.251.255.254"
```

The value for `IP` comes from `$new_ip_address` and the value for `OPTION 6RD` comes from `$new_option_6rd`, both available in a dhclient hook.

Example output:

```
PREFIX="2001:2003:f400::/38"
RELAY_PREFIX="84.240.0.0/14"
BORDER_RELAY="::84.251.255.254"
DELEGATED_PREFIX="2001:2003:f464:6400::/56"
TUNNEL_ADDRESS="2001:2003:f464:6400::/128"
```

Usage in scripts:

```shell
eval "$(6rdcalc tunnel 84.240.100.100 "14 38 8193 8195 62464 0 0 0 0 0 84.251.255.254")"
echo "$DELEGATED_PREFIX"
```

### Network configuration

Use the network subcommand to slice a `/64` out of your delegated prefix.

```shell
6rdcalc network 2001:2003:f464:6400::/56 0:0:0:10::1/64
```

The `DELEGATED PREFIX` comes from the tunnel configuration and the `IP` is the wanted network IP in CIDR notation, without the delegated prefix.

Example output:

```shell
NETWORK="2001:2003:f464:6410::/64"
ADDRESS="2001:2003:f464:6410::1/64"
IP="2001:2003:f464:6410::1"
```

Usage in scripts:

```shell
eval "$(6rdcalc network 2001:2003:f464:6400::/56 0:0:0:10::1/64)"
echo "$IP"
```

### Verification

Use the `contains` subcommand to figure out if your existing configuration needs updating.

```shell
DELEGATED_PREFIX="2001:2003:f464:6400::/56"
ADDRESS="2001:2003:f464:6410::1/64"
IP="2001:2003:f464:6410::1"

$ 6rdcalc contains "$DELEGATED_PREFIX" "$IP"
true
$ 6rdcalc contains "$DELEGATED_PREFIX" "$ADDRESS"
true
6rdcalc contains "$DELEGATED_PREFIX" "2001:2003:f464:0000::1"
false
```

## Configuration of dhclient

```
option option-6rd code 212 = { integer 8, integer 8, ip6-address, array of ip-address };
# OR
option option-6rd code 212 = {
	integer 8, integer 8,
	unsigned integer 16, unsigned integer 16, unsigned integer 16, unsigned integer 16,
	unsigned integer 16, unsigned integer 16, unsigned integer 16, unsigned integer 16,
	array of ip-address
};

interface "eth0" {
	also request option-6rd;
}
```

## Resources

- [RFC5969 - IPv6 Rapid Deployment on IPv4 Infrastructures (6rd)](https://tools.ietf.org/html/rfc5969)
- [bonan/dhcp6rd](https://github.com/bonan/dhcp6rd)
