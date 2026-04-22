goog.provide('shadow.cljs.devtools.client.browser');
shadow.cljs.devtools.client.browser.devtools_msg = (function shadow$cljs$devtools$client$browser$devtools_msg(var_args){
var args__5732__auto__ = [];
var len__5726__auto___38066 = arguments.length;
var i__5727__auto___38067 = (0);
while(true){
if((i__5727__auto___38067 < len__5726__auto___38066)){
args__5732__auto__.push((arguments[i__5727__auto___38067]));

var G__38068 = (i__5727__auto___38067 + (1));
i__5727__auto___38067 = G__38068;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic = (function (msg,args){
if(shadow.cljs.devtools.client.env.log){
if(cljs.core.seq(shadow.cljs.devtools.client.env.log_style)){
return console.log.apply(console,cljs.core.into_array.cljs$core$IFn$_invoke$arity$1(cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [["%cshadow-cljs: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(msg)].join(''),shadow.cljs.devtools.client.env.log_style], null),args)));
} else {
return console.log.apply(console,cljs.core.into_array.cljs$core$IFn$_invoke$arity$1(cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [["shadow-cljs: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(msg)].join('')], null),args)));
}
} else {
return null;
}
}));

(shadow.cljs.devtools.client.browser.devtools_msg.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(shadow.cljs.devtools.client.browser.devtools_msg.cljs$lang$applyTo = (function (seq37452){
var G__37453 = cljs.core.first(seq37452);
var seq37452__$1 = cljs.core.next(seq37452);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__37453,seq37452__$1);
}));

shadow.cljs.devtools.client.browser.script_eval = (function shadow$cljs$devtools$client$browser$script_eval(code){
return goog.globalEval(code);
});
shadow.cljs.devtools.client.browser.do_js_load = (function shadow$cljs$devtools$client$browser$do_js_load(sources){
var seq__37454 = cljs.core.seq(sources);
var chunk__37455 = null;
var count__37456 = (0);
var i__37457 = (0);
while(true){
if((i__37457 < count__37456)){
var map__37472 = chunk__37455.cljs$core$IIndexed$_nth$arity$2(null, i__37457);
var map__37472__$1 = cljs.core.__destructure_map(map__37472);
var src = map__37472__$1;
var resource_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37472__$1,new cljs.core.Keyword(null,"resource-id","resource-id",-1308422582));
var output_name = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37472__$1,new cljs.core.Keyword(null,"output-name","output-name",-1769107767));
var resource_name = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37472__$1,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100));
var js = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37472__$1,new cljs.core.Keyword(null,"js","js",1768080579));
$CLJS.SHADOW_ENV.setLoaded(output_name);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load JS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([resource_name], 0));

shadow.cljs.devtools.client.env.before_load_src(src);

try{shadow.cljs.devtools.client.browser.script_eval([cljs.core.str.cljs$core$IFn$_invoke$arity$1(js),"\n//# sourceURL=",cljs.core.str.cljs$core$IFn$_invoke$arity$1($CLJS.SHADOW_ENV.scriptBase),cljs.core.str.cljs$core$IFn$_invoke$arity$1(output_name)].join(''));
}catch (e37479){var e_38076 = e37479;
if(shadow.cljs.devtools.client.env.log){
console.error(["Failed to load ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name)].join(''),e_38076);
} else {
}

throw (new Error(["Failed to load ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name),": ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(e_38076.message)].join('')));
}

var G__38077 = seq__37454;
var G__38078 = chunk__37455;
var G__38079 = count__37456;
var G__38080 = (i__37457 + (1));
seq__37454 = G__38077;
chunk__37455 = G__38078;
count__37456 = G__38079;
i__37457 = G__38080;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__37454);
if(temp__5825__auto__){
var seq__37454__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__37454__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__37454__$1);
var G__38081 = cljs.core.chunk_rest(seq__37454__$1);
var G__38082 = c__5525__auto__;
var G__38083 = cljs.core.count(c__5525__auto__);
var G__38084 = (0);
seq__37454 = G__38081;
chunk__37455 = G__38082;
count__37456 = G__38083;
i__37457 = G__38084;
continue;
} else {
var map__37480 = cljs.core.first(seq__37454__$1);
var map__37480__$1 = cljs.core.__destructure_map(map__37480);
var src = map__37480__$1;
var resource_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37480__$1,new cljs.core.Keyword(null,"resource-id","resource-id",-1308422582));
var output_name = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37480__$1,new cljs.core.Keyword(null,"output-name","output-name",-1769107767));
var resource_name = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37480__$1,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100));
var js = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37480__$1,new cljs.core.Keyword(null,"js","js",1768080579));
$CLJS.SHADOW_ENV.setLoaded(output_name);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load JS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([resource_name], 0));

shadow.cljs.devtools.client.env.before_load_src(src);

try{shadow.cljs.devtools.client.browser.script_eval([cljs.core.str.cljs$core$IFn$_invoke$arity$1(js),"\n//# sourceURL=",cljs.core.str.cljs$core$IFn$_invoke$arity$1($CLJS.SHADOW_ENV.scriptBase),cljs.core.str.cljs$core$IFn$_invoke$arity$1(output_name)].join(''));
}catch (e37481){var e_38085 = e37481;
if(shadow.cljs.devtools.client.env.log){
console.error(["Failed to load ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name)].join(''),e_38085);
} else {
}

throw (new Error(["Failed to load ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name),": ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(e_38085.message)].join('')));
}

var G__38086 = cljs.core.next(seq__37454__$1);
var G__38087 = null;
var G__38088 = (0);
var G__38089 = (0);
seq__37454 = G__38086;
chunk__37455 = G__38087;
count__37456 = G__38088;
i__37457 = G__38089;
continue;
}
} else {
return null;
}
}
break;
}
});
shadow.cljs.devtools.client.browser.do_js_reload = (function shadow$cljs$devtools$client$browser$do_js_reload(msg,sources,complete_fn,failure_fn){
return shadow.cljs.devtools.client.env.do_js_reload.cljs$core$IFn$_invoke$arity$4(cljs.core.assoc.cljs$core$IFn$_invoke$arity$variadic(msg,new cljs.core.Keyword(null,"log-missing-fn","log-missing-fn",732676765),(function (fn_sym){
return null;
}),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"log-call-async","log-call-async",183826192),(function (fn_sym){
return shadow.cljs.devtools.client.browser.devtools_msg(["call async ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym)].join(''));
}),new cljs.core.Keyword(null,"log-call","log-call",412404391),(function (fn_sym){
return shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym)].join(''));
})], 0)),(function (){
return shadow.cljs.devtools.client.browser.do_js_load(sources);
}),complete_fn,failure_fn);
});
/**
 * when (require '["some-str" :as x]) is done at the REPL we need to manually call the shadow.js.require for it
 * since the file only adds the shadow$provide. only need to do this for shadow-js.
 */
