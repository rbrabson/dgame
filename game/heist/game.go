package heist

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/rbrabson/dgame/bank"
	"github.com/rbrabson/dgame/guild"
	"github.com/rbrabson/dgame/internal/format"
	log "github.com/sirupsen/logrus"
)

const (
	VAULT_UPDATE_TIME     = 1 * time.Minute // Update the vault once every minute
	VAULT_RECOVER_PERCENT = 1.04            // Percentage of valuts total that is recovered every update
)

var (
	heists = Heists{
		heists: make(map[string]*Heist),
		mutex:  sync.Mutex{},
	}
)

type Heists struct {
	heists map[string]*Heist
	mutex  sync.Mutex
}

// plan begins the planning phase for the heist. During this phase, other members may
// join the heist.
func plan(heist *Heist) error {
	log.Trace("--> heist.plan")
	defer log.Trace("<-- heist.plan")

	// TODO: implement
	return nil
}

// start begins the execution phase for the heist. During this phase, no one can join the
// heist and the heist plays out until it is completed.
func start(heist *Heist) error {
	log.Trace("--> heist.start")
	defer log.Trace("<-- heist.start")

	// TODO: implement

	return nil
}

// join adds the given memmber to the heist crew
func join(heist *Heist, member *Member) error {
	log.Trace("--> heist.join")
	defer log.Trace("<-- heist.join")

	// TODO: implement

	return nil
}

// reset resets a hung heist
func reset(guild *guild.Guild) error {
	log.Trace("--> heist.reset")
	defer log.Trace("<-- heist.reset")

	heist := heists.heists[guild.GuildID]
	if heist == nil {
		log.WithFields(log.Fields{"guild": guild.GuildID}).Warn("heist not found")
		return ErrNoHeist
	}

	delete(heists.heists, guild.GuildID)
	log.WithFields(log.Fields{"guild": guild.GuildID}).Info("heist reset")

	return nil
}

// newHeist creates a new heist for the guild if one does not already exist
func newHeist(guild *guild.Guild, member *guild.Member) (*Heist, error) {
	log.Trace("--> heist.newHeist")
	defer log.Trace("<-- heist.newHeist")

	heists.mutex.Lock()
	defer heists.mutex.Unlock()

	// Check if the heist is already in progress
	heist := heists.heists[guild.GuildID]
	if heist != nil {
		log.WithFields(log.Fields{"guild": guild.GuildID}).Warn("heist already in progress")
		return nil, ErrHeistInProgress
	}

	// Create a new heist for the guild
	heist = NewHeist(guild, member)
	heist.CrewIDs = append(heist.CrewIDs, member.MemberID)
	heists.heists[guild.GuildID] = heist
	log.WithFields(log.Fields{"guild": guild.GuildID}).Info("heist created")

	return heist, nil
}

// deleteHeist deletes the heist from the list of heists
func deleteHeist(heist *Heist) {
	log.Trace("--> heist.deleteHeist")
	defer log.Trace("<-- heist.deleteHeist")

	heists.mutex.Lock()
	defer heists.mutex.Unlock()

	delete(heists.heists, heist.GuildID)
	log.WithFields(log.Fields{"guild": heist.GuildID}).Info("heist deleted")
}

// getTarget returns the heist target given the number of crew members
func getTarget(heist *Heist, targets []*Target) *Target {
	log.Trace("--> heist.getTarget")
	defer log.Trace("<-- heist.getTarget")

	crewSize := len(heist.CrewIDs)
	var target *Target
	for _, possibleTarget := range targets {
		if possibleTarget.CrewSize >= crewSize {
			if target == nil || target.CrewSize > possibleTarget.CrewSize {
				target = possibleTarget
			}
		}
	}

	log.WithFields(log.Fields{"guild": target.GuildID, "target": target.TargetID}).Debug("heist target selected")
	return target
}

