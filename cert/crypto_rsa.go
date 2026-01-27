package cert

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
)

// RSACryptoOperator RSA 证书加密操作器.
type RSACryptoOperator struct {
	cert *Certificate
}

// GetKeyAlgorithm 获取密钥算法.
func (r *RSACryptoOperator) GetKeyAlgorithm() KeyAlgorithm {
	return KeyAlgorithmRSA
}

// GetCertificate 获取底层证书.
func (r *RSACryptoOperator) GetCertificate() *Certificate {
	return r.cert
}

// Sign 使用 RSA 私钥对数据进行签名(PKCS1v15 with SHA-256).
func (r *RSACryptoOperator) Sign(data []byte) ([]byte, error) {
	// 检查是否有私钥.
	if !r.cert.HasPrivateKey() {
		return nil, ErrNoPrivateKey
	}

	// 获取 RSA 私钥.
	rsaKey, ok := r.cert.privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKeyType
	}

	// 对数据进行哈希.
	hashed := sha256.Sum256(data)

	// 签名.
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("rsa sign failed: %w", err)
	}

	return signature, nil
}

// Verify 使用 RSA 公钥验证签名(PKCS1v15 with SHA-256).
func (r *RSACryptoOperator) Verify(data []byte, signature []byte) error {
	// 获取 RSA 公钥.
	pubKey, ok := r.cert.cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return ErrInvalidKeyType
	}

	// 对数据进行哈希.
	hashed := sha256.Sum256(data)

	// 验证签名.
	err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return fmt.Errorf("rsa verify failed: %w", err)
	}

	return nil
}

// HybridEncrypt 混合加密: 使用 AES 加密数据, 使用 RSA 加密 AES 密钥.
// 返回密文和 nonce, 如果 plaintext 为 nil, 则返回 nil 密文和有效的 nonce.
func (r *RSACryptoOperator) HybridEncrypt(plaintext []byte) ([]byte, []byte, error) {
	// 生成随机 AES 密钥.
	aesKey := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, nil, fmt.Errorf("generate aes key failed: %w", err)
	}

	// 使用 AES-GCM 加密数据.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create aes cipher failed: %w", err)
	}

	// 创建 GCM 模式.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("create gcm failed: %w", err)
	}

	// 生成随机 nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, randErr := io.ReadFull(rand.Reader, nonce); randErr != nil {
		return nil, nil, fmt.Errorf("generate nonce failed: %w", randErr)
	}

	// 如果 plaintext 为 nil, 只返回 nonce.
	if plaintext == nil {
		return nil, nonce, nil
	}

	// 加密数据.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 使用 RSA-OAEP 加密 AES 密钥.
	pubKey, ok := r.cert.cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, nil, ErrInvalidKeyType
	}

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, aesKey, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt aes key failed: %w", err)
	}

	// 组合加密包: [加密密钥长度(2字节)][加密密钥][nonce][加密数据].
	result := make([]byte, 2+len(encryptedKey)+len(nonce)+len(ciphertext))
	result[0] = byte(len(encryptedKey) >> 8)
	result[1] = byte(len(encryptedKey))
	copy(result[2:], encryptedKey)
	copy(result[2+len(encryptedKey):], nonce)
	copy(result[2+len(encryptedKey)+len(nonce):], ciphertext)

	return result, nonce, nil
}

// HybridDecrypt 混合解密.
func (r *RSACryptoOperator) HybridDecrypt(encryptedPackage []byte) ([]byte, error) {
	// 检查加密包长度.
	if len(encryptedPackage) < 2 {
		return nil, ErrInvalidCiphertext
	}

	// 解析加密包.
	keyLen := int(encryptedPackage[0])<<8 | int(encryptedPackage[1])
	if len(encryptedPackage) < 2+keyLen {
		return nil, ErrInvalidCiphertext
	}

	encryptedKey := encryptedPackage[2 : 2+keyLen]
	remaining := encryptedPackage[2+keyLen:]

	// 使用 RSA-OAEP 解密 AES 密钥.
	if !r.cert.HasPrivateKey() {
		return nil, ErrNoPrivateKey
	}

	rsaKey, ok := r.cert.privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKeyType
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt aes key failed: %w", err)
	}

	// 使用 AES-GCM 解密数据.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher failed: %w", err)
	}

	// 创建 GCM 模式.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm failed: %w", err)
	}

	// 检查剩余数据长度.
	if len(remaining) < gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	nonce := remaining[:gcm.NonceSize()]
	ciphertext := remaining[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt data failed: %w", err)
	}

	return plaintext, nil
}
