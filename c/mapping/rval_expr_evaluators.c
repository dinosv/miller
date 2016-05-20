#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <ctype.h> // for tolower(), toupper()
#include "lib/mlr_globals.h"
#include "lib/mlrutil.h"
#include "lib/mlrregex.h"
#include "lib/mtrand.h"
#include "mapping/mapper.h"
#include "mapping/rval_evaluators.h"

// ================================================================
// See comments in rval_evaluators.h
// ================================================================

static rval_evaluator_t* rval_evaluator_alloc_from_ast_aux(mlr_dsl_ast_node_t* pnode,
	int type_inferencing, function_lookup_t* fcn_lookup_table);

// ================================================================
rval_evaluator_t* rval_evaluator_alloc_from_ast(mlr_dsl_ast_node_t* pnode, int type_inferencing) {
	return rval_evaluator_alloc_from_ast_aux(pnode, type_inferencing, FUNCTION_LOOKUP_TABLE);
}

static rval_evaluator_t* rval_evaluator_alloc_from_ast_aux(mlr_dsl_ast_node_t* pnode,
	int type_inferencing, function_lookup_t* fcn_lookup_table)
{
	if (pnode->pchildren == NULL) { // leaf node
		if (pnode->type == MD_AST_NODE_TYPE_FIELD_NAME) {
			return rval_evaluator_alloc_from_field_name(pnode->text, type_inferencing);
		} else if (pnode->type == MD_AST_NODE_TYPE_STRNUM_LITERAL) {
			return rval_evaluator_alloc_from_strnum_literal(pnode->text, type_inferencing);
		} else if (pnode->type == MD_AST_NODE_TYPE_BOOLEAN_LITERAL) {
			return rval_evaluator_alloc_from_boolean_literal(pnode->text);
		} else if (pnode->type == MD_AST_NODE_TYPE_REGEXI) {
			return rval_evaluator_alloc_from_strnum_literal(pnode->text, type_inferencing);
		} else if (pnode->type == MD_AST_NODE_TYPE_CONTEXT_VARIABLE) {
			return rval_evaluator_alloc_from_context_variable(pnode->text);
		} else if (pnode->type == MD_AST_NODE_TYPE_BOUND_VARIABLE) {
			return rval_evaluator_alloc_from_bound_variable(pnode->text);
		} else {
			fprintf(stderr, "%s: internal coding error detected in file %s at line %d.\n",
				MLR_GLOBALS.argv0, __FILE__, __LINE__);
			exit(1);
		}

	} else if (pnode->type == MD_AST_NODE_TYPE_INDIRECT_FIELD_NAME) { // xxx rid of
		return rval_evaluator_alloc_from_indirect_field_name(pnode->pchildren->phead->pvvalue, type_inferencing);

	} else if (pnode->type == MD_AST_NODE_TYPE_OOSVAR_KEYLIST) {
		return rval_evaluator_alloc_from_oosvar_keylist(pnode, type_inferencing);

	} else if (pnode->type == MD_AST_NODE_TYPE_ENV) {
		return rval_evaluator_alloc_from_environment(pnode, type_inferencing);

	} else { // operator/function
		if ((pnode->type != MD_AST_NODE_TYPE_NON_SIGIL_NAME)
		&& (pnode->type != MD_AST_NODE_TYPE_OPERATOR)) {
			fprintf(stderr, "%s: internal coding error detected in file %s at line %d (node type %s).\n",
				MLR_GLOBALS.argv0, __FILE__, __LINE__, mlr_dsl_ast_node_describe_type(pnode->type));
			exit(1);
		}
		char* func_name = pnode->text;

		int user_provided_arity = pnode->pchildren->length;

		check_arity_with_report(fcn_lookup_table, func_name, user_provided_arity);

		rval_evaluator_t* pevaluator = NULL;
		if (user_provided_arity == 0) {
			pevaluator = rval_evaluator_alloc_from_zary_func_name(func_name);
		} else if (user_provided_arity == 1) {
			mlr_dsl_ast_node_t* parg1_node = pnode->pchildren->phead->pvvalue;
			rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing, fcn_lookup_table);
			pevaluator = rval_evaluator_alloc_from_unary_func_name(func_name, parg1);
		} else if (user_provided_arity == 2) {
			mlr_dsl_ast_node_t* parg1_node = pnode->pchildren->phead->pvvalue;
			mlr_dsl_ast_node_t* parg2_node = pnode->pchildren->phead->pnext->pvvalue;
			int type2 = parg2_node->type;

			if ((streq(func_name, "=~") || streq(func_name, "!=~")) && type2 == MD_AST_NODE_TYPE_STRNUM_LITERAL) {
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_binary_regex_arg2_func_name(func_name,
					parg1, parg2_node->text, FALSE);
			} else if ((streq(func_name, "=~") || streq(func_name, "!=~")) && type2 == MD_AST_NODE_TYPE_REGEXI) {
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_binary_regex_arg2_func_name(func_name, parg1, parg2_node->text,
					TYPE_INFER_STRING_FLOAT_INT);
			} else {
				// regexes can still be applied here, e.g. if the 2nd argument is a non-terminal AST: however
				// the regexes will be compiled record-by-record rather than once at alloc time, which will
				// be slower.
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				rval_evaluator_t* parg2 = rval_evaluator_alloc_from_ast_aux(parg2_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_binary_func_name(func_name, parg1, parg2);
			}

		} else if (user_provided_arity == 3) {
			mlr_dsl_ast_node_t* parg1_node = pnode->pchildren->phead->pvvalue;
			mlr_dsl_ast_node_t* parg2_node = pnode->pchildren->phead->pnext->pvvalue;
			mlr_dsl_ast_node_t* parg3_node = pnode->pchildren->phead->pnext->pnext->pvvalue;
			int type2 = parg2_node->type;

			if ((streq(func_name, "sub") || streq(func_name, "gsub")) && type2 == MD_AST_NODE_TYPE_STRNUM_LITERAL) {
				// sub/gsub-regex special case:
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				rval_evaluator_t* parg3 = rval_evaluator_alloc_from_ast_aux(parg3_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_ternary_regex_arg2_func_name(func_name, parg1, parg2_node->text,
					FALSE, parg3);

			} else if ((streq(func_name, "sub") || streq(func_name, "gsub")) && type2 == MD_AST_NODE_TYPE_REGEXI) {
				// sub/gsub-regex special case:
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				rval_evaluator_t* parg3 = rval_evaluator_alloc_from_ast_aux(parg3_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_ternary_regex_arg2_func_name(func_name, parg1, parg2_node->text,
					TYPE_INFER_STRING_FLOAT_INT, parg3);

			} else {
				// regexes can still be applied here, e.g. if the 2nd argument is a non-terminal AST: however
				// the regexes will be compiled record-by-record rather than once at alloc time, which will
				// be slower.
				rval_evaluator_t* parg1 = rval_evaluator_alloc_from_ast_aux(parg1_node, type_inferencing,
					fcn_lookup_table);
				rval_evaluator_t* parg2 = rval_evaluator_alloc_from_ast_aux(parg2_node, type_inferencing,
					fcn_lookup_table);
				rval_evaluator_t* parg3 = rval_evaluator_alloc_from_ast_aux(parg3_node, type_inferencing,
					fcn_lookup_table);
				pevaluator = rval_evaluator_alloc_from_ternary_func_name(func_name, parg1, parg2, parg3);
			}
		} else {
			fprintf(stderr, "Miller: internal coding error:  arity for function name \"%s\" misdetected.\n",
				func_name);
			exit(1);
		}
		if (pevaluator == NULL) {
			fprintf(stderr, "Miller: unrecognized function name \"%s\".\n", func_name);
			exit(1);
		}
		return pevaluator;
	}
}

// ================================================================
typedef struct _rval_evaluator_field_name_state_t {
	char* field_name;
} rval_evaluator_field_name_state_t;

static mv_t rval_evaluator_field_name_func_string_only(void* pvstate, variables_t* pvars) {
	rval_evaluator_field_name_state_t* pstate = pvstate;
	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, pstate->field_name);
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		return mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, pstate->field_name);
		if (string == NULL) {
			return mv_absent();
		} else if (*string == 0) {
			return mv_empty();
		} else {
			// string points into lrec memory and is valid as long as the lrec is.
			return mv_from_string_no_free(string);
		}
	}
}

