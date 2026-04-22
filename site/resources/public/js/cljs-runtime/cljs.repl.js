goog.provide('cljs.repl');
cljs.repl.print_doc = (function cljs$repl$print_doc(p__35431){
var map__35432 = p__35431;
var map__35432__$1 = cljs.core.__destructure_map(map__35432);
var m = map__35432__$1;
var n = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35432__$1,new cljs.core.Keyword(null,"ns","ns",441598760));
var nm = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35432__$1,new cljs.core.Keyword(null,"name","name",1843675177));
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["-------------------------"], 0));

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([(function (){var or__5002__auto__ = new cljs.core.Keyword(null,"spec","spec",347520401).cljs$core$IFn$_invoke$arity$1(m);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return [(function (){var temp__5825__auto__ = new cljs.core.Keyword(null,"ns","ns",441598760).cljs$core$IFn$_invoke$arity$1(m);
if(cljs.core.truth_(temp__5825__auto__)){
var ns = temp__5825__auto__;
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(ns),"/"].join('');
} else {
return null;
}
})(),cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"name","name",1843675177).cljs$core$IFn$_invoke$arity$1(m))].join('');
}
})()], 0));

if(cljs.core.truth_(new cljs.core.Keyword(null,"protocol","protocol",652470118).cljs$core$IFn$_invoke$arity$1(m))){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["Protocol"], 0));
} else {
}

if(cljs.core.truth_(new cljs.core.Keyword(null,"forms","forms",2045992350).cljs$core$IFn$_invoke$arity$1(m))){
var seq__35433_35718 = cljs.core.seq(new cljs.core.Keyword(null,"forms","forms",2045992350).cljs$core$IFn$_invoke$arity$1(m));
var chunk__35434_35719 = null;
var count__35435_35720 = (0);
var i__35436_35721 = (0);
while(true){
if((i__35436_35721 < count__35435_35720)){
var f_35722 = chunk__35434_35719.cljs$core$IIndexed$_nth$arity$2(null, i__35436_35721);
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["  ",f_35722], 0));


var G__35723 = seq__35433_35718;
var G__35724 = chunk__35434_35719;
var G__35725 = count__35435_35720;
var G__35726 = (i__35436_35721 + (1));
seq__35433_35718 = G__35723;
chunk__35434_35719 = G__35724;
count__35435_35720 = G__35725;
i__35436_35721 = G__35726;
continue;
} else {
var temp__5825__auto___35727 = cljs.core.seq(seq__35433_35718);
if(temp__5825__auto___35727){
var seq__35433_35728__$1 = temp__5825__auto___35727;
if(cljs.core.chunked_seq_QMARK_(seq__35433_35728__$1)){
var c__5525__auto___35729 = cljs.core.chunk_first(seq__35433_35728__$1);
var G__35730 = cljs.core.chunk_rest(seq__35433_35728__$1);
var G__35731 = c__5525__auto___35729;
var G__35732 = cljs.core.count(c__5525__auto___35729);
var G__35733 = (0);
seq__35433_35718 = G__35730;
chunk__35434_35719 = G__35731;
count__35435_35720 = G__35732;
i__35436_35721 = G__35733;
continue;
} else {
var f_35744 = cljs.core.first(seq__35433_35728__$1);
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["  ",f_35744], 0));


var G__35747 = cljs.core.next(seq__35433_35728__$1);
var G__35748 = null;
var G__35750 = (0);
var G__35751 = (0);
seq__35433_35718 = G__35747;
chunk__35434_35719 = G__35748;
count__35435_35720 = G__35750;
i__35436_35721 = G__35751;
continue;
}
} else {
}
}
break;
}
} else {
if(cljs.core.truth_(new cljs.core.Keyword(null,"arglists","arglists",1661989754).cljs$core$IFn$_invoke$arity$1(m))){
var arglists_35756 = new cljs.core.Keyword(null,"arglists","arglists",1661989754).cljs$core$IFn$_invoke$arity$1(m);
if(cljs.core.truth_((function (){var or__5002__auto__ = new cljs.core.Keyword(null,"macro","macro",-867863404).cljs$core$IFn$_invoke$arity$1(m);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return new cljs.core.Keyword(null,"repl-special-function","repl-special-function",1262603725).cljs$core$IFn$_invoke$arity$1(m);
}
})())){
cljs.core.prn.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([arglists_35756], 0));
} else {
cljs.core.prn.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Symbol(null,"quote","quote",1377916282,null),cljs.core.first(arglists_35756)))?cljs.core.second(arglists_35756):arglists_35756)], 0));
}
} else {
}
}

