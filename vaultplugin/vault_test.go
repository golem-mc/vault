package vaultplugin_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/golem-mc/golem/plugintest"
	"github.com/golem-mc/vault"
	_ "github.com/golem-mc/vault/vaultplugin"
)

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
