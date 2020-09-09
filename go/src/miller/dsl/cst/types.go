package cst

import (
	"miller/lib"
)

// ================================================================
// Main type definitions for CST build/execute
// ================================================================

// ----------------------------------------------------------------
// There are three CST roots: begin-block, body-block, and end-block.
//
// Next-level items are:
// * srec assignments
// * oosvar assignments
// * localvar assignments
// * emit et al.
// * bare-boolean
// * break/continue/return
// * statement block (if-body, for-body, etc)
// ----------------------------------------------------------------

// ----------------------------------------------------------------
// When we do mlr put '...DSL expression here...', this state is what is needed
// to execute the expression. That includes the current record, AWK-like variables
// such as FILENAME and NR, and out-of-stream variables.
type State struct {
	Inrec   *lib.Mlrmap
	Context *lib.Context
	Oosvars *lib.Mlrmap
	// TODO: stack frames will go into individual statement-block nodes
}

func NewEmptyState() *State {
	oosvars := lib.NewMlrmap()
	return &State{
		Inrec:   nil,
		Context: nil,
		Oosvars: oosvars,
	}
}

func (this *State) Update(
	inrec *lib.Mlrmap,
	context *lib.Context,
) {
	this.Inrec = inrec
	this.Context = context
}

// ----------------------------------------------------------------
type Root struct {
	// Statements/blocks
	executables []IExecutable
}

// ----------------------------------------------------------------
// This is for all statements and statemnt blocks within the CST.
type IExecutable interface {
	Execute(state *State) error
}

// ----------------------------------------------------------------
type StatementBlockNode struct {
	// TODO: list of statement, once begin/end/main are in the DSL
}

// ================================================================
// This is for any left-hand side (LHS or Lvalue) of an assignment statement.
type IAssignable interface {
	Assign(rvalue *lib.Mlrval, state *State) error

	// 'foo = "bar"' or 'foo[3]["abc"] = "bar"'
	// For non-indexed assignment, which is the normal case, indices can be
	// zero-length or nil.
	AssignIndexed(rvalue *lib.Mlrval, indices []*lib.Mlrval, state *State) error
}

// ================================================================
// This is for any right-hand side (RHS or Rvalue) of an assignment statement.
// Also, for computed field names on the left-hand side, like '$a . $b' in mlr
// put '$[$a . $b]' = $x + $y'.
type IEvaluable interface {
	Evaluate(state *State) lib.Mlrval
}