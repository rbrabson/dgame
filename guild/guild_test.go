package guild

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/dgame/database/mongo"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestGetGuild(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	guild := GetGuild("12345")
	if guild == nil {
		t.Errorf("GetGuild() guild not found or created")
	}
}

func TestGetMember(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	guild := GetGuild("12345")
	member := guild.GetMember("67890")
	if member == nil {
		t.Errorf("GetMember() member not found or created")
	}
}

func TestSetName(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	guild := GetGuild("12345")
	member := guild.GetMember("67890").SetName("userName", "displayName")
	if member == nil {
		t.Errorf("GetMember() member not found or created")
	}
	member.SetName("userName", "")
	if member == nil || member.Name != "userName" {
		t.Errorf("SetName() member name not set")
	}
	member.SetName("userName", "displayName")
	if member == nil || member.Name != "displayName" {
		t.Errorf("SetName() member name not set")
	}
}
