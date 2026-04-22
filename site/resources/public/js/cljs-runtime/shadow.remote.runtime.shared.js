goog.provide('shadow.remote.runtime.shared');
shadow.remote.runtime.shared.init_state = (function shadow$remote$runtime$shared$init_state(client_info){
return new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"extensions","extensions",-1103629196),cljs.core.PersistentArrayMap.EMPTY,new cljs.core.Keyword(null,"ops","ops",1237330063),cljs.core.PersistentArrayMap.EMPTY,new cljs.core.Keyword(null,"client-info","client-info",1958982504),client_info,new cljs.core.Keyword(null,"call-id-seq","call-id-seq",-1679248218),(0),new cljs.core.Keyword(null,"call-handlers","call-handlers",386605551),cljs.core.PersistentArrayMap.EMPTY], null);
});
shadow.remote.runtime.shared.now = (function shadow$remote$runtime$shared$now(){
return Date.now();
});
shadow.remote.runtime.shared.get_client_id = (function shadow$remote$runtime$shared$get_client_id(p__30739){
var map__30740 = p__30739;
var map__30740__$1 = cljs.core.__destructure_map(map__30740);
var runtime = map__30740__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30740__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
var or__5002__auto__ = new cljs.core.Keyword(null,"client-id","client-id",-464622140).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref));
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("runtime has no assigned runtime-id",new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"runtime","runtime",-1331573996),runtime], null));
}
});
shadow.remote.runtime.shared.relay_msg = (function shadow$remote$runtime$shared$relay_msg(runtime,msg){
var self_id_31049 = shadow.remote.runtime.shared.get_client_id(runtime);
if(cljs.core.not_EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"to","to",192099007).cljs$core$IFn$_invoke$arity$1(msg),self_id_31049)){
shadow.remote.runtime.api.relay_msg(runtime,msg);
} else {
Promise.resolve((1)).then((function (){
var G__30748 = runtime;
var G__30749 = cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(msg,new cljs.core.Keyword(null,"from","from",1815293044),self_id_31049);
return (shadow.remote.runtime.shared.process.cljs$core$IFn$_invoke$arity$2 ? shadow.remote.runtime.shared.process.cljs$core$IFn$_invoke$arity$2(G__30748,G__30749) : shadow.remote.runtime.shared.process.call(null, G__30748,G__30749));
}));
}

return msg;
});
shadow.remote.runtime.shared.reply = (function shadow$remote$runtime$shared$reply(runtime,p__30751,res){
var map__30753 = p__30751;
var map__30753__$1 = cljs.core.__destructure_map(map__30753);
var call_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30753__$1,new cljs.core.Keyword(null,"call-id","call-id",1043012968));
var from = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30753__$1,new cljs.core.Keyword(null,"from","from",1815293044));
var res__$1 = (function (){var G__30762 = res;
var G__30762__$1 = (cljs.core.truth_(call_id)?cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__30762,new cljs.core.Keyword(null,"call-id","call-id",1043012968),call_id):G__30762);
if(cljs.core.truth_(from)){
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(G__30762__$1,new cljs.core.Keyword(null,"to","to",192099007),from);
} else {
return G__30762__$1;
}
})();
return shadow.remote.runtime.api.relay_msg(runtime,res__$1);
});
shadow.remote.runtime.shared.call = (function shadow$remote$runtime$shared$call(var_args){
var G__30769 = arguments.length;
switch (G__30769) {
case 3:
return shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
case 4:
return shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$3 = (function (runtime,msg,handlers){
return shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$4(runtime,msg,handlers,(0));
}));

(shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$4 = (function (p__30771,msg,handlers,timeout_after_ms){
var map__30772 = p__30771;
var map__30772__$1 = cljs.core.__destructure_map(map__30772);
var runtime = map__30772__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30772__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
if(cljs.core.map_QMARK_(msg)){
} else {
throw (new Error("Assert failed: (map? msg)"));
}

if(cljs.core.map_QMARK_(handlers)){
} else {
throw (new Error("Assert failed: (map? handlers)"));
}

if(cljs.core.nat_int_QMARK_(timeout_after_ms)){
} else {
throw (new Error("Assert failed: (nat-int? timeout-after-ms)"));
}

var call_id = new cljs.core.Keyword(null,"call-id-seq","call-id-seq",-1679248218).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref));
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(state_ref,cljs.core.update,new cljs.core.Keyword(null,"call-id-seq","call-id-seq",-1679248218),cljs.core.inc);

cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(state_ref,cljs.core.assoc_in,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"call-handlers","call-handlers",386605551),call_id], null),new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"handlers","handlers",79528781),handlers,new cljs.core.Keyword(null,"called-at","called-at",607081160),shadow.remote.runtime.shared.now(),new cljs.core.Keyword(null,"msg","msg",-1386103444),msg,new cljs.core.Keyword(null,"timeout","timeout",-318625318),timeout_after_ms], null));