static mv_t rval_evaluator_field_name_func_string_float(void* pvstate, variables_t* pvars) {
	rval_evaluator_field_name_state_t* pstate = pvstate;
	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, pstate->field_name);
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		return mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, pstate->field_name);
		if (string == NULL) {
			return mv_absent();
		} else if (*string == 0) {
			return mv_empty();
		} else {
			double fltv;
			if (mlr_try_float_from_string(string, &fltv)) {
				return mv_from_float(fltv);
			} else {
				// string points into lrec memory and is valid as long as the lrec is.
				return mv_from_string_no_free(string);
			}
		}
	}
}

static mv_t rval_evaluator_field_name_func_string_float_int(void* pvstate, variables_t* pvars) {
	rval_evaluator_field_name_state_t* pstate = pvstate;
	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, pstate->field_name);
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		return mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, pstate->field_name);
		if (string == NULL) {
			return mv_absent();
		} else if (*string == 0) {
			return mv_empty();
		} else {
			long long intv;
			double fltv;
			if (mlr_try_int_from_string(string, &intv)) {
				return mv_from_int(intv);
			} else if (mlr_try_float_from_string(string, &fltv)) {
				return mv_from_float(fltv);
			} else {
				// string points into AST memory and is valid as long as the AST is.
				return mv_from_string_no_free(string);
			}
		}
	}
}
static void rval_evaluator_field_name_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_field_name_state_t* pstate = pevaluator->pvstate;
	free(pstate->field_name);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_field_name(char* field_name, int type_inferencing) {
	rval_evaluator_field_name_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_field_name_state_t));
	pstate->field_name = mlr_strdup_or_die(field_name);

	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = pstate;
	pevaluator->pprocess_func = NULL;
	switch (type_inferencing) {
	case TYPE_INFER_STRING_ONLY:
		pevaluator->pprocess_func = rval_evaluator_field_name_func_string_only;
		break;
	case TYPE_INFER_STRING_FLOAT:
		pevaluator->pprocess_func = rval_evaluator_field_name_func_string_float;
		break;
	case TYPE_INFER_STRING_FLOAT_INT:
		pevaluator->pprocess_func = rval_evaluator_field_name_func_string_float_int;
		break;
	default:
		fprintf(stderr, "%s: internal coding error detected in file %s at line %d.\n",
			MLR_GLOBALS.argv0, __FILE__, __LINE__);
		exit(1);
		break;
	}
	pevaluator->pfree_func = rval_evaluator_field_name_free;

	return pevaluator;
}

