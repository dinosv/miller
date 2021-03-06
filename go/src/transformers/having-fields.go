package transformers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"miller/src/cliutil"
	"miller/src/lib"
	"miller/src/transforming"
	"miller/src/types"
)

type tHavingFieldsCriterion int

const (
	havingFieldsCriterionUnspecified tHavingFieldsCriterion = iota
	havingFieldsAtLeast
	havingFieldsWhichAre
	havingFieldsAtMost
	havingAllFieldsMatching
	havingAnyFieldsMatching
	havingNoFieldsMatching
)

// ----------------------------------------------------------------
const verbNameHavingFields = "having-fields"

var HavingFieldsSetup = transforming.TransformerSetup{
	Verb:         verbNameHavingFields,
	UsageFunc:    transformerHavingFieldsUsage,
	ParseCLIFunc: transformerHavingFieldsParseCLI,

	IgnoresInput: false,
}

func transformerHavingFieldsUsage(
	o *os.File,
	doExit bool,
	exitCode int,
) {
	exeName := lib.MlrExeName()
	verb := verbNameHavingFields
	fmt.Fprintf(o, "Usage: %s %s [options]\n", lib.MlrExeName(), verbNameHavingFields)

	fmt.Fprintf(o, "Conditionally passes through records depending on each record's field names.\n")
	fmt.Fprintf(o, "Options:\n")
	fmt.Fprintf(o, "  --at-least      {comma-separated names}\n")
	fmt.Fprintf(o, "  --which-are     {comma-separated names}\n")
	fmt.Fprintf(o, "  --at-most       {comma-separated names}\n")
	fmt.Fprintf(o, "  --all-matching  {regular expression}\n")
	fmt.Fprintf(o, "  --any-matching  {regular expression}\n")
	fmt.Fprintf(o, "  --none-matching {regular expression}\n")
	fmt.Fprintf(o, "Examples:\n")
	fmt.Fprintf(o, "  %s %s --which-are amount,status,owner\n", exeName, verb)
	fmt.Fprintf(o, "  %s %s --any-matching 'sda[0-9]'\n", exeName, verb)
	fmt.Fprintf(o, "  %s %s --any-matching '\"sda[0-9]\"'\n", exeName, verb)
	fmt.Fprintf(o, "  %s %s --any-matching '\"sda[0-9]\"i' (this is case-insensitive)\n", exeName, verb)

	if doExit {
		os.Exit(exitCode)
	}
}

func transformerHavingFieldsParseCLI(
	pargi *int,
	argc int,
	args []string,
	_ *cliutil.TReaderOptions,
	__ *cliutil.TWriterOptions,
) transforming.IRecordTransformer {

	havingFieldsCriterion := havingFieldsCriterionUnspecified
	var fieldNames []string = nil
	regexString := ""

	// Skip the verb name from the current spot in the mlr command line
	argi := *pargi
	verb := args[argi]
	argi++

	for argi < argc /* variable increment: 1 or 2 depending on flag */ {
		opt := args[argi]
		if !strings.HasPrefix(opt, "-") {
			break // No more flag options to process
		}
		argi++

		if opt == "-h" || opt == "--help" {
			transformerHavingFieldsUsage(os.Stdout, true, 0)

		} else if opt == "--at-least" {
			havingFieldsCriterion = havingFieldsAtLeast
			fieldNames = cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			regexString = ""

		} else if opt == "--which-are" {
			havingFieldsCriterion = havingFieldsWhichAre
			fieldNames = cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			regexString = ""

		} else if opt == "--at-most" {
			havingFieldsCriterion = havingFieldsAtMost
			fieldNames = cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			regexString = ""

		} else if opt == "--all-matching" {
			havingFieldsCriterion = havingAllFieldsMatching
			regexString = cliutil.VerbGetStringArgOrDie(verb, opt, args, &argi, argc)
			fieldNames = nil

		} else if opt == "--any-matching" {
			havingFieldsCriterion = havingAnyFieldsMatching
			regexString = cliutil.VerbGetStringArgOrDie(verb, opt, args, &argi, argc)
			fieldNames = nil

		} else if opt == "--none-matching" {
			havingFieldsCriterion = havingNoFieldsMatching
			regexString = cliutil.VerbGetStringArgOrDie(verb, opt, args, &argi, argc)
			fieldNames = nil

		} else {
			transformerHavingFieldsUsage(os.Stderr, true, 1)
		}
	}

	if havingFieldsCriterion == havingFieldsCriterionUnspecified {
		transformerHavingFieldsUsage(os.Stderr, true, 1)
	}
	if fieldNames == nil && regexString == "" {
		transformerHavingFieldsUsage(os.Stderr, true, 1)
	}

	transformer, _ := NewTransformerHavingFields(
		havingFieldsCriterion,
		fieldNames,
		regexString,
	)

	*pargi = argi
	return transformer
}

