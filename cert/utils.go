package cert

import (
	"net"
	"strings"
)

// SplitCommaList 按照指定分隔符拆分逗号分隔的字符串列表.
func SplitCommaList(value string, delimiter string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, delimiter)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

// ParseIPListFromStr 从逗号分隔的字符串解析 IP 地址列表.
func ParseIPListFromStr(value string, delimiter string) []net.IP {
	items := SplitCommaList(value, delimiter)
	ips := make([]net.IP, 0, len(items))

	for _, item := range items {
		if ip := net.ParseIP(item); ip != nil {
			ips = append(ips, ip)
		}
	}

	return ips
}

// ParseSANFromStr 从逗号分隔的字符串解析 SAN 配置.
func ParseSANFromStr(dnsNames, ipAddrs string) SANConfig {
	return SANConfig{
		DNSNames:    SplitCommaList(dnsNames, ","),
		IPAddresses: ParseIPListFromStr(ipAddrs, ","),
	}
}
