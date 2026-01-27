package cert

import (
	"net"
	"testing"
)

// ============================================
// 工具函数测试
// ============================================

// TestSplitCommaList 测试逗号分隔列表解析.
func TestSplitCommaList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "正常输入",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "带空格",
			input:    "a, b, c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "空字符串",
			input:    "",
			expected: nil,
		},
		{
			name:     "只有空格",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "单个元素",
			input:    "single",
			expected: []string{"single"},
		},
		{
			name:     "带空元素",
			input:    "a,,b",
			expected: []string{"a", "b"},
		},
		{
			name:     "前后空格",
			input:    "  a  ,  b  ,  c  ",
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitCommaList(tt.input, ",")

			if len(result) != len(tt.expected) {
				t.Errorf("期望 %d 个元素, 实际 %d 个", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %s, 实际 %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestParseIPList 测试 IP 列表解析.
func TestParseIPList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "IPv4 地址",
			input:    "127.0.0.1,192.168.1.1",
			expected: 2,
		},
		{
			name:     "IPv6 地址",
			input:    "::1,fe80::1",
			expected: 2,
		},
		{
			name:     "混合地址",
			input:    "127.0.0.1,::1",
			expected: 2,
		},
		{
			name:     "无效地址被忽略",
			input:    "127.0.0.1,invalid,192.168.1.1",
			expected: 2,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseIPListFromStr(tt.input, ",")

			if len(result) != tt.expected {
				t.Errorf("期望 %d 个 IP, 实际 %d 个", tt.expected, len(result))
			}
		})
	}
}

// TestParseSANFromStr 测试从字符串解析 SAN.
func TestParseSANFromStr(t *testing.T) {
	san := ParseSANFromStr("a.com,b.com", "127.0.0.1,192.168.1.1")

	if len(san.DNSNames) != 2 {
		t.Errorf("期望 2 个 DNS 名称, 实际 %d 个", len(san.DNSNames))
	}
	if len(san.IPAddresses) != 2 {
		t.Errorf("期望 2 个 IP 地址, 实际 %d 个", len(san.IPAddresses))
	}
}

// TestParseSANFromStr_Empty 测试空输入.
func TestParseSANFromStr_Empty(t *testing.T) {
	san := ParseSANFromStr("", "")

	if len(san.DNSNames) != 0 {
		t.Errorf("期望 0 个 DNS 名称, 实际 %d 个", len(san.DNSNames))
	}
	if len(san.IPAddresses) != 0 {
		t.Errorf("期望 0 个 IP 地址, 实际 %d 个", len(san.IPAddresses))
	}
}

// ============================================
// SANConfig 测试
// ============================================

// TestSANConfig_DNSNames 测试 DNS 名称配置.
func TestSANConfig_DNSNames(t *testing.T) {
	caCfg := &CACertConfig{
		DaysValid:    365,
		Subject:      Subject{CommonName: "Test CA"},
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
	}
	if err := GenCACert(caCfg); err != nil {
		t.Fatalf("生成 CA 证书失败: %v", err)
	}

	serverCfg := &CASignedCertConfig{
		CACert:       caCfg.Cert,
		CAKey:        caCfg.Key,
		Name:         "server",
		Subject:      Subject{CommonName: "server"},
		DaysValid:    365,
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
		SAN: SANConfig{
			DNSNames: []string{"api.example.com", "*.example.com", "localhost"},
		},
		Usage: UsageServer,
	}
	if err := GenerateCASignedCert(serverCfg); err != nil {
		t.Fatalf("生成服务器证书失败: %v", err)
	}

	info, err := GetCertInfo(serverCfg.Cert)
	if err != nil {
		t.Fatalf("获取证书信息失败: %v", err)
	}

	if len(info.DNSNames) != 3 {
		t.Errorf("期望 3 个 DNS 名称, 实际 %d 个", len(info.DNSNames))
	}
}

// TestSANConfig_IPAddresses 测试 IP 地址配置.
func TestSANConfig_IPAddresses(t *testing.T) {
	caCfg := &CACertConfig{
		DaysValid:    365,
		Subject:      Subject{CommonName: "Test CA"},
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
	}
	if err := GenCACert(caCfg); err != nil {
		t.Fatalf("生成 CA 证书失败: %v", err)
	}

	serverCfg := &CASignedCertConfig{
		CACert:       caCfg.Cert,
		CAKey:        caCfg.Key,
		Name:         "server",
		Subject:      Subject{CommonName: "server"},
		DaysValid:    365,
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
		SAN: SANConfig{
			IPAddresses: []net.IP{
				net.ParseIP("127.0.0.1"),
				net.ParseIP("192.168.1.100"),
				net.ParseIP("::1"),
			},
		},
		Usage: UsageServer,
	}
	if err := GenerateCASignedCert(serverCfg); err != nil {
		t.Fatalf("生成服务器证书失败: %v", err)
	}

	info, err := GetCertInfo(serverCfg.Cert)
	if err != nil {
		t.Fatalf("获取证书信息失败: %v", err)
	}

	if len(info.IPAddresses) != 3 {
		t.Errorf("期望 3 个 IP 地址, 实际 %d 个", len(info.IPAddresses))
	}
}

// TestSANConfig_EmailAddresses 测试电子邮件地址配置.
func TestSANConfig_EmailAddresses(t *testing.T) {
	caCfg := &CACertConfig{
		DaysValid:    365,
		Subject:      Subject{CommonName: "Test CA"},
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
	}
	if err := GenCACert(caCfg); err != nil {
		t.Fatalf("生成 CA 证书失败: %v", err)
	}

	emailCfg := &CASignedCertConfig{
		CACert:       caCfg.Cert,
		CAKey:        caCfg.Key,
		Name:         "email",
		Subject:      Subject{CommonName: "user@example.com"},
		DaysValid:    365,
		KeyAlgorithm: KeyAlgorithmECDSA,
		ECDSACurve:   CurveP256,
		SAN: SANConfig{
			EmailAddrs: []string{"user@example.com", "admin@example.com"},
		},
		Usage: UsageEmailProtection,
	}
	if err := GenerateCASignedCert(emailCfg); err != nil {
		t.Fatalf("生成邮件证书失败: %v", err)
	}

	// 证书生成成功即可.
	if emailCfg.Cert == "" {
		t.Error("证书为空")
	}
}
