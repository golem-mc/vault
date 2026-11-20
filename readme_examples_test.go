package vault_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/vault"
)

// The functions below are the README's "For plugin developers" examples,
// verbatim; TestReadmeExamplesCompile fails when the README and this file
// differ, so the examples are always the ones the compiler has seen.

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

func canBuild(s plugin.Server, p plugin.Player) bool {
	perm, ok := vault.LookupPermission(s)
	if !ok {
		return false
	}
	return perm.PlayerHas(p.World(), p.UUID(), "myplugin.build")
}

func chatLine(s plugin.Server, p plugin.Player) string {
	chat, ok := vault.LookupChat(s)
	if !ok {
		return p.Name()
	}
	homes := vault.PlayerInfoOr(chat, p.World(), p.UUID(), "homes", 3)
	_ = homes
	return chat.PlayerPrefix(p.World(), p.UUID()) + p.Name() + chat.PlayerSuffix(p.World(), p.UUID())
}

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

// MyEconomy stands for a type that implements vault.Economy.
type MyEconomy struct{ vault.Economy }

func (m *MyEconomy) Enable(s plugin.Server) error {
	vault.ProvideEconomy(s, m, vault.PriorityNormal)
	return nil
}

func (m *MyEconomy) Disable() {}

type MyPermissions struct{ vault.Permission }

func (m *MyPermissions) Enable(s plugin.Server) error {
	vault.ProvidePermission(s, m, vault.PriorityHigh)
	return nil
}

func (m *MyPermissions) Disable() {}

var (
	_ plugin.Plugin = (*MyEconomy)(nil)
	_ plugin.Plugin = (*MyPermissions)(nil)
)

func TestReadmeExamplesCompile(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("readme_examples_test.go")
	if err != nil {
		t.Fatal(err)
	}
	blocks := regexp.MustCompile("(?s)```go\n(.*?)```").FindAllStringSubmatch(string(readme), -1)
	if len(blocks) < 5 {
		t.Fatalf("README has %d Go examples, want at least 5", len(blocks))
	}
	for _, b := range blocks {
		if code := strings.TrimSpace(b[1]); !strings.Contains(string(src), code) {
			t.Errorf("README example is not in readme_examples_test.go:\n%s", code)
		}
	}
}
