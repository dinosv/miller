run_mlr --from $indir/s.dkvp head -n 2 then put -q 'for (k,v in $*) { emit { "foo" : "bar" } }'
run_mlr --from $indir/s.dkvp head -n 2 then put -q 'for (k,v in $*) { emit { "foo" : v } }'
run_mlr --from $indir/s.dkvp head -n 2 then put -q 'for (k,v in $*) { emit { k: "bar" } }'
run_mlr --from $indir/s.dkvp head -n 2 then put -q 'for (k,v in $*) { emit { k : v } }'
run_mlr --from $indir/s.dkvp head -n 1 then put -q 'for (i,e in [3,4,5]) { emit { "foo" : "bar" } }'
run_mlr --from $indir/s.dkvp head -n 1 then put -q 'for (i,e in [3,4,5]) { emit { "foo" : i } }'
run_mlr --from $indir/s.dkvp head -n 1 then put -q 'for (i,e in [3,4,5]) { emit { "foo" : e } }'