// ================================================================
// xxx reduce code-dup x 6

typedef struct _rval_evaluator_indirect_field_name_state_t {
	rval_evaluator_t* pname_evaluator;
} rval_evaluator_indirect_field_name_state_t;

static mv_t rval_evaluator_indirect_field_name_func_string_only(void* pvstate, variables_t* pvars) {
	rval_evaluator_indirect_field_name_state_t* pstate = pvstate;

	mv_t mvname = pstate->pname_evaluator->pprocess_func(pstate->pname_evaluator->pvstate, pvars);
	if (mv_is_null(&mvname)) {
		return mv_absent();
	}
	char free_flags = NO_FREE;
	char* indirect_field_name = mv_maybe_alloc_format_val(&mvname, &free_flags);

	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, indirect_field_name);
	mv_t rv;
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		rv = mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, indirect_field_name);
		if (string == NULL) {
			rv = mv_absent();
		} else if (*string == 0) {
			rv = mv_empty();
		} else {
			// string points into lrec memory and is valid as long as the lrec is.
			rv = mv_from_string_no_free(string);
		}
	}

	if (free_flags & FREE_ENTRY_VALUE)
		free(indirect_field_name);
	return rv;
}

static mv_t rval_evaluator_indirect_field_name_func_string_float(void* pvstate, variables_t* pvars) {
	rval_evaluator_indirect_field_name_state_t* pstate = pvstate;

	mv_t mvname = pstate->pname_evaluator->pprocess_func(pstate->pname_evaluator->pvstate, pvars);
	if (mv_is_null(&mvname)) {
		return mv_absent();
	}
	char free_flags = NO_FREE;
	char* indirect_field_name = mv_maybe_alloc_format_val(&mvname, &free_flags);

	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, indirect_field_name);
	mv_t rv;
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		rv = mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, indirect_field_name);
		if (string == NULL) {
			rv = mv_absent();
		} else if (*string == 0) {
			rv = mv_empty();
		} else {
			double fltv;
			if (mlr_try_float_from_string(string, &fltv)) {
				rv = mv_from_float(fltv);
			} else {
				// string points into lrec memory and is valid as long as the lrec is.
				rv = mv_from_string_no_free(string);
			}
		}
	}

	if (free_flags & FREE_ENTRY_VALUE)
		free(indirect_field_name);
	return rv;
}