shadow.cljs.devtools.client.browser.do_js_requires = (function shadow$cljs$devtools$client$browser$do_js_requires(js_requires){
var seq__37482 = cljs.core.seq(js_requires);
var chunk__37483 = null;
var count__37484 = (0);
var i__37485 = (0);
while(true){
if((i__37485 < count__37484)){
var js_ns = chunk__37483.cljs$core$IIndexed$_nth$arity$2(null, i__37485);
var require_str_38095 = ["var ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(js_ns)," = shadow.js.require(\"",cljs.core.str.cljs$core$IFn$_invoke$arity$1(js_ns),"\");"].join('');
shadow.cljs.devtools.client.browser.script_eval(require_str_38095);


var G__38096 = seq__37482;
var G__38097 = chunk__37483;
var G__38098 = count__37484;
var G__38099 = (i__37485 + (1));
seq__37482 = G__38096;
chunk__37483 = G__38097;
count__37484 = G__38098;
i__37485 = G__38099;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__37482);
if(temp__5825__auto__){
var seq__37482__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__37482__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__37482__$1);
var G__38104 = cljs.core.chunk_rest(seq__37482__$1);
var G__38105 = c__5525__auto__;
var G__38106 = cljs.core.count(c__5525__auto__);
var G__38107 = (0);
seq__37482 = G__38104;
chunk__37483 = G__38105;
count__37484 = G__38106;
i__37485 = G__38107;
continue;
} else {
var js_ns = cljs.core.first(seq__37482__$1);
var require_str_38108 = ["var ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(js_ns)," = shadow.js.require(\"",cljs.core.str.cljs$core$IFn$_invoke$arity$1(js_ns),"\");"].join('');
shadow.cljs.devtools.client.browser.script_eval(require_str_38108);


var G__38109 = cljs.core.next(seq__37482__$1);
var G__38110 = null;
var G__38111 = (0);
var G__38112 = (0);
seq__37482 = G__38109;
chunk__37483 = G__38110;
count__37484 = G__38111;
i__37485 = G__38112;
continue;
}
} else {
return null;
}
}
break;
}
});
shadow.cljs.devtools.client.browser.handle_build_complete = (function shadow$cljs$devtools$client$browser$handle_build_complete(runtime,p__37487){
var map__37488 = p__37487;
var map__37488__$1 = cljs.core.__destructure_map(map__37488);
var msg = map__37488__$1;
var info = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37488__$1,new cljs.core.Keyword(null,"info","info",-317069002));
var reload_info = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37488__$1,new cljs.core.Keyword(null,"reload-info","reload-info",1648088086));
var warnings = cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentVector.EMPTY,cljs.core.distinct.cljs$core$IFn$_invoke$arity$1((function (){var iter__5480__auto__ = (function shadow$cljs$devtools$client$browser$handle_build_complete_$_iter__37489(s__37490){
return (new cljs.core.LazySeq(null,(function (){
var s__37490__$1 = s__37490;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__37490__$1);
if(temp__5825__auto__){
var xs__6385__auto__ = temp__5825__auto__;
var map__37496 = cljs.core.first(xs__6385__auto__);
var map__37496__$1 = cljs.core.__destructure_map(map__37496);
var src = map__37496__$1;
var resource_name = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37496__$1,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100));
var warnings = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37496__$1,new cljs.core.Keyword(null,"warnings","warnings",-735437651));
if(cljs.core.not(new cljs.core.Keyword(null,"from-jar","from-jar",1050932827).cljs$core$IFn$_invoke$arity$1(src))){
var iterys__5476__auto__ = ((function (s__37490__$1,map__37496,map__37496__$1,src,resource_name,warnings,xs__6385__auto__,temp__5825__auto__,map__37488,map__37488__$1,msg,info,reload_info){
return (function shadow$cljs$devtools$client$browser$handle_build_complete_$_iter__37489_$_iter__37491(s__37492){
return (new cljs.core.LazySeq(null,((function (s__37490__$1,map__37496,map__37496__$1,src,resource_name,warnings,xs__6385__auto__,temp__5825__auto__,map__37488,map__37488__$1,msg,info,reload_info){
return (function (){
var s__37492__$1 = s__37492;
while(true){
var temp__5825__auto____$1 = cljs.core.seq(s__37492__$1);
if(temp__5825__auto____$1){
var s__37492__$2 = temp__5825__auto____$1;
if(cljs.core.chunked_seq_QMARK_(s__37492__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__37492__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__37494 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__37493 = (0);
while(true){
if((i__37493 < size__5479__auto__)){
var warning = cljs.core._nth(c__5478__auto__,i__37493);
cljs.core.chunk_append(b__37494,cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(warning,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100),resource_name));

var G__38114 = (i__37493 + (1));
i__37493 = G__38114;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__37494),shadow$cljs$devtools$client$browser$handle_build_complete_$_iter__37489_$_iter__37491(cljs.core.chunk_rest(s__37492__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__37494),null);
}
} else {
var warning = cljs.core.first(s__37492__$2);
return cljs.core.cons(cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(warning,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100),resource_name),shadow$cljs$devtools$client$browser$handle_build_complete_$_iter__37489_$_iter__37491(cljs.core.rest(s__37492__$2)));
}
} else {
return null;
}
break;
}
});})(s__37490__$1,map__37496,map__37496__$1,src,resource_name,warnings,xs__6385__auto__,temp__5825__auto__,map__37488,map__37488__$1,msg,info,reload_info))
,null,null));
});})(s__37490__$1,map__37496,map__37496__$1,src,resource_name,warnings,xs__6385__auto__,temp__5825__auto__,map__37488,map__37488__$1,msg,info,reload_info))
;
var fs__5477__auto__ = cljs.core.seq(iterys__5476__auto__(warnings));
if(fs__5477__auto__){
return cljs.core.concat.cljs$core$IFn$_invoke$arity$2(fs__5477__auto__,shadow$cljs$devtools$client$browser$handle_build_complete_$_iter__37489(cljs.core.rest(s__37490__$1)));
} else {
var G__38115 = cljs.core.rest(s__37490__$1);
s__37490__$1 = G__38115;
continue;
}
} else {
var G__38116 = cljs.core.rest(s__37490__$1);
s__37490__$1 = G__38116;
continue;
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(new cljs.core.Keyword(null,"sources","sources",-321166424).cljs$core$IFn$_invoke$arity$1(info));
})()));
if(shadow.cljs.devtools.client.env.log){
var seq__37497_38117 = cljs.core.seq(warnings);
var chunk__37498_38118 = null;
var count__37499_38119 = (0);
var i__37500_38120 = (0);
while(true){
if((i__37500_38120 < count__37499_38119)){
var map__37503_38121 = chunk__37498_38118.cljs$core$IIndexed$_nth$arity$2(null, i__37500_38120);
var map__37503_38122__$1 = cljs.core.__destructure_map(map__37503_38121);
var w_38123 = map__37503_38122__$1;
var msg_38124__$1 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37503_38122__$1,new cljs.core.Keyword(null,"msg","msg",-1386103444));
var line_38125 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37503_38122__$1,new cljs.core.Keyword(null,"line","line",212345235));
var column_38126 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37503_38122__$1,new cljs.core.Keyword(null,"column","column",2078222095));
var resource_name_38127 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37503_38122__$1,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100));
console.warn(["BUILD-WARNING in ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name_38127)," at [",cljs.core.str.cljs$core$IFn$_invoke$arity$1(line_38125),":",cljs.core.str.cljs$core$IFn$_invoke$arity$1(column_38126),"]\n\t",cljs.core.str.cljs$core$IFn$_invoke$arity$1(msg_38124__$1)].join(''));


var G__38128 = seq__37497_38117;
var G__38129 = chunk__37498_38118;
var G__38130 = count__37499_38119;
var G__38131 = (i__37500_38120 + (1));
seq__37497_38117 = G__38128;
chunk__37498_38118 = G__38129;
count__37499_38119 = G__38130;
i__37500_38120 = G__38131;
continue;
} else {
var temp__5825__auto___38132 = cljs.core.seq(seq__37497_38117);
if(temp__5825__auto___38132){
var seq__37497_38133__$1 = temp__5825__auto___38132;
if(cljs.core.chunked_seq_QMARK_(seq__37497_38133__$1)){
var c__5525__auto___38134 = cljs.core.chunk_first(seq__37497_38133__$1);
var G__38135 = cljs.core.chunk_rest(seq__37497_38133__$1);
var G__38136 = c__5525__auto___38134;
var G__38137 = cljs.core.count(c__5525__auto___38134);
var G__38138 = (0);
seq__37497_38117 = G__38135;
chunk__37498_38118 = G__38136;
count__37499_38119 = G__38137;
i__37500_38120 = G__38138;
continue;
} else {
var map__37504_38139 = cljs.core.first(seq__37497_38133__$1);
var map__37504_38140__$1 = cljs.core.__destructure_map(map__37504_38139);
var w_38141 = map__37504_38140__$1;
var msg_38142__$1 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37504_38140__$1,new cljs.core.Keyword(null,"msg","msg",-1386103444));
var line_38143 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37504_38140__$1,new cljs.core.Keyword(null,"line","line",212345235));
var column_38144 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37504_38140__$1,new cljs.core.Keyword(null,"column","column",2078222095));
var resource_name_38145 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37504_38140__$1,new cljs.core.Keyword(null,"resource-name","resource-name",2001617100));
console.warn(["BUILD-WARNING in ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(resource_name_38145)," at [",cljs.core.str.cljs$core$IFn$_invoke$arity$1(line_38143),":",cljs.core.str.cljs$core$IFn$_invoke$arity$1(column_38144),"]\n\t",cljs.core.str.cljs$core$IFn$_invoke$arity$1(msg_38142__$1)].join(''));


var G__38146 = cljs.core.next(seq__37497_38133__$1);
var G__38147 = null;
var G__38148 = (0);
var G__38149 = (0);
seq__37497_38117 = G__38146;
chunk__37498_38118 = G__38147;
count__37499_38119 = G__38148;
i__37500_38120 = G__38149;
continue;
}
} else {
}
}
break;
}
} else {
}

