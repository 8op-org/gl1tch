goog.provide('shadow.test');
/**
 * like ct/test-vars-block but more generic
 * groups vars by namespace, executes fixtures
 */
shadow.test.test_vars_grouped_block = (function shadow$test$test_vars_grouped_block(vars){
return cljs.core.mapcat.cljs$core$IFn$_invoke$arity$variadic((function (p__25626){
var vec__25627 = p__25626;
var ns = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25627,(0),null);
var vars__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25627,(1),null);
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){
return cljs.test.report.call(null, new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"begin-test-ns","begin-test-ns",-1701237033),new cljs.core.Keyword(null,"ns","ns",441598760),ns], null));
}),(function (){
return cljs.test.block((function (){var env = cljs.test.get_current_env();
var once_fixtures = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(env,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"once-fixtures","once-fixtures",1253947167),ns], null));
var each_fixtures = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(env,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"each-fixtures","each-fixtures",802243977),ns], null));
var G__25632 = cljs.test.execution_strategy(once_fixtures,each_fixtures);
var G__25632__$1 = (((G__25632 instanceof cljs.core.Keyword))?G__25632.fqn:null);
switch (G__25632__$1) {
case "async":
return cljs.test.wrap_map_fixtures(once_fixtures,cljs.core.mapcat.cljs$core$IFn$_invoke$arity$variadic(cljs.core.comp.cljs$core$IFn$_invoke$arity$2(cljs.core.partial.cljs$core$IFn$_invoke$arity$2(cljs.test.wrap_map_fixtures,each_fixtures),cljs.test.test_var_block),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.filter.cljs$core$IFn$_invoke$arity$2(cljs.core.comp.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"test","test",577538877),cljs.core.meta),vars__$1)], 0)));

break;
case "sync":
var each_fixture_fn = cljs.test.join_fixtures(each_fixtures);
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){
var G__25639 = (function (){
var seq__25640 = cljs.core.seq(vars__$1);
var chunk__25641 = null;
var count__25642 = (0);
var i__25643 = (0);
while(true){
if((i__25643 < count__25642)){
var v = chunk__25641.cljs$core$IIndexed$_nth$arity$2(null, i__25643);
var temp__5825__auto___25743 = new cljs.core.Keyword(null,"test","test",577538877).cljs$core$IFn$_invoke$arity$1(cljs.core.meta(v));
if(cljs.core.truth_(temp__5825__auto___25743)){
var t_25744 = temp__5825__auto___25743;
var G__25647_25745 = ((function (seq__25640,chunk__25641,count__25642,i__25643,t_25744,temp__5825__auto___25743,v,each_fixture_fn,G__25632,G__25632__$1,env,once_fixtures,each_fixtures,vec__25627,ns,vars__$1){
return (function (){
return cljs.test.run_block(cljs.test.test_var_block_STAR_(v,cljs.test.disable_async(t_25744)));
});})(seq__25640,chunk__25641,count__25642,i__25643,t_25744,temp__5825__auto___25743,v,each_fixture_fn,G__25632,G__25632__$1,env,once_fixtures,each_fixtures,vec__25627,ns,vars__$1))
;
(each_fixture_fn.cljs$core$IFn$_invoke$arity$1 ? each_fixture_fn.cljs$core$IFn$_invoke$arity$1(G__25647_25745) : each_fixture_fn.call(null, G__25647_25745));
} else {
}


var G__25747 = seq__25640;
var G__25748 = chunk__25641;
var G__25749 = count__25642;
var G__25750 = (i__25643 + (1));
seq__25640 = G__25747;
chunk__25641 = G__25748;
count__25642 = G__25749;
i__25643 = G__25750;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__25640);
if(temp__5825__auto__){
var seq__25640__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__25640__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__25640__$1);
var G__25751 = cljs.core.chunk_rest(seq__25640__$1);
var G__25752 = c__5525__auto__;
var G__25753 = cljs.core.count(c__5525__auto__);
var G__25754 = (0);
seq__25640 = G__25751;
chunk__25641 = G__25752;
count__25642 = G__25753;
i__25643 = G__25754;
continue;
} else {
var v = cljs.core.first(seq__25640__$1);
var temp__5825__auto___25756__$1 = new cljs.core.Keyword(null,"test","test",577538877).cljs$core$IFn$_invoke$arity$1(cljs.core.meta(v));
if(cljs.core.truth_(temp__5825__auto___25756__$1)){
var t_25758 = temp__5825__auto___25756__$1;
var G__25648_25759 = ((function (seq__25640,chunk__25641,count__25642,i__25643,t_25758,temp__5825__auto___25756__$1,v,seq__25640__$1,temp__5825__auto__,each_fixture_fn,G__25632,G__25632__$1,env,once_fixtures,each_fixtures,vec__25627,ns,vars__$1){
return (function (){
return cljs.test.run_block(cljs.test.test_var_block_STAR_(v,cljs.test.disable_async(t_25758)));
});})(seq__25640,chunk__25641,count__25642,i__25643,t_25758,temp__5825__auto___25756__$1,v,seq__25640__$1,temp__5825__auto__,each_fixture_fn,G__25632,G__25632__$1,env,once_fixtures,each_fixtures,vec__25627,ns,vars__$1))
;
(each_fixture_fn.cljs$core$IFn$_invoke$arity$1 ? each_fixture_fn.cljs$core$IFn$_invoke$arity$1(G__25648_25759) : each_fixture_fn.call(null, G__25648_25759));
} else {
}


var G__25760 = cljs.core.next(seq__25640__$1);
var G__25761 = null;
var G__25762 = (0);
var G__25763 = (0);
seq__25640 = G__25760;
chunk__25641 = G__25761;
count__25642 = G__25762;
i__25643 = G__25763;
continue;
}
} else {
return null;
}
}
break;
}
});
var fexpr__25638 = cljs.test.join_fixtures(once_fixtures);
return (fexpr__25638.cljs$core$IFn$_invoke$arity$1 ? fexpr__25638.cljs$core$IFn$_invoke$arity$1(G__25639) : fexpr__25638.call(null, G__25639));
})], null);

