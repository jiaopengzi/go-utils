package cert

import "errors"

// 错误定义.
var (
	// ErrNoPrivateKey 没有私钥.
	ErrNoPrivateKey = errors.New("no private key available")

	// ErrInvalidKeyType 无效的密钥类型.
	ErrInvalidKeyType = errors.New("invalid key type for this operation")

	// ErrInvalidCiphertext 无效的密文.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")

	// ErrInvalidSignature 无效的签名.
	ErrInvalidSignature = errors.New("invalid signature")
)