if((!(shadow.cljs.devtools.client.env.autoload))){
return shadow.cljs.devtools.client.hud.load_end_success();
} else {
if(((cljs.core.empty_QMARK_(warnings)) || (shadow.cljs.devtools.client.env.ignore_warnings))){
var sources_to_get = shadow.cljs.devtools.client.env.filter_reload_sources(info,reload_info);
if(cljs.core.not(cljs.core.seq(sources_to_get))){
return shadow.cljs.devtools.client.hud.load_end_success();
} else {
if(cljs.core.seq(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(msg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"reload-info","reload-info",1648088086),new cljs.core.Keyword(null,"after-load","after-load",-1278503285)], null)))){
} else {
shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("reloading code but no :after-load hooks are configured!",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2(["https://shadow-cljs.github.io/docs/UsersGuide.html#_lifecycle_hooks"], 0));
}

return shadow.cljs.devtools.client.shared.load_sources(runtime,sources_to_get,(function (p1__37486_SHARP_){
return shadow.cljs.devtools.client.browser.do_js_reload(msg,p1__37486_SHARP_,shadow.cljs.devtools.client.hud.load_end_success,shadow.cljs.devtools.client.hud.load_failure);
}));
}
} else {
return null;
}
}
});
shadow.cljs.devtools.client.browser.page_load_uri = (cljs.core.truth_(goog.global.document)?goog.Uri.parse(document.location.href):null);
shadow.cljs.devtools.client.browser.match_paths = (function shadow$cljs$devtools$client$browser$match_paths(old,new$){
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2("file",shadow.cljs.devtools.client.browser.page_load_uri.getScheme())){
var rel_new = cljs.core.subs.cljs$core$IFn$_invoke$arity$2(new$,(1));
if(((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(old,rel_new)) || (clojure.string.starts_with_QMARK_(old,[rel_new,"?"].join(''))))){
return rel_new;
} else {
return null;
}
} else {
var node_uri = goog.Uri.parse(old);
var node_uri_resolved = shadow.cljs.devtools.client.browser.page_load_uri.resolve(node_uri);
var node_abs = node_uri_resolved.getPath();
var and__5000__auto__ = ((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$1(shadow.cljs.devtools.client.browser.page_load_uri.hasSameDomainAs(node_uri))) || (cljs.core.not(node_uri.hasDomain())));
if(and__5000__auto__){
var and__5000__auto____$1 = cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(node_abs,new$);
if(and__5000__auto____$1){
return new$;
} else {
return and__5000__auto____$1;
}
} else {
return and__5000__auto__;
}
}
});
shadow.cljs.devtools.client.browser.handle_asset_update = (function shadow$cljs$devtools$client$browser$handle_asset_update(p__37513){
var map__37514 = p__37513;
var map__37514__$1 = cljs.core.__destructure_map(map__37514);
var msg = map__37514__$1;
var updates = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37514__$1,new cljs.core.Keyword(null,"updates","updates",2013983452));
var reload_info = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37514__$1,new cljs.core.Keyword(null,"reload-info","reload-info",1648088086));
var seq__37515 = cljs.core.seq(updates);
var chunk__37517 = null;
var count__37518 = (0);
var i__37519 = (0);
while(true){
if((i__37519 < count__37518)){
var path = chunk__37517.cljs$core$IIndexed$_nth$arity$2(null, i__37519);
if(clojure.string.ends_with_QMARK_(path,"css")){
var seq__37768_38150 = cljs.core.seq(cljs.core.array_seq.cljs$core$IFn$_invoke$arity$1(document.querySelectorAll("link[rel=\"stylesheet\"]")));
var chunk__37772_38151 = null;
var count__37773_38152 = (0);
var i__37774_38153 = (0);
while(true){
if((i__37774_38153 < count__37773_38152)){
var node_38154 = chunk__37772_38151.cljs$core$IIndexed$_nth$arity$2(null, i__37774_38153);
if(cljs.core.not(node_38154.shadow$old)){
var path_match_38155 = shadow.cljs.devtools.client.browser.match_paths(node_38154.getAttribute("href"),path);
if(cljs.core.truth_(path_match_38155)){
var new_link_38156 = (function (){var G__37855 = node_38154.cloneNode(true);
G__37855.setAttribute("href",[cljs.core.str.cljs$core$IFn$_invoke$arity$1(path_match_38155),"?r=",cljs.core.str.cljs$core$IFn$_invoke$arity$1(cljs.core.rand.cljs$core$IFn$_invoke$arity$0())].join(''));

return G__37855;
})();
(node_38154.shadow$old = true);

(new_link_38156.onload = ((function (seq__37768_38150,chunk__37772_38151,count__37773_38152,i__37774_38153,seq__37515,chunk__37517,count__37518,i__37519,new_link_38156,path_match_38155,node_38154,path,map__37514,map__37514__$1,msg,updates,reload_info){
return (function (e){
var seq__37858_38157 = cljs.core.seq(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(msg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"reload-info","reload-info",1648088086),new cljs.core.Keyword(null,"asset-load","asset-load",-1925902322)], null)));
var chunk__37860_38158 = null;
var count__37861_38159 = (0);
var i__37862_38160 = (0);
while(true){
if((i__37862_38160 < count__37861_38159)){
var map__37893_38161 = chunk__37860_38158.cljs$core$IIndexed$_nth$arity$2(null, i__37862_38160);
var map__37893_38162__$1 = cljs.core.__destructure_map(map__37893_38161);
var task_38163 = map__37893_38162__$1;
var fn_str_38164 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37893_38162__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38165 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37893_38162__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38205 = goog.getObjectByName(fn_str_38164,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38165)].join(''));

(fn_obj_38205.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38205.cljs$core$IFn$_invoke$arity$2(path,new_link_38156) : fn_obj_38205.call(null, path,new_link_38156));


var G__38206 = seq__37858_38157;
var G__38207 = chunk__37860_38158;
var G__38208 = count__37861_38159;
var G__38209 = (i__37862_38160 + (1));
seq__37858_38157 = G__38206;
chunk__37860_38158 = G__38207;
count__37861_38159 = G__38208;
i__37862_38160 = G__38209;
continue;
} else {
var temp__5825__auto___38214 = cljs.core.seq(seq__37858_38157);
if(temp__5825__auto___38214){
var seq__37858_38215__$1 = temp__5825__auto___38214;
if(cljs.core.chunked_seq_QMARK_(seq__37858_38215__$1)){
var c__5525__auto___38217 = cljs.core.chunk_first(seq__37858_38215__$1);
var G__38218 = cljs.core.chunk_rest(seq__37858_38215__$1);
var G__38219 = c__5525__auto___38217;
var G__38220 = cljs.core.count(c__5525__auto___38217);
var G__38221 = (0);
seq__37858_38157 = G__38218;
chunk__37860_38158 = G__38219;
count__37861_38159 = G__38220;
i__37862_38160 = G__38221;
continue;
} else {
var map__37895_38227 = cljs.core.first(seq__37858_38215__$1);
var map__37895_38228__$1 = cljs.core.__destructure_map(map__37895_38227);
var task_38229 = map__37895_38228__$1;
var fn_str_38230 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37895_38228__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38231 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37895_38228__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38234 = goog.getObjectByName(fn_str_38230,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38231)].join(''));

(fn_obj_38234.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38234.cljs$core$IFn$_invoke$arity$2(path,new_link_38156) : fn_obj_38234.call(null, path,new_link_38156));


var G__38238 = cljs.core.next(seq__37858_38215__$1);
var G__38239 = null;
var G__38240 = (0);
var G__38241 = (0);
seq__37858_38157 = G__38238;
chunk__37860_38158 = G__38239;
count__37861_38159 = G__38240;
i__37862_38160 = G__38241;
continue;
}
} else {
}
}
break;
}

return goog.dom.removeNode(node_38154);
});})(seq__37768_38150,chunk__37772_38151,count__37773_38152,i__37774_38153,seq__37515,chunk__37517,count__37518,i__37519,new_link_38156,path_match_38155,node_38154,path,map__37514,map__37514__$1,msg,updates,reload_info))
);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load CSS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([path_match_38155], 0));

