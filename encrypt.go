package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var (
	superSecret []byte
)

func SHA1(text string) string {
	h := sha1.New()
	io.WriteString(h, text)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func RandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)[:length]
}

func SetKey(text string) {
	h := sha256.New()
	io.WriteString(h, text)
	superSecret = h.Sum(nil)
}

func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func Decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func StringEncrypt(text string) (string, error) {
	s, err := Encrypt(superSecret, []byte(text))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(s), nil
}

func StringDecrypt(text string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	d, err := Decrypt(superSecret, data)
	return string(d), nil
}

// load encryption secret from file
// if it doesn't exist, create one
func LoadKey(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		key := RandomString(128)
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(f, key)
		f.Close()
		SetKey(key)
		return
	}
	key, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	SetKey(strings.TrimSpace(string(key)))
}
