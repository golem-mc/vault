package vaultplugin_test

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/golem-mc/golem/plugin"
	"github.com/golem-mc/golem/plugintest"
	"github.com/golem-mc/vault"
	"github.com/golem-mc/vault/vaultplugin"
)

func TestInfoShowsOneVersionPrefix(t *testing.T) {
	for _, version := range []string{"v0.1.0", "0.1.0"} {
		t.Run(version, func(t *testing.T) {
			srv := plugintest.Start(t, plugintest.Ops("Alice"), plugintest.WithPlugins(
				plugin.Registration{Name: "vault", Plugin: &versionedVault{version: version}},
			))
			alice := srv.Join("Alice")

			alice.Run("/vault-info")
			const want = "[Vault] Vault v0.1.0 Information"
			if got := alice.Messages(2); len(got) == 0 || got[0] != want {
				t.Fatalf("/vault-info: got %q, want header %q", got, want)
			}
		})
	}
}

type versionedVault struct {
	vaultplugin.Vault
	version string
}

func (v *versionedVault) Enable(s plugin.Server) error {
	return v.Vault.Enable(versionedPluginServer{Server: s, version: v.version})
}

// versionedPluginServer supplies the version that golem normally reads from build information.
type versionedPluginServer struct {
	plugin.Server
	version string
}

func (s versionedPluginServer) Plugins() []plugin.Info {
	plugins := s.Server.Plugins()
	for i := range plugins {
		if plugins[i].Name == s.Name() {
			plugins[i].Version = s.version
		}
	}
	return plugins
}

func TestInfoListsTheHookedProviders(t *testing.T) {
	srv := plugintest.Start(t, plugintest.Ops("Alice"))
	alice := srv.Join("Alice")
	bob := srv.Join("Bob")
	alice.Messages(2) // Bob's join message

	alice.Run("/vault-info")
	got := alice.Messages(2)
	want := []string{"[Vault] Economy: None [null]", "[Vault] Permission: SuperPerms [SuperPerms]", "[Vault] Chat: None [null]"}
	if len(got) != 4 || !strings.HasPrefix(got[0], "[Vault] Vault v") || !slices.Equal(got[1:], want) {
		t.Fatalf("/vault-info: %q", got)
	}

	bob.Run("/vault-info")
	if got := bob.Messages(2); len(got) == 0 || !strings.Contains(got[0], "Unknown or incomplete command") {
		t.Fatalf("/vault-info as a non-operator: %q", got)
	}

	// The fallback answers golem's own check for an online player only.
	perm, ok := vault.LookupPermission(srv.Host())
	if !ok || perm.Name() != "SuperPerms" {
		t.Fatalf("the fallback Permission is %v", perm)
	}
	if !perm.PlayerHas(nil, alice.Handle().UUID(), "vault.admin") || perm.PlayerHas(nil, bob.Handle().UUID(), "vault.admin") {
		t.Fatal("SuperPerms does not follow HasPermission")
	}
	if err := perm.PlayerAdd(nil, bob.Handle().UUID(), "vault.admin"); !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("SuperPerms PlayerAdd: %v", err)
	}
}

func TestSuperPermsTransientPermissions(t *testing.T) {
	srv := plugintest.Start(t)
	bob := srv.Join("Bob")
	id := bob.Handle().UUID()
	perm, _ := vault.LookupPermission(srv.Host())
	const node = "vault.transient"
	if perm.PlayerHas(nil, id, node) {
		t.Fatal("node set before it was added")
	}
	if err := perm.PlayerAddTransient(nil, id, node); err != nil {
		t.Fatalf("PlayerAddTransient: %v", err)
	}
	if !perm.PlayerHas(nil, id, node) {
		t.Fatal("transient node not granted")
	}
	if err := perm.PlayerRemoveTransient(nil, id, node); err != nil {
		t.Fatalf("PlayerRemoveTransient: %v", err)
	}
	if perm.PlayerHas(nil, id, node) {
		t.Fatal("transient node still granted after removal")
	}
}

