package vault

import "github.com/golem-mc/golem/plugin"

// Economy is an economy plugin's player accounts, for other plugins to
// charge and pay. Amounts are in the economy's currency and never negative.
// A provider is removed when its plugin is disabled, so there is no
// isEnabled. Vault: Economy
type Economy interface {
	// Name answers the economy's name, as /vault-info shows it and
	// /vault-convert matches it with its spaces removed. Vault: getName
	Name() string
	// FractionalDigits answers how many digits after the decimal point the
	// economy keeps, or -1 when it does not round. Vault: fractionalDigits
	FractionalDigits() int
	// Format writes amount as the economy shows money, e.g. "$1,000.50".
	// Vault: format
	Format(amount float64) string
	// CurrencyNameSingular and CurrencyNamePlural answer the currency's
	// name, e.g. "dollar" and "dollars"; either may be empty.
	// Vault: currencyNameSingular, currencyNamePlural
	CurrencyNameSingular() string
	CurrencyNamePlural() string

	// HasAccount reports whether player has an account.
	// Vault: hasAccount(OfflinePlayer)
	HasAccount(player plugin.UUID) bool
	// CreateAccount opens an account for player, with the starting balance,
	// and reports whether it made one: false when it already existed. name
	// is the player's name, or an NPC account's (a town, a shop) that no
	// player owns. Vault: createPlayerAccount(OfflinePlayer)
	CreateAccount(player plugin.UUID, name string) bool
	// Balance answers player's balance. Vault: getBalance(OfflinePlayer)
	Balance(player plugin.UUID) float64
	// Has reports whether player's balance is at least amount.
	// Vault: has(OfflinePlayer, double)
	Has(player plugin.UUID, amount float64) bool
	// Withdraw takes amount from player's balance. It fails, changing
	// nothing, when the balance is too low and the economy allows no debt.
	// Vault: withdrawPlayer(OfflinePlayer, double)
	Withdraw(player plugin.UUID, amount float64) Response
	// Deposit adds amount to player's balance.
	// Vault: depositPlayer(OfflinePlayer, double)
	Deposit(player plugin.UUID, amount float64) Response
}

// Response is the outcome of a balance change. Vault: EconomyResponse
type Response struct {
	// Amount is the amount moved; Balance is the balance afterwards, as far
	// as the economy knows it.
	Amount, Balance float64
	// Err is nil when the change was made, and otherwise says why not, to
	// show the player; errors.ErrUnsupported for an operation the economy
	// does not have. Vault: type and errorMessage
	Err error
}

// Success reports whether the change was made. Vault: transactionSuccess
func (r Response) Success() bool { return r.Err == nil }

// Banks is the shared accounts an Economy may also keep, which players own
// or are members of. A provider with banks implements it on the same value
// as its Economy; BanksOf finds it. Vault: Economy's bank methods
type Banks interface {
	// Banks answers the names of every bank. Vault: getBanks
	Banks() []string
	// CreateBank opens the bank name, owned by owner. Vault: createBank
	CreateBank(name string, owner plugin.UUID) error
	// DeleteBank closes the bank name. Vault: deleteBank
	DeleteBank(name string) error
	// BankBalance answers the bank's balance. Vault: bankBalance
	BankBalance(name string) (float64, error)
	// BankHas reports whether the bank holds at least amount. Vault: bankHas
	BankHas(name string, amount float64) (bool, error)
	// BankWithdraw and BankDeposit change the bank's balance as Withdraw and
	// Deposit change a player's. Vault: bankWithdraw, bankDeposit
	BankWithdraw(name string, amount float64) Response
	BankDeposit(name string, amount float64) Response
	// IsBankOwner and IsBankMember report whether player owns, or is a
	// member of, the bank. Vault: isBankOwner, isBankMember
	IsBankOwner(name string, player plugin.UUID) (bool, error)
	IsBankMember(name string, player plugin.UUID) (bool, error)
}

// BanksOf answers e's banks, and false when it keeps none.
// Vault: hasBankSupport
func BanksOf(e Economy) (Banks, bool) {
	b, ok := e.(Banks)
	return b, ok
}
