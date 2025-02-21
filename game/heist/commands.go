package heist

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/olekukonko/tablewriter"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/bank"
	"github.com/rbrabson/dgame/channel"
	"github.com/rbrabson/dgame/guild"
	"github.com/rbrabson/dgame/internal/discmsg"
	"github.com/rbrabson/dgame/internal/format"
	log "github.com/sirupsen/logrus"
)

// componentHandlers are the buttons that appear on messages sent by this bot.
var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"heist":       heist,
		"heist-admin": admin,
		"join_heist":  joinHeist,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist-admin",
			Description: "Heist admin commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "clear",
					Description: "Clears the criminal settings for the user.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to clear",
							Required:    true,
						},
					},
				},
				{
					Name:        "config",
					Description: "Configures the Heist bot.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "info",
							Description: "Returns the configuration information for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "bail",
							Description: "Sets the base cost of bail.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The base cost of bail.",
									Required:    true,
								},
							},
						},
						{
							Name:        "cost",
							Description: "Sets the cost to plan or join a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The cost to plan or join a heist.",
									Required:    true,
								},
							},
						},
						{
							Name:        "death",
							Description: "Sets how long players remain dead.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the player remains dead, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "patrol",
							Description: "Sets the time the authorities will prevent a new heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the authorities will patrol, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "payday",
							Description: "Sets how many credits a player gets for each payday.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The amount deposited in a players account for each payday.",
									Required:    true,
								},
							},
						},
						{
							Name:        "sentence",
							Description: "Sets the base apprehension time when caught.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The base time, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "wait",
							Description: "Sets how long players can gather others for a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time to wait for players to join the heist, in seconds.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "theme",
					Description: "Commands that interact with the heist themes.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Gets the list of available heist themes.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "set",
							Description: "Sets the current heist theme.",
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "Name of the theme to set.",
									Required:    true,
								},
							},
							Type: discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
				{
					Name:        "reset",
					Description: "Resets a new heist that is hung.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist",
			Description: "Heist game commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "bail",
					Description: "Bail a player out of jail.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to bail. Defaults to you.",
							Required:    false,
						},
					},
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "start",
					Description: "Plans a new heist.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "targets",
					Description: "Gets the list of available heist targets.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// config routes the configuration commands to the proper handlers.
func config(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> config")
	defer log.Trace("<-- config")

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "cost":
		configCost(s, i)
	case "sentence":
		configSentence(s, i)
	case "patrol":
		configPatrol(s, i)
	case "bail":
		configBail(s, i)
	case "death":
		configDeath(s, i)
	case "wait":
		configWait(s, i)
	case "info":
		configInfo(s, i)
	}
}

// admin routes the commands to the subcommand and subcommandgroup handlers
func admin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> admin")
	defer log.Trace("<-- admin")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "clear":
		clearMember(s, i)
	case "config":
		config(s, i)
	case "reset":
		resetHeist(s, i)
	case "theme":
		theme(s, i)
	}
}

// heist routes the commands to the subcommand and subcommandgroup handlers
func heist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist")
	defer log.Trace("<-- heist")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "bail":
		bailoutPlayer(s, i)
	case "start":
		planHeist(s, i)
	case "stats":
		playerStats(s, i)
	case "targets":
		listTargets(s, i)
	}
}

// theme routes the theme commands to the proper handlers.
func theme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> theme")
	defer log.Trace("<-- theme")

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "list":
		listThemes(s, i)
	case "set":
		setTheme(s, i)
	}
}

// planHeist plans a new heist
func planHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> planHeist")
	defer log.Trace("<-- planHeist")

	// Create a new heist
	g := guild.GetGuild(i.GuildID)
	guildMember := g.GetMember(i.Member.User.ID).SetName(i.Member.User.Username, i.Member.DisplayName())
	heist, err := newHeist(g, guildMember)
	if err != nil {
		log.WithField("error", err).Error("unable to create the heist")
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}
	heist.session = s
	heist.interaction = i
	defer deleteHeist(heist)

	// Plan the heist
	err = plan(heist)
	if err != nil {
		log.WithField("error", err).Error("unable to plan the heist")
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}

	heistMessage(heist, "plan")

	// Wait for the heist to start
	err = waitHeist(heist)
	if err != nil {
		log.WithField("error", err).Error("unable to wait for the heist")
		discmsg.SendEphemeralResponse(s, i, err.Error())
		return
	}

	// Start the heist once the wait period expires
	startHeist(heist)
}

