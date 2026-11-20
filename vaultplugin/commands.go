package vaultplugin

import (
	"fmt"
	"strings"

	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/vault"
)

func registerCommands(s plugin.Server) error {
	info := func(ctx *plugin.CommandContext) (int, error) {
		infoCommand(s, ctx.Sender)
		return 1, nil
	}
	return s.Commands().Register(plugin.Command{Name: "vault-info",
		Description: "Displays information about Vault",
		Usage:       "/vault-info - Displays Vault information",
		Permission:  AdminPermission,
		Run:         info, Children: []plugin.Node{{Arg: plugin.GreedyArg("args"), Run: info}},
	})
}

// infoCommand lists the service in use and every provider of each.
// Vault: Vault.infoCommand
func infoCommand(s plugin.Server, to plugin.CommandSender) {
	version := ""
	for _, p := range s.Plugins() {
		if p.Name == s.Name() {
			version = p.Version
		}
	}
	say(to, "[Vault] Vault v%s Information", version)
	say(to, "[Vault] Economy: %s", serviceLine(vault.Economies(s)))
	say(to, "[Vault] Permission: %s", serviceLine(vault.Permissions(s)))
	say(to, "[Vault] Chat: %s", serviceLine(vault.Chats(s)))
}

// serviceLine is "<in use> [<all>]": "None" for none in use and, as
// Vault's string concatenation of a null list prints it, "null" for an
// empty list.
func serviceLine[S interface{ Name() string }](all []S) string {
	if len(all) == 0 {
		return "None [null]"
	}
	names := make([]string, len(all))
	for i, p := range all {
		names[i] = p.Name()
	}
	return fmt.Sprintf("%s [%s]", names[0], strings.Join(names, ", "))
}

func say(to plugin.CommandSender, format string, args ...any) {
	to.SendMessage(plugin.Plain(fmt.Sprintf(format, args...)))
}
