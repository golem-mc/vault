# Vault for golem

An unofficial port of [Vault](https://github.com/MilkBowl/Vault) and its API
([VaultAPI](https://github.com/MilkBowl/VaultAPI)) to
[golem](https://github.com/golem-mc/golem), the Minecraft: Java Edition server written in Go.
Vault is the shared economy, permission and chat API: an economy, permission or chat plugin
provides it, and other plugins use it, so a plugin that charges money works with whichever
economy plugin a server runs.

It has two parts:

- **The API**, package `github.com/golem-mc/vault`: the `Economy`, `Permission` and `Chat`
  interfaces. Importing it registers nothing, so any plugin can depend on it.
- **The Vault plugin**, package `github.com/golem-mc/vault/vaultplugin`, registered as
  `vault`: `/vault-info`, `/vault-convert` and a fallback permission service.

## What's included

- **Economy** (`vault.Economy`): balances, `Has`, `Withdraw` / `Deposit` with a
  `vault.Response` (amount, new balance, error), account creation (players and NPC
  accounts such as towns), currency names, `Format` and fractional digits. Banks are an
  optional interface (`vault.Banks`, found with `vault.BanksOf`).
- **Permission** (`vault.Permission`): player and group permission queries and changes,
  transient permissions, group membership, primary group and the group list, per world
  or everywhere, for online and offline players.
- **Chat** (`vault.Chat`): player and group prefixes, suffixes and info nodes (meta), read
  as strings or, with `vault.PlayerInfoOr` / `vault.GroupInfoOr`, as ints, floats and bools.
- **`/vault-info`** shows the economy, permission and chat service in use and every
  provider of each; **`/vault-convert <economy1> <economy2>`** copies every player's
  balance from one economy plugin into another that has no account for them yet. Both
  need `vault.admin` (operators by default) and print Vault's messages.
- **Fallback permissions**: with no permission plugin, the Vault plugin provides a
  `vault.Permission` ("SuperPerms") over golem's own permission checks, as Vault falls
  back to Bukkit's SuperPerms, so plugins that only check permissions keep working.

### Which plugins provide what

| Service | Provided by |
|---|---|
| `vault.Economy` | [Essentials](https://github.com/golem-mc/essentials) ("EssentialsX Economy") |
| `vault.Permission` | [LuckPerms](https://github.com/golem-mc/perms); the Vault plugin's "SuperPerms" fallback otherwise |
| `vault.Chat` | [LuckPerms](https://github.com/golem-mc/perms) |

## Not yet

- **Vault's own hooks for other plugins.** On Bukkit, Vault ships adapters for dozens of
  economy and permission plugins that predate it. Here each plugin provides the service
  itself, as Essentials and LuckPerms do; there is nothing to adapt.
- **Transient permissions in the fallback.** golem has no permission attachments, so
  SuperPerms answers `errors.ErrUnsupported` to `PlayerAddTransient`.
- **The update checker and metrics.**

## Installing

### For server owners

Add the plugin with
[`golem-launcher`](https://github.com/golem-mc/golem#installing-plugins):

```bash
./golem-launcher plugin add github.com/golem-mc/vault
./golem-launcher
```

The Vault plugin is optional: a plugin that uses Vault's services finds them as soon as
an economy or permission plugin that provides them is installed. Install it for
`/vault-info`, `/vault-convert` and the fallback permission service.

### As a ready-built server

```bash
git clone https://github.com/golem-mc/vault.git
cd vault
mise run dev        # or: go run ./cmd/server
```

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

See [CONTRIBUTING.md](CONTRIBUTING.md) and [PORTING.md](PORTING.md).

## License

LGPL-3.0, like Vault. See `LICENSE` and `COPYING`. This project is not affiliated with or
endorsed by the Vault authors.
