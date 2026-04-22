goog.provide('shadow.dom');
shadow.dom.transition_supported_QMARK_ = true;

/**
 * @interface
 */
shadow.dom.IElement = function(){};

var shadow$dom$IElement$_to_dom$dyn_29330 = (function (this$){
var x__5350__auto__ = (((this$ == null))?null:this$);
var m__5351__auto__ = (shadow.dom._to_dom[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5351__auto__.call(null, this$));
} else {
var m__5349__auto__ = (shadow.dom._to_dom["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5349__auto__.call(null, this$));
} else {
throw cljs.core.missing_protocol("IElement.-to-dom",this$);
}
}
});
shadow.dom._to_dom = (function shadow$dom$_to_dom(this$){
if((((!((this$ == null)))) && ((!((this$.shadow$dom$IElement$_to_dom$arity$1 == null)))))){
return this$.shadow$dom$IElement$_to_dom$arity$1(this$);
} else {
return shadow$dom$IElement$_to_dom$dyn_29330(this$);
}
});


/**
 * @interface
 */
shadow.dom.SVGElement = function(){};

var shadow$dom$SVGElement$_to_svg$dyn_29331 = (function (this$){
var x__5350__auto__ = (((this$ == null))?null:this$);
var m__5351__auto__ = (shadow.dom._to_svg[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5351__auto__.call(null, this$));
} else {
var m__5349__auto__ = (shadow.dom._to_svg["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(this$) : m__5349__auto__.call(null, this$));
} else {
throw cljs.core.missing_protocol("SVGElement.-to-svg",this$);
}
}
});
shadow.dom._to_svg = (function shadow$dom$_to_svg(this$){
if((((!((this$ == null)))) && ((!((this$.shadow$dom$SVGElement$_to_svg$arity$1 == null)))))){
return this$.shadow$dom$SVGElement$_to_svg$arity$1(this$);
} else {
return shadow$dom$SVGElement$_to_svg$dyn_29331(this$);
}
});

shadow.dom.lazy_native_coll_seq = (function shadow$dom$lazy_native_coll_seq(coll,idx){
if((idx < coll.length)){
return (new cljs.core.LazySeq(null,(function (){
return cljs.core.cons((coll[idx]),(function (){var G__28206 = coll;
var G__28207 = (idx + (1));
return (shadow.dom.lazy_native_coll_seq.cljs$core$IFn$_invoke$arity$2 ? shadow.dom.lazy_native_coll_seq.cljs$core$IFn$_invoke$arity$2(G__28206,G__28207) : shadow.dom.lazy_native_coll_seq.call(null, G__28206,G__28207));
})());
}),null,null));
} else {
return null;
}
});

/**
* @constructor
 * @implements {cljs.core.IIndexed}
 * @implements {cljs.core.ICounted}
 * @implements {cljs.core.ISeqable}
 * @implements {cljs.core.IDeref}
 * @implements {shadow.dom.IElement}
*/
shadow.dom.NativeColl = (function (coll){
this.coll = coll;
this.cljs$lang$protocol_mask$partition0$ = 8421394;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(shadow.dom.NativeColl.prototype.cljs$core$IDeref$_deref$arity$1 = (function (this$){
var self__ = this;
var this$__$1 = this;
return self__.coll;
}));

(shadow.dom.NativeColl.prototype.cljs$core$IIndexed$_nth$arity$2 = (function (this$,n){
var self__ = this;
var this$__$1 = this;
return (self__.coll[n]);
}));

(shadow.dom.NativeColl.prototype.cljs$core$IIndexed$_nth$arity$3 = (function (this$,n,not_found){
var self__ = this;
var this$__$1 = this;
var or__5002__auto__ = (self__.coll[n]);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return not_found;
}
}));

(shadow.dom.NativeColl.prototype.cljs$core$ICounted$_count$arity$1 = (function (this$){
var self__ = this;
var this$__$1 = this;
return self__.coll.length;
}));

(shadow.dom.NativeColl.prototype.cljs$core$ISeqable$_seq$arity$1 = (function (this$){
var self__ = this;
var this$__$1 = this;
return shadow.dom.lazy_native_coll_seq(self__.coll,(0));
}));

(shadow.dom.NativeColl.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(shadow.dom.NativeColl.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var self__ = this;
var this$__$1 = this;
return self__.coll;
}));

(shadow.dom.NativeColl.getBasis = (function (){
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"coll","coll",-1006698606,null)], null);
}));

(shadow.dom.NativeColl.cljs$lang$type = true);

(shadow.dom.NativeColl.cljs$lang$ctorStr = "shadow.dom/NativeColl");

(shadow.dom.NativeColl.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"shadow.dom/NativeColl");
}));

/**
 * Positional factory function for shadow.dom/NativeColl.
 */
shadow.dom.__GT_NativeColl = (function shadow$dom$__GT_NativeColl(coll){
return (new shadow.dom.NativeColl(coll));
});