if(cljs.core.truth_(new cljs.core.Keyword(null,"special-form","special-form",-1326536374).cljs$core$IFn$_invoke$arity$1(m))){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["Special Form"], 0));

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",new cljs.core.Keyword(null,"doc","doc",1913296891).cljs$core$IFn$_invoke$arity$1(m)], 0));

if(cljs.core.contains_QMARK_(m,new cljs.core.Keyword(null,"url","url",276297046))){
if(cljs.core.truth_(new cljs.core.Keyword(null,"url","url",276297046).cljs$core$IFn$_invoke$arity$1(m))){
return cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([["\n  Please see http://clojure.org/",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"url","url",276297046).cljs$core$IFn$_invoke$arity$1(m))].join('')], 0));
} else {
return null;
}
} else {
return cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([["\n  Please see http://clojure.org/special_forms#",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"name","name",1843675177).cljs$core$IFn$_invoke$arity$1(m))].join('')], 0));
}
} else {
if(cljs.core.truth_(new cljs.core.Keyword(null,"macro","macro",-867863404).cljs$core$IFn$_invoke$arity$1(m))){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["Macro"], 0));
} else {
}

if(cljs.core.truth_(new cljs.core.Keyword(null,"spec","spec",347520401).cljs$core$IFn$_invoke$arity$1(m))){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["Spec"], 0));
} else {
}

if(cljs.core.truth_(new cljs.core.Keyword(null,"repl-special-function","repl-special-function",1262603725).cljs$core$IFn$_invoke$arity$1(m))){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["REPL Special Function"], 0));
} else {
}

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",new cljs.core.Keyword(null,"doc","doc",1913296891).cljs$core$IFn$_invoke$arity$1(m)], 0));

if(cljs.core.truth_(new cljs.core.Keyword(null,"protocol","protocol",652470118).cljs$core$IFn$_invoke$arity$1(m))){
var seq__35467_35859 = cljs.core.seq(new cljs.core.Keyword(null,"methods","methods",453930866).cljs$core$IFn$_invoke$arity$1(m));
var chunk__35468_35860 = null;
var count__35469_35861 = (0);
var i__35470_35862 = (0);
while(true){
if((i__35470_35862 < count__35469_35861)){
var vec__35483_35882 = chunk__35468_35860.cljs$core$IIndexed$_nth$arity$2(null, i__35470_35862);
var name_35883 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35483_35882,(0),null);
var map__35486_35884 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35483_35882,(1),null);
var map__35486_35885__$1 = cljs.core.__destructure_map(map__35486_35884);
var doc_35886 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35486_35885__$1,new cljs.core.Keyword(null,"doc","doc",1913296891));
var arglists_35887 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35486_35885__$1,new cljs.core.Keyword(null,"arglists","arglists",1661989754));
cljs.core.println();

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",name_35883], 0));

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",arglists_35887], 0));

if(cljs.core.truth_(doc_35886)){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",doc_35886], 0));
} else {
}