goog.dom.insertSiblingAfter(new_link_38156,node_38154);


var G__38249 = seq__37768_38150;
var G__38250 = chunk__37772_38151;
var G__38251 = count__37773_38152;
var G__38252 = (i__37774_38153 + (1));
seq__37768_38150 = G__38249;
chunk__37772_38151 = G__38250;
count__37773_38152 = G__38251;
i__37774_38153 = G__38252;
continue;
} else {
var G__38253 = seq__37768_38150;
var G__38254 = chunk__37772_38151;
var G__38255 = count__37773_38152;
var G__38256 = (i__37774_38153 + (1));
seq__37768_38150 = G__38253;
chunk__37772_38151 = G__38254;
count__37773_38152 = G__38255;
i__37774_38153 = G__38256;
continue;
}
} else {
var G__38257 = seq__37768_38150;
var G__38258 = chunk__37772_38151;
var G__38259 = count__37773_38152;
var G__38260 = (i__37774_38153 + (1));
seq__37768_38150 = G__38257;
chunk__37772_38151 = G__38258;
count__37773_38152 = G__38259;
i__37774_38153 = G__38260;
continue;
}
} else {
var temp__5825__auto___38261 = cljs.core.seq(seq__37768_38150);
if(temp__5825__auto___38261){
var seq__37768_38262__$1 = temp__5825__auto___38261;
if(cljs.core.chunked_seq_QMARK_(seq__37768_38262__$1)){
var c__5525__auto___38263 = cljs.core.chunk_first(seq__37768_38262__$1);
var G__38264 = cljs.core.chunk_rest(seq__37768_38262__$1);
var G__38265 = c__5525__auto___38263;
var G__38266 = cljs.core.count(c__5525__auto___38263);
var G__38267 = (0);
seq__37768_38150 = G__38264;
chunk__37772_38151 = G__38265;
count__37773_38152 = G__38266;
i__37774_38153 = G__38267;
continue;
} else {
var node_38268 = cljs.core.first(seq__37768_38262__$1);
if(cljs.core.not(node_38268.shadow$old)){
var path_match_38269 = shadow.cljs.devtools.client.browser.match_paths(node_38268.getAttribute("href"),path);
if(cljs.core.truth_(path_match_38269)){
var new_link_38270 = (function (){var G__37897 = node_38268.cloneNode(true);
G__37897.setAttribute("href",[cljs.core.str.cljs$core$IFn$_invoke$arity$1(path_match_38269),"?r=",cljs.core.str.cljs$core$IFn$_invoke$arity$1(cljs.core.rand.cljs$core$IFn$_invoke$arity$0())].join(''));

return G__37897;
})();
(node_38268.shadow$old = true);

(new_link_38270.onload = ((function (seq__37768_38150,chunk__37772_38151,count__37773_38152,i__37774_38153,seq__37515,chunk__37517,count__37518,i__37519,new_link_38270,path_match_38269,node_38268,seq__37768_38262__$1,temp__5825__auto___38261,path,map__37514,map__37514__$1,msg,updates,reload_info){
return (function (e){
var seq__37898_38271 = cljs.core.seq(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(msg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"reload-info","reload-info",1648088086),new cljs.core.Keyword(null,"asset-load","asset-load",-1925902322)], null)));
var chunk__37900_38272 = null;
var count__37901_38273 = (0);
var i__37902_38274 = (0);
while(true){
if((i__37902_38274 < count__37901_38273)){
var map__37910_38275 = chunk__37900_38272.cljs$core$IIndexed$_nth$arity$2(null, i__37902_38274);
var map__37910_38276__$1 = cljs.core.__destructure_map(map__37910_38275);
var task_38277 = map__37910_38276__$1;
var fn_str_38278 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37910_38276__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38279 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37910_38276__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38280 = goog.getObjectByName(fn_str_38278,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38279)].join(''));

(fn_obj_38280.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38280.cljs$core$IFn$_invoke$arity$2(path,new_link_38270) : fn_obj_38280.call(null, path,new_link_38270));


var G__38281 = seq__37898_38271;
var G__38282 = chunk__37900_38272;
var G__38283 = count__37901_38273;
var G__38284 = (i__37902_38274 + (1));
seq__37898_38271 = G__38281;
chunk__37900_38272 = G__38282;
count__37901_38273 = G__38283;
i__37902_38274 = G__38284;
continue;
} else {
var temp__5825__auto___38285__$1 = cljs.core.seq(seq__37898_38271);
if(temp__5825__auto___38285__$1){
var seq__37898_38286__$1 = temp__5825__auto___38285__$1;
if(cljs.core.chunked_seq_QMARK_(seq__37898_38286__$1)){
var c__5525__auto___38292 = cljs.core.chunk_first(seq__37898_38286__$1);
var G__38293 = cljs.core.chunk_rest(seq__37898_38286__$1);
var G__38294 = c__5525__auto___38292;
var G__38295 = cljs.core.count(c__5525__auto___38292);
var G__38296 = (0);
seq__37898_38271 = G__38293;
chunk__37900_38272 = G__38294;
count__37901_38273 = G__38295;
i__37902_38274 = G__38296;
continue;
} else {
var map__37912_38297 = cljs.core.first(seq__37898_38286__$1);
var map__37912_38298__$1 = cljs.core.__destructure_map(map__37912_38297);
var task_38299 = map__37912_38298__$1;
var fn_str_38300 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37912_38298__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38301 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__37912_38298__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38302 = goog.getObjectByName(fn_str_38300,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38301)].join(''));

(fn_obj_38302.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38302.cljs$core$IFn$_invoke$arity$2(path,new_link_38270) : fn_obj_38302.call(null, path,new_link_38270));


var G__38304 = cljs.core.next(seq__37898_38286__$1);
var G__38305 = null;
var G__38306 = (0);
var G__38307 = (0);
seq__37898_38271 = G__38304;
chunk__37900_38272 = G__38305;
count__37901_38273 = G__38306;
i__37902_38274 = G__38307;
continue;
}
} else {
}
}
break;
}

return goog.dom.removeNode(node_38268);
});})(seq__37768_38150,chunk__37772_38151,count__37773_38152,i__37774_38153,seq__37515,chunk__37517,count__37518,i__37519,new_link_38270,path_match_38269,node_38268,seq__37768_38262__$1,temp__5825__auto___38261,path,map__37514,map__37514__$1,msg,updates,reload_info))
);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load CSS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([path_match_38269], 0));

