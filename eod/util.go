package eod

import (
	_ "embed"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/bwmarrin/discordgo"
)

// Unneeded for now
/*func (b *EoD) getRoles(userID string, guild string) ([]*discordgo.Role, error) {
	user, err := b.dg.GuildMember(guild, userID)
	if err != nil {
		return nil, err
	}
	hasLoadedRoles := false
	var roles []*discordgo.Role
	out := make([]*discordgo.Role, len(user.Roles))

	for i, roleID := range user.Roles {
		role, err := b.dg.State.Role(guild, roleID)
		if err != nil {
			if !hasLoadedRoles {
				roles, err = b.dg.GuildRoles(guild)
				if err != nil {
					return nil, err
				}
			}

			for _, role := range roles {
				if role.ID == roleID {
					roles[i] = role
				}
			}
		} else {
			roles[i] = role
		}
	}
	return out, nil
}*/

func (b *EoD) isMod(userID string, guildID string, m types.Msg) (bool, error) {
	lock.RLock()
	dat, inited := b.dat[guildID]
	lock.RUnlock()

	user, err := b.dg.GuildMember(m.GuildID, userID)
	if err != nil {
		return false, err
	}
	if (user.Permissions * discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator {
		return true, nil
	}

	hasLoadedRoles := false
	var roles []*discordgo.Role

	for _, roleID := range user.Roles {
		if inited && (roleID == dat.ModRole) {
			return true, nil
		}
		role, err := b.dg.State.Role(guildID, roleID)
		if err != nil {
			if !hasLoadedRoles {
				roles, err = b.dg.GuildRoles(m.GuildID)
				if err != nil {
					return false, err
				}
				hasLoadedRoles = true
			}

			for _, role := range roles {
				if role.ID == roleID && ((role.Permissions & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator) {
					return true, nil
				}
			}
		} else {
			if (role.Permissions & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator {
				return true, nil
			}
		}
	}
	return false, nil
}

func splitByCombs(inp string) []string {
	for _, val := range combs {
		if strings.Contains(inp, val) {
			return strings.Split(inp, val)
		}
	}
	return []string{inp}
}

func (b *EoD) getMessageElem(id string, guild string) (string, bool) {
	lock.RLock()
	dat, exists := b.dat[guild]
	lock.RUnlock()
	if !exists {
		return "Guild not setup yet!", false
	}
	el, res := dat.GetMsgElem(id)
	if !res.Exists {
		return res.Message, false
	}
	return el, true
}

//go:embed fools.txt
var foolsRaw string