var G__35890 = seq__35467_35859;
var G__35891 = chunk__35468_35860;
var G__35892 = count__35469_35861;
var G__35893 = (i__35470_35862 + (1));
seq__35467_35859 = G__35890;
chunk__35468_35860 = G__35891;
count__35469_35861 = G__35892;
i__35470_35862 = G__35893;
continue;
} else {
var temp__5825__auto___35895 = cljs.core.seq(seq__35467_35859);
if(temp__5825__auto___35895){
var seq__35467_35896__$1 = temp__5825__auto___35895;
if(cljs.core.chunked_seq_QMARK_(seq__35467_35896__$1)){
var c__5525__auto___35897 = cljs.core.chunk_first(seq__35467_35896__$1);
var G__35901 = cljs.core.chunk_rest(seq__35467_35896__$1);
var G__35902 = c__5525__auto___35897;
var G__35903 = cljs.core.count(c__5525__auto___35897);
var G__35904 = (0);
seq__35467_35859 = G__35901;
chunk__35468_35860 = G__35902;
count__35469_35861 = G__35903;
i__35470_35862 = G__35904;
continue;
} else {
var vec__35487_35905 = cljs.core.first(seq__35467_35896__$1);
var name_35906 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35487_35905,(0),null);
var map__35490_35907 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35487_35905,(1),null);
var map__35490_35908__$1 = cljs.core.__destructure_map(map__35490_35907);
var doc_35909 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35490_35908__$1,new cljs.core.Keyword(null,"doc","doc",1913296891));
var arglists_35910 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35490_35908__$1,new cljs.core.Keyword(null,"arglists","arglists",1661989754));
cljs.core.println();

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",name_35906], 0));

cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",arglists_35910], 0));

if(cljs.core.truth_(doc_35909)){
cljs.core.println.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([" ",doc_35909], 0));
} else {
}


var G__35920 = cljs.core.next(seq__35467_35896__$1);
var G__35921 = null;
var G__35922 = (0);
var G__35923 = (0);
seq__35467_35859 = G__35920;
chunk__35468_35860 = G__35921;
count__35469_35861 = G__35922;
i__35470_35862 = G__35923;
continue;
}
} else {
}
}
break;
}
} else {
}

if(cljs.core.truth_(n)){
var temp__5825__auto__ = cljs.spec.alpha.get_spec(cljs.core.symbol.cljs$core$IFn$_invoke$arity$2(cljs.core.str.cljs$core$IFn$_invoke$arity$1(cljs.core.ns_name(n)),cljs.core.name(nm)));
if(cljs.core.truth_(temp__5825__auto__)){
var fnspec = temp__5825__auto__;
cljs.core.print.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["Spec"], 0));

var seq__35503 = cljs.core.seq(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"args","args",1315556576),new cljs.core.Keyword(null,"ret","ret",-468222814),new cljs.core.Keyword(null,"fn","fn",-1175266204)], null));
var chunk__35504 = null;
var count__35505 = (0);
var i__35506 = (0);
while(true){
if((i__35506 < count__35505)){
var role = chunk__35504.cljs$core$IIndexed$_nth$arity$2(null, i__35506);
var temp__5825__auto___35926__$1 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(fnspec,role);
if(cljs.core.truth_(temp__5825__auto___35926__$1)){
var spec_35927 = temp__5825__auto___35926__$1;
cljs.core.print.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([["\n ",cljs.core.name(role),":"].join(''),cljs.spec.alpha.describe(spec_35927)], 0));
} else {
}


var G__35928 = seq__35503;
var G__35929 = chunk__35504;
var G__35930 = count__35505;
var G__35931 = (i__35506 + (1));
seq__35503 = G__35928;
chunk__35504 = G__35929;
count__35505 = G__35930;
i__35506 = G__35931;
continue;
} else {
var temp__5825__auto____$1 = cljs.core.seq(seq__35503);
if(temp__5825__auto____$1){
var seq__35503__$1 = temp__5825__auto____$1;
if(cljs.core.chunked_seq_QMARK_(seq__35503__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__35503__$1);
var G__35932 = cljs.core.chunk_rest(seq__35503__$1);
var G__35933 = c__5525__auto__;
var G__35934 = cljs.core.count(c__5525__auto__);
var G__35935 = (0);
seq__35503 = G__35932;
chunk__35504 = G__35933;
count__35505 = G__35934;
i__35506 = G__35935;
continue;
} else {
var role = cljs.core.first(seq__35503__$1);
var temp__5825__auto___35936__$2 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(fnspec,role);
if(cljs.core.truth_(temp__5825__auto___35936__$2)){
var spec_35937 = temp__5825__auto___35936__$2;
cljs.core.print.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([["\n ",cljs.core.name(role),":"].join(''),cljs.spec.alpha.describe(spec_35937)], 0));
} else {
}


var G__35938 = cljs.core.next(seq__35503__$1);
var G__35939 = null;
var G__35940 = (0);
var G__35941 = (0);
seq__35503 = G__35938;
chunk__35504 = G__35939;
count__35505 = G__35940;
i__35506 = G__35941;
continue;
}
} else {
return null;
}
}
break;
}
} else {
return null;
}
} else {
return null;
}
}
});
/**
 * Constructs a data representation for a Error with keys:
 *  :cause - root cause message
 *  :phase - error phase
 *  :via - cause chain, with cause keys:
 *           :type - exception class symbol
 *           :message - exception message
 *           :data - ex-data
 *           :at - top stack element
 *  :trace - root cause stack elements
 */