static mv_t rval_evaluator_indirect_field_name_func_string_float_int(void* pvstate, variables_t* pvars) {
	rval_evaluator_indirect_field_name_state_t* pstate = pvstate;

	mv_t mvname = pstate->pname_evaluator->pprocess_func(pstate->pname_evaluator->pvstate, pvars);
	if (mv_is_null(&mvname)) {
		return mv_absent();
	}
	char free_flags = NO_FREE;
	char* indirect_field_name = mv_maybe_alloc_format_val(&mvname, &free_flags);

	// See comments in rval_evaluator.h and mapper_put.c regarding the typed-overlay map.
	mv_t* poverlay = lhmsv_get(pvars->ptyped_overlay, indirect_field_name);
	mv_t rv;
	if (poverlay != NULL) {
		// The lrec-evaluator logic will free its inputs and allocate new outputs, so we must copy
		// a value here to feed into that. Otherwise the typed-overlay map would have its contents
		// freed out from underneath it by the evaluator functions.
		rv = mv_copy(poverlay);
	} else {
		char* string = lrec_get(pvars->pinrec, indirect_field_name);
		if (string == NULL) {
			rv = mv_absent();
		} else if (*string == 0) {
			rv = mv_empty();
		} else {
			long long intv;
			double fltv;
			if (mlr_try_int_from_string(string, &intv)) {
				rv = mv_from_int(intv);
			} else if (mlr_try_float_from_string(string, &fltv)) {
				rv = mv_from_float(fltv);
			} else {
				// string points into AST memory and is valid as long as the AST is.
				rv = mv_from_string_no_free(string);
			}
		}
	}

	if (free_flags & FREE_ENTRY_VALUE)
		free(indirect_field_name);
	return rv;
}

static void rval_evaluator_indirect_field_name_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_indirect_field_name_state_t* pstate = pevaluator->pvstate;
	pstate->pname_evaluator->pfree_func(pstate->pname_evaluator);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_indirect_field_name(mlr_dsl_ast_node_t* pnamenode, int type_inferencing) {
	rval_evaluator_indirect_field_name_state_t* pstate = mlr_malloc_or_die(
		sizeof(rval_evaluator_indirect_field_name_state_t));

	pstate->pname_evaluator = rval_evaluator_alloc_from_ast(pnamenode, type_inferencing);

	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = pstate;
	pevaluator->pprocess_func = NULL;
	switch (type_inferencing) {
	case TYPE_INFER_STRING_ONLY:
		pevaluator->pprocess_func = rval_evaluator_indirect_field_name_func_string_only;
		break;
	case TYPE_INFER_STRING_FLOAT:
		pevaluator->pprocess_func = rval_evaluator_indirect_field_name_func_string_float;
		break;
	case TYPE_INFER_STRING_FLOAT_INT:
		pevaluator->pprocess_func = rval_evaluator_indirect_field_name_func_string_float_int;
		break;
	default:
		fprintf(stderr, "%s: internal coding error detected in file %s at line %d.\n",
			MLR_GLOBALS.argv0, __FILE__, __LINE__);
		exit(1);
		break;
	}
	pevaluator->pfree_func = rval_evaluator_indirect_field_name_free;

	return pevaluator;
}

