package heist

import (
	"encoding/json"
	"time"

	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	MEMBER = "heist_member"
)

type CriminalLevel int

const (
	GREENHORN CriminalLevel = 0
	RENEGADE  CriminalLevel = 1
	VETERAN   CriminalLevel = 10
	COMMANDER CriminalLevel = 25
	WAR_CHIEF CriminalLevel = 50
	LEGEND    CriminalLevel = 75
	IMMORTAL  CriminalLevel = 100
)

type MemberStatus int

const (
	APPREHENDED MemberStatus = iota
	DEAD
	FREE
	OOB
)

var (
	members = make(map[string]map[string]*Member)
)

// Member is the status of a member who has participated in at least one heist
type Member struct {
	ID            interface{}   `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string        `json:"guild_id" bson:"guild_id"`
	MemberID      string        `json:"member_id" bson:"member_id"`
	BailCost      int           `json:"bail_cost" bson:"bail_cost"`
	CriminalLevel CriminalLevel `json:"criminal_level" bson:"criminal_level"`
	Deaths        int           `json:"deaths" bson:"deaths"`
	DeathTimer    time.Time     `json:"death_timer" bson:"death_timer"`
	JailCounter   int           `json:"jail_counter" bson:"jail_counter"`
	JailTimer     time.Time     `json:"time_served" bson:"time_served"`
	Sentence      time.Duration `json:"sentence" bson:"sentence"`
	Spree         int           `json:"spree" bson:"spree"`
	Status        MemberStatus  `json:"status" bson:"status"`
	TotalJail     int           `json:"total_jail" bson:"total_jail"`
}

// GetMember gets a member for heists. If the member does not exist, then nil is returned.
func GetMember(guild *guild.Guild, guildMember *guild.Member) *Member {
	log.Trace("--> heist.GetMember")
	defer log.Trace("<-- heist.GetMember")

	guildMembers := members[guild.GuildID]
	if guildMembers == nil {
		members[guild.GuildID] = make(map[string]*Member)
		log.WithField("guild", guild.GuildID).Debug("creating heist members for the guild")
	}

	member := guildMembers[guildMember.MemberID]

	return member
}

// NewMember creates a new member for heists. It is called when guild member
// first plans or joins a heist.
func NewMember(guild *guild.Guild, guildMember *guild.Member) *Member {
	log.Trace("--> heist.NewMember")
	defer log.Trace("<-- heist.NewMember")

	member := &Member{
		MemberID:      guildMember.MemberID,
		CriminalLevel: GREENHORN,
		Status:        FREE,
		GuildID:       guild.GuildID,
	}

	writeMember(member)
	log.WithFields(log.Fields{"guild": guild.GuildID, "member": member.MemberID}).Debug("create heist member")

	return member
}

// Release frees the member from jail and removes the jail and death times. It is used when a
// member is released from jail or is revived after dying in a heist.
func (member *Member) Release() {
	log.Trace("--> heist.Member.Release")
	log.Trace("<-- heist.Member.Release")

	member.Status = FREE
	member.DeathTimer = time.Time{}
	member.BailCost = 0
	member.Sentence = 0
	member.JailTimer = time.Time{}
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Info("released from jail")

	writeMember(member)
}

// Clear resets the status of the member and releases them from jail, so they can participate in heists.
func (member *Member) Clear() {
	log.Trace("--> heist.Member.Clear")
	log.Trace("<-- heist.Member.Clear")

	member.CriminalLevel = GREENHORN
	member.JailCounter = 0
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Info("status cleared")

	member.Release()
}

// RemainingJailTime returns the amount of time remaining on the player's sentence has been served
func (member *Member) RemainingJailTime() time.Duration {
	log.Trace("--> heist.Member.RemainingJailTime")
	log.Trace("<-- heist.Member.RemainingJailTime")

	if member.JailTimer.IsZero() || member.JailTimer.After(time.Now()) {
		return 0
	}
	return time.Until(member.JailTimer)
}

// RemainingDeathTime returns the amount of time before the member can be resurected.
func (member *Member) RemainingDeathTime() time.Duration {
	log.Trace("--> heist.Member.RemainingDeathTime")
	log.Trace("<-- heist.Member.RemainingDeathTime")

	if member.DeathTimer.IsZero() || member.DeathTimer.After(time.Now()) {
		return 0
	}
	return time.Until(member.DeathTimer)
}

// Reset clears the jain and death settings for a member.
func (m *Member) Reset() {
	m.BailCost = 0
	m.CriminalLevel = GREENHORN
	m.DeathTimer = time.Time{}
	m.JailCounter = 0
	m.JailTimer = time.Time{}
	m.Sentence = 0
	m.Status = FREE
	m.Status = FREE
}

// String returns a string representation of the member of the heist
func (member *Member) String() string {
	out, _ := json.Marshal(member)
	return string(out)
}

// String returns the string representation for a criminal level
func (level CriminalLevel) String() string {
	switch {
	case level >= IMMORTAL:
		return "Immortal"
	case level >= LEGEND:
		return "Legend"
	case level >= WAR_CHIEF:
		return "WarChief"
	case level >= COMMANDER:
		return "Commander"
	case level >= VETERAN:
		return "Veteran"
	case level >= RENEGADE:
		return "Renegade"
	case level >= GREENHORN:
		return "Greenhorn"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the status of the member of a heist
func (status MemberStatus) String() string {
	switch status {
	case FREE:
		return "Free"
	case DEAD:
		return "Dead"
	case APPREHENDED:
		return "Apprehended"
	case OOB:
		return "Out on Bail"
	default:
		return "Unknownn"
	}
}
