package geth

import (
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/Tosca/go/ct"
	"github.com/Fantom-foundation/Tosca/go/ct/common"
	"github.com/Fantom-foundation/Tosca/go/ct/st"
	"github.com/Fantom-foundation/Tosca/go/ct/utils"
	"github.com/Fantom-foundation/Tosca/go/vm"
	geth_common "github.com/ethereum/go-ethereum/common"
	geth_vm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
)

func NewConformanceTestingTarget() ct.Evm {
	return ctAdapter{}
}

type ctAdapter struct{}

func (a ctAdapter) StepN(state *st.State, numSteps int) (*st.State, error) {
	// Hack: Special handling for unknown revision, because geth cannot represent an invalid revision.
	// So we mark the status as failed already.
	// TODO: Fix this once we add full revision support to the CT and geth.
	if state.Revision > common.R10_London {
		state.Status = st.Failed
		return state, nil
	}

	// No need to run anything that is not in a running state.
	if state.Status != st.Running {
		return state, nil
	}

	op, err := state.Code.GetOperation(int(state.Pc))
	isStopInstruction := false
	if err == nil && op == common.STOP {
		isStopInstruction = true
	}

	parameters := utils.ToVmParameters(state)

	evm, contract, stateDb := createGethInterpreterContext(parameters)
	stateDb.refund = uint64(state.GasRefund)

	evm.CallContext = &callInterceptor{parameters.Context, stateDb}

	interpreterState := geth_vm.NewGethState(
		contract,
		convertCtMemoryToGethMemory(state),
		convertCtStackToGethStack(state),
		uint64(state.Pc),
	)
	interpreterState.Result = state.LastCallReturnData.ToBytes()
	interpreterState.ReadOnly = state.ReadOnly

	interpreter := evm.Interpreter().(*geth_vm.GethEVMInterpreter)
	for i := 0; i < numSteps && !interpreterState.Halted; i++ {
		if !interpreter.Step(interpreterState) {
			break
		}
	}

	// Update the resulting state.
	state.Status, err = convertGethStatusToCtStatus(interpreterState)
	if err != nil {
		return nil, err
	}
	if state.Status == st.Running {
		state.Pc = uint16(interpreterState.Pc)
	}

	state.Gas = vm.Gas(contract.Gas)
	state.GasRefund = vm.Gas(stateDb.GetRefund())
	state.Stack = convertGethStackToCtStack(interpreterState)
	state.Memory = convertGethMemoryToCtMemory(interpreterState)
	state.LastCallReturnData = common.NewBytes(interpreterState.Result)
	if state.Status == st.Stopped || state.Status == st.Reverted {
		// Right now, the interpreter state does not allow to decide whether the
		// stopped state was reached through a STOP or RETURN instruction. Only
		// in the latter case the interpreterState.Result should be assigned to
		// the resulting state.ReturnData.
		// In general, this should be fixed in the go-ethereum-substate repository
		// by providing the necessary information in the state. However, the CT
		// integration in this repository will have to be re-done in the future,
		// when upgrading to a newer go-ethereum version. Thus, for now, this
		// local check is performed to determine whether the result should be
		// copied or not.
		if !isStopInstruction {
			state.ReturnData = common.NewBytes(interpreterState.Result)
		}
	}
	return state, nil

}

func convertGethStatusToCtStatus(state *geth_vm.GethState) (st.StatusCode, error) {
	if !state.Halted && state.Err == nil {
		return st.Running, nil
	}

	if state.Err == geth_vm.ErrExecutionReverted {
		return st.Reverted, nil
	}

	if state.Err != nil {
		return st.Failed, nil
	}

	if state.Halted {
		return st.Stopped, nil
	}

	return st.Failed, fmt.Errorf("unable to convert geth status to ct status")
}

func convertCtMemoryToGethMemory(state *st.State) *geth_vm.Memory {
	data := state.Memory.Read(0, uint64(state.Memory.Size()))
	memory := geth_vm.NewMemory()
	// Set internal memory gas cost state so future grow operations compute the correct cost.
	geth_vm.MemoryGasCost(memory, uint64(len(data)))
	memory.Resize(uint64(len(data)))
	memory.Set(0, uint64(len(data)), data)
	return memory
}

func convertGethMemoryToCtMemory(state *geth_vm.GethState) *st.Memory {
	memory := st.NewMemory()
	memory.Set(state.Memory.Data())
	return memory
}

func convertCtStackToGethStack(state *st.State) *geth_vm.Stack {
	stack := geth_vm.NewStack()
	for i := state.Stack.Size() - 1; i >= 0; i-- {
		val := state.Stack.Get(i).Uint256()
		stack.Push(&val)
	}
	return stack
}

func convertGethStackToCtStack(state *geth_vm.GethState) *st.Stack {
	stack := st.NewStack()
	for i := 0; i < state.Stack.Len(); i++ {
		val := state.Stack.Data()[i]
		stack.Push(common.NewU256(val[3], val[2], val[1], val[0]))
	}
	return stack
}

type callInterceptor struct {
	context vm.RunContext
	stateDb *stateDbAdapter
}

func (i *callInterceptor) Call(env *geth_vm.EVM, me geth_vm.ContractRef, addr geth_common.Address, data []byte, gas uint64, value *big.Int) ([]byte, uint64, error) {
	have := i.stateDb.GetBalance(me.Address())
	if value.Cmp(have) > 0 {
		return nil, gas, geth_vm.ErrInsufficientBalance
	}

	var vmValue vm.Value
	value.FillBytes(vmValue[:])
	res, _ := i.context.Call(vm.Call, vm.CallParameter{
		Sender:    vm.Address(me.Address()),
		Recipient: vm.Address(addr),
		Value:     vmValue,
		Input:     data,
		Gas:       vm.Gas(gas),
	})

	i.handleGasRefund(res.GasRefund)
	err := geth_vm.ErrExecutionReverted
	if res.Success {
		err = nil
	}
	return res.Output, uint64(res.GasLeft), err
}

func (i *callInterceptor) CallCode(env *geth_vm.EVM, me geth_vm.ContractRef, addr geth_common.Address, data []byte, gas uint64, value *big.Int) ([]byte, uint64, error) {
	panic("not implemented")
}

func (i *callInterceptor) DelegateCall(env *geth_vm.EVM, me geth_vm.ContractRef, addr geth_common.Address, data []byte, gas uint64) ([]byte, uint64, error) {
	panic("not implemented")
}

func (i *callInterceptor) StaticCall(env *geth_vm.EVM, me geth_vm.ContractRef, addr geth_common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	panic("not implemented")
}

func (i *callInterceptor) Create(env *geth_vm.EVM, me geth_vm.ContractRef, code []byte, gas uint64, value *big.Int) ([]byte, geth_common.Address, uint64, error) {
	panic("not implemented")
}

func (i *callInterceptor) Create2(env *geth_vm.EVM, me geth_vm.ContractRef, code []byte, gas uint64, endowment *big.Int, salt *uint256.Int) ([]byte, geth_common.Address, uint64, error) {
	panic("not implemented")
}

func (i *callInterceptor) handleGasRefund(refund vm.Gas) {
	if refund < 0 {
		i.stateDb.SubRefund(uint64(-refund))
	} else {
		i.stateDb.AddRefund(uint64(refund))
	}
}