break;
default:
throw (new Error(["No matching clause: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(G__25632__$1)].join('')));

}
})());
}),(function (){
return cljs.test.report.call(null, new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"end-test-ns","end-test-ns",1620675645),new cljs.core.Keyword(null,"ns","ns",441598760),ns], null));
})], null);
}),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.sort_by.cljs$core$IFn$_invoke$arity$2(cljs.core.first,cljs.core.group_by((function (p1__25621_SHARP_){
return new cljs.core.Keyword(null,"ns","ns",441598760).cljs$core$IFn$_invoke$arity$1(cljs.core.meta(p1__25621_SHARP_));
}),vars))], 0));
});
/**
 * Like test-ns, but returns a block for further composition and
 *   later execution.  Does not clear the current env.
 */
shadow.test.test_ns_block = (function shadow$test$test_ns_block(ns){
if((ns instanceof cljs.core.Symbol)){
} else {
throw (new Error("Assert failed: (symbol? ns)"));
}

var map__25663 = shadow.test.env.get_test_ns_info(ns);
var map__25663__$1 = cljs.core.__destructure_map(map__25663);
var test_ns = map__25663__$1;
var vars = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25663__$1,new cljs.core.Keyword(null,"vars","vars",-2046957217));
if(cljs.core.not(test_ns)){
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){
return cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([["Namespace: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(ns)," not found, no tests to run."].join('')], 0));
})], null);
} else {
return shadow.test.test_vars_grouped_block(vars);
}
});
shadow.test.prepare_test_run = (function shadow$test$prepare_test_run(p__25666,vars){
var map__25669 = p__25666;
var map__25669__$1 = cljs.core.__destructure_map(map__25669);
var env = map__25669__$1;
var report_fn = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25669__$1,new cljs.core.Keyword(null,"report-fn","report-fn",-549046115));
var orig_report = cljs.test.report;
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){
cljs.test.set_env_BANG_(cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(env,new cljs.core.Keyword("shadow.test","report-fn","shadow.test/report-fn",1075704061),orig_report));

if(cljs.core.truth_(report_fn)){
(cljs.test.report = report_fn);
} else {
}

var seq__25673_25764 = cljs.core.seq(shadow.test.env.get_tests());
var chunk__25675_25765 = null;
var count__25676_25766 = (0);
var i__25677_25767 = (0);
while(true){
if((i__25677_25767 < count__25676_25766)){
var vec__25703_25768 = chunk__25675_25765.cljs$core$IIndexed$_nth$arity$2(null, i__25677_25767);
var test_ns_25769 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25703_25768,(0),null);
var ns_info_25770 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25703_25768,(1),null);
var map__25706_25771 = ns_info_25770;
var map__25706_25772__$1 = cljs.core.__destructure_map(map__25706_25771);
var fixtures_25773 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25706_25772__$1,new cljs.core.Keyword(null,"fixtures","fixtures",1009814994));
var temp__5825__auto___25774 = new cljs.core.Keyword(null,"once","once",-262568523).cljs$core$IFn$_invoke$arity$1(fixtures_25773);
if(cljs.core.truth_(temp__5825__auto___25774)){
var fix_25775 = temp__5825__auto___25774;
cljs.test.update_current_env_BANG_.cljs$core$IFn$_invoke$arity$variadic(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"once-fixtures","once-fixtures",1253947167)], null),cljs.core.assoc,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([test_ns_25769,fix_25775], 0));
} else {
}

