// Copyright (c) 2024 Fantom Foundation
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at fantom.foundation/bsl11.
//
// Change Date: 2028-4-16
//
// On the date above, in accordance with the Business Source License, use of
// this software will be governed by the GNU Lesser General Public License v3.

package interpreter_test

// This file contains a few initial shake-down tests or a Processor implementation.
// Right now, the tested features are minimal. Follow-up work is needed to systematically
// establish a set of test cases for Processor features.
//
// TODO:
// - test gas price charging
// - test gas refunding
// - test left-over gas refunding
// - test recursive calls
// - test roll-back on revert
// - improve test setup
// - find better place for those tests

import (
	"fmt"
	"testing"

	"github.com/Fantom-foundation/Tosca/go/integration_test"
	"github.com/Fantom-foundation/Tosca/go/tosca"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/mock/gomock"

	_ "github.com/Fantom-foundation/Tosca/go/processor/opera"

	// This is only imported to get the EVM opcode definitions.
	// TODO: write up our own op-code definition and remove this dependency.
	op "github.com/ethereum/go-ethereum/core/vm"
)

func TestProcessor_SimpleValueTransfer(t *testing.T) {
	const transferCosts = tosca.Gas(21_000)

	// Transfer 3 tokens from account 1 to account 2.
	transaction := tosca.Transaction{
		Sender:    tosca.Address{1},
		Recipient: &tosca.Address{2},
		Value:     tosca.ValueFromUint64(3),
		Nonce:     4,
		GasLimit:  transferCosts,
	}

	for name, processor := range getProcessors() {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			blockParams := tosca.BlockParameters{}

			// TODO: clean up expectations
			// TODO: provide a better way to define expectations that is less
			// sensitive to implementation details; focus on effects
			// - use a before/after pattern
			context := tosca.NewMockTransactionContext(ctrl)

			context.EXPECT().CreateSnapshot()
			context.EXPECT().GetBalance(tosca.Address{1}).Return(tosca.ValueFromUint64(10)).AnyTimes()
			context.EXPECT().GetBalance(tosca.Address{2}).Return(tosca.ValueFromUint64(5)).AnyTimes()

			context.EXPECT().AccountExists(tosca.Address{2}).Return(true)
			context.EXPECT().GetCode(tosca.Address{2}).Return([]byte{})
			context.EXPECT().GetNonce(tosca.Address{1}).Return(uint64(4)).Times(2)
			context.EXPECT().SetNonce(tosca.Address{1}, uint64(5)).Return()
			context.EXPECT().GetCodeHash(tosca.Address{1}).Return(tosca.Hash{})

			context.EXPECT().SetBalance(tosca.Address{1}, tosca.ValueFromUint64(10)).MinTimes(1) // < charging gas, but price is zero
			context.EXPECT().SetBalance(tosca.Address{1}, tosca.ValueFromUint64(7))              // < withdraw 3 tokens
			context.EXPECT().SetBalance(tosca.Address{2}, tosca.ValueFromUint64(8))              // < deposit 3 tokens

			context.EXPECT().GetLogs()

			// Execute the transaction.
			receipt, err := processor.Run(blockParams, transaction, context)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			// Check the result.
			if got, want := receipt.Success, true; got != want {
				t.Errorf("unexpected success: got %v, want %v", got, want)
			}
			if want, got := transferCosts, receipt.GasUsed; want != got {
				t.Errorf("unexpected gas costs: want %d, got %d", want, got)
			}
		})
	}
}

func TestProcessor_ContractCallThatSucceeds(t *testing.T) {
	const gasCosts = tosca.Gas(21_000 + 2*3)

	// A call to the contract at address 2 paid by account 1.
	transaction := tosca.Transaction{
		Sender:    tosca.Address{1},
		Recipient: &tosca.Address{2},
		Nonce:     4,
		GasLimit:  gasCosts,
	}

	for name, processor := range getProcessors() {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			blockParams := tosca.BlockParameters{}

			context := tosca.NewMockTransactionContext(ctrl)

			context.EXPECT().CreateSnapshot()
			context.EXPECT().GetBalance(tosca.Address{1}).Return(tosca.ValueFromUint64(10)).AnyTimes()
			context.EXPECT().GetBalance(tosca.Address{2}).Return(tosca.ValueFromUint64(5)).AnyTimes()

			context.EXPECT().GetNonce(tosca.Address{1}).Return(uint64(4)).Times(2)
			context.EXPECT().SetNonce(tosca.Address{1}, uint64(5)).Return()
			context.EXPECT().GetCodeHash(tosca.Address{1}).Return(tosca.Hash{})

			context.EXPECT().AccountExists(tosca.Address{2}).Return(true)

			code := []byte{
				byte(op.PUSH1), byte(0), // < push 0
				byte(op.PUSH1), byte(0), // < push 0
				byte(op.RETURN),
			}
			context.EXPECT().GetCode(tosca.Address{2}).Return(code)
			context.EXPECT().GetCodeHash(tosca.Address{2}).Return(integration_test.Keccak256Hash(code))

			context.EXPECT().SetBalance(tosca.Address{1}, tosca.ValueFromUint64(10)).MinTimes(1) // < charging gas, but price is zero

			context.EXPECT().GetLogs()

			// Execute the transaction.
			receipt, err := processor.Run(blockParams, transaction, context)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			// Check the result.
			if got, want := receipt.Success, true; got != want {
				t.Errorf("unexpected success: got %v, want %v", got, want)
			}
			if want, got := gasCosts, receipt.GasUsed; want != got {
				t.Errorf("unexpected gas costs: want %d, got %d", want, got)
			}
		})
	}
}

