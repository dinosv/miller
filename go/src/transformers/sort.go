// ================================================================
// OVERVIEW
//
// * Suppose we are sorting records lexically ascending on field "a" and then
//   numerically descending on field "x".
//
// * CLI syntax is "mlr sort -f a -nr x".
//
// * We first consume all input records and for each extract the string values
//   of fields a and x. For each uniq combination of a-value (e.g. "red",
//   "green", "blue") and x-value (e.g. "1", "1.0", "2.4") -- e.g.
//   pairs ["red","1"], and so on -- we keep a linked list of all the records
//   having those sort-key values, in the order encountered.
//
// * For each of those unique sort-key-value combinations, we also parse the
//   numerical fields at this point into an array of union-string-double.
//   E.g. the list ["red", "1.0"] maps to the array ["red", 1.0].
//
// * The pairing of parsed-value array the linked list of same-key-value records
//   is called a *bucket* or a *group*. E.g the records
//     {"a":"red","b":"circle","x":"1.0","y":"3.9"}
//     {"a":"red","b":"square","x":"1.0","z":"5.7", "q":"even"}
//   would both land in the ["red","1.0"] group.
//
// * Groups are retained in a hash map: the key is the string-list of the form
//   ["red","1.0"] and the value is the pairing of parsed-value array ["red",1.0]
//   and linked list of records.
//
// * Once all the input records are ingested into this hash map, we copy the
//   group-pointers into an array and sort it: this being the pairing of
//   parsed-value array and linked list of records. The comparator callback for
//   the sort walks through the parsed-value arrays one slot at a time,
//   looking at the first difference, e.g. if one has "a"="red" and the other
//   has "a"="blue". If the first field matches then the sort moves to the
//   second field, and so on.
//
// * Note in particular that string keys ["a":"red","x":"1"] and
//   ["a":"red","x":"1.0"] map to different groups, but will sort equally.
//
// ================================================================

package transformers

import (
	"container/list"
	"fmt"
	"os"
	"sort"
	"strings"

	"miller/src/cliutil"
	"miller/src/lib"
	"miller/src/transforming"
	"miller/src/types"
)

// ----------------------------------------------------------------
const verbNameSort = "sort"

var SortSetup = transforming.TransformerSetup{
	Verb:         verbNameSort,
	UsageFunc:    transformerSortUsage,
	ParseCLIFunc: transformerSortParseCLI,
	IgnoresInput: false,
}

func transformerSortUsage(
	o *os.File,
	doExit bool,
	exitCode int,
) {
	fmt.Fprintf(o, "Usage: %s %s {flags}\n", lib.MlrExeName(), verbNameSort)
	fmt.Fprintf(o, "Sorts records primarily by the first specified field, secondarily by the second\n")
	fmt.Fprintf(o, "field, and so on.  (Any records not having all specified sort keys will appear\n")
	fmt.Fprintf(o, "at the end of the output, in the order they were encountered, regardless of the\n")
	fmt.Fprintf(o, "specified sort order.) The sort is stable: records that compare equal will sort\n")
	fmt.Fprintf(o, "in the order they were encountered in the input record stream.\n")
	fmt.Fprintf(o, "\n")
	fmt.Fprintf(o, "Flags:\n")
	fmt.Fprintf(o, "-f  {comma-separated field names}  Lexical ascending\n")
	fmt.Fprintf(o, "-n  {comma-separated field names}  Numerical ascending; nulls sort last\n")
	fmt.Fprintf(o, "-nf {comma-separated field names}  Same as -n\n")
	fmt.Fprintf(o, "-r  {comma-separated field names}  Lexical descending\n")
	fmt.Fprintf(o, "-nr {comma-separated field names}  Numerical descending; nulls sort first\n")
	fmt.Fprintf(o, "-h|--help Show this message.\n")
	fmt.Fprintf(o, "\n")
	fmt.Fprintf(o, "Example:\n")
	fmt.Fprintf(o, "  %s %s -f a,b -nr x,y,z\n", lib.MlrExeName(), verbNameSort)
	fmt.Fprintf(o, "which is the same as:\n")
	fmt.Fprintf(o, "  %s %s -f a -f b -nr x -nr y -nr z\n", lib.MlrExeName(), verbNameSort)

	if doExit {
		os.Exit(exitCode)
	}
}

