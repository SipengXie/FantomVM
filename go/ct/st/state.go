package st

import (
	"fmt"
	"strings"

	. "github.com/Fantom-foundation/Tosca/go/ct/common"
)

////////////////////////////////////////////////////////////

type StatusCode int

const (
	Running        StatusCode = iota // still running
	Stopped                          // stopped execution successfully
	Returned                         // finished successfully
	Reverted                         // finished with revert signal
	Failed                           // failed (for any reason)
	NumStatusCodes                   // not an actual status
)

func (s StatusCode) String() string {
	switch s {
	case Running:
		return "running"
	case Stopped:
		return "stopped"
	case Returned:
		return "returned"
	case Reverted:
		return "reverted"
	case Failed:
		return "failed"
	default:
		return fmt.Sprintf("StatusCode(%d)", s)
	}
}

////////////////////////////////////////////////////////////

// State represents an EVM's execution state.
type State struct {
	Status    StatusCode
	Revision  Revision
	Pc        uint16
	Gas       uint64
	GasRefund uint64
	Code      *Code
	Stack     *Stack
	Memory    *Memory
	Storage   *Storage
}

// NewState creates a new State instance with the given code.
func NewState(code *Code) *State {
	return &State{
		Status:   Running,
		Revision: R07_Istanbul,
		Code:     code,
		Stack:    NewStack(),
		Memory:   NewMemory(),
		Storage:  NewStorage(),
	}
}

func (s *State) Clone() *State {
	clone := NewState(s.Code.Clone())
	clone.Status = s.Status
	clone.Revision = s.Revision
	clone.Pc = s.Pc
	clone.Gas = s.Gas
	clone.GasRefund = s.GasRefund
	clone.Stack = s.Stack.Clone()
	clone.Memory = s.Memory.Clone()
	clone.Storage = s.Storage.Clone()
	return clone
}

func (s *State) Eq(other *State) bool {
	// All failure states are considered equal.
	if s.Status == Failed && other.Status == Failed {
		return true
	}
	return s.Status == other.Status &&
		s.Revision == other.Revision &&
		s.Pc == other.Pc &&
		s.Gas == other.Gas &&
		s.GasRefund == other.GasRefund &&
		s.Code.Eq(other.Code) &&
		s.Stack.Eq(other.Stack) &&
		s.Memory.Eq(other.Memory) &&
		s.Storage.Eq(other.Storage)
}

const codeCutoffLength = 20
const stackCutOffLength = 5

func (s *State) String() string {
	builder := strings.Builder{}
	builder.WriteString("{\n")
	builder.WriteString(fmt.Sprintf("\tStatus: %v\n", s.Status))
	builder.WriteString(fmt.Sprintf("\tRevision: %v\n", s.Revision))
	builder.WriteString(fmt.Sprintf("\tPc: %d (0x%04x)\n", s.Pc, s.Pc))
	if !s.Code.IsCode(int(s.Pc)) {
		builder.WriteString("\t    (points to data)\n")
	} else if s.Pc < uint16(len(s.Code.code)) {
		builder.WriteString(fmt.Sprintf("\t    (operation: %v)\n", OpCode(s.Code.code[s.Pc])))
	} else {
		builder.WriteString("\t    (out of bounds)\n")
	}
	builder.WriteString(fmt.Sprintf("\tGas: %d\n", s.Gas))
	builder.WriteString(fmt.Sprintf("\tGas refund: %d\n", s.GasRefund))
	if len(s.Code.code) > codeCutoffLength {
		builder.WriteString(fmt.Sprintf("\tCode: %x... (size: %d)\n", s.Code.code[:codeCutoffLength], len(s.Code.code)))
	} else {
		builder.WriteString(fmt.Sprintf("\tCode: %v\n", s.Code))
	}
	builder.WriteString(fmt.Sprintf("\tStack size: %d\n", s.Stack.Size()))
	for i := 0; i < s.Stack.Size() && i < stackCutOffLength; i++ {
		builder.WriteString(fmt.Sprintf("\t    %d: %v\n", i, s.Stack.Get(i)))
	}
	if s.Stack.Size() > stackCutOffLength {
		builder.WriteString("\t    ...\n")
	}
	builder.WriteString(fmt.Sprintf("\tMemory size: %d\n", s.Memory.Size()))
	builder.WriteString("\tStorage.Current:\n")
	for k, v := range s.Storage.Current {
		builder.WriteString(fmt.Sprintf("\t    [%v]=%v\n", k, v))
	}
	builder.WriteString("\tStorage.Original:\n")
	for k, v := range s.Storage.Original {
		builder.WriteString(fmt.Sprintf("\t    [%v]=%v\n", k, v))
	}
	builder.WriteString("\tStorage.Warm:\n")
	for k := range s.Storage.warm {
		builder.WriteString(fmt.Sprintf("\t    [%v]\n", k))
	}

	builder.WriteString("}")
	return builder.String()
}

func (s *State) Diff(o *State) []string {
	res := []string{}

	if s.Status != o.Status {
		res = append(res, fmt.Sprintf("Different status: %v vs %v", s.Status, o.Status))
	}

	if s.Revision != o.Revision {
		res = append(res, fmt.Sprintf("Different revision: %v vs %v", s.Revision, o.Revision))
	}

	if s.Pc != o.Pc {
		res = append(res, fmt.Sprintf("Different pc: %v vs %v", s.Pc, o.Pc))
	}

	if s.Gas != o.Gas {
		res = append(res, fmt.Sprintf("Different gas: %v vs %v", s.Gas, o.Gas))
	}

	if s.GasRefund != o.GasRefund {
		res = append(res, fmt.Sprintf("Different gas refund: %v vs %v", s.GasRefund, o.GasRefund))
	}

	if !s.Code.Eq(o.Code) {
		res = append(res, s.Code.Diff(o.Code)...)
	}

	if !s.Stack.Eq(o.Stack) {
		res = append(res, s.Stack.Diff(o.Stack)...)
	}

	if !s.Memory.Eq(o.Memory) {
		res = append(res, s.Memory.Diff(o.Memory)...)
	}

	if !s.Storage.Eq(o.Storage) {
		res = append(res, s.Storage.Diff(o.Storage)...)
	}

	return res
}
