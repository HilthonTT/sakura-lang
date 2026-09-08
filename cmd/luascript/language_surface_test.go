package main

import (
	"testing"

	"github.com/hilthontt/luascript/internal/compiler"
	"github.com/hilthontt/luascript/internal/compiler/parser"
	"github.com/hilthontt/luascript/internal/vm"
)
func runScript(t *testing.T, src string) *vm.VM {
	t.Helper()
	chunks, err := compiler.CompileToInstructions(src, parser.NormalMode)
	if err != nil {
		t.Fatalf("compile error: %v\nsource:\n%s", err, src)
	}
	v := vm.New()
	registerAllNatives(v)
	if err := v.Run(chunks[0]); err != nil {
		t.Fatalf("vm error: %v\nsource:\n%s", err, src)
	}
	return v
}
func assertGlobal(t *testing.T, v *vm.VM, name string, want vm.Value) {
	t.Helper()
	got := v.Globals.Get(name)
	if !vm.Equal(got, want) {
		t.Errorf("global %q = %v (%T), want %v (%T)", name, got, got, want, want)
	}
}
func TestDestructureNamed(t *testing.T) {
	v := runScript(t, `
		local cfg = { host = "h", port = 80 }
		local { host, port } = cfg
		a, b = host, port`)
	assertGlobal(t, v, "a", "h")
	assertGlobal(t, v, "b", int64(80))
}
func TestDestructureRenameAndDefault(t *testing.T) {
	v := runScript(t, `
		local cfg = { host = "h", flag = false }
		local { host = addr, timeout or 30, flag or true } = cfg
		a, b, c = addr, timeout, flag`)
	assertGlobal(t, v, "a", "h")
	assertGlobal(t, v, "b", int64(30))
	assertGlobal(t, v, "c", false)
}
func TestDestructureNamedRest(t *testing.T) {
	v := runScript(t, `
		local cfg = { a = 1, b = 2, c = 3 }
		local { a, ...others } = cfg
		x, y, z = others.a, others.b, others.c`)
	assertGlobal(t, v, "x", nil)
	assertGlobal(t, v, "y", int64(2))
	assertGlobal(t, v, "z", int64(3))
}
func TestDestructurePositional(t *testing.T) {
	v := runScript(t, `
		local list = { 10, 20, 30, 40 }
		local [ first, second, ...tail ] = list
		a, b, n, t1, t2 = first, second, #tail, tail[1], tail[2]`)
	assertGlobal(t, v, "a", int64(10))
	assertGlobal(t, v, "b", int64(20))
	assertGlobal(t, v, "n", int64(2))
	assertGlobal(t, v, "t1", int64(30))
	assertGlobal(t, v, "t2", int64(40))
}
func TestDestructureEvaluatesSourceOnce(t *testing.T) {
	v := runScript(t, `
		calls = 0
		local function source()
			calls = calls + 1
			return { a = 1, b = 2 }
		end
		local { a, b } = source()
		sum = a + b`)
	assertGlobal(t, v, "calls", int64(1))
	assertGlobal(t, v, "sum", int64(3))
}
func TestTableSpread(t *testing.T) {
	v := runScript(t, `
		local base = { x = 1, y = 2 }
		local over = { ...base, y = 99, z = 3 }
		a, b, c = over.x, over.y, over.z

		local head, tail = { 1, 2 }, { 3, 4 }
		local all = { 0, ...head, ...tail, 5 }
		n, first, last = #all, all[1], all[6]`)
	assertGlobal(t, v, "a", int64(1))
	assertGlobal(t, v, "b", int64(99))
	assertGlobal(t, v, "c", int64(3))
	assertGlobal(t, v, "n", int64(6))
	assertGlobal(t, v, "first", int64(0))
	assertGlobal(t, v, "last", int64(5))
}
func TestSpreadDoesNotMutateSource(t *testing.T) {
	v := runScript(t, `
		local src = { 1, 2 }
		local copy = { ...src, 3 }
		srcLen, copyLen = #src, #copy`)
	assertGlobal(t, v, "srcLen", int64(2))
	assertGlobal(t, v, "copyLen", int64(3))
}
func TestOptionalChainShortCircuitsWholeChain(t *testing.T) {
	v := runScript(t, `
		local t = { inner = { value = 42 } }
		present = t?.inner?.value
		absent = t?.missing?.value
		none = nil
		fromNil = none?.inner?.value
		bracket = t?["inner"]?["value"]`)
	assertGlobal(t, v, "present", int64(42))
	assertGlobal(t, v, "absent", nil)
	assertGlobal(t, v, "fromNil", nil)
	assertGlobal(t, v, "bracket", int64(42))
}
func TestOptionalMethodCall(t *testing.T) {
	v := runScript(t, `
		local o = { n = 5, get = function(self) return self.n end }
		hit = o?:get()
		local none = nil
		miss = none?:get()`)
	assertGlobal(t, v, "hit", int64(5))
	assertGlobal(t, v, "miss", nil)
}
func TestOptionalChainEvaluatesReceiverOnce(t *testing.T) {
	v := runScript(t, `
		calls = 0
		local function source()
			calls = calls + 1
			return { a = { b = 1 } }
		end
		value = source()?.a?.b`)
	assertGlobal(t, v, "calls", int64(1))
	assertGlobal(t, v, "value", int64(1))
}
func TestCoalesceDistinguishesNilFromFalse(t *testing.T) {
	v := runScript(t, `
		fromNil = nil ?? "d"
		fromFalse = false ?? "d"
		fromValue = 42 ?? "d"
		fromOr = false or "d"`)
	assertGlobal(t, v, "fromNil", "d")
	assertGlobal(t, v, "fromFalse", false)
	assertGlobal(t, v, "fromValue", int64(42))
	assertGlobal(t, v, "fromOr", "d")
}
func TestCoalesceSkipsRightWhenLeftPresent(t *testing.T) {
	v := runScript(t, `
		calls = 0
		local function side()
			calls = calls + 1
			return 0
		end
		r = 1 ?? side()`)
	assertGlobal(t, v, "calls", int64(0))
	assertGlobal(t, v, "r", int64(1))
}
func TestPipeline(t *testing.T) {
	v := runScript(t, `
		local function double(n) return n * 2 end
		local function add(a, b) return a + b end
		local obj = { base = 10, plus = function(self, n) return self.base + n end }

		a = 5 |> double
		b = 5 |> add(3)
		c = 5 |> double |> double
		d = 1 |> obj:plus()
		e = 2 + 3 |> double`)
	assertGlobal(t, v, "a", int64(10))
	assertGlobal(t, v, "b", int64(8))
	assertGlobal(t, v, "c", int64(20))
	assertGlobal(t, v, "d", int64(11))
	assertGlobal(t, v, "e", int64(10))
}
