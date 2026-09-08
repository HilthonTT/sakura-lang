package bytecode

import (
	"github.com/hilthontt/luascript/internal/compiler/ast"
)

const (
	spreadGlobal     = "__tbl_spread"
	pushGlobal       = "__tbl_push"
	restArrayGlobal  = "__rest_array"
	restRecordGlobal = "__rest_record"
)
func tableHasSpread(t *ast.TableConstructor) bool {
	for _, f := range t.Fields {
		if f.IsSpread {
			return true
		}
	}
	return false
}
func (g *Generator) compileSpreadTable(is *InstructionSet, t *ast.TableConstructor) {
	line := t.Line()
	is.define(NewTable, line, len(t.Fields), 0)
	slot := g.current.locals.define("(table spread)")
	is.define(SetLocal, line, slot)

	for _, f := range t.Fields {
		switch {
		case f.IsSpread:
			g.emitMergeCall(is, line, spreadGlobal, slot, f.Value)
		case f.Key == nil:
			g.emitMergeCall(is, line, pushGlobal, slot, f.Value)
		case f.IsBracketed:
			is.define(GetLocal, line, slot)
			g.compileExpression(is, f.Key)
			g.compileExpression(is, f.Value)
			is.define(SetTable, line)
		default:
			ident, ok := f.Key.(*ast.Identifier)
			if !ok {
				panic("table record key must be *ast.Identifier")
			}
			is.define(GetLocal, line, slot)
			g.compileExpression(is, f.Value)
			is.define(SetField, line, ident.Name)
		}
	}
	is.define(GetLocal, line, slot)
}
func (g *Generator) emitMergeCall(is *InstructionSet, line int, global string, slot int, value ast.Expression) {
	is.define(GetGlobal, line, global)
	is.define(GetLocal, line, slot)
	g.compileExpression(is, value)
	is.define(Call, line, 2, 0)
}
func (g *Generator) compileLocalDestructure(is *InstructionSet, s *ast.LocalDestructureStatement) {
	line := s.Line()
	g.emitValue(is, s.Value)
	srcSlot := g.current.locals.define("(destructure src)")
	is.define(SetLocal, line, srcSlot)

	for _, b := range s.Binds {
		switch {
		case b.Rest && s.IsArray:
			is.define(GetGlobal, line, restArrayGlobal)
			is.define(GetLocal, line, srcSlot)
			is.define(LoadInt, line, int64(b.Index))
			is.define(Call, line, 2, 1)
		case b.Rest:
			is.define(GetGlobal, line, restRecordGlobal)
			is.define(GetLocal, line, srcSlot)
			g.emitKeyList(is, line, s.Binds)
			is.define(Call, line, 2, 1)
		case s.IsArray:
			is.define(GetLocal, line, srcSlot)
			is.define(LoadInt, line, int64(b.Index))
			is.define(GetTable, line)
		default:
			is.define(GetLocal, line, srcSlot)
			is.define(GetField, line, b.Key)
		}

		if b.Default != nil {
			g.emitNilDefault(is, line, b.Default)
		}

		if g.isReplTopLevel() {
			is.define(SetGlobal, line, b.Bind)
			continue
		}
		is.define(SetLocal, line, g.current.locals.define(b.Bind))
	}
}
func (g *Generator) emitKeyList(is *InstructionSet, line int, binds []ast.DestructureBind) {
	is.define(NewTable, line, len(binds), 0)
	idx := int64(0)
	for _, b := range binds {
		if b.Rest {
			continue
		}
		idx++
		is.define(Dup, line)
		is.define(LoadInt, line, idx)
		is.define(LoadString, line, b.Key)
		is.define(SetTable, line)
	}
}
func (g *Generator) emitNilDefault(is *InstructionSet, line int, def ast.Expression) {
	useDefault := &anchor{}
	jn := is.define(JumpIfNil, line, useDefault)
	g.current.recordPending(jn)
	end := &anchor{}
	j := is.define(Jump, line, end)
	g.current.recordPending(j)
	useDefault.line = is.count
	is.define(Pop, line, 1)
	g.compileExpression(is, def)
	end.line = is.count
}
