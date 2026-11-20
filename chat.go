package vault

import (
	"strconv"

	"github.com/golem-mc/golem/plugin"
)

// Chat is the chat prefixes, suffixes and info nodes (LuckPerms' meta) a
// permission or chat plugin keeps for players and groups, for chat
// formatters and other plugins to read and set. world is as for Permission.
// Setting an empty value removes the player's or group's own. Vault: Chat
type Chat interface {
	// Name answers the chat plugin's name, as /vault-info shows it.
	// Vault: getName
	Name() string

	// PlayerPrefix and PlayerSuffix answer player's prefix and suffix, their
	// own or inherited from a group, and "" for none.
	// Vault: getPlayerPrefix, getPlayerSuffix
	PlayerPrefix(world plugin.World, player plugin.UUID) string
	PlayerSuffix(world plugin.World, player plugin.UUID) string
	// SetPlayerPrefix and SetPlayerSuffix set player's own prefix and suffix.
	// Vault: setPlayerPrefix, setPlayerSuffix
	SetPlayerPrefix(world plugin.World, player plugin.UUID, prefix string) error
	SetPlayerSuffix(world plugin.World, player plugin.UUID, suffix string) error

	// GroupPrefix and GroupSuffix answer group's prefix and suffix, and ""
	// for none. Vault: getGroupPrefix, getGroupSuffix
	GroupPrefix(world plugin.World, group string) string
	GroupSuffix(world plugin.World, group string) string
	// SetGroupPrefix and SetGroupSuffix set group's own prefix and suffix.
	// Vault: setGroupPrefix, setGroupSuffix
	SetGroupPrefix(world plugin.World, group, prefix string) error
	SetGroupSuffix(world plugin.World, group, suffix string) error

	// PlayerInfo answers player's info node, their own or inherited, and
	// whether it is set; PlayerInfoOr reads it as a number or bool.
	// Vault: getPlayerInfoString
	PlayerInfo(world plugin.World, player plugin.UUID, node string) (string, bool)
	// SetPlayerInfo sets player's own info node; a number or bool is
	// written as strconv formats it. Vault: setPlayerInfoString and the
	// other typed setters
	SetPlayerInfo(world plugin.World, player plugin.UUID, node, value string) error
	// GroupInfo and SetGroupInfo are PlayerInfo and SetPlayerInfo for a
	// group. Vault: getGroupInfoString, setGroupInfoString
	GroupInfo(world plugin.World, group, node string) (string, bool)
	SetGroupInfo(world plugin.World, group, node, value string) error
}

// InfoValue is a type an info node can be read as.
type InfoValue interface {
	string | int | float64 | bool
}

// PlayerInfoOr answers player's info node read as a T, or def when it is
// unset or does not parse as one. Vault: getPlayerInfoInteger,
// getPlayerInfoDouble, getPlayerInfoBoolean, getPlayerInfoString
func PlayerInfoOr[T InfoValue](c Chat, world plugin.World, player plugin.UUID, node string, def T) T {
	v, ok := c.PlayerInfo(world, player, node)
	return parseInfo(v, ok, def)
}

// GroupInfoOr answers group's info node read as a T, or def when it is
// unset or does not parse as one. Vault: getGroupInfoInteger,
// getGroupInfoDouble, getGroupInfoBoolean, getGroupInfoString
func GroupInfoOr[T InfoValue](c Chat, world plugin.World, group, node string, def T) T {
	v, ok := c.GroupInfo(world, group, node)
	return parseInfo(v, ok, def)
}

func parseInfo[T InfoValue](v string, ok bool, def T) T {
	if !ok {
		return def
	}
	var out any
	var err error
	switch any(def).(type) {
	case string:
		out = v
	case int:
		out, err = strconv.Atoi(v)
	case float64:
		out, err = strconv.ParseFloat(v, 64)
	case bool:
		out, err = strconv.ParseBool(v)
	}
	if err != nil {
		return def
	}
	return out.(T)
}