shadow.dom.native_coll = (function shadow$dom$native_coll(coll){
return (new shadow.dom.NativeColl(coll));
});
shadow.dom.dom_node = (function shadow$dom$dom_node(el){
if((el == null)){
return null;
} else {
if((((!((el == null))))?((((false) || ((cljs.core.PROTOCOL_SENTINEL === el.shadow$dom$IElement$))))?true:false):false)){
return el.shadow$dom$IElement$_to_dom$arity$1(null, );
} else {
if(typeof el === 'string'){
return document.createTextNode(el);
} else {
if(typeof el === 'number'){
return document.createTextNode(cljs.core.str.cljs$core$IFn$_invoke$arity$1(el));
} else {
return el;

}
}
}
}
});
shadow.dom.query_one = (function shadow$dom$query_one(var_args){
var G__28283 = arguments.length;
switch (G__28283) {
case 1:
return shadow.dom.query_one.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.query_one.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.query_one.cljs$core$IFn$_invoke$arity$1 = (function (sel){
return document.querySelector(sel);
}));

(shadow.dom.query_one.cljs$core$IFn$_invoke$arity$2 = (function (sel,root){
return shadow.dom.dom_node(root).querySelector(sel);
}));

(shadow.dom.query_one.cljs$lang$maxFixedArity = 2);

shadow.dom.query = (function shadow$dom$query(var_args){
var G__28289 = arguments.length;
switch (G__28289) {
case 1:
return shadow.dom.query.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.query.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.query.cljs$core$IFn$_invoke$arity$1 = (function (sel){
return (new shadow.dom.NativeColl(document.querySelectorAll(sel)));
}));

(shadow.dom.query.cljs$core$IFn$_invoke$arity$2 = (function (sel,root){
return (new shadow.dom.NativeColl(shadow.dom.dom_node(root).querySelectorAll(sel)));
}));

(shadow.dom.query.cljs$lang$maxFixedArity = 2);

shadow.dom.by_id = (function shadow$dom$by_id(var_args){
var G__28292 = arguments.length;
switch (G__28292) {
case 2:
return shadow.dom.by_id.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 1:
return shadow.dom.by_id.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.by_id.cljs$core$IFn$_invoke$arity$2 = (function (id,el){
return shadow.dom.dom_node(el).getElementById(id);
}));

(shadow.dom.by_id.cljs$core$IFn$_invoke$arity$1 = (function (id){
return document.getElementById(id);
}));

(shadow.dom.by_id.cljs$lang$maxFixedArity = 2);

shadow.dom.build = shadow.dom.dom_node;
shadow.dom.ev_stop = (function shadow$dom$ev_stop(var_args){
var G__28304 = arguments.length;
switch (G__28304) {
case 1:
return shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 4:
return shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$1 = (function (e){
if(cljs.core.truth_(e.stopPropagation)){
e.stopPropagation();

e.preventDefault();
} else {
(e.cancelBubble = true);

(e.returnValue = false);
}

return e;
}));

(shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$2 = (function (e,el){
shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$1(e);

return el;
}));

(shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$4 = (function (e,el,scope,owner){
shadow.dom.ev_stop.cljs$core$IFn$_invoke$arity$1(e);

return el;
}));

(shadow.dom.ev_stop.cljs$lang$maxFixedArity = 4);

/**
 * check wether a parent node (or the document) contains the child
 */
shadow.dom.contains_QMARK_ = (function shadow$dom$contains_QMARK_(var_args){
var G__28310 = arguments.length;
switch (G__28310) {
case 1:
return shadow.dom.contains_QMARK_.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.contains_QMARK_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.contains_QMARK_.cljs$core$IFn$_invoke$arity$1 = (function (el){
return goog.dom.contains(document,shadow.dom.dom_node(el));
}));

(shadow.dom.contains_QMARK_.cljs$core$IFn$_invoke$arity$2 = (function (parent,el){
return goog.dom.contains(shadow.dom.dom_node(parent),shadow.dom.dom_node(el));
}));

(shadow.dom.contains_QMARK_.cljs$lang$maxFixedArity = 2);

shadow.dom.add_class = (function shadow$dom$add_class(el,cls){
return goog.dom.classlist.add(shadow.dom.dom_node(el),cls);
});
shadow.dom.remove_class = (function shadow$dom$remove_class(el,cls){
return goog.dom.classlist.remove(shadow.dom.dom_node(el),cls);
});
shadow.dom.toggle_class = (function shadow$dom$toggle_class(var_args){
var G__28343 = arguments.length;
switch (G__28343) {
case 2:
return shadow.dom.toggle_class.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return shadow.dom.toggle_class.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.toggle_class.cljs$core$IFn$_invoke$arity$2 = (function (el,cls){
return goog.dom.classlist.toggle(shadow.dom.dom_node(el),cls);
}));

(shadow.dom.toggle_class.cljs$core$IFn$_invoke$arity$3 = (function (el,cls,v){
if(cljs.core.truth_(v)){
return shadow.dom.add_class(el,cls);
} else {
return shadow.dom.remove_class(el,cls);
}
}));

(shadow.dom.toggle_class.cljs$lang$maxFixedArity = 3);

shadow.dom.dom_listen = (cljs.core.truth_((function (){var or__5002__auto__ = (!((typeof document !== 'undefined')));
if(or__5002__auto__){
return or__5002__auto__;
} else {
return document.addEventListener;
}
})())?(function shadow$dom$dom_listen_good(el,ev,handler){
return el.addEventListener(ev,handler,false);
}):(function shadow$dom$dom_listen_ie(el,ev,handler){
try{return el.attachEvent(["on",cljs.core.str.cljs$core$IFn$_invoke$arity$1(ev)].join(''),(function (e){
return (handler.cljs$core$IFn$_invoke$arity$2 ? handler.cljs$core$IFn$_invoke$arity$2(e,el) : handler.call(null, e,el));
}));
}catch (e28345){if((e28345 instanceof Object)){
var e = e28345;
return console.log("didnt support attachEvent",el,e);
} else {
throw e28345;

}
}}));
shadow.dom.dom_listen_remove = (cljs.core.truth_((function (){var or__5002__auto__ = (!((typeof document !== 'undefined')));
if(or__5002__auto__){
return or__5002__auto__;
} else {
return document.removeEventListener;
}
})())?(function shadow$dom$dom_listen_remove_good(el,ev,handler){
return el.removeEventListener(ev,handler,false);
}):(function shadow$dom$dom_listen_remove_ie(el,ev,handler){
return el.detachEvent(["on",cljs.core.str.cljs$core$IFn$_invoke$arity$1(ev)].join(''),handler);
}));
shadow.dom.on_query = (function shadow$dom$on_query(root_el,ev,selector,handler){
var seq__28346 = cljs.core.seq(shadow.dom.query.cljs$core$IFn$_invoke$arity$2(selector,root_el));
var chunk__28347 = null;
var count__28348 = (0);
var i__28349 = (0);
while(true){
if((i__28349 < count__28348)){
var el = chunk__28347.cljs$core$IIndexed$_nth$arity$2(null, i__28349);
var handler_29370__$1 = ((function (seq__28346,chunk__28347,count__28348,i__28349,el){
return (function (e){
return (handler.cljs$core$IFn$_invoke$arity$2 ? handler.cljs$core$IFn$_invoke$arity$2(e,el) : handler.call(null, e,el));
});})(seq__28346,chunk__28347,count__28348,i__28349,el))
;
shadow.dom.dom_listen(el,cljs.core.name(ev),handler_29370__$1);


var G__29373 = seq__28346;
var G__29374 = chunk__28347;
var G__29375 = count__28348;
var G__29376 = (i__28349 + (1));
seq__28346 = G__29373;
chunk__28347 = G__29374;
count__28348 = G__29375;
i__28349 = G__29376;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__28346);
if(temp__5825__auto__){
var seq__28346__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__28346__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__28346__$1);
var G__29377 = cljs.core.chunk_rest(seq__28346__$1);
var G__29378 = c__5525__auto__;
var G__29379 = cljs.core.count(c__5525__auto__);
var G__29380 = (0);
seq__28346 = G__29377;
chunk__28347 = G__29378;
count__28348 = G__29379;
i__28349 = G__29380;
continue;
} else {
var el = cljs.core.first(seq__28346__$1);
var handler_29381__$1 = ((function (seq__28346,chunk__28347,count__28348,i__28349,el,seq__28346__$1,temp__5825__auto__){
return (function (e){
return (handler.cljs$core$IFn$_invoke$arity$2 ? handler.cljs$core$IFn$_invoke$arity$2(e,el) : handler.call(null, e,el));
});})(seq__28346,chunk__28347,count__28348,i__28349,el,seq__28346__$1,temp__5825__auto__))
;
shadow.dom.dom_listen(el,cljs.core.name(ev),handler_29381__$1);


var G__29382 = cljs.core.next(seq__28346__$1);
var G__29383 = null;
var G__29384 = (0);
var G__29385 = (0);
seq__28346 = G__29382;
chunk__28347 = G__29383;
count__28348 = G__29384;
i__28349 = G__29385;
continue;
}
} else {
return null;
}
}
break;
}
});
shadow.dom.on = (function shadow$dom$on(var_args){
var G__28359 = arguments.length;
switch (G__28359) {
case 3:
return shadow.dom.on.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
case 4:
return shadow.dom.on.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.on.cljs$core$IFn$_invoke$arity$3 = (function (el,ev,handler){
return shadow.dom.on.cljs$core$IFn$_invoke$arity$4(el,ev,handler,false);
}));

(shadow.dom.on.cljs$core$IFn$_invoke$arity$4 = (function (el,ev,handler,capture){
if(cljs.core.vector_QMARK_(ev)){
return shadow.dom.on_query(el,cljs.core.first(ev),cljs.core.second(ev),handler);
} else {
var handler__$1 = (function (e){
return (handler.cljs$core$IFn$_invoke$arity$2 ? handler.cljs$core$IFn$_invoke$arity$2(e,el) : handler.call(null, e,el));
});
return shadow.dom.dom_listen(shadow.dom.dom_node(el),cljs.core.name(ev),handler__$1);
}
}));

(shadow.dom.on.cljs$lang$maxFixedArity = 4);

shadow.dom.remove_event_handler = (function shadow$dom$remove_event_handler(el,ev,handler){
return shadow.dom.dom_listen_remove(shadow.dom.dom_node(el),cljs.core.name(ev),handler);
});
shadow.dom.add_event_listeners = (function shadow$dom$add_event_listeners(el,events){
var seq__28361 = cljs.core.seq(events);
var chunk__28362 = null;
var count__28363 = (0);
var i__28364 = (0);
while(true){
if((i__28364 < count__28363)){
var vec__28371 = chunk__28362.cljs$core$IIndexed$_nth$arity$2(null, i__28364);
var k = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28371,(0),null);
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28371,(1),null);
shadow.dom.on.cljs$core$IFn$_invoke$arity$3(el,k,v);


var G__29400 = seq__28361;
var G__29401 = chunk__28362;
var G__29402 = count__28363;
var G__29403 = (i__28364 + (1));
seq__28361 = G__29400;
chunk__28362 = G__29401;
count__28363 = G__29402;
i__28364 = G__29403;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__28361);
if(temp__5825__auto__){
var seq__28361__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__28361__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__28361__$1);
var G__29405 = cljs.core.chunk_rest(seq__28361__$1);
var G__29406 = c__5525__auto__;
var G__29407 = cljs.core.count(c__5525__auto__);
var G__29408 = (0);
seq__28361 = G__29405;
chunk__28362 = G__29406;
count__28363 = G__29407;
i__28364 = G__29408;
continue;
} else {
var vec__28374 = cljs.core.first(seq__28361__$1);
var k = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28374,(0),null);
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28374,(1),null);
shadow.dom.on.cljs$core$IFn$_invoke$arity$3(el,k,v);


var G__29409 = cljs.core.next(seq__28361__$1);
var G__29410 = null;
var G__29411 = (0);
var G__29412 = (0);
seq__28361 = G__29409;
chunk__28362 = G__29410;
count__28363 = G__29411;
i__28364 = G__29412;
continue;
}
} else {
return null;
}
}
break;
}
});
shadow.dom.set_style = (function shadow$dom$set_style(el,styles){
var dom = shadow.dom.dom_node(el);
var seq__28377 = cljs.core.seq(styles);
var chunk__28378 = null;
var count__28379 = (0);
var i__28380 = (0);
while(true){
if((i__28380 < count__28379)){
var vec__28387 = chunk__28378.cljs$core$IIndexed$_nth$arity$2(null, i__28380);
var k = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28387,(0),null);
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28387,(1),null);
goog.style.setStyle(dom,cljs.core.name(k),(((v == null))?"":v));


var G__29414 = seq__28377;
var G__29415 = chunk__28378;
var G__29416 = count__28379;
var G__29417 = (i__28380 + (1));
seq__28377 = G__29414;
chunk__28378 = G__29415;
count__28379 = G__29416;
i__28380 = G__29417;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__28377);
if(temp__5825__auto__){
var seq__28377__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__28377__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__28377__$1);
var G__29419 = cljs.core.chunk_rest(seq__28377__$1);
var G__29420 = c__5525__auto__;
var G__29421 = cljs.core.count(c__5525__auto__);
var G__29422 = (0);
seq__28377 = G__29419;
chunk__28378 = G__29420;
count__28379 = G__29421;
i__28380 = G__29422;
continue;
} else {
var vec__28401 = cljs.core.first(seq__28377__$1);
var k = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28401,(0),null);
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28401,(1),null);
goog.style.setStyle(dom,cljs.core.name(k),(((v == null))?"":v));


var G__29423 = cljs.core.next(seq__28377__$1);
var G__29424 = null;
var G__29425 = (0);
var G__29426 = (0);
seq__28377 = G__29423;
chunk__28378 = G__29424;
count__28379 = G__29425;
i__28380 = G__29426;
continue;
}
} else {
return null;
}
}
break;
}
});
shadow.dom.set_attr_STAR_ = (function shadow$dom$set_attr_STAR_(el,key,value){
var G__28404_29427 = key;
var G__28404_29428__$1 = (((G__28404_29427 instanceof cljs.core.Keyword))?G__28404_29427.fqn:null);
switch (G__28404_29428__$1) {
case "id":
(el.id = cljs.core.str.cljs$core$IFn$_invoke$arity$1(value));

break;
case "class":
(el.className = cljs.core.str.cljs$core$IFn$_invoke$arity$1(value));

break;
case "for":
(el.htmlFor = value);

break;
case "cellpadding":
el.setAttribute("cellPadding",value);

break;
case "cellspacing":
el.setAttribute("cellSpacing",value);

break;
case "colspan":
el.setAttribute("colSpan",value);

break;
case "frameborder":
el.setAttribute("frameBorder",value);

break;
case "height":
el.setAttribute("height",value);

break;
case "maxlength":
el.setAttribute("maxLength",value);

break;
case "role":
el.setAttribute("role",value);

break;
case "rowspan":
el.setAttribute("rowSpan",value);

break;
case "type":
el.setAttribute("type",value);

break;
case "usemap":
el.setAttribute("useMap",value);

break;
case "valign":
el.setAttribute("vAlign",value);

break;
case "width":
el.setAttribute("width",value);

break;
case "on":
shadow.dom.add_event_listeners(el,value);

break;
case "style":
if((value == null)){
} else {
if(typeof value === 'string'){
el.setAttribute("style",value);
} else {
if(cljs.core.map_QMARK_(value)){
shadow.dom.set_style(el,value);
} else {
goog.style.setStyle(el,value);

}
}
}

break;
default:
var ks_29432 = cljs.core.name(key);
if(cljs.core.truth_((function (){var or__5002__auto__ = goog.string.startsWith(ks_29432,"data-");
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return goog.string.startsWith(ks_29432,"aria-");
}
})())){
el.setAttribute(ks_29432,value);
} else {
(el[ks_29432] = value);
}

}

return el;
});
shadow.dom.set_attrs = (function shadow$dom$set_attrs(el,attrs){
return cljs.core.reduce_kv((function (el__$1,key,value){
shadow.dom.set_attr_STAR_(el__$1,key,value);

return el__$1;
}),shadow.dom.dom_node(el),attrs);
});
shadow.dom.set_attr = (function shadow$dom$set_attr(el,key,value){
return shadow.dom.set_attr_STAR_(shadow.dom.dom_node(el),key,value);
});
shadow.dom.has_class_QMARK_ = (function shadow$dom$has_class_QMARK_(el,cls){
return goog.dom.classlist.contains(shadow.dom.dom_node(el),cls);
});
shadow.dom.merge_class_string = (function shadow$dom$merge_class_string(current,extra_class){
if(cljs.core.seq(current)){
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(current)," ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(extra_class)].join('');
} else {
return extra_class;
}
});
shadow.dom.parse_tag = (function shadow$dom$parse_tag(spec){
var spec__$1 = cljs.core.name(spec);
var fdot = spec__$1.indexOf(".");
var fhash = spec__$1.indexOf("#");
if(((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2((-1),fdot)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2((-1),fhash)))){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [spec__$1,null,null], null);
} else {
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2((-1),fhash)){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [spec__$1.substring((0),fdot),null,clojure.string.replace(spec__$1.substring((fdot + (1))),/\./," ")], null);
} else {
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2((-1),fdot)){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [spec__$1.substring((0),fhash),spec__$1.substring((fhash + (1))),null], null);
} else {
if((fhash > fdot)){
throw ["cant have id after class?",spec__$1].join('');
} else {
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [spec__$1.substring((0),fhash),spec__$1.substring((fhash + (1)),fdot),clojure.string.replace(spec__$1.substring((fdot + (1))),/\./," ")], null);

}
}
}
}
});
shadow.dom.create_dom_node = (function shadow$dom$create_dom_node(tag_def,p__28463){
var map__28464 = p__28463;
var map__28464__$1 = cljs.core.__destructure_map(map__28464);
var props = map__28464__$1;
var class$ = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__28464__$1,new cljs.core.Keyword(null,"class","class",-2030961996));
var tag_props = ({});
var vec__28465 = shadow.dom.parse_tag(tag_def);
var tag_name = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28465,(0),null);
var tag_id = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28465,(1),null);
var tag_classes = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28465,(2),null);
if(cljs.core.truth_(tag_id)){
(tag_props["id"] = tag_id);
} else {
}

if(cljs.core.truth_(tag_classes)){
(tag_props["class"] = shadow.dom.merge_class_string(class$,tag_classes));
} else {
}

var G__28468 = goog.dom.createDom(tag_name,tag_props);
shadow.dom.set_attrs(G__28468,cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(props,new cljs.core.Keyword(null,"class","class",-2030961996)));

