// Package vaultplugin is the Vault plugin, registered as "vault": /vault-info
// and the fallback Permission service over golem's own
// permission checks. Vault: net.milkbowl.vault.Vault
package vaultplugin

import (
	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/vault"
)

// AdminPermission lets a sender run Vault's commands.
const AdminPermission = "vault.admin"

func init() { plugin.Register("vault", &Vault{}) }

// Vault is the plugin.
type Vault struct{}

// Enable registers the commands and provides the fallback Permission.
func (*Vault) Enable(s plugin.Server) error {
	// plugin.yml: default op, any operator.
	s.Permissions().Register(plugin.Permission{Name: AdminPermission,
		Description: "Notifies the player when vault is in need of an update.",
		Default:     plugin.PermissionDefault{OpLevel: 1}})
	// vault: Vault.loadPermission registers SuperPerms at ServicePriority.Lowest.
	vault.ProvidePermission(s, superPerms{s}, vault.PriorityLowest)
	return registerCommands(s)
}

// Disable does nothing: golem removes the commands and services.
func (*Vault) Disable() {}
