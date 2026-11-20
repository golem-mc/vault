package vault

import "github.com/golem-mc/golem/plugin"

// Permission is a permission plugin's players and groups, for other plugins
// to query and change, online players or not. A plain permission check of
// an online player or the console needs none of it:
// CommandSender.HasPermission asks the permission plugin through golem's
// plugin.PermissionProvider. Vault: Permission
//
// world is the world a query or change is about. nil means no particular
// world: a query then uses the player's current context (their world when
// online), and a change applies in every world. A change answers
// errors.ErrUnsupported when the plugin cannot make it, such as a group
// change without group support.
type Permission interface {
	// Name answers the permission plugin's name, as /vault-info shows it.
	// Vault: getName
	Name() string
	// SuperPermsCompat reports whether the plugin also decides golem's own
	// permission checks, so that CommandSender.HasPermission agrees with
	// PlayerHas. Vault: hasSuperPermsCompat
	SuperPermsCompat() bool
	// GroupSupport reports whether the plugin has groups. Without, the group
	// queries answer nothing and the group changes fail.
	// Vault: hasGroupSupport
	GroupSupport() bool

	// PlayerHas reports whether player has node. Vault: playerHas
	PlayerHas(world plugin.World, player plugin.UUID, node string) bool
	// PlayerAdd and PlayerRemove grant or unset node for player, saved.
	// Vault: playerAdd, playerRemove
	PlayerAdd(world plugin.World, player plugin.UUID, node string) error
	PlayerRemove(world plugin.World, player plugin.UUID, node string) error
	// PlayerAddTransient and PlayerRemoveTransient grant or unset node for
	// player until they quit or the server stops, without saving it.
	// Vault: playerAddTransient, playerRemoveTransient
	PlayerAddTransient(world plugin.World, player plugin.UUID, node string) error
	PlayerRemoveTransient(world plugin.World, player plugin.UUID, node string) error

	// GroupHas reports whether group has node. Vault: groupHas
	GroupHas(world plugin.World, group, node string) bool
	// GroupAdd and GroupRemove grant or unset node for group.
	// Vault: groupAdd, groupRemove
	GroupAdd(world plugin.World, group, node string) error
	GroupRemove(world plugin.World, group, node string) error

	// PlayerInGroup reports whether player is a direct member of group, not
	// through another group. Vault: playerInGroup
	PlayerInGroup(world plugin.World, player plugin.UUID, group string) bool
	// PlayerAddGroup and PlayerRemoveGroup add player to group or take them
	// out of it. Vault: playerAddGroup, playerRemoveGroup
	PlayerAddGroup(world plugin.World, player plugin.UUID, group string) error
	PlayerRemoveGroup(world plugin.World, player plugin.UUID, group string) error
	// PlayerGroups answers the groups player is a direct member of.
	// Vault: getPlayerGroups
	PlayerGroups(world plugin.World, player plugin.UUID) []string
	// PrimaryGroup answers player's main group, and "" for none.
	// Vault: getPrimaryGroup
	PrimaryGroup(world plugin.World, player plugin.UUID) string
	// Groups answers every group. Vault: getGroups
	Groups() []string
}