return G__28468;
});
shadow.dom.append = (function shadow$dom$append(var_args){
var G__28470 = arguments.length;
switch (G__28470) {
case 1:
return shadow.dom.append.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.append.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.append.cljs$core$IFn$_invoke$arity$1 = (function (node){
if(cljs.core.truth_(node)){
var temp__5825__auto__ = shadow.dom.dom_node(node);
if(cljs.core.truth_(temp__5825__auto__)){
var n = temp__5825__auto__;
document.body.appendChild(n);

return n;
} else {
return null;
}
} else {
return null;
}
}));

(shadow.dom.append.cljs$core$IFn$_invoke$arity$2 = (function (el,node){
if(cljs.core.truth_(node)){
var temp__5825__auto__ = shadow.dom.dom_node(node);
if(cljs.core.truth_(temp__5825__auto__)){
var n = temp__5825__auto__;
shadow.dom.dom_node(el).appendChild(n);

return n;
} else {
return null;
}
} else {
return null;
}
}));

(shadow.dom.append.cljs$lang$maxFixedArity = 2);

shadow.dom.destructure_node = (function shadow$dom$destructure_node(create_fn,p__28471){
var vec__28472 = p__28471;
var seq__28473 = cljs.core.seq(vec__28472);
var first__28474 = cljs.core.first(seq__28473);
var seq__28473__$1 = cljs.core.next(seq__28473);
var nn = first__28474;
var first__28474__$1 = cljs.core.first(seq__28473__$1);
var seq__28473__$2 = cljs.core.next(seq__28473__$1);
var np = first__28474__$1;
var nc = seq__28473__$2;
var node = vec__28472;
if((nn instanceof cljs.core.Keyword)){
} else {
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("invalid dom node",new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"node","node",581201198),node], null));
}

if((((np == null)) && ((nc == null)))){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){var G__28475 = nn;
var G__28476 = cljs.core.PersistentArrayMap.EMPTY;
return (create_fn.cljs$core$IFn$_invoke$arity$2 ? create_fn.cljs$core$IFn$_invoke$arity$2(G__28475,G__28476) : create_fn.call(null, G__28475,G__28476));
})(),cljs.core.List.EMPTY], null);
} else {
if(cljs.core.map_QMARK_(np)){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(create_fn.cljs$core$IFn$_invoke$arity$2 ? create_fn.cljs$core$IFn$_invoke$arity$2(nn,np) : create_fn.call(null, nn,np)),nc], null);
} else {
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(function (){var G__28477 = nn;
var G__28478 = cljs.core.PersistentArrayMap.EMPTY;
return (create_fn.cljs$core$IFn$_invoke$arity$2 ? create_fn.cljs$core$IFn$_invoke$arity$2(G__28477,G__28478) : create_fn.call(null, G__28477,G__28478));
})(),cljs.core.conj.cljs$core$IFn$_invoke$arity$2(nc,np)], null);

}
}
});
shadow.dom.make_dom_node = (function shadow$dom$make_dom_node(structure){
var vec__28484 = shadow.dom.destructure_node(shadow.dom.create_dom_node,structure);
var node = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28484,(0),null);
var node_children = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28484,(1),null);
var seq__28487_29485 = cljs.core.seq(node_children);
var chunk__28488_29486 = null;
var count__28489_29487 = (0);
var i__28490_29488 = (0);
while(true){
if((i__28490_29488 < count__28489_29487)){
var child_struct_29489 = chunk__28488_29486.cljs$core$IIndexed$_nth$arity$2(null, i__28490_29488);
var children_29490 = shadow.dom.dom_node(child_struct_29489);
if(cljs.core.seq_QMARK_(children_29490)){
var seq__28528_29491 = cljs.core.seq(cljs.core.map.cljs$core$IFn$_invoke$arity$2(shadow.dom.dom_node,children_29490));
var chunk__28530_29492 = null;
var count__28531_29493 = (0);
var i__28532_29494 = (0);
while(true){
if((i__28532_29494 < count__28531_29493)){
var child_29495 = chunk__28530_29492.cljs$core$IIndexed$_nth$arity$2(null, i__28532_29494);
if(cljs.core.truth_(child_29495)){
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,child_29495);


var G__29497 = seq__28528_29491;
var G__29498 = chunk__28530_29492;
var G__29499 = count__28531_29493;
var G__29500 = (i__28532_29494 + (1));
seq__28528_29491 = G__29497;
chunk__28530_29492 = G__29498;
count__28531_29493 = G__29499;
i__28532_29494 = G__29500;
continue;
} else {
var G__29505 = seq__28528_29491;
var G__29506 = chunk__28530_29492;
var G__29507 = count__28531_29493;
var G__29508 = (i__28532_29494 + (1));
seq__28528_29491 = G__29505;
chunk__28530_29492 = G__29506;
count__28531_29493 = G__29507;
i__28532_29494 = G__29508;
continue;
}
} else {
var temp__5825__auto___29510 = cljs.core.seq(seq__28528_29491);
if(temp__5825__auto___29510){
var seq__28528_29512__$1 = temp__5825__auto___29510;
if(cljs.core.chunked_seq_QMARK_(seq__28528_29512__$1)){
var c__5525__auto___29513 = cljs.core.chunk_first(seq__28528_29512__$1);
var G__29514 = cljs.core.chunk_rest(seq__28528_29512__$1);
var G__29515 = c__5525__auto___29513;
var G__29516 = cljs.core.count(c__5525__auto___29513);
var G__29517 = (0);
seq__28528_29491 = G__29514;
chunk__28530_29492 = G__29515;
count__28531_29493 = G__29516;
i__28532_29494 = G__29517;
continue;
} else {
var child_29518 = cljs.core.first(seq__28528_29512__$1);
if(cljs.core.truth_(child_29518)){
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,child_29518);


var G__29519 = cljs.core.next(seq__28528_29512__$1);
var G__29520 = null;
var G__29521 = (0);
var G__29522 = (0);
seq__28528_29491 = G__29519;
chunk__28530_29492 = G__29520;
count__28531_29493 = G__29521;
i__28532_29494 = G__29522;
continue;
} else {
var G__29523 = cljs.core.next(seq__28528_29512__$1);
var G__29524 = null;
var G__29525 = (0);
var G__29526 = (0);
seq__28528_29491 = G__29523;
chunk__28530_29492 = G__29524;
count__28531_29493 = G__29525;
i__28532_29494 = G__29526;
continue;
}
}
} else {
}
}
break;
}
} else {
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,children_29490);
}


var G__29528 = seq__28487_29485;
var G__29529 = chunk__28488_29486;
var G__29530 = count__28489_29487;
var G__29531 = (i__28490_29488 + (1));
seq__28487_29485 = G__29528;
chunk__28488_29486 = G__29529;
count__28489_29487 = G__29530;
i__28490_29488 = G__29531;
continue;
} else {
var temp__5825__auto___29532 = cljs.core.seq(seq__28487_29485);
if(temp__5825__auto___29532){
var seq__28487_29534__$1 = temp__5825__auto___29532;
if(cljs.core.chunked_seq_QMARK_(seq__28487_29534__$1)){
var c__5525__auto___29535 = cljs.core.chunk_first(seq__28487_29534__$1);
var G__29537 = cljs.core.chunk_rest(seq__28487_29534__$1);
var G__29538 = c__5525__auto___29535;
var G__29539 = cljs.core.count(c__5525__auto___29535);
var G__29540 = (0);
seq__28487_29485 = G__29537;
chunk__28488_29486 = G__29538;
count__28489_29487 = G__29539;
i__28490_29488 = G__29540;
continue;
} else {
var child_struct_29541 = cljs.core.first(seq__28487_29534__$1);
var children_29543 = shadow.dom.dom_node(child_struct_29541);
if(cljs.core.seq_QMARK_(children_29543)){
var seq__28545_29544 = cljs.core.seq(cljs.core.map.cljs$core$IFn$_invoke$arity$2(shadow.dom.dom_node,children_29543));
var chunk__28547_29545 = null;
var count__28548_29547 = (0);
var i__28549_29548 = (0);
while(true){
if((i__28549_29548 < count__28548_29547)){
var child_29549 = chunk__28547_29545.cljs$core$IIndexed$_nth$arity$2(null, i__28549_29548);
if(cljs.core.truth_(child_29549)){
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,child_29549);


var G__29550 = seq__28545_29544;
var G__29551 = chunk__28547_29545;
var G__29552 = count__28548_29547;
var G__29553 = (i__28549_29548 + (1));
seq__28545_29544 = G__29550;
chunk__28547_29545 = G__29551;
count__28548_29547 = G__29552;
i__28549_29548 = G__29553;
continue;
} else {
var G__29555 = seq__28545_29544;
var G__29556 = chunk__28547_29545;
var G__29557 = count__28548_29547;
var G__29558 = (i__28549_29548 + (1));
seq__28545_29544 = G__29555;
chunk__28547_29545 = G__29556;
count__28548_29547 = G__29557;
i__28549_29548 = G__29558;
continue;
}
} else {
var temp__5825__auto___29559__$1 = cljs.core.seq(seq__28545_29544);
if(temp__5825__auto___29559__$1){
var seq__28545_29562__$1 = temp__5825__auto___29559__$1;
if(cljs.core.chunked_seq_QMARK_(seq__28545_29562__$1)){
var c__5525__auto___29563 = cljs.core.chunk_first(seq__28545_29562__$1);
var G__29565 = cljs.core.chunk_rest(seq__28545_29562__$1);
var G__29566 = c__5525__auto___29563;
var G__29567 = cljs.core.count(c__5525__auto___29563);
var G__29568 = (0);
seq__28545_29544 = G__29565;
chunk__28547_29545 = G__29566;
count__28548_29547 = G__29567;
i__28549_29548 = G__29568;
continue;
} else {
var child_29571 = cljs.core.first(seq__28545_29562__$1);
if(cljs.core.truth_(child_29571)){
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,child_29571);


var G__29573 = cljs.core.next(seq__28545_29562__$1);
var G__29574 = null;
var G__29575 = (0);
var G__29576 = (0);
seq__28545_29544 = G__29573;
chunk__28547_29545 = G__29574;
count__28548_29547 = G__29575;
i__28549_29548 = G__29576;
continue;
} else {
var G__29577 = cljs.core.next(seq__28545_29562__$1);
var G__29578 = null;
var G__29579 = (0);
var G__29580 = (0);
seq__28545_29544 = G__29577;
chunk__28547_29545 = G__29578;
count__28548_29547 = G__29579;
i__28549_29548 = G__29580;
continue;
}
}
} else {
}
}
break;
}
} else {
shadow.dom.append.cljs$core$IFn$_invoke$arity$2(node,children_29543);
}


var G__29582 = cljs.core.next(seq__28487_29534__$1);
var G__29583 = null;
var G__29584 = (0);
var G__29585 = (0);
seq__28487_29485 = G__29582;
chunk__28488_29486 = G__29583;
count__28489_29487 = G__29584;
i__28490_29488 = G__29585;
continue;
}
} else {
}
}
break;
}

return node;
});
(cljs.core.Keyword.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.Keyword.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var this$__$1 = this;
return shadow.dom.make_dom_node(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [this$__$1], null));
}));

(cljs.core.PersistentVector.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.PersistentVector.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var this$__$1 = this;
return shadow.dom.make_dom_node(this$__$1);
}));

(cljs.core.LazySeq.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.LazySeq.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var this$__$1 = this;
return cljs.core.map.cljs$core$IFn$_invoke$arity$2(shadow.dom._to_dom,this$__$1);
}));
if(cljs.core.truth_(((typeof HTMLElement) != 'undefined'))){
(HTMLElement.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(HTMLElement.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var this$__$1 = this;
return this$__$1;
}));
} else {
}
if(cljs.core.truth_(((typeof DocumentFragment) != 'undefined'))){
(DocumentFragment.prototype.shadow$dom$IElement$ = cljs.core.PROTOCOL_SENTINEL);

(DocumentFragment.prototype.shadow$dom$IElement$_to_dom$arity$1 = (function (this$){
var this$__$1 = this;
return this$__$1;
}));
} else {
}
/**
 * clear node children
 */
