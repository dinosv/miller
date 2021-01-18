// ================================================================
// This handles dump and edump statements.
// See print.go for comments; this is similar.
//
// Differences between print and dump:
//
// * 'print $x' and 'dump $x' are the same.
//
// * 'print' and 'dump' with no specific value: print outputs a newline; dump
//   outputs a JSON representation of all out-of-stream variables.
//
// * 'print $x,$y,$z' prints all items on one line; 'dump $x,$y,$z' prints each on
//   its own line.
// ================================================================

package cst

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"miller/dsl"
	"miller/lib"
	"miller/types"
)

// ================================================================
type tDumpToRedirectFunc func(
	outputString string,
	state *State,
) error

type DumpStatementNode struct {
	expressionEvaluables      []IEvaluable
	dumpToRedirectFunc        tDumpToRedirectFunc
	redirectorTargetEvaluable IEvaluable           // for file/pipe targets
	outputHandlerManager      OutputHandlerManager // for file/pipe targets
}

// ----------------------------------------------------------------
func (this *RootNode) BuildDumpStatementNode(astNode *dsl.ASTNode) (IExecutable, error) {
	lib.InternalCodingErrorIf(astNode.Type != dsl.NodeTypeDumpStatement)
	return this.buildDumpxStatementNode(
		astNode,
		os.Stdout,
	)
}

func (this *RootNode) BuildEdumpStatementNode(astNode *dsl.ASTNode) (IExecutable, error) {
	lib.InternalCodingErrorIf(astNode.Type != dsl.NodeTypeEdumpStatement)
	return this.buildDumpxStatementNode(
		astNode,
		os.Stderr,
	)
}

// ----------------------------------------------------------------
// Common code for building dump/edump nodes

func (this *RootNode) buildDumpxStatementNode(
	astNode *dsl.ASTNode,
	defaultOutputStream *os.File,
) (IExecutable, error) {
	lib.InternalCodingErrorIf(len(astNode.Children) != 2)
	expressionsNode := astNode.Children[0]
	redirectorNode := astNode.Children[1]

	//  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	// Things to be dumped, e.g. $a and $b in 'dump > "foo.dat", $a, $b'.

	var expressionEvaluables []IEvaluable = nil

	if expressionsNode.Type == dsl.NodeTypeNoOp {
		// Just 'dump' without 'dump $something'
		expressionEvaluables = make([]IEvaluable, 1)
		expressionEvaluable := this.BuildFullOosvarRvalueNode()
		expressionEvaluables[0] = expressionEvaluable
	} else if expressionsNode.Type == dsl.NodeTypeFunctionCallsite {
		expressionEvaluables = make([]IEvaluable, len(expressionsNode.Children))
		for i, childNode := range expressionsNode.Children {
			expressionEvaluable, err := this.BuildEvaluableNode(childNode)
			if err != nil {
				return nil, err
			}
			expressionEvaluables[i] = expressionEvaluable
		}
	} else {
		lib.InternalCodingErrorIf(true)
	}

	//  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	// Redirection targets (the thing after > >> |, if any).

	retval := &DumpStatementNode{
		expressionEvaluables:      expressionEvaluables,
		dumpToRedirectFunc:        nil,
		redirectorTargetEvaluable: nil,
		outputHandlerManager:      nil,
	}

	if redirectorNode.Type == dsl.NodeTypeNoOp {
		// No > >> or | was provided.
		if defaultOutputStream == os.Stdout {
			retval.dumpToRedirectFunc = retval.dumpToStdout
		} else if defaultOutputStream == os.Stderr {
			retval.dumpToRedirectFunc = retval.dumpToStderr
		} else {
			lib.InternalCodingErrorIf(true)
		}
	} else {
		// There is > >> or | provided.
		lib.InternalCodingErrorIf(redirectorNode.Children == nil)
		lib.InternalCodingErrorIf(len(redirectorNode.Children) != 1)
		redirectorTargetNode := redirectorNode.Children[0]
		var err error = nil

		if redirectorTargetNode.Type == dsl.NodeTypeRedirectTargetStdout {
			retval.dumpToRedirectFunc = retval.dumpToStdout
		} else if redirectorTargetNode.Type == dsl.NodeTypeRedirectTargetStderr {
			retval.dumpToRedirectFunc = retval.dumpToStderr
		} else {
			retval.dumpToRedirectFunc = retval.dumpToFileOrPipe

			retval.redirectorTargetEvaluable, err = this.BuildEvaluableNode(redirectorTargetNode)
			if err != nil {
				return nil, err
			}

			if redirectorNode.Type == dsl.NodeTypeRedirectWrite {
				retval.outputHandlerManager = NewFileWritetHandlerManager()
			} else if redirectorNode.Type == dsl.NodeTypeRedirectAppend {
				retval.outputHandlerManager = NewFileAppendHandlerManager()
			} else if redirectorNode.Type == dsl.NodeTypeRedirectPipe {
				retval.outputHandlerManager = NewPipeWriteHandlerManager()
			} else {
				return nil, errors.New(
					fmt.Sprintf(
						"%s: unhandled redirector node type %s.",
						os.Args[0], string(redirectorNode.Type),
					),
				)
			}
		}
	}

	// TODO: root node register outputHandlerManager to add to close-handles list

	return retval, nil
}

