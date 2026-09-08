package typecheck

import "testing"
func TestDestructureResolvesFieldTypes(t *testing.T) {
	expectOK(t, `
		type Server = { host: string, port: number }
		local function f(s: Server): string
			local { host, port } = s
			return host .. port
		end`)
}
func TestDestructureRejectsUnknownField(t *testing.T) {
	expectErrContains(t, `
		local s: { host: string } = { host = "h" }
		local { missing } = s`,
		`has no field "missing"`)
}
func TestDestructureRejectsNonTable(t *testing.T) {
	expectErrContains(t, `
		local n: number = 1
		local { x } = n`,
		"cannot destructure")
}
func TestDestructureRenameBindsUnderNewName(t *testing.T) {
	expectErrContains(t, `
		local s: { host: string } = { host = "h" }
		local { host = addr } = s
		local n: number = addr`,
		`could not be converted`)
}
func TestDestructureAnnotationIsChecked(t *testing.T) {
	expectErrContains(t, `
		local s: { host: string } = { host = "h" }
		local { host: number } = s`,
		"could not be converted")
}
func TestDestructureDefaultDropsNil(t *testing.T) {
	expectOK(t, `
		local s: { port: number? } = { }
		local { port or 80 } = s
		local n: number = port`)
}
func TestSpreadRejectsNonTable(t *testing.T) {
	expectErrContains(t, `
		local n: number = 1
		local t = { ...n }`,
		"cannot spread")
}
func TestOptionalAccessAddsNil(t *testing.T) {
	expectErrContains(t, `
		local t: { a: number } = { a = 1 }
		local n: number = t?.a`,
		"could not be converted")
}
func TestOptionalAccessCoalescedIsNotOptional(t *testing.T) {
	expectOK(t, `
		local t: { a: number } = { a = 1 }
		local n: number = t?.a ?? 0`)
}
func TestCoalesceKeepsRightWhenLeftIsNil(t *testing.T) {
	expectOK(t, `
		local a: string? = nil
		local s: string = a ?? "fallback"`)
}
func TestCoalesceUnionsBothSides(t *testing.T) {
	expectErrContains(t, `
		local a: string? = nil
		local s: string = a ?? 1`,
		"could not be converted")
}
func TestPipelineChecksArguments(t *testing.T) {
	expectErrContains(t, `
		local function f(n: number): number return n end
		local r = "s" |> f`,
		"could not be converted")
}
func TestTypeParamConstraintRejectsBadArgument(t *testing.T) {
	expectErrContains(t, `
		interface Named { name: string }
		local function f<T: Named>(x: T): T return x end
		local r = f(42)`,
		"does not satisfy the constraint")
}
func TestTypeParamConstraintAcceptsGoodArgument(t *testing.T) {
	expectOK(t, `
		interface Named { name: string }
		local function f<T: Named>(x: T): T return x end
		local r = f({ name = "ok" })`)
}
func TestConstraintOnGenericAliasIsChecked(t *testing.T) {
	expectErrContains(t, `
		interface Named { name: string }
		type Box<T: Named> = { value: T }
		local b: Box<number> = { value = 1 }`,
		"does not satisfy the constraint")
}
func TestBoundedTypeParamResolvesFields(t *testing.T) {
	expectOK(t, `
		interface Named { name: string }
		local function f<T: Named>(x: T): string
			return x.name
		end`)
}
func TestIntersectionMergesFields(t *testing.T) {
	expectOK(t, `
		interface Named { name: string }
		interface Aged { age: number }
		type Person = Named & Aged
		local function f(p: Person): string
			return p.name .. p.age
		end`)
}
func TestIntersectionRejectsNonTables(t *testing.T) {
	expectErrContains(t, `
		type Bad = number & string
		local b: Bad = 1`,
		"only table types can be intersected")
}
func TestIntersectionRequiresEveryMember(t *testing.T) {
	expectErrContains(t, `
		interface Named { name: string }
		interface Aged { age: number }
		local function f(p: Named & Aged): number return p.age end
		local n = f({ name = "x" } :: Named)`,
		"could not be converted")
}