// ================================================================
typedef struct _rval_evaluator_oosvar_keylist_state_t {
	sllv_t* poosvar_rhs_keylist_evaluators;
} rval_evaluator_oosvar_keylist_state_t;

mv_t rval_evaluator_oosvar_keylist_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_oosvar_keylist_state_t* pstate = pvstate;

	int all_non_null_or_error = TRUE;
	sllmv_t* pmvkeys = evaluate_list(pstate->poosvar_rhs_keylist_evaluators, pvars, &all_non_null_or_error);

	mv_t rv = mv_absent();
	if (all_non_null_or_error) {
		int error = 0;
		mv_t* pval = mlhmmv_get_terminal(pvars->poosvars, pmvkeys, &error);
		if (pval != NULL) {
			if (pval->type == MT_STRING && *pval->u.strv == 0)
				rv = mv_empty();
			else
				rv = mv_copy(pval);
		}
	}

	sllmv_free(pmvkeys);
	return rv;
}

static void rval_evaluator_oosvar_keylist_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_oosvar_keylist_state_t* pstate = pevaluator->pvstate;
	for (sllve_t* pe = pstate->poosvar_rhs_keylist_evaluators->phead; pe != NULL; pe = pe->pnext) {
		rval_evaluator_t* pevaluator = pe->pvvalue;
		pevaluator->pfree_func(pevaluator);
	}
	sllv_free(pstate->poosvar_rhs_keylist_evaluators);
	free(pstate);
	free(pevaluator);
}

// Example AST:
//
// $ mlr -n put -v '$y = @x[1]["two"][$3+4][@5]'
// list (statement_list):
//     = (srec_assignment):
//         y (field_name).
//         oosvar_keylist (oosvar_keylist):
//             x (string_literal).
//             1 (strnum_literal).
//             two (strnum_literal).
//             + (operator):
//                 3 (field_name).
//                 4 (strnum_literal).
//             oosvar_keylist (oosvar_keylist):
//                 5 (string_literal).

rval_evaluator_t* rval_evaluator_alloc_from_oosvar_keylist(mlr_dsl_ast_node_t* pnode, int type_inferencing) {
	rval_evaluator_oosvar_keylist_state_t* pstate = mlr_malloc_or_die(
		sizeof(rval_evaluator_oosvar_keylist_state_t));

	sllv_t* pkeylist_evaluators = sllv_alloc();

	for (sllve_t* pe = pnode->pchildren->phead; pe != NULL; pe = pe->pnext) {
		mlr_dsl_ast_node_t* pkeynode = pe->pvvalue;
		if (pkeynode->type == MD_AST_NODE_TYPE_STRING_LITERAL) {
			sllv_append(pkeylist_evaluators, rval_evaluator_alloc_from_string(pkeynode->text));
		} else {
			sllv_append(pkeylist_evaluators, rval_evaluator_alloc_from_ast(pkeynode, type_inferencing));
		}
	}
	pstate->poosvar_rhs_keylist_evaluators = pkeylist_evaluators;

	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = pstate;
	pevaluator->pprocess_func = NULL;
	pevaluator->pprocess_func = rval_evaluator_oosvar_keylist_func;
	pevaluator->pfree_func = rval_evaluator_oosvar_keylist_free;

	return pevaluator;
}

// ================================================================
// This is used for evaluating strings and numbers in literal expressions, e.g. '$x = "abc"'
// or '$x = "left_\1". The values are subject to replacement with regex captures. See comments
// in mapper_put for more information.
//
// Compare rval_evaluator_alloc_from_string which doesn't do regex replacement: it is intended for
// oosvar names on expression left-hand sides (outside of this file).

typedef struct _rval_evaluator_strnum_literal_state_t {
	mv_t literal;
} rval_evaluator_strnum_literal_state_t;

mv_t rval_evaluator_non_string_literal_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_strnum_literal_state_t* pstate = pvstate;
	return pstate->literal;
}