func transformerSortParseCLI(
	pargi *int,
	argc int,
	args []string,
	_ *cliutil.TReaderOptions,
	__ *cliutil.TWriterOptions,
) transforming.IRecordTransformer {

	// Skip the verb name from the current spot in the mlr command line
	argi := *pargi
	verb := args[argi]
	argi++

	groupByFieldNames := make([]string, 0)
	comparatorFuncs := make([]types.ComparatorFunc, 0)

	for argi < argc /* variable increment: 1 or 2 depending on flag */ {
		opt := args[argi]
		if !strings.HasPrefix(opt, "-") {
			break // No more flag options to process
		}
		argi++

		if opt == "-h" || opt == "--help" {
			transformerSortUsage(os.Stdout, true, 0)

		} else if opt == "-f" {
			subList := cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			for _, item := range subList {
				groupByFieldNames = append(groupByFieldNames, item)
				comparatorFuncs = append(comparatorFuncs, types.LexicalAscendingComparator)
			}

		} else if opt == "-r" {
			subList := cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			for _, item := range subList {
				groupByFieldNames = append(groupByFieldNames, item)
				comparatorFuncs = append(comparatorFuncs, types.LexicalDescendingComparator)
			}

		} else if opt == "-n" {
			subList := cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			for _, item := range subList {
				groupByFieldNames = append(groupByFieldNames, item)
				comparatorFuncs = append(comparatorFuncs, types.NumericAscendingComparator)
			}

		} else if opt == "-nf" {
			subList := cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			for _, item := range subList {
				groupByFieldNames = append(groupByFieldNames, item)
				comparatorFuncs = append(comparatorFuncs, types.NumericAscendingComparator)
			}

		} else if opt == "-nr" {
			subList := cliutil.VerbGetStringArrayArgOrDie(verb, opt, args, &argi, argc)
			for _, item := range subList {
				groupByFieldNames = append(groupByFieldNames, item)
				comparatorFuncs = append(comparatorFuncs, types.NumericDescendingComparator)
			}

		} else {
			transformerSortUsage(os.Stderr, true, 1)
		}
	}

	if len(groupByFieldNames) == 0 {
		transformerSortUsage(os.Stderr, true, 1)
	}

	transformer, _ := NewTransformerSort(
		groupByFieldNames,
		comparatorFuncs,
	)

	*pargi = argi
	return transformer
}

// ----------------------------------------------------------------
// Example:
// * mlr sort -f a -n i
// * group-by field-name list is "a,i"
// * input record 'a=pan,b=pan,i=1,x=0.3467,y=0.7268'
//   o values at a,i are "pan",1
//   o grouping key for the ordered map from string to record-group is the string "pan,1"
//   o we also need a map from "pan,1" to the array of mlrvals ["pan", 1].
// * next input record 'a=eks,b=pan,i=2,x=0.7586,y=0.5221'
//   o values at a,i are "eks",2
//   o grouping key for the ordered map from string to record-group is the string "eks,2"
//   o we also need a map from "eks,2" to the array of mlrvals ["eks", 2].
// * what gets sorted are the group-heading arrays of mlrvals:
//   o make an array [ ("pan,1", ["pan", 1]), ("eks,2", ["eks", 2])
//   o sort that
// * output is simply for each slot in the array, emit each record in the group

