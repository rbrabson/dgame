package heist

import (
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

// readConfig loads the heist configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guild *guild.Guild) *Config {
	log.Trace("--> heist.readConfig")
	defer log.Trace("<-- heist.readConfig")

	filter := make(map[string]interface{})
	filter["guild_id"] = guild.GuildID
	var config Config
	db.Read(CONFIG_COLLECTION, filter, &config)

	return &config
}

// writeConfig stores the configuration in the database.
func writeConfig(config *Config) {
	log.Trace("--> heist.writeConfig")
	defer log.Trace("<-- heist.writeConfig")

	filter := make(map[string]interface{})
	if config.ID != nil {
		filter["_id"] = config.ID
	} else {
		filter["guild_id"] = config.GuildID
	}
	db.Write(CONFIG_COLLECTION, filter, config)
}

// Write creates or updates the heist member in the database
func writeMember(member *Member) {
	log.Trace("--> heist.writeMember")
	defer log.Trace("<-- heist.writeMember")

	filter := make(map[string]interface{})
	if member.ID != nil {
		filter["_id"] = member.ID
	} else {
		filter["guild_id"] = member.GuildID
		filter["member_id"] = member.MemberID
	}
	db.Write(MEMBER, filter, member)
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Debug("write member to the database")
}
