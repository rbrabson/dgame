package leaderboard

import (
	"time"

	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/bank"
	"github.com/rbrabson/dgame/internal/discmsg"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/language"
)

// A Leaderboard is used to send a monthly leaderboard to the Discord server for each guild.
type Leaderboard struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	ChannelID  string             `json:"channel_id" bson:"channel_id"`
	LastSeason time.Time          `json:"last_season" bson:"last_season"`
}

// getLeaderboards returns all the leaderboards for all guilds known to the bot.
func getLeaderboards() []*Leaderboard {
	log.Trace("--> leaderboard.getLeaderboards")
	defer log.Trace("<-- leaderboard.getLeaderboards")

	// TODO: read all leaderboards from the DB
	return nil
}

// getLeaderboard returns the leaderbord for the given guild
func getLeaderboard(guildID string) *Leaderboard {
	log.Trace("--> leaderboard.getLeaderboard")
	defer log.Trace("<-- leaderboard.getLeaderboard")

	lb := readLeaderboard(guildID)

	return lb
}

func (lb *Leaderboard) setChannel(channelID string) {
	log.Trace("--> leaderboard.setChannel")
	defer log.Trace("<-- leaderboard.setChannel")

	lb.ChannelID = channelID
	writeLeaderboard(lb)
}

// GetCurrentRanking returns the global rankings based on the current balance.
func getCurrentLeaderboard(lb *Leaderboard) []*bank.Account {
	log.Trace("--> leaderboard.getCurrentLeaderboard")
	defer log.Trace("<-- leaderboard.getCurrentLeaderboard")

	b := bank.GetBank(lb.GuildID)
	filter := bson.M{"guild_id": lb.GuildID}
	sort := bson.M{"balance": -1}
	limit := int64(10)

	accounts := b.GetAccounts(filter, sort, limit)

	return accounts
}

// getMontlyLeaderboard returns the global rankings based on the monthly balance.
func getMontlyLeaderboard(lb *Leaderboard) []*bank.Account {
	log.Trace("--> leaderboard.getMontlyLeaderboard")
	defer log.Trace("<-- leaderboard.getMontlyLeaderboard")

	b := bank.GetBank(lb.GuildID)
	filter := bson.M{"guild_id": lb.GuildID}
	sort := bson.M{"monthly_balance": -1}
	limit := int64(10)

	accounts := b.GetAccounts(filter, sort, limit)

	return accounts
}

// getLifetimeLeaderboard returns the global rankings based on the monthly balance.
func getLifetimeLeaderboard(lb *Leaderboard) []*bank.Account {
	log.Trace("--> leaderboard.getLifetimeLeaderboard")
	defer log.Trace("<-- leaderboard.getLifetimeLeaderboard")

	b := bank.GetBank(lb.GuildID)
	filter := bson.M{"guild_id": lb.GuildID}
	sort := bson.M{"lifetime_balance": -1}
	limit := int64(10)

	accounts := b.GetAccounts(filter, sort, limit)

	return accounts
}

// getMonthlyRanking returns the monthly global ranking on the server for a given player.
func getMonthlyRanking(lb *Leaderboard, account *bank.Account) int {
	log.Trace("--> leaderboard.getMonthlyRanking")
	defer log.Trace("<-- leaderboard.getMonthlyRanking")

	accounts := getMontlyLeaderboard(lb)
	rank := slices.Index(accounts, account) + 1

	return rank
}

// getLifetimeRanking returns the lifetime global ranking on the server for a given player.
func getLifetimeRanking(lb *Leaderboard, account *bank.Account) int {
	log.Trace("--> leaderboard.getLifetimeRanking")
	defer log.Trace("<-- leaderboard.getLifetimeRanking")

	accounts := getLifetimeLeaderboard(lb)
	rank := slices.Index(accounts, account) + 1

	return rank
}

// sendMonthlyLeaderboard publishes the monthly leaderboard to the bank channel.
func sendhMonthlyLeaderboard(lb *Leaderboard) {
	log.Trace("--> bank.sendMonthlyLeaderboard")
	defer log.Trace("<-- bank.sendMonthlyLeaderboard")

	// Get the top 10 accounts for this month
	sortedAccounts := getMontlyLeaderboard(lb)
	leaderboardSize := max(10, len(sortedAccounts))
	sortedAccounts = sortedAccounts[:leaderboardSize]

	year, month := prevMonth(time.Now())
	if lb.ChannelID != "" {
		p := discmsg.GetPrinter(language.AmericanEnglish)
		embeds := formatAccounts(p, fmt.Sprintf("%s %d Top 10", month, year), sortedAccounts)
		_, err := bot.Session.ChannelMessageSendComplex(lb.ChannelID, &discordgo.MessageSend{
			Embeds: embeds,
		})
		if err != nil {
			log.Error("Unable to send montly leaderboard, err:", err)
		}
	} else {
		log.WithField("guildID", lb.ChannelID).Warning("no leaderboard channel set for server")
	}
}

// publishMonthlyLeaderboard sends the monthly leaderboard to each guild.
func sendMonthlyLeaderboard() {
	log.Trace("--> bank.sendMonthlyLeaderboard")
	defer log.Trace("<-- bank.sendMonthlyLeaderboard")

	// Get the last season for the banks, defaulting to the current time if there are no banks.
	// This handles the off-chance that the server crashed and a new month starts before the
	// server is restarted.
	lastSeason := time.Now()
	leaderboards := getLeaderboards()
	for _, lb := range leaderboards {
		lastSeason = lb.LastSeason
		break
	}

	for {
		y, m := nextMonth(lastSeason)
		nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
		time.Sleep(time.Until(nextMonth))
		lastSeason = nextMonth

		leaderboards := getLeaderboards()
		for _, lb := range leaderboards {
			sendhMonthlyLeaderboard(lb)
			lb.LastSeason = lastSeason
			writeLeaderboard(lb)
		}

		bank.ResetMonthlyBalances()
	}
}

// prevMonth returns the previous year and month
func prevMonth(t time.Time) (int, time.Month) {
	y, m, _ := t.Date()
	y, m, _ = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC).Date()
	return y, m
}

// nextMonth returns the next year and month
func nextMonth(t time.Time) (int, time.Month) {
	y, m, _ := t.Date()
	y, m, _ = time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC).Date()
	return y, m
}