// heistChecks returns an error, with appropriate message, if a heist cannot be started.
func heistChecks(g *guild.Guild, member *guild.Member, theme *Theme) (string, bool) {
	log.Trace("--> heist.heistChecks")
	defer log.Trace("<-- heist.heistChecks")

	targets := targets[g.GuildID]
	if len(targets) == 0 {
		msg := "Oh no! There are no targets!"
		return msg, false
	}

	heists.mutex.Lock()
	defer heists.mutex.Unlock()
	heist := heists.heists[g.GuildID]
	if heist == nil {
		msg := "There is no heist in progress. You need to start a heist first."
		return msg, false
	}
	heist.mutex.Lock()
	isMember := slices.Contains(heist.CrewIDs, member.MemberID)
	heist.mutex.Unlock()

	if isMember {
		msg := fmt.Sprintf("You are already in the %s.", theme.Crew)
		return msg, false
	}

	config := GetConfig(g)
	b := bank.GetBank(g.GuildID)
	account := b.GetAccount(member.MemberID)

	if account.CurrentBalance < config.HeistCost {
		msg := fmt.Sprintf("You do not have enough credits to cover the cost of entry. You need %d credits to participate", config.HeistCost)
		return msg, false
	}

	if config.AlertTime.After(time.Now()) {
		remainingTime := time.Until(config.AlertTime)
		msg := fmt.Sprintf("The %s are on high alert after the last target. We should wait for things to cool off before hitting another target. Time remaining: %s.", theme.Police, format.Duration(remainingTime))
		return msg,
			false
	}

	player := GetMember(g, member)
	if player.Status == APPREHENDED {
		if player.Status == OOB {
			if player.JailTimer.Before(time.Now()) {
				msg := fmt.Sprintf("Your %s is over, and you are no longer on probation! 3x penalty removed.", theme.Sentence)
				player.Clear()
				return msg, true
			}
			return "", true
		}
		if player.JailTimer.After(time.Now()) {
			remainingTime := time.Until(player.JailTimer)
			msg := fmt.Sprintf("You are in %s. You are serving a %s of %s.\nYou can wait out your remaining %s of %s, or pay %d credits to be released on %s.",
				theme.Jail, theme.Sentence, format.Duration(player.Sentence), theme.Sentence, format.Duration(remainingTime), player.BailCost, theme.Bail)
			return msg, false
		}
		msg := "You served your time. Enjoy the fresh air of freedom while you can."
		player.Clear()
		return msg, true
	}
	if player.Status == DEAD {
		if player.DeathTimer.After(time.Now()) {
			remainingTime := time.Until(player.DeathTimer)
			msg := fmt.Sprintf("You are dead. You will revive in %s", format.Duration(remainingTime))
			return msg, false
		}
		msg := "You have risen from the dead!."
		player.Clear()
		return msg, true
	}

	return "", true
}

// getHeistResults returns the results of the heist
func getHeistResults(guild *guild.Guild) *HeistResult {
	log.Trace("--> heist.getHeistResults")
	defer log.Trace("<-- heist.getHeistResults")

	// TODO: implement

	return nil
}

// handleHeistFailure updates the status of a player who is apprehended or killed during a heist.
func handleHeistFailure(g *guild.Guild, player *Member, result *HeistMemberResult) {
	log.Trace("--> handleHeistFailure")
	defer log.Trace("<-- handleHeistFailure")

	member := g.GetMember(player.MemberID)
	config := GetConfig(g)
	if result.Status == "apprehended" {
		sentence := int(config.SentenceBase) * (player.JailCounter + 1)
		bail := config.BailBase
		if player.Status == OOB {
			bail *= 3
		}
		player.BailCost = bail
		player.CriminalLevel++
		player.JailCounter++
		player.TotalJail++
		player.Sentence = time.Duration(sentence)
		player.JailTimer = time.Now().Add(player.Sentence)
		player.Spree = 0
		player.Status = APPREHENDED

		log.WithFields(log.Fields{
			"player":        member.Name,
			"bail":          player.BailCost,
			"criminalLevel": player.CriminalLevel,
			"jailCounter":   player.JailCounter,
			"jailTimier":    player.JailTimer,
			"sentence":      player.Sentence,
			"spree":         player.Spree,
			"status":        player.Status,
			"totalJail":     player.TotalJail,
		}).Debug("Apprehended")

		return
	}

	player.BailCost = 0
	player.CriminalLevel = 0
	player.DeathTimer = time.Now().Add(config.DeathTimer)
	player.JailCounter = 0
	player.JailTimer = time.Time{}
	player.Sentence = 0
	player.Spree = 0
	player.Status = DEAD
	player.Deaths++

	log.WithFields(log.Fields{
		"player":        member.Name,
		"bail":          player.BailCost,
		"criminalLevel": player.CriminalLevel,
		"deathTimer":    player.DeathTimer,
		"totalDeaths":   player.Deaths,
		"jailTimer":     player.JailTimer,
		"sentence":      player.Sentence,
		"spree":         player.Spree,
		"status":        player.Status,
	}).Debug("Dead")
}