var temp__5825__auto___25776 = new cljs.core.Keyword(null,"each","each",940016129).cljs$core$IFn$_invoke$arity$1(fixtures_25773);
if(cljs.core.truth_(temp__5825__auto___25776)){
var fix_25777 = temp__5825__auto___25776;
cljs.test.update_current_env_BANG_.cljs$core$IFn$_invoke$arity$variadic(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"each-fixtures","each-fixtures",802243977)], null),cljs.core.assoc,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([test_ns_25769,fix_25777], 0));
} else {
}


var G__25778 = seq__25673_25764;
var G__25779 = chunk__25675_25765;
var G__25780 = count__25676_25766;
var G__25781 = (i__25677_25767 + (1));
seq__25673_25764 = G__25778;
chunk__25675_25765 = G__25779;
count__25676_25766 = G__25780;
i__25677_25767 = G__25781;
continue;
} else {
var temp__5825__auto___25782 = cljs.core.seq(seq__25673_25764);
if(temp__5825__auto___25782){
var seq__25673_25783__$1 = temp__5825__auto___25782;
if(cljs.core.chunked_seq_QMARK_(seq__25673_25783__$1)){
var c__5525__auto___25784 = cljs.core.chunk_first(seq__25673_25783__$1);
var G__25785 = cljs.core.chunk_rest(seq__25673_25783__$1);
var G__25786 = c__5525__auto___25784;
var G__25787 = cljs.core.count(c__5525__auto___25784);
var G__25788 = (0);
seq__25673_25764 = G__25785;
chunk__25675_25765 = G__25786;
count__25676_25766 = G__25787;
i__25677_25767 = G__25788;
continue;
} else {
var vec__25710_25789 = cljs.core.first(seq__25673_25783__$1);
var test_ns_25790 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25710_25789,(0),null);
var ns_info_25791 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__25710_25789,(1),null);
var map__25713_25792 = ns_info_25791;
var map__25713_25793__$1 = cljs.core.__destructure_map(map__25713_25792);
var fixtures_25794 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25713_25793__$1,new cljs.core.Keyword(null,"fixtures","fixtures",1009814994));
var temp__5825__auto___25795__$1 = new cljs.core.Keyword(null,"once","once",-262568523).cljs$core$IFn$_invoke$arity$1(fixtures_25794);
if(cljs.core.truth_(temp__5825__auto___25795__$1)){
var fix_25796 = temp__5825__auto___25795__$1;
cljs.test.update_current_env_BANG_.cljs$core$IFn$_invoke$arity$variadic(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"once-fixtures","once-fixtures",1253947167)], null),cljs.core.assoc,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([test_ns_25790,fix_25796], 0));
} else {
}

