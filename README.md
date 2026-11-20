# Vault for golem

An unofficial port of [Vault](https://github.com/MilkBowl/Vault) to
[golem](https://github.com/golem-mc/golem). Vault is the shared economy, permission and
chat API: plugins that charge money, check permissions or show chat prefixes work with
whichever economy or permission plugin the server runs.

## Install

```bash
golem-launcher plugin add github.com/golem-mc/vault/vaultplugin
```

This is the only module to install. It is optional: plugins find the economy, permission
and chat services of the plugins that provide them without it. Install it for the
commands below and for a permission fallback when no permission plugin is installed.

Vault itself has no economy or permissions; add plugins that provide them:

```bash
golem-launcher plugin add github.com/golem-mc/essentials   # economy
golem-launcher plugin add github.com/golem-mc/perms        # permissions and chat
```

## Getting started

Install Vault together with an economy and a permission plugin, start the server and run
`/vault-info` as an operator:

```
/vault-info
[Vault] Vault v1.0.0 Information
[Vault] Economy: Essentials Economy [Essentials Economy]
[Vault] Permission: LuckPerms [LuckPerms, SuperPerms]
[Vault] Chat: LuckPerms [LuckPerms]
```

The first name on each line is the provider plugins get; the list is every installed
provider. `None [null]` means no plugin provides that service.

## Commands

Both commands need `vault.admin`.

| Command | What it does | Permission |
|---|---|---|
| `/vault-info` | Shows the Vault version and the economy, permission and chat provider in use, with every installed provider | `vault.admin` |
| `/vault-convert <economy1> <economy2>` | Copies the player balances of `economy1` into `economy2` for players that have no account there yet. Needs two economy plugins; names are written without spaces and tab-complete | `vault.admin` |

Neither command has aliases.

## Permissions

| Permission | Default | Allows |
|---|---|---|
| `vault.admin` | operators | `/vault-info` and `/vault-convert` |

With no permission plugin installed, Vault's "SuperPerms" answers permission checks from
what golem itself grants a player (`vault.admin` for operators), so plugins that only
check permissions keep working. It cannot add or remove permissions, groups or prefixes;
a plugin that tries is told the operation is unsupported. Install
[LuckPerms](https://github.com/golem-mc/perms) for that.

## Configuration

Vault has no configuration file.

## Migrating from Bukkit or Paper

Remove Vault's jar and install the plugins above. Plugins written for golem find the
services through Vault's API, as Bukkit plugins do through `ServicesManager`. Bukkit
plugins' jars do not run on golem, so each plugin that used Vault needs its golem port.

## Works with

- [Essentials](https://github.com/golem-mc/essentials) provides the economy.
- [LuckPerms](https://github.com/golem-mc/perms) provides permissions, groups, prefixes
  and suffixes.
- Any plugin that provides or uses an economy, permission or chat service through the API
  below.

## Differences from Vault

- Economy, permission and chat plugins provide their service directly; there are no
  adapters for other plugins.
- Players are named by UUID, not by name.
- Vault's update checker and metrics are not included.

## For plugin developers

Add the module and import the package `vault`:

```bash
go get github.com/golem-mc/vault
```

Importing it registers nothing, so any plugin can depend on it. Look a service up when you
need it, not once in `Enable`: providers come and go with their plugins. Players are named
by UUID, and a `world` argument names the world a query or change is about (`nil` means
none in particular).

### Using a service

`vault.LookupEconomy`, `LookupPermission` and `LookupChat` answer the highest-priority
provider, and false when there is none:

```go
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

```go
func canBuild(s plugin.Server, p plugin.Player) bool {
	perm, ok := vault.LookupPermission(s)
	if !ok {
		return false
	}
	return perm.PlayerHas(p.World(), p.UUID(), "myplugin.build")
}
```

Chat info nodes are read as a string, int, float or bool with `vault.PlayerInfoOr` and
`GroupInfoOr`, which answer the default when the node is unset or does not parse:

```go
func chatLine(s plugin.Server, p plugin.Player) string {
	chat, ok := vault.LookupChat(s)
	if !ok {
		return p.Name()
	}
	homes := vault.PlayerInfoOr(chat, p.World(), p.UUID(), "homes", 3)
	_ = homes
	return chat.PlayerPrefix(p.World(), p.UUID()) + p.Name() + chat.PlayerSuffix(p.World(), p.UUID())
}
```

Banks are optional for an economy; `vault.BanksOf` finds them:

```go
func bankBalance(s plugin.Server, bank string) (float64, bool) {
	eco, ok := vault.LookupEconomy(s)
	if !ok {
		return 0, false
	}
	banks, ok := vault.BanksOf(eco)
	if !ok {
		return 0, false // this economy has no banks
	}
	balance, err := banks.BankBalance(bank)
	return balance, err == nil
}
```

### Providing a service

Implement `vault.Economy`, `vault.Permission` or `vault.Chat` and provide it in `Enable`
with the matching `Provide` function, which also takes a priority
(`vault.PriorityLowest` to `PriorityHighest`, as Bukkit's `ServicePriority`). Here
`MyEconomy` implements `vault.Economy`:

```go
func (m *MyEconomy) Enable(s plugin.Server) error {
	vault.ProvideEconomy(s, m, vault.PriorityNormal)
	return nil
}
```

```go
func (m *MyPermissions) Enable(s plugin.Server) error {
	vault.ProvidePermission(s, m, vault.PriorityHigh)
	return nil
}
```

Use `vault.ProvideEconomy`, `ProvidePermission` and `ProvideChat` rather than
`plugin.Provide` without a type argument, which registers the value where nobody looks.
Answer `errors.ErrUnsupported` for what your plugin cannot do. An economy that keeps
banks also implements `vault.Banks` on the same value.

## License

LGPL-3.0, like Vault. See `LICENSE` and `COPYING`. This project is not affiliated with or
endorsed by the Vault authors.
