package vaultplugin

import (
	"errors"

	"github.com/golem-mc/golem/plugin"
)

// superPerms is the Permission service when no permission plugin provides
// one: it answers an online player's golem permission checks and changes
// nothing. Vault: Permission_SuperPerms
type superPerms struct{ s plugin.Server }

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

// PlayerAddTransient and PlayerRemoveTransient are unsupported: golem has no
// permission attachments to put them on.
func (superPerms) PlayerAddTransient(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (superPerms) PlayerRemoveTransient(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
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
