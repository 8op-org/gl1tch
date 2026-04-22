goog.provide('reagent.debug');
reagent.debug.has_console = (typeof console !== 'undefined');
reagent.debug.tracking = false;
if((typeof reagent !== 'undefined') && (typeof reagent.debug !== 'undefined') && (typeof reagent.debug.warnings !== 'undefined')){
} else {
reagent.debug.warnings = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(null);
}
if((typeof reagent !== 'undefined') && (typeof reagent.debug !== 'undefined') && (typeof reagent.debug.track_console !== 'undefined')){
} else {
reagent.debug.track_console = (function (){var o = ({});
(o.warn = (function() { 
var G__24635__delegate = function (args){
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(reagent.debug.warnings,cljs.core.update_in,new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"warn","warn",-436710552)], null),cljs.core.conj,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,args)], 0));
};
var G__24635 = function (var_args){
var args = null;
if (arguments.length > 0) {
var G__24637__i = 0, G__24637__a = new Array(arguments.length -  0);
while (G__24637__i < G__24637__a.length) {G__24637__a[G__24637__i] = arguments[G__24637__i + 0]; ++G__24637__i;}
  args = new cljs.core.IndexedSeq(G__24637__a,0,null);
} 
return G__24635__delegate.call(this,args);};
G__24635.cljs$lang$maxFixedArity = 0;
G__24635.cljs$lang$applyTo = (function (arglist__24639){
var args = cljs.core.seq(arglist__24639);
return G__24635__delegate(args);
});
G__24635.cljs$core$IFn$_invoke$arity$variadic = G__24635__delegate;
return G__24635;
})()
);

(o.error = (function() { 
var G__24640__delegate = function (args){
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(reagent.debug.warnings,cljs.core.update_in,new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"error","error",-978969032)], null),cljs.core.conj,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,args)], 0));
};
var G__24640 = function (var_args){
var args = null;
if (arguments.length > 0) {
var G__24641__i = 0, G__24641__a = new Array(arguments.length -  0);
while (G__24641__i < G__24641__a.length) {G__24641__a[G__24641__i] = arguments[G__24641__i + 0]; ++G__24641__i;}
  args = new cljs.core.IndexedSeq(G__24641__a,0,null);
} 
return G__24640__delegate.call(this,args);};
G__24640.cljs$lang$maxFixedArity = 0;
G__24640.cljs$lang$applyTo = (function (arglist__24642){
var args = cljs.core.seq(arglist__24642);
return G__24640__delegate(args);
});
G__24640.cljs$core$IFn$_invoke$arity$variadic = G__24640__delegate;
return G__24640;
})()
);

return o;
})();
}
reagent.debug.track_warnings = (function reagent$debug$track_warnings(f){
(reagent.debug.tracking = true);

cljs.core.reset_BANG_(reagent.debug.warnings,null);

(f.cljs$core$IFn$_invoke$arity$0 ? f.cljs$core$IFn$_invoke$arity$0() : f.call(null, ));

var warns = cljs.core.deref(reagent.debug.warnings);
cljs.core.reset_BANG_(reagent.debug.warnings,null);

(reagent.debug.tracking = false);

return warns;
});

//# sourceMappingURL=reagent.debug.js.map
