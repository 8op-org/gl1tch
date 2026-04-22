goog.provide('garden.util');
/**
 * Formats a string using goog.string.format.
 */
garden.util.format = (function garden$util$format(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40008 = arguments.length;
var i__5727__auto___40009 = (0);
while(true){
if((i__5727__auto___40009 < len__5726__auto___40008)){
args__5732__auto__.push((arguments[i__5727__auto___40009]));

var G__40011 = (i__5727__auto___40009 + (1));
i__5727__auto___40009 = G__40011;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.util.format.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.util.format.cljs$core$IFn$_invoke$arity$variadic = (function (fmt,args){
return cljs.core.apply.cljs$core$IFn$_invoke$arity$3(goog.string.format,fmt,args);
}));

(garden.util.format.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.util.format.cljs$lang$applyTo = (function (seq39896){
var G__39897 = cljs.core.first(seq39896);
var seq39896__$1 = cljs.core.next(seq39896);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__39897,seq39896__$1);
}));


/**
 * @interface
 */
garden.util.ToString = function(){};

var garden$util$ToString$to_str$dyn_40022 = (function (this$){
var x__5350__auto__ = (((this$ == null))?null:this$);
var m__5351__auto__ = (garden.util.to_str[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5351__auto__.call(null, this$));
} else {
var m__5349__auto__ = (garden.util.to_str["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5349__auto__.call(null, this$));
} else {
throw cljs.core.missing_protocol("ToString.to-str",this$);
}
}
});
/**
 * Convert a value into a string.
 */
garden.util.to_str = (function garden$util$to_str(this$){
if((((!((this$ == null)))) && ((!((this$.garden$util$ToString$to_str$arity$1 == null)))))){
return this$.garden$util$ToString$to_str$arity$1(this$);
} else {
return garden$util$ToString$to_str$dyn_40022(this$);
}
});

(cljs.core.Keyword.prototype.garden$util$ToString$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.Keyword.prototype.garden$util$ToString$to_str$arity$1 = (function (this$){
var this$__$1 = this;
return cljs.core.name(this$__$1);
}));

(garden.util.ToString["_"] = true);

(garden.util.to_str["_"] = (function (this$){
return cljs.core.str.cljs$core$IFn$_invoke$arity$1(this$);
}));

(garden.util.ToString["null"] = true);

(garden.util.to_str["null"] = (function (this$){
return "";
}));
/**
 * Convert a variable number of values into strings.
 */
garden.util.as_str = (function garden$util$as_str(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40037 = arguments.length;
var i__5727__auto___40039 = (0);
while(true){
if((i__5727__auto___40039 < len__5726__auto___40037)){
args__5732__auto__.push((arguments[i__5727__auto___40039]));

var G__40048 = (i__5727__auto___40039 + (1));
i__5727__auto___40039 = G__40048;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((0) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((0)),(0),null)):null);
return garden.util.as_str.cljs$core$IFn$_invoke$arity$variadic(argseq__5733__auto__);
});

(garden.util.as_str.cljs$core$IFn$_invoke$arity$variadic = (function (args){
return cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,cljs.core.map.cljs$core$IFn$_invoke$arity$2(garden.util.to_str,args));
}));

(garden.util.as_str.cljs$lang$maxFixedArity = (0));

/** @this {Function} */
(garden.util.as_str.cljs$lang$applyTo = (function (seq39899){
var self__5712__auto__ = this;
return self__5712__auto__.cljs$core$IFn$_invoke$arity$variadic(cljs.core.seq(seq39899));
}));

/**
 * Convert a string to an integer with optional base.
 */
garden.util.string__GT_int = (function garden$util$string__GT_int(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40055 = arguments.length;
var i__5727__auto___40056 = (0);
while(true){
if((i__5727__auto___40056 < len__5726__auto___40055)){
args__5732__auto__.push((arguments[i__5727__auto___40056]));

var G__40057 = (i__5727__auto___40056 + (1));
i__5727__auto___40056 = G__40057;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.util.string__GT_int.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.util.string__GT_int.cljs$core$IFn$_invoke$arity$variadic = (function (s,p__39912){
var vec__39913 = p__39912;
var radix = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__39913,(0),null);
var radix__$1 = (function (){var or__5002__auto__ = radix;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (10);
}
})();
return parseInt(s,radix__$1);
}));

(garden.util.string__GT_int.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.util.string__GT_int.cljs$lang$applyTo = (function (seq39908){
var G__39909 = cljs.core.first(seq39908);
var seq39908__$1 = cljs.core.next(seq39908);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__39909,seq39908__$1);
}));

/**
 * Convert an integer to a string with optional base.
 */
garden.util.int__GT_string = (function garden$util$int__GT_string(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40060 = arguments.length;
var i__5727__auto___40061 = (0);
while(true){
if((i__5727__auto___40061 < len__5726__auto___40060)){
args__5732__auto__.push((arguments[i__5727__auto___40061]));

var G__40063 = (i__5727__auto___40061 + (1));
i__5727__auto___40061 = G__40063;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.util.int__GT_string.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.util.int__GT_string.cljs$core$IFn$_invoke$arity$variadic = (function (i,p__39919){
var vec__39920 = p__39919;
var radix = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__39920,(0),null);
var radix__$1 = (function (){var or__5002__auto__ = radix;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (10);
}
})();
return i.toString(radix__$1);
}));

(garden.util.int__GT_string.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.util.int__GT_string.cljs$lang$applyTo = (function (seq39917){
var G__39918 = cljs.core.first(seq39917);
var seq39917__$1 = cljs.core.next(seq39917);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__39918,seq39917__$1);
}));

/**
 * Return a space separated list of values.
 */
garden.util.space_join = (function garden$util$space_join(xs){
return clojure.string.join.cljs$core$IFn$_invoke$arity$2(" ",cljs.core.map.cljs$core$IFn$_invoke$arity$2(garden.util.to_str,xs));
});
/**
 * Return a comma separated list of values. Subsequences are joined with
 * spaces.
 */
garden.util.comma_join = (function garden$util$comma_join(xs){
var ys = (function (){var iter__5480__auto__ = (function garden$util$comma_join_$_iter__39963(s__39964){
return (new cljs.core.LazySeq(null,(function (){
var s__39964__$1 = s__39964;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__39964__$1);
if(temp__5825__auto__){
var s__39964__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__39964__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__39964__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__39966 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__39965 = (0);
while(true){
if((i__39965 < size__5479__auto__)){
var x = cljs.core._nth(c__5478__auto__,i__39965);
cljs.core.chunk_append(b__39966,((cljs.core.sequential_QMARK_(x))?garden.util.space_join(x):garden.util.to_str(x)));

var G__40072 = (i__39965 + (1));
i__39965 = G__40072;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__39966),garden$util$comma_join_$_iter__39963(cljs.core.chunk_rest(s__39964__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__39966),null);
}
} else {
var x = cljs.core.first(s__39964__$2);
return cljs.core.cons(((cljs.core.sequential_QMARK_(x))?garden.util.space_join(x):garden.util.to_str(x)),garden$util$comma_join_$_iter__39963(cljs.core.rest(s__39964__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(xs);
})();
return clojure.string.join.cljs$core$IFn$_invoke$arity$2(", ",ys);
});
/**
 * Wrap a string with double quotes.
 */
garden.util.wrap_quotes = (function garden$util$wrap_quotes(s){
return ["\"",cljs.core.str.cljs$core$IFn$_invoke$arity$1(s),"\""].join('');
});
/**
 * True if `(map? x)` and `x` does not satisfy `clojure.lang.IRecord`.
 */
garden.util.hash_map_QMARK_ = (function garden$util$hash_map_QMARK_(x){
return ((cljs.core.map_QMARK_(x)) && ((!(cljs.core.record_QMARK_(x)))));
});
/**
 * Alias to `vector?`.
 */
garden.util.rule_QMARK_ = cljs.core.vector_QMARK_;
/**
 * Alias to `hash-map?`.
 */
garden.util.declaration_QMARK_ = garden.util.hash_map_QMARK_;
garden.util.at_rule_QMARK_ = (function garden$util$at_rule_QMARK_(x){
return (x instanceof garden.types.CSSAtRule);
});
/**
 * True if `x` is a CSS `@media` rule.
 */
garden.util.at_media_QMARK_ = (function garden$util$at_media_QMARK_(x){
return ((garden.util.at_rule_QMARK_(x)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"identifier","identifier",-805503498).cljs$core$IFn$_invoke$arity$1(x),new cljs.core.Keyword(null,"media","media",-1066138403))));
});
/**
 * True if `x` is a CSS `@supports` rule.
 */
garden.util.at_supports_QMARK_ = (function garden$util$at_supports_QMARK_(x){
return ((garden.util.at_rule_QMARK_(x)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"identifier","identifier",-805503498).cljs$core$IFn$_invoke$arity$1(x),new cljs.core.Keyword(null,"feature","feature",27242652))));
});
/**
 * True if `x` is a CSS `@keyframes` rule.
 */
garden.util.at_keyframes_QMARK_ = (function garden$util$at_keyframes_QMARK_(x){
return ((garden.util.at_rule_QMARK_(x)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"identifier","identifier",-805503498).cljs$core$IFn$_invoke$arity$1(x),new cljs.core.Keyword(null,"keyframes","keyframes",-1437976012))));
});
/**
 * True if `x` is a CSS `@import` rule.
 */
garden.util.at_import_QMARK_ = (function garden$util$at_import_QMARK_(x){
return ((garden.util.at_rule_QMARK_(x)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"identifier","identifier",-805503498).cljs$core$IFn$_invoke$arity$1(x),new cljs.core.Keyword(null,"import","import",-1399500709))));
});
/**
 * Attach a CSS style prefix to s.
 */
garden.util.prefix = (function garden$util$prefix(p,s){
var p__$1 = garden.util.to_str(p);
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2("-",cljs.core.last(p__$1))){
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(p__$1),cljs.core.str.cljs$core$IFn$_invoke$arity$1(s)].join('');
} else {
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(p__$1),"-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(s)].join('');
}
});
/**
 * Attach a CSS vendor prefix to s.
 */
garden.util.vendor_prefix = (function garden$util$vendor_prefix(p,s){
var p__$1 = garden.util.to_str(p);
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2("-",cljs.core.first(p__$1))){
return garden.util.prefix(p__$1,s);
} else {
return garden.util.prefix(["-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(p__$1)].join(''),s);
}
});
/**
 * True if n is a natural number.
 */
garden.util.natural_QMARK_ = (function garden$util$natural_QMARK_(n){
return ((cljs.core.integer_QMARK_(n)) && ((n > (0))));
});
/**
 * True if n is a number between a and b.
 */
garden.util.between_QMARK_ = (function garden$util$between_QMARK_(n,a,b){
var bottom = (function (){var x__5090__auto__ = a;
var y__5091__auto__ = b;
return ((x__5090__auto__ < y__5091__auto__) ? x__5090__auto__ : y__5091__auto__);
})();
var top = (function (){var x__5087__auto__ = a;
var y__5088__auto__ = b;
return ((x__5087__auto__ > y__5088__auto__) ? x__5087__auto__ : y__5088__auto__);
})();
return (((n >= bottom)) && ((n <= top)));
});
/**
 * Return a number such that n is no less than a and no more than b.
 */
garden.util.clip = (function garden$util$clip(a,b,n){
var vec__39997 = (((a <= b))?new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [a,b], null):new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [b,a], null));
var a__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__39997,(0),null);
var b__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__39997,(1),null);
var x__5087__auto__ = a__$1;
var y__5088__auto__ = (function (){var x__5090__auto__ = b__$1;
var y__5091__auto__ = n;
return ((x__5090__auto__ < y__5091__auto__) ? x__5090__auto__ : y__5091__auto__);
})();
return ((x__5087__auto__ > y__5088__auto__) ? x__5087__auto__ : y__5088__auto__);
});
/**
 * Return the average of two or more numbers.
 */