var temp__5825__auto___25797__$1 = new cljs.core.Keyword(null,"each","each",940016129).cljs$core$IFn$_invoke$arity$1(fixtures_25794);
if(cljs.core.truth_(temp__5825__auto___25797__$1)){
var fix_25798 = temp__5825__auto___25797__$1;
cljs.test.update_current_env_BANG_.cljs$core$IFn$_invoke$arity$variadic(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"each-fixtures","each-fixtures",802243977)], null),cljs.core.assoc,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([test_ns_25790,fix_25798], 0));
} else {
}


var G__25799 = cljs.core.next(seq__25673_25783__$1);
var G__25800 = null;
var G__25801 = (0);
var G__25802 = (0);
seq__25673_25764 = G__25799;
chunk__25675_25765 = G__25800;
count__25676_25766 = G__25801;
i__25677_25767 = G__25802;
continue;
}
} else {
}
}
break;
}

return cljs.test.report.call(null, new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"begin-run-tests","begin-run-tests",309363062),new cljs.core.Keyword(null,"var-count","var-count",-1513152110),cljs.core.count(vars),new cljs.core.Keyword(null,"ns-count","ns-count",-1269070724),cljs.core.count(cljs.core.set(cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p1__25665_SHARP_){
return new cljs.core.Keyword(null,"ns","ns",441598760).cljs$core$IFn$_invoke$arity$1(cljs.core.meta(p1__25665_SHARP_));
}),vars)))], null));
})], null);
});
shadow.test.finish_test_run = (function shadow$test$finish_test_run(block){
if(cljs.core.vector_QMARK_(block)){
} else {
throw (new Error("Assert failed: (vector? block)"));
}

return cljs.core.conj.cljs$core$IFn$_invoke$arity$2(block,(function (){
var map__25719 = cljs.test.get_current_env();
var map__25719__$1 = cljs.core.__destructure_map(map__25719);
var env = map__25719__$1;
var report_fn = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25719__$1,new cljs.core.Keyword("shadow.test","report-fn","shadow.test/report-fn",1075704061));
var report_counters = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25719__$1,new cljs.core.Keyword(null,"report-counters","report-counters",-1702609242));
cljs.test.report.call(null, cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(report_counters,new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"summary","summary",380847952)));

cljs.test.report.call(null, cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(report_counters,new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"end-run-tests","end-run-tests",267300563)));

return (cljs.test.report = report_fn);
}));
});
/**
 * tests all vars grouped by namespace, expects seq of test vars, can be obtained from env
 */
