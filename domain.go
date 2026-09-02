//
// FilePath    : go-utils\domain.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 域名与 CSR 校验工具
//

package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/netip"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

// NormalizeDomainName 规范化域名或 IP 地址, 统一为可持久化和比较的形式.
//   - domainName, 域名或 IP 地址.
//
// 返回值 string, 小写 ASCII 域名或规范化 IP 地址; error, 域名不合法时非 nil.
func NormalizeDomainName(domainName string) (string, error) {
	domainName = strings.TrimSpace(domainName)
	domainName = strings.TrimSuffix(domainName, ".")
	if domainName == "" {
		return "", fmt.Errorf("域名不能为空")
	}

	if ipAddress, err := netip.ParseAddr(domainName); err == nil {
		return ipAddress.String(), nil
	}

	asciiDomainName, err := idna.Lookup.ToASCII(domainName)
	if err != nil {
		return "", fmt.Errorf("国际化域名转换失败: %w", err)
	}

	return strings.ToLower(asciiDomainName), nil
}

// GetURLDomainName 从带协议的 URL 中提取并规范化主机名.
//   - rawURL, 包含协议和主机名的 URL, 可以携带端口、路径、查询参数或片段.
//
// 返回值 string, 小写 ASCII 域名或规范化 IP 地址; error, URL 无效时非 nil.
func GetURLDomainName(rawURL string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Hostname() == "" {
		return "", fmt.Errorf("URL 无效, 需要包含协议和主机名")
	}
	if parsedURL.User != nil {
		return "", fmt.Errorf("URL 不能包含用户信息")
	}

	domainName, err := NormalizeDomainName(parsedURL.Hostname())
	if err != nil {
		return "", fmt.Errorf("规范化 URL 主机名失败: %w", err)
	}

	return domainName, nil
}

// ValidateCSRDomain 校验 CSR 签名和其中唯一的域名标识是否与登记域名一致.
//   - csrPEM, PEM 编码的证书签名请求.
//   - domainName, 需要绑定的登记域名或 IP 地址.
//
// 返回值 error, CSR 无效或 CN/SAN 与登记域名不一致时非 nil.
func ValidateCSRDomain(csrPEM string, domainName string) error {
	normalizedDomainName, err := NormalizeDomainName(domainName)
	if err != nil {
		return fmt.Errorf("规范化登记域名失败: %w", err)
	}

	certificateRequest, err := parseCertificateRequest(csrPEM)
	if err != nil {
		return err
	}

	normalizedCommonName, err := NormalizeDomainName(certificateRequest.Subject.CommonName)
	if err != nil || normalizedCommonName != normalizedDomainName {
		return fmt.Errorf("CSR CommonName 与登记域名不一致")
	}

	if len(certificateRequest.EmailAddresses) != 0 {
		return fmt.Errorf("CSR 不允许包含邮箱 SAN")
	}

	if expectedIP, parseErr := netip.ParseAddr(normalizedDomainName); parseErr == nil {
		return validateIPCSRDomain(certificateRequest, expectedIP)
	}

	return validateDNSCSRDomain(certificateRequest, normalizedDomainName)
}

// parseCertificateRequest 解析并验证 PEM 编码 CSR 的签名.
//   - csrPEM, PEM 编码的证书签名请求.
//
// 返回值 *x509.CertificateRequest, 已验证签名的 CSR; error, CSR 格式或签名无效时非 nil.
func parseCertificateRequest(csrPEM string) (*x509.CertificateRequest, error) {
	block, rest := pem.Decode([]byte(csrPEM))
	if block == nil || block.Type != "CERTIFICATE REQUEST" || len(strings.TrimSpace(string(rest))) != 0 {
		return nil, fmt.Errorf("CSR PEM 格式无效")
	}

	certificateRequest, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析 CSR 失败: %w", err)
	}

	if err = certificateRequest.CheckSignature(); err != nil {
		return nil, fmt.Errorf("CSR 签名校验失败: %w", err)
	}

	return certificateRequest, nil
}

// validateIPCSRDomain 校验 CSR 中唯一的 DNS 与 IP SAN 都等于登记 IP.
//   - certificateRequest, 已验证签名的 CSR.
//   - expectedIP, 经过规范化的登记 IP.
//
// 返回值 error, CSR SAN 与登记 IP 不一致时非 nil.
func validateIPCSRDomain(certificateRequest *x509.CertificateRequest, expectedIP netip.Addr) error {
	if len(certificateRequest.DNSNames) != 1 || len(certificateRequest.IPAddresses) != 1 || certificateRequest.DNSNames[0] != expectedIP.String() || certificateRequest.IPAddresses[0].String() != expectedIP.String() {
		return fmt.Errorf("CSR IP SAN 与登记域名不一致")
	}

	return nil
}

// validateDNSCSRDomain 校验 CSR 中唯一的 DNS SAN 等于登记域名.
//   - certificateRequest, 已验证签名的 CSR.
//   - domainName, 经过规范化的登记域名.
//
// 返回值 error, CSR SAN 与登记域名不一致时非 nil.
func validateDNSCSRDomain(certificateRequest *x509.CertificateRequest, domainName string) error {
	if len(certificateRequest.IPAddresses) != 0 || len(certificateRequest.DNSNames) != 1 {
		return fmt.Errorf("CSR DNS SAN 与登记域名不一致")
	}

	normalizedDNSName, err := NormalizeDomainName(certificateRequest.DNSNames[0])
	if err != nil || normalizedDNSName != domainName {
		return fmt.Errorf("CSR DNS SAN 与登记域名不一致")
	}

	return nil
}
