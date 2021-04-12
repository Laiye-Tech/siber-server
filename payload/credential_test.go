package payload

import (
	"api-test/configs"
	"os"
	"testing"
)

func Test_GetHeaders(t *testing.T) {
	secret := configs.GetGlobalConfig().Auth.TestPubKey
	pubKey := configs.GetGlobalConfig().Auth.TestSecret
	credential := NewCredential(secret, pubKey)
	headers := credential.GetHeaders()

	if headers["Api-Auth-pubkey"] == "" {
		t.Error("[Test_PostWithNetError]=> pubkey is missed.")
	}

	if headers["Api-Auth-sign"] == "" {
		t.Error("[Test_PostWithNetError]=> sign is missed.")
	}

	if headers["Api-Auth-timestamp"] == "" {
		t.Error("[Test_PostWithNetError]=> timestamp is missed.")
	}

	if headers["Api-Auth-nonce"] == "" {
		t.Error("[Test_PostWithNetError]=> nonce is missed.")
	}
}

func Test_GetRandomString(t *testing.T) {
	secret, pubkey := os.Getenv("secret"), os.Getenv("pubkey")
	credential := NewCredential(secret, pubkey)
	str := credential.GetRandomString(15)
	if len(str) != 15 {
		t.Error("[Test_GetRandomString]=> random strint len is wrong.")
	}
}

func Benchmark_GetHeaders(t *testing.B) {
	secret, pubkey := os.Getenv("secret"), os.Getenv("pubkey")
	credential := NewCredential(secret, pubkey)

	for i := 0; i < t.N; i++ {
		headers := credential.GetHeaders()
		if headers["Api-Auth-pubkey"] == "" {
			t.Error("[Benchmark_GetHeaders]=> pubkey is missed.")
		}
	}
}

func Benchmark_GetRandomString(t *testing.B) {
	secret, pubkey := os.Getenv("secret"), os.Getenv("pubkey")
	credential := NewCredential(secret, pubkey)

	for i := 0; i < t.N; i++ {
		str := credential.GetRandomString(15)
		if len(str) != 15 {
			t.Error("[Benchmark_GetRandomString]=> random strint len is wrong.")
		}
	}
}
