package leaderboard

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
	"github.com/rbrabson/dgame/bank"
	"github.com/rbrabson/dgame/internal/discmsg"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"leaderboard": leaderboard,
		"current":     currentLeaderboard,
		"monthly":     monthlyLeaderboard,
		"lifetime":    lifetimeLeaderboard,
		"rank":        rank,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "leaderboard",
			Description: "Commands used to interact with the leaderboard for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "channel",
					Description: "Sets the channel ID where the monthly leaderboard is published at the end of the month.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "The channel ID.",
							Required:    true,
						},
					},
				},
			},
		},
	}
	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "current",
			Description: "Gets the current economy leaderboard.",
		},
		{
			Name:        "monthly",
			Description: "Gets the monthly economy leaderboard.",
		},
		{
			Name:        "lifetime",
			Description: "Gets the lifetime economy leaderboard.",
		},
		{
			Name:        "rank",
			Description: "Gets the member rank for the leaderboards.",
		},
	}
)

// leaderboard updates the leaderboard channel.
func leaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> leaderboard.leaderboard")
	defer log.Trace("<-- leaderboard.leaderboard")

	options := i.ApplicationCommandData().Options
	if options[0].Name == "channel" {
		setLeaderboardChannel(s, i)
	}
}

// currentLeaderboard returns the top ranked accounts for the current balance.
func currentLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> leader.currentLeaderboard")
	defer log.Trace("<-- leaderboard.currentLeaderboard")

	lb := getLeaderboard(i.GuildID)
	leaderboard := getCurrentLeaderboard(lb)
	sendLeaderboard(s, i, "Current Leaderboard", leaderboard)
}

// monthlyLeaderboard returns the top ranked accounts for the current months.
func monthlyLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> leaderboard.monthlyLeaderboard")
	defer log.Trace("<-- leaderboard.monthlyLeaderboard")

	lb := getLeaderboard(i.GuildID)
	leaderboard := getMontlyLeaderboard(lb)
	sendLeaderboard(s, i, "Monthly Leaderboard", leaderboard)
}

// lifetimeLeaderboard returns the top ranked accounts for the lifetime of the server.
func lifetimeLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> leaderboard.lifetimeLeaderboard")
	defer log.Trace("<-- leaderboard.lifetimeLeaderboard")

	lb := getLeaderboard(i.GuildID)
	leaderboard := getLifetimeLeaderboard(lb)
	sendLeaderboard(s, i, "Lifetime Leaderboard", leaderboard)
}

// setLeaderboardChannel sets the server channel to which the monthly leaderboard is published.
func setLeaderboardChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setLeaderboardChannel")
	defer log.Trace("<-- setLeaderboardChannel")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	lb := getLeaderboard(i.GuildID)
	channelID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	lb.setChannel(channelID)

	resp := p.Sprintf("Channel ID for the monthly leaderboard set to %s.", lb.ChannelID)
	discmsg.SendResponse(s, i, resp)
}

// sendLeaderboard is a utility function that sends an economy leaderboard to Discord.
func sendLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, title string, accounts []*bank.Account) {
	log.Trace("--> sendLeaderboard")
	defer log.Trace("<-- sendLeaderboard")

	p := discmsg.GetPrinter(language.AmericanEnglish)
	embeds := formatAccounts(p, title, accounts)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// rank returns the rank of the member in the leaderboard.
func rank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> leaderboard.rank")
	defer log.Trace("<-- leaderboard.rank")

	p := discmsg.GetPrinter(language.AmericanEnglish)
	b := bank.GetBank(i.GuildID)
	account := b.GetAccount(i.Member.User.ID)
	lb := getLeaderboard(i.GuildID)
	monthlyRank := getMonthlyRanking(lb, account)
	lifetimeRank := getLifetimeRanking(lb, account)

	resp := p.Sprintf("Your monthly rank is %d and your lifetime rank is %d.", monthlyRank, lifetimeRank)
	discmsg.SendEphemeralResponse(s, i, resp)
}

// formatAccounts formats the leaderboard to be sent to a Discord server
func formatAccounts(p *message.Printer, title string, accounts []*bank.Account) []*discordgo.MessageEmbed {
	log.Trace("--> leaderboard.formatAccounts")
	defer log.Trace("<-- leaderboard.formatAccounts")

	var tableBuffer strings.Builder
	table := tablewriter.NewWriter(&tableBuffer)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"#", "Name", "Balance"})
	for i, account := range accounts {
		data := []string{strconv.Itoa(i + 1), account.GuildID, p.Sprintf("%d", account.CurrentBalance)}
		table.Append(data)
	}
	table.Render()
	embeds := []*discordgo.MessageEmbed{
		{
			Type:  discordgo.EmbedTypeRich,
			Title: title,
			Fields: []*discordgo.MessageEmbedField{
				{
					Value: p.Sprintf("```\n%s```\n", tableBuffer.String()),
				},
			},
		},
	}

	return embeds
}
