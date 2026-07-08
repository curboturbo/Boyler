package networkservice


import (
	stnet "net"
	"fmt"
)


func hostCIDR(network string, host byte) (string, error) {
    ip, ipNet, err := stnet.ParseCIDR(network)
    if err != nil {
        return "", err
    }
    ip = ip.To4()
    if ip == nil {
        return "", fmt.Errorf("only IPv4 is supported")
    }
    ip = append(stnet.IP(nil), ip...)
    ip[3] = host
    ones, _ := ipNet.Mask.Size()
    return fmt.Sprintf("%s/%d", ip.String(), ones), nil
}

func isValidIPv4(ip string) bool {
    parsedIP := stnet.ParseIP(ip)
    if parsedIP == nil {return false}
    return parsedIP.To4() != nil
}

func SubnetSize(ipCidr string) (int, error) {
	_, ipNet, err := stnet.ParseCIDR(ipCidr)
	if err != nil {return 0, err}
	ones, bits := ipNet.Mask.Size()
	if bits == 0 {return 0, fmt.Errorf("invalid mask for %s", ipCidr)}
	hostBits := bits - ones
	return 1 << hostBits, nil
}

func UsableHosts(cidr string) (int, error) {
	total, err := SubnetSize(cidr)
	if err != nil {return 0, err}
	if total <= 2 {return 0, fmt.Errorf("empty net")}
	return total - 2, nil
}

func AllocateIpTable(num int) []byte {
	table := make([]byte, num)
	table[0],table[1] = 1,1
	return table
}