return shadow.remote.runtime.api.relay_msg(runtime,cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(msg,new cljs.core.Keyword(null,"call-id","call-id",1043012968),call_id));
}));

(shadow.remote.runtime.shared.call.cljs$lang$maxFixedArity = 4);

shadow.remote.runtime.shared.trigger_BANG_ = (function shadow$remote$runtime$shared$trigger_BANG_(var_args){
var args__5732__auto__ = [];
var len__5726__auto___31090 = arguments.length;
var i__5727__auto___31091 = (0);
while(true){
if((i__5727__auto___31091 < len__5726__auto___31090)){
args__5732__auto__.push((arguments[i__5727__auto___31091]));

var G__31092 = (i__5727__auto___31091 + (1));
i__5727__auto___31091 = G__31092;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((2) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((2)),(0),null)):null);
return shadow.remote.runtime.shared.trigger_BANG_.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),(arguments[(1)]),argseq__5733__auto__);
});

(shadow.remote.runtime.shared.trigger_BANG_.cljs$core$IFn$_invoke$arity$variadic = (function (p__30778,ev,args){
var map__30779 = p__30778;
var map__30779__$1 = cljs.core.__destructure_map(map__30779);
var runtime = map__30779__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30779__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
var seq__30780 = cljs.core.seq(cljs.core.vals(new cljs.core.Keyword(null,"extensions","extensions",-1103629196).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref))));
var chunk__30783 = null;
var count__30784 = (0);
var i__30785 = (0);
while(true){
if((i__30785 < count__30784)){
var ext = chunk__30783.cljs$core$IIndexed$_nth$arity$2(null, i__30785);
var ev_fn = cljs.core.get.cljs$core$IFn$_invoke$arity$2(ext,ev);
if(cljs.core.truth_(ev_fn)){
cljs.core.apply.cljs$core$IFn$_invoke$arity$2(ev_fn,args);


var G__31094 = seq__30780;
var G__31095 = chunk__30783;
var G__31096 = count__30784;
var G__31097 = (i__30785 + (1));
seq__30780 = G__31094;
chunk__30783 = G__31095;
count__30784 = G__31096;
i__30785 = G__31097;
continue;
} else {
var G__31101 = seq__30780;
var G__31102 = chunk__30783;
var G__31103 = count__30784;
var G__31104 = (i__30785 + (1));
seq__30780 = G__31101;
chunk__30783 = G__31102;
count__30784 = G__31103;
i__30785 = G__31104;
continue;
}
} else {
var temp__5825__auto__ = cljs.core.seq(seq__30780);
if(temp__5825__auto__){
var seq__30780__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__30780__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__30780__$1);
var G__31106 = cljs.core.chunk_rest(seq__30780__$1);
var G__31107 = c__5525__auto__;
var G__31108 = cljs.core.count(c__5525__auto__);
var G__31109 = (0);
seq__30780 = G__31106;
chunk__30783 = G__31107;
count__30784 = G__31108;
i__30785 = G__31109;
continue;
} else {
var ext = cljs.core.first(seq__30780__$1);
var ev_fn = cljs.core.get.cljs$core$IFn$_invoke$arity$2(ext,ev);
if(cljs.core.truth_(ev_fn)){
cljs.core.apply.cljs$core$IFn$_invoke$arity$2(ev_fn,args);


var G__31115 = cljs.core.next(seq__30780__$1);
var G__31116 = null;
var G__31117 = (0);
var G__31118 = (0);
seq__30780 = G__31115;
chunk__30783 = G__31116;
count__30784 = G__31117;
i__30785 = G__31118;
continue;
} else {
var G__31119 = cljs.core.next(seq__30780__$1);
var G__31120 = null;
var G__31121 = (0);
var G__31122 = (0);
seq__30780 = G__31119;
chunk__30783 = G__31120;
count__30784 = G__31121;
i__30785 = G__31122;
continue;
}
}
} else {
return null;
}
}
break;
}
}));

