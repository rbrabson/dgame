package bank

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"

	"github.com/rbrabson/dgame/internal/discmsg"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"bank":    bank,
		"account": account,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "bank",
			Description: "Commands used to interact with the economy for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "account",
					Description: "Sets the amount of credits for a given member.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The member ID.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "amount",
							Description: "The amount to set the account to.",
							Required:    true,
						},
					},
				},
				{
					Name:        "balance",
					Description: "Set the default balance for the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the default balance for the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "name",
					Description: "Set the name of the bank for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The the name of the bank for the server.",
							Required:    true,
						},
					},
				},
				{
					Name:        "currency",
					Description: "Set the currency for the server.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "The currency to set for the server.",
							Required:    true,
						},
					},
				},
			},
		},
	}
	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "account",
			Description: "Bank account balance for the member",
		},
	}
)

// bank routes the bank commands to the proper handers.
func bank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> bank.bank")
	defer log.Trace("<-- bank.bank")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "balance":
		setDefaultBalance(s, i)
	case "name":
		setBankName(s, i)
	case "currency":
		setBankCurrency(s, i)
	case "account":
		setAccountBalance(s, i)
	}
}

// account gets information about the member's bank account.
func account(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> bankAccount")
	defer log.Trace("<-- bankAccount")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	account := bank.GetAccount(i.User.ID)

	resp := p.Sprintf("**Balance**: %d\n**MonthlyBalance**: %d\n**LifetimeBalance**: %d\n**Created**: %s\n",
		account.CurrentBalance,
		account.MonthlyBalance,
		account.LifetimeBalance,
		account.CreatedAt,
	)
	discmsg.SendEphemeralResponse(s, i, resp)
}

// setAccountBalance sets the balance of the account for the member of the guild to the specified amount
func setAccountBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setAccountBalance")
	defer log.Trace("<-- setAccountBalance")

	var id string
	var amount int
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "id":
			id = strings.TrimSpace(option.StringValue())
		case "amount":
			amount = int(option.IntValue())
		}
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	member, err := s.GuildMember(i.GuildID, id)
	if err != nil {
		resp := p.Sprintf("An account with ID `%s` is not a member of this server", id)
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	g := guild.GetGuild(i.GuildID)
	m := g.GetMember(member.User.ID).SetName(i.User.Username, i.Member.DisplayName())
	bank := GetBank(i.GuildID)
	account := bank.GetAccount(id)

	account.SetBalance(amount)

	log.WithFields(log.Fields{
		"guild":   i.GuildID,
		"account": member.User.ID,
		"mName":   m.Name,
		"balance": amount,
	}).Debug("/bank set account")

	resp := p.Sprintf("Account balance for %s was set to %d", m.Name, account.CurrentBalance)
	discmsg.SendResponse(s, i, resp)
}

// setDefaultBalance sets the default balance for bank for the guild (server).
func setDefaultBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setDefaultBalance")
	defer log.Trace("<-- setDefaultBalance")

	var balance int
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			balance = int(option.IntValue())
			break
		}
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetDefaultBalance(balance)

	log.WithFields(log.Fields{
		"guild":   i.GuildID,
		"balance": balance,
	}).Debug("/bank balance")

	resp := p.Sprintf("Bank default balance was set to %s", bank.DefaultBalance)
	discmsg.SendResponse(s, i, resp)
}

// setBankName sets the name of the bank for the guild (server).
func setBankName(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setBankName")
	defer log.Trace("<-- setBankName")

	var name string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			name = strings.TrimSpace(option.StringValue())
			break
		}
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetName(name)

	log.WithFields(log.Fields{
		"guild": i.GuildID,
		"name":  name,
	}).Debug("/bank name")

	resp := p.Sprintf("Bank name was set to %s", bank.Name)
	discmsg.SendResponse(s, i, resp)
}

// setBankCurrency sets the name of the currency used by the bank for the guild (server).
func setBankCurrency(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setBankCurrency")
	defer log.Trace("<-- setBankCurrency")

	var currency string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "value" {
			currency = strings.TrimSpace(option.StringValue())
			break
		}
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	bank := GetBank(i.GuildID)
	bank.SetCurrency(currency)

	log.WithFields(log.Fields{
		"guild": i.GuildID,
		"name":  currency,
	}).Debug("/bank currency")

	resp := p.Sprintf("Bank name was set to %s", bank.Currency)
	discmsg.SendResponse(s, i, resp)
}
