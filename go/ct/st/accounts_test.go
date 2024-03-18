package st

import (
	"strings"
	"testing"

	. "github.com/Fantom-foundation/Tosca/go/ct/common"
)

func TestAccounts_MarkWarmMarksAddressesAsWarm(t *testing.T) {
	b := NewAccounts()
	b.MarkWarm(NewAddressFromInt(42))

	if want, got := true, b.IsWarm(NewAddressFromInt(42)); want != got {
		t.Fatalf("IsWarm is broken, want %v, got %v", want, got)
	}
	if want, got := false, b.IsWarm(NewAddressFromInt(8)); want != got {
		t.Fatalf("IsWarm is broken, want %v, got %v", want, got)
	}
}

func TestAccounts_Clone(t *testing.T) {
	a := NewAddressFromInt(42)
	b := NewAddressFromInt(48)
	tests := map[string]struct {
		change func(*Accounts)
	}{
		"add-balance": {func(accounts *Accounts) {
			accounts.Balance[b] = NewU256(3)
		}},
		"modify-balance": {func(accounts *Accounts) {
			accounts.Balance[a] = NewU256(3)
		}},
		"remove-balance": {func(accounts *Accounts) {
			delete(accounts.Balance, a)
		}},
		"add-code": {func(accounts *Accounts) {
			accounts.Code[b] = []byte{byte(ADD), byte(PUSH1), 5, byte(PUSH2)}
		}},
		"modify-code": {func(accounts *Accounts) {
			accounts.Code[a] = []byte{byte(SUB), byte(BALANCE), 5, byte(SHA3)}
		}},
		"remove-code": {func(accounts *Accounts) {
			delete(accounts.Code, a)
		}},
		"mark-cold": {func(accounts *Accounts) {
			accounts.MarkCold(a)
		}},
		"mark-warm": {func(accounts *Accounts) {
			accounts.MarkWarm(b)
		}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			b1 := NewAccounts()
			b1.Balance[a] = NewU256(1)
			b1.Code[a] = []byte{byte(SUB), byte(SWAP1), 5, byte(PUSH2)}
			b1.MarkWarm(a)
			b2 := b1.Clone()
			if !b1.Eq(b2) {
				t.Fatalf("clones are not equal")
			}
			test.change(b2)
			if b1.Eq(b2) {
				t.Errorf("clones are not independent")
			}
		})
	}
}

func TestAccounts_AccountsWithZeroBalanceAreTreatedTheSameByEqAndDiff(t *testing.T) {
	a1 := NewAccounts()
	a1.Balance[Address{1}] = NewU256(0)
	a2 := NewAccounts()

	equal := a1.Eq(a2)
	diff := a1.Diff(a2)

	if equal != (len(diff) == 0) {
		t.Errorf("Eq and Diff not compatible, Eq returns %t, Diff %v", equal, diff)
	}
}

func TestAccounts_Diff(t *testing.T) {
	a := NewAddressFromInt(42)
	b := NewAddressFromInt(48)
	tests := map[string]struct {
		change  func(*Accounts)
		outcome string
	}{
		"add-balance": {func(accounts *Accounts) {
			accounts.Balance[b] = NewU256(3)
		}, "Different balance entry"},
		"modify-balance": {func(accounts *Accounts) {
			accounts.Balance[a] = NewU256(3)
		}, "Different balance entry"},
		"remove-balance": {func(accounts *Accounts) {
			delete(accounts.Balance, a)
		}, "Different balance entry"},
		"add-code": {func(accounts *Accounts) {
			accounts.Code[b] = []byte{byte(ADD), byte(PUSH1), 5, byte(PUSH2)}
		}, "Different code entry"},
		"modify-code": {func(accounts *Accounts) {
			accounts.Code[a] = []byte{byte(SUB), byte(BALANCE), 5, byte(SHA3)}
		}, "Different code entry"},
		"remove-code": {func(accounts *Accounts) {
			delete(accounts.Code, a)
		}, "Different code entry"},
		"mark-cold": {func(accounts *Accounts) {
			accounts.MarkCold(a)
		}, "Different account warm entry"},
		"mark-warm": {func(accounts *Accounts) {
			accounts.MarkWarm(b)
		}, "Different account warm entry"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a1 := NewAccounts()
			a1.Balance[a] = NewU256(1)
			a1.Code[a] = []byte{byte(SUB), byte(SWAP1), 5, byte(PUSH2)}
			a1.MarkWarm(a)
			a2 := a1.Clone()
			diff := a1.Diff(a2)
			if len(diff) != 0 {
				t.Errorf("clones are different: %v", diff)
			}
			test.change(a2)
			diff = a1.Diff(a2)
			if !strings.Contains(diff[0], test.outcome) {
				t.Errorf("difference in accounts not found: %v", diff)
			}
		})
	}
}