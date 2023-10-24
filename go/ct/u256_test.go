package ct

import (
	"bytes"
	"fmt"
	"testing"
)

func TestU256IsZero(t *testing.T) {
	zero := NewU256()
	if !zero.IsZero() {
		t.Fail()
	}
	one := NewU256(1)
	if one.IsZero() {
		t.Fail()
	}
}

func TestU256Bytes32be(t *testing.T) {
	x := NewU256(1, 2, 3, 4)
	xBytes := x.Bytes32be()
	if !bytes.Equal(xBytes[:], []byte{0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Fail()
	}
}

func TestU256Bytes20be(t *testing.T) {
	x := NewU256(1, 2, 3, 4)
	xBytes := x.Bytes20be()
	if !bytes.Equal(xBytes[:], []byte{0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Fail()
	}
}

func TestU256Eq(t *testing.T) {
	a := NewU256(1, 2, 3, 4)
	b := NewU256(0, 0, 0, 4)
	if !a.Eq(a) {
		t.Fail()
	}
	if a.Eq(b) {
		t.Fail()
	}
}

func TestU256Ne(t *testing.T) {
	a := NewU256(1, 2, 3, 4)
	b := NewU256(0, 0, 0, 4)
	if a.Ne(a) {
		t.Fail()
	}
	if !a.Ne(b) {
		t.Fail()
	}
}

func TestU256Lt(t *testing.T) {
	a := NewU256(1, 2, 3, 4)
	b := NewU256(0, 0, 0, 4)
	if a.Lt(a) {
		t.Fail()
	}
	if a.Lt(b) {
		t.Fail()
	}
	if !b.Lt(a) {
		t.Fail()
	}
}

func TestU256Slt(t *testing.T) {
	if !MaxU256().Slt(NewU256(0)) {
		t.Fail()
	}
}

func TestU256Gt(t *testing.T) {
	a := NewU256(1, 2, 3, 4)
	b := NewU256(0, 0, 0, 4)
	if a.Gt(a) {
		t.Fail()
	}
	if !a.Gt(b) {
		t.Fail()
	}
	if b.Gt(a) {
		t.Fail()
	}
}

func TestU256Sgt(t *testing.T) {
	zero := NewU256(0)
	if !zero.Sgt(MaxU256()) {
		t.Fail()
	}
}

func TestU256Add(t *testing.T) {
	a := NewU256(17)
	b := NewU256(13)
	if a.Add(b).Ne(NewU256(17 + 13)) {
		t.Fail()
	}
	if MaxU256().Add(NewU256(1)).Ne(NewU256(0)) {
		t.Fail()
	}
}

func TestU256AddMod(t *testing.T) {
	a := NewU256(10)
	if a.AddMod(NewU256(10), NewU256(8)).Ne(NewU256(4)) {
		t.Fail()
	}
	if MaxU256().AddMod(NewU256(2), NewU256(2)).Ne(NewU256(1)) {
		t.Fail()
	}
}

func TestU256Sub(t *testing.T) {
	a := NewU256(17)
	b := NewU256(13)
	if a.Sub(b).Ne(NewU256(17 - 13)) {
		t.Fail()
	}
	zero := NewU256(0)
	if zero.Sub(NewU256(1)).Ne(MaxU256()) {
		t.Fail()
	}
}

func TestU256Mul(t *testing.T) {
	a := NewU256(17)
	b := NewU256(13)
	if a.Mul(b).Ne(NewU256(17 * 13)) {
		t.Fail()
	}
}

func TestU256MulMod(t *testing.T) {
	a := NewU256(10)
	if a.MulMod(NewU256(10), NewU256(8)).Ne(NewU256(4)) {
		t.Fail()
	}
	if MaxU256().MulMod(MaxU256(), NewU256(12)).Ne(NewU256(9)) {
		t.Fail()
	}
}

func TestU256Div(t *testing.T) {
	a := NewU256(24)
	b := NewU256(8)
	if a.Div(b).Ne(NewU256(24 / 8)) {
		t.Fail()
	}
}

func TestU256Mod(t *testing.T) {
	a := NewU256(25)
	b := NewU256(8)
	if a.Mod(b).Ne(NewU256(25 % 8)) {
		t.Fail()
	}
}

func TestU256SDiv(t *testing.T) {
	a := MaxU256().Sub(NewU256(1))
	b := MaxU256()
	if a.SDiv(b).Ne(NewU256(2)) {
		t.Fail()
	}
}

func TestU256SMod(t *testing.T) {
	a := MaxU256().Sub(NewU256(7))
	b := MaxU256().Sub(NewU256(2))
	if a.SMod(b).Ne(MaxU256().Sub(NewU256(1))) {
		t.Fail()
	}
}

func TestU256Exp(t *testing.T) {
	a := NewU256(7)
	b := NewU256(5)
	if a.Exp(b).Ne(NewU256(16807)) {
		t.Fail()
	}
}

func TestU256Not(t *testing.T) {
	zero := NewU256(0)
	if zero.Not().Ne(MaxU256()) {
		t.Fail()
	}
}

func TestU256Shl(t *testing.T) {
	x := NewU256(42)
	if x.Shl(NewU256(64)).Ne(NewU256(0, 42)) {
		t.Fail()
	}
}
func TestU256Shr(t *testing.T) {
	x := NewU256(0, 42)
	if x.Shr(NewU256(64)).Ne(NewU256(42)) {
		t.Fail()
	}
}

func TestU256String(t *testing.T) {
	x := NewU256(42, 13, 47, 1)
	if fmt.Sprint(x) != "000000000000002a 000000000000000d 000000000000002f 0000000000000001" {
		t.Fail()
	}
}