// ----------------------------------------------------------------
type TransformerHavingFields struct {
	fieldNames    []string
	numFieldNames int
	fieldNameSet  map[string]bool

	regex *regexp.Regexp

	recordTransformerFunc transforming.RecordTransformerFunc
}

// ----------------------------------------------------------------
func NewTransformerHavingFields(
	havingFieldsCriterion tHavingFieldsCriterion,
	fieldNames []string,
	regexString string,
) (*TransformerHavingFields, error) {

	this := &TransformerHavingFields{}

	if fieldNames != nil {
		this.fieldNames = fieldNames
		this.numFieldNames = len(fieldNames)
		this.fieldNameSet = lib.StringListToSet(fieldNames)

		if havingFieldsCriterion == havingFieldsAtLeast {
			this.recordTransformerFunc = this.transformHavingFieldsAtLeast
		} else if havingFieldsCriterion == havingFieldsWhichAre {
			this.recordTransformerFunc = this.transformHavingFieldsWhichAre
		} else if havingFieldsCriterion == havingFieldsAtMost {
			this.recordTransformerFunc = this.transformHavingFieldsAtMost
		} else {
			lib.InternalCodingErrorIf(true)
		}

	} else {
		// Let them type in a.*b if they want, or "a.*b", or "a.*b"i.
		// Strip off the leading " and trailing " or "i.
		regex, err := lib.CompileMillerRegex(regexString)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"%s %s: cannot compile regex \"%s\"\n",
				lib.MlrExeName(),
				verbNameHavingFields,
				regexString,
			)
			os.Exit(1)
			// return nil, err
		}
		this.regex = regex

		if havingFieldsCriterion == havingAllFieldsMatching {
			this.recordTransformerFunc = this.transformHavingAllFieldsMatching
		} else if havingFieldsCriterion == havingAnyFieldsMatching {
			this.recordTransformerFunc = this.transformHavingAnyFieldsMatching
		} else if havingFieldsCriterion == havingNoFieldsMatching {
			this.recordTransformerFunc = this.transformHavingNoFieldsMatching
		} else {
			lib.InternalCodingErrorIf(true)
		}
	}

	return this, nil
}

// ----------------------------------------------------------------
func (this *TransformerHavingFields) Transform(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	this.recordTransformerFunc(inrecAndContext, outputChannel)
}

// ----------------------------------------------------------------
func (this *TransformerHavingFields) transformHavingFieldsAtLeast(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		numFound := 0
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if this.fieldNameSet[pe.Key] {
				numFound++
				if numFound == this.numFieldNames {
					outputChannel <- inrecAndContext
					return
				}
			}
		}

	} else {
		outputChannel <- inrecAndContext
	}
}

func (this *TransformerHavingFields) transformHavingFieldsWhichAre(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		if inrec.FieldCount != this.numFieldNames {
			return
		}
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if !this.fieldNameSet[pe.Key] {
				return
			}
		}
		outputChannel <- inrecAndContext
	} else {
		outputChannel <- inrecAndContext
	}
}

func (this *TransformerHavingFields) transformHavingFieldsAtMost(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if !this.fieldNameSet[pe.Key] {
				return
			}
		}
		outputChannel <- inrecAndContext
	} else {
		outputChannel <- inrecAndContext
	}
}

// ----------------------------------------------------------------
func (this *TransformerHavingFields) transformHavingAllFieldsMatching(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if !this.regex.MatchString(pe.Key) {
				return
			}
		}
		outputChannel <- inrecAndContext
	} else {
		outputChannel <- inrecAndContext
	}
}

func (this *TransformerHavingFields) transformHavingAnyFieldsMatching(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if this.regex.MatchString(pe.Key) {
				outputChannel <- inrecAndContext
				return
			}
		}
	} else {
		outputChannel <- inrecAndContext
	}
}

func (this *TransformerHavingFields) transformHavingNoFieldsMatching(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record
		for pe := inrec.Head; pe != nil; pe = pe.Next {
			if this.regex.MatchString(pe.Key) {
				return
			}
		}
		outputChannel <- inrecAndContext
	} else {
		outputChannel <- inrecAndContext
	}
}
