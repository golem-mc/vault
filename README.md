# Vault for golem

An unofficial port of [Vault](https://github.com/MilkBowl/Vault) to
[golem](https://github.com/golem-mc/golem). Vault is the shared economy, permission and
chat API: plugins that charge money or check permissions work with whichever economy or
permission plugin the server runs.

```bash
golem-launcher plugin add github.com/golem-mc/vault/vaultplugin
```

It is optional: plugins find the services of an installed economy or permission plugin
without it. Install it for the commands below.

## What's included

Commands work as in [Vault](https://dev.bukkit.org/projects/vault) and need `vault.admin`
(operators by default).

- `/vault-info`: the economy, permission and chat service in use and their providers
- `/vault-convert`: copy player balances from one economy plugin into another
- Fallback permissions: with no permission plugin, plugins that only check permissions
  keep working, as with Vault's SuperPerms

Economy comes from [Essentials](https://github.com/golem-mc/essentials), permissions and
chat from [LuckPerms](https://github.com/golem-mc/perms).

## Not yet

- Transient permissions in the fallback: waits on [golem-mc/golem#154](https://github.com/golem-mc/golem/issues/154)

## Configuration

None.

## License

LGPL-3.0, like Vault. See `LICENSE` and `COPYING`. This project is not affiliated with or
endorsed by the Vault authors.