garden.util.average = (function garden$util$average(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40079 = arguments.length;
var i__5727__auto___40080 = (0);
while(true){
if((i__5727__auto___40080 < len__5726__auto___40079)){
args__5732__auto__.push((arguments[i__5727__auto___40080]));

var G__40083 = (i__5727__auto___40080 + (1));
i__5727__auto___40080 = G__40083;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((2) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((2)),(0),null)):null);
return garden.util.average.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),(arguments[(1)]),argseq__5733__auto__);
});

(garden.util.average.cljs$core$IFn$_invoke$arity$variadic = (function (n,m,more){
return (cljs.core.apply.cljs$core$IFn$_invoke$arity$4(cljs.core._PLUS_,n,m,more) / (2.0 + cljs.core.count(more)));
}));

(garden.util.average.cljs$lang$maxFixedArity = (2));

/** @this {Function} */
(garden.util.average.cljs$lang$applyTo = (function (seq40000){
var G__40001 = cljs.core.first(seq40000);
var seq40000__$1 = cljs.core.next(seq40000);
var G__40002 = cljs.core.first(seq40000__$1);
var seq40000__$2 = cljs.core.next(seq40000__$1);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__40001,G__40002,seq40000__$2);
}));

/**
 * All the ways to take one item from each sequence.
 */