// waitHeist waits until the planning stage for the heist expires.
func waitHeist(heist *Heist) error {
	log.Trace("--> waitHeist")
	defer log.Trace("<-- waitHeist")

	// Get the theme for the heist
	guild := guild.GetGuild(heist.GuildID)
	theme, err := GetTheme(guild)
	if err != nil {
		return err
	}

	// Wait for the heist to be ready to start
	discmsg.SendEphemeralResponse(heist.session, heist.interaction, "Starting "+theme.Heist+"...")
	for !time.Now().After(heist.StartTime) {
		maximumWait := time.Until(heist.StartTime)
		timeToWait := min(maximumWait, 5*time.Second)
		if timeToWait < 0 {
			break
		}
		time.Sleep(timeToWait)
		err := heistMessage(heist, "update")
		if err != nil {
			log.WithField("error", err).Error("Unable to update the time for the heist message")
			continue
		}
	}

	return nil
}

// joinHeist attempts to join a heist that is being planned
func joinHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> joinHeist")
	defer log.Trace("<-- joinHeist")

	g := guild.GetGuild(i.GuildID)
	member := g.GetMember(i.Member.User.ID).SetName(i.Member.User.Username, i.Member.DisplayName())
	player := GetMember(g, member)
	config := GetConfig(g)
	themes := themes[g.GuildID]
	theme := themes[config.Theme]

	heist := heists.heists[g.GuildID]
	if heist == nil {
		discmsg.SendEphemeralResponse(s, i, "No "+theme.Heist+" is planned.")
		return
	}

	discmsg.SendEphemeralResponse(s, i, "Joining "+theme.Heist+"...")

	heist.mutex.Lock()
	isMember := slices.Contains(heist.CrewIDs, player.MemberID)
	heist.mutex.Unlock()
	if isMember {
		discmsg.EditResponse(s, i, "You are already a member of the "+theme.Heist+".")
		return
	}
	msg, ok := heistChecks(g, member, theme)
	if !ok {
		discmsg.EditResponse(s, i, msg)
		return
	}
	if heist.State == STARTED {
		discmsg.EditResponse(s, i, "The heist has already been started")
		return
	}

	heist.mutex.Lock()
	heist.CrewIDs = append(heist.CrewIDs, player.MemberID)
	heist.mutex.Unlock()
	err := heistMessage(heist, "join")
	if err != nil {
		log.Error("Unable to update the heist message, error:", err)
	}

	// Withdraw the cost of the heist from the player's account. We know the player already
	// as the required number of credits as this is verified in `heistChecks`.
	b := bank.GetBank(i.GuildID)
	account := b.GetAccount(player.MemberID)
	account.Withdraw(config.HeistCost)

	if msg != "" {
		resp := fmt.Sprintf("%s You have joined the %s at a cost of %d credits.", msg, theme.Heist, config.HeistCost)
		discmsg.EditResponse(s, i, resp)
	} else {
		resp := fmt.Sprintf("You have joined the %s at a cost of %d credits.", theme.Heist, config.HeistCost)
		discmsg.EditResponse(s, i, resp)
	}
}