goog.dom.insertSiblingAfter(new_link_38270,node_38268);


var G__38311 = cljs.core.next(seq__37768_38262__$1);
var G__38312 = null;
var G__38313 = (0);
var G__38314 = (0);
seq__37768_38150 = G__38311;
chunk__37772_38151 = G__38312;
count__37773_38152 = G__38313;
i__37774_38153 = G__38314;
continue;
} else {
var G__38322 = cljs.core.next(seq__37768_38262__$1);
var G__38323 = null;
var G__38324 = (0);
var G__38325 = (0);
seq__37768_38150 = G__38322;
chunk__37772_38151 = G__38323;
count__37773_38152 = G__38324;
i__37774_38153 = G__38325;
continue;
}
} else {
var G__38329 = cljs.core.next(seq__37768_38262__$1);
var G__38330 = null;
var G__38331 = (0);
var G__38332 = (0);
seq__37768_38150 = G__38329;
chunk__37772_38151 = G__38330;
count__37773_38152 = G__38331;
i__37774_38153 = G__38332;
continue;
}
}
} else {
}
}
break;
}


var G__38333 = seq__37515;
var G__38334 = chunk__37517;
var G__38335 = count__37518;
var G__38336 = (i__37519 + (1));
seq__37515 = G__38333;
chunk__37517 = G__38334;
count__37518 = G__38335;
i__37519 = G__38336;
continue;
} else {
var G__38337 = seq__37515;
var G__38338 = chunk__37517;
var G__38339 = count__37518;
var G__38340 = (i__37519 + (1));
seq__37515 = G__38337;
chunk__37517 = G__38338;
count__37518 = G__38339;
i__37519 = G__38340;
continue;
}
} else {
var temp__5825__auto__ = cljs.core.seq(seq__37515);
if(temp__5825__auto__){
var seq__37515__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__37515__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__37515__$1);
var G__38341 = cljs.core.chunk_rest(seq__37515__$1);
var G__38342 = c__5525__auto__;
var G__38343 = cljs.core.count(c__5525__auto__);
var G__38344 = (0);
seq__37515 = G__38341;
chunk__37517 = G__38342;
count__37518 = G__38343;
i__37519 = G__38344;
continue;
} else {
var path = cljs.core.first(seq__37515__$1);
if(clojure.string.ends_with_QMARK_(path,"css")){
var seq__37927_38345 = cljs.core.seq(cljs.core.array_seq.cljs$core$IFn$_invoke$arity$1(document.querySelectorAll("link[rel=\"stylesheet\"]")));
var chunk__37931_38346 = null;
var count__37932_38347 = (0);
var i__37933_38348 = (0);
while(true){
if((i__37933_38348 < count__37932_38347)){
var node_38349 = chunk__37931_38346.cljs$core$IIndexed$_nth$arity$2(null, i__37933_38348);
if(cljs.core.not(node_38349.shadow$old)){
var path_match_38350 = shadow.cljs.devtools.client.browser.match_paths(node_38349.getAttribute("href"),path);
if(cljs.core.truth_(path_match_38350)){
var new_link_38352 = (function (){var G__37991 = node_38349.cloneNode(true);
G__37991.setAttribute("href",[cljs.core.str.cljs$core$IFn$_invoke$arity$1(path_match_38350),"?r=",cljs.core.str.cljs$core$IFn$_invoke$arity$1(cljs.core.rand.cljs$core$IFn$_invoke$arity$0())].join(''));

return G__37991;
})();
(node_38349.shadow$old = true);

(new_link_38352.onload = ((function (seq__37927_38345,chunk__37931_38346,count__37932_38347,i__37933_38348,seq__37515,chunk__37517,count__37518,i__37519,new_link_38352,path_match_38350,node_38349,path,seq__37515__$1,temp__5825__auto__,map__37514,map__37514__$1,msg,updates,reload_info){
return (function (e){
var seq__37992_38354 = cljs.core.seq(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(msg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"reload-info","reload-info",1648088086),new cljs.core.Keyword(null,"asset-load","asset-load",-1925902322)], null)));
var chunk__37994_38355 = null;
var count__37995_38356 = (0);
var i__37996_38357 = (0);
while(true){
if((i__37996_38357 < count__37995_38356)){
var map__38000_38358 = chunk__37994_38355.cljs$core$IIndexed$_nth$arity$2(null, i__37996_38357);
var map__38000_38359__$1 = cljs.core.__destructure_map(map__38000_38358);
var task_38360 = map__38000_38359__$1;
var fn_str_38361 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38000_38359__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38362 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38000_38359__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38363 = goog.getObjectByName(fn_str_38361,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38362)].join(''));

(fn_obj_38363.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38363.cljs$core$IFn$_invoke$arity$2(path,new_link_38352) : fn_obj_38363.call(null, path,new_link_38352));


var G__38364 = seq__37992_38354;
var G__38365 = chunk__37994_38355;
var G__38366 = count__37995_38356;
var G__38367 = (i__37996_38357 + (1));
seq__37992_38354 = G__38364;
chunk__37994_38355 = G__38365;
count__37995_38356 = G__38366;
i__37996_38357 = G__38367;
continue;
} else {
var temp__5825__auto___38368__$1 = cljs.core.seq(seq__37992_38354);
if(temp__5825__auto___38368__$1){
var seq__37992_38369__$1 = temp__5825__auto___38368__$1;
if(cljs.core.chunked_seq_QMARK_(seq__37992_38369__$1)){
var c__5525__auto___38370 = cljs.core.chunk_first(seq__37992_38369__$1);
var G__38372 = cljs.core.chunk_rest(seq__37992_38369__$1);
var G__38373 = c__5525__auto___38370;
var G__38374 = cljs.core.count(c__5525__auto___38370);
var G__38375 = (0);
seq__37992_38354 = G__38372;
chunk__37994_38355 = G__38373;
count__37995_38356 = G__38374;
i__37996_38357 = G__38375;
continue;
} else {
var map__38001_38377 = cljs.core.first(seq__37992_38369__$1);
var map__38001_38378__$1 = cljs.core.__destructure_map(map__38001_38377);
var task_38379 = map__38001_38378__$1;
var fn_str_38380 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38001_38378__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38381 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38001_38378__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38382 = goog.getObjectByName(fn_str_38380,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38381)].join(''));

(fn_obj_38382.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38382.cljs$core$IFn$_invoke$arity$2(path,new_link_38352) : fn_obj_38382.call(null, path,new_link_38352));


var G__38383 = cljs.core.next(seq__37992_38369__$1);
var G__38384 = null;
var G__38385 = (0);
var G__38386 = (0);
seq__37992_38354 = G__38383;
chunk__37994_38355 = G__38384;
count__37995_38356 = G__38385;
i__37996_38357 = G__38386;
continue;
}
} else {
}
}
break;
}

return goog.dom.removeNode(node_38349);
});})(seq__37927_38345,chunk__37931_38346,count__37932_38347,i__37933_38348,seq__37515,chunk__37517,count__37518,i__37519,new_link_38352,path_match_38350,node_38349,path,seq__37515__$1,temp__5825__auto__,map__37514,map__37514__$1,msg,updates,reload_info))
);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load CSS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([path_match_38350], 0));

