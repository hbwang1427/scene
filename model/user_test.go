package model

import (
	"testing"

	"github.com/aitour/scene/auth"
)

func TestUser(t *testing.T) {
	SetDbArgs("localhost", "dbuser", "kingwang", "testdb")

	user, err := CreateUser("kingwang", "luckykw99@gmail.com", "12345678")
	if err != nil {
		t.Fatalf("create user error:%v", err)
	}

	user2, err := GetUserByName("kingwang")
	if err != nil {
		t.Fatalf("get user by name error:%v", err)
	}
	if user.Id != user2.Id {
		t.Fatalf("user id error %d != %d", user.Id, user2.Id)
	}
	if user.CreateAt.UnixNano() != user2.Salt {
		t.Fatalf("createAt check failed: %v != %v", user.CreateAt.UnixNano(), user2.Salt)
	}

	authPass := auth.VerifyPassword("12345678", user2.Salt, user2.Password)
	if !authPass {
		t.Fatal("auth check failed")
	}
}
