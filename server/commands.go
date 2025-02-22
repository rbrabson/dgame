package server

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/internal/discmsg"
	log "github.com/sirupsen/logrus"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"server": server,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "server",
			Description: "Commands used to configure the bot for a given server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "role",
					Description: "Manages the admin roles for the bot for this server.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Returns the list of admin roles for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "add",
							Description: "Adds an admin role for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "rname",
									Description: "The name of the role to add.",
									Required:    true,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Removes an admin role for this server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to remove.",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
	memberCommands = []*discordgo.ApplicationCommand{}
)

// server handles the server command.
func server(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.server")
	defer log.Trace("<-- server.server")

	options := i.ApplicationCommandData().Options
	if options[0].Name == "role" {
		role(s, i)
	}
}

// role handles the role subcommands for the server command.
func role(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.role")
	defer log.Trace("<-- server.role")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "add":
		addRole(s, i)
	case "list":
		listRoles(s, i)
	case "remove":
		removeRole(s, i)
	}
}

// addRole adds a role to the list of admin roles for the server.
func addRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.addRole")
	defer log.Trace("<-- server.addRole")

	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := GetServer(guildID)

	// Add the role to the server configuration
	server.AddAdminRole(roleName)

	// Save the server configuration
	writeServer(server)
	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/server role add")
	discmsg.SendResponse(s, i, fmt.Sprintf("Role \"%s\" added", roleName))
}

// removeRole removes a role from the list of admin roles for the server.
func removeRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.removeRole")
	defer log.Trace("<-- server.removeRole")

	guildID := i.GuildID
	options := i.ApplicationCommandData().Options[0].Options
	roleName := options[0].StringValue()

	// Get the server configuration
	server := GetServer(guildID)

	// Remove the role from the server configuration
	server.RemoveAdminRole(roleName)

	// Save the server configuration
	writeServer(server)

	log.WithFields(log.Fields{"guild": guildID, "role": roleName}).Debug("/server role remove")
	discmsg.SendResponse(s, i, fmt.Sprintf("Role \"%s\" removed", roleName))
}

// listRoles lists the admin roles for the server.
func listRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> server.listRoles")
	defer log.Trace("<-- server.listRoles")

	guildID := i.GuildID

	// Get the server configuration
	server := GetServer(guildID)

	// Get the list of admin roles
	roles := server.GetAdminRoles()

	// Send the list of admin roles to the user
	var sb strings.Builder
	sb.WriteString("Admin Roles:\n")
	for _, role := range roles {
		sb.WriteString(role + "\n")
	}
	roleList := sb.String()
	log.WithFields(log.Fields{"guild": guildID, "roles": roleList}).Debug("/server role list")

	discmsg.SendResponse(s, i, roleList)
}
