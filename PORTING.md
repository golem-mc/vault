# Porting Vault to golem: layout and conventions

This module is Vault and VaultAPI (Java, Bukkit) ported to golem (golem-mc/golem; Go).
`go.mod` requires a golem commit; to move to a newer one, run
`mise exec -- go get github.com/golem-mc/golem@<commit>`. To build against a local golem
checkout instead, add a `go.work` (ignored by git): `go work init . ../golem`. The Java
sources are checked out, ignored by git, in `Vault/` and `VaultAPI/` for reference.

## Layout

| Path | What | Java counterpart |
|---|---|---|
| `economy.go` | `Economy`, `Response`, the optional `Banks` | `economy/Economy`, `EconomyResponse` |
| `permission.go` | `Permission` | `permission/Permission` |
| `chat.go` | `Chat`, `PlayerInfoOr` / `GroupInfoOr` | `chat/Chat` |
| `services.go` | `Provide*`, `Lookup*`, the provider lists, `Priority*` | Bukkit's `ServicesManager`, `ServicePriority` |
| `vaultplugin/` | The `vault` plugin: commands, the SuperPerms fallback | `Vault.java`, `Permission_SuperPerms` |

## How the API maps

- **Players are UUIDs.** Every `OfflinePlayer` and name overload became one method taking
  a `plugin.UUID`. `CreateAccount` also takes a name, which Essentials needs for NPC
  accounts.
- **Worlds are `plugin.World`.** nil stands for Vault's null world: the player's current
  context for a query, every world for a change. Economy has no world variants.
- **Errors, not booleans.** A change answers an `error`, `errors.ErrUnsupported` where
  Vault returns false or throws `UnsupportedOperationException`. `EconomyResponse`'s type
  and message became `Response.Err`.
- **One string for info nodes.** Chat's typed getters and setters became `PlayerInfo` /
  `SetPlayerInfo` (strings) and the generic `PlayerInfoOr[T]`.
- **Dropped:** `isEnabled` (a disabled plugin's services are removed), `has(CommandSender)`
  (that is `CommandSender.HasPermission`), and Chat's group-membership shortcuts (use
  `Permission`).

Doc comments name the Vault counterpart as `Vault: methodName`.

## Tests

Tests start a real golem server in-process with golem's `plugintest` and plugins built for
the test: a provider and a consumer that only knows package `vault`.

## Differences from Vault

- **No hooks for other plugins.** Vault ships adapters for dozens of economy and
  permission plugins that predate it. Here each plugin provides the service itself, as
  Essentials and LuckPerms do; there is nothing to adapt.
- **Transient permissions in the fallback.** golem has no permission attachments, so
  SuperPerms answers `errors.ErrUnsupported` to `PlayerAddTransient`.
- **No update checker or metrics.**
- **Packages.** The API is package `github.com/golem-mc/vault` (`Economy`, `Permission`,
  `Chat`); importing it registers nothing, so any plugin can depend on it. The plugin is
  `github.com/golem-mc/vault/vaultplugin`, registered as `vault`.
- **Providers.** Essentials provides `vault.Economy`; LuckPerms provides
  `vault.Permission` and `vault.Chat`; the `vault` plugin's "SuperPerms" is the
  permission fallback. `/vault-info` and `/vault-convert` print Vault's messages and
  need `vault.admin`.
- **Banks** are an optional interface (`vault.Banks`, found with `vault.BanksOf`); info
  nodes are read as ints, floats and bools with `vault.PlayerInfoOr` / `GroupInfoOr`.

## For plugin authors

Add the module (`go get github.com/golem-mc/vault`) and import the package `vault`. Look a
service up when you need it, not once in `Enable`: providers come and go with their
plugins.

### Using a service

```go
import (
	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/vault"
)

func chargeEntry(s plugin.Server, p plugin.Player, fee float64) bool {
	eco, ok := vault.LookupEconomy(s)
	if !ok {
		return true // no economy plugin: entry is free
	}
	if !eco.Has(p.UUID(), fee) {
		p.SendMessage(plugin.Plain("You need " + eco.Format(fee)).Color("red"))
		return false
	}
	if r := eco.Withdraw(p.UUID(), fee); !r.Success() {
		p.SendMessage(plugin.Plain(r.Err.Error()).Color("red"))
		return false
	}
	return true
}
```

`vault.LookupPermission` and `vault.LookupChat` work the same way:

```go
if chat, ok := vault.LookupChat(s); ok {
	prefix := chat.PlayerPrefix(p.World(), p.UUID())
	homes := vault.PlayerInfoOr(chat, p.World(), p.UUID(), "homes", 3)
}
```

A `world` argument names the world a query or change is about; `nil` means none in
particular. Players are always named by UUID; Vault's deprecated name-based methods are
gone.

### Providing a service

Implement the interface and provide it in `Enable`:

```go
func (m *MyEconomy) Enable(s plugin.Server) error {
	vault.ProvideEconomy(s, m, vault.PriorityNormal)
	return nil
}
```

Use `vault.ProvideEconomy` / `ProvidePermission` / `ProvideChat` (or
`plugin.Provide[vault.Economy]` with the type argument): golem finds services by their
exact interface type, and `plugin.Provide(s, m, 0)` without one registers `m` under its
own type, where nobody looks. The highest priority wins; the `vault.Priority*` constants
are Bukkit's `ServicePriority`. Answer `errors.ErrUnsupported` for what your plugin cannot
do. Calls come from plugins' commands and listeners, on the tick goroutine, unless the
caller documents otherwise.

## Development

```bash
mise run build
mise run test:changed
mise run dev        # a golem server with Vault built in
```

See [CONTRIBUTING.md](CONTRIBUTING.md).
