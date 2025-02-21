package guild

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/database/mongo"
	"github.com/rbrabson/dgame/discord"
)

const (
	PLUGIN_NAME = "bank"
)

var (
	plugin *Plugin
	db     *mongo.MongoDB
)

// Plugin is the plugin for the banking system used by the bot
type Plugin struct{}

// init creates and registers the plugin for the banking system
func init() {
	plugin = &Plugin{}
	discord.RegisterPlugin(plugin)
}

// Initialize saves the Discord bot to be used by the banking system
func (plugin *Plugin) Initialize(b *discord.Bot, d *mongo.MongoDB) {
	db = d
}

// GetCommands returns the commands for the banking system
func (plugin *Plugin) GetCommands() []*discordgo.ApplicationCommand {
	return nil
}

// GetCommandHandlers returns the command handlers for the banking system
func (plugin *Plugin) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return nil
}

// GetComponentHandlers returns the component handlers for the banking system
func (plugin *Plugin) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return nil
}

// GetName returns the name of the banking system plugin
func (plugin *Plugin) GetName() string {
	return PLUGIN_NAME
}

// GetHelp returns the member help for the banking system
func (plugin *Plugin) GetHelp() []string {
	return nil
}

// GetAdminHelp returns the admin help for the banking system
func (plugin *Plugin) GetAdminHelp() []string {
	return nil
}
