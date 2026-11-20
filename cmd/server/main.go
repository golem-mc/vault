// Command server is the golem server with Vault built in.
package main

import (
	"github.com/golem-mc/golem/app"
	_ "github.com/golem-mc/vault/vaultplugin"
)

func main() { app.Main() }