// startHeist is called once the wait time for planning the heist completes
func startHeist(heist *Heist) {
	log.Trace("--> startHeist")
	defer log.Trace("<-- startHeist")

	g := guild.GetGuild(heist.GuildID)
	targets := targets[g.GuildID]
	if len(targets) == 1 {
		discmsg.SendEphemeralResponse(heist.session, heist.interaction, "There are no heist targets.")
		return
	}

	mute := channel.NewChannelMute(heist.session, heist.interaction)
	mute.MuteChannel()
	defer mute.UnmuteChannel()

	heist.State = STARTED

	err := heistMessage(heist, "start")
	if err != nil {
		log.Error("Unable to mark the heist message as started, error:", err)
	}
	theme, _ := GetTheme(g)
	if len(heist.CrewIDs) <= 1 {
		heistMessage(heist, "ended")
		msg := fmt.Sprintf("You tried to rally a %s, but no one wanted to follow you. The %s has been cancelled.", theme.Crew, theme.Heist)
		heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
		return
	}
	log.Debug("Heist is starting")
	msg := fmt.Sprintf("Get ready! The %s is starting with %d members.", theme.Heist, len(heist.CrewIDs))
	heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
	time.Sleep(3 * time.Second)
	heistMessage(heist, "start")
	target := getTarget(heist, targets)
	results := getHeistResults(g)
	log.Debug("Hitting " + target.ID)
	msg = fmt.Sprintf("The %s has decided to hit **%s**.", theme.Crew, target.ID)
	heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
	time.Sleep(3 * time.Second)

	// Process the results
	for _, result := range results.MemberResults {
		g := guild.GetGuild(heist.GuildID)
		member := g.GetMember(result.Player.MemberID)
		msg = fmt.Sprintf(result.Message+"\n", "**"+member.Name+"**")
		if result.Status == "apprehended" {
			msg += fmt.Sprintf("`%s dropped out of the game.`", member.Name)
		}
		heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
		time.Sleep(3 * time.Second)
	}

	if results.Escaped == 0 {
		msg = "\nNo one made it out safe."
		heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
	} else {
		msg = "\nThe raid is now over. Distributing player spoils."
		heist.session.ChannelMessageSend(heist.interaction.ChannelID, msg)
		// Render the results into a table and returnt he results.
		var tableBuffer strings.Builder
		table := tablewriter.NewWriter(&tableBuffer)
		table.SetBorder(false)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
		table.SetHeader([]string{"Player", "Loot", "Bonus", "Total"})
		for _, result := range results.SurvivingCrew {
			member := g.GetMember(result.Player.MemberID)
			data := []string{member.Name, fmt.Sprintf("%d", result.StolenCredits), fmt.Sprintf("%d", result.BonusCredits), fmt.Sprintf("%d", result.StolenCredits+result.BonusCredits)}
			table.Append(data)
		}
		table.Render()
		heist.session.ChannelMessageSend(heist.interaction.ChannelID, "```\n"+tableBuffer.String()+"```")
	}

	// Update the status for each player and then save the information
	for _, result := range results.MemberResults {
		member := g.GetMember(result.Player.MemberID)
		player := GetMember(g, member)
		if result.Status == "apprehended" || result.Status == "dead" {
			handleHeistFailure(g, player, result)
		} else {
			player.Spree++
		}
		b := bank.GetBank(heist.GuildID)
		if results.Escaped > 0 && result.StolenCredits != 0 {
			account := b.GetAccount(player.MemberID)
			account.Deposit(result.StolenCredits + result.BonusCredits)
			target.Vault -= int(result.StolenCredits)
			log.WithFields(log.Fields{"Member": member.Name, "Stolen": result.StolenCredits, "Bonus": result.BonusCredits}).Debug("heist loot")
		}
	}
	target.Vault = max(target.Vault, target.VaultMax*4/100)

	heistMessage(heist, "ended")

	// Update the heist status information
	config := GetConfig(g)
	config.AlertTime = time.Now().Add(config.PoliceAlert) // TODO: need to store this somewhere outside of the heist itself....
}