goog.dom.insertSiblingAfter(new_link_38352,node_38349);


var G__38387 = seq__37927_38345;
var G__38388 = chunk__37931_38346;
var G__38389 = count__37932_38347;
var G__38390 = (i__37933_38348 + (1));
seq__37927_38345 = G__38387;
chunk__37931_38346 = G__38388;
count__37932_38347 = G__38389;
i__37933_38348 = G__38390;
continue;
} else {
var G__38391 = seq__37927_38345;
var G__38392 = chunk__37931_38346;
var G__38393 = count__37932_38347;
var G__38394 = (i__37933_38348 + (1));
seq__37927_38345 = G__38391;
chunk__37931_38346 = G__38392;
count__37932_38347 = G__38393;
i__37933_38348 = G__38394;
continue;
}
} else {
var G__38395 = seq__37927_38345;
var G__38396 = chunk__37931_38346;
var G__38397 = count__37932_38347;
var G__38398 = (i__37933_38348 + (1));
seq__37927_38345 = G__38395;
chunk__37931_38346 = G__38396;
count__37932_38347 = G__38397;
i__37933_38348 = G__38398;
continue;
}
} else {
var temp__5825__auto___38399__$1 = cljs.core.seq(seq__37927_38345);
if(temp__5825__auto___38399__$1){
var seq__37927_38401__$1 = temp__5825__auto___38399__$1;
if(cljs.core.chunked_seq_QMARK_(seq__37927_38401__$1)){
var c__5525__auto___38402 = cljs.core.chunk_first(seq__37927_38401__$1);
var G__38403 = cljs.core.chunk_rest(seq__37927_38401__$1);
var G__38404 = c__5525__auto___38402;
var G__38405 = cljs.core.count(c__5525__auto___38402);
var G__38406 = (0);
seq__37927_38345 = G__38403;
chunk__37931_38346 = G__38404;
count__37932_38347 = G__38405;
i__37933_38348 = G__38406;
continue;
} else {
var node_38407 = cljs.core.first(seq__37927_38401__$1);
if(cljs.core.not(node_38407.shadow$old)){
var path_match_38408 = shadow.cljs.devtools.client.browser.match_paths(node_38407.getAttribute("href"),path);
if(cljs.core.truth_(path_match_38408)){
var new_link_38409 = (function (){var G__38002 = node_38407.cloneNode(true);
G__38002.setAttribute("href",[cljs.core.str.cljs$core$IFn$_invoke$arity$1(path_match_38408),"?r=",cljs.core.str.cljs$core$IFn$_invoke$arity$1(cljs.core.rand.cljs$core$IFn$_invoke$arity$0())].join(''));

return G__38002;
})();
(node_38407.shadow$old = true);

(new_link_38409.onload = ((function (seq__37927_38345,chunk__37931_38346,count__37932_38347,i__37933_38348,seq__37515,chunk__37517,count__37518,i__37519,new_link_38409,path_match_38408,node_38407,seq__37927_38401__$1,temp__5825__auto___38399__$1,path,seq__37515__$1,temp__5825__auto__,map__37514,map__37514__$1,msg,updates,reload_info){
return (function (e){
var seq__38007_38411 = cljs.core.seq(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(msg,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"reload-info","reload-info",1648088086),new cljs.core.Keyword(null,"asset-load","asset-load",-1925902322)], null)));
var chunk__38009_38412 = null;
var count__38010_38413 = (0);
var i__38011_38414 = (0);
while(true){
if((i__38011_38414 < count__38010_38413)){
var map__38015_38415 = chunk__38009_38412.cljs$core$IIndexed$_nth$arity$2(null, i__38011_38414);
var map__38015_38416__$1 = cljs.core.__destructure_map(map__38015_38415);
var task_38417 = map__38015_38416__$1;
var fn_str_38418 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38015_38416__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38419 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38015_38416__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38420 = goog.getObjectByName(fn_str_38418,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38419)].join(''));

(fn_obj_38420.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38420.cljs$core$IFn$_invoke$arity$2(path,new_link_38409) : fn_obj_38420.call(null, path,new_link_38409));


var G__38421 = seq__38007_38411;
var G__38422 = chunk__38009_38412;
var G__38423 = count__38010_38413;
var G__38424 = (i__38011_38414 + (1));
seq__38007_38411 = G__38421;
chunk__38009_38412 = G__38422;
count__38010_38413 = G__38423;
i__38011_38414 = G__38424;
continue;
} else {
var temp__5825__auto___38425__$2 = cljs.core.seq(seq__38007_38411);
if(temp__5825__auto___38425__$2){
var seq__38007_38426__$1 = temp__5825__auto___38425__$2;
if(cljs.core.chunked_seq_QMARK_(seq__38007_38426__$1)){
var c__5525__auto___38427 = cljs.core.chunk_first(seq__38007_38426__$1);
var G__38428 = cljs.core.chunk_rest(seq__38007_38426__$1);
var G__38429 = c__5525__auto___38427;
var G__38430 = cljs.core.count(c__5525__auto___38427);
var G__38431 = (0);
seq__38007_38411 = G__38428;
chunk__38009_38412 = G__38429;
count__38010_38413 = G__38430;
i__38011_38414 = G__38431;
continue;
} else {
var map__38016_38432 = cljs.core.first(seq__38007_38426__$1);
var map__38016_38433__$1 = cljs.core.__destructure_map(map__38016_38432);
var task_38434 = map__38016_38433__$1;
var fn_str_38435 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38016_38433__$1,new cljs.core.Keyword(null,"fn-str","fn-str",-1348506402));
var fn_sym_38436 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38016_38433__$1,new cljs.core.Keyword(null,"fn-sym","fn-sym",1423988510));
var fn_obj_38437 = goog.getObjectByName(fn_str_38435,$CLJS);
shadow.cljs.devtools.client.browser.devtools_msg(["call ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(fn_sym_38436)].join(''));

(fn_obj_38437.cljs$core$IFn$_invoke$arity$2 ? fn_obj_38437.cljs$core$IFn$_invoke$arity$2(path,new_link_38409) : fn_obj_38437.call(null, path,new_link_38409));


var G__38438 = cljs.core.next(seq__38007_38426__$1);
var G__38439 = null;
var G__38440 = (0);
var G__38441 = (0);
seq__38007_38411 = G__38438;
chunk__38009_38412 = G__38439;
count__38010_38413 = G__38440;
i__38011_38414 = G__38441;
continue;
}
} else {
}
}
break;
}

return goog.dom.removeNode(node_38407);
});})(seq__37927_38345,chunk__37931_38346,count__37932_38347,i__37933_38348,seq__37515,chunk__37517,count__37518,i__37519,new_link_38409,path_match_38408,node_38407,seq__37927_38401__$1,temp__5825__auto___38399__$1,path,seq__37515__$1,temp__5825__auto__,map__37514,map__37514__$1,msg,updates,reload_info))
);

shadow.cljs.devtools.client.browser.devtools_msg.cljs$core$IFn$_invoke$arity$variadic("load CSS",cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([path_match_38408], 0));

goog.dom.insertSiblingAfter(new_link_38409,node_38407);