cljs.repl.Error__GT_map = (function cljs$repl$Error__GT_map(o){
return cljs.core.Throwable__GT_map(o);
});
/**
 * Returns an analysis of the phase, error, cause, and location of an error that occurred
 *   based on Throwable data, as returned by Throwable->map. All attributes other than phase
 *   are optional:
 *  :clojure.error/phase - keyword phase indicator, one of:
 *    :read-source :compile-syntax-check :compilation :macro-syntax-check :macroexpansion
 *    :execution :read-eval-result :print-eval-result
 *  :clojure.error/source - file name (no path)
 *  :clojure.error/line - integer line number
 *  :clojure.error/column - integer column number
 *  :clojure.error/symbol - symbol being expanded/compiled/invoked
 *  :clojure.error/class - cause exception class symbol
 *  :clojure.error/cause - cause exception message
 *  :clojure.error/spec - explain-data for spec error
 */
cljs.repl.ex_triage = (function cljs$repl$ex_triage(datafied_throwable){
var map__35516 = datafied_throwable;
var map__35516__$1 = cljs.core.__destructure_map(map__35516);
var via = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35516__$1,new cljs.core.Keyword(null,"via","via",-1904457336));
var trace = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35516__$1,new cljs.core.Keyword(null,"trace","trace",-1082747415));
var phase = cljs.core.get.cljs$core$IFn$_invoke$arity$3(map__35516__$1,new cljs.core.Keyword(null,"phase","phase",575722892),new cljs.core.Keyword(null,"execution","execution",253283524));
var map__35517 = cljs.core.last(via);
var map__35517__$1 = cljs.core.__destructure_map(map__35517);
var type = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35517__$1,new cljs.core.Keyword(null,"type","type",1174270348));
var message = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35517__$1,new cljs.core.Keyword(null,"message","message",-406056002));
var data = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35517__$1,new cljs.core.Keyword(null,"data","data",-232669377));
var map__35518 = data;
var map__35518__$1 = cljs.core.__destructure_map(map__35518);
var problems = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35518__$1,new cljs.core.Keyword("cljs.spec.alpha","problems","cljs.spec.alpha/problems",447400814));
var fn = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35518__$1,new cljs.core.Keyword("cljs.spec.alpha","fn","cljs.spec.alpha/fn",408600443));
var caller = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35518__$1,new cljs.core.Keyword("cljs.spec.test.alpha","caller","cljs.spec.test.alpha/caller",-398302390));
var map__35519 = new cljs.core.Keyword(null,"data","data",-232669377).cljs$core$IFn$_invoke$arity$1(cljs.core.first(via));
var map__35519__$1 = cljs.core.__destructure_map(map__35519);
var top_data = map__35519__$1;
var source = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35519__$1,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397));
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3((function (){var G__35532 = phase;
var G__35532__$1 = (((G__35532 instanceof cljs.core.Keyword))?G__35532.fqn:null);
switch (G__35532__$1) {
case "read-source":
var map__35536 = data;
var map__35536__$1 = cljs.core.__destructure_map(map__35536);
var line = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35536__$1,new cljs.core.Keyword("clojure.error","line","clojure.error/line",-1816287471));
var column = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35536__$1,new cljs.core.Keyword("clojure.error","column","clojure.error/column",304721553));
var G__35537 = cljs.core.merge.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"data","data",-232669377).cljs$core$IFn$_invoke$arity$1(cljs.core.second(via)),top_data], 0));
var G__35537__$1 = (cljs.core.truth_(source)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35537,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397),source):G__35537);
var G__35537__$2 = (cljs.core.truth_((function (){var fexpr__35538 = new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, ["NO_SOURCE_PATH",null,"NO_SOURCE_FILE",null], null), null);
return (fexpr__35538.cljs$core$IFn$_invoke$arity$1 ? fexpr__35538.cljs$core$IFn$_invoke$arity$1(source) : fexpr__35538.call(null, source));
})())?cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(G__35537__$1,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397)):G__35537__$1);
if(cljs.core.truth_(message)){
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35537__$2,new cljs.core.Keyword("clojure.error","cause","clojure.error/cause",-1879175742),message);
} else {
return G__35537__$2;
}