// playerStats shows a player's heist stats
func playerStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> playerStats")
	defer log.Trace("<-- playerStats")

	g := guild.GetGuild(i.GuildID)
	theme, _ := GetTheme(g)
	member := g.GetMember(i.Member.User.ID)
	player := GetMember(g, member)
	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))

	b := bank.GetBank(i.GuildID)
	account := b.GetAccount(member.MemberID)

	var sentence string
	if player.Status == APPREHENDED {
		if player.JailTimer.Before(time.Now()) {
			sentence = "Served"
		} else {
			timeRemaining := time.Until(player.JailTimer)
			sentence = format.Duration(timeRemaining)
		}
	} else {
		sentence = "None"
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       member.Name,
			Description: player.CriminalLevel.String(),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  strconv.Itoa(int(player.Status)),
					Inline: true,
				},
				{
					Name:   "Spree",
					Value:  fmt.Sprintf("%d", player.Spree),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Bail),
					Value:  fmt.Sprintf("%d", player.BailCost),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Sentence),
					Value:  sentence,
					Inline: true,
				},
				{
					Name:   "apprehended",
					Value:  fmt.Sprintf("%d", player.JailCounter),
					Inline: true,
				},
				{
					Name:   "Total Deaths",
					Value:  fmt.Sprintf("%d", player.Deaths),
					Inline: true,
				},
				{
					Name:   "Lifetime Apprehensions",
					Value:  fmt.Sprintf("%d", player.TotalJail),
					Inline: true,
				},
				{
					Name:   "Credits",
					Value:  fmt.Sprintf("%d", account.CurrentBalance),
					Inline: true,
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error("Unable to send the player stats to Discord, error:", err)
	}
}

// bailoutPlayer bails a player player out from jail. This defaults to the player initiating the command, but can
// be another player as well.
func bailoutPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> bailoutPlayer")
	log.Trace("<-- bailoutPlayer")

	var playerID string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "id" {
			playerID = strings.TrimSpace(option.StringValue())
		}
	}

	g := guild.GetGuild(i.GuildID)
	member := g.GetMember(i.Member.User.ID)
	initiatingPlayer := GetMember(g, member)
	b := bank.GetBank(i.GuildID)
	account := b.GetAccount(member.MemberID)

	discmsg.SendEphemeralResponse(s, i, "Bailing "+playerID+"...")
	var player *Member
	if playerID != "" {
		player = GetMember(g, member)
	} else {
		player = initiatingPlayer
	}

	if player.Status != APPREHENDED && player.Status != OOB {
		var msg string
		if player.ID == i.Member.User.ID {
			msg = "You are not in jail"
		} else {
			msg = fmt.Sprintf("%s is not in jail", member.Name)
		}
		discmsg.EditResponse(s, i, msg)
		return
	}
	if player.Status == APPREHENDED && player.JailTimer.Before(time.Now()) {
		discmsg.EditResponse(s, i, "You have already served your sentence.")
		player.Reset()
		return
	}
	if account.CurrentBalance < player.BailCost {
		msg := fmt.Sprintf("You do not have enough credits to play the bail of %d", player.BailCost)
		discmsg.EditResponse(s, i, msg)
		return
	}

	account.Withdraw(player.BailCost)
	player.Status = OOB

	var msg string
	if player.ID == initiatingPlayer.ID {
		msg = fmt.Sprintf("Congratulations, you are now free! You spent %d credits on your bail. Enjoy your freedom while it lasts.", player.BailCost)
		discmsg.EditResponse(s, i, msg)
	} else {
		playerMember := g.GetMember(player.MemberID)
		initiatingPlayerMember := g.GetMember(initiatingPlayer.MemberID)
		msg = fmt.Sprintf("Congratulations, %s, %s bailed you out by spending %d credits and now you are free!. Enjoy your freedom while it lasts.", playerMember.Name, initiatingPlayerMember.Name, player.BailCost)
		discmsg.EditResponse(s, i, msg)
	}
}