type TransformerSort struct {
	// -- Input
	groupByFieldNames []string
	comparatorFuncs   []types.ComparatorFunc

	// -- State
	// Map from string to *list.List:
	recordListsByGroup *lib.OrderedMap
	// Map from string to []*lib.Mlrval:
	groupHeads *lib.OrderedMap
	spillGroup *list.List // e.g. sort by field "a" -- this is for records lacking a field named "a"
}

func NewTransformerSort(
	groupByFieldNames []string,
	comparatorFuncs []types.ComparatorFunc,
) (*TransformerSort, error) {

	this := &TransformerSort{
		groupByFieldNames: groupByFieldNames,
		comparatorFuncs:   comparatorFuncs,

		recordListsByGroup: lib.NewOrderedMap(),
		groupHeads:         lib.NewOrderedMap(),
		spillGroup:         list.New(),
	}

	return this, nil
}

// ----------------------------------------------------------------
type GroupingKeysAndMlrvals struct {
	groupingKey string
	mlrvals     []*types.Mlrval
}

func (this *TransformerSort) Transform(
	inrecAndContext *types.RecordAndContext,
	outputChannel chan<- *types.RecordAndContext,
) {
	if !inrecAndContext.EndOfStream {
		inrec := inrecAndContext.Record

		groupingKey, selectedValues, ok := inrec.GetSelectedValuesAndJoined(
			this.groupByFieldNames,
		)
		if !ok {
			this.spillGroup.PushBack(inrecAndContext)
			return
		}

		recordListForGroup := this.recordListsByGroup.Get(groupingKey)
		if recordListForGroup == nil {
			recordListForGroup = list.New()
			this.recordListsByGroup.Put(groupingKey, recordListForGroup)
			this.groupHeads.Put(groupingKey, selectedValues)
		}

		recordListForGroup.(*list.List).PushBack(inrecAndContext)

	} else { // End of record stream

		// At this point, in the above example, groupHeads is:
		//
		// {
		//   "pan,1" : ["pan", 1],
		//   "eks,2" : ["eks", 2]
		// }
		//
		// We need to make an array like
		//
		// [
		//   [ "pan,1", ["pan', 1],
		//   [ "eks,2", ["eks', 2]
		// ]

		groupingKeysAndMlrvals := groupHeadsToArray(this.groupHeads)

		// Go sort API: for ascending sort, return true if element i < element j.
		sort.Slice(groupingKeysAndMlrvals, func(i, j int) bool {
			for k, comparator := range this.comparatorFuncs {
				result := comparator(
					groupingKeysAndMlrvals[i].mlrvals[k],
					groupingKeysAndMlrvals[j].mlrvals[k],
				)
				if result < 0 {
					return true
				} else if result > 0 {
					return false
				}
			}
			return false
		})

		// Now output the groups
		for _, groupingKeyAndMlrvals := range groupingKeysAndMlrvals {
			iRecordsInGroup := this.recordListsByGroup.Get(groupingKeyAndMlrvals.groupingKey)
			recordsInGroup := iRecordsInGroup.(*list.List)
			for iRecord := recordsInGroup.Front(); iRecord != nil; iRecord = iRecord.Next() {
				outputChannel <- iRecord.Value.(*types.RecordAndContext)
			}
		}

		for iRecord := this.spillGroup.Front(); iRecord != nil; iRecord = iRecord.Next() {
			outputChannel <- iRecord.Value.(*types.RecordAndContext)
		}

		outputChannel <- inrecAndContext // end-of-stream marker
	}
}

func groupHeadsToArray(groupHeads *lib.OrderedMap) []GroupingKeysAndMlrvals {
	retval := make([]GroupingKeysAndMlrvals, groupHeads.FieldCount)

	i := 0
	for entry := groupHeads.Head; entry != nil; entry = entry.Next {
		retval[i] = GroupingKeysAndMlrvals{
			groupingKey: entry.Key,
			mlrvals:     entry.Value.([]*types.Mlrval),
		}
		i++
	}

	return retval
}
