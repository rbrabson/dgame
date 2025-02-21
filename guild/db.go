package guild

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	MEMBER_COLLECTION = "guild_members"
)

// readMember reads the member from the database and returns the value, if it exists, or returns nil if the
// member does not exist in the database
func readMember(guild *Guild, memberID string) *Member {
	log.Trace("--> guild.loadMember")
	defer log.Trace("<-- guild.loadMember")

	filter := bson.M{
		"guild_id":  guild.GuildID,
		"member_id": memberID,
	}
	var member Member
	err := db.Read(MEMBER_COLLECTION, filter, &member)
	if err != nil {
		log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID}).Info("guild member not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "name": member.Name}).Info("read guild member from the database")
	return &member
}

// writeMember creates or updates the member data in the database.
func writeMember(member *Member) error {
	log.Trace("--> guild.Member.writeMember")
	defer log.Trace("<-- guild.Member.writeMember")

	filter := bson.M{
		"guild_id":  member.GuildID,
		"member_id": member.MemberID,
	}
	err := db.Write(MEMBER_COLLECTION, filter, member)
	if err != nil {
		log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Info("guild member not found in the database")
		return err
	}

	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "name": member.Name}).Info("write guild member to the database")
	return nil
}
