package heist

import (
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	THEME = "theme"
)

var (
	themes = make(map[string]map[string]*Theme)
)

// A Theme is a set of messages that provide a "flavor" for a heist
type Theme struct {
	ID       string        `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID  string        `json:"guild_id" bson:"guild_id"`
	ThemeID  string        `json:"theme_id" bson:"theme_id"`
	Good     []GoodMessage `json:"good"`
	Bad      []BadMessage  `json:"bad"`
	Jail     string        `json:"jail" bson:"jail"`
	OOB      string        `json:"oob" bson:"oob"`
	Police   string        `json:"police" bson:"police"`
	Bail     string        `json:"bail" bson:"bail"`
	Crew     string        `json:"crew" bson:"crew"`
	Sentence string        `json:"sentence" bson:"sentence"`
	Heist    string        `json:"heist" bson:"heist"`
	Vault    string        `json:"vault" bson:"vault"`
}

// A GoodMessage is a message for a successful heist outcome
type GoodMessage struct {
	Message string `json:"message" bson:"message"`
	Amount  int    `json:"amount" bson:"amount"`
}

// A BadMessage is a message for a failed heist outcome
type BadMessage struct {
	Message string       `json:"message" bson:"message"`
	Result  MemberStatus `json:"result" bson:"result"`
}

// LoadThemes loads all available themes for a guild
func LoadThemes(guild *guild.Guild) map[string]*Theme {
	log.Trace("--> heist.LoadThemes")
	defer log.Trace("<-- heist.LoadThemes")

	heistTheme := make(map[string]*Theme)
	themes[guild.GuildID] = heistTheme

	filter := make(map[string]interface{})
	filter["guild_id"] = guild.GuildID

	themeIDs, _ := db.ListDocuments(THEME, filter)
	for _, themeID := range themeIDs {
		filter["_id"] = themeID
		var theme Theme
		db.Read(THEME, filter, &theme)
		theme.GuildID = guild.GuildID
		heistTheme[theme.ThemeID] = &theme
		log.WithFields(log.Fields{"guild": theme.GuildID, "theme": theme.ThemeID}).Debug("load theme from database")
	}

	return heistTheme
}

// GetThemeNames returns a list of available themes
func GetThemeNames(guild *guild.Guild) ([]string, error) {
	var fileNames []string
	for _, theme := range themes[guild.GuildID] {
		fileNames = append(fileNames, theme.ThemeID)
	}

	return fileNames, nil
}

// GetTheme returns the theme for a guild
func GetTheme(g *guild.Guild) (*Theme, error) {
	log.Trace("--> heist.GetTheme")
	defer log.Trace("<-- heist.GetTheme")

	config := GetConfig(g)
	theme := themes[g.GuildID][config.Theme]
	if theme == nil {
		return nil, ErrThemeNotFound
	}

	return theme, nil
}

// Write creates or updates the theme in the database
func (theme *Theme) Write() {
	log.Trace("--> heist.Theme.Write")
	defer log.Trace("<-- heist.Theme.Write")

	filter := make(map[string]interface{})
	filter["guild_id"] = theme.GuildID
	filter["theme_id"] = theme.ThemeID
	db.Write(THEME, filter, theme)
}
