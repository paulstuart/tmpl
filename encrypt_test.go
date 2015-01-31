package main

import (
	"testing"
)

var (
	testText   = "my little pony"
	testSecret string
)

func TestEncrypt(t *testing.T) {
	var err error
	testSecret, err = StringEncrypt(testText)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecrypt(t *testing.T) {
	revealed, err := StringDecrypt(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	if revealed != testText {
		t.Fatal("Decrypted", revealed, "Should be", testText)
	}
}
