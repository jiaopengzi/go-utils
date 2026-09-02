//
// FilePath    : go-utils\domain_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 域名与 CSR 校验工具测试
//

package utils

import (
	"net"
	"testing"

	cert "github.com/jiaopengzi/cert/core"
)

// TestNormalizeDomainName 测试域名、国际化域名和 IP 地址的规范化结果.
func TestNormalizeDomainName(t *testing.T) {
	tests := []struct {
		name       string
		domainName string
		want       string
		wantErr    bool
	}{
		{name: "ascii_domain", domainName: "Blog.Example.COM.", want: "blog.example.com"},
		{name: "international_domain", domainName: "例子.测试", want: "xn--fsqu00a.xn--0zwm56d"},
		{name: "ipv4_without_port", domainName: "192.0.2.10", want: "192.0.2.10"},
		{name: "ipv6", domainName: "2001:db8::1", want: "2001:db8::1"},
		{name: "ipv4_with_port", domainName: "192.0.2.10:7364", wantErr: true},
		{name: "ipv4_with_path", domainName: "192.0.2.10:7364/blog", wantErr: true},
		{name: "ipv6_with_port", domainName: "[2001:db8::1]:7364", wantErr: true},
		{name: "missing", wantErr: true},
		{name: "missing_scheme", domainName: "example.com:443", wantErr: true},
		{name: "userinfo", domainName: "user@example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeDomainName(tt.domainName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeDomainName(%q) error = %v, wantErr %v", tt.domainName, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NormalizeDomainName(%q) = %q, want %q", tt.domainName, got, tt.want)
			}
		})
	}
}

// TestGetURLDomainName 测试 URL 主机名解析会忽略协议外的端口、路径等组成部分.
func TestGetURLDomainName(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{name: "ascii_domain", rawURL: "https://Blog.Example.COM.:7364", want: "blog.example.com"},
		{name: "international_domain", rawURL: "https://例子.测试:7364", want: "xn--fsqu00a.xn--0zwm56d"},
		{name: "ipv4_without_port", rawURL: "https://192.0.2.10", want: "192.0.2.10"},
		{name: "ipv4_with_port", rawURL: "https://192.0.2.10:7364", want: "192.0.2.10"},
		{name: "ipv4_with_path", rawURL: "https://192.0.2.10:7364/blog", want: "192.0.2.10"},
		{name: "ipv6", rawURL: "https://[2001:db8::1]:7364", want: "2001:db8::1"},
		{name: "missing", wantErr: true},
		{name: "missing_scheme", rawURL: "example.com:443", wantErr: true},
		{name: "userinfo", rawURL: "https://user@example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetURLDomainName(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetURLDomainName(%q) error = %v, wantErr %v", tt.rawURL, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("GetURLDomainName(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}

// TestValidateCSRDomain 测试 CSR 必须只携带与登记域名一致的 CN 和 SAN.
func TestValidateCSRDomain(t *testing.T) {
	createCSR := func(t *testing.T, commonName string, dnsNames []string, ipAddresses []net.IP) string {
		t.Helper()

		config := &cert.CSRConfig{
			Subject:      cert.Subject{CommonName: commonName},
			KeyAlgorithm: cert.KeyAlgorithmEd25519,
			SAN: cert.SANConfig{
				DNSNames:    dnsNames,
				IPAddresses: ipAddresses,
			},
		}
		if err := cert.GenerateCSR(config); err != nil {
			t.Fatalf("GenerateCSR() error = %v", err)
		}

		return config.CSR
	}

	t.Run("matching_international_domain", func(t *testing.T) {
		csr := createCSR(t, "xn--fsqu00a.xn--0zwm56d", []string{"xn--fsqu00a.xn--0zwm56d"}, nil)
		if err := ValidateCSRDomain(csr, "例子.测试"); err != nil {
			t.Fatalf("ValidateCSRDomain() error = %v", err)
		}
	})

	t.Run("matching_ip", func(t *testing.T) {
		csr := createCSR(t, "127.0.0.1", nil, []net.IP{net.ParseIP("127.0.0.1")})
		if err := ValidateCSRDomain(csr, "127.0.0.1"); err != nil {
			t.Fatalf("ValidateCSRDomain() error = %v", err)
		}
	})

	t.Run("mismatching_san", func(t *testing.T) {
		csr := createCSR(t, "example.com", []string{"other.example.com"}, nil)
		if err := ValidateCSRDomain(csr, "example.com"); err == nil {
			t.Fatal("ValidateCSRDomain() should reject mismatching SAN")
		}
	})
}
