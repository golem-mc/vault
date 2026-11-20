package vault_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/golem/plugintest"
	"github.com/golem-mc/vault"
)

func TestImportingRegistersNothing(t *testing.T) {
	if regs := plugin.Registered(); len(regs) != 0 {
		t.Fatalf("importing vault registered %v, want nothing", regs)
	}
}

// TestConsumerFindsEachService runs a provider plugin and a shop plugin that
// only knows the vault interfaces: /buy checks a permission, charges and
// shows the buyer's prefix.
func TestConsumerFindsEachService(t *testing.T) {
	srv := plugintest.Start(t, plugintest.WithPlugins(
		plugin.Registration{Name: "provider", Plugin: &provider{}},
		plugin.Registration{Name: "shop", Plugin: shop{}},
	))
	alice := srv.Join("Alice")

	alice.Run("/buy")
	if got := strings.Join(alice.Messages(2), "\n"); got != "You may not buy" {
		t.Fatalf("/buy without the node: %q", got)
	}

	host := srv.Host()
	id := alice.Handle().UUID()
	perm, _ := vault.LookupPermission(host)
	if err := perm.PlayerAdd(nil, id, "shop.buy"); err != nil {
		t.Fatal(err)
	}
	chat, _ := vault.LookupChat(host)
	if err := chat.SetPlayerPrefix(nil, id, "[VIP] "); err != nil {
		t.Fatal(err)
	}
	eco, _ := vault.LookupEconomy(host)
	eco.CreateAccount(id, "Alice")
	eco.Deposit(id, 15)

	alice.Run("/buy")
	if got := strings.Join(alice.Messages(2), "\n"); got != "[VIP] Alice bought a pass for $10.00" {
		t.Fatalf("/buy: %q", got)
	}
	alice.Run("/buy")
	if got := strings.Join(alice.Messages(2), "\n"); got != "Not enough money" {
		t.Fatalf("/buy with $5 left: %q", got)
	}
	if got := eco.Balance(id); got != 5 {
		t.Fatalf("balance %v, want 5", got)
	}
	if _, ok := vault.BanksOf(eco); ok {
		t.Fatal("an economy without bank methods has banks")
	}
}

func TestHigherPriorityProviderWins(t *testing.T) {
	srv := plugintest.Start(t, plugintest.WithPlugins(
		plugin.Registration{Name: "low", Plugin: economyAt{"Low", vault.PriorityLowest}},
		plugin.Registration{Name: "high", Plugin: economyAt{"High", vault.PriorityHigh}},
	))
	if eco, ok := vault.LookupEconomy(srv.Host()); !ok || eco.Name() != "High" {
		t.Fatalf("LookupEconomy answered %v, want the PriorityHigh one", eco)
	}
	if _, ok := vault.LookupChat(srv.Host()); ok {
		t.Fatal("LookupChat found a chat nobody provides")
	}
}

func TestInfoOrParsesTheNode(t *testing.T) {
	c := newMemory()
	id := plugin.UUID{1}
	for node, v := range map[string]string{"homes": "3", "rate": "1.5", "vip": "true", "bad": "x"} {
		if err := c.SetPlayerInfo(nil, id, node, v); err != nil {
			t.Fatal(err)
		}
	}
	if got := vault.PlayerInfoOr(c, nil, id, "homes", 1); got != 3 {
		t.Errorf("homes = %d", got)
	}
	if got := vault.PlayerInfoOr(c, nil, id, "rate", 0.0); got != 1.5 {
		t.Errorf("rate = %v", got)
	}
	if got := vault.PlayerInfoOr(c, nil, id, "vip", false); !got {
		t.Error("vip = false")
	}
	if got := vault.PlayerInfoOr(c, nil, id, "bad", 7); got != 7 {
		t.Errorf("an unparsable node read %d, want the default", got)
	}
	if got := vault.PlayerInfoOr(c, nil, id, "unset", "none"); got != "none" {
		t.Errorf("an unset node read %q, want the default", got)
	}
	if got := vault.GroupInfoOr(c, nil, "admin", "homes", 2); got != 2 {
		t.Errorf("group homes = %d", got)
	}
}

// shop is a consumer: it knows only package vault.
type shop struct{}

func (shop) Enable(s plugin.Server) error {
	return s.Commands().Register(plugin.Command{Name: "buy", Run: func(ctx *plugin.CommandContext) (int, error) {
		p := ctx.Sender.(plugin.Player)
		perm, ok1 := vault.LookupPermission(s)
		eco, ok2 := vault.LookupEconomy(s)
		chat, ok3 := vault.LookupChat(s)
		switch {
		case !ok1 || !ok2 || !ok3:
			p.SendMessage(plugin.Plain("No services"))
		case !perm.PlayerHas(p.World(), p.UUID(), "shop.buy"):
			p.SendMessage(plugin.Plain("You may not buy"))
		case !eco.Has(p.UUID(), 10):
			p.SendMessage(plugin.Plain("Not enough money"))
		default:
			if r := eco.Withdraw(p.UUID(), 10); !r.Success() {
				p.SendMessage(plugin.Plain(r.Err.Error()))
				break
			}
			p.SendMessage(plugin.Plain(chat.PlayerPrefix(p.World(), p.UUID()) + p.Name() + " bought a pass for " + eco.Format(10)))
		}
		return 1, nil
	}})
}

func (shop) Disable() {}

// provider provides all three services from one in-memory store.
type provider struct{}