// ----------------------------------------------------------------
func (this *DumpStatementNode) Execute(state *State) (*BlockExitPayload, error) {
	// 5x faster than fmt.Dump() separately: note that os.Stdout is
	// non-buffered in Go whereas stdout is buffered in C.
	//
	// Minus: we need to do our own buffering for performance.
	//
	// Plus: we never have to worry about forgetting to do fflush(). :)
	var buffer bytes.Buffer

	for _, expressionEvaluable := range this.expressionEvaluables {
		evaluation := expressionEvaluable.Evaluate(state)
		if !evaluation.IsAbsent() {
			s := evaluation.String()
			buffer.WriteString(s)
			// Dump of 1 is "1", needs newline; similar for other atomics.
			// Dump of JSON objects already ends in newline and doesn't need
			// another.
			if !strings.HasSuffix(s, "\n") {
				buffer.WriteString("\n")
			}
		}
	}
	outputString := buffer.String()
	this.dumpToRedirectFunc(outputString, state)
	return nil, nil
}

// ----------------------------------------------------------------
type FullOosvarDumpNode struct {
}

func (this *RootNode) BuildFullOosvarDumpNode() *FullOosvarDumpNode {
	return &FullOosvarDumpNode{}
}
func (this *FullOosvarDumpNode) Evaluate(state *State) types.Mlrval {
	return types.MlrvalFromString(state.Oosvars.String())
}

// ----------------------------------------------------------------
func (this *DumpStatementNode) dumpToStdout(
	outputString string,
	state *State,
) error {
	// Insert the string into the record-output stream, so that goroutine can
	// print it, resulting in deterministic output-ordering.
	state.OutputChannel <- types.NewOutputString(outputString, state.Context)
	return nil
}

// ----------------------------------------------------------------
func (this *DumpStatementNode) dumpToStderr(
	outputString string,
	state *State,
) error {
	fmt.Fprintf(os.Stderr, outputString)
	return nil
}

// ----------------------------------------------------------------
func (this *DumpStatementNode) dumpToFileOrPipe(
	outputString string,
	state *State,
) error {
	redirectorTarget := this.redirectorTargetEvaluable.Evaluate(state)
	if !redirectorTarget.IsString() {
		return errors.New(
			fmt.Sprintf(
				"%s: output redirection yielded %s, not string.",
				os.Args[0], redirectorTarget.GetTypeName(),
			),
		)
	}
	outputFileName := redirectorTarget.String()

	this.outputHandlerManager.Print(outputString, outputFileName)
	return nil
}
