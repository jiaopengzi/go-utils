//
// FilePath    : go-utils\req\crypto_json_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : JSON 加密解密功能单元测试
//

package req

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
	"time"
)

// generateTestCertECDSA 生成测试用的 ECDSA 自签名证书和私钥 PEM
func generateTestCertECDSA(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()

	// 生成 ECDSA 私钥
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA private key: %v", err)
	}

	certPEM = generateCertWithKey(t, &priv.PublicKey, priv)

	// 编码私钥为 PEM
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal ECDSA private key: %v", err)
	}
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDER,
	})

	return certPEM, string(keyPEMBytes)
}

// generateTestCertRSA 生成测试用的 RSA 自签名证书和私钥 PEM
func generateTestCertRSA(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()

	// 生成 RSA 私钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA private key: %v", err)
	}

	certPEM = generateCertWithKey(t, &priv.PublicKey, priv)

	// 编码私钥为 PEM
	keyDER := x509.MarshalPKCS1PrivateKey(priv)
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyDER,
	})

	return certPEM, string(keyPEMBytes)
}

// generateTestCertEd25519 生成测试用的 Ed25519 自签名证书和私钥 PEM
func generateTestCertEd25519(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()

	// 生成 Ed25519 密钥对
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate Ed25519 private key: %v", err)
	}

	certPEM = generateCertWithKey(t, pub, priv)

	// 编码私钥为 PEM (使用 PKCS8 格式)
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal Ed25519 private key: %v", err)
	}
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	return certPEM, string(keyPEMBytes)
}

// generateCertWithKey 使用给定的公钥和私钥生成自签名证书 PEM
func generateCertWithKey(t *testing.T, pub crypto.PublicKey, priv crypto.Signer) string {
	t.Helper()

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	// 编码证书为 PEM
	certPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return string(certPEMBytes)
}

// certGenerator 证书生成器类型
type certGenerator struct {
	name     string
	generate func(t *testing.T) (certPEM, keyPEM string)
}

// allCertGenerators 所有证书类型的生成器
var allCertGenerators = []certGenerator{
	{"ECDSA", generateTestCertECDSA},
	{"RSA", generateTestCertRSA},
	{"Ed25519", generateTestCertEd25519},
}

func TestEncryptDecryptJSON_RoundTrip(t *testing.T) {
	type Payload struct {
		Name  string
		Age   int
		Notes []string
	}
	orig := Payload{
		Name:  "alice",
		Age:   30,
		Notes: []string{"note1", "note2"},
	}

	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			certPEM, keyPEM := cg.generate(t)

			enc, nonce, err := EncryptJSON(orig, certPEM)
			if err != nil {
				t.Fatalf("EncryptJSON error: %v", err)
			}
			if enc == "" || nonce == "" {
				t.Fatalf("expected non-empty enc and nonce")
			}

			var out Payload
			if err := DecryptJSON(enc, keyPEM, &out); err != nil {
				t.Fatalf("DecryptJSON error: %v", err)
			}
			if !reflect.DeepEqual(orig, out) {
				t.Fatalf("roundtrip mismatch: got %+v want %+v", out, orig)
			}
		})
	}
}

func TestEncryptJSON_NilData(t *testing.T) {
	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			certPEM, _ := cg.generate(t)

			enc, nonce, err := EncryptJSON(nil, certPEM)
			if err != nil {
				t.Fatalf("EncryptJSON nil data error: %v", err)
			}
			if enc != "" {
				t.Fatalf("expected empty ciphertext for nil data, got: %s", enc)
			}
			if nonce == "" {
				t.Fatalf("expected non-empty nonce for nil data")
			}
		})
	}
}

func TestDecryptJSON_EmptyEncryptedData(t *testing.T) {
	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			_, keyPEM := cg.generate(t)

			type Payload struct{ X string }
			var out Payload

			// 空字符串密文应该直接返回 nil
			err := DecryptJSON("", keyPEM, &out)
			if err != nil {
				t.Fatalf("expected nil error for empty encrypted data, got: %v", err)
			}
		})
	}
}

func TestDecryptJSON_NonPointerDst(t *testing.T) {
	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			_, keyPEM := cg.generate(t)

			type Payload struct{ X string }
			var nonPtr Payload

			err := DecryptJSON("", keyPEM, nonPtr)
			if err == nil {
				t.Fatalf("expected error for non-pointer dst")
			}
			if err.Error() == "" {
				t.Fatalf("expected descriptive error for non-pointer dst")
			}
		})
	}
}

func TestEncryptJSON_InvalidCert(t *testing.T) {
	type Payload struct{ X string }

	_, _, err := EncryptJSON(Payload{X: "x"}, "invalid-cert-pem")
	if err == nil {
		t.Fatalf("expected error for invalid cert PEM")
	}
}

func TestDecryptJSON_InvalidBase64(t *testing.T) {
	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			_, keyPEM := cg.generate(t)

			type Payload struct{ X string }
			var out Payload

			err := DecryptJSON("not-valid-base64!!!", keyPEM, &out)
			if err == nil {
				t.Fatalf("expected error for invalid base64 ciphertext")
			}
		})
	}
}

func TestDecryptJSON_WrongKeyFails(t *testing.T) {
	type Payload struct{ X string }
	orig := Payload{X: "secret"}

	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			certPEM1, _ := cg.generate(t)
			_, keyPEM2 := cg.generate(t)

			enc, _, err := EncryptJSON(orig, certPEM1)
			if err != nil {
				t.Fatalf("EncryptJSON error: %v", err)
			}

			var out Payload
			if err := DecryptJSON(enc, keyPEM2, &out); err == nil {
				t.Fatalf("expected decrypt error with wrong key")
			}
		})
	}
}