mv_t rval_evaluator_string_literal_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_strnum_literal_state_t* pstate = pvstate;
	char* input = pstate->literal.u.strv;

	if (pvars->ppregex_captures == NULL || *pvars->ppregex_captures == NULL) {
		return mv_from_string_no_free(input);
	} else {
		int was_allocated = FALSE;
		char* output = interpolate_regex_captures(input, *pvars->ppregex_captures, &was_allocated);
		if (was_allocated)
			return mv_from_string_with_free(output);
		else
			return mv_from_string_no_free(output);
	}
}
static void rval_evaluator_strnum_literal_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_strnum_literal_state_t* pstate = pevaluator->pvstate;
	mv_free(&pstate->literal);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_strnum_literal(char* string, int type_inferencing) {
	rval_evaluator_strnum_literal_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_strnum_literal_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	if (string == NULL) {
		pstate->literal = mv_absent();
		pevaluator->pprocess_func = rval_evaluator_non_string_literal_func;
	} else {
		long long intv;
		double fltv;

		pevaluator->pprocess_func = NULL;
		switch (type_inferencing) {
		case TYPE_INFER_STRING_ONLY:
			pstate->literal = mv_from_string_no_free(string);
			pevaluator->pprocess_func = rval_evaluator_string_literal_func;
			break;

		case TYPE_INFER_STRING_FLOAT:
			if (mlr_try_float_from_string(string, &fltv)) {
				pstate->literal = mv_from_float(fltv);
				pevaluator->pprocess_func = rval_evaluator_non_string_literal_func;
			} else {
				pstate->literal = mv_from_string_no_free(string);
				pevaluator->pprocess_func = rval_evaluator_string_literal_func;
			}
			break;

		case TYPE_INFER_STRING_FLOAT_INT:
			if (mlr_try_int_from_string(string, &intv)) {
				pstate->literal = mv_from_int(intv);
				pevaluator->pprocess_func = rval_evaluator_non_string_literal_func;
			} else if (mlr_try_float_from_string(string, &fltv)) {
				pstate->literal = mv_from_float(fltv);
				pevaluator->pprocess_func = rval_evaluator_non_string_literal_func;
			} else {
				pstate->literal = mv_from_string_no_free(string);
				pevaluator->pprocess_func = rval_evaluator_string_literal_func;
			}
			break;
		default:
			fprintf(stderr, "%s: internal coding error detected in file %s at line %d.\n",
				MLR_GLOBALS.argv0, __FILE__, __LINE__);
			exit(1);
			break;
		}
	}
	pevaluator->pfree_func = rval_evaluator_strnum_literal_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

// ================================================================
// This is intended only for oosvar names on expression left-hand sides (outside of this file).
// Compare rval_evaluator_alloc_from_strnum_literal.

typedef struct _rval_evaluator_string_state_t {
	char* string;
} rval_evaluator_string_state_t;

mv_t rval_evaluator_string_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_string_state_t* pstate = pvstate;
	return mv_from_string_no_free(pstate->string);
}
static void rval_evaluator_string_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_string_state_t* pstate = pevaluator->pvstate;
	free(pstate->string);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_string(char* string) {
	rval_evaluator_string_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_string_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	pstate->string            = mlr_strdup_or_die(string);
	pevaluator->pprocess_func = rval_evaluator_string_func;
	pevaluator->pfree_func    = rval_evaluator_string_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

// ----------------------------------------------------------------
typedef struct _rval_evaluator_boolean_literal_state_t {
	mv_t literal;
} rval_evaluator_boolean_literal_state_t;

mv_t rval_evaluator_boolean_literal_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_boolean_literal_state_t* pstate = pvstate;
	return pstate->literal;
}

