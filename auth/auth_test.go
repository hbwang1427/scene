package auth

import (
	"log"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "@3sdf&K"
	salt := time.Now().UnixNano()
	hashedPassword := HashPassword(password, salt)
	log.Printf("%s", hashedPassword)
	pass := VerifyPassword(password, salt, hashedPassword)
	if !pass {
		t.Fatalf("VerifyPassword failed")
	}
}

func TestSimpleTokenProvider(t *testing.T) {
	provider, err := CreateTokenProvider("simple", map[string]interface{}{
		"tokenTTL": 10 * time.Second,
		"tokenLen": 16,
	})
	if err != nil {
		t.Fatalf("create provider error:%v", err)
	}

	token, err := provider.AssignToken("kingwang")
	if err != nil {
		t.Fatalf("assign token error:%v", err)
	}

	//log.Printf("token generated:%s", token)

	authInfo, err := provider.GetAuthInfo(token)
	if err != nil {
		t.Fatalf("get auth info error#1:%v", err)
	}

	if authInfo.UserName != "kingwang" {
		t.Fatalf("get auth info failed. expect %s, got %s", "kingwang", authInfo.UserName)
	}

	if err := provider.RevokeToken(token); err != nil {
		t.Fatalf("remove token failed. error:%v", err)
	}

	authInfo, err = provider.GetAuthInfo(token)
	if err != nil {
		t.Fatalf("get auth info error#2:%v", err)
	}
	if authInfo != nil {
		t.Fatalf("revoke token failed. revoked token still exist")
	}
}

func TestJwtTokenProvider(t *testing.T) {
	provider, err := CreateTokenProvider("jwt", map[string]interface{}{
		"key":      "hmacsecretkey",
		"tokenTTL": 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("create provider error:%v", err)
	}

	token, err := provider.AssignToken("kingwang")
	if err != nil {
		t.Fatalf("assign token error:%v", err)
	}

	//log.Printf("token generated:%s", token)

	authInfo, err := provider.GetAuthInfo(token)
	if err != nil {
		t.Fatalf("get auth info error#1:%v", err)
	}

	if authInfo.UserName != "kingwang" {
		t.Fatalf("get auth info failed. expect %s, got %s", "kingwang", authInfo.UserName)
	}

	if err := provider.RevokeToken(token); err != nil {
		t.Fatalf("remove token failed. error:%v", err)
	}

	authInfo, err = provider.GetAuthInfo(token)
	if err != nil {
		t.Fatalf("get auth info error#2:%v", err)
	}
	if authInfo != nil {
		t.Fatalf("revoke token failed. revoked token still exist")
	}

	//test expire
	token, err = provider.AssignToken("kingwang")
	if err != nil {
		t.Fatalf("assign token error:%v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if authInfo, err = provider.GetAuthInfo(token); authInfo != nil {
		t.Fatalf("token was not expire as expected")
	}
}