func TestConvertCopiesBalancesIntoTheOtherEconomy(t *testing.T) {
	old, fresh := newEconomy("Old Eco"), newEconomy("New")
	regs := append(plugin.Registered(),
		plugin.Registration{Name: "old", Plugin: provides{old, vault.PriorityHigh}},
		plugin.Registration{Name: "fresh", Plugin: provides{fresh, vault.PriorityNormal}})
	srv := plugintest.Start(t, plugintest.WithPlugins(regs...), plugintest.Ops("Alice"))
	alice := srv.Join("Alice")
	bob := srv.Join("Bob")
	bobID := bob.Handle().UUID()
	aliceID := alice.Handle().UUID()
	old.CreateAccount(bobID, "Bob")
	old.Deposit(bobID, 42.5)
	old.CreateAccount(aliceID, "Alice")
	old.Deposit(aliceID, 7)
	fresh.CreateAccount(aliceID, "Alice")
	bob.Disconnect()
	alice.Messages(2) // Bob's join and leave messages

	alice.Run("/vault-info")
	if got := alice.Messages(2); len(got) != 4 || got[1] != "[Vault] Economy: Old Eco [Old Eco, New]" {
		t.Fatalf("/vault-info: %q", got)
	}

	alice.Run("/vault-convert OldEco")
	if got := alice.Messages(2); len(got) != 1 || !strings.HasPrefix(got[0], "You must specify only the economy") {
		t.Fatalf("/vault-convert with one name: %q", got)
	}
	alice.Run("/vault-convert Nope New")
	if got := alice.Messages(2); !slices.Equal(got, []string{"Could not find Nope loaded on the server, check your spelling.", "Valid economies are: OldEco, New"}) {
		t.Fatalf("/vault-convert with an unknown name: %q", got)
	}

	alice.Run("/vault-convert oldeco new")
	want := []string{"This may take some time to convert, expect server lag.", "Converson complete, please verify the data before using it."}
	if got := alice.Messages(2); !slices.Equal(got, want) {
		t.Fatalf("/vault-convert: %q", got)
	}
	if got := fresh.Balance(bobID); got != 42.5 {
		t.Fatalf("Bob's converted balance is %v, want 42.5", got)
	}
	if got := fresh.Balance(aliceID); got != 0 {
		t.Fatalf("Alice already had an account in New, yet it holds %v", got)
	}
}

type provides struct {
	e        *economy
	priority int
}

func (p provides) Enable(s plugin.Server) error {
	vault.ProvideEconomy(s, p.e, p.priority)
	return nil
}

func (provides) Disable() {}

type economy struct {
	name     string
	mu       sync.Mutex
	balances map[plugin.UUID]float64
}

func newEconomy(name string) *economy {
	return &economy{name: name, balances: map[plugin.UUID]float64{}}
}

func (e *economy) Name() string                 { return e.name }
func (e *economy) FractionalDigits() int        { return -1 }
func (e *economy) Format(amount float64) string { return "" }
func (e *economy) CurrencyNameSingular() string { return "" }
func (e *economy) CurrencyNamePlural() string   { return "" }

func (e *economy) HasAccount(id plugin.UUID) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, ok := e.balances[id]
	return ok
}

func (e *economy) CreateAccount(id plugin.UUID, _ string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.balances[id]; ok {
		return false
	}
	e.balances[id] = 0
	return true
}

func (e *economy) Balance(id plugin.UUID) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.balances[id]
}

func (e *economy) Has(id plugin.UUID, amount float64) bool { return e.Balance(id) >= amount }

func (e *economy) Withdraw(id plugin.UUID, amount float64) vault.Response {
	return e.Deposit(id, -amount)
}

func (e *economy) Deposit(id plugin.UUID, amount float64) vault.Response {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.balances[id] += amount
	return vault.Response{Amount: amount, Balance: e.balances[id]}
}
