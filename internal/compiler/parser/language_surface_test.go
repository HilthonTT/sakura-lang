package parser

import (
	"testing"

	"github.com/hilthontt/luascript/internal/compiler/ast"
)
func TestParseInterfaceStatement(t *testing.T) {
	ta := parseExpect1(t, "interface Named { name: string }").(*ast.TypeAliasStatement)
	if !ta.IsInterface || ta.Name != "Named" {
		t.Errorf("alias = %+v, want interface Named", ta)
	}
	if _, ok := ta.Target.(*ast.TypeTable); !ok {
		t.Errorf("target = %T, want *ast.TypeTable", ta.Target)
	}
}
func TestInterfaceIsContextual(t *testing.T) {
	parse(t, "local interface = 1 print(interface)")
}
func TestParseTypeParamConstraint(t *testing.T) {
	lf := parseExpect1(t, "local function f<T: Named, U>(x: T): T return x end").(*ast.LocalFunctionStatement)
	if len(lf.Func.TypeParams) != 2 {
		t.Fatalf("type params = %d, want 2", len(lf.Func.TypeParams))
	}
	if lf.Func.TypeParams[0].Constraint == nil || lf.Func.TypeParams[0].Constraint.String() != "Named" {
		t.Errorf("constraint = %v, want Named", lf.Func.TypeParams[0].Constraint)
	}
	if lf.Func.TypeParams[1].Constraint != nil {
		t.Errorf("second param is constrained, want unconstrained")
	}
}
func TestParseIntersectionType(t *testing.T) {
	ta := parseExpect1(t, "type Both = A & B").(*ast.TypeAliasStatement)
	if _, ok := ta.Target.(*ast.TypeIntersection); !ok {
		t.Fatalf("target = %T, want *ast.TypeIntersection", ta.Target)
	}
	if got := ta.Target.String(); got != "A & B" {
		t.Errorf("target = %q, want A & B", got)
	}
}
func TestIntersectionBindsTighterThanUnion(t *testing.T) {
	ta := parseExpect1(t, "type T = A & B | C").(*ast.TypeAliasStatement)
	if got := ta.Target.String(); got != "A & B | C" {
		t.Errorf("target = %q, want A & B | C", got)
	}
	if _, ok := ta.Target.(*ast.TypeUnion); !ok {
		t.Errorf("target = %T, want the union to be outermost", ta.Target)
	}
}