static void rval_evaluator_boolean_literal_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_boolean_literal_state_t* pstate = pevaluator->pvstate;
	mv_free(&pstate->literal);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_boolean_literal(char* string) {
	rval_evaluator_boolean_literal_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_boolean_literal_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	if (streq(string, "true")) {
		pstate->literal = mv_from_true();
	} else if (streq(string, "false")) {
		pstate->literal = mv_from_false();
	} else {
		fprintf(stderr, "%s: internal coding error detected in file %s at line %d.\n",
			MLR_GLOBALS.argv0, __FILE__, __LINE__);
		exit(1);
	}
	pevaluator->pprocess_func = rval_evaluator_boolean_literal_func;
	pevaluator->pfree_func = rval_evaluator_boolean_literal_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

rval_evaluator_t* rval_evaluator_alloc_from_boolean(int boolval) {
	rval_evaluator_boolean_literal_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_boolean_literal_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	pstate->literal = mv_from_bool(boolval);
	pevaluator->pprocess_func = rval_evaluator_boolean_literal_func;
	pevaluator->pfree_func = rval_evaluator_boolean_literal_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

// ================================================================
// Example:
// $ mlr put -v '$y=ENV["X"]' ...
// AST BEGIN STATEMENTS (0):
// AST MAIN STATEMENTS (1):
// = (srec_assignment):
//     y (field_name).
//     env (env):
//         ENV (env).
//         X (strnum_literal).
// AST END STATEMENTS (0):

// ----------------------------------------------------------------
typedef struct _rval_evaluator_environment_state_t {
	rval_evaluator_t* pname_evaluator;
} rval_evaluator_environment_state_t;

mv_t rval_evaluator_environment_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_environment_state_t* pstate = pvstate;

	mv_t mvname = pstate->pname_evaluator->pprocess_func(pstate->pname_evaluator->pvstate, pvars);
	if (mv_is_null(&mvname)) {
		return mv_absent();
	}
	char free_flags;
	char* strname = mv_format_val(&mvname, &free_flags);
	char* strvalue = getenv(strname);
	if (strvalue == NULL) {
		mv_free(&mvname);
		if (free_flags & FREE_ENTRY_VALUE)
			free(strname);
		return mv_empty();
	}
	mv_t rv = mv_from_string(strvalue, NO_FREE);
	mv_free(&mvname);
	if (free_flags & FREE_ENTRY_VALUE)
		free(strname);
	return rv;
}

static void rval_evaluator_environment_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_environment_state_t* pstate = pevaluator->pvstate;
	pstate->pname_evaluator->pfree_func(pstate->pname_evaluator);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_environment(mlr_dsl_ast_node_t* pnode, int type_inferencing) {
	rval_evaluator_environment_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_environment_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	mlr_dsl_ast_node_t* pnamenode = pnode->pchildren->phead->pnext->pvvalue;

	pstate->pname_evaluator = rval_evaluator_alloc_from_ast(pnamenode, type_inferencing);
	pevaluator->pprocess_func = rval_evaluator_environment_func;
	pevaluator->pfree_func = rval_evaluator_environment_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

// ================================================================
mv_t rval_evaluator_NF_func(void* pvstate, variables_t* pvars) {
	return mv_from_int(pvars->pinrec->field_count);
}
static void rval_evaluator_NF_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_NF() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_NF_func;
	pevaluator->pfree_func = rval_evaluator_NF_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_NR_func(void* pvstate, variables_t* pvars) {
	return mv_from_int(pvars->pctx->nr);
}
static void rval_evaluator_NR_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_NR() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_NR_func;
	pevaluator->pfree_func = rval_evaluator_NR_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_FNR_func(void* pvstate, variables_t* pvars) {
	return mv_from_int(pvars->pctx->fnr);
}
static void rval_evaluator_FNR_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_FNR() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_FNR_func;
	pevaluator->pfree_func = rval_evaluator_FNR_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_FILENAME_func(void* pvstate, variables_t* pvars) {
	return mv_from_string_no_free(pvars->pctx->filename);
}
static void rval_evaluator_FILENAME_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_FILENAME() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_FILENAME_func;
	pevaluator->pfree_func = rval_evaluator_FILENAME_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_FILENUM_func(void* pvstate, variables_t* pvars) {
	return mv_from_int(pvars->pctx->filenum);
}
static void rval_evaluator_FILENUM_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_FILENUM() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_FILENUM_func;
	pevaluator->pfree_func = rval_evaluator_FILENUM_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_PI_func(void* pvstate, variables_t* pvars) {
	return mv_from_float(M_PI);
}
static void rval_evaluator_PI_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_PI() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_PI_func;
	pevaluator->pfree_func = rval_evaluator_PI_free;
	return pevaluator;
}

// ----------------------------------------------------------------
mv_t rval_evaluator_E_func(void* pvstate, variables_t* pvars) {
	return mv_from_float(M_E);
}
static void rval_evaluator_E_free(rval_evaluator_t* pevaluator) {
	free(pevaluator);
}
rval_evaluator_t* rval_evaluator_alloc_from_E() {
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));
	pevaluator->pvstate = NULL;
	pevaluator->pprocess_func = rval_evaluator_E_func;
	pevaluator->pfree_func = rval_evaluator_E_free;
	return pevaluator;
}