func (*provider) Enable(s plugin.Server) error {
	m := newMemory()
	vault.ProvideEconomy(s, m, vault.PriorityNormal)
	vault.ProvidePermission(s, m, vault.PriorityNormal)
	vault.ProvideChat(s, m, vault.PriorityNormal)
	return nil
}

func (*provider) Disable() {}

type economyAt struct {
	name     string
	priority int
}

func (e economyAt) Enable(s plugin.Server) error {
	m := newMemory()
	m.name = e.name
	vault.ProvideEconomy(s, m, e.priority)
	return nil
}

func (economyAt) Disable() {}

// memory is an Economy, Permission and Chat without groups.
type memory struct {
	name     string
	mu       sync.Mutex
	balances map[plugin.UUID]float64
	nodes    map[plugin.UUID]map[string]bool
	info     map[plugin.UUID]map[string]string
}

func newMemory() *memory {
	return &memory{name: "Memory", balances: map[plugin.UUID]float64{}, nodes: map[plugin.UUID]map[string]bool{}, info: map[plugin.UUID]map[string]string{}}
}

func (m *memory) Name() string                 { return m.name }
func (m *memory) FractionalDigits() int        { return 2 }
func (m *memory) Format(amount float64) string { return fmt.Sprintf("$%.2f", amount) }
func (m *memory) CurrencyNameSingular() string { return "dollar" }
func (m *memory) CurrencyNamePlural() string   { return "dollars" }

func (m *memory) HasAccount(id plugin.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.balances[id]
	return ok
}

func (m *memory) CreateAccount(id plugin.UUID, _ string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.balances[id]; ok {
		return false
	}
	m.balances[id] = 0
	return true
}

func (m *memory) Balance(id plugin.UUID) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.balances[id]
}

func (m *memory) Has(id plugin.UUID, amount float64) bool { return m.Balance(id) >= amount }

func (m *memory) Withdraw(id plugin.UUID, amount float64) vault.Response {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances[id] < amount {
		return vault.Response{Balance: m.balances[id], Err: errors.New("insufficient funds")}
	}
	m.balances[id] -= amount
	return vault.Response{Amount: amount, Balance: m.balances[id]}
}

func (m *memory) Deposit(id plugin.UUID, amount float64) vault.Response {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances[id] += amount
	return vault.Response{Amount: amount, Balance: m.balances[id]}
}

func (m *memory) SuperPermsCompat() bool { return false }
func (m *memory) GroupSupport() bool     { return false }

func (m *memory) PlayerHas(_ plugin.World, id plugin.UUID, node string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nodes[id][node]
}

func (m *memory) PlayerAdd(_ plugin.World, id plugin.UUID, node string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes[id] == nil {
		m.nodes[id] = map[string]bool{}
	}
	m.nodes[id][node] = true
	return nil
}

func (m *memory) PlayerRemove(_ plugin.World, id plugin.UUID, node string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.nodes[id], node)
	return nil
}

func (m *memory) PlayerAddTransient(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (m *memory) PlayerRemoveTransient(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}
func (m *memory) GroupHas(plugin.World, string, string) bool           { return false }
func (m *memory) GroupAdd(plugin.World, string, string) error          { return errors.ErrUnsupported }
func (m *memory) GroupRemove(plugin.World, string, string) error       { return errors.ErrUnsupported }
func (m *memory) PlayerInGroup(plugin.World, plugin.UUID, string) bool { return false }
func (m *memory) PlayerAddGroup(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}

func (m *memory) PlayerRemoveGroup(plugin.World, plugin.UUID, string) error {
	return errors.ErrUnsupported
}
func (m *memory) PlayerGroups(plugin.World, plugin.UUID) []string { return nil }
func (m *memory) PrimaryGroup(plugin.World, plugin.UUID) string   { return "" }
func (m *memory) Groups() []string                                { return nil }

func (m *memory) PlayerPrefix(w plugin.World, id plugin.UUID) string {
	v, _ := m.PlayerInfo(w, id, "prefix")
	return v
}

func (m *memory) PlayerSuffix(w plugin.World, id plugin.UUID) string {
	v, _ := m.PlayerInfo(w, id, "suffix")
	return v
}

func (m *memory) SetPlayerPrefix(w plugin.World, id plugin.UUID, prefix string) error {
	return m.SetPlayerInfo(w, id, "prefix", prefix)
}

func (m *memory) SetPlayerSuffix(w plugin.World, id plugin.UUID, suffix string) error {
	return m.SetPlayerInfo(w, id, "suffix", suffix)
}
func (m *memory) GroupPrefix(plugin.World, string) string               { return "" }
func (m *memory) GroupSuffix(plugin.World, string) string               { return "" }
func (m *memory) SetGroupPrefix(plugin.World, string, string) error     { return errors.ErrUnsupported }
func (m *memory) SetGroupSuffix(plugin.World, string, string) error     { return errors.ErrUnsupported }
func (m *memory) GroupInfo(plugin.World, string, string) (string, bool) { return "", false }
func (m *memory) SetGroupInfo(plugin.World, string, string, string) error {
	return errors.ErrUnsupported
}

func (m *memory) PlayerInfo(_ plugin.World, id plugin.UUID, node string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.info[id][node]
	return v, ok
}

func (m *memory) SetPlayerInfo(_ plugin.World, id plugin.UUID, node, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.info[id] == nil {
		m.info[id] = map[string]string{}
	}
	if value == "" {
		delete(m.info[id], node)
	} else {
		m.info[id][node] = value
	}
	return nil
}