break;
case "compile-syntax-check":
case "compilation":
case "macro-syntax-check":
case "macroexpansion":
var G__35545 = top_data;
var G__35545__$1 = (cljs.core.truth_(source)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35545,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397),source):G__35545);
var G__35545__$2 = (cljs.core.truth_((function (){var fexpr__35548 = new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, ["NO_SOURCE_PATH",null,"NO_SOURCE_FILE",null], null), null);
return (fexpr__35548.cljs$core$IFn$_invoke$arity$1 ? fexpr__35548.cljs$core$IFn$_invoke$arity$1(source) : fexpr__35548.call(null, source));
})())?cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(G__35545__$1,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397)):G__35545__$1);
var G__35545__$3 = (cljs.core.truth_(type)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35545__$2,new cljs.core.Keyword("clojure.error","class","clojure.error/class",278435890),type):G__35545__$2);
var G__35545__$4 = (cljs.core.truth_(message)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35545__$3,new cljs.core.Keyword("clojure.error","cause","clojure.error/cause",-1879175742),message):G__35545__$3);
if(cljs.core.truth_(problems)){
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35545__$4,new cljs.core.Keyword("clojure.error","spec","clojure.error/spec",2055032595),data);
} else {
return G__35545__$4;
}

break;
case "read-eval-result":
case "print-eval-result":
var vec__35556 = cljs.core.first(trace);
var source__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35556,(0),null);
var method = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35556,(1),null);
var file = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35556,(2),null);
var line = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35556,(3),null);
var G__35562 = top_data;
var G__35562__$1 = (cljs.core.truth_(line)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35562,new cljs.core.Keyword("clojure.error","line","clojure.error/line",-1816287471),line):G__35562);
var G__35562__$2 = (cljs.core.truth_(file)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35562__$1,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397),file):G__35562__$1);
var G__35562__$3 = (cljs.core.truth_((function (){var and__5000__auto__ = source__$1;
if(cljs.core.truth_(and__5000__auto__)){
return method;
} else {
return and__5000__auto__;
}
})())?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35562__$2,new cljs.core.Keyword("clojure.error","symbol","clojure.error/symbol",1544821994),(new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[source__$1,method],null))):G__35562__$2);
var G__35562__$4 = (cljs.core.truth_(type)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35562__$3,new cljs.core.Keyword("clojure.error","class","clojure.error/class",278435890),type):G__35562__$3);
if(cljs.core.truth_(message)){
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35562__$4,new cljs.core.Keyword("clojure.error","cause","clojure.error/cause",-1879175742),message);
} else {
return G__35562__$4;
}

