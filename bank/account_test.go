package bank

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

func TestDeposit(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	account := bank.GetAccount("54321")
	account.SetBalance(0)
	account.Deposit(100)
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestWithdraw(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	account := bank.GetAccount("54321")
	account.SetBalance(200)
	account.Withdraw(100)
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestSetBalance(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	account := bank.GetAccount("54321")
	account.SetBalance(500)
	if account.CurrentBalance != 500 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestResetMonthlyBalances(t *testing.T) {
	db = mongo.NewDatabase()
	defer t.Cleanup(func() {
		db.Close()
	})

	bank := GetBank("12345")
	account := bank.GetAccount("54321")
	account.CurrentBalance = 500
	account.MonthlyBalance = 750
	writeAccount(account)

	ResetMonthlyBalances()

	account = bank.GetAccount("54321")
	if account.MonthlyBalance != 0 {
		t.Errorf("Expected monthly balance to be 0, got %d", account.MonthlyBalance)
	}
}
