package bank

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/dgame/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestGetBank(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
	}
}

func TestGetAccounts(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	bank.GetAccount("67890")
	filter := bson.D{
		{Key: "guild_id", Value: bank.GuildID},
	}
	sort := bson.D{
		{Key: "member_id", Value: 1},
	}
	accounts := bank.GetAccounts(filter, sort, 10)
	if len(accounts) == 0 {
		t.Error("no accounts returned")
	}
}
