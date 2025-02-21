package heist

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

type HeistState int

const (
	PLANNED HeistState = iota
	STARTED
	COMPLETED
)

// Heist is a heist that is being planned, is in progress, or has completed
type Heist struct {
	ID          interface{}                  `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string                       `json:"guild_id" bson:"guild_id"`
	MemberID    string                       `json:"member_id" bson:"member_id"`
	CrewIDs     []string                     `json:"crew_ids" bson:"crew_ids"`
	State       HeistState                   `json:"state" bson:"state"`
	StartTime   time.Time                    `json:"start_time" bson:"start_time"`
	session     *discordgo.Session           `json:"-" bson:"-"`
	interaction *discordgo.InteractionCreate `json:"-" bson:"-"`
	mutex       sync.Mutex                   `json:"-" bson:"-"`
}

// HeistResult are the results of a heist
type HeistResult struct {
	Escaped       int
	Apprehended   int
	Dead          int
	MemberResults []*HeistMemberResult
	SurvivingCrew []*HeistMemberResult
	Target        *Target
}

// HeistMemberResult is the result for a single member of the heist
type HeistMemberResult struct {
	Player        *Member
	Status        string
	Message       string
	StolenCredits int
	BonusCredits  int
}

// NewHeist creates a new heist that is being planned.
func NewHeist(guild *guild.Guild, member *guild.Member) *Heist {
	log.Trace("--> heist.NewHeist")
	defer log.Trace("<-- heist.NewHeist")

	heist := &Heist{
		MemberID:  member.MemberID,
		CrewIDs:   make([]string, 0, 5),
		State:     PLANNED,
		StartTime: time.Now(),
		GuildID:   guild.GuildID,
		mutex:     sync.Mutex{},
	}

	return heist
}

// String returns a string representation of the state of a heist
func (state HeistState) String() string {
	switch state {
	case PLANNED:
		return "Planned"
	case STARTED:
		return "Started"
	case COMPLETED:
		return "Completed"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the heist
func (heist *Heist) String() string {
	out, _ := json.Marshal(heist)
	return string(out)
}

// String returns a string representation of the resuilt of a heist
func (result *HeistResult) String() string {
	out, _ := json.Marshal(result)
	return string(out)
}

// String returns a string representation of the resuilt for a single member of a heist
func (result *HeistMemberResult) String() string {
	out, _ := json.Marshal(result)
	return string(out)
}