shadow.dom.reset = (function shadow$dom$reset(node){
return goog.dom.removeChildren(shadow.dom.dom_node(node));
});
shadow.dom.remove = (function shadow$dom$remove(node){
if((((!((node == null))))?(((((node.cljs$lang$protocol_mask$partition0$ & (8388608))) || ((cljs.core.PROTOCOL_SENTINEL === node.cljs$core$ISeqable$))))?true:false):false)){
var seq__28707 = cljs.core.seq(node);
var chunk__28708 = null;
var count__28709 = (0);
var i__28710 = (0);
while(true){
if((i__28710 < count__28709)){
var n = chunk__28708.cljs$core$IIndexed$_nth$arity$2(null, i__28710);
(shadow.dom.remove.cljs$core$IFn$_invoke$arity$1 ? shadow.dom.remove.cljs$core$IFn$_invoke$arity$1(n) : shadow.dom.remove.call(null, n));


var G__29655 = seq__28707;
var G__29656 = chunk__28708;
var G__29657 = count__28709;
var G__29658 = (i__28710 + (1));
seq__28707 = G__29655;
chunk__28708 = G__29656;
count__28709 = G__29657;
i__28710 = G__29658;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__28707);
if(temp__5825__auto__){
var seq__28707__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__28707__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__28707__$1);
var G__29666 = cljs.core.chunk_rest(seq__28707__$1);
var G__29667 = c__5525__auto__;
var G__29668 = cljs.core.count(c__5525__auto__);
var G__29669 = (0);
seq__28707 = G__29666;
chunk__28708 = G__29667;
count__28709 = G__29668;
i__28710 = G__29669;
continue;
} else {
var n = cljs.core.first(seq__28707__$1);
(shadow.dom.remove.cljs$core$IFn$_invoke$arity$1 ? shadow.dom.remove.cljs$core$IFn$_invoke$arity$1(n) : shadow.dom.remove.call(null, n));


var G__29731 = cljs.core.next(seq__28707__$1);
var G__29732 = null;
var G__29733 = (0);
var G__29734 = (0);
seq__28707 = G__29731;
chunk__28708 = G__29732;
count__28709 = G__29733;
i__28710 = G__29734;
continue;
}
} else {
return null;
}
}
break;
}
} else {
return goog.dom.removeNode(node);
}
});
shadow.dom.replace_node = (function shadow$dom$replace_node(old,new$){
return goog.dom.replaceNode(shadow.dom.dom_node(new$),shadow.dom.dom_node(old));
});
shadow.dom.text = (function shadow$dom$text(var_args){
var G__28849 = arguments.length;
switch (G__28849) {
case 2:
return shadow.dom.text.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 1:
return shadow.dom.text.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.text.cljs$core$IFn$_invoke$arity$2 = (function (el,new_text){
return (shadow.dom.dom_node(el).innerText = new_text);
}));

(shadow.dom.text.cljs$core$IFn$_invoke$arity$1 = (function (el){
return shadow.dom.dom_node(el).innerText;
}));

(shadow.dom.text.cljs$lang$maxFixedArity = 2);

shadow.dom.check = (function shadow$dom$check(var_args){
var G__28908 = arguments.length;
switch (G__28908) {
case 1:
return shadow.dom.check.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.check.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.check.cljs$core$IFn$_invoke$arity$1 = (function (el){
return shadow.dom.check.cljs$core$IFn$_invoke$arity$2(el,true);
}));

(shadow.dom.check.cljs$core$IFn$_invoke$arity$2 = (function (el,checked){
return (shadow.dom.dom_node(el).checked = checked);
}));

(shadow.dom.check.cljs$lang$maxFixedArity = 2);

shadow.dom.checked_QMARK_ = (function shadow$dom$checked_QMARK_(el){
return shadow.dom.dom_node(el).checked;
});
shadow.dom.form_elements = (function shadow$dom$form_elements(el){
return (new shadow.dom.NativeColl(shadow.dom.dom_node(el).elements));
});
shadow.dom.children = (function shadow$dom$children(el){
return (new shadow.dom.NativeColl(shadow.dom.dom_node(el).children));
});
shadow.dom.child_nodes = (function shadow$dom$child_nodes(el){
return (new shadow.dom.NativeColl(shadow.dom.dom_node(el).childNodes));
});
shadow.dom.attr = (function shadow$dom$attr(var_args){
var G__28925 = arguments.length;
switch (G__28925) {
case 2:
return shadow.dom.attr.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return shadow.dom.attr.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.attr.cljs$core$IFn$_invoke$arity$2 = (function (el,key){
return shadow.dom.dom_node(el).getAttribute(cljs.core.name(key));
}));

(shadow.dom.attr.cljs$core$IFn$_invoke$arity$3 = (function (el,key,default$){
var or__5002__auto__ = shadow.dom.dom_node(el).getAttribute(cljs.core.name(key));
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return default$;
}
}));

(shadow.dom.attr.cljs$lang$maxFixedArity = 3);

shadow.dom.del_attr = (function shadow$dom$del_attr(el,key){
return shadow.dom.dom_node(el).removeAttribute(cljs.core.name(key));
});
shadow.dom.data = (function shadow$dom$data(el,key){
return shadow.dom.dom_node(el).getAttribute(["data-",cljs.core.name(key)].join(''));
});
shadow.dom.set_data = (function shadow$dom$set_data(el,key,value){
return shadow.dom.dom_node(el).setAttribute(["data-",cljs.core.name(key)].join(''),cljs.core.str.cljs$core$IFn$_invoke$arity$1(value));
});
shadow.dom.set_html = (function shadow$dom$set_html(node,text){
return (shadow.dom.dom_node(node).innerHTML = text);
});
shadow.dom.get_html = (function shadow$dom$get_html(node){
return shadow.dom.dom_node(node).innerHTML;
});
shadow.dom.fragment = (function shadow$dom$fragment(var_args){
var args__5732__auto__ = [];
var len__5726__auto___29809 = arguments.length;
var i__5727__auto___29813 = (0);
while(true){
if((i__5727__auto___29813 < len__5726__auto___29809)){
args__5732__auto__.push((arguments[i__5727__auto___29813]));

var G__29816 = (i__5727__auto___29813 + (1));
i__5727__auto___29813 = G__29816;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((0) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((0)),(0),null)):null);
return shadow.dom.fragment.cljs$core$IFn$_invoke$arity$variadic(argseq__5733__auto__);
});

(shadow.dom.fragment.cljs$core$IFn$_invoke$arity$variadic = (function (nodes){
var fragment = document.createDocumentFragment();
var seq__28933_29849 = cljs.core.seq(nodes);
var chunk__28934_29851 = null;
var count__28935_29852 = (0);
var i__28936_29853 = (0);
while(true){
if((i__28936_29853 < count__28935_29852)){
var node_29866 = chunk__28934_29851.cljs$core$IIndexed$_nth$arity$2(null, i__28936_29853);
fragment.appendChild(shadow.dom._to_dom(node_29866));


var G__29868 = seq__28933_29849;
var G__29869 = chunk__28934_29851;
var G__29870 = count__28935_29852;
var G__29871 = (i__28936_29853 + (1));
seq__28933_29849 = G__29868;
chunk__28934_29851 = G__29869;
count__28935_29852 = G__29870;
i__28936_29853 = G__29871;
continue;
} else {
var temp__5825__auto___29873 = cljs.core.seq(seq__28933_29849);
if(temp__5825__auto___29873){
var seq__28933_29874__$1 = temp__5825__auto___29873;
if(cljs.core.chunked_seq_QMARK_(seq__28933_29874__$1)){
var c__5525__auto___29876 = cljs.core.chunk_first(seq__28933_29874__$1);
var G__29878 = cljs.core.chunk_rest(seq__28933_29874__$1);
var G__29879 = c__5525__auto___29876;
var G__29880 = cljs.core.count(c__5525__auto___29876);
var G__29881 = (0);
seq__28933_29849 = G__29878;
chunk__28934_29851 = G__29879;
count__28935_29852 = G__29880;
i__28936_29853 = G__29881;
continue;
} else {
var node_29882 = cljs.core.first(seq__28933_29874__$1);
fragment.appendChild(shadow.dom._to_dom(node_29882));


var G__29895 = cljs.core.next(seq__28933_29874__$1);
var G__29896 = null;
var G__29897 = (0);
var G__29898 = (0);
seq__28933_29849 = G__29895;
chunk__28934_29851 = G__29896;
count__28935_29852 = G__29897;
i__28936_29853 = G__29898;
continue;
}
} else {
}
}
break;
}

return (new shadow.dom.NativeColl(fragment));
}));

(shadow.dom.fragment.cljs$lang$maxFixedArity = (0));

/** @this {Function} */
(shadow.dom.fragment.cljs$lang$applyTo = (function (seq28931){
var self__5712__auto__ = this;
return self__5712__auto__.cljs$core$IFn$_invoke$arity$variadic(cljs.core.seq(seq28931));
}));

/**
 * given a html string, eval all <script> tags and return the html without the scripts
 * don't do this for everything, only content you trust.
 */
shadow.dom.eval_scripts = (function shadow$dom$eval_scripts(s){
var scripts = cljs.core.re_seq(/<script[^>]*?>(.+?)<\/script>/,s);
var seq__28943_29899 = cljs.core.seq(scripts);
var chunk__28944_29900 = null;
var count__28945_29901 = (0);
var i__28946_29902 = (0);
while(true){
if((i__28946_29902 < count__28945_29901)){
var vec__28956_29904 = chunk__28944_29900.cljs$core$IIndexed$_nth$arity$2(null, i__28946_29902);
var script_tag_29905 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28956_29904,(0),null);
var script_body_29906 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28956_29904,(1),null);
eval(script_body_29906);


var G__29908 = seq__28943_29899;
var G__29909 = chunk__28944_29900;
var G__29910 = count__28945_29901;
var G__29911 = (i__28946_29902 + (1));
seq__28943_29899 = G__29908;
chunk__28944_29900 = G__29909;
count__28945_29901 = G__29910;
i__28946_29902 = G__29911;
continue;
} else {
var temp__5825__auto___29912 = cljs.core.seq(seq__28943_29899);
if(temp__5825__auto___29912){
var seq__28943_29913__$1 = temp__5825__auto___29912;
if(cljs.core.chunked_seq_QMARK_(seq__28943_29913__$1)){
var c__5525__auto___29914 = cljs.core.chunk_first(seq__28943_29913__$1);
var G__29915 = cljs.core.chunk_rest(seq__28943_29913__$1);
var G__29916 = c__5525__auto___29914;
var G__29917 = cljs.core.count(c__5525__auto___29914);
var G__29918 = (0);
seq__28943_29899 = G__29915;
chunk__28944_29900 = G__29916;
count__28945_29901 = G__29917;
i__28946_29902 = G__29918;
continue;
} else {
var vec__28961_29920 = cljs.core.first(seq__28943_29913__$1);
var script_tag_29921 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28961_29920,(0),null);
var script_body_29922 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28961_29920,(1),null);
eval(script_body_29922);


var G__29924 = cljs.core.next(seq__28943_29913__$1);
var G__29925 = null;
var G__29926 = (0);
var G__29927 = (0);
seq__28943_29899 = G__29924;
chunk__28944_29900 = G__29925;
count__28945_29901 = G__29926;
i__28946_29902 = G__29927;
continue;
}
} else {
}
}
break;
}

return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (s__$1,p__28964){
var vec__28967 = p__28964;
var script_tag = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28967,(0),null);
var script_body = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__28967,(1),null);
return clojure.string.replace(s__$1,script_tag,"");
}),s,scripts);
});
shadow.dom.str__GT_fragment = (function shadow$dom$str__GT_fragment(s){
var el = document.createElement("div");
(el.innerHTML = s);

return (new shadow.dom.NativeColl(goog.dom.childrenToNode_(document,el)));
});
shadow.dom.node_name = (function shadow$dom$node_name(el){
return shadow.dom.dom_node(el).nodeName;
});
shadow.dom.ancestor_by_class = (function shadow$dom$ancestor_by_class(el,cls){
return goog.dom.getAncestorByClass(shadow.dom.dom_node(el),cls);
});
shadow.dom.ancestor_by_tag = (function shadow$dom$ancestor_by_tag(var_args){
var G__28973 = arguments.length;
switch (G__28973) {
case 2:
return shadow.dom.ancestor_by_tag.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return shadow.dom.ancestor_by_tag.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.ancestor_by_tag.cljs$core$IFn$_invoke$arity$2 = (function (el,tag){
return goog.dom.getAncestorByTagNameAndClass(shadow.dom.dom_node(el),cljs.core.name(tag));
}));

(shadow.dom.ancestor_by_tag.cljs$core$IFn$_invoke$arity$3 = (function (el,tag,cls){
return goog.dom.getAncestorByTagNameAndClass(shadow.dom.dom_node(el),cljs.core.name(tag),cljs.core.name(cls));
}));

(shadow.dom.ancestor_by_tag.cljs$lang$maxFixedArity = 3);

shadow.dom.get_value = (function shadow$dom$get_value(dom){
return goog.dom.forms.getValue(shadow.dom.dom_node(dom));
});
shadow.dom.set_value = (function shadow$dom$set_value(dom,value){
return goog.dom.forms.setValue(shadow.dom.dom_node(dom),value);
});
shadow.dom.px = (function shadow$dom$px(value){
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1((value | (0))),"px"].join('');
});
shadow.dom.pct = (function shadow$dom$pct(value){
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(value),"%"].join('');
});
shadow.dom.remove_style_STAR_ = (function shadow$dom$remove_style_STAR_(el,style){
return el.style.removeProperty(cljs.core.name(style));
});
shadow.dom.remove_style = (function shadow$dom$remove_style(el,style){
var el__$1 = shadow.dom.dom_node(el);
return shadow.dom.remove_style_STAR_(el__$1,style);
});
shadow.dom.remove_styles = (function shadow$dom$remove_styles(el,style_keys){
var el__$1 = shadow.dom.dom_node(el);
var seq__28998 = cljs.core.seq(style_keys);
var chunk__28999 = null;
var count__29000 = (0);
var i__29001 = (0);
while(true){
if((i__29001 < count__29000)){
var it = chunk__28999.cljs$core$IIndexed$_nth$arity$2(null, i__29001);
shadow.dom.remove_style_STAR_(el__$1,it);


var G__29962 = seq__28998;
var G__29963 = chunk__28999;
var G__29964 = count__29000;
var G__29965 = (i__29001 + (1));
seq__28998 = G__29962;
chunk__28999 = G__29963;
count__29000 = G__29964;
i__29001 = G__29965;
continue;
} else {
var temp__5825__auto__ = cljs.core.seq(seq__28998);
if(temp__5825__auto__){
var seq__28998__$1 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(seq__28998__$1)){
var c__5525__auto__ = cljs.core.chunk_first(seq__28998__$1);
var G__29966 = cljs.core.chunk_rest(seq__28998__$1);
var G__29967 = c__5525__auto__;
var G__29968 = cljs.core.count(c__5525__auto__);
var G__29969 = (0);
seq__28998 = G__29966;
chunk__28999 = G__29967;
count__29000 = G__29968;
i__29001 = G__29969;
continue;
} else {
var it = cljs.core.first(seq__28998__$1);
shadow.dom.remove_style_STAR_(el__$1,it);


var G__29970 = cljs.core.next(seq__28998__$1);
var G__29971 = null;
var G__29972 = (0);
var G__29973 = (0);
seq__28998 = G__29970;
chunk__28999 = G__29971;
count__29000 = G__29972;
i__29001 = G__29973;
continue;
}
} else {
return null;
}
}
break;
}
});