break;
case "execution":
var vec__35570 = cljs.core.first(trace);
var source__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35570,(0),null);
var method = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35570,(1),null);
var file = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35570,(2),null);
var line = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__35570,(3),null);
var file__$1 = cljs.core.first(cljs.core.remove.cljs$core$IFn$_invoke$arity$2((function (p1__35515_SHARP_){
var or__5002__auto__ = (p1__35515_SHARP_ == null);
if(or__5002__auto__){
return or__5002__auto__;
} else {
var fexpr__35578 = new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, ["NO_SOURCE_PATH",null,"NO_SOURCE_FILE",null], null), null);
return (fexpr__35578.cljs$core$IFn$_invoke$arity$1 ? fexpr__35578.cljs$core$IFn$_invoke$arity$1(p1__35515_SHARP_) : fexpr__35578.call(null, p1__35515_SHARP_));
}
}),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"file","file",-1269645878).cljs$core$IFn$_invoke$arity$1(caller),file], null)));
var err_line = (function (){var or__5002__auto__ = new cljs.core.Keyword(null,"line","line",212345235).cljs$core$IFn$_invoke$arity$1(caller);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return line;
}
})();
var G__35602 = new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword("clojure.error","class","clojure.error/class",278435890),type], null);
var G__35602__$1 = (cljs.core.truth_(err_line)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35602,new cljs.core.Keyword("clojure.error","line","clojure.error/line",-1816287471),err_line):G__35602);
var G__35602__$2 = (cljs.core.truth_(message)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35602__$1,new cljs.core.Keyword("clojure.error","cause","clojure.error/cause",-1879175742),message):G__35602__$1);
var G__35602__$3 = (cljs.core.truth_((function (){var or__5002__auto__ = fn;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
var and__5000__auto__ = source__$1;
if(cljs.core.truth_(and__5000__auto__)){
return method;
} else {
return and__5000__auto__;
}
}
})())?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35602__$2,new cljs.core.Keyword("clojure.error","symbol","clojure.error/symbol",1544821994),(function (){var or__5002__auto__ = fn;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[source__$1,method],null));
}
})()):G__35602__$2);
var G__35602__$4 = (cljs.core.truth_(file__$1)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35602__$3,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397),file__$1):G__35602__$3);
if(cljs.core.truth_(problems)){
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__35602__$4,new cljs.core.Keyword("clojure.error","spec","clojure.error/spec",2055032595),data);
} else {
return G__35602__$4;
}

break;
default:
throw (new Error(["No matching clause: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(G__35532__$1)].join('')));

}
})(),new cljs.core.Keyword("clojure.error","phase","clojure.error/phase",275140358),phase);
});
/**
 * Returns a string from exception data, as produced by ex-triage.
 *   The first line summarizes the exception phase and location.
 *   The subsequent lines describe the cause.
 */
