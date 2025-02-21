package heist

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rbrabson/dgame/guild"
)

const (
	GAME_ID           = "heist"
	CONFIG_COLLECTION = "config"
)

const (
	BAIL_BASE     = 250
	CREW_OUTPUT   = "None"
	DEATH_TIMER   = 45
	HEIST_COST    = 1500
	POLICE_ALERT  = 60
	SENTENCE_BASE = 5
	WAIT_TIME     = time.Duration(60 * time.Second)
)

var (
	configs = make(map[string]*Config)
)

// Configuration data for new heists
type Config struct {
	ID           interface{}   `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID      string        `json:"guild_id" bson:"guild_id"`
	AlertTime    time.Time     `json:"alert_time" bson:"alert_time"`
	BailBase     int           `json:"bail_base" bson:"bail_base"`
	CrewOutput   string        `json:"crew_output" bson:"crew_output"`
	DeathTimer   time.Duration `json:"death_timer" bson:"death_timer"`
	HeistCost    int           `json:"heist_cost" bson:"heist_cost"`
	PoliceAlert  time.Duration `json:"police_alert" bson:"police_alert"`
	SentenceBase time.Duration `json:"sentence_base" bson:"sentence_base"`
	Targets      string        `json:"targets" bson:"targets"`
	Theme        string        `json:"theme" bson:"theme"`
	WaitTime     time.Duration `json:"wait_time" bson:"wait_time"`
}

// GetConfig retrieves the heist configuration for the specified guild. If
// the configuration does not exist, nil is returned.
func GetConfig(guild *guild.Guild) *Config {
	config := configs[guild.GuildID]
	if config != nil {
		return config
	}
	config = readConfig(guild)
	if config != nil {
		configs[config.GuildID] = config
		return config
	}
	return nil
}

// NewConfig creates a new default configuration for the specified guild.
func NewConfig(guild *guild.Guild) *Config {
	defaultTheme := os.Getenv("HEIST_DEFAULT_THEME")
	config := &Config{
		GuildID:      guild.GuildID,
		AlertTime:    time.Time{},
		BailBase:     BAIL_BASE,
		CrewOutput:   CREW_OUTPUT,
		DeathTimer:   DEATH_TIMER,
		HeistCost:    HEIST_COST,
		PoliceAlert:  POLICE_ALERT,
		SentenceBase: SENTENCE_BASE,
		Targets:      defaultTheme,
		Theme:        defaultTheme,
		WaitTime:     WAIT_TIME,
	}
	configs[config.GuildID] = config

	writeConfig(config)

	return config
}

// String returns a string representation of the heist configuration
func (config *Config) String() string {
	out, _ := json.Marshal(config)
	return string(out)
}