/**
* @constructor
 * @implements {cljs.core.IRecord}
 * @implements {cljs.core.IKVReduce}
 * @implements {cljs.core.IEquiv}
 * @implements {cljs.core.IHash}
 * @implements {cljs.core.ICollection}
 * @implements {cljs.core.ICounted}
 * @implements {cljs.core.ISeqable}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.ICloneable}
 * @implements {cljs.core.IPrintWithWriter}
 * @implements {cljs.core.IIterable}
 * @implements {cljs.core.IWithMeta}
 * @implements {cljs.core.IAssociative}
 * @implements {cljs.core.IMap}
 * @implements {cljs.core.ILookup}
*/
shadow.dom.Coordinate = (function (x,y,__meta,__extmap,__hash){
this.x = x;
this.y = y;
this.__meta = __meta;
this.__extmap = __extmap;
this.__hash = __hash;
this.cljs$lang$protocol_mask$partition0$ = 2230716170;
this.cljs$lang$protocol_mask$partition1$ = 139264;
});
(shadow.dom.Coordinate.prototype.cljs$core$ILookup$_lookup$arity$2 = (function (this__5300__auto__,k__5301__auto__){
var self__ = this;
var this__5300__auto____$1 = this;
return this__5300__auto____$1.cljs$core$ILookup$_lookup$arity$3(null, k__5301__auto__,null);
}));

(shadow.dom.Coordinate.prototype.cljs$core$ILookup$_lookup$arity$3 = (function (this__5302__auto__,k29029,else__5303__auto__){
var self__ = this;
var this__5302__auto____$1 = this;
var G__29050 = k29029;
var G__29050__$1 = (((G__29050 instanceof cljs.core.Keyword))?G__29050.fqn:null);
switch (G__29050__$1) {
case "x":
return self__.x;

break;
case "y":
return self__.y;

break;
default:
return cljs.core.get.cljs$core$IFn$_invoke$arity$3(self__.__extmap,k29029,else__5303__auto__);

}
}));

(shadow.dom.Coordinate.prototype.cljs$core$IKVReduce$_kv_reduce$arity$3 = (function (this__5320__auto__,f__5321__auto__,init__5322__auto__){
var self__ = this;
var this__5320__auto____$1 = this;
return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (ret__5323__auto__,p__29057){
var vec__29058 = p__29057;
var k__5324__auto__ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29058,(0),null);
var v__5325__auto__ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29058,(1),null);
return (f__5321__auto__.cljs$core$IFn$_invoke$arity$3 ? f__5321__auto__.cljs$core$IFn$_invoke$arity$3(ret__5323__auto__,k__5324__auto__,v__5325__auto__) : f__5321__auto__.call(null, ret__5323__auto__,k__5324__auto__,v__5325__auto__));
}),init__5322__auto__,this__5320__auto____$1);
}));

(shadow.dom.Coordinate.prototype.cljs$core$IPrintWithWriter$_pr_writer$arity$3 = (function (this__5315__auto__,writer__5316__auto__,opts__5317__auto__){
var self__ = this;
var this__5315__auto____$1 = this;
var pr_pair__5318__auto__ = (function (keyval__5319__auto__){
return cljs.core.pr_sequential_writer(writer__5316__auto__,cljs.core.pr_writer,""," ","",opts__5317__auto__,keyval__5319__auto__);
});
return cljs.core.pr_sequential_writer(writer__5316__auto__,pr_pair__5318__auto__,"#shadow.dom.Coordinate{",", ","}",opts__5317__auto__,cljs.core.concat.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[new cljs.core.Keyword(null,"x","x",2099068185),self__.x],null)),(new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[new cljs.core.Keyword(null,"y","y",-1757859776),self__.y],null))], null),self__.__extmap));
}));

(shadow.dom.Coordinate.prototype.cljs$core$IIterable$_iterator$arity$1 = (function (G__29028){
var self__ = this;
var G__29028__$1 = this;
return (new cljs.core.RecordIter((0),G__29028__$1,2,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"x","x",2099068185),new cljs.core.Keyword(null,"y","y",-1757859776)], null),(cljs.core.truth_(self__.__extmap)?cljs.core._iterator(self__.__extmap):cljs.core.nil_iter())));
}));

(shadow.dom.Coordinate.prototype.cljs$core$IMeta$_meta$arity$1 = (function (this__5298__auto__){
var self__ = this;
var this__5298__auto____$1 = this;
return self__.__meta;
}));

(shadow.dom.Coordinate.prototype.cljs$core$ICloneable$_clone$arity$1 = (function (this__5295__auto__){
var self__ = this;
var this__5295__auto____$1 = this;
return (new shadow.dom.Coordinate(self__.x,self__.y,self__.__meta,self__.__extmap,self__.__hash));
}));

(shadow.dom.Coordinate.prototype.cljs$core$ICounted$_count$arity$1 = (function (this__5304__auto__){
var self__ = this;
var this__5304__auto____$1 = this;
return (2 + cljs.core.count(self__.__extmap));
}));

(shadow.dom.Coordinate.prototype.cljs$core$IHash$_hash$arity$1 = (function (this__5296__auto__){
var self__ = this;
var this__5296__auto____$1 = this;
var h__5111__auto__ = self__.__hash;
if((!((h__5111__auto__ == null)))){
return h__5111__auto__;
} else {
var h__5111__auto____$1 = (function (coll__5297__auto__){
return (145542109 ^ cljs.core.hash_unordered_coll(coll__5297__auto__));
})(this__5296__auto____$1);
(self__.__hash = h__5111__auto____$1);

return h__5111__auto____$1;
}
}));

(shadow.dom.Coordinate.prototype.cljs$core$IEquiv$_equiv$arity$2 = (function (this29030,other29031){
var self__ = this;
var this29030__$1 = this;
return (((!((other29031 == null)))) && ((((this29030__$1.constructor === other29031.constructor)) && (((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29030__$1.x,other29031.x)) && (((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29030__$1.y,other29031.y)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29030__$1.__extmap,other29031.__extmap)))))))));
}));

(shadow.dom.Coordinate.prototype.cljs$core$IMap$_dissoc$arity$2 = (function (this__5310__auto__,k__5311__auto__){
var self__ = this;
var this__5310__auto____$1 = this;
if(cljs.core.contains_QMARK_(new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"y","y",-1757859776),null,new cljs.core.Keyword(null,"x","x",2099068185),null], null), null),k__5311__auto__)){
return cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(cljs.core._with_meta(cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentArrayMap.EMPTY,this__5310__auto____$1),self__.__meta),k__5311__auto__);
} else {
return (new shadow.dom.Coordinate(self__.x,self__.y,self__.__meta,cljs.core.not_empty(cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(self__.__extmap,k__5311__auto__)),null));
}
}));

(shadow.dom.Coordinate.prototype.cljs$core$IAssociative$_contains_key_QMARK_$arity$2 = (function (this__5307__auto__,k29029){
var self__ = this;
var this__5307__auto____$1 = this;
var G__29082 = k29029;
var G__29082__$1 = (((G__29082 instanceof cljs.core.Keyword))?G__29082.fqn:null);
switch (G__29082__$1) {
case "x":
case "y":
return true;

break;
default:
return cljs.core.contains_QMARK_(self__.__extmap,k29029);

}
}));

(shadow.dom.Coordinate.prototype.cljs$core$IAssociative$_assoc$arity$3 = (function (this__5308__auto__,k__5309__auto__,G__29028){
var self__ = this;
var this__5308__auto____$1 = this;
var pred__29083 = cljs.core.keyword_identical_QMARK_;
var expr__29084 = k__5309__auto__;
if(cljs.core.truth_((pred__29083.cljs$core$IFn$_invoke$arity$2 ? pred__29083.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"x","x",2099068185),expr__29084) : pred__29083.call(null, new cljs.core.Keyword(null,"x","x",2099068185),expr__29084)))){
return (new shadow.dom.Coordinate(G__29028,self__.y,self__.__meta,self__.__extmap,null));
} else {
if(cljs.core.truth_((pred__29083.cljs$core$IFn$_invoke$arity$2 ? pred__29083.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"y","y",-1757859776),expr__29084) : pred__29083.call(null, new cljs.core.Keyword(null,"y","y",-1757859776),expr__29084)))){
return (new shadow.dom.Coordinate(self__.x,G__29028,self__.__meta,self__.__extmap,null));
} else {
return (new shadow.dom.Coordinate(self__.x,self__.y,self__.__meta,cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(self__.__extmap,k__5309__auto__,G__29028),null));
}
}
}));

(shadow.dom.Coordinate.prototype.cljs$core$ISeqable$_seq$arity$1 = (function (this__5313__auto__){
var self__ = this;
var this__5313__auto____$1 = this;
return cljs.core.seq(cljs.core.concat.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.MapEntry(new cljs.core.Keyword(null,"x","x",2099068185),self__.x,null)),(new cljs.core.MapEntry(new cljs.core.Keyword(null,"y","y",-1757859776),self__.y,null))], null),self__.__extmap));
}));

