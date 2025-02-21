package heist

import (
	"cmp"
	"encoding/json"

	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	TARGET = "target"
)

var (
	targets = make(map[string][]*Target)

	// Sort the targets alphabetically
	sortTargets = func(t1, t2 *Target) int {
		return cmp.Compare(t1.CrewSize, t2.CrewSize)
	}
)

// Target is a target of a heist.
type Target struct {
	ID       string  `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID  string  `json:"guild_id" bson:"guild_id"`
	TargetID string  `json:"target_id" bson:"target_id"`
	CrewSize int     `json:"crew" bson:"crew"`
	Success  float64 `json:"success" bson:"success"`
	Vault    int     `json:"vault" bson:"vault"`
	VaultMax int     `json:"vault_max" bson:"vault_max"`
}

// NewTarget creates a new target for a heist
func NewTarget(guild *guild.Guild, themeID string, targetID string, maxCrewSize int, success float64, vaultCurrent int, maxVault int) *Target {
	log.Debug("--> heist.NewTarget")
	defer log.Debug("<-- heist.NewTarget")

	target := Target{
		GuildID:  guild.GuildID,
		TargetID: targetID,
		CrewSize: maxCrewSize,
		Success:  success,
		Vault:    vaultCurrent,
		VaultMax: maxVault,
	}
	return &target
}

// ReadTargets loads the targets that may be used in heists by the given guild
func ReadTargets(guild *guild.Guild) {
	log.Debug("--> heist.ReadTargets")
	defer log.Debug("<-- heist.ReadTargets")

	filter := make(map[string]interface{})
	filter["guild_id"] = guild.GuildID
	targetIDs, _ := db.ListDocuments(TARGET, filter)

	heistTargets := make([]*Target, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		filter["_id"] = targetID
		var target *Target
		db.Read(TARGET, filter, target)
		heistTargets = append(heistTargets, target)
	}

	log.WithFields(log.Fields{"guild": guild.GuildID, "targets": targets}).Trace("load targets")
	targets[guild.GuildID] = heistTargets
}

// WriteTarget writes the set of targets to the database. If they already exist, the are updated; otherwise, the set is created.
func WriteTarget(target *Target) {
	log.Debug("--> heist.Target.WriteTarget")
	defer log.Debug("<-- heist.Target.WriteTarget")

	filter := make(map[string]interface{})
	filter["guild_id"] = target.GuildID
	filter["target_id"] = target.TargetID
	db.Write(TARGET, filter, targets)
	log.WithFields(log.Fields{"guild": target.GuildID, "target": target.TargetID}).Trace("save target")
}

// GetTargets returns the list of targets for the server
func GetTargets(guild *guild.Guild) []*Target {
	log.Debug("--> heist.GetTargets")
	defer log.Debug("<-- heist.GetTargets")

	heistTargets := targets[guild.GuildID]
	if heistTargets == nil {
		log.WithFields(log.Fields{"guild": guild.GuildID}).Warning("targets not found")
		return nil
	}

	log.WithFields(log.Fields{"guild": guild.GuildID, "targets": heistTargets}).Trace("get targets")
	return heistTargets
}

// vaultUpdater updates the vault balance for any target whose vault is not at the maximum value
func vaultUpdater() {
	// Update the vaults forever
	for {
		for _, guildTargets := range targets {
			for _, target := range guildTargets {
				newVaultAmount := int(float64(target.Vault) * VAULT_RECOVER_PERCENT)
				newVaultAmount = min(newVaultAmount, target.VaultMax)
				if newVaultAmount != target.Vault {
					WriteTarget(target)
					log.WithFields(log.Fields{"guild": target.GuildID, "target": target.TargetID, "Old": target.Vault, "New": newVaultAmount, "Max": target.VaultMax}).Debug("update vault")
				}
			}
		}
	}
}

// String returns a string representation of the target.
func (t *Target) String() string {
	out, _ := json.Marshal(t)
	return string(out)
}
