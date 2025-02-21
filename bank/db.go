package bank

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	BANK_COLLECTION    = "banks"
	ACCOUNT_COLLECTION = "bank_accounts"
)

// Resets the monthly balances for all accounts in all banks.
func ResetMonthlyBalances() {
	log.Trace("--> bank.ResetMonthlyBalances")
	defer log.Trace("<-- bank.ResetMonthlyBalances")

	filter := bson.M{}
	update := bson.D{{Key: "monthly_balance", Value: 0}}
	err := db.WriteAll(ACCOUNT_COLLECTION, filter, update)
	if err != nil {
		log.WithError(err).Error("unable to reset monthly balances for all accounts")
	}
}

// readBank gets the bank from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database.
func readBank(guildID string) *Bank {
	log.Trace("--> bank.readBank")
	defer log.Trace("<-- bank.readBank")

	filter := bson.M{"guild_id": guildID}
	var bank Bank
	err := db.Read(BANK_COLLECTION, filter, &bank)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Info("bank not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": bank.GuildID}).Info("read bank from the database")
	return &bank
}

// writeBank creates or updates the bank data in the database being used by the Discord bot.
func writeBank(bank *Bank) error {
	log.Trace("--> bank.writeBank")
	defer log.Trace("<-- bank.writeBank")

	filter := bson.M{"guild_id": bank.GuildID}
	err := db.Write(BANK_COLLECTION, filter, bank)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID}).Error("unable to save bank to the database")
		return err
	}
	log.WithFields(log.Fields{"guild": bank.GuildID}).Info("save bank to the database")

	return nil
}

// Get all the matching accounts for the given bank.
func readAccounts(bank *Bank, filter interface{}, sortBy interface{}, limit int64) []*Account {
	log.Trace("--> bank.readAccounts")
	defer log.Trace("<-- bank.readAccounts")

	var accounts []*Account
	err := db.ReadAll(ACCOUNT_COLLECTION, filter, &accounts, sortBy, limit)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID}).Error("unable to read accounts from the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": bank.GuildID, "count": len(accounts)}).Info("read accounts from the database")

	return accounts
}

// readAccount reads the account from the database and returns the value, if it exists, or returns nil if the
// account does not exist in the database
func readAccount(bank *Bank, memberID string) *Account {
	log.Trace("--> bank.readAccount")
	defer log.Trace("<-- bank.readAccount")

	filter := bson.M{
		"guild_id":  bank.GuildID,
		"member_id": memberID,
	}
	var account Account
	err := db.Read(ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID, "bank": memberID}).Info("account not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Info("read account from the database")

	return &account
}

// writeAccount creates or updates the member data in the database being used by the Discord bot.
func writeAccount(account *Account) error {
	log.Trace("--> bank.writeAccount")
	defer log.Trace("<-- bank.writeAccount")

	filter := bson.M{
		"guild_id":  account.GuildID,
		"member_id": account.MemberID,
	}
	err := db.Write(ACCOUNT_COLLECTION, filter, account)
	if err != nil {
		log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Error("unable to save bank account to the database")
		return err
	}
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Info("save bank account to the database")

	return nil
}