cljs.repl.ex_str = (function cljs$repl$ex_str(p__35636){
var map__35637 = p__35636;
var map__35637__$1 = cljs.core.__destructure_map(map__35637);
var triage_data = map__35637__$1;
var phase = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","phase","clojure.error/phase",275140358));
var source = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","source","clojure.error/source",-2011936397));
var line = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","line","clojure.error/line",-1816287471));
var column = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","column","clojure.error/column",304721553));
var symbol = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","symbol","clojure.error/symbol",1544821994));
var class$ = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","class","clojure.error/class",278435890));
var cause = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","cause","clojure.error/cause",-1879175742));
var spec = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__35637__$1,new cljs.core.Keyword("clojure.error","spec","clojure.error/spec",2055032595));
var loc = [cljs.core.str.cljs$core$IFn$_invoke$arity$1((function (){var or__5002__auto__ = source;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return "<cljs repl>";
}
})()),":",cljs.core.str.cljs$core$IFn$_invoke$arity$1((function (){var or__5002__auto__ = line;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (1);
}
})()),(cljs.core.truth_(column)?[":",cljs.core.str.cljs$core$IFn$_invoke$arity$1(column)].join(''):"")].join('');
var class_name = cljs.core.name((function (){var or__5002__auto__ = class$;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return "";
}
})());
var simple_class = class_name;
var cause_type = ((cljs.core.contains_QMARK_(new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, ["RuntimeException",null,"Exception",null], null), null),simple_class))?"":[" (",simple_class,")"].join(''));
var format = goog.string.format;
var G__35643 = phase;
var G__35643__$1 = (((G__35643 instanceof cljs.core.Keyword))?G__35643.fqn:null);
switch (G__35643__$1) {
case "read-source":
return (format.cljs$core$IFn$_invoke$arity$3 ? format.cljs$core$IFn$_invoke$arity$3("Syntax error reading source at (%s).\n%s\n",loc,cause) : format.call(null, "Syntax error reading source at (%s).\n%s\n",loc,cause));

break;
case "macro-syntax-check":
var G__35644 = "Syntax error macroexpanding %sat (%s).\n%s";
var G__35645 = (cljs.core.truth_(symbol)?[cljs.core.str.cljs$core$IFn$_invoke$arity$1(symbol)," "].join(''):"");
var G__35646 = loc;
var G__35647 = (cljs.core.truth_(spec)?(function (){var sb__5647__auto__ = (new goog.string.StringBuffer());
var _STAR_print_newline_STAR__orig_val__35648_36009 = cljs.core._STAR_print_newline_STAR_;
var _STAR_print_fn_STAR__orig_val__35649_36010 = cljs.core._STAR_print_fn_STAR_;
var _STAR_print_newline_STAR__temp_val__35650_36011 = true;
var _STAR_print_fn_STAR__temp_val__35651_36012 = (function (x__5648__auto__){
return sb__5647__auto__.append(x__5648__auto__);
});
(cljs.core._STAR_print_newline_STAR_ = _STAR_print_newline_STAR__temp_val__35650_36011);

(cljs.core._STAR_print_fn_STAR_ = _STAR_print_fn_STAR__temp_val__35651_36012);

try{cljs.spec.alpha.explain_out(cljs.core.update.cljs$core$IFn$_invoke$arity$3(spec,new cljs.core.Keyword("cljs.spec.alpha","problems","cljs.spec.alpha/problems",447400814),(function (probs){
return cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p1__35634_SHARP_){
return cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(p1__35634_SHARP_,new cljs.core.Keyword(null,"in","in",-1531184865));
}),probs);
}))
);
}finally {(cljs.core._STAR_print_fn_STAR_ = _STAR_print_fn_STAR__orig_val__35649_36010);

(cljs.core._STAR_print_newline_STAR_ = _STAR_print_newline_STAR__orig_val__35648_36009);
}
return cljs.core.str.cljs$core$IFn$_invoke$arity$1(sb__5647__auto__);
})():(format.cljs$core$IFn$_invoke$arity$2 ? format.cljs$core$IFn$_invoke$arity$2("%s\n",cause) : format.call(null, "%s\n",cause)));
return (format.cljs$core$IFn$_invoke$arity$4 ? format.cljs$core$IFn$_invoke$arity$4(G__35644,G__35645,G__35646,G__35647) : format.call(null, G__35644,G__35645,G__35646,G__35647));

break;
case "macroexpansion":
var G__35658 = "Unexpected error%s macroexpanding %sat (%s).\n%s\n";
var G__35659 = cause_type;
var G__35660 = (cljs.core.truth_(symbol)?[cljs.core.str.cljs$core$IFn$_invoke$arity$1(symbol)," "].join(''):"");
var G__35661 = loc;
var G__35662 = cause;
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5(G__35658,G__35659,G__35660,G__35661,G__35662) : format.call(null, G__35658,G__35659,G__35660,G__35661,G__35662));

break;
case "compile-syntax-check":
var G__35663 = "Syntax error%s compiling %sat (%s).\n%s\n";
var G__35664 = cause_type;
var G__35665 = (cljs.core.truth_(symbol)?[cljs.core.str.cljs$core$IFn$_invoke$arity$1(symbol)," "].join(''):"");
var G__35666 = loc;
var G__35667 = cause;
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5(G__35663,G__35664,G__35665,G__35666,G__35667) : format.call(null, G__35663,G__35664,G__35665,G__35666,G__35667));

