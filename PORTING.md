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
