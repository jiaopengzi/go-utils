//
// FilePath    : go-utils\dtovalidator\common_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : dtovalidator 公共方法测试
//

package dtovalidator

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestStringHasPrefixIgnoreCase(t *testing.T) {
	cases := []struct {
		s, prefix string
		want      bool
	}{
		{"HelloWorld", "hello", true},
		{"HelloWorld", "HELLO", true},
		{"abc", "abcd", false},
		{"", "", true},
	}

	for _, c := range cases {
		if got := stringHasPrefixIgnoreCase(c.s, c.prefix); got != c.want {
			t.Fatalf("stringHasPrefixIgnoreCase(%q, %q) = %v; want %v", c.s, c.prefix, got, c.want)
		}
	}
}

func TestStringHasSuffixIgnoreCase(t *testing.T) {
	cases := []struct {
		s, suffix string
		want      bool
	}{
		{"HelloWorld", "world", true},
		{"HelloWorld", "WORLD", true},
		{"abc", "z", false},
		{"", "", true},
	}

	for _, c := range cases {
		if got := stringHasSuffixIgnoreCase(c.s, c.suffix); got != c.want {
			t.Fatalf("stringHasSuffixIgnoreCase(%q, %q) = %v; want %v", c.s, c.suffix, got, c.want)
		}
	}
}

func TestValidateCSR(t *testing.T) {
	v := validator.New()
	if err := v.RegisterValidation("ValidateCSR", ValidateCSR); err != nil {
		t.Fatalf("register validation failed: %v", err)
	}

	type S struct {
		CSR string `validate:"ValidateCSR"`
	}

	validCSR := `-----BEGIN CERTIFICATE REQUEST-----
MIG6MG4CAQAwFDESMBAGA1UEAxMJbG9jYWxob3N0MCowBQYDK2VwAyEAr2h/kLhK
6e0FsbWcOjyBYr6dewt95bS9TBZ95Dm9jTWgJzAlBgkqhkiG9w0BCQ4xGDAWMBQG
A1UdEQQNMAuCCWxvY2FsaG9zdDAFBgMrZXADQQDi/X6l3MkbAWkeYPSBjJGR/zxH
b0ywI0a+em51y5dgH/o6Ud052pGWysNVx7FRaLQoG/DQfT4ofSasRwqeIRoN
-----END CERTIFICATE REQUEST-----`

	s := S{CSR: validCSR}
	if err := v.Struct(s); err != nil {
		t.Fatalf("valid CSR flagged invalid: %v", err)
	}

	// missing header
	s2 := S{CSR: "some body\n-----END CERTIFICATE REQUEST-----"}
	if err := v.Struct(s2); err == nil {
		t.Fatalf("invalid CSR (missing header) was accepted")
	}

	// missing footer
	s3 := S{CSR: "-----BEGIN CERTIFICATE REQUEST-----\nsome body"}
	if err := v.Struct(s3); err == nil {
		t.Fatalf("invalid CSR (missing footer) was accepted")
	}
}