break;
case "compilation":
var G__35668 = "Unexpected error%s compiling %sat (%s).\n%s\n";
var G__35669 = cause_type;
var G__35670 = (cljs.core.truth_(symbol)?[cljs.core.str.cljs$core$IFn$_invoke$arity$1(symbol)," "].join(''):"");
var G__35671 = loc;
var G__35672 = cause;
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5(G__35668,G__35669,G__35670,G__35671,G__35672) : format.call(null, G__35668,G__35669,G__35670,G__35671,G__35672));

break;
case "read-eval-result":
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5("Error reading eval result%s at %s (%s).\n%s\n",cause_type,symbol,loc,cause) : format.call(null, "Error reading eval result%s at %s (%s).\n%s\n",cause_type,symbol,loc,cause));

break;
case "print-eval-result":
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5("Error printing return value%s at %s (%s).\n%s\n",cause_type,symbol,loc,cause) : format.call(null, "Error printing return value%s at %s (%s).\n%s\n",cause_type,symbol,loc,cause));

break;
case "execution":
if(cljs.core.truth_(spec)){
var G__35673 = "Execution error - invalid arguments to %s at (%s).\n%s";
var G__35674 = symbol;
var G__35675 = loc;
var G__35676 = (function (){var sb__5647__auto__ = (new goog.string.StringBuffer());
var _STAR_print_newline_STAR__orig_val__35677_36051 = cljs.core._STAR_print_newline_STAR_;
var _STAR_print_fn_STAR__orig_val__35678_36052 = cljs.core._STAR_print_fn_STAR_;
var _STAR_print_newline_STAR__temp_val__35679_36053 = true;
var _STAR_print_fn_STAR__temp_val__35680_36054 = (function (x__5648__auto__){
return sb__5647__auto__.append(x__5648__auto__);
});
(cljs.core._STAR_print_newline_STAR_ = _STAR_print_newline_STAR__temp_val__35679_36053);

(cljs.core._STAR_print_fn_STAR_ = _STAR_print_fn_STAR__temp_val__35680_36054);

try{cljs.spec.alpha.explain_out(cljs.core.update.cljs$core$IFn$_invoke$arity$3(spec,new cljs.core.Keyword("cljs.spec.alpha","problems","cljs.spec.alpha/problems",447400814),(function (probs){
return cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p1__35635_SHARP_){
return cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(p1__35635_SHARP_,new cljs.core.Keyword(null,"in","in",-1531184865));
}),probs);
}))
);
}finally {(cljs.core._STAR_print_fn_STAR_ = _STAR_print_fn_STAR__orig_val__35678_36052);

(cljs.core._STAR_print_newline_STAR_ = _STAR_print_newline_STAR__orig_val__35677_36051);
}
return cljs.core.str.cljs$core$IFn$_invoke$arity$1(sb__5647__auto__);
})();
return (format.cljs$core$IFn$_invoke$arity$4 ? format.cljs$core$IFn$_invoke$arity$4(G__35673,G__35674,G__35675,G__35676) : format.call(null, G__35673,G__35674,G__35675,G__35676));
} else {
var G__35684 = "Execution error%s at %s(%s).\n%s\n";
var G__35685 = cause_type;
var G__35686 = (cljs.core.truth_(symbol)?[cljs.core.str.cljs$core$IFn$_invoke$arity$1(symbol)," "].join(''):"");
var G__35687 = loc;
var G__35688 = cause;
return (format.cljs$core$IFn$_invoke$arity$5 ? format.cljs$core$IFn$_invoke$arity$5(G__35684,G__35685,G__35686,G__35687,G__35688) : format.call(null, G__35684,G__35685,G__35686,G__35687,G__35688));
}

break;
default:
throw (new Error(["No matching clause: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(G__35643__$1)].join('')));

}
});
cljs.repl.error__GT_str = (function cljs$repl$error__GT_str(error){
return cljs.repl.ex_str(cljs.repl.ex_triage(cljs.repl.Error__GT_map(error)));
});

//# sourceMappingURL=cljs.repl.js.map