func TestProcessor_ContractCallThatReverts(t *testing.T) {
	const gasCosts = tosca.Gas(21_000 + 2*3)

	// A call to the contract at address 2 paid by account 1.
	transaction := tosca.Transaction{
		Sender:    tosca.Address{1},
		Recipient: &tosca.Address{2},
		Nonce:     4,
		GasLimit:  gasCosts,
	}

	for name, processor := range getProcessors() {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			blockParams := tosca.BlockParameters{}

			context := tosca.NewMockTransactionContext(ctrl)

			context.EXPECT().CreateSnapshot().Return(tosca.Snapshot(12))
			context.EXPECT().RestoreSnapshot(tosca.Snapshot(12))

			context.EXPECT().GetBalance(tosca.Address{1}).Return(tosca.ValueFromUint64(10)).AnyTimes()
			context.EXPECT().GetBalance(tosca.Address{2}).Return(tosca.ValueFromUint64(5)).AnyTimes()

			context.EXPECT().GetNonce(tosca.Address{1}).Return(uint64(4)).Times(2)
			context.EXPECT().SetNonce(tosca.Address{1}, uint64(5)).Return()

			context.EXPECT().GetCodeHash(tosca.Address{1}).Return(tosca.Hash{})

			context.EXPECT().AccountExists(tosca.Address{2}).Return(true)

			code := []byte{
				byte(op.PUSH1), byte(0), // < push 0
				byte(op.PUSH1), byte(0), // < push 0
				byte(op.REVERT),
			}
			context.EXPECT().GetCode(tosca.Address{2}).Return(code)
			context.EXPECT().GetCodeHash(tosca.Address{2}).Return(integration_test.Keccak256Hash(code))

			context.EXPECT().SetBalance(tosca.Address{1}, tosca.ValueFromUint64(10)).MinTimes(1) // < charging gas, but price is zero

			context.EXPECT().GetLogs()

			// Execute the transaction.
			receipt, err := processor.Run(blockParams, transaction, context)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			// Check the result.
			if got, want := receipt.Success, false; got != want {
				t.Errorf("unexpected success: got %v, want %v", got, want)
			}

			if want, got := gasCosts, receipt.GasUsed; want != got {
				t.Errorf("unexpected gas costs: want %d, got %d", want, got)
			}
		})
	}
}

func TestProcessor_ContractCreation(t *testing.T) {
	const gasCosts = tosca.Gas(53_000)

	// Transfer 3*2^(31*8) tokens from account 1 to account 2.
	transaction := tosca.Transaction{
		Sender:   tosca.Address{1},
		Nonce:    4,
		GasLimit: gasCosts,
	}

	// The new contract address is derived from the sender's address and the nonce.
	newContractAddress := tosca.Address(crypto.CreateAddress(common.Address{1}, 4))

	for name, processor := range getProcessors() {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			blockParams := tosca.BlockParameters{}

			context := tosca.NewMockTransactionContext(ctrl)

			context.EXPECT().CreateSnapshot()
			context.EXPECT().GetBalance(tosca.Address{1}).Return(tosca.ValueFromUint64(10)).AnyTimes()
			context.EXPECT().SetBalance(tosca.Address{1}, tosca.ValueFromUint64(10)).MinTimes(1) // < charging gas, but price is zero

			context.EXPECT().GetNonce(tosca.Address{1}).Return(uint64(4)).Times(3)
			context.EXPECT().SetNonce(tosca.Address{1}, uint64(5)).Return()

			context.EXPECT().AccountExists(newContractAddress).Return(false)
			context.EXPECT().GetNonce(newContractAddress).Return(uint64(0))
			context.EXPECT().SetNonce(newContractAddress, uint64(1)).Return()

			context.EXPECT().GetCodeHash(tosca.Address{1}).Return(tosca.Hash{})
			context.EXPECT().GetCodeHash(newContractAddress).Return(tosca.Hash{})

			context.EXPECT().SetCode(newContractAddress, gomock.Any()).Do(func(address tosca.Address, code []byte) {
				if len(code) != 0 {
					t.Fatalf("unexpected code: %x", code)
				}
			})

			context.EXPECT().GetLogs()

			// Execute the transaction.
			receipt, err := processor.Run(blockParams, transaction, context)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			// Check the result.
			if got, want := receipt.Success, true; got != want {
				t.Errorf("unexpected success: got %v, want %v", got, want)
			}

			if want, got := gasCosts, receipt.GasUsed; want != got {
				t.Errorf("unexpected gas costs: want %d, got %d", want, got)
			}
			if receipt.ContractAddress == nil {
				t.Fatalf("created contract address not set in result")
			}
			if *receipt.ContractAddress != newContractAddress {
				t.Errorf("unexpected result for created contract address, wanted %v, got %v", newContractAddress, *receipt.ContractAddress)
			}
		})
	}
}

// getProcessors returns a map containing all registered processors instantiated
// with all registered interpreters.
func getProcessors() map[string]tosca.Processor {
	interpreter := tosca.GetAllRegisteredInterpreters()
	factories := tosca.GetAllRegisteredProcessorFactories()

	res := map[string]tosca.Processor{}
	for processorName, factory := range factories {
		for interpreterName, interpreter := range interpreter {
			processor := factory(interpreter)
			res[fmt.Sprintf("%s/%s", processorName, interpreterName)] = processor
		}
	}
	return res
}