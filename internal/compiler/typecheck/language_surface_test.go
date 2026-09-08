package typecheck

import "testing"
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