var G__38442 = cljs.core.next(seq__37927_38401__$1);
var G__38443 = null;
var G__38444 = (0);
var G__38445 = (0);
seq__37927_38345 = G__38442;
chunk__37931_38346 = G__38443;
count__37932_38347 = G__38444;
i__37933_38348 = G__38445;
continue;
} else {
var G__38446 = cljs.core.next(seq__37927_38401__$1);
var G__38447 = null;
var G__38448 = (0);
var G__38449 = (0);
seq__37927_38345 = G__38446;
chunk__37931_38346 = G__38447;
count__37932_38347 = G__38448;
i__37933_38348 = G__38449;
continue;
}
} else {
var G__38450 = cljs.core.next(seq__37927_38401__$1);
var G__38451 = null;
var G__38452 = (0);
var G__38453 = (0);
seq__37927_38345 = G__38450;
chunk__37931_38346 = G__38451;
count__37932_38347 = G__38452;
i__37933_38348 = G__38453;
continue;
}
}
} else {
}
}
break;
}


var G__38454 = cljs.core.next(seq__37515__$1);
var G__38455 = null;
var G__38456 = (0);
var G__38457 = (0);
seq__37515 = G__38454;
chunk__37517 = G__38455;
count__37518 = G__38456;
i__37519 = G__38457;
continue;
} else {
var G__38458 = cljs.core.next(seq__37515__$1);
var G__38459 = null;
var G__38460 = (0);
var G__38461 = (0);
seq__37515 = G__38458;
chunk__37517 = G__38459;
count__37518 = G__38460;
i__37519 = G__38461;
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
shadow.cljs.devtools.client.browser.global_eval = (function shadow$cljs$devtools$client$browser$global_eval(js){
if(cljs.core.not_EQ_.cljs$core$IFn$_invoke$arity$2("undefined",typeof(module))){
return eval(js);
} else {
return (0,eval)(js);;
}
});
shadow.cljs.devtools.client.browser.runtime_info = (((typeof SHADOW_CONFIG !== 'undefined'))?shadow.json.to_clj.cljs$core$IFn$_invoke$arity$1(SHADOW_CONFIG):null);
shadow.cljs.devtools.client.browser.client_info = cljs.core.merge.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([shadow.cljs.devtools.client.browser.runtime_info,new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"host","host",-1558485167),(cljs.core.truth_(goog.global.document)?new cljs.core.Keyword(null,"browser","browser",828191719):new cljs.core.Keyword(null,"browser-worker","browser-worker",1638998282)),new cljs.core.Keyword(null,"user-agent","user-agent",1220426212),[(cljs.core.truth_(goog.userAgent.OPERA)?"Opera":(cljs.core.truth_(goog.userAgent.product.CHROME)?"Chrome":(cljs.core.truth_(goog.userAgent.IE)?"MSIE":(cljs.core.truth_(goog.userAgent.EDGE)?"Edge":(cljs.core.truth_(goog.userAgent.GECKO)?"Firefox":(cljs.core.truth_(goog.userAgent.SAFARI)?"Safari":(cljs.core.truth_(goog.userAgent.WEBKIT)?"Webkit":null)))))))," ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(goog.userAgent.VERSION)," [",cljs.core.str.cljs$core$IFn$_invoke$arity$1(goog.userAgent.PLATFORM),"]"].join(''),new cljs.core.Keyword(null,"dom","dom",-1236537922),(!((goog.global.document == null)))], null)], 0));
if((typeof shadow !== 'undefined') && (typeof shadow.cljs !== 'undefined') && (typeof shadow.cljs.devtools !== 'undefined') && (typeof shadow.cljs.devtools.client !== 'undefined') && (typeof shadow.cljs.devtools.client.browser !== 'undefined') && (typeof shadow.cljs.devtools.client.browser.ws_was_welcome_ref !== 'undefined')){
} else {
shadow.cljs.devtools.client.browser.ws_was_welcome_ref = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(false);
}
if(((shadow.cljs.devtools.client.env.enabled) && ((shadow.cljs.devtools.client.env.worker_client_id > (0))))){
(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$remote$runtime$api$IEvalJS$ = cljs.core.PROTOCOL_SENTINEL);

(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$remote$runtime$api$IEvalJS$_js_eval$arity$2 = (function (this$,code){
var this$__$1 = this;
return shadow.cljs.devtools.client.browser.global_eval(code);
}));

(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$cljs$devtools$client$shared$IHostSpecific$ = cljs.core.PROTOCOL_SENTINEL);

(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$cljs$devtools$client$shared$IHostSpecific$do_invoke$arity$3 = (function (this$,ns,p__38024){
var map__38025 = p__38024;
var map__38025__$1 = cljs.core.__destructure_map(map__38025);
var js = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38025__$1,new cljs.core.Keyword(null,"js","js",1768080579));
var this$__$1 = this;
return shadow.cljs.devtools.client.browser.global_eval(js);
}));

(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$cljs$devtools$client$shared$IHostSpecific$do_repl_init$arity$4 = (function (runtime,p__38027,done,error){
var map__38028 = p__38027;
var map__38028__$1 = cljs.core.__destructure_map(map__38028);
var repl_sources = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38028__$1,new cljs.core.Keyword(null,"repl-sources","repl-sources",723867535));
var runtime__$1 = this;
return shadow.cljs.devtools.client.shared.load_sources(runtime__$1,cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentVector.EMPTY,cljs.core.remove.cljs$core$IFn$_invoke$arity$2(shadow.cljs.devtools.client.env.src_is_loaded_QMARK_,repl_sources)),(function (sources){
shadow.cljs.devtools.client.browser.do_js_load(sources);

return (done.cljs$core$IFn$_invoke$arity$0 ? done.cljs$core$IFn$_invoke$arity$0() : done.call(null, ));
}));
}));

(shadow.cljs.devtools.client.shared.Runtime.prototype.shadow$cljs$devtools$client$shared$IHostSpecific$do_repl_require$arity$4 = (function (runtime,p__38030,done,error){
var map__38031 = p__38030;
var map__38031__$1 = cljs.core.__destructure_map(map__38031);
var msg = map__38031__$1;
var sources = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38031__$1,new cljs.core.Keyword(null,"sources","sources",-321166424));
var reload_namespaces = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38031__$1,new cljs.core.Keyword(null,"reload-namespaces","reload-namespaces",250210134));
var js_requires = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38031__$1,new cljs.core.Keyword(null,"js-requires","js-requires",-1311472051));
var runtime__$1 = this;
var sources_to_load = cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentVector.EMPTY,cljs.core.remove.cljs$core$IFn$_invoke$arity$2((function (p__38039){
var map__38040 = p__38039;
var map__38040__$1 = cljs.core.__destructure_map(map__38040);
var src = map__38040__$1;
var provides = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38040__$1,new cljs.core.Keyword(null,"provides","provides",-1634397992));
var and__5000__auto__ = shadow.cljs.devtools.client.env.src_is_loaded_QMARK_(src);
if(cljs.core.truth_(and__5000__auto__)){
return cljs.core.not(cljs.core.some(reload_namespaces,provides));
} else {
return and__5000__auto__;
}
}),sources));
if(cljs.core.not(cljs.core.seq(sources_to_load))){
var G__38041 = cljs.core.PersistentVector.EMPTY;
return (done.cljs$core$IFn$_invoke$arity$1 ? done.cljs$core$IFn$_invoke$arity$1(G__38041) : done.call(null, G__38041));
} else {
return shadow.remote.runtime.shared.call.cljs$core$IFn$_invoke$arity$3(runtime__$1,new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"op","op",-1882987955),new cljs.core.Keyword(null,"cljs-load-sources","cljs-load-sources",-1458295962),new cljs.core.Keyword(null,"to","to",192099007),shadow.cljs.devtools.client.env.worker_client_id,new cljs.core.Keyword(null,"sources","sources",-321166424),cljs.core.into.cljs$core$IFn$_invoke$arity$3(cljs.core.PersistentVector.EMPTY,cljs.core.map.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"resource-id","resource-id",-1308422582)),sources_to_load)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"cljs-sources","cljs-sources",31121610),(function (p__38042){
var map__38043 = p__38042;
var map__38043__$1 = cljs.core.__destructure_map(map__38043);
var msg__$1 = map__38043__$1;
var sources__$1 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38043__$1,new cljs.core.Keyword(null,"sources","sources",-321166424));
try{shadow.cljs.devtools.client.browser.do_js_load(sources__$1);

if(cljs.core.seq(js_requires)){
shadow.cljs.devtools.client.browser.do_js_requires(js_requires);
} else {
}

return (done.cljs$core$IFn$_invoke$arity$1 ? done.cljs$core$IFn$_invoke$arity$1(sources_to_load) : done.call(null, sources_to_load));
}catch (e38044){var ex = e38044;
return (error.cljs$core$IFn$_invoke$arity$1 ? error.cljs$core$IFn$_invoke$arity$1(ex) : error.call(null, ex));
}})], null));
}
}));

