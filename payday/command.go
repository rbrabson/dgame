package payday

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/bank"
	"github.com/rbrabson/dgame/internal/discmsg"
	"github.com/rbrabson/dgame/internal/format"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"payday": payday,
	}

	adminCommands = []*discordgo.ApplicationCommand{}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "payday",
			Description: "Deposits your daily check into your bank account.",
		},
	}
)

// payday gives some credits to the player every 24 hours.
func payday(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> payday")
	defer log.Trace("<-- payday")

	p := discmsg.GetPrinter(language.AmericanEnglish)
	payday := GetPayday(i.GuildID)
	paydayAccount := payday.GetAccount(i.Member.User.ID)

	if paydayAccount.NextPayday.After(time.Now()) {
		remainingTime := time.Until(paydayAccount.NextPayday)
		resp := p.Sprintf("You can't get another payday yet. You need to wait %s.", format.Duration(remainingTime))
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	b := bank.GetBank(i.GuildID)
	account := b.GetAccount(i.User.ID)
	account.Deposit(payday.Amount)

	paydayAccount.NextPayday = time.Now().Add(payday.PaydayFrequency)
	writeAccount(paydayAccount)

	resp := p.Sprintf("You deposited your check of %d into your bank account. You now have %d credits.", payday.Amount, account.CurrentBalance)
	discmsg.SendEphemeralResponse(s, i, resp)
}