(shadow.remote.runtime.shared.trigger_BANG_.cljs$lang$maxFixedArity = (2));

/** @this {Function} */
(shadow.remote.runtime.shared.trigger_BANG_.cljs$lang$applyTo = (function (seq30773){
var G__30774 = cljs.core.first(seq30773);
var seq30773__$1 = cljs.core.next(seq30773);
var G__30775 = cljs.core.first(seq30773__$1);
var seq30773__$2 = cljs.core.next(seq30773__$1);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__30774,G__30775,seq30773__$2);
}));

shadow.remote.runtime.shared.welcome = (function shadow$remote$runtime$shared$welcome(p__30795,p__30796){
var map__30799 = p__30795;
var map__30799__$1 = cljs.core.__destructure_map(map__30799);
var runtime = map__30799__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30799__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
var map__30802 = p__30796;
var map__30802__$1 = cljs.core.__destructure_map(map__30802);
var msg = map__30802__$1;
var client_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30802__$1,new cljs.core.Keyword(null,"client-id","client-id",-464622140));
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(state_ref,cljs.core.assoc,new cljs.core.Keyword(null,"client-id","client-id",-464622140),client_id,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"welcome","welcome",-578152123),true], 0));

var map__30804 = cljs.core.deref(state_ref);
var map__30804__$1 = cljs.core.__destructure_map(map__30804);
var client_info = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30804__$1,new cljs.core.Keyword(null,"client-info","client-info",1958982504));
var extensions = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30804__$1,new cljs.core.Keyword(null,"extensions","extensions",-1103629196));
shadow.remote.runtime.shared.relay_msg(runtime,new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"op","op",-1882987955),new cljs.core.Keyword(null,"hello","hello",-245025397),new cljs.core.Keyword(null,"client-info","client-info",1958982504),client_info], null));

return shadow.remote.runtime.shared.trigger_BANG_(runtime,new cljs.core.Keyword(null,"on-welcome","on-welcome",1895317125));
});
shadow.remote.runtime.shared.ping = (function shadow$remote$runtime$shared$ping(runtime,msg){
return shadow.remote.runtime.shared.reply(runtime,msg,new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"op","op",-1882987955),new cljs.core.Keyword(null,"pong","pong",-172484958)], null));
});
shadow.remote.runtime.shared.request_supported_ops = (function shadow$remote$runtime$shared$request_supported_ops(p__30821,msg){
var map__30822 = p__30821;
var map__30822__$1 = cljs.core.__destructure_map(map__30822);
var runtime = map__30822__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30822__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
return shadow.remote.runtime.shared.reply(runtime,msg,new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"op","op",-1882987955),new cljs.core.Keyword(null,"supported-ops","supported-ops",337914702),new cljs.core.Keyword(null,"ops","ops",1237330063),cljs.core.disj.cljs$core$IFn$_invoke$arity$variadic(cljs.core.set(cljs.core.keys(new cljs.core.Keyword(null,"ops","ops",1237330063).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref)))),new cljs.core.Keyword(null,"welcome","welcome",-578152123),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"unknown-relay-op","unknown-relay-op",170832753),new cljs.core.Keyword(null,"unknown-op","unknown-op",1900385996),new cljs.core.Keyword(null,"request-supported-ops","request-supported-ops",-1034994502),new cljs.core.Keyword(null,"tool-disconnect","tool-disconnect",189103996)], 0))], null));
});
shadow.remote.runtime.shared.unknown_relay_op = (function shadow$remote$runtime$shared$unknown_relay_op(msg){
return console.warn("unknown-relay-op",msg);
});
shadow.remote.runtime.shared.unknown_op = (function shadow$remote$runtime$shared$unknown_op(msg){
return console.warn("unknown-op",msg);
});
shadow.remote.runtime.shared.add_extension_STAR_ = (function shadow$remote$runtime$shared$add_extension_STAR_(p__30830,key,p__30831){
var map__30832 = p__30830;
var map__30832__$1 = cljs.core.__destructure_map(map__30832);
var state = map__30832__$1;
var extensions = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30832__$1,new cljs.core.Keyword(null,"extensions","extensions",-1103629196));
var map__30833 = p__30831;
var map__30833__$1 = cljs.core.__destructure_map(map__30833);
var spec = map__30833__$1;
var ops = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30833__$1,new cljs.core.Keyword(null,"ops","ops",1237330063));
var transit_write_handlers = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30833__$1,new cljs.core.Keyword(null,"transit-write-handlers","transit-write-handlers",1886308716));
if(cljs.core.contains_QMARK_(extensions,key)){
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("extension already registered",new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"key","key",-1516042587),key,new cljs.core.Keyword(null,"spec","spec",347520401),spec], null));
} else {
}

