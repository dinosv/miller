package transforming

import (
	"os"

	"miller/src/cliutil"
	"miller/src/types"
)

type IRecordTransformer interface {
	Transform(
		inrecAndContext *types.RecordAndContext,
		outputChannel chan<- *types.RecordAndContext,
	)
}

type RecordTransformerFunc func(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
)

type TransformerUsageFunc func(
	ostream *os.File,
	doExit bool,
	exitCode int,
)

type TransformerParseCLIFunc func(
	pargi *int,
	argc int,
	args []string,
	readerOptions *cliutil.TReaderOptions,
	writerOptions *cliutil.TWriterOptions,
) IRecordTransformer

type TransformerSetup struct {
	Verb         string
	UsageFunc    TransformerUsageFunc
	ParseCLIFunc TransformerParseCLIFunc
	IgnoresInput bool
}