shadow.test.run_test_vars = (function shadow$test$run_test_vars(var_args){
var G__25722 = arguments.length;
switch (G__25722) {
case 1:
return shadow.test.run_test_vars.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.test.run_test_vars.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.test.run_test_vars.cljs$core$IFn$_invoke$arity$1 = (function (test_vars){
return shadow.test.run_test_vars.cljs$core$IFn$_invoke$arity$2(cljs.test.empty_env.cljs$core$IFn$_invoke$arity$0(),test_vars);
}));

(shadow.test.run_test_vars.cljs$core$IFn$_invoke$arity$2 = (function (env,vars){
return cljs.test.run_block(shadow.test.finish_test_run(cljs.core.into.cljs$core$IFn$_invoke$arity$2(shadow.test.prepare_test_run(env,vars),shadow.test.test_vars_grouped_block(vars))));
}));

(shadow.test.run_test_vars.cljs$lang$maxFixedArity = 2);

/**
 * test all vars for given namespace symbol
 */
shadow.test.test_ns = (function shadow$test$test_ns(var_args){
var G__25724 = arguments.length;
switch (G__25724) {
case 1:
return shadow.test.test_ns.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.test.test_ns.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.test.test_ns.cljs$core$IFn$_invoke$arity$1 = (function (ns){
return shadow.test.test_ns.cljs$core$IFn$_invoke$arity$2(cljs.test.empty_env.cljs$core$IFn$_invoke$arity$0(),ns);
}));

(shadow.test.test_ns.cljs$core$IFn$_invoke$arity$2 = (function (env,ns){
var map__25725 = shadow.test.env.get_test_ns_info(ns);
var map__25725__$1 = cljs.core.__destructure_map(map__25725);
var vars = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__25725__$1,new cljs.core.Keyword(null,"vars","vars",-2046957217));
return cljs.test.run_block(shadow.test.finish_test_run(cljs.core.into.cljs$core$IFn$_invoke$arity$2(shadow.test.prepare_test_run(env,vars),shadow.test.test_vars_grouped_block(vars))));
}));

(shadow.test.test_ns.cljs$lang$maxFixedArity = 2);

/**
 * test all vars in specified namespace symbol set
 */
shadow.test.run_tests = (function shadow$test$run_tests(var_args){
var G__25730 = arguments.length;
switch (G__25730) {
case 0:
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$0();

break;
case 1:
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.test.run_tests.cljs$core$IFn$_invoke$arity$0 = (function (){
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$1(cljs.test.empty_env.cljs$core$IFn$_invoke$arity$0());
}));

(shadow.test.run_tests.cljs$core$IFn$_invoke$arity$1 = (function (env){
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$2(env,shadow.test.env.get_test_namespaces());
}));

(shadow.test.run_tests.cljs$core$IFn$_invoke$arity$2 = (function (env,namespaces){
if(cljs.core.set_QMARK_(namespaces)){
} else {
throw (new Error("Assert failed: (set? namespaces)"));
}

var vars = cljs.core.filter.cljs$core$IFn$_invoke$arity$2((function (p1__25726_SHARP_){
return cljs.core.contains_QMARK_(namespaces,new cljs.core.Keyword(null,"ns","ns",441598760).cljs$core$IFn$_invoke$arity$1(cljs.core.meta(p1__25726_SHARP_)));
}),shadow.test.env.get_test_vars());
return cljs.test.run_block(shadow.test.finish_test_run(cljs.core.into.cljs$core$IFn$_invoke$arity$2(shadow.test.prepare_test_run(env,vars),shadow.test.test_vars_grouped_block(vars))));
}));

(shadow.test.run_tests.cljs$lang$maxFixedArity = 2);

/**
 * Runs all tests in all namespaces; prints results.
 *   Optional argument is a regular expression; only namespaces with
 *   names matching the regular expression (with re-matches) will be
 *   tested.
 */
shadow.test.run_all_tests = (function shadow$test$run_all_tests(var_args){
var G__25735 = arguments.length;
switch (G__25735) {
case 0:
return shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$0();

break;
case 1:
return shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$0 = (function (){
return shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$2(cljs.test.empty_env.cljs$core$IFn$_invoke$arity$0(),null);
}));

(shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$1 = (function (env){
return shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$2(env,null);
}));

(shadow.test.run_all_tests.cljs$core$IFn$_invoke$arity$2 = (function (env,re){
return shadow.test.run_tests.cljs$core$IFn$_invoke$arity$2(env,cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentHashSet.EMPTY,cljs.core.filter.cljs$core$IFn$_invoke$arity$2((function (p1__25733_SHARP_){
var or__5002__auto__ = (re == null);
if(or__5002__auto__){
return or__5002__auto__;
} else {
return cljs.core.re_matches(re,cljs.core.str.cljs$core$IFn$_invoke$arity$1(p1__25733_SHARP_));
}
}),shadow.test.env.get_test_namespaces())));
}));

(shadow.test.run_all_tests.cljs$lang$maxFixedArity = 2);


//# sourceMappingURL=shadow.test.js.map