return cljs.core.reduce_kv((function (state__$1,op_kw,op_handler){
if(cljs.core.truth_(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(state__$1,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ops","ops",1237330063),op_kw], null)))){
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("op already registered",new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"key","key",-1516042587),key,new cljs.core.Keyword(null,"op","op",-1882987955),op_kw], null));
} else {
}

return cljs.core.assoc_in(state__$1,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ops","ops",1237330063),op_kw], null),op_handler);
}),cljs.core.assoc_in(state,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"extensions","extensions",-1103629196),key], null),spec),ops);
});
shadow.remote.runtime.shared.add_extension = (function shadow$remote$runtime$shared$add_extension(p__30855,key,spec){
var map__30856 = p__30855;
var map__30856__$1 = cljs.core.__destructure_map(map__30856);
var runtime = map__30856__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30856__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(state_ref,shadow.remote.runtime.shared.add_extension_STAR_,key,spec);

var temp__5829__auto___31131 = new cljs.core.Keyword(null,"on-welcome","on-welcome",1895317125).cljs$core$IFn$_invoke$arity$1(spec);
if((temp__5829__auto___31131 == null)){
} else {
var on_welcome_31132 = temp__5829__auto___31131;
if(cljs.core.truth_(new cljs.core.Keyword(null,"welcome","welcome",-578152123).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref)))){
(on_welcome_31132.cljs$core$IFn$_invoke$arity$0 ? on_welcome_31132.cljs$core$IFn$_invoke$arity$0() : on_welcome_31132.call(null, ));
} else {
}
}

