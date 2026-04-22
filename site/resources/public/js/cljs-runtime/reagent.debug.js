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
var G__36961__delegate = function (args){
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(reagent.debug.warnings,cljs.core.update_in,new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"warn","warn",-436710552)], null),cljs.core.conj,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,args)], 0));
};
var G__36961 = function (var_args){
var args = null;
if (arguments.length > 0) {
var G__36962__i = 0, G__36962__a = new Array(arguments.length -  0);
while (G__36962__i < G__36962__a.length) {G__36962__a[G__36962__i] = arguments[G__36962__i + 0]; ++G__36962__i;}
  args = new cljs.core.IndexedSeq(G__36962__a,0,null);
} 
return G__36961__delegate.call(this,args);};
G__36961.cljs$lang$maxFixedArity = 0;
G__36961.cljs$lang$applyTo = (function (arglist__36963){
var args = cljs.core.seq(arglist__36963);
return G__36961__delegate(args);
});
G__36961.cljs$core$IFn$_invoke$arity$variadic = G__36961__delegate;
return G__36961;
})()
);

(o.error = (function() { 
var G__36968__delegate = function (args){
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(reagent.debug.warnings,cljs.core.update_in,new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"error","error",-978969032)], null),cljs.core.conj,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,args)], 0));
};
var G__36968 = function (var_args){
var args = null;
if (arguments.length > 0) {
var G__36972__i = 0, G__36972__a = new Array(arguments.length -  0);
while (G__36972__i < G__36972__a.length) {G__36972__a[G__36972__i] = arguments[G__36972__i + 0]; ++G__36972__i;}
  args = new cljs.core.IndexedSeq(G__36972__a,0,null);
} 
return G__36968__delegate.call(this,args);};
G__36968.cljs$lang$maxFixedArity = 0;
G__36968.cljs$lang$applyTo = (function (arglist__36977){
var args = cljs.core.seq(arglist__36977);
return G__36968__delegate(args);
});
G__36968.cljs$core$IFn$_invoke$arity$variadic = G__36968__delegate;
return G__36968;
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
