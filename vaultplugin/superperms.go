package vaultplugin

import (
	"errors"
	"sync"

	"github.com/golem-mc/golem/plugin"
)

// superPerms is the Permission service when no permission plugin provides
// one: it answers an online player's golem permission checks and changes
// nothing but its transient permissions. Vault: Permission_SuperPerms
type superPerms struct {
	s plugin.Server
	*attachments
}

// attachments holds the one attachment each online player's transient
// permissions live on; a rejoined player's old one is replaced.
type attachments struct {
	mu       sync.Mutex
	attached map[plugin.UUID]attached
}

type attached struct {
	player plugin.Player
	att    plugin.PermissionAttachment
}

func newSuperPerms(s plugin.Server) superPerms {
	return superPerms{s, &attachments{attached: map[plugin.UUID]attached{}}}
}

func (superPerms) Name() string           { return "SuperPerms" }
func (superPerms) SuperPermsCompat() bool { return true }
func (superPerms) GroupSupport() bool     { return false }

// PlayerHas asks an online player's HasPermission; an offline player has
// nothing.
func (p superPerms) PlayerHas(_ plugin.World, player plugin.UUID, node string) bool {
	online, ok := p.s.Player(player)
	return ok && online.HasPermission(node)
}

func (superPerms) PlayerAdd(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (superPerms) PlayerRemove(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

// PlayerAddTransient sets node to true on the player's attachment; it lasts
// until PlayerRemoveTransient, or until the player leaves. An offline player
// has no attachment to put it on.
// Vault: Permission_SuperPerms.playerAddTransient
func (p superPerms) PlayerAddTransient(_ plugin.World, player plugin.UUID, node string) error {
	online, ok := p.s.Player(player)
	if !ok {
		return errors.ErrUnsupported
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	held, ok := p.attached[player]
	if !ok || held.player != online {
		att := online.AddAttachment()
		if att == nil {
			return errors.ErrUnsupported
		}
		held = attached{online, att}
		p.attached[player] = held
	}
	held.att.Set(node, true)
	return nil
}

// PlayerRemoveTransient unsets node from the player's attachment.
// Vault: Permission_SuperPerms.playerRemoveTransient
func (p superPerms) PlayerRemoveTransient(_ plugin.World, player plugin.UUID, node string) error {
	online, ok := p.s.Player(player)
	if !ok {
		return errors.ErrUnsupported
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if held, ok := p.attached[player]; ok && held.player == online {
		held.att.Unset(node)
	}
	return nil
}

func (superPerms) GroupHas(plugin.World, string, string) bool { return false }
func (superPerms) GroupAdd(plugin.World, string, string) error {
	return errors.ErrUnsupported
}

func (superPerms) GroupRemove(plugin.World, string, string) error {
	return errors.ErrUnsupported
}

// PlayerInGroup reads membership as the node groups.<group>.
func (p superPerms) PlayerInGroup(w plugin.World, player plugin.UUID, group string) bool {
	return p.PlayerHas(w, player, "groups."+group)
}

func (superPerms) PlayerAddGroup(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (superPerms) PlayerRemoveGroup(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (superPerms) PlayerGroups(plugin.World, plugin.UUID) []string { return nil }
func (superPerms) PrimaryGroup(plugin.World, plugin.UUID) string   { return "" }
func (superPerms) Groups() []string                                { return nil }