// ================================================================
rval_evaluator_t* rval_evaluator_alloc_from_context_variable(char* variable_name) {
	if        (streq(variable_name, "NF"))       { return rval_evaluator_alloc_from_NF();
	} else if (streq(variable_name, "NR"))       { return rval_evaluator_alloc_from_NR();
	} else if (streq(variable_name, "FNR"))      { return rval_evaluator_alloc_from_FNR();
	} else if (streq(variable_name, "FILENAME")) { return rval_evaluator_alloc_from_FILENAME();
	} else if (streq(variable_name, "FILENUM"))  { return rval_evaluator_alloc_from_FILENUM();
	} else if (streq(variable_name, "PI"))       { return rval_evaluator_alloc_from_PI();
	} else if (streq(variable_name, "E"))        { return rval_evaluator_alloc_from_E();
	} else  { return NULL;
	}
}

// ================================================================
typedef struct _rval_evaluator_from_bound_variable_state_t {
	char* variable_name;
} rval_evaluator_from_bound_variable_state_t;

mv_t rval_evaluator_from_bound_variable_func(void* pvstate, variables_t* pvars) {
	rval_evaluator_from_bound_variable_state_t* pstate = pvstate;
	mv_t* pval = bind_stack_resolve(pvars->pbind_stack, pstate->variable_name);
	if (pval == NULL) {
		return mv_absent();
	} else {
		return mv_copy(pval);
	}
}
static void rval_evaluator_from_bound_variable_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_from_bound_variable_state_t* pstate = pevaluator->pvstate;
	free(pstate->variable_name);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_bound_variable(char* variable_name) {
	rval_evaluator_from_bound_variable_state_t* pstate = mlr_malloc_or_die(
		sizeof(rval_evaluator_from_bound_variable_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	pstate->variable_name     = mlr_strdup_or_die(variable_name);
	pevaluator->pprocess_func = rval_evaluator_from_bound_variable_func;
	pevaluator->pfree_func    = rval_evaluator_from_bound_variable_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}

// ----------------------------------------------------------------
typedef struct _rval_evaluator_mv_state_t {
	mv_t literal;
} rval_evaluator_mv_state_t;

mv_t rval_evaluator_mv_process(void* pvstate, variables_t* pvars) {
	rval_evaluator_mv_state_t* pstate = pvstate;
	return mv_copy(&pstate->literal);

}
static void rval_evaluator_mv_free(rval_evaluator_t* pevaluator) {
	rval_evaluator_mv_state_t* pstate = pevaluator->pvstate;
	mv_free(&pstate->literal);
	free(pstate);
	free(pevaluator);
}

rval_evaluator_t* rval_evaluator_alloc_from_mlrval(mv_t* pval) {
	rval_evaluator_mv_state_t* pstate = mlr_malloc_or_die(sizeof(rval_evaluator_mv_state_t));
	rval_evaluator_t* pevaluator = mlr_malloc_or_die(sizeof(rval_evaluator_t));

	pstate->literal = mv_copy(pval);
	pevaluator->pprocess_func = rval_evaluator_mv_process;
	pevaluator->pfree_func = rval_evaluator_mv_free;

	pevaluator->pvstate = pstate;
	return pevaluator;
}
