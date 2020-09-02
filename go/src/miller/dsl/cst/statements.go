package cst

import (
	"miller/dsl"
)

// ================================================================
// CST build/execute for statements: assignments, bare booleans,
// break/continue/return, etc.
// ================================================================

// ----------------------------------------------------------------
func NewSrecDirectFieldAssignment(
	astNode *dsl.ASTNode,
) (*SrecDirectFieldAssignment, error) {

	err := astNode.CheckArity(2)
	if err != nil {
		return nil, err
	}

	lhsASTNode := astNode.Children[0]
	rhsASTNode := astNode.Children[1]

	lhsFieldName := string(lhsASTNode.Token.Lit)
	rhs, err := NewEvaluable(rhsASTNode)
	if err != nil {
		return nil, err
	}

	return &SrecDirectFieldAssignment{
		lhsFieldName: lhsFieldName,
		rhs:          rhs,
	}, nil
}

func (this *SrecDirectFieldAssignment) Execute(state *State) {
	value := this.rhs.Evaluate(state)
	if !value.IsAbsent() {
		state.Inrec.Put(&this.lhsFieldName, &value)
	}
}
