package payday

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PAYDAY_COLLECTION         = "paydays"
	PAYDAY_ACCOUNT_COLLECTION = "payday_accounts"
)

// readPayday loads payday information for the guild from the database.
func readPayday(guildID string) *Payday {
	log.Trace("--> payday.readPayday")
	defer log.Trace("<-- payday.readPayday")

	filter := bson.M{
		"guild_id": guildID,
	}
	var payday *Payday
	err := db.Read(PAYDAY_COLLECTION, filter, &payday)
	if err != nil {
		log.WithField("guild", payday.GuildID).Info("payday not found in the database")
		return nil
	}
	log.WithField("guild", payday.GuildID).Info("read payday from the database")

	return payday
}

// writePayday saves the payday information for the guild into the database.
func writePayday(payday *Payday) error {
	log.Trace("--> payday.writePayday")
	defer log.Trace("<-- payday.writePayday")

	filter := bson.M{
		"guild_id": payday.GuildID,
	}
	db.Write(PAYDAY_COLLECTION, filter, payday)

	err := db.Write(PAYDAY_COLLECTION, filter, payday)
	if err != nil {
		log.WithField("guild", payday.GuildID).Error("unable to save payday to the database")
		return err
	}
	log.WithField("guild", payday.GuildID).Info("save bank account to the database")
	return nil
}

// readAccount loads payday information for a given account in the guild from the database.
func readAccount(payday *Payday, accountID string) *Account {
	log.Trace("--> payday.readAccount")
	defer log.Trace("<-- payday.readAccount")

	filter := bson.M{
		"guild_id":  payday.GuildID,
		"member_id": accountID,
	}
	var account *Account
	err := db.Read(PAYDAY_ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		log.WithFields(log.Fields{"guild": payday.GuildID, "member": accountID}).Info("payday account not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Info("read payday account from the database")
	account.GuildID = payday.GuildID

	return account
}

// writeAccount saves the payday information for a given account in the guild into the database.
func writeAccount(account *Account) error {
	log.Trace("--> payday.writeAccount")
	defer log.Trace("<-- payday.writeAccount")

	filter := bson.M{
		"guild_id":  account.GuildID,
		"member_id": account.MemberID,
	}
	err := db.Write(PAYDAY_ACCOUNT_COLLECTION, filter, payday)
	if err != nil {
		log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Info("payday account not found in the database")
		return err
	}
	if account.ID == primitive.NilObjectID {
		p := GetPayday(account.GuildID)
		p.GetAccount(account.MemberID)
		updated := readAccount(p, account.MemberID)
		if updated == nil {
			log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Error("unable to read payday from the database after saving it")
			return ErrUnableToSaveAccount
		}
		account.ID = updated.ID
	}
	return nil
}