return runtime;
});
shadow.remote.runtime.shared.add_defaults = (function shadow$remote$runtime$shared$add_defaults(runtime){
return shadow.remote.runtime.shared.add_extension(runtime,new cljs.core.Keyword("shadow.remote.runtime.shared","defaults","shadow.remote.runtime.shared/defaults",-1821257543),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"ops","ops",1237330063),new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"welcome","welcome",-578152123),(function (p1__30863_SHARP_){
return shadow.remote.runtime.shared.welcome(runtime,p1__30863_SHARP_);
}),new cljs.core.Keyword(null,"unknown-relay-op","unknown-relay-op",170832753),(function (p1__30868_SHARP_){
return shadow.remote.runtime.shared.unknown_relay_op(p1__30868_SHARP_);
}),new cljs.core.Keyword(null,"unknown-op","unknown-op",1900385996),(function (p1__30869_SHARP_){
return shadow.remote.runtime.shared.unknown_op(p1__30869_SHARP_);
}),new cljs.core.Keyword(null,"ping","ping",-1670114784),(function (p1__30870_SHARP_){
return shadow.remote.runtime.shared.ping(runtime,p1__30870_SHARP_);
}),new cljs.core.Keyword(null,"request-supported-ops","request-supported-ops",-1034994502),(function (p1__30871_SHARP_){
return shadow.remote.runtime.shared.request_supported_ops(runtime,p1__30871_SHARP_);
})], null)], null));
});
shadow.remote.runtime.shared.del_extension_STAR_ = (function shadow$remote$runtime$shared$del_extension_STAR_(state,key){
var ext = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(state,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"extensions","extensions",-1103629196),key], null));
if(cljs.core.not(ext)){
return state;
} else {
return cljs.core.reduce_kv((function (state__$1,op_kw,op_handler){
return cljs.core.update_in.cljs$core$IFn$_invoke$arity$4(state__$1,new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ops","ops",1237330063)], null),cljs.core.dissoc,op_kw);
}),cljs.core.update.cljs$core$IFn$_invoke$arity$4(state,new cljs.core.Keyword(null,"extensions","extensions",-1103629196),cljs.core.dissoc,key),new cljs.core.Keyword(null,"ops","ops",1237330063).cljs$core$IFn$_invoke$arity$1(ext));
}
});
shadow.remote.runtime.shared.del_extension = (function shadow$remote$runtime$shared$del_extension(p__30884,key){
var map__30885 = p__30884;
var map__30885__$1 = cljs.core.__destructure_map(map__30885);
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30885__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(state_ref,shadow.remote.runtime.shared.del_extension_STAR_,key);
});
shadow.remote.runtime.shared.unhandled_call_result = (function shadow$remote$runtime$shared$unhandled_call_result(call_config,msg){
return console.warn("unhandled call result",msg,call_config);
});
shadow.remote.runtime.shared.unhandled_client_not_found = (function shadow$remote$runtime$shared$unhandled_client_not_found(p__30888,msg){
var map__30889 = p__30888;
var map__30889__$1 = cljs.core.__destructure_map(map__30889);
var runtime = map__30889__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30889__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
return shadow.remote.runtime.shared.trigger_BANG_.cljs$core$IFn$_invoke$arity$variadic(runtime,new cljs.core.Keyword(null,"on-client-not-found","on-client-not-found",-642452849),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([msg], 0));
});
shadow.remote.runtime.shared.reply_unknown_op = (function shadow$remote$runtime$shared$reply_unknown_op(runtime,msg){
return shadow.remote.runtime.shared.reply(runtime,msg,new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"op","op",-1882987955),new cljs.core.Keyword(null,"unknown-op","unknown-op",1900385996),new cljs.core.Keyword(null,"msg","msg",-1386103444),msg], null));
});
shadow.remote.runtime.shared.process = (function shadow$remote$runtime$shared$process(p__30911,p__30912){
var map__30914 = p__30911;
var map__30914__$1 = cljs.core.__destructure_map(map__30914);
var runtime = map__30914__$1;
var state_ref = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30914__$1,new cljs.core.Keyword(null,"state-ref","state-ref",2127874952));
var map__30917 = p__30912;
var map__30917__$1 = cljs.core.__destructure_map(map__30917);
var msg = map__30917__$1;
var op = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30917__$1,new cljs.core.Keyword(null,"op","op",-1882987955));
var call_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30917__$1,new cljs.core.Keyword(null,"call-id","call-id",1043012968));
var state = cljs.core.deref(state_ref);
var op_handler = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(state,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ops","ops",1237330063),op], null));
if(cljs.core.truth_(call_id)){
var cfg = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(state,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"call-handlers","call-handlers",386605551),call_id], null));
var call_handler = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(cfg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"handlers","handlers",79528781),op], null));
if(cljs.core.truth_(call_handler)){
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$variadic(state_ref,cljs.core.update,new cljs.core.Keyword(null,"call-handlers","call-handlers",386605551),cljs.core.dissoc,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([call_id], 0));

return (call_handler.cljs$core$IFn$_invoke$arity$1 ? call_handler.cljs$core$IFn$_invoke$arity$1(msg) : call_handler.call(null, msg));
} else {
if(cljs.core.truth_(op_handler)){
return (op_handler.cljs$core$IFn$_invoke$arity$1 ? op_handler.cljs$core$IFn$_invoke$arity$1(msg) : op_handler.call(null, msg));
} else {
return shadow.remote.runtime.shared.unhandled_call_result(cfg,msg);

}
}
} else {
if(cljs.core.truth_(op_handler)){
return (op_handler.cljs$core$IFn$_invoke$arity$1 ? op_handler.cljs$core$IFn$_invoke$arity$1(msg) : op_handler.call(null, msg));
} else {
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"client-not-found","client-not-found",-1754042614),op)){
return shadow.remote.runtime.shared.unhandled_client_not_found(runtime,msg);
} else {
return shadow.remote.runtime.shared.reply_unknown_op(runtime,msg);

}
}
}
});
shadow.remote.runtime.shared.run_on_idle = (function shadow$remote$runtime$shared$run_on_idle(state_ref){
var seq__30940 = cljs.core.seq(cljs.core.vals(new cljs.core.Keyword(null,"extensions","extensions",-1103629196).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(state_ref))));
var chunk__30942 = null;
var count__30943 = (0);
var i__30944 = (0);
while(true){
if((i__30944 < count__30943)){
var map__30978 = chunk__30942.cljs$core$IIndexed$_nth$arity$2(null, i__30944);
var map__30978__$1 = cljs.core.__destructure_map(map__30978);
var on_idle = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30978__$1,new cljs.core.Keyword(null,"on-idle","on-idle",2044706602));
if(cljs.core.truth_(on_idle)){
(on_idle.cljs$core$IFn$_invoke$arity$0 ? on_idle.cljs$core$IFn$_invoke$arity$0() : on_idle.call(null, ));


var G__31202 = seq__30940;
var G__31203 = chunk__30942;
var G__31204 = count__30943;
var G__31205 = (i__30944 + (1));
seq__30940 = G__31202;
chunk__30942 = G__31203;
count__30943 = G__31204;
i__30944 = G__31205;
continue;
} else {
var G__31206 = seq__30940;
var G__31207 = chunk__30942;
var G__31208 = count__30943;
var G__31209 = (i__30944 + (1));
seq__30940 = G__31206;
chunk__30942 = G__31207;
count__30943 = G__31208;
i__30944 = G__31209;
continue;
}
} else {
var temp__5825__auto__ = cljs.core.seq(seq__30940);
if(temp__5825__auto__){
var seq__30940__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__30940__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__30940__$1);
var G__31210 = cljs.core.chunk_rest(seq__30940__$1);
var G__31211 = c__5525__auto__;
var G__31212 = cljs.core.count(c__5525__auto__);
var G__31213 = (0);
seq__30940 = G__31210;
chunk__30942 = G__31211;
count__30943 = G__31212;
i__30944 = G__31213;
continue;
} else {
var map__30985 = cljs.core.first(seq__30940__$1);
var map__30985__$1 = cljs.core.__destructure_map(map__30985);
var on_idle = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__30985__$1,new cljs.core.Keyword(null,"on-idle","on-idle",2044706602));
if(cljs.core.truth_(on_idle)){
(on_idle.cljs$core$IFn$_invoke$arity$0 ? on_idle.cljs$core$IFn$_invoke$arity$0() : on_idle.call(null, ));


var G__31214 = cljs.core.next(seq__30940__$1);
var G__31215 = null;
var G__31216 = (0);
var G__31217 = (0);
seq__30940 = G__31214;
chunk__30942 = G__31215;
count__30943 = G__31216;
i__30944 = G__31217;
continue;
} else {
var G__31218 = cljs.core.next(seq__30940__$1);
var G__31219 = null;
var G__31220 = (0);
var G__31221 = (0);
seq__30940 = G__31218;
chunk__30942 = G__31219;
count__30943 = G__31220;
i__30944 = G__31221;
continue;
}
}
} else {
return null;
}
}
break;
}
});

//# sourceMappingURL=shadow.remote.runtime.shared.js.map