(shadow.dom.Coordinate.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (this__5299__auto__,G__29028){
var self__ = this;
var this__5299__auto____$1 = this;
return (new shadow.dom.Coordinate(self__.x,self__.y,G__29028,self__.__extmap,self__.__hash));
}));

(shadow.dom.Coordinate.prototype.cljs$core$ICollection$_conj$arity$2 = (function (this__5305__auto__,entry__5306__auto__){
var self__ = this;
var this__5305__auto____$1 = this;
if(cljs.core.vector_QMARK_(entry__5306__auto__)){
return this__5305__auto____$1.cljs$core$IAssociative$_assoc$arity$3(null, cljs.core._nth(entry__5306__auto__,(0)),cljs.core._nth(entry__5306__auto__,(1)));
} else {
return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3(cljs.core._conj,this__5305__auto____$1,entry__5306__auto__);
}
}));

(shadow.dom.Coordinate.getBasis = (function (){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"x","x",-555367584,null),new cljs.core.Symbol(null,"y","y",-117328249,null)], null);
}));

(shadow.dom.Coordinate.cljs$lang$type = true);

(shadow.dom.Coordinate.cljs$lang$ctorPrSeq = (function (this__5346__auto__){
return (new cljs.core.List(null,"shadow.dom/Coordinate",null,(1),null));
}));

(shadow.dom.Coordinate.cljs$lang$ctorPrWriter = (function (this__5346__auto__,writer__5347__auto__){
return cljs.core._write(writer__5347__auto__,"shadow.dom/Coordinate");
}));

/**
 * Positional factory function for shadow.dom/Coordinate.
 */
shadow.dom.__GT_Coordinate = (function shadow$dom$__GT_Coordinate(x,y){
return (new shadow.dom.Coordinate(x,y,null,null,null));
});

/**
 * Factory function for shadow.dom/Coordinate, taking a map of keywords to field values.
 */
shadow.dom.map__GT_Coordinate = (function shadow$dom$map__GT_Coordinate(G__29041){
var extmap__5342__auto__ = (function (){var G__29086 = cljs.core.dissoc.cljs$core$IFn$_invoke$arity$variadic(G__29041,new cljs.core.Keyword(null,"x","x",2099068185),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"y","y",-1757859776)], 0));
if(cljs.core.record_QMARK_(G__29041)){
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentArrayMap.EMPTY,G__29086);
} else {
return G__29086;
}
})();
return (new shadow.dom.Coordinate(new cljs.core.Keyword(null,"x","x",2099068185).cljs$core$IFn$_invoke$arity$1(G__29041),new cljs.core.Keyword(null,"y","y",-1757859776).cljs$core$IFn$_invoke$arity$1(G__29041),null,cljs.core.not_empty(extmap__5342__auto__),null));
});

shadow.dom.get_position = (function shadow$dom$get_position(el){
var pos = goog.style.getPosition(shadow.dom.dom_node(el));
return shadow.dom.__GT_Coordinate(pos.x,pos.y);
});
shadow.dom.get_client_position = (function shadow$dom$get_client_position(el){
var pos = goog.style.getClientPosition(shadow.dom.dom_node(el));
return shadow.dom.__GT_Coordinate(pos.x,pos.y);
});
shadow.dom.get_page_offset = (function shadow$dom$get_page_offset(el){
var pos = goog.style.getPageOffset(shadow.dom.dom_node(el));
return shadow.dom.__GT_Coordinate(pos.x,pos.y);
});

/**
* @constructor
 * @implements {cljs.core.IRecord}
 * @implements {cljs.core.IKVReduce}
 * @implements {cljs.core.IEquiv}
 * @implements {cljs.core.IHash}
 * @implements {cljs.core.ICollection}
 * @implements {cljs.core.ICounted}
 * @implements {cljs.core.ISeqable}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.ICloneable}
 * @implements {cljs.core.IPrintWithWriter}
 * @implements {cljs.core.IIterable}
 * @implements {cljs.core.IWithMeta}
 * @implements {cljs.core.IAssociative}
 * @implements {cljs.core.IMap}
 * @implements {cljs.core.ILookup}
*/
shadow.dom.Size = (function (w,h,__meta,__extmap,__hash){
this.w = w;
this.h = h;
this.__meta = __meta;
this.__extmap = __extmap;
this.__hash = __hash;
this.cljs$lang$protocol_mask$partition0$ = 2230716170;
this.cljs$lang$protocol_mask$partition1$ = 139264;
});
(shadow.dom.Size.prototype.cljs$core$ILookup$_lookup$arity$2 = (function (this__5300__auto__,k__5301__auto__){
var self__ = this;
var this__5300__auto____$1 = this;
return this__5300__auto____$1.cljs$core$ILookup$_lookup$arity$3(null, k__5301__auto__,null);
}));

(shadow.dom.Size.prototype.cljs$core$ILookup$_lookup$arity$3 = (function (this__5302__auto__,k29088,else__5303__auto__){
var self__ = this;
var this__5302__auto____$1 = this;
var G__29096 = k29088;
var G__29096__$1 = (((G__29096 instanceof cljs.core.Keyword))?G__29096.fqn:null);
switch (G__29096__$1) {
case "w":
return self__.w;

break;
case "h":
return self__.h;

break;
default:
return cljs.core.get.cljs$core$IFn$_invoke$arity$3(self__.__extmap,k29088,else__5303__auto__);

}
}));

(shadow.dom.Size.prototype.cljs$core$IKVReduce$_kv_reduce$arity$3 = (function (this__5320__auto__,f__5321__auto__,init__5322__auto__){
var self__ = this;
var this__5320__auto____$1 = this;
return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (ret__5323__auto__,p__29099){
var vec__29100 = p__29099;
var k__5324__auto__ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29100,(0),null);
var v__5325__auto__ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29100,(1),null);
return (f__5321__auto__.cljs$core$IFn$_invoke$arity$3 ? f__5321__auto__.cljs$core$IFn$_invoke$arity$3(ret__5323__auto__,k__5324__auto__,v__5325__auto__) : f__5321__auto__.call(null, ret__5323__auto__,k__5324__auto__,v__5325__auto__));
}),init__5322__auto__,this__5320__auto____$1);
}));

(shadow.dom.Size.prototype.cljs$core$IPrintWithWriter$_pr_writer$arity$3 = (function (this__5315__auto__,writer__5316__auto__,opts__5317__auto__){
var self__ = this;
var this__5315__auto____$1 = this;
var pr_pair__5318__auto__ = (function (keyval__5319__auto__){
return cljs.core.pr_sequential_writer(writer__5316__auto__,cljs.core.pr_writer,""," ","",opts__5317__auto__,keyval__5319__auto__);
});
return cljs.core.pr_sequential_writer(writer__5316__auto__,pr_pair__5318__auto__,"#shadow.dom.Size{",", ","}",opts__5317__auto__,cljs.core.concat.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[new cljs.core.Keyword(null,"w","w",354169001),self__.w],null)),(new cljs.core.PersistentVector(null,2,(5),cljs.core.PersistentVector.EMPTY_NODE,[new cljs.core.Keyword(null,"h","h",1109658740),self__.h],null))], null),self__.__extmap));
}));

(shadow.dom.Size.prototype.cljs$core$IIterable$_iterator$arity$1 = (function (G__29087){
var self__ = this;
var G__29087__$1 = this;
return (new cljs.core.RecordIter((0),G__29087__$1,2,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"w","w",354169001),new cljs.core.Keyword(null,"h","h",1109658740)], null),(cljs.core.truth_(self__.__extmap)?cljs.core._iterator(self__.__extmap):cljs.core.nil_iter())));
}));

(shadow.dom.Size.prototype.cljs$core$IMeta$_meta$arity$1 = (function (this__5298__auto__){
var self__ = this;
var this__5298__auto____$1 = this;
return self__.__meta;
}));

(shadow.dom.Size.prototype.cljs$core$ICloneable$_clone$arity$1 = (function (this__5295__auto__){
var self__ = this;
var this__5295__auto____$1 = this;
return (new shadow.dom.Size(self__.w,self__.h,self__.__meta,self__.__extmap,self__.__hash));
}));

(shadow.dom.Size.prototype.cljs$core$ICounted$_count$arity$1 = (function (this__5304__auto__){
var self__ = this;
var this__5304__auto____$1 = this;
return (2 + cljs.core.count(self__.__extmap));
}));

(shadow.dom.Size.prototype.cljs$core$IHash$_hash$arity$1 = (function (this__5296__auto__){
var self__ = this;
var this__5296__auto____$1 = this;
var h__5111__auto__ = self__.__hash;
if((!((h__5111__auto__ == null)))){
return h__5111__auto__;
} else {
var h__5111__auto____$1 = (function (coll__5297__auto__){
return (-1228019642 ^ cljs.core.hash_unordered_coll(coll__5297__auto__));
})(this__5296__auto____$1);
(self__.__hash = h__5111__auto____$1);

return h__5111__auto____$1;
}
}));

(shadow.dom.Size.prototype.cljs$core$IEquiv$_equiv$arity$2 = (function (this29089,other29090){
var self__ = this;
var this29089__$1 = this;
return (((!((other29090 == null)))) && ((((this29089__$1.constructor === other29090.constructor)) && (((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29089__$1.w,other29090.w)) && (((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29089__$1.h,other29090.h)) && (cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(this29089__$1.__extmap,other29090.__extmap)))))))));
}));

(shadow.dom.Size.prototype.cljs$core$IMap$_dissoc$arity$2 = (function (this__5310__auto__,k__5311__auto__){
var self__ = this;
var this__5310__auto____$1 = this;
if(cljs.core.contains_QMARK_(new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"w","w",354169001),null,new cljs.core.Keyword(null,"h","h",1109658740),null], null), null),k__5311__auto__)){
return cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(cljs.core._with_meta(cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentArrayMap.EMPTY,this__5310__auto____$1),self__.__meta),k__5311__auto__);
} else {
return (new shadow.dom.Size(self__.w,self__.h,self__.__meta,cljs.core.not_empty(cljs.core.dissoc.cljs$core$IFn$_invoke$arity$2(self__.__extmap,k__5311__auto__)),null));
}
}));

(shadow.dom.Size.prototype.cljs$core$IAssociative$_contains_key_QMARK_$arity$2 = (function (this__5307__auto__,k29088){
var self__ = this;
var this__5307__auto____$1 = this;
var G__29106 = k29088;
var G__29106__$1 = (((G__29106 instanceof cljs.core.Keyword))?G__29106.fqn:null);
switch (G__29106__$1) {
case "w":
case "h":
return true;

break;
default:
return cljs.core.contains_QMARK_(self__.__extmap,k29088);

}
}));

(shadow.dom.Size.prototype.cljs$core$IAssociative$_assoc$arity$3 = (function (this__5308__auto__,k__5309__auto__,G__29087){
var self__ = this;
var this__5308__auto____$1 = this;
var pred__29108 = cljs.core.keyword_identical_QMARK_;
var expr__29109 = k__5309__auto__;
if(cljs.core.truth_((pred__29108.cljs$core$IFn$_invoke$arity$2 ? pred__29108.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"w","w",354169001),expr__29109) : pred__29108.call(null, new cljs.core.Keyword(null,"w","w",354169001),expr__29109)))){
return (new shadow.dom.Size(G__29087,self__.h,self__.__meta,self__.__extmap,null));
} else {
if(cljs.core.truth_((pred__29108.cljs$core$IFn$_invoke$arity$2 ? pred__29108.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"h","h",1109658740),expr__29109) : pred__29108.call(null, new cljs.core.Keyword(null,"h","h",1109658740),expr__29109)))){
return (new shadow.dom.Size(self__.w,G__29087,self__.__meta,self__.__extmap,null));
} else {
return (new shadow.dom.Size(self__.w,self__.h,self__.__meta,cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(self__.__extmap,k__5309__auto__,G__29087),null));
}
}
}));