// heistMessage sends the main command used to plan, join and leave a heist. It also handles the case where
// the heist starts, disabling the buttons to join/leave/cancel the heist.
func heistMessage(heist *Heist, action string) error {
	log.Trace("--> heistMessage")
	defer log.Trace("<-- heistMessage")

	g := guild.GetGuild(heist.GuildID)
	member := g.GetMember(heist.MemberID)
	var status string
	var buttonDisabled bool
	switch action {
	case "plan", "join", "leave":
		until := time.Until(heist.StartTime)
		status = "Starts in " + format.Duration(until)
		buttonDisabled = false
	case "update":
		until := time.Until(heist.StartTime)
		status = "Starts in " + format.Duration(until)
		buttonDisabled = false
	case "start":
		status = "Started"
		buttonDisabled = true
	case "cancel":
		status = "Canceled"
		buttonDisabled = true
	default:
		status = "Ended"
		buttonDisabled = true
	}

	heist.mutex.Lock()
	crew := make([]string, 0, len(heist.CrewIDs))
	for _, id := range heist.CrewIDs {
		crewMember := g.GetMember(id)
		crew = append(crew, crewMember.Name)
	}
	heist.mutex.Unlock()

	config := GetConfig(g)
	theme, _ := GetTheme(g)
	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))
	msg := fmt.Sprintf("A new %s is being planned by %s. You can join the %s for a cost of %d credits at any time prior to the %s starting.", theme.Heist, member.Name, theme.Heist, config.HeistCost, theme.Heist)
	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Heist",
			Description: msg,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  status,
					Inline: true,
				},
				{
					Name:   fmt.Sprintf("%s (%d members)", caser.String(theme.Crew), len(crew)),
					Value:  strings.Join(crew, ", "),
					Inline: true,
				},
			},
		},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Join",
				Style:    discordgo.SuccessButton,
				Disabled: buttonDisabled,
				CustomID: "join_heist",
				Emoji:    nil,
			},
		}},
	}
	emptymsg := ""
	_, err := heist.session.InteractionResponseEdit(heist.interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
		Content:    &emptymsg,
	})
	if err != nil {
		return err
	}

	return nil
}

/******** ADMIN COMMANDS ********/

// Reset resets the heist in case it hangs
func resetHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> resetHeist")
	defer log.Trace("<-- resetHeist")

	mute := channel.NewChannelMute(s, i)
	defer mute.UnmuteChannel()

	g := guild.GetGuild(i.GuildID)
	theme, _ := GetTheme(g)
	heist := heists.heists[g.GuildID]
	if heist == nil {
		discmsg.SendEphemeralResponse(s, i, "No "+theme.Heist+" is being planned.")
		return
	}

	heistMessage(heist, "cancel")
	deleteHeist(heist)
	discmsg.SendResponse(s, i, "The "+theme.Heist+" has been reset.")
}

// listTargets displays a list of available heist targets.
func listTargets(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> listTargets")
	defer log.Trace("<-- listTargets")

	g := guild.GetGuild(i.GuildID)
	theme, _ := GetTheme(g)
	targets := GetTargets(g)

	if len(targets) == 0 {
		msg := "There aren't any targets!"
		discmsg.SendEphemeralResponse(s, i, msg)
		return
	}

	sortedTargets := slices.Clone(targets)
	slices.SortFunc(sortedTargets, sortTargets)

	// Lets return the data in an Ascii table. Ideally, it would be using a Discord embed, but unfortunately
	// Discord only puts three columns per row, which isn't enough for our purposes.
	var tableBuffer strings.Builder
	table := tablewriter.NewWriter(&tableBuffer)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"ID", "Max Crew", theme.Vault, "Max " + theme.Vault, "Success Rate"})
	for _, target := range sortedTargets {
		data := []string{target.ID, fmt.Sprintf("%d", target.CrewSize), fmt.Sprintf("%d", target.Vault), fmt.Sprintf("%d", target.VaultMax), fmt.Sprintf("%.2f", target.Success)}
		table.Append(data)
	}
	table.Render()

	discmsg.SendEphemeralResponse(s, i, "```\n"+tableBuffer.String()+"\n```")
}

// clearMember clears the criminal state of the player.
func clearMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> clearMember")
	log.Trace("<-- clearMember")

	memberID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	g := guild.GetGuild(i.GuildID)
	member := g.GetMember(memberID)
	player := GetMember(g, member)
	player.Reset()
	discmsg.SendResponse(s, i, "Player \""+member.Name+"\"'s settings cleared.")
}

// listThemes returns the list of available themes that may be used for heists
func listThemes(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> listThemes")
	defer log.Trace("<-- listThemes")

	g := guild.GetGuild(i.GuildID)
	themes, err := GetThemeNames(g)
	if err != nil {
		log.Warning("Unable to get the themes, error:", err)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Available Themes",
			Description: "Available Themes for the Heist bot",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Themes",
					Value:  strings.Join(themes[:], ", "),
					Inline: true,
				},
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
		},
	})
	if err != nil {
		log.Error("Unable to send list of themes to the user, error:", err)
	}
}

