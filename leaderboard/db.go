package leaderboard

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	LEADERBOARD_COLLECTION = "leaderboards"
)

// readLeaderboard reads the leaderboard from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database
func readLeaderboard(guildID string) *Leaderboard {
	log.Trace("--> leaderboard.readLeaderboard")
	defer log.Trace("<-- leaderboard.readLeaderboard")

	filter := make(map[string]interface{})
	filter["guild_id"] = guildID
	var lb Leaderboard
	err := db.Read(LEADERBOARD_COLLECTION, filter, &lb)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Info("leaderboard not found in the database")
		return nil
	}

	return &lb
}

// writeBank creates or updates the bank for a guild in the database being used by the Discord bot.
func writeLeaderboard(lb *Leaderboard) error {
	log.Trace("--> leaderboard.writeLeaderboard")
	defer log.Trace("<-- leaderboard.writeLeaderboard")

	filter := make(map[string]interface{})
	if lb.ID != primitive.NilObjectID {
		filter["_id"] = lb.ID
	} else {
		filter["guild_id"] = lb.GuildID
	}

	err := db.Write(LEADERBOARD_COLLECTION, filter, lb)
	if err != nil {
		log.WithField("guild", lb.GuildID).Error("unable to save leaderboard to the database")
		return err
	}
	log.WithField("guild", lb.GuildID).Debug("save leaderboard to the database")

	return nil
}
