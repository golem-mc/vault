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
	if err := s.Commands().Register(plugin.Command{Name: "vault-info",
		Description: "Displays information about Vault",
		Usage:       "/vault-info - Displays Vault information",
		Permission:  AdminPermission,
		Run:         info, Children: []plugin.Node{{Arg: plugin.GreedyArg("args"), Run: info}},
	}); err != nil {
		return err
	}
	convert := func(ctx *plugin.CommandContext) (int, error) {
		convertCommand(s, ctx.Sender, strings.Fields(plugin.Arg[string](ctx, "args")))
		return 1, nil
	}
	return s.Commands().Register(plugin.Command{Name: "vault-convert",
		Description: "Converts all data in economy1 and dumps it into economy2",
		Usage:       "/vault-convert [economy1] [economy2]",
		Permission:  AdminPermission,
		Run:         convert, Children: []plugin.Node{{Arg: plugin.GreedyArg("args"), Run: convert,
			Suggest: func(*plugin.CommandContext, string) []string { return economyNames(vault.Economies(s)) }}},
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

// convertCommand copies every player's balance from one economy to another
// that has no account for them yet. Vault: Vault.convertCommand
func convertCommand(s plugin.Server, to plugin.CommandSender, args []string) {
	econs := vault.Economies(s)
	if len(econs) < 2 {
		say(to, "You must have at least 2 economies loaded to convert.")
		return
	}
	if len(args) != 2 {
		say(to, "You must specify only the economy to convert from and the economy to convert to. (names should not contain spaces)")
		return
	}
	var from, into vault.Economy
	for _, e := range econs {
		name := strings.ReplaceAll(e.Name(), " ", "")
		if strings.EqualFold(name, args[0]) {
			from = e
		} else if strings.EqualFold(name, args[1]) {
			into = e
		}
	}
	valid := strings.Join(economyNames(econs), ", ")
	if from == nil {
		say(to, "Could not find %s loaded on the server, check your spelling.", args[0])
		say(to, "Valid economies are: %s", valid)
		return
	}
	if into == nil {
		say(to, "Could not find %s loaded on the server, check your spelling.", args[1])
		say(to, "Valid economies are: %s", valid)
		return
	}

	say(to, "This may take some time to convert, expect server lag.")
	for _, p := range s.OfflinePlayers() {
		id := p.UUID()
		if !from.HasAccount(id) || into.HasAccount(id) {
			continue
		}
		into.CreateAccount(id, p.Name())
		switch diff := from.Balance(id) - into.Balance(id); {
		case diff > 0:
			into.Deposit(id, diff)
		case diff < 0:
			into.Withdraw(id, -diff)
		}
	}
	// vault: the typo is Vault's.
	say(to, "Converson complete, please verify the data before using it.")
}

// economyNames answers the names /vault-convert matches: spaces removed.
func economyNames(econs []vault.Economy) []string {
	out := make([]string, len(econs))
	for i, e := range econs {
		out[i] = strings.ReplaceAll(e.Name(), " ", "")
	}
	return out
}

func say(to plugin.CommandSender, format string, args ...any) {
	to.SendMessage(plugin.Plain(fmt.Sprintf(format, args...)))
}