garden.util.cartesian_product = (function garden$util$cartesian_product(var_args){
var args__5732__auto__ = [];
var len__5726__auto___40085 = arguments.length;
var i__5727__auto___40086 = (0);
while(true){
if((i__5727__auto___40086 < len__5726__auto___40085)){
args__5732__auto__.push((arguments[i__5727__auto___40086]));

var G__40087 = (i__5727__auto___40086 + (1));
i__5727__auto___40086 = G__40087;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((0) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((0)),(0),null)):null);
return garden.util.cartesian_product.cljs$core$IFn$_invoke$arity$variadic(argseq__5733__auto__);
});

(garden.util.cartesian_product.cljs$core$IFn$_invoke$arity$variadic = (function (seqs){
var v_original_seqs = cljs.core.vec(seqs);
var step = (function garden$util$step(v_seqs){
var increment = (function (v_seqs__$1){
var i = (cljs.core.count(v_seqs__$1) - (1));
var v_seqs__$2 = v_seqs__$1;
while(true){
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(i,(-1))){
return null;
} else {
var temp__5823__auto__ = cljs.core.next((v_seqs__$2.cljs$core$IFn$_invoke$arity$1 ? v_seqs__$2.cljs$core$IFn$_invoke$arity$1(i) : v_seqs__$2.call(null, i)));
if(temp__5823__auto__){
var rst = temp__5823__auto__;
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(v_seqs__$2,i,rst);
} else {
var G__40088 = (i - (1));
var G__40089 = cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(v_seqs__$2,i,(v_original_seqs.cljs$core$IFn$_invoke$arity$1 ? v_original_seqs.cljs$core$IFn$_invoke$arity$1(i) : v_original_seqs.call(null, i)));
i = G__40088;
v_seqs__$2 = G__40089;
continue;
}
}
break;
}
});
if(cljs.core.truth_(v_seqs)){
return cljs.core.cons(cljs.core.map.cljs$core$IFn$_invoke$arity$2(cljs.core.first,v_seqs),(new cljs.core.LazySeq(null,(function (){
return garden$util$step(increment(v_seqs));
}),null,null)));
} else {
return null;
}
});
if(cljs.core.every_QMARK_(cljs.core.seq,seqs)){
return (new cljs.core.LazySeq(null,(function (){
return step(v_original_seqs);
}),null,null));
} else {
return null;
}
}));

(garden.util.cartesian_product.cljs$lang$maxFixedArity = (0));

/** @this {Function} */
(garden.util.cartesian_product.cljs$lang$applyTo = (function (seq40003){
var self__5712__auto__ = this;
return self__5712__auto__.cljs$core$IFn$_invoke$arity$variadic(cljs.core.seq(seq40003));
}));


//# sourceMappingURL=garden.util.js.map