shadow.cljs.devtools.client.shared.add_plugin_BANG_(new cljs.core.Keyword("shadow.cljs.devtools.client.browser","client","shadow.cljs.devtools.client.browser/client",-1461019282),cljs.core.PersistentHashSet.EMPTY,(function (p__38045){
var map__38046 = p__38045;
var map__38046__$1 = cljs.core.__destructure_map(map__38046);
var env = map__38046__$1;
var runtime = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38046__$1,new cljs.core.Keyword(null,"runtime","runtime",-1331573996));
var svc = new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"runtime","runtime",-1331573996),runtime], null);
shadow.remote.runtime.api.add_extension(runtime,new cljs.core.Keyword("shadow.cljs.devtools.client.browser","client","shadow.cljs.devtools.client.browser/client",-1461019282),new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"on-welcome","on-welcome",1895317125),(function (){
cljs.core.reset_BANG_(shadow.cljs.devtools.client.browser.ws_was_welcome_ref,true);

shadow.cljs.devtools.client.hud.connection_error_clear_BANG_();

shadow.cljs.devtools.client.env.patch_goog_BANG_();

return shadow.cljs.devtools.client.browser.devtools_msg(["#",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"client-id","client-id",-464622140).cljs$core$IFn$_invoke$arity$1(cljs.core.deref(new cljs.core.Keyword(null,"state-ref","state-ref",2127874952).cljs$core$IFn$_invoke$arity$1(runtime))))," ready!"].join(''));
}),new cljs.core.Keyword(null,"on-disconnect","on-disconnect",-809021814),(function (e){
if(cljs.core.truth_(cljs.core.deref(shadow.cljs.devtools.client.browser.ws_was_welcome_ref))){
shadow.cljs.devtools.client.hud.connection_error("The Websocket connection was closed!");

return cljs.core.reset_BANG_(shadow.cljs.devtools.client.browser.ws_was_welcome_ref,false);
} else {
return null;
}
}),new cljs.core.Keyword(null,"on-reconnect","on-reconnect",1239988702),(function (e){
return shadow.cljs.devtools.client.hud.connection_error("Reconnecting ...");
}),new cljs.core.Keyword(null,"ops","ops",1237330063),new cljs.core.PersistentArrayMap(null, 7, [new cljs.core.Keyword(null,"access-denied","access-denied",959449406),(function (msg){
cljs.core.reset_BANG_(shadow.cljs.devtools.client.browser.ws_was_welcome_ref,false);

return shadow.cljs.devtools.client.hud.connection_error(["Stale Output! Your loaded JS was not produced by the running shadow-cljs instance."," Is the watch for this build running?"].join(''));
}),new cljs.core.Keyword(null,"cljs-asset-update","cljs-asset-update",1224093028),(function (msg){
return shadow.cljs.devtools.client.browser.handle_asset_update(msg);
}),new cljs.core.Keyword(null,"cljs-build-configure","cljs-build-configure",-2089891268),(function (msg){
return null;
}),new cljs.core.Keyword(null,"cljs-build-start","cljs-build-start",-725781241),(function (msg){
shadow.cljs.devtools.client.hud.hud_hide();

shadow.cljs.devtools.client.hud.load_start();

return shadow.cljs.devtools.client.env.run_custom_notify_BANG_(cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(msg,new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"build-start","build-start",-959649480)));
}),new cljs.core.Keyword(null,"cljs-build-complete","cljs-build-complete",273626153),(function (msg){
var msg__$1 = shadow.cljs.devtools.client.env.add_warnings_to_info(msg);
shadow.cljs.devtools.client.hud.connection_error_clear_BANG_();

shadow.cljs.devtools.client.hud.hud_warnings(msg__$1);

shadow.cljs.devtools.client.browser.handle_build_complete(runtime,msg__$1);

return shadow.cljs.devtools.client.env.run_custom_notify_BANG_(cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(msg__$1,new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"build-complete","build-complete",-501868472)));
}),new cljs.core.Keyword(null,"cljs-build-failure","cljs-build-failure",1718154990),(function (msg){
shadow.cljs.devtools.client.hud.load_end();

shadow.cljs.devtools.client.hud.hud_error(msg);

return shadow.cljs.devtools.client.env.run_custom_notify_BANG_(cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(msg,new cljs.core.Keyword(null,"type","type",1174270348),new cljs.core.Keyword(null,"build-failure","build-failure",-2107487466)));
}),new cljs.core.Keyword("shadow.cljs.devtools.client.env","worker-notify","shadow.cljs.devtools.client.env/worker-notify",-1456820670),(function (p__38055){
var map__38056 = p__38055;
var map__38056__$1 = cljs.core.__destructure_map(map__38056);
var event_op = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38056__$1,new cljs.core.Keyword(null,"event-op","event-op",200358057));
var client_id = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38056__$1,new cljs.core.Keyword(null,"client-id","client-id",-464622140));
if(((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"client-disconnect","client-disconnect",640227957),event_op)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(client_id,shadow.cljs.devtools.client.env.worker_client_id)))){
shadow.cljs.devtools.client.hud.connection_error_clear_BANG_();

return shadow.cljs.devtools.client.hud.connection_error("The watch for this build was stopped!");
} else {
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"client-connect","client-connect",-1113973888),event_op)){
shadow.cljs.devtools.client.hud.connection_error_clear_BANG_();

return shadow.cljs.devtools.client.hud.connection_error("The watch for this build was restarted. Reload required!");
} else {
return null;
}
}
})], null)], null));

return svc;
}),(function (p__38057){
var map__38058 = p__38057;
var map__38058__$1 = cljs.core.__destructure_map(map__38058);
var svc = map__38058__$1;
var runtime = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__38058__$1,new cljs.core.Keyword(null,"runtime","runtime",-1331573996));
return shadow.remote.runtime.api.del_extension(runtime,new cljs.core.Keyword("shadow.cljs.devtools.client.browser","client","shadow.cljs.devtools.client.browser/client",-1461019282));
}));

shadow.cljs.devtools.client.shared.init_runtime_BANG_(shadow.cljs.devtools.client.browser.client_info,shadow.cljs.devtools.client.websocket.start,shadow.cljs.devtools.client.websocket.send,shadow.cljs.devtools.client.websocket.stop);
} else {
}

//# sourceMappingURL=shadow.cljs.devtools.client.browser.js.map