(shadow.dom.Size.prototype.cljs$core$ISeqable$_seq$arity$1 = (function (this__5313__auto__){
var self__ = this;
var this__5313__auto____$1 = this;
return cljs.core.seq(cljs.core.concat.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.MapEntry(new cljs.core.Keyword(null,"w","w",354169001),self__.w,null)),(new cljs.core.MapEntry(new cljs.core.Keyword(null,"h","h",1109658740),self__.h,null))], null),self__.__extmap));
}));

(shadow.dom.Size.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (this__5299__auto__,G__29087){
var self__ = this;
var this__5299__auto____$1 = this;
return (new shadow.dom.Size(self__.w,self__.h,G__29087,self__.__extmap,self__.__hash));
}));

(shadow.dom.Size.prototype.cljs$core$ICollection$_conj$arity$2 = (function (this__5305__auto__,entry__5306__auto__){
var self__ = this;
var this__5305__auto____$1 = this;
if(cljs.core.vector_QMARK_(entry__5306__auto__)){
return this__5305__auto____$1.cljs$core$IAssociative$_assoc$arity$3(null, cljs.core._nth(entry__5306__auto__,(0)),cljs.core._nth(entry__5306__auto__,(1)));
} else {
return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3(cljs.core._conj,this__5305__auto____$1,entry__5306__auto__);
}
}));

(shadow.dom.Size.getBasis = (function (){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"w","w",1994700528,null),new cljs.core.Symbol(null,"h","h",-1544777029,null)], null);
}));

(shadow.dom.Size.cljs$lang$type = true);

(shadow.dom.Size.cljs$lang$ctorPrSeq = (function (this__5346__auto__){
return (new cljs.core.List(null,"shadow.dom/Size",null,(1),null));
}));

(shadow.dom.Size.cljs$lang$ctorPrWriter = (function (this__5346__auto__,writer__5347__auto__){
return cljs.core._write(writer__5347__auto__,"shadow.dom/Size");
}));

/**
 * Positional factory function for shadow.dom/Size.
 */
shadow.dom.__GT_Size = (function shadow$dom$__GT_Size(w,h){
return (new shadow.dom.Size(w,h,null,null,null));
});

/**
 * Factory function for shadow.dom/Size, taking a map of keywords to field values.
 */
shadow.dom.map__GT_Size = (function shadow$dom$map__GT_Size(G__29092){
var extmap__5342__auto__ = (function (){var G__29120 = cljs.core.dissoc.cljs$core$IFn$_invoke$arity$variadic(G__29092,new cljs.core.Keyword(null,"w","w",354169001),cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([new cljs.core.Keyword(null,"h","h",1109658740)], 0));
if(cljs.core.record_QMARK_(G__29092)){
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentArrayMap.EMPTY,G__29120);
} else {
return G__29120;
}
})();
return (new shadow.dom.Size(new cljs.core.Keyword(null,"w","w",354169001).cljs$core$IFn$_invoke$arity$1(G__29092),new cljs.core.Keyword(null,"h","h",1109658740).cljs$core$IFn$_invoke$arity$1(G__29092),null,cljs.core.not_empty(extmap__5342__auto__),null));
});

shadow.dom.size__GT_clj = (function shadow$dom$size__GT_clj(size){
return (new shadow.dom.Size(size.width,size.height,null,null,null));
});
shadow.dom.get_size = (function shadow$dom$get_size(el){
return shadow.dom.size__GT_clj(goog.style.getSize(shadow.dom.dom_node(el)));
});
shadow.dom.get_height = (function shadow$dom$get_height(el){
return shadow.dom.get_size(el).h;
});
shadow.dom.get_viewport_size = (function shadow$dom$get_viewport_size(){
return shadow.dom.size__GT_clj(goog.dom.getViewportSize());
});
shadow.dom.first_child = (function shadow$dom$first_child(el){
return (shadow.dom.dom_node(el).children[(0)]);
});
shadow.dom.select_option_values = (function shadow$dom$select_option_values(el){
var native$ = shadow.dom.dom_node(el);
var opts = (native$["options"]);
var a__5590__auto__ = opts;
var l__5591__auto__ = a__5590__auto__.length;
var i = (0);
var ret = cljs.core.PersistentVector.EMPTY;
while(true){
if((i < l__5591__auto__)){
var G__30266 = (i + (1));
var G__30267 = cljs.core.conj.cljs$core$IFn$_invoke$arity$2(ret,(opts[i]["value"]));
i = G__30266;
ret = G__30267;
continue;
} else {
return ret;
}
break;
}
});
shadow.dom.build_url = (function shadow$dom$build_url(path,query_params){
if(cljs.core.empty_QMARK_(query_params)){
return path;
} else {
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(path),"?",clojure.string.join.cljs$core$IFn$_invoke$arity$2("&",cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p__29148){
var vec__29149 = p__29148;
var k = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29149,(0),null);
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29149,(1),null);
return [cljs.core.name(k),"=",cljs.core.str.cljs$core$IFn$_invoke$arity$1(encodeURIComponent(cljs.core.str.cljs$core$IFn$_invoke$arity$1(v)))].join('');
}),query_params))].join('');
}
});
shadow.dom.redirect = (function shadow$dom$redirect(var_args){
var G__29158 = arguments.length;
switch (G__29158) {
case 1:
return shadow.dom.redirect.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return shadow.dom.redirect.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(shadow.dom.redirect.cljs$core$IFn$_invoke$arity$1 = (function (path){
return shadow.dom.redirect.cljs$core$IFn$_invoke$arity$2(path,cljs.core.PersistentArrayMap.EMPTY);
}));

(shadow.dom.redirect.cljs$core$IFn$_invoke$arity$2 = (function (path,query_params){
return (document["location"]["href"] = shadow.dom.build_url(path,query_params));
}));

(shadow.dom.redirect.cljs$lang$maxFixedArity = 2);

shadow.dom.reload_BANG_ = (function shadow$dom$reload_BANG_(){
return (document.location.href = document.location.href);
});
shadow.dom.tag_name = (function shadow$dom$tag_name(el){
var dom = shadow.dom.dom_node(el);
return dom.tagName;
});
shadow.dom.insert_after = (function shadow$dom$insert_after(ref,new$){
var new_node = shadow.dom.dom_node(new$);
goog.dom.insertSiblingAfter(new_node,shadow.dom.dom_node(ref));

return new_node;
});
shadow.dom.insert_before = (function shadow$dom$insert_before(ref,new$){
var new_node = shadow.dom.dom_node(new$);
goog.dom.insertSiblingBefore(new_node,shadow.dom.dom_node(ref));

return new_node;
});
shadow.dom.insert_first = (function shadow$dom$insert_first(ref,new$){
var temp__5823__auto__ = shadow.dom.dom_node(ref).firstChild;
if(cljs.core.truth_(temp__5823__auto__)){
var child = temp__5823__auto__;
return shadow.dom.insert_before(child,new$);
} else {
return shadow.dom.append.cljs$core$IFn$_invoke$arity$2(ref,new$);
}
});
shadow.dom.index_of = (function shadow$dom$index_of(el){
var el__$1 = shadow.dom.dom_node(el);
var i = (0);
while(true){
var ps = el__$1.previousSibling;
if((ps == null)){
return i;
} else {
var G__30334 = ps;
var G__30335 = (i + (1));
el__$1 = G__30334;
i = G__30335;
continue;
}
break;
}
});
shadow.dom.get_parent = (function shadow$dom$get_parent(el){
return goog.dom.getParentElement(shadow.dom.dom_node(el));
});
shadow.dom.parents = (function shadow$dom$parents(el){
var parent = shadow.dom.get_parent(el);
if(cljs.core.truth_(parent)){
return cljs.core.cons(parent,(new cljs.core.LazySeq(null,(function (){
return (shadow.dom.parents.cljs$core$IFn$_invoke$arity$1 ? shadow.dom.parents.cljs$core$IFn$_invoke$arity$1(parent) : shadow.dom.parents.call(null, parent));
}),null,null)));
} else {
return null;
}
});
shadow.dom.matches = (function shadow$dom$matches(el,sel){
return shadow.dom.dom_node(el).matches(sel);
});
shadow.dom.get_next_sibling = (function shadow$dom$get_next_sibling(el){
return goog.dom.getNextElementSibling(shadow.dom.dom_node(el));
});
shadow.dom.get_previous_sibling = (function shadow$dom$get_previous_sibling(el){
return goog.dom.getPreviousElementSibling(shadow.dom.dom_node(el));
});
shadow.dom.xmlns = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(new cljs.core.PersistentArrayMap(null, 2, ["svg","http://www.w3.org/2000/svg","xlink","http://www.w3.org/1999/xlink"], null));
shadow.dom.create_svg_node = (function shadow$dom$create_svg_node(tag_def,props){
var vec__29180 = shadow.dom.parse_tag(tag_def);
var tag_name = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29180,(0),null);
var tag_id = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29180,(1),null);
var tag_classes = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29180,(2),null);
var el = document.createElementNS("http://www.w3.org/2000/svg",tag_name);
if(cljs.core.truth_(tag_id)){
el.setAttribute("id",tag_id);
} else {
}

if(cljs.core.truth_(tag_classes)){
el.setAttribute("class",shadow.dom.merge_class_string(new cljs.core.Keyword(null,"class","class",-2030961996).cljs$core$IFn$_invoke$arity$1(props),tag_classes));
} else {
}

var seq__29184_30374 = cljs.core.seq(props);
var chunk__29185_30375 = null;
var count__29186_30376 = (0);
var i__29187_30377 = (0);
while(true){
if((i__29187_30377 < count__29186_30376)){
var vec__29195_30379 = chunk__29185_30375.cljs$core$IIndexed$_nth$arity$2(null, i__29187_30377);
var k_30380 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29195_30379,(0),null);
var v_30381 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29195_30379,(1),null);
el.setAttributeNS((function (){var temp__5825__auto__ = cljs.core.namespace(k_30380);
if(cljs.core.truth_(temp__5825__auto__)){
var ns = temp__5825__auto__;
return cljs.core.get.cljs$core$IFn$_invoke$arity$2(cljs.core.deref(shadow.dom.xmlns),ns);
} else {
return null;
}
})(),cljs.core.name(k_30380),v_30381);


var G__30389 = seq__29184_30374;
var G__30390 = chunk__29185_30375;
var G__30391 = count__29186_30376;
var G__30392 = (i__29187_30377 + (1));
seq__29184_30374 = G__30389;
chunk__29185_30375 = G__30390;
count__29186_30376 = G__30391;
i__29187_30377 = G__30392;
continue;
} else {
var temp__5825__auto___30394 = cljs.core.seq(seq__29184_30374);
if(temp__5825__auto___30394){
var seq__29184_30396__$1 = temp__5825__auto___30394;
if(cljs.core.chunked_seq_QMARK_(seq__29184_30396__$1)){
var c__5525__auto___30397 = cljs.core.chunk_first(seq__29184_30396__$1);
var G__30398 = cljs.core.chunk_rest(seq__29184_30396__$1);
var G__30399 = c__5525__auto___30397;
var G__30400 = cljs.core.count(c__5525__auto___30397);
var G__30401 = (0);
seq__29184_30374 = G__30398;
chunk__29185_30375 = G__30399;
count__29186_30376 = G__30400;
i__29187_30377 = G__30401;
continue;
} else {
var vec__29198_30403 = cljs.core.first(seq__29184_30396__$1);
var k_30404 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29198_30403,(0),null);
var v_30405 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29198_30403,(1),null);
el.setAttributeNS((function (){var temp__5825__auto____$1 = cljs.core.namespace(k_30404);
if(cljs.core.truth_(temp__5825__auto____$1)){
var ns = temp__5825__auto____$1;
return cljs.core.get.cljs$core$IFn$_invoke$arity$2(cljs.core.deref(shadow.dom.xmlns),ns);
} else {
return null;
}
})(),cljs.core.name(k_30404),v_30405);


var G__30406 = cljs.core.next(seq__29184_30396__$1);
var G__30407 = null;
var G__30408 = (0);
var G__30409 = (0);
seq__29184_30374 = G__30406;
chunk__29185_30375 = G__30407;
count__29186_30376 = G__30408;
i__29187_30377 = G__30409;
continue;
}
} else {
}
}
break;
}

