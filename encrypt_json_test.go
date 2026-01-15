//
// FilePath    : go-utils\encrypt_json_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 单测
//

package utils

import (
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"testing"
)

// randomKeyB64 生成一个随机的 Base64 编码密钥, 长度为 n 字节
func randomKeyB64(t *testing.T, n int) string {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
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

	key := randomKeyB64(t, 32)
	enc, nonce, err := EncryptJSON(orig, key)
	if err != nil {
		t.Fatalf("EncryptJSON error: %v", err)
	}
	if enc == "" || nonce == "" {
		t.Fatalf("expected non-empty enc and nonce")
	}

	var out Payload
	if err := DecryptJSON(enc, key, &out); err != nil {
		t.Fatalf("DecryptJSON error: %v", err)
	}
	if !reflect.DeepEqual(orig, out) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", out, orig)
	}
}

func TestDecryptJSON_NonPointerDst(t *testing.T) {
	type Payload struct{ X string }
	key := randomKeyB64(t, 32)

	var nonPtr Payload
	err := DecryptJSON("", key, nonPtr)
	if err == nil {
		t.Fatalf("expected error for non-pointer dst")
	}
	if err.Error() == "" {
		t.Fatalf("expected descriptive error for non-pointer dst")
	}
}

func TestEncryptJSON_InvalidKey(t *testing.T) {
	type Payload struct{ X string }
	_, _, err := EncryptJSON(Payload{X: "x"}, "not-base64!!!")
	if err == nil {
		t.Fatalf("expected error for invalid base64 key")
	}
}

func TestDecryptJSON_WrongKeyFails(t *testing.T) {
	type Payload struct{ X string }
	orig := Payload{X: "secret"}

	key1 := randomKeyB64(t, 32)
	key2 := randomKeyB64(t, 32)

	enc, _, err := EncryptJSON(orig, key1)
	if err != nil {
		t.Fatalf("EncryptJSON error: %v", err)
	}

	var out Payload
	if err := DecryptJSON(enc, key2, &out); err == nil {
		t.Fatalf("expected decrypt error with wrong key")
	}
}