// setTheme sets the heist theme to the one specified in the command
func setTheme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setTheme")
	defer log.Trace("<-- setTheme")

	g := guild.GetGuild(i.GuildID)
	var themeName string
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	for _, option := range options {
		if option.Name == "name" {
			themeName = strings.TrimSpace(option.StringValue())
		}
	}

	config := GetConfig(g)
	if themeName == config.Theme {
		discmsg.SendEphemeralResponse(s, i, "Theme `"+themeName+"` is already being used.")
		return
	}
	theme, err := GetTheme(g)
	if err != nil {
		r := []rune(err.Error())
		r[0] = unicode.ToUpper(r[0])
		str := string(r)
		discmsg.SendEphemeralResponse(s, i, str)
		return
	}
	config.Theme = theme.ID
	log.Debug("Now using theme ", config.Theme)

	discmsg.SendResponse(s, i, "Theme "+themeName+" is now being used.")
}

// configCost sets the cost to plan or join a heist
func configCost(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configCost")
	defer log.Trace("<-- configCost")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	cost := options[0].IntValue()
	config.HeistCost = int(cost)

	discmsg.SendResponse(s, i, fmt.Sprintf("Cost set to %d", cost))
	// TODO: save the configuration
}

// configSentence sets the base aprehension time when a player is apprehended.
func configSentence(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configSentence")
	defer log.Trace("<-- configSentence")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	sentence := i.ApplicationCommandData().Options[0].Options[0].IntValue()
	config.SentenceBase = time.Duration(sentence * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Sentence set to %d", sentence))

	// TODO: save the config
}

// configPatrol sets the time authorities will prevent a new heist following one being completed.
func configPatrol(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configPatrol")
	defer log.Trace("<-- configPatrol")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	patrol := options[0].IntValue()
	config.PoliceAlert = time.Duration(patrol * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Patrol set to %d", patrol))

	// TODO: save the config
}

// configBail sets the base cost of bail.
func configBail(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configBail")
	defer log.Trace("<-- configBail")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	bail := options[0].IntValue()
	config.BailBase = int(bail)

	discmsg.SendResponse(s, i, fmt.Sprintf("Bail set to %d", bail))

	// TODO: save the config
}

// configDeath sets how long players remain dead.
func configDeath(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configDeath")
	defer log.Trace("<-- configDeath")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	death := options[0].IntValue()
	config.PoliceAlert = time.Duration(death * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Death set to %d", death))

	// TODO: save the config
}

// configWait sets how long players wait for others to join the heist.
func configWait(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configWait")
	defer log.Trace("<-- configWait")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	wait := options[0].IntValue()
	config.WaitTime = time.Duration(wait * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Wait set to %d", wait))

	// TODO: save the config
}

// configInfo returns the configuration for the Heist bot on this server.
func configInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configInfo")
	defer log.Trace("<-- configInfo")

	g := guild.GetGuild(i.GuildID)
	config := GetConfig(g)

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "bail",
				Value:  fmt.Sprintf("%d", config.BailBase),
				Inline: true,
			},
			{
				Name:   "cost",
				Value:  fmt.Sprintf("%d", config.HeistCost),
				Inline: true,
			},
			{
				Name:   "death",
				Value:  fmt.Sprintf("%.f", config.DeathTimer.Seconds()),
				Inline: true,
			},
			{
				Name:   "patrol",
				Value:  fmt.Sprintf("%.f", config.PoliceAlert.Seconds()),
				Inline: true,
			},
			{
				Name:   "sentence",
				Value:  fmt.Sprintf("%.f", config.SentenceBase.Seconds()),
				Inline: true,
			},
			{
				Name:   "wait",
				Value:  fmt.Sprintf("%.f", config.WaitTime.Seconds()),
				Inline: true,
			},
		},
	}

	embeds := []*discordgo.MessageEmbed{
		embed,
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Heist Configuration",
			Embeds:  embeds,
		},
	})
	if err != nil {
		log.Error("Unable to send a response, error:", err)
	}
}
