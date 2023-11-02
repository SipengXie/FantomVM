package st

import (
	"testing"

	"github.com/Fantom-foundation/Tosca/go/ct"
)

func TestStack_NewStack(t *testing.T) {
	stack := NewStack()
	if want, got := 0, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}

	stack = NewStack(ct.NewU256(1))
	if want, got := 1, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}
}

func TestStack_NewStackWithSize(t *testing.T) {
	stack := NewStackWithSize(5)
	if want, got := 5, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}
	for i := 0; i < stack.Size(); i++ {
		if !stack.Get(i).Eq(ct.NewU256(0)) {
			t.Errorf("unexpected non-zero value at index %d", i)
		}
	}
}

func TestStack_Clone(t *testing.T) {
	stack := NewStack(ct.NewU256(42))
	clone := stack.Clone()

	if stack.Size() != clone.Size() {
		t.Error("Clone does not have the same size")
	}

	stack.Push(ct.NewU256(21))
	if clone.Size() != 1 {
		t.Error("Clone is not independent from original")
	}

	stack.Set(1, ct.NewU256(43))
	if !clone.Get(0).Eq(ct.NewU256(42)) {
		t.Error("Clone is not independent from original")
	}
}

func TestStack_Get(t *testing.T) {
	stack := NewStack(ct.NewU256(1), ct.NewU256(2), ct.NewU256(3))
	if want, got := uint64(3), stack.Get(0).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 0, want %v, got %v", want, got)
	}
	if want, got := uint64(2), stack.Get(1).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 1, want %v, got %v", want, got)
	}
	if want, got := uint64(1), stack.Get(2).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 2, want %v, got %v", want, got)
	}
}

func TestStack_Set(t *testing.T) {
	stack := NewStack(ct.NewU256(2))
	stack.Set(0, ct.NewU256(4))
	if want, got := uint64(4), stack.Get(0).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 0, want %v, got %v", want, got)
	}
}

func TestStack_Push(t *testing.T) {
	stack := NewStack()

	stack.Push(ct.NewU256(42))
	if want, got := 1, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}
	if want, got := uint64(42), stack.Get(0).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 0, want %v, got %v", want, got)
	}

	stack.Push(ct.NewU256(16))
	if want, got := 2, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}
	if want, got := uint64(16), stack.Get(0).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 0, want %v, got %v", want, got)
	}
	if want, got := uint64(42), stack.Get(1).Uint64(); want != got {
		t.Errorf("unexpected stack value at position 1, want %v, got %v", want, got)
	}
}

func TestStack_Pop(t *testing.T) {
	stack := NewStack(ct.NewU256(1), ct.NewU256(2))

	value := stack.Pop().Uint64()
	if value != 2 {
		t.Errorf("unexpected value popped, want 2, got %v", value)
	}
	if want, got := 1, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}

	value = stack.Pop().Uint64()
	if value != 1 {
		t.Errorf("unexpected value popped, want 1, got %v", value)
	}
	if want, got := 0, stack.Size(); want != got {
		t.Errorf("unexpected stack size, want %v, got %v", want, got)
	}
}

func TestStack_Eq(t *testing.T) {
	stack1 := NewStack(ct.NewU256(1), ct.NewU256(2))
	stack2 := NewStack(ct.NewU256(1), ct.NewU256(2))
	if !stack1.Eq(stack2) {
		t.Errorf("unexpected stack inequality %v vs. %v", stack1.stack, stack2.stack)
	}

	stack2.Set(0, ct.NewU256(42))
	if stack1.Eq(stack2) {
		t.Errorf("unexpected stack equality %v vs. %v", stack1.stack, stack2.stack)
	}

	stack2.Pop()
	if stack1.Eq(stack2) {
		t.Errorf("unexpected stack equality %v vs. %v", stack1.stack, stack2.stack)
	}
}