return el;
});
shadow.dom.svg_node = (function shadow$dom$svg_node(el){
if((el == null)){
return null;
} else {
if((((!((el == null))))?((((false) || ((cljs.core.PROTOCOL_SENTINEL === el.shadow$dom$SVGElement$))))?true:false):false)){
return el.shadow$dom$SVGElement$_to_svg$arity$1(null, );
} else {
return el;

}
}
});
shadow.dom.make_svg_node = (function shadow$dom$make_svg_node(structure){
var vec__29202 = shadow.dom.destructure_node(shadow.dom.create_svg_node,structure);
var node = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29202,(0),null);
var node_children = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__29202,(1),null);
var seq__29205_30413 = cljs.core.seq(node_children);
var chunk__29207_30414 = null;
var count__29208_30415 = (0);
var i__29209_30416 = (0);
while(true){
if((i__29209_30416 < count__29208_30415)){
var child_struct_30417 = chunk__29207_30414.cljs$core$IIndexed$_nth$arity$2(null, i__29209_30416);
if((!((child_struct_30417 == null)))){
if(typeof child_struct_30417 === 'string'){
var text_30418 = (node["textContent"]);
(node["textContent"] = [cljs.core.str.cljs$core$IFn$_invoke$arity$1(text_30418),child_struct_30417].join(''));
} else {
var children_30419 = shadow.dom.svg_node(child_struct_30417);
if(cljs.core.seq_QMARK_(children_30419)){
var seq__29255_30420 = cljs.core.seq(children_30419);
var chunk__29257_30421 = null;
var count__29258_30422 = (0);
var i__29259_30423 = (0);
while(true){
if((i__29259_30423 < count__29258_30422)){
var child_30424 = chunk__29257_30421.cljs$core$IIndexed$_nth$arity$2(null, i__29259_30423);
if(cljs.core.truth_(child_30424)){
node.appendChild(child_30424);


var G__30427 = seq__29255_30420;
var G__30428 = chunk__29257_30421;
var G__30429 = count__29258_30422;
var G__30430 = (i__29259_30423 + (1));
seq__29255_30420 = G__30427;
chunk__29257_30421 = G__30428;
count__29258_30422 = G__30429;
i__29259_30423 = G__30430;
continue;
} else {
var G__30432 = seq__29255_30420;
var G__30433 = chunk__29257_30421;
var G__30434 = count__29258_30422;
var G__30435 = (i__29259_30423 + (1));
seq__29255_30420 = G__30432;
chunk__29257_30421 = G__30433;
count__29258_30422 = G__30434;
i__29259_30423 = G__30435;
continue;
}
} else {
var temp__5825__auto___30436 = cljs.core.seq(seq__29255_30420);
if(temp__5825__auto___30436){
var seq__29255_30438__$1 = temp__5825__auto___30436;
if(cljs.core.chunked_seq_QMARK_(seq__29255_30438__$1)){
var c__5525__auto___30439 = cljs.core.chunk_first(seq__29255_30438__$1);
var G__30440 = cljs.core.chunk_rest(seq__29255_30438__$1);
var G__30441 = c__5525__auto___30439;
var G__30442 = cljs.core.count(c__5525__auto___30439);
var G__30443 = (0);
seq__29255_30420 = G__30440;
chunk__29257_30421 = G__30441;
count__29258_30422 = G__30442;
i__29259_30423 = G__30443;
continue;
} else {
var child_30465 = cljs.core.first(seq__29255_30438__$1);
if(cljs.core.truth_(child_30465)){
node.appendChild(child_30465);


var G__30478 = cljs.core.next(seq__29255_30438__$1);
var G__30479 = null;
var G__30480 = (0);
var G__30481 = (0);
seq__29255_30420 = G__30478;
chunk__29257_30421 = G__30479;
count__29258_30422 = G__30480;
i__29259_30423 = G__30481;
continue;
} else {
var G__30482 = cljs.core.next(seq__29255_30438__$1);
var G__30483 = null;
var G__30484 = (0);
var G__30485 = (0);
seq__29255_30420 = G__30482;
chunk__29257_30421 = G__30483;
count__29258_30422 = G__30484;
i__29259_30423 = G__30485;
continue;
}
}
} else {
}
}
break;
}
} else {
node.appendChild(children_30419);
}
}


var G__30490 = seq__29205_30413;
var G__30491 = chunk__29207_30414;
var G__30492 = count__29208_30415;
var G__30493 = (i__29209_30416 + (1));
seq__29205_30413 = G__30490;
chunk__29207_30414 = G__30491;
count__29208_30415 = G__30492;
i__29209_30416 = G__30493;
continue;
} else {
var G__30496 = seq__29205_30413;
var G__30497 = chunk__29207_30414;
var G__30498 = count__29208_30415;
var G__30499 = (i__29209_30416 + (1));
seq__29205_30413 = G__30496;
chunk__29207_30414 = G__30497;
count__29208_30415 = G__30498;
i__29209_30416 = G__30499;
continue;
}
} else {
var temp__5825__auto___30501 = cljs.core.seq(seq__29205_30413);
if(temp__5825__auto___30501){
var seq__29205_30507__$1 = temp__5825__auto___30501;
if(cljs.core.chunked_seq_QMARK_(seq__29205_30507__$1)){
var c__5525__auto___30516 = cljs.core.chunk_first(seq__29205_30507__$1);
var G__30522 = cljs.core.chunk_rest(seq__29205_30507__$1);
var G__30523 = c__5525__auto___30516;
var G__30524 = cljs.core.count(c__5525__auto___30516);
var G__30525 = (0);
seq__29205_30413 = G__30522;
chunk__29207_30414 = G__30523;
count__29208_30415 = G__30524;
i__29209_30416 = G__30525;
continue;
} else {
var child_struct_30542 = cljs.core.first(seq__29205_30507__$1);
if((!((child_struct_30542 == null)))){
if(typeof child_struct_30542 === 'string'){
var text_30554 = (node["textContent"]);
(node["textContent"] = [cljs.core.str.cljs$core$IFn$_invoke$arity$1(text_30554),child_struct_30542].join(''));
} else {
var children_30555 = shadow.dom.svg_node(child_struct_30542);
if(cljs.core.seq_QMARK_(children_30555)){
var seq__29270_30556 = cljs.core.seq(children_30555);
var chunk__29272_30557 = null;
var count__29273_30558 = (0);
var i__29274_30559 = (0);
while(true){
if((i__29274_30559 < count__29273_30558)){
var child_30564 = chunk__29272_30557.cljs$core$IIndexed$_nth$arity$2(null, i__29274_30559);
if(cljs.core.truth_(child_30564)){
node.appendChild(child_30564);


var G__30567 = seq__29270_30556;
var G__30568 = chunk__29272_30557;
var G__30569 = count__29273_30558;
var G__30570 = (i__29274_30559 + (1));
seq__29270_30556 = G__30567;
chunk__29272_30557 = G__30568;
count__29273_30558 = G__30569;
i__29274_30559 = G__30570;
continue;
} else {
var G__30571 = seq__29270_30556;
var G__30572 = chunk__29272_30557;
var G__30573 = count__29273_30558;
var G__30574 = (i__29274_30559 + (1));
seq__29270_30556 = G__30571;
chunk__29272_30557 = G__30572;
count__29273_30558 = G__30573;
i__29274_30559 = G__30574;
continue;
}
} else {
var temp__5825__auto___30575__$1 = cljs.core.seq(seq__29270_30556);
if(temp__5825__auto___30575__$1){
var seq__29270_30576__$1 = temp__5825__auto___30575__$1;
if(cljs.core.chunked_seq_QMARK_(seq__29270_30576__$1)){
var c__5525__auto___30577 = cljs.core.chunk_first(seq__29270_30576__$1);
var G__30578 = cljs.core.chunk_rest(seq__29270_30576__$1);
var G__30579 = c__5525__auto___30577;
var G__30580 = cljs.core.count(c__5525__auto___30577);
var G__30581 = (0);
seq__29270_30556 = G__30578;
chunk__29272_30557 = G__30579;
count__29273_30558 = G__30580;
i__29274_30559 = G__30581;
continue;
} else {
var child_30586 = cljs.core.first(seq__29270_30576__$1);
if(cljs.core.truth_(child_30586)){
node.appendChild(child_30586);


var G__30587 = cljs.core.next(seq__29270_30576__$1);
var G__30588 = null;
var G__30589 = (0);
var G__30590 = (0);
seq__29270_30556 = G__30587;
chunk__29272_30557 = G__30588;
count__29273_30558 = G__30589;
i__29274_30559 = G__30590;
continue;
} else {
var G__30591 = cljs.core.next(seq__29270_30576__$1);
var G__30592 = null;
var G__30593 = (0);
var G__30594 = (0);
seq__29270_30556 = G__30591;
chunk__29272_30557 = G__30592;
count__29273_30558 = G__30593;
i__29274_30559 = G__30594;
continue;
}
}
} else {
}
}
break;
}
} else {
node.appendChild(children_30555);
}
}


var G__30595 = cljs.core.next(seq__29205_30507__$1);
var G__30596 = null;
var G__30597 = (0);
var G__30598 = (0);
seq__29205_30413 = G__30595;
chunk__29207_30414 = G__30596;
count__29208_30415 = G__30597;
i__29209_30416 = G__30598;
continue;
} else {
var G__30599 = cljs.core.next(seq__29205_30507__$1);
var G__30600 = null;
var G__30601 = (0);
var G__30602 = (0);
seq__29205_30413 = G__30599;
chunk__29207_30414 = G__30600;
count__29208_30415 = G__30601;
i__29209_30416 = G__30602;
continue;
}
}
} else {
}
}
break;
}

return node;
});
(shadow.dom.SVGElement["string"] = true);

(shadow.dom._to_svg["string"] = (function (this$){
if((this$ instanceof cljs.core.Keyword)){
return shadow.dom.make_svg_node(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [this$], null));
} else {
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("strings cannot be in svgs",new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"this","this",-611633625),this$], null));
}
}));

(cljs.core.PersistentVector.prototype.shadow$dom$SVGElement$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.PersistentVector.prototype.shadow$dom$SVGElement$_to_svg$arity$1 = (function (this$){
var this$__$1 = this;
return shadow.dom.make_svg_node(this$__$1);
}));

(cljs.core.LazySeq.prototype.shadow$dom$SVGElement$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.LazySeq.prototype.shadow$dom$SVGElement$_to_svg$arity$1 = (function (this$){
var this$__$1 = this;
return cljs.core.map.cljs$core$IFn$_invoke$arity$2(shadow.dom._to_svg,this$__$1);
}));

(shadow.dom.SVGElement["null"] = true);

(shadow.dom._to_svg["null"] = (function (_){
return null;
}));
shadow.dom.svg = (function shadow$dom$svg(var_args){
var args__5732__auto__ = [];
var len__5726__auto___30614 = arguments.length;
var i__5727__auto___30615 = (0);
while(true){
if((i__5727__auto___30615 < len__5726__auto___30614)){
args__5732__auto__.push((arguments[i__5727__auto___30615]));

var G__30616 = (i__5727__auto___30615 + (1));
i__5727__auto___30615 = G__30616;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return shadow.dom.svg.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(shadow.dom.svg.cljs$core$IFn$_invoke$arity$variadic = (function (attrs,children){
return shadow.dom._to_svg(cljs.core.vec(cljs.core.concat.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"svg","svg",856789142),attrs], null),children)));
}));

(shadow.dom.svg.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(shadow.dom.svg.cljs$lang$applyTo = (function (seq29300){
var G__29301 = cljs.core.first(seq29300);
var seq29300__$1 = cljs.core.next(seq29300);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__29301,seq29300__$1);
}));


//# sourceMappingURL=shadow.dom.js.map
