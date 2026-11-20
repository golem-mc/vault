package vault

import (
	"reflect"

	"github.com/golem-mc/golem/plugin"
)

// Service priorities, Bukkit's ServicePriority on golem's scale, where the
// highest-priority provider of a service is the one plugins get. The Vault
// plugin's own fallbacks are PriorityLowest, so any plugin's provider wins.
const (
	PriorityLowest  = -2
	PriorityLow     = -1
	PriorityNormal  = 0
	PriorityHigh    = 1
	PriorityHighest = 2
)

// ProvideEconomy makes e the plugin's Economy service. Prefer it to
// plugin.Provide, which without its type argument registers e under its
// concrete type, where no plugin looks.
func ProvideEconomy(s plugin.Server, e Economy, priority int) {
	plugin.Provide[Economy](s, e, priority)
}

// ProvidePermission makes p the plugin's Permission service.
func ProvidePermission(s plugin.Server, p Permission, priority int) {
	plugin.Provide[Permission](s, p, priority)
}

// ProvideChat makes c the plugin's Chat service.
func ProvideChat(s plugin.Server, c Chat, priority int) {
	plugin.Provide[Chat](s, c, priority)
}

// LookupEconomy answers the highest-priority Economy any plugin provides.
// Look it up when it is needed rather than once in Enable: the economy
// plugin may be enabled later, or disabled.
func LookupEconomy(s plugin.Server) (Economy, bool) { return plugin.Service[Economy](s) }

// LookupPermission answers the highest-priority Permission any plugin
// provides; with the Vault plugin installed there is always one.
func LookupPermission(s plugin.Server) (Permission, bool) { return plugin.Service[Permission](s) }

// LookupChat answers the highest-priority Chat any plugin provides.
func LookupChat(s plugin.Server) (Chat, bool) { return plugin.Service[Chat](s) }

// Economies answers every Economy the plugins provide, highest priority
// first. Vault: ServicesManager.getRegistrations(Economy.class)
func Economies(s plugin.Server) []Economy { return providers[Economy](s) }

// Permissions answers every Permission the plugins provide, highest
// priority first.
func Permissions(s plugin.Server) []Permission { return providers[Permission](s) }

// Chats answers every Chat the plugins provide, highest priority first.
func Chats(s plugin.Server) []Chat { return providers[Chat](s) }

func providers[S any](s plugin.Server) []S {
	ps := s.Services().Providers(reflect.TypeFor[S]())
	out := make([]S, len(ps))
	for i, p := range ps {
		out[i] = p.Impl.(S)
	}
	return out
}
