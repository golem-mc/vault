// Package vault is the service API of Vault, ported to golem: the interfaces an economy,
// permission or chat plugin provides and other plugins use. Importing it registers nothing.
//
// A plugin that keeps balances provides an Economy; one that charges looks it up when it
// charges:
//
//	vault.ProvideEconomy(s, myEconomy, vault.PriorityNormal)
//
//	if eco, ok := vault.LookupEconomy(s); ok {
//		if r := eco.Withdraw(p.UUID(), 50); !r.Success() {
//			p.SendMessage(plugin.Plain(r.Err.Error()).Color("red"))
//		}
//	}
//
// Permission and Chat work the same way, through ProvidePermission, LookupPermission,
// ProvideChat and LookupChat. The Vault plugin, package vaultplugin, adds /vault-info,
// /vault-convert and a fallback Permission over golem's own permission checks.
package vault
