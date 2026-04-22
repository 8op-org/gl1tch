goog.provide('cljs.core.async');
goog.scope(function(){
  cljs.core.async.goog$module$goog$array = goog.module.get('goog.array');
});

/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Handler}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async31046 = (function (f,blockable,meta31047){
this.f = f;
this.blockable = blockable;
this.meta31047 = meta31047;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_31048,meta31047__$1){
var self__ = this;
var _31048__$1 = this;
return (new cljs.core.async.t_cljs$core$async31046(self__.f,self__.blockable,meta31047__$1));
}));

(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_31048){
var self__ = this;
var _31048__$1 = this;
return self__.meta31047;
}));

(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$async$impl$protocols$Handler$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$async$impl$protocols$Handler$active_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return true;
}));

(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$async$impl$protocols$Handler$blockable_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return self__.blockable;
}));

(cljs.core.async.t_cljs$core$async31046.prototype.cljs$core$async$impl$protocols$Handler$commit$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return self__.f;
}));

(cljs.core.async.t_cljs$core$async31046.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"f","f",43394975,null),new cljs.core.Symbol(null,"blockable","blockable",-28395259,null),new cljs.core.Symbol(null,"meta31047","meta31047",-1866129131,null)], null);
}));

(cljs.core.async.t_cljs$core$async31046.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async31046.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async31046");

(cljs.core.async.t_cljs$core$async31046.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async31046");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async31046.
 */
cljs.core.async.__GT_t_cljs$core$async31046 = (function cljs$core$async$__GT_t_cljs$core$async31046(f,blockable,meta31047){
return (new cljs.core.async.t_cljs$core$async31046(f,blockable,meta31047));
});


cljs.core.async.fn_handler = (function cljs$core$async$fn_handler(var_args){
var G__31044 = arguments.length;
switch (G__31044) {
case 1:
return cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$1 = (function (f){
return cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$2(f,true);
}));

(cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$2 = (function (f,blockable){
return (new cljs.core.async.t_cljs$core$async31046(f,blockable,cljs.core.PersistentArrayMap.EMPTY));
}));

(cljs.core.async.fn_handler.cljs$lang$maxFixedArity = 2);

/**
 * Returns a fixed buffer of size n. When full, puts will block/park.
 */
cljs.core.async.buffer = (function cljs$core$async$buffer(n){
return cljs.core.async.impl.buffers.fixed_buffer(n);
});
/**
 * Returns a buffer of size n. When full, puts will complete but
 *   val will be dropped (no transfer).
 */
cljs.core.async.dropping_buffer = (function cljs$core$async$dropping_buffer(n){
return cljs.core.async.impl.buffers.dropping_buffer(n);
});
/**
 * Returns a buffer of size n. When full, puts will complete, and be
 *   buffered, but oldest elements in buffer will be dropped (not
 *   transferred).
 */
cljs.core.async.sliding_buffer = (function cljs$core$async$sliding_buffer(n){
return cljs.core.async.impl.buffers.sliding_buffer(n);
});
/**
 * Returns true if a channel created with buff will never block. That is to say,
 * puts into this buffer will never cause the buffer to be full. 
 */
cljs.core.async.unblocking_buffer_QMARK_ = (function cljs$core$async$unblocking_buffer_QMARK_(buff){
if((!((buff == null)))){
if(((false) || ((cljs.core.PROTOCOL_SENTINEL === buff.cljs$core$async$impl$protocols$UnblockingBuffer$)))){
return true;
} else {
if((!buff.cljs$lang$protocol_mask$partition$)){
return cljs.core.native_satisfies_QMARK_(cljs.core.async.impl.protocols.UnblockingBuffer,buff);
} else {
return false;
}
}
} else {
return cljs.core.native_satisfies_QMARK_(cljs.core.async.impl.protocols.UnblockingBuffer,buff);
}
});
/**
 * Creates a channel with an optional buffer, an optional transducer (like (map f),
 *   (filter p) etc or a composition thereof), and an optional exception handler.
 *   If buf-or-n is a number, will create and use a fixed buffer of that size. If a
 *   transducer is supplied a buffer must be specified. ex-handler must be a
 *   fn of one argument - if an exception occurs during transformation it will be called
 *   with the thrown value as an argument, and any non-nil return value will be placed
 *   in the channel.
 */
cljs.core.async.chan = (function cljs$core$async$chan(var_args){
var G__31075 = arguments.length;
switch (G__31075) {
case 0:
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$0();

break;
case 1:
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.chan.cljs$core$IFn$_invoke$arity$0 = (function (){
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(null);
}));

(cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1 = (function (buf_or_n){
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3(buf_or_n,null,null);
}));

(cljs.core.async.chan.cljs$core$IFn$_invoke$arity$2 = (function (buf_or_n,xform){
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3(buf_or_n,xform,null);
}));

(cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3 = (function (buf_or_n,xform,ex_handler){
var buf_or_n__$1 = ((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(buf_or_n,(0)))?null:buf_or_n);
if(cljs.core.truth_(xform)){
if(cljs.core.truth_(buf_or_n__$1)){
} else {
throw (new Error(["Assert failed: ","buffer must be supplied when transducer is","\n","buf-or-n"].join('')));
}
} else {
}

return cljs.core.async.impl.channels.chan.cljs$core$IFn$_invoke$arity$3(((typeof buf_or_n__$1 === 'number')?cljs.core.async.buffer(buf_or_n__$1):buf_or_n__$1),xform,ex_handler);
}));

(cljs.core.async.chan.cljs$lang$maxFixedArity = 3);

/**
 * Creates a promise channel with an optional transducer, and an optional
 *   exception-handler. A promise channel can take exactly one value that consumers
 *   will receive. Once full, puts complete but val is dropped (no transfer).
 *   Consumers will block until either a value is placed in the channel or the
 *   channel is closed. See chan for the semantics of xform and ex-handler.
 */
cljs.core.async.promise_chan = (function cljs$core$async$promise_chan(var_args){
var G__31085 = arguments.length;
switch (G__31085) {
case 0:
return cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$0();

break;
case 1:
return cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$0 = (function (){
return cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$1(null);
}));

(cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$1 = (function (xform){
return cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$2(xform,null);
}));

(cljs.core.async.promise_chan.cljs$core$IFn$_invoke$arity$2 = (function (xform,ex_handler){
return cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3(cljs.core.async.impl.buffers.promise_buffer(),xform,ex_handler);
}));

(cljs.core.async.promise_chan.cljs$lang$maxFixedArity = 2);

/**
 * Returns a channel that will close after msecs
 */
cljs.core.async.timeout = (function cljs$core$async$timeout(msecs){
return cljs.core.async.impl.timers.timeout(msecs);
});
/**
 * takes a val from port. Must be called inside a (go ...) block. Will
 *   return nil if closed. Will park if nothing is available.
 *   Returns true unless port is already closed
 */
cljs.core.async._LT__BANG_ = (function cljs$core$async$_LT__BANG_(port){
throw (new Error("<! used not in (go ...) block"));
});
/**
 * Asynchronously takes a val from port, passing to fn1. Will pass nil
 * if closed. If on-caller? (default true) is true, and value is
 * immediately available, will call fn1 on calling thread.
 * Returns nil.
 */
cljs.core.async.take_BANG_ = (function cljs$core$async$take_BANG_(var_args){
var G__31100 = arguments.length;
switch (G__31100) {
case 2:
return cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$2 = (function (port,fn1){
return cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$3(port,fn1,true);
}));

(cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$3 = (function (port,fn1,on_caller_QMARK_){
var ret = cljs.core.async.impl.protocols.take_BANG_(port,cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$1(fn1));
if(cljs.core.truth_(ret)){
var val_34344 = cljs.core.deref(ret);
if(cljs.core.truth_(on_caller_QMARK_)){
(fn1.cljs$core$IFn$_invoke$arity$1 ? fn1.cljs$core$IFn$_invoke$arity$1(val_34344) : fn1.call(null, val_34344));
} else {
cljs.core.async.impl.dispatch.run((function (){
return (fn1.cljs$core$IFn$_invoke$arity$1 ? fn1.cljs$core$IFn$_invoke$arity$1(val_34344) : fn1.call(null, val_34344));
}));
}
} else {
}

return null;
}));

(cljs.core.async.take_BANG_.cljs$lang$maxFixedArity = 3);

cljs.core.async.nop = (function cljs$core$async$nop(_){
return null;
});
cljs.core.async.fhnop = cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$1(cljs.core.async.nop);
/**
 * puts a val into port. nil values are not allowed. Must be called
 *   inside a (go ...) block. Will park if no buffer space is available.
 *   Returns true unless port is already closed.
 */
cljs.core.async._GT__BANG_ = (function cljs$core$async$_GT__BANG_(port,val){
throw (new Error(">! used not in (go ...) block"));
});
/**
 * Asynchronously puts a val into port, calling fn1 (if supplied) when
 * complete. nil values are not allowed. Will throw if closed. If
 * on-caller? (default true) is true, and the put is immediately
 * accepted, will call fn1 on calling thread.  Returns nil.
 */
cljs.core.async.put_BANG_ = (function cljs$core$async$put_BANG_(var_args){
var G__31128 = arguments.length;
switch (G__31128) {
case 2:
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
case 4:
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2 = (function (port,val){
var temp__5823__auto__ = cljs.core.async.impl.protocols.put_BANG_(port,val,cljs.core.async.fhnop);
if(cljs.core.truth_(temp__5823__auto__)){
var ret = temp__5823__auto__;
return cljs.core.deref(ret);
} else {
return true;
}
}));

(cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$3 = (function (port,val,fn1){
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$4(port,val,fn1,true);
}));

(cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$4 = (function (port,val,fn1,on_caller_QMARK_){
var temp__5823__auto__ = cljs.core.async.impl.protocols.put_BANG_(port,val,cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$1(fn1));
if(cljs.core.truth_(temp__5823__auto__)){
var retb = temp__5823__auto__;
var ret = cljs.core.deref(retb);
if(cljs.core.truth_(on_caller_QMARK_)){
(fn1.cljs$core$IFn$_invoke$arity$1 ? fn1.cljs$core$IFn$_invoke$arity$1(ret) : fn1.call(null, ret));
} else {
cljs.core.async.impl.dispatch.run((function (){
return (fn1.cljs$core$IFn$_invoke$arity$1 ? fn1.cljs$core$IFn$_invoke$arity$1(ret) : fn1.call(null, ret));
}));
}

return ret;
} else {
return true;
}
}));

(cljs.core.async.put_BANG_.cljs$lang$maxFixedArity = 4);

cljs.core.async.close_BANG_ = (function cljs$core$async$close_BANG_(port){
return cljs.core.async.impl.protocols.close_BANG_(port);
});
cljs.core.async.random_array = (function cljs$core$async$random_array(n){
var a = (new Array(n));
var n__5593__auto___34372 = n;
var x_34373 = (0);
while(true){
if((x_34373 < n__5593__auto___34372)){
(a[x_34373] = x_34373);

var G__34374 = (x_34373 + (1));
x_34373 = G__34374;
continue;
} else {
}
break;
}

cljs.core.async.goog$module$goog$array.shuffle(a);

return a;
});

/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Handler}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async31171 = (function (flag,meta31172){
this.flag = flag;
this.meta31172 = meta31172;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_31173,meta31172__$1){
var self__ = this;
var _31173__$1 = this;
return (new cljs.core.async.t_cljs$core$async31171(self__.flag,meta31172__$1));
}));

(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_31173){
var self__ = this;
var _31173__$1 = this;
return self__.meta31172;
}));

(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$async$impl$protocols$Handler$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$async$impl$protocols$Handler$active_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.deref(self__.flag);
}));

(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$async$impl$protocols$Handler$blockable_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return true;
}));

(cljs.core.async.t_cljs$core$async31171.prototype.cljs$core$async$impl$protocols$Handler$commit$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
cljs.core.reset_BANG_(self__.flag,null);

return true;
}));

(cljs.core.async.t_cljs$core$async31171.getBasis = (function (){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"flag","flag",-1565787888,null),new cljs.core.Symbol(null,"meta31172","meta31172",966406259,null)], null);
}));

(cljs.core.async.t_cljs$core$async31171.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async31171.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async31171");

(cljs.core.async.t_cljs$core$async31171.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async31171");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async31171.
 */
cljs.core.async.__GT_t_cljs$core$async31171 = (function cljs$core$async$__GT_t_cljs$core$async31171(flag,meta31172){
return (new cljs.core.async.t_cljs$core$async31171(flag,meta31172));
});


cljs.core.async.alt_flag = (function cljs$core$async$alt_flag(){
var flag = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(true);
return (new cljs.core.async.t_cljs$core$async31171(flag,cljs.core.PersistentArrayMap.EMPTY));
});

/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Handler}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async31185 = (function (flag,cb,meta31186){
this.flag = flag;
this.cb = cb;
this.meta31186 = meta31186;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_31187,meta31186__$1){
var self__ = this;
var _31187__$1 = this;
return (new cljs.core.async.t_cljs$core$async31185(self__.flag,self__.cb,meta31186__$1));
}));

(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_31187){
var self__ = this;
var _31187__$1 = this;
return self__.meta31186;
}));

(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$async$impl$protocols$Handler$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$async$impl$protocols$Handler$active_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.active_QMARK_(self__.flag);
}));

(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$async$impl$protocols$Handler$blockable_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return true;
}));

(cljs.core.async.t_cljs$core$async31185.prototype.cljs$core$async$impl$protocols$Handler$commit$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
cljs.core.async.impl.protocols.commit(self__.flag);

return self__.cb;
}));

(cljs.core.async.t_cljs$core$async31185.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"flag","flag",-1565787888,null),new cljs.core.Symbol(null,"cb","cb",-2064487928,null),new cljs.core.Symbol(null,"meta31186","meta31186",-1865458415,null)], null);
}));

(cljs.core.async.t_cljs$core$async31185.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async31185.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async31185");

(cljs.core.async.t_cljs$core$async31185.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async31185");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async31185.
 */
cljs.core.async.__GT_t_cljs$core$async31185 = (function cljs$core$async$__GT_t_cljs$core$async31185(flag,cb,meta31186){
return (new cljs.core.async.t_cljs$core$async31185(flag,cb,meta31186));
});


cljs.core.async.alt_handler = (function cljs$core$async$alt_handler(flag,cb){
return (new cljs.core.async.t_cljs$core$async31185(flag,cb,cljs.core.PersistentArrayMap.EMPTY));
});
/**
 * returns derefable [val port] if immediate, nil if enqueued
 */
cljs.core.async.do_alts = (function cljs$core$async$do_alts(fret,ports,opts){
if((cljs.core.count(ports) > (0))){
} else {
throw (new Error(["Assert failed: ","alts must have at least one channel operation","\n","(pos? (count ports))"].join('')));
}

var flag = cljs.core.async.alt_flag();
var n = cljs.core.count(ports);
var idxs = cljs.core.async.random_array(n);
var priority = new cljs.core.Keyword(null,"priority","priority",1431093715).cljs$core$IFn$_invoke$arity$1(opts);
var ret = (function (){var i = (0);
while(true){
if((i < n)){
var idx = (cljs.core.truth_(priority)?i:(idxs[i]));
var port = cljs.core.nth.cljs$core$IFn$_invoke$arity$2(ports,idx);
var wport = ((cljs.core.vector_QMARK_(port))?(port.cljs$core$IFn$_invoke$arity$1 ? port.cljs$core$IFn$_invoke$arity$1((0)) : port.call(null, (0))):null);
var vbox = (cljs.core.truth_(wport)?(function (){var val = (port.cljs$core$IFn$_invoke$arity$1 ? port.cljs$core$IFn$_invoke$arity$1((1)) : port.call(null, (1)));
return cljs.core.async.impl.protocols.put_BANG_(wport,val,cljs.core.async.alt_handler(flag,((function (i,val,idx,port,wport,flag,n,idxs,priority){
return (function (p1__31188_SHARP_){
var G__31196 = new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [p1__31188_SHARP_,wport], null);
return (fret.cljs$core$IFn$_invoke$arity$1 ? fret.cljs$core$IFn$_invoke$arity$1(G__31196) : fret.call(null, G__31196));
});})(i,val,idx,port,wport,flag,n,idxs,priority))
));
})():cljs.core.async.impl.protocols.take_BANG_(port,cljs.core.async.alt_handler(flag,((function (i,idx,port,wport,flag,n,idxs,priority){
return (function (p1__31189_SHARP_){
var G__31197 = new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [p1__31189_SHARP_,port], null);
return (fret.cljs$core$IFn$_invoke$arity$1 ? fret.cljs$core$IFn$_invoke$arity$1(G__31197) : fret.call(null, G__31197));
});})(i,idx,port,wport,flag,n,idxs,priority))
)));
if(cljs.core.truth_(vbox)){
return cljs.core.async.impl.channels.box(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [cljs.core.deref(vbox),(function (){var or__5002__auto__ = wport;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return port;
}
})()], null));
} else {
var G__34405 = (i + (1));
i = G__34405;
continue;
}
} else {
return null;
}
break;
}
})();
var or__5002__auto__ = ret;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
if(cljs.core.contains_QMARK_(opts,new cljs.core.Keyword(null,"default","default",-1987822328))){
var temp__5825__auto__ = (function (){var and__5000__auto__ = flag.cljs$core$async$impl$protocols$Handler$active_QMARK_$arity$1(null, );
if(cljs.core.truth_(and__5000__auto__)){
return flag.cljs$core$async$impl$protocols$Handler$commit$arity$1(null, );
} else {
return and__5000__auto__;
}
})();
if(cljs.core.truth_(temp__5825__auto__)){
var got = temp__5825__auto__;
return cljs.core.async.impl.channels.box(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"default","default",-1987822328).cljs$core$IFn$_invoke$arity$1(opts),new cljs.core.Keyword(null,"default","default",-1987822328)], null));
} else {
return null;
}
} else {
return null;
}
}
});
/**
 * Completes at most one of several channel operations. Must be called
 * inside a (go ...) block. ports is a vector of channel endpoints,
 * which can be either a channel to take from or a vector of
 *   [channel-to-put-to val-to-put], in any combination. Takes will be
 *   made as if by <!, and puts will be made as if by >!. Unless
 *   the :priority option is true, if more than one port operation is
 *   ready a non-deterministic choice will be made. If no operation is
 *   ready and a :default value is supplied, [default-val :default] will
 *   be returned, otherwise alts! will park until the first operation to
 *   become ready completes. Returns [val port] of the completed
 *   operation, where val is the value taken for takes, and a
 *   boolean (true unless already closed, as per put!) for puts.
 * 
 *   opts are passed as :key val ... Supported options:
 * 
 *   :default val - the value to use if none of the operations are immediately ready
 *   :priority true - (default nil) when true, the operations will be tried in order.
 * 
 *   Note: there is no guarantee that the port exps or val exprs will be
 *   used, nor in what order should they be, so they should not be
 *   depended upon for side effects.
 */
cljs.core.async.alts_BANG_ = (function cljs$core$async$alts_BANG_(var_args){
var args__5732__auto__ = [];
var len__5726__auto___34408 = arguments.length;
var i__5727__auto___34410 = (0);
while(true){
if((i__5727__auto___34410 < len__5726__auto___34408)){
args__5732__auto__.push((arguments[i__5727__auto___34410]));

var G__34412 = (i__5727__auto___34410 + (1));
i__5727__auto___34410 = G__34412;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return cljs.core.async.alts_BANG_.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(cljs.core.async.alts_BANG_.cljs$core$IFn$_invoke$arity$variadic = (function (ports,p__31241){
var map__31242 = p__31241;
var map__31242__$1 = cljs.core.__destructure_map(map__31242);
var opts = map__31242__$1;
throw (new Error("alts! used not in (go ...) block"));
}));

(cljs.core.async.alts_BANG_.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(cljs.core.async.alts_BANG_.cljs$lang$applyTo = (function (seq31201){
var G__31222 = cljs.core.first(seq31201);
var seq31201__$1 = cljs.core.next(seq31201);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__31222,seq31201__$1);
}));

/**
 * Puts a val into port if it's possible to do so immediately.
 *   nil values are not allowed. Never blocks. Returns true if offer succeeds.
 */
cljs.core.async.offer_BANG_ = (function cljs$core$async$offer_BANG_(port,val){
var ret = cljs.core.async.impl.protocols.put_BANG_(port,val,cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$2(cljs.core.async.nop,false));
if(cljs.core.truth_(ret)){
return cljs.core.deref(ret);
} else {
return null;
}
});
/**
 * Takes a val from port if it's possible to do so immediately.
 *   Never blocks. Returns value if successful, nil otherwise.
 */
cljs.core.async.poll_BANG_ = (function cljs$core$async$poll_BANG_(port){
var ret = cljs.core.async.impl.protocols.take_BANG_(port,cljs.core.async.fn_handler.cljs$core$IFn$_invoke$arity$2(cljs.core.async.nop,false));
if(cljs.core.truth_(ret)){
return cljs.core.deref(ret);
} else {
return null;
}
});
/**
 * Takes elements from the from channel and supplies them to the to
 * channel. By default, the to channel will be closed when the from
 * channel closes, but can be determined by the close?  parameter. Will
 * stop consuming the from channel if the to channel closes
 */
cljs.core.async.pipe = (function cljs$core$async$pipe(var_args){
var G__31284 = arguments.length;
switch (G__31284) {
case 2:
return cljs.core.async.pipe.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.pipe.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.pipe.cljs$core$IFn$_invoke$arity$2 = (function (from,to){
return cljs.core.async.pipe.cljs$core$IFn$_invoke$arity$3(from,to,true);
}));

(cljs.core.async.pipe.cljs$core$IFn$_invoke$arity$3 = (function (from,to,close_QMARK_){
var c__30907__auto___34437 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_31417){
var state_val_31419 = (state_31417[(1)]);
if((state_val_31419 === (7))){
var inst_31399 = (state_31417[(2)]);
var state_31417__$1 = state_31417;
var statearr_31445_34446 = state_31417__$1;
(statearr_31445_34446[(2)] = inst_31399);

(statearr_31445_34446[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (1))){
var state_31417__$1 = state_31417;
var statearr_31446_34447 = state_31417__$1;
(statearr_31446_34447[(2)] = null);

(statearr_31446_34447[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (4))){
var inst_31314 = (state_31417[(7)]);
var inst_31314__$1 = (state_31417[(2)]);
var inst_31320 = (inst_31314__$1 == null);
var state_31417__$1 = (function (){var statearr_31449 = state_31417;
(statearr_31449[(7)] = inst_31314__$1);

return statearr_31449;
})();
if(cljs.core.truth_(inst_31320)){
var statearr_31454_34448 = state_31417__$1;
(statearr_31454_34448[(1)] = (5));

} else {
var statearr_31455_34449 = state_31417__$1;
(statearr_31455_34449[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (13))){
var state_31417__$1 = state_31417;
var statearr_31462_34450 = state_31417__$1;
(statearr_31462_34450[(2)] = null);

(statearr_31462_34450[(1)] = (14));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (6))){
var inst_31314 = (state_31417[(7)]);
var state_31417__$1 = state_31417;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31417__$1,(11),to,inst_31314);
} else {
if((state_val_31419 === (3))){
var inst_31408 = (state_31417[(2)]);
var state_31417__$1 = state_31417;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31417__$1,inst_31408);
} else {
if((state_val_31419 === (12))){
var state_31417__$1 = state_31417;
var statearr_31485_34469 = state_31417__$1;
(statearr_31485_34469[(2)] = null);

(statearr_31485_34469[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (2))){
var state_31417__$1 = state_31417;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31417__$1,(4),from);
} else {
if((state_val_31419 === (11))){
var inst_31333 = (state_31417[(2)]);
var state_31417__$1 = state_31417;
if(cljs.core.truth_(inst_31333)){
var statearr_31486_34474 = state_31417__$1;
(statearr_31486_34474[(1)] = (12));

} else {
var statearr_31487_34475 = state_31417__$1;
(statearr_31487_34475[(1)] = (13));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (9))){
var state_31417__$1 = state_31417;
var statearr_31488_34476 = state_31417__$1;
(statearr_31488_34476[(2)] = null);

(statearr_31488_34476[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (5))){
var state_31417__$1 = state_31417;
if(cljs.core.truth_(close_QMARK_)){
var statearr_31489_34478 = state_31417__$1;
(statearr_31489_34478[(1)] = (8));

} else {
var statearr_31490_34479 = state_31417__$1;
(statearr_31490_34479[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (14))){
var inst_31396 = (state_31417[(2)]);
var state_31417__$1 = state_31417;
var statearr_31491_34483 = state_31417__$1;
(statearr_31491_34483[(2)] = inst_31396);

(statearr_31491_34483[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (10))){
var inst_31330 = (state_31417[(2)]);
var state_31417__$1 = state_31417;
var statearr_31492_34484 = state_31417__$1;
(statearr_31492_34484[(2)] = inst_31330);

(statearr_31492_34484[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31419 === (8))){
var inst_31323 = cljs.core.async.close_BANG_(to);
var state_31417__$1 = state_31417;
var statearr_31493_34485 = state_31417__$1;
(statearr_31493_34485[(2)] = inst_31323);

(statearr_31493_34485[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_31494 = [null,null,null,null,null,null,null,null];
(statearr_31494[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_31494[(1)] = (1));

return statearr_31494;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_31417){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31417);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31495){var ex__30547__auto__ = e31495;
var statearr_31496_34486 = state_31417;
(statearr_31496_34486[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31417[(4)]))){
var statearr_31497_34487 = state_31417;
(statearr_31497_34487[(1)] = cljs.core.first((state_31417[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34488 = state_31417;
state_31417 = G__34488;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_31417){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_31417);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_31499 = f__30908__auto__();
(statearr_31499[(6)] = c__30907__auto___34437);

return statearr_31499;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return to;
}));

(cljs.core.async.pipe.cljs$lang$maxFixedArity = 3);

cljs.core.async.pipeline_STAR_ = (function cljs$core$async$pipeline_STAR_(n,to,xf,from,close_QMARK_,ex_handler,type){
if((n > (0))){
} else {
throw (new Error("Assert failed: (pos? n)"));
}

var jobs = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(n);
var results = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(n);
var process__$1 = (function (p__31504){
var vec__31505 = p__31504;
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__31505,(0),null);
var p = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__31505,(1),null);
var job = vec__31505;
if((job == null)){
cljs.core.async.close_BANG_(results);

return null;
} else {
var res = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$3((1),xf,ex_handler);
var c__30907__auto___34489 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_31513){
var state_val_31514 = (state_31513[(1)]);
if((state_val_31514 === (1))){
var state_31513__$1 = state_31513;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31513__$1,(2),res,v);
} else {
if((state_val_31514 === (2))){
var inst_31510 = (state_31513[(2)]);
var inst_31511 = cljs.core.async.close_BANG_(res);
var state_31513__$1 = (function (){var statearr_31520 = state_31513;
(statearr_31520[(7)] = inst_31510);

return statearr_31520;
})();
return cljs.core.async.impl.ioc_helpers.return_chan(state_31513__$1,inst_31511);
} else {
return null;
}
}
});
return (function() {
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_31530 = [null,null,null,null,null,null,null,null];
(statearr_31530[(0)] = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__);

(statearr_31530[(1)] = (1));

return statearr_31530;
});
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1 = (function (state_31513){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31513);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31535){var ex__30547__auto__ = e31535;
var statearr_31538_34495 = state_31513;
(statearr_31538_34495[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31513[(4)]))){
var statearr_31545_34500 = state_31513;
(statearr_31545_34500[(1)] = cljs.core.first((state_31513[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34501 = state_31513;
state_31513 = G__34501;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = function(state_31513){
switch(arguments.length){
case 0:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1.call(this,state_31513);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0;
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_31548 = f__30908__auto__();
(statearr_31548[(6)] = c__30907__auto___34489);

return statearr_31548;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2(p,res);

return true;
}
});
var async = (function (p__31550){
var vec__31551 = p__31550;
var v = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__31551,(0),null);
var p = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__31551,(1),null);
var job = vec__31551;
if((job == null)){
cljs.core.async.close_BANG_(results);

return null;
} else {
var res = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
(xf.cljs$core$IFn$_invoke$arity$2 ? xf.cljs$core$IFn$_invoke$arity$2(v,res) : xf.call(null, v,res));

cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2(p,res);

return true;
}
});
var n__5593__auto___34505 = n;
var __34506 = (0);
while(true){
if((__34506 < n__5593__auto___34505)){
var G__31554_34507 = type;
var G__31554_34508__$1 = (((G__31554_34507 instanceof cljs.core.Keyword))?G__31554_34507.fqn:null);
switch (G__31554_34508__$1) {
case "compute":
var c__30907__auto___34510 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run(((function (__34506,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = ((function (__34506,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function (state_31568){
var state_val_31569 = (state_31568[(1)]);
if((state_val_31569 === (1))){
var state_31568__$1 = state_31568;
var statearr_31570_34514 = state_31568__$1;
(statearr_31570_34514[(2)] = null);

(statearr_31570_34514[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31569 === (2))){
var state_31568__$1 = state_31568;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31568__$1,(4),jobs);
} else {
if((state_val_31569 === (3))){
var inst_31566 = (state_31568[(2)]);
var state_31568__$1 = state_31568;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31568__$1,inst_31566);
} else {
if((state_val_31569 === (4))){
var inst_31558 = (state_31568[(2)]);
var inst_31559 = process__$1(inst_31558);
var state_31568__$1 = state_31568;
if(cljs.core.truth_(inst_31559)){
var statearr_31574_34515 = state_31568__$1;
(statearr_31574_34515[(1)] = (5));

} else {
var statearr_31575_34516 = state_31568__$1;
(statearr_31575_34516[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31569 === (5))){
var state_31568__$1 = state_31568;
var statearr_31576_34517 = state_31568__$1;
(statearr_31576_34517[(2)] = null);

(statearr_31576_34517[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31569 === (6))){
var state_31568__$1 = state_31568;
var statearr_31577_34518 = state_31568__$1;
(statearr_31577_34518[(2)] = null);

(statearr_31577_34518[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31569 === (7))){
var inst_31564 = (state_31568[(2)]);
var state_31568__$1 = state_31568;
var statearr_31578_34519 = state_31568__$1;
(statearr_31578_34519[(2)] = inst_31564);

(statearr_31578_34519[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
});})(__34506,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
;
return ((function (__34506,switch__30543__auto__,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function() {
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_31581 = [null,null,null,null,null,null,null];
(statearr_31581[(0)] = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__);

(statearr_31581[(1)] = (1));

return statearr_31581;
});
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1 = (function (state_31568){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31568);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31582){var ex__30547__auto__ = e31582;
var statearr_31583_34523 = state_31568;
(statearr_31583_34523[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31568[(4)]))){
var statearr_31585_34524 = state_31568;
(statearr_31585_34524[(1)] = cljs.core.first((state_31568[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34525 = state_31568;
state_31568 = G__34525;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = function(state_31568){
switch(arguments.length){
case 0:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1.call(this,state_31568);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0;
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__;
})()
;})(__34506,switch__30543__auto__,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
})();
var state__30909__auto__ = (function (){var statearr_31587 = f__30908__auto__();
(statearr_31587[(6)] = c__30907__auto___34510);

return statearr_31587;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
});})(__34506,c__30907__auto___34510,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
);


break;
case "async":
var c__30907__auto___34526 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run(((function (__34506,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = ((function (__34506,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function (state_31621){
var state_val_31622 = (state_31621[(1)]);
if((state_val_31622 === (1))){
var state_31621__$1 = state_31621;
var statearr_31648_34528 = state_31621__$1;
(statearr_31648_34528[(2)] = null);

(statearr_31648_34528[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31622 === (2))){
var state_31621__$1 = state_31621;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31621__$1,(4),jobs);
} else {
if((state_val_31622 === (3))){
var inst_31619 = (state_31621[(2)]);
var state_31621__$1 = state_31621;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31621__$1,inst_31619);
} else {
if((state_val_31622 === (4))){
var inst_31590 = (state_31621[(2)]);
var inst_31607 = async(inst_31590);
var state_31621__$1 = state_31621;
if(cljs.core.truth_(inst_31607)){
var statearr_31681_34539 = state_31621__$1;
(statearr_31681_34539[(1)] = (5));

} else {
var statearr_31682_34540 = state_31621__$1;
(statearr_31682_34540[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31622 === (5))){
var state_31621__$1 = state_31621;
var statearr_31688_34541 = state_31621__$1;
(statearr_31688_34541[(2)] = null);

(statearr_31688_34541[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31622 === (6))){
var state_31621__$1 = state_31621;
var statearr_31690_34542 = state_31621__$1;
(statearr_31690_34542[(2)] = null);

(statearr_31690_34542[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31622 === (7))){
var inst_31617 = (state_31621[(2)]);
var state_31621__$1 = state_31621;
var statearr_31692_34543 = state_31621__$1;
(statearr_31692_34543[(2)] = inst_31617);

(statearr_31692_34543[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
});})(__34506,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
;
return ((function (__34506,switch__30543__auto__,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async){
return (function() {
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_31696 = [null,null,null,null,null,null,null];
(statearr_31696[(0)] = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__);

(statearr_31696[(1)] = (1));

return statearr_31696;
});
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1 = (function (state_31621){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31621);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31701){var ex__30547__auto__ = e31701;
var statearr_31705_34545 = state_31621;
(statearr_31705_34545[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31621[(4)]))){
var statearr_31706_34546 = state_31621;
(statearr_31706_34546[(1)] = cljs.core.first((state_31621[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34547 = state_31621;
state_31621 = G__34547;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = function(state_31621){
switch(arguments.length){
case 0:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1.call(this,state_31621);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0;
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__;
})()
;})(__34506,switch__30543__auto__,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
})();
var state__30909__auto__ = (function (){var statearr_31707 = f__30908__auto__();
(statearr_31707[(6)] = c__30907__auto___34526);

return statearr_31707;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
});})(__34506,c__30907__auto___34526,G__31554_34507,G__31554_34508__$1,n__5593__auto___34505,jobs,results,process__$1,async))
);


break;
default:
throw (new Error(["No matching clause: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(G__31554_34508__$1)].join('')));

}

var G__34548 = (__34506 + (1));
__34506 = G__34548;
continue;
} else {
}
break;
}

var c__30907__auto___34549 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_31733){
var state_val_31734 = (state_31733[(1)]);
if((state_val_31734 === (7))){
var inst_31729 = (state_31733[(2)]);
var state_31733__$1 = state_31733;
var statearr_31744_34550 = state_31733__$1;
(statearr_31744_34550[(2)] = inst_31729);

(statearr_31744_34550[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31734 === (1))){
var state_31733__$1 = state_31733;
var statearr_31745_34551 = state_31733__$1;
(statearr_31745_34551[(2)] = null);

(statearr_31745_34551[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31734 === (4))){
var inst_31713 = (state_31733[(7)]);
var inst_31713__$1 = (state_31733[(2)]);
var inst_31714 = (inst_31713__$1 == null);
var state_31733__$1 = (function (){var statearr_31747 = state_31733;
(statearr_31747[(7)] = inst_31713__$1);

return statearr_31747;
})();
if(cljs.core.truth_(inst_31714)){
var statearr_31748_34554 = state_31733__$1;
(statearr_31748_34554[(1)] = (5));

} else {
var statearr_31750_34557 = state_31733__$1;
(statearr_31750_34557[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31734 === (6))){
var inst_31713 = (state_31733[(7)]);
var inst_31718 = (state_31733[(8)]);
var inst_31718__$1 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
var inst_31720 = cljs.core.PersistentVector.EMPTY_NODE;
var inst_31721 = [inst_31713,inst_31718__$1];
var inst_31722 = (new cljs.core.PersistentVector(null,2,(5),inst_31720,inst_31721,null));
var state_31733__$1 = (function (){var statearr_31755 = state_31733;
(statearr_31755[(8)] = inst_31718__$1);

return statearr_31755;
})();
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31733__$1,(8),jobs,inst_31722);
} else {
if((state_val_31734 === (3))){
var inst_31731 = (state_31733[(2)]);
var state_31733__$1 = state_31733;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31733__$1,inst_31731);
} else {
if((state_val_31734 === (2))){
var state_31733__$1 = state_31733;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31733__$1,(4),from);
} else {
if((state_val_31734 === (9))){
var inst_31726 = (state_31733[(2)]);
var state_31733__$1 = (function (){var statearr_31757 = state_31733;
(statearr_31757[(9)] = inst_31726);

return statearr_31757;
})();
var statearr_31758_34562 = state_31733__$1;
(statearr_31758_34562[(2)] = null);

(statearr_31758_34562[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31734 === (5))){
var inst_31716 = cljs.core.async.close_BANG_(jobs);
var state_31733__$1 = state_31733;
var statearr_31760_34563 = state_31733__$1;
(statearr_31760_34563[(2)] = inst_31716);

(statearr_31760_34563[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31734 === (8))){
var inst_31718 = (state_31733[(8)]);
var inst_31724 = (state_31733[(2)]);
var state_31733__$1 = (function (){var statearr_31762 = state_31733;
(statearr_31762[(10)] = inst_31724);

return statearr_31762;
})();
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31733__$1,(9),results,inst_31718);
} else {
return null;
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_31763 = [null,null,null,null,null,null,null,null,null,null,null];
(statearr_31763[(0)] = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__);

(statearr_31763[(1)] = (1));

return statearr_31763;
});
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1 = (function (state_31733){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31733);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31764){var ex__30547__auto__ = e31764;
var statearr_31765_34564 = state_31733;
(statearr_31765_34564[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31733[(4)]))){
var statearr_31766_34565 = state_31733;
(statearr_31766_34565[(1)] = cljs.core.first((state_31733[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34572 = state_31733;
state_31733 = G__34572;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = function(state_31733){
switch(arguments.length){
case 0:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1.call(this,state_31733);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0;
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_31773 = f__30908__auto__();
(statearr_31773[(6)] = c__30907__auto___34549);

return statearr_31773;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


var c__30907__auto__ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_31816){
var state_val_31817 = (state_31816[(1)]);
if((state_val_31817 === (7))){
var inst_31809 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
var statearr_31818_34580 = state_31816__$1;
(statearr_31818_34580[(2)] = inst_31809);

(statearr_31818_34580[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (20))){
var state_31816__$1 = state_31816;
var statearr_31819_34581 = state_31816__$1;
(statearr_31819_34581[(2)] = null);

(statearr_31819_34581[(1)] = (21));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (1))){
var state_31816__$1 = state_31816;
var statearr_31820_34584 = state_31816__$1;
(statearr_31820_34584[(2)] = null);

(statearr_31820_34584[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (4))){
var inst_31777 = (state_31816[(7)]);
var inst_31777__$1 = (state_31816[(2)]);
var inst_31778 = (inst_31777__$1 == null);
var state_31816__$1 = (function (){var statearr_31821 = state_31816;
(statearr_31821[(7)] = inst_31777__$1);

return statearr_31821;
})();
if(cljs.core.truth_(inst_31778)){
var statearr_31822_34587 = state_31816__$1;
(statearr_31822_34587[(1)] = (5));

} else {
var statearr_31823_34588 = state_31816__$1;
(statearr_31823_34588[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (15))){
var inst_31790 = (state_31816[(8)]);
var state_31816__$1 = state_31816;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31816__$1,(18),to,inst_31790);
} else {
if((state_val_31817 === (21))){
var inst_31803 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
var statearr_31824_34589 = state_31816__$1;
(statearr_31824_34589[(2)] = inst_31803);

(statearr_31824_34589[(1)] = (13));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (13))){
var inst_31806 = (state_31816[(2)]);
var state_31816__$1 = (function (){var statearr_31829 = state_31816;
(statearr_31829[(9)] = inst_31806);

return statearr_31829;
})();
var statearr_31830_34590 = state_31816__$1;
(statearr_31830_34590[(2)] = null);

(statearr_31830_34590[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (6))){
var inst_31777 = (state_31816[(7)]);
var state_31816__$1 = state_31816;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31816__$1,(11),inst_31777);
} else {
if((state_val_31817 === (17))){
var inst_31798 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
if(cljs.core.truth_(inst_31798)){
var statearr_31832_34591 = state_31816__$1;
(statearr_31832_34591[(1)] = (19));

} else {
var statearr_31833_34592 = state_31816__$1;
(statearr_31833_34592[(1)] = (20));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (3))){
var inst_31811 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31816__$1,inst_31811);
} else {
if((state_val_31817 === (12))){
var inst_31787 = (state_31816[(10)]);
var state_31816__$1 = state_31816;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31816__$1,(14),inst_31787);
} else {
if((state_val_31817 === (2))){
var state_31816__$1 = state_31816;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31816__$1,(4),results);
} else {
if((state_val_31817 === (19))){
var state_31816__$1 = state_31816;
var statearr_31845_34593 = state_31816__$1;
(statearr_31845_34593[(2)] = null);

(statearr_31845_34593[(1)] = (12));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (11))){
var inst_31787 = (state_31816[(2)]);
var state_31816__$1 = (function (){var statearr_31846 = state_31816;
(statearr_31846[(10)] = inst_31787);

return statearr_31846;
})();
var statearr_31847_34594 = state_31816__$1;
(statearr_31847_34594[(2)] = null);

(statearr_31847_34594[(1)] = (12));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (9))){
var state_31816__$1 = state_31816;
var statearr_31848_34595 = state_31816__$1;
(statearr_31848_34595[(2)] = null);

(statearr_31848_34595[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (5))){
var state_31816__$1 = state_31816;
if(cljs.core.truth_(close_QMARK_)){
var statearr_31850_34596 = state_31816__$1;
(statearr_31850_34596[(1)] = (8));

} else {
var statearr_31851_34597 = state_31816__$1;
(statearr_31851_34597[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (14))){
var inst_31792 = (state_31816[(11)]);
var inst_31790 = (state_31816[(8)]);
var inst_31790__$1 = (state_31816[(2)]);
var inst_31791 = (inst_31790__$1 == null);
var inst_31792__$1 = cljs.core.not(inst_31791);
var state_31816__$1 = (function (){var statearr_31853 = state_31816;
(statearr_31853[(11)] = inst_31792__$1);

(statearr_31853[(8)] = inst_31790__$1);

return statearr_31853;
})();
if(inst_31792__$1){
var statearr_31854_34598 = state_31816__$1;
(statearr_31854_34598[(1)] = (15));

} else {
var statearr_31855_34599 = state_31816__$1;
(statearr_31855_34599[(1)] = (16));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (16))){
var inst_31792 = (state_31816[(11)]);
var state_31816__$1 = state_31816;
var statearr_31857_34600 = state_31816__$1;
(statearr_31857_34600[(2)] = inst_31792);

(statearr_31857_34600[(1)] = (17));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (10))){
var inst_31784 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
var statearr_31861_34604 = state_31816__$1;
(statearr_31861_34604[(2)] = inst_31784);

(statearr_31861_34604[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (18))){
var inst_31795 = (state_31816[(2)]);
var state_31816__$1 = state_31816;
var statearr_31862_34605 = state_31816__$1;
(statearr_31862_34605[(2)] = inst_31795);

(statearr_31862_34605[(1)] = (17));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31817 === (8))){
var inst_31781 = cljs.core.async.close_BANG_(to);
var state_31816__$1 = state_31816;
var statearr_31863_34606 = state_31816__$1;
(statearr_31863_34606[(2)] = inst_31781);

(statearr_31863_34606[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_31864 = [null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_31864[(0)] = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__);

(statearr_31864[(1)] = (1));

return statearr_31864;
});
var cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1 = (function (state_31816){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31816);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e31866){var ex__30547__auto__ = e31866;
var statearr_31867_34607 = state_31816;
(statearr_31867_34607[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31816[(4)]))){
var statearr_31870_34608 = state_31816;
(statearr_31870_34608[(1)] = cljs.core.first((state_31816[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34609 = state_31816;
state_31816 = G__34609;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__ = function(state_31816){
switch(arguments.length){
case 0:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1.call(this,state_31816);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____0;
cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$pipeline_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$pipeline_STAR__$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_31871 = f__30908__auto__();
(statearr_31871[(6)] = c__30907__auto__);

return statearr_31871;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

return c__30907__auto__;
});
/**
 * Takes elements from the from channel and supplies them to the to
 *   channel, subject to the async function af, with parallelism n. af
 *   must be a function of two arguments, the first an input value and
 *   the second a channel on which to place the result(s). The
 *   presumption is that af will return immediately, having launched some
 *   asynchronous operation whose completion/callback will put results on
 *   the channel, then close! it. Outputs will be returned in order
 *   relative to the inputs. By default, the to channel will be closed
 *   when the from channel closes, but can be determined by the close?
 *   parameter. Will stop consuming the from channel if the to channel
 *   closes. See also pipeline, pipeline-blocking.
 */
cljs.core.async.pipeline_async = (function cljs$core$async$pipeline_async(var_args){
var G__31876 = arguments.length;
switch (G__31876) {
case 4:
return cljs.core.async.pipeline_async.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
case 5:
return cljs.core.async.pipeline_async.cljs$core$IFn$_invoke$arity$5((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]),(arguments[(4)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.pipeline_async.cljs$core$IFn$_invoke$arity$4 = (function (n,to,af,from){
return cljs.core.async.pipeline_async.cljs$core$IFn$_invoke$arity$5(n,to,af,from,true);
}));

(cljs.core.async.pipeline_async.cljs$core$IFn$_invoke$arity$5 = (function (n,to,af,from,close_QMARK_){
return cljs.core.async.pipeline_STAR_(n,to,af,from,close_QMARK_,null,new cljs.core.Keyword(null,"async","async",1050769601));
}));

(cljs.core.async.pipeline_async.cljs$lang$maxFixedArity = 5);

/**
 * Takes elements from the from channel and supplies them to the to
 *   channel, subject to the transducer xf, with parallelism n. Because
 *   it is parallel, the transducer will be applied independently to each
 *   element, not across elements, and may produce zero or more outputs
 *   per input.  Outputs will be returned in order relative to the
 *   inputs. By default, the to channel will be closed when the from
 *   channel closes, but can be determined by the close?  parameter. Will
 *   stop consuming the from channel if the to channel closes.
 * 
 *   Note this is supplied for API compatibility with the Clojure version.
 *   Values of N > 1 will not result in actual concurrency in a
 *   single-threaded runtime.
 */
cljs.core.async.pipeline = (function cljs$core$async$pipeline(var_args){
var G__31883 = arguments.length;
switch (G__31883) {
case 4:
return cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
case 5:
return cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$5((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]),(arguments[(4)]));

break;
case 6:
return cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$6((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]),(arguments[(4)]),(arguments[(5)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$4 = (function (n,to,xf,from){
return cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$5(n,to,xf,from,true);
}));

(cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$5 = (function (n,to,xf,from,close_QMARK_){
return cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$6(n,to,xf,from,close_QMARK_,null);
}));

(cljs.core.async.pipeline.cljs$core$IFn$_invoke$arity$6 = (function (n,to,xf,from,close_QMARK_,ex_handler){
return cljs.core.async.pipeline_STAR_(n,to,xf,from,close_QMARK_,ex_handler,new cljs.core.Keyword(null,"compute","compute",1555393130));
}));

(cljs.core.async.pipeline.cljs$lang$maxFixedArity = 6);

/**
 * Takes a predicate and a source channel and returns a vector of two
 *   channels, the first of which will contain the values for which the
 *   predicate returned true, the second those for which it returned
 *   false.
 * 
 *   The out channels will be unbuffered by default, or two buf-or-ns can
 *   be supplied. The channels will close after the source channel has
 *   closed.
 */
cljs.core.async.split = (function cljs$core$async$split(var_args){
var G__31892 = arguments.length;
switch (G__31892) {
case 2:
return cljs.core.async.split.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 4:
return cljs.core.async.split.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.split.cljs$core$IFn$_invoke$arity$2 = (function (p,ch){
return cljs.core.async.split.cljs$core$IFn$_invoke$arity$4(p,ch,null,null);
}));

(cljs.core.async.split.cljs$core$IFn$_invoke$arity$4 = (function (p,ch,t_buf_or_n,f_buf_or_n){
var tc = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(t_buf_or_n);
var fc = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(f_buf_or_n);
var c__30907__auto___34633 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_31922){
var state_val_31923 = (state_31922[(1)]);
if((state_val_31923 === (7))){
var inst_31917 = (state_31922[(2)]);
var state_31922__$1 = state_31922;
var statearr_31929_34634 = state_31922__$1;
(statearr_31929_34634[(2)] = inst_31917);

(statearr_31929_34634[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (1))){
var state_31922__$1 = state_31922;
var statearr_31934_34637 = state_31922__$1;
(statearr_31934_34637[(2)] = null);

(statearr_31934_34637[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (4))){
var inst_31897 = (state_31922[(7)]);
var inst_31897__$1 = (state_31922[(2)]);
var inst_31898 = (inst_31897__$1 == null);
var state_31922__$1 = (function (){var statearr_31948 = state_31922;
(statearr_31948[(7)] = inst_31897__$1);

return statearr_31948;
})();
if(cljs.core.truth_(inst_31898)){
var statearr_31951_34644 = state_31922__$1;
(statearr_31951_34644[(1)] = (5));

} else {
var statearr_31952_34645 = state_31922__$1;
(statearr_31952_34645[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (13))){
var state_31922__$1 = state_31922;
var statearr_31958_34646 = state_31922__$1;
(statearr_31958_34646[(2)] = null);

(statearr_31958_34646[(1)] = (14));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (6))){
var inst_31897 = (state_31922[(7)]);
var inst_31904 = (p.cljs$core$IFn$_invoke$arity$1 ? p.cljs$core$IFn$_invoke$arity$1(inst_31897) : p.call(null, inst_31897));
var state_31922__$1 = state_31922;
if(cljs.core.truth_(inst_31904)){
var statearr_31961_34647 = state_31922__$1;
(statearr_31961_34647[(1)] = (9));

} else {
var statearr_31962_34648 = state_31922__$1;
(statearr_31962_34648[(1)] = (10));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (3))){
var inst_31919 = (state_31922[(2)]);
var state_31922__$1 = state_31922;
return cljs.core.async.impl.ioc_helpers.return_chan(state_31922__$1,inst_31919);
} else {
if((state_val_31923 === (12))){
var state_31922__$1 = state_31922;
var statearr_31965_34649 = state_31922__$1;
(statearr_31965_34649[(2)] = null);

(statearr_31965_34649[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (2))){
var state_31922__$1 = state_31922;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_31922__$1,(4),ch);
} else {
if((state_val_31923 === (11))){
var inst_31897 = (state_31922[(7)]);
var inst_31908 = (state_31922[(2)]);
var state_31922__$1 = state_31922;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_31922__$1,(8),inst_31908,inst_31897);
} else {
if((state_val_31923 === (9))){
var state_31922__$1 = state_31922;
var statearr_31997_34654 = state_31922__$1;
(statearr_31997_34654[(2)] = tc);

(statearr_31997_34654[(1)] = (11));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (5))){
var inst_31900 = cljs.core.async.close_BANG_(tc);
var inst_31901 = cljs.core.async.close_BANG_(fc);
var state_31922__$1 = (function (){var statearr_32001 = state_31922;
(statearr_32001[(8)] = inst_31900);

return statearr_32001;
})();
var statearr_32005_34659 = state_31922__$1;
(statearr_32005_34659[(2)] = inst_31901);

(statearr_32005_34659[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (14))){
var inst_31915 = (state_31922[(2)]);
var state_31922__$1 = state_31922;
var statearr_32018_34661 = state_31922__$1;
(statearr_32018_34661[(2)] = inst_31915);

(statearr_32018_34661[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (10))){
var state_31922__$1 = state_31922;
var statearr_32020_34685 = state_31922__$1;
(statearr_32020_34685[(2)] = fc);

(statearr_32020_34685[(1)] = (11));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_31923 === (8))){
var inst_31910 = (state_31922[(2)]);
var state_31922__$1 = state_31922;
if(cljs.core.truth_(inst_31910)){
var statearr_32022_34687 = state_31922__$1;
(statearr_32022_34687[(1)] = (12));

} else {
var statearr_32023_34688 = state_31922__$1;
(statearr_32023_34688[(1)] = (13));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_32030 = [null,null,null,null,null,null,null,null,null];
(statearr_32030[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_32030[(1)] = (1));

return statearr_32030;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_31922){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_31922);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e32034){var ex__30547__auto__ = e32034;
var statearr_32035_34692 = state_31922;
(statearr_32035_34692[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_31922[(4)]))){
var statearr_32036_34696 = state_31922;
(statearr_32036_34696[(1)] = cljs.core.first((state_31922[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34697 = state_31922;
state_31922 = G__34697;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_31922){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_31922);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_32038 = f__30908__auto__();
(statearr_32038[(6)] = c__30907__auto___34633);

return statearr_32038;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [tc,fc], null);
}));

(cljs.core.async.split.cljs$lang$maxFixedArity = 4);

/**
 * f should be a function of 2 arguments. Returns a channel containing
 *   the single result of applying f to init and the first item from the
 *   channel, then applying f to that result and the 2nd item, etc. If
 *   the channel closes without yielding items, returns init and f is not
 *   called. ch must close before reduce produces a result.
 */
cljs.core.async.reduce = (function cljs$core$async$reduce(f,init,ch){
var c__30907__auto__ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_32065){
var state_val_32066 = (state_32065[(1)]);
if((state_val_32066 === (7))){
var inst_32057 = (state_32065[(2)]);
var state_32065__$1 = state_32065;
var statearr_32077_34702 = state_32065__$1;
(statearr_32077_34702[(2)] = inst_32057);

(statearr_32077_34702[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (1))){
var inst_32039 = init;
var inst_32040 = inst_32039;
var state_32065__$1 = (function (){var statearr_32080 = state_32065;
(statearr_32080[(7)] = inst_32040);

return statearr_32080;
})();
var statearr_32081_34703 = state_32065__$1;
(statearr_32081_34703[(2)] = null);

(statearr_32081_34703[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (4))){
var inst_32043 = (state_32065[(8)]);
var inst_32043__$1 = (state_32065[(2)]);
var inst_32044 = (inst_32043__$1 == null);
var state_32065__$1 = (function (){var statearr_32082 = state_32065;
(statearr_32082[(8)] = inst_32043__$1);

return statearr_32082;
})();
if(cljs.core.truth_(inst_32044)){
var statearr_32083_34704 = state_32065__$1;
(statearr_32083_34704[(1)] = (5));

} else {
var statearr_32084_34705 = state_32065__$1;
(statearr_32084_34705[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (6))){
var inst_32043 = (state_32065[(8)]);
var inst_32048 = (state_32065[(9)]);
var inst_32040 = (state_32065[(7)]);
var inst_32048__$1 = (f.cljs$core$IFn$_invoke$arity$2 ? f.cljs$core$IFn$_invoke$arity$2(inst_32040,inst_32043) : f.call(null, inst_32040,inst_32043));
var inst_32049 = cljs.core.reduced_QMARK_(inst_32048__$1);
var state_32065__$1 = (function (){var statearr_32085 = state_32065;
(statearr_32085[(9)] = inst_32048__$1);

return statearr_32085;
})();
if(inst_32049){
var statearr_32086_34706 = state_32065__$1;
(statearr_32086_34706[(1)] = (8));

} else {
var statearr_32087_34707 = state_32065__$1;
(statearr_32087_34707[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (3))){
var inst_32059 = (state_32065[(2)]);
var state_32065__$1 = state_32065;
return cljs.core.async.impl.ioc_helpers.return_chan(state_32065__$1,inst_32059);
} else {
if((state_val_32066 === (2))){
var state_32065__$1 = state_32065;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_32065__$1,(4),ch);
} else {
if((state_val_32066 === (9))){
var inst_32048 = (state_32065[(9)]);
var inst_32040 = inst_32048;
var state_32065__$1 = (function (){var statearr_32088 = state_32065;
(statearr_32088[(7)] = inst_32040);

return statearr_32088;
})();
var statearr_32089_34710 = state_32065__$1;
(statearr_32089_34710[(2)] = null);

(statearr_32089_34710[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (5))){
var inst_32040 = (state_32065[(7)]);
var state_32065__$1 = state_32065;
var statearr_32091_34713 = state_32065__$1;
(statearr_32091_34713[(2)] = inst_32040);

(statearr_32091_34713[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (10))){
var inst_32055 = (state_32065[(2)]);
var state_32065__$1 = state_32065;
var statearr_32092_34714 = state_32065__$1;
(statearr_32092_34714[(2)] = inst_32055);

(statearr_32092_34714[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32066 === (8))){
var inst_32048 = (state_32065[(9)]);
var inst_32051 = cljs.core.deref(inst_32048);
var state_32065__$1 = state_32065;
var statearr_32095_34722 = state_32065__$1;
(statearr_32095_34722[(2)] = inst_32051);

(statearr_32095_34722[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$reduce_$_state_machine__30544__auto__ = null;
var cljs$core$async$reduce_$_state_machine__30544__auto____0 = (function (){
var statearr_32107 = [null,null,null,null,null,null,null,null,null,null];
(statearr_32107[(0)] = cljs$core$async$reduce_$_state_machine__30544__auto__);

(statearr_32107[(1)] = (1));

return statearr_32107;
});
var cljs$core$async$reduce_$_state_machine__30544__auto____1 = (function (state_32065){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_32065);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e32111){var ex__30547__auto__ = e32111;
var statearr_32112_34731 = state_32065;
(statearr_32112_34731[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_32065[(4)]))){
var statearr_32114_34732 = state_32065;
(statearr_32114_34732[(1)] = cljs.core.first((state_32065[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34736 = state_32065;
state_32065 = G__34736;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$reduce_$_state_machine__30544__auto__ = function(state_32065){
switch(arguments.length){
case 0:
return cljs$core$async$reduce_$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$reduce_$_state_machine__30544__auto____1.call(this,state_32065);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$reduce_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$reduce_$_state_machine__30544__auto____0;
cljs$core$async$reduce_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$reduce_$_state_machine__30544__auto____1;
return cljs$core$async$reduce_$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_32124 = f__30908__auto__();
(statearr_32124[(6)] = c__30907__auto__);

return statearr_32124;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

return c__30907__auto__;
});
/**
 * async/reduces a channel with a transformation (xform f).
 *   Returns a channel containing the result.  ch must close before
 *   transduce produces a result.
 */
cljs.core.async.transduce = (function cljs$core$async$transduce(xform,f,init,ch){
var f__$1 = (xform.cljs$core$IFn$_invoke$arity$1 ? xform.cljs$core$IFn$_invoke$arity$1(f) : xform.call(null, f));
var c__30907__auto__ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_32132){
var state_val_32133 = (state_32132[(1)]);
if((state_val_32133 === (1))){
var inst_32127 = cljs.core.async.reduce(f__$1,init,ch);
var state_32132__$1 = state_32132;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_32132__$1,(2),inst_32127);
} else {
if((state_val_32133 === (2))){
var inst_32129 = (state_32132[(2)]);
var inst_32130 = (f__$1.cljs$core$IFn$_invoke$arity$1 ? f__$1.cljs$core$IFn$_invoke$arity$1(inst_32129) : f__$1.call(null, inst_32129));
var state_32132__$1 = state_32132;
return cljs.core.async.impl.ioc_helpers.return_chan(state_32132__$1,inst_32130);
} else {
return null;
}
}
});
return (function() {
var cljs$core$async$transduce_$_state_machine__30544__auto__ = null;
var cljs$core$async$transduce_$_state_machine__30544__auto____0 = (function (){
var statearr_32138 = [null,null,null,null,null,null,null];
(statearr_32138[(0)] = cljs$core$async$transduce_$_state_machine__30544__auto__);

(statearr_32138[(1)] = (1));

return statearr_32138;
});
var cljs$core$async$transduce_$_state_machine__30544__auto____1 = (function (state_32132){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_32132);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e32144){var ex__30547__auto__ = e32144;
var statearr_32146_34762 = state_32132;
(statearr_32146_34762[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_32132[(4)]))){
var statearr_32147_34763 = state_32132;
(statearr_32147_34763[(1)] = cljs.core.first((state_32132[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34767 = state_32132;
state_32132 = G__34767;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$transduce_$_state_machine__30544__auto__ = function(state_32132){
switch(arguments.length){
case 0:
return cljs$core$async$transduce_$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$transduce_$_state_machine__30544__auto____1.call(this,state_32132);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$transduce_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$transduce_$_state_machine__30544__auto____0;
cljs$core$async$transduce_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$transduce_$_state_machine__30544__auto____1;
return cljs$core$async$transduce_$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_32152 = f__30908__auto__();
(statearr_32152[(6)] = c__30907__auto__);

return statearr_32152;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

return c__30907__auto__;
});
/**
 * Puts the contents of coll into the supplied channel.
 * 
 *   By default the channel will be closed after the items are copied,
 *   but can be determined by the close? parameter.
 * 
 *   Returns a channel which will close after the items are copied.
 */
cljs.core.async.onto_chan_BANG_ = (function cljs$core$async$onto_chan_BANG_(var_args){
var G__32159 = arguments.length;
switch (G__32159) {
case 2:
return cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$2 = (function (ch,coll){
return cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$3(ch,coll,true);
}));

(cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$3 = (function (ch,coll,close_QMARK_){
var c__30907__auto__ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_32193){
var state_val_32194 = (state_32193[(1)]);
if((state_val_32194 === (7))){
var inst_32174 = (state_32193[(2)]);
var state_32193__$1 = state_32193;
var statearr_32199_34782 = state_32193__$1;
(statearr_32199_34782[(2)] = inst_32174);

(statearr_32199_34782[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (1))){
var inst_32168 = cljs.core.seq(coll);
var inst_32169 = inst_32168;
var state_32193__$1 = (function (){var statearr_32203 = state_32193;
(statearr_32203[(7)] = inst_32169);

return statearr_32203;
})();
var statearr_32204_34783 = state_32193__$1;
(statearr_32204_34783[(2)] = null);

(statearr_32204_34783[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (4))){
var inst_32169 = (state_32193[(7)]);
var inst_32172 = cljs.core.first(inst_32169);
var state_32193__$1 = state_32193;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_32193__$1,(7),ch,inst_32172);
} else {
if((state_val_32194 === (13))){
var inst_32187 = (state_32193[(2)]);
var state_32193__$1 = state_32193;
var statearr_32206_34785 = state_32193__$1;
(statearr_32206_34785[(2)] = inst_32187);

(statearr_32206_34785[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (6))){
var inst_32177 = (state_32193[(2)]);
var state_32193__$1 = state_32193;
if(cljs.core.truth_(inst_32177)){
var statearr_32209_34786 = state_32193__$1;
(statearr_32209_34786[(1)] = (8));

} else {
var statearr_32212_34787 = state_32193__$1;
(statearr_32212_34787[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (3))){
var inst_32191 = (state_32193[(2)]);
var state_32193__$1 = state_32193;
return cljs.core.async.impl.ioc_helpers.return_chan(state_32193__$1,inst_32191);
} else {
if((state_val_32194 === (12))){
var state_32193__$1 = state_32193;
var statearr_32213_34790 = state_32193__$1;
(statearr_32213_34790[(2)] = null);

(statearr_32213_34790[(1)] = (13));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (2))){
var inst_32169 = (state_32193[(7)]);
var state_32193__$1 = state_32193;
if(cljs.core.truth_(inst_32169)){
var statearr_32215_34793 = state_32193__$1;
(statearr_32215_34793[(1)] = (4));

} else {
var statearr_32218_34794 = state_32193__$1;
(statearr_32218_34794[(1)] = (5));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (11))){
var inst_32183 = cljs.core.async.close_BANG_(ch);
var state_32193__$1 = state_32193;
var statearr_32220_34797 = state_32193__$1;
(statearr_32220_34797[(2)] = inst_32183);

(statearr_32220_34797[(1)] = (13));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (9))){
var state_32193__$1 = state_32193;
if(cljs.core.truth_(close_QMARK_)){
var statearr_32221_34801 = state_32193__$1;
(statearr_32221_34801[(1)] = (11));

} else {
var statearr_32222_34803 = state_32193__$1;
(statearr_32222_34803[(1)] = (12));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (5))){
var inst_32169 = (state_32193[(7)]);
var state_32193__$1 = state_32193;
var statearr_32223_34805 = state_32193__$1;
(statearr_32223_34805[(2)] = inst_32169);

(statearr_32223_34805[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (10))){
var inst_32189 = (state_32193[(2)]);
var state_32193__$1 = state_32193;
var statearr_32225_34808 = state_32193__$1;
(statearr_32225_34808[(2)] = inst_32189);

(statearr_32225_34808[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32194 === (8))){
var inst_32169 = (state_32193[(7)]);
var inst_32179 = cljs.core.next(inst_32169);
var inst_32169__$1 = inst_32179;
var state_32193__$1 = (function (){var statearr_32229 = state_32193;
(statearr_32229[(7)] = inst_32169__$1);

return statearr_32229;
})();
var statearr_32232_34809 = state_32193__$1;
(statearr_32232_34809[(2)] = null);

(statearr_32232_34809[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_32233 = [null,null,null,null,null,null,null,null];
(statearr_32233[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_32233[(1)] = (1));

return statearr_32233;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_32193){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_32193);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e32234){var ex__30547__auto__ = e32234;
var statearr_32236_34816 = state_32193;
(statearr_32236_34816[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_32193[(4)]))){
var statearr_32237_34837 = state_32193;
(statearr_32237_34837[(1)] = cljs.core.first((state_32193[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__34838 = state_32193;
state_32193 = G__34838;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_32193){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_32193);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_32239 = f__30908__auto__();
(statearr_32239[(6)] = c__30907__auto__);

return statearr_32239;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

return c__30907__auto__;
}));

(cljs.core.async.onto_chan_BANG_.cljs$lang$maxFixedArity = 3);

/**
 * Creates and returns a channel which contains the contents of coll,
 *   closing when exhausted.
 */
cljs.core.async.to_chan_BANG_ = (function cljs$core$async$to_chan_BANG_(coll){
var ch = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(cljs.core.bounded_count((100),coll));
cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$2(ch,coll);

return ch;
});
/**
 * Deprecated - use onto-chan!
 */
cljs.core.async.onto_chan = (function cljs$core$async$onto_chan(var_args){
var G__32250 = arguments.length;
switch (G__32250) {
case 2:
return cljs.core.async.onto_chan.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.onto_chan.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.onto_chan.cljs$core$IFn$_invoke$arity$2 = (function (ch,coll){
return cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$3(ch,coll,true);
}));

(cljs.core.async.onto_chan.cljs$core$IFn$_invoke$arity$3 = (function (ch,coll,close_QMARK_){
return cljs.core.async.onto_chan_BANG_.cljs$core$IFn$_invoke$arity$3(ch,coll,close_QMARK_);
}));

(cljs.core.async.onto_chan.cljs$lang$maxFixedArity = 3);

/**
 * Deprecated - use to-chan!
 */
cljs.core.async.to_chan = (function cljs$core$async$to_chan(coll){
return cljs.core.async.to_chan_BANG_(coll);
});

/**
 * @interface
 */
cljs.core.async.Mux = function(){};

var cljs$core$async$Mux$muxch_STAR_$dyn_34879 = (function (_){
var x__5350__auto__ = (((_ == null))?null:_);
var m__5351__auto__ = (cljs.core.async.muxch_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(_) : m__5351__auto__.call(null, _));
} else {
var m__5349__auto__ = (cljs.core.async.muxch_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(_) : m__5349__auto__.call(null, _));
} else {
throw cljs.core.missing_protocol("Mux.muxch*",_);
}
}
});
cljs.core.async.muxch_STAR_ = (function cljs$core$async$muxch_STAR_(_){
if((((!((_ == null)))) && ((!((_.cljs$core$async$Mux$muxch_STAR_$arity$1 == null)))))){
return _.cljs$core$async$Mux$muxch_STAR_$arity$1(_);
} else {
return cljs$core$async$Mux$muxch_STAR_$dyn_34879(_);
}
});


/**
 * @interface
 */
cljs.core.async.Mult = function(){};

var cljs$core$async$Mult$tap_STAR_$dyn_34883 = (function (m,ch,close_QMARK_){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.tap_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$3 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$3(m,ch,close_QMARK_) : m__5351__auto__.call(null, m,ch,close_QMARK_));
} else {
var m__5349__auto__ = (cljs.core.async.tap_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$3 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$3(m,ch,close_QMARK_) : m__5349__auto__.call(null, m,ch,close_QMARK_));
} else {
throw cljs.core.missing_protocol("Mult.tap*",m);
}
}
});
cljs.core.async.tap_STAR_ = (function cljs$core$async$tap_STAR_(m,ch,close_QMARK_){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mult$tap_STAR_$arity$3 == null)))))){
return m.cljs$core$async$Mult$tap_STAR_$arity$3(m,ch,close_QMARK_);
} else {
return cljs$core$async$Mult$tap_STAR_$dyn_34883(m,ch,close_QMARK_);
}
});

var cljs$core$async$Mult$untap_STAR_$dyn_34887 = (function (m,ch){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.untap_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5351__auto__.call(null, m,ch));
} else {
var m__5349__auto__ = (cljs.core.async.untap_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5349__auto__.call(null, m,ch));
} else {
throw cljs.core.missing_protocol("Mult.untap*",m);
}
}
});
cljs.core.async.untap_STAR_ = (function cljs$core$async$untap_STAR_(m,ch){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mult$untap_STAR_$arity$2 == null)))))){
return m.cljs$core$async$Mult$untap_STAR_$arity$2(m,ch);
} else {
return cljs$core$async$Mult$untap_STAR_$dyn_34887(m,ch);
}
});

var cljs$core$async$Mult$untap_all_STAR_$dyn_34888 = (function (m){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.untap_all_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(m) : m__5351__auto__.call(null, m));
} else {
var m__5349__auto__ = (cljs.core.async.untap_all_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(m) : m__5349__auto__.call(null, m));
} else {
throw cljs.core.missing_protocol("Mult.untap-all*",m);
}
}
});
cljs.core.async.untap_all_STAR_ = (function cljs$core$async$untap_all_STAR_(m){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mult$untap_all_STAR_$arity$1 == null)))))){
return m.cljs$core$async$Mult$untap_all_STAR_$arity$1(m);
} else {
return cljs$core$async$Mult$untap_all_STAR_$dyn_34888(m);
}
});


/**
* @constructor
 * @implements {cljs.core.async.Mult}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.async.Mux}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async32285 = (function (ch,cs,meta32286){
this.ch = ch;
this.cs = cs;
this.meta32286 = meta32286;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_32287,meta32286__$1){
var self__ = this;
var _32287__$1 = this;
return (new cljs.core.async.t_cljs$core$async32285(self__.ch,self__.cs,meta32286__$1));
}));

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_32287){
var self__ = this;
var _32287__$1 = this;
return self__.meta32286;
}));

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mux$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mux$muxch_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return self__.ch;
}));

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mult$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mult$tap_STAR_$arity$3 = (function (_,ch__$1,close_QMARK_){
var self__ = this;
var ___$1 = this;
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(self__.cs,cljs.core.assoc,ch__$1,close_QMARK_);

return null;
}));

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mult$untap_STAR_$arity$2 = (function (_,ch__$1){
var self__ = this;
var ___$1 = this;
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(self__.cs,cljs.core.dissoc,ch__$1);

return null;
}));

(cljs.core.async.t_cljs$core$async32285.prototype.cljs$core$async$Mult$untap_all_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
cljs.core.reset_BANG_(self__.cs,cljs.core.PersistentArrayMap.EMPTY);

return null;
}));

(cljs.core.async.t_cljs$core$async32285.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"cs","cs",-117024463,null),new cljs.core.Symbol(null,"meta32286","meta32286",-725120250,null)], null);
}));

(cljs.core.async.t_cljs$core$async32285.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async32285.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async32285");

(cljs.core.async.t_cljs$core$async32285.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async32285");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async32285.
 */
cljs.core.async.__GT_t_cljs$core$async32285 = (function cljs$core$async$__GT_t_cljs$core$async32285(ch,cs,meta32286){
return (new cljs.core.async.t_cljs$core$async32285(ch,cs,meta32286));
});


/**
 * Creates and returns a mult(iple) of the supplied channel. Channels
 *   containing copies of the channel can be created with 'tap', and
 *   detached with 'untap'.
 * 
 *   Each item is distributed to all taps in parallel and synchronously,
 *   i.e. each tap must accept before the next item is distributed. Use
 *   buffering/windowing to prevent slow taps from holding up the mult.
 * 
 *   Items received when there are no taps get dropped.
 * 
 *   If a tap puts to a closed channel, it will be removed from the mult.
 */
cljs.core.async.mult = (function cljs$core$async$mult(ch){
var cs = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(cljs.core.PersistentArrayMap.EMPTY);
var m = (new cljs.core.async.t_cljs$core$async32285(ch,cs,cljs.core.PersistentArrayMap.EMPTY));
var dchan = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
var dctr = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(null);
var done = (function (_){
if((cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$2(dctr,cljs.core.dec) === (0))){
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2(dchan,true);
} else {
return null;
}
});
var c__30907__auto___34896 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_32438){
var state_val_32439 = (state_32438[(1)]);
if((state_val_32439 === (7))){
var inst_32432 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32441_34897 = state_32438__$1;
(statearr_32441_34897[(2)] = inst_32432);

(statearr_32441_34897[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (20))){
var inst_32324 = (state_32438[(7)]);
var inst_32336 = cljs.core.first(inst_32324);
var inst_32337 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32336,(0),null);
var inst_32338 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32336,(1),null);
var state_32438__$1 = (function (){var statearr_32443 = state_32438;
(statearr_32443[(8)] = inst_32337);

return statearr_32443;
})();
if(cljs.core.truth_(inst_32338)){
var statearr_32447_34898 = state_32438__$1;
(statearr_32447_34898[(1)] = (22));

} else {
var statearr_32451_34899 = state_32438__$1;
(statearr_32451_34899[(1)] = (23));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (27))){
var inst_32369 = (state_32438[(9)]);
var inst_32367 = (state_32438[(10)]);
var inst_32292 = (state_32438[(11)]);
var inst_32375 = (state_32438[(12)]);
var inst_32375__$1 = cljs.core._nth(inst_32367,inst_32369);
var inst_32376 = cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$3(inst_32375__$1,inst_32292,done);
var state_32438__$1 = (function (){var statearr_32454 = state_32438;
(statearr_32454[(12)] = inst_32375__$1);

return statearr_32454;
})();
if(cljs.core.truth_(inst_32376)){
var statearr_32455_34900 = state_32438__$1;
(statearr_32455_34900[(1)] = (30));

} else {
var statearr_32456_34901 = state_32438__$1;
(statearr_32456_34901[(1)] = (31));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (1))){
var state_32438__$1 = state_32438;
var statearr_32457_34902 = state_32438__$1;
(statearr_32457_34902[(2)] = null);

(statearr_32457_34902[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (24))){
var inst_32324 = (state_32438[(7)]);
var inst_32343 = (state_32438[(2)]);
var inst_32344 = cljs.core.next(inst_32324);
var inst_32302 = inst_32344;
var inst_32303 = null;
var inst_32304 = (0);
var inst_32305 = (0);
var state_32438__$1 = (function (){var statearr_32464 = state_32438;
(statearr_32464[(13)] = inst_32343);

(statearr_32464[(14)] = inst_32303);

(statearr_32464[(15)] = inst_32302);

(statearr_32464[(16)] = inst_32304);

(statearr_32464[(17)] = inst_32305);

return statearr_32464;
})();
var statearr_32465_34906 = state_32438__$1;
(statearr_32465_34906[(2)] = null);

(statearr_32465_34906[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (39))){
var state_32438__$1 = state_32438;
var statearr_32470_34907 = state_32438__$1;
(statearr_32470_34907[(2)] = null);

(statearr_32470_34907[(1)] = (41));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (4))){
var inst_32292 = (state_32438[(11)]);
var inst_32292__$1 = (state_32438[(2)]);
var inst_32293 = (inst_32292__$1 == null);
var state_32438__$1 = (function (){var statearr_32472 = state_32438;
(statearr_32472[(11)] = inst_32292__$1);

return statearr_32472;
})();
if(cljs.core.truth_(inst_32293)){
var statearr_32473_34908 = state_32438__$1;
(statearr_32473_34908[(1)] = (5));

} else {
var statearr_32474_34909 = state_32438__$1;
(statearr_32474_34909[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (15))){
var inst_32303 = (state_32438[(14)]);
var inst_32302 = (state_32438[(15)]);
var inst_32304 = (state_32438[(16)]);
var inst_32305 = (state_32438[(17)]);
var inst_32320 = (state_32438[(2)]);
var inst_32321 = (inst_32305 + (1));
var tmp32466 = inst_32303;
var tmp32467 = inst_32302;
var tmp32468 = inst_32304;
var inst_32302__$1 = tmp32467;
var inst_32303__$1 = tmp32466;
var inst_32304__$1 = tmp32468;
var inst_32305__$1 = inst_32321;
var state_32438__$1 = (function (){var statearr_32475 = state_32438;
(statearr_32475[(18)] = inst_32320);

(statearr_32475[(14)] = inst_32303__$1);

(statearr_32475[(15)] = inst_32302__$1);

(statearr_32475[(16)] = inst_32304__$1);

(statearr_32475[(17)] = inst_32305__$1);

return statearr_32475;
})();
var statearr_32476_34914 = state_32438__$1;
(statearr_32476_34914[(2)] = null);

(statearr_32476_34914[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (21))){
var inst_32347 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32481_34916 = state_32438__$1;
(statearr_32481_34916[(2)] = inst_32347);

(statearr_32481_34916[(1)] = (18));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (31))){
var inst_32375 = (state_32438[(12)]);
var inst_32379 = m.cljs$core$async$Mult$untap_STAR_$arity$2(null, inst_32375);
var state_32438__$1 = state_32438;
var statearr_32484_34921 = state_32438__$1;
(statearr_32484_34921[(2)] = inst_32379);

(statearr_32484_34921[(1)] = (32));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (32))){
var inst_32369 = (state_32438[(9)]);
var inst_32368 = (state_32438[(19)]);
var inst_32367 = (state_32438[(10)]);
var inst_32366 = (state_32438[(20)]);
var inst_32381 = (state_32438[(2)]);
var inst_32382 = (inst_32369 + (1));
var tmp32477 = inst_32368;
var tmp32478 = inst_32367;
var tmp32479 = inst_32366;
var inst_32366__$1 = tmp32479;
var inst_32367__$1 = tmp32478;
var inst_32368__$1 = tmp32477;
var inst_32369__$1 = inst_32382;
var state_32438__$1 = (function (){var statearr_32486 = state_32438;
(statearr_32486[(9)] = inst_32369__$1);

(statearr_32486[(19)] = inst_32368__$1);

(statearr_32486[(10)] = inst_32367__$1);

(statearr_32486[(21)] = inst_32381);

(statearr_32486[(20)] = inst_32366__$1);

return statearr_32486;
})();
var statearr_32489_34944 = state_32438__$1;
(statearr_32489_34944[(2)] = null);

(statearr_32489_34944[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (40))){
var inst_32403 = (state_32438[(22)]);
var inst_32409 = m.cljs$core$async$Mult$untap_STAR_$arity$2(null, inst_32403);
var state_32438__$1 = state_32438;
var statearr_32490_34960 = state_32438__$1;
(statearr_32490_34960[(2)] = inst_32409);

(statearr_32490_34960[(1)] = (41));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (33))){
var inst_32385 = (state_32438[(23)]);
var inst_32387 = cljs.core.chunked_seq_QMARK_(inst_32385);
var state_32438__$1 = state_32438;
if(inst_32387){
var statearr_32491_34964 = state_32438__$1;
(statearr_32491_34964[(1)] = (36));

} else {
var statearr_32492_34965 = state_32438__$1;
(statearr_32492_34965[(1)] = (37));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (13))){
var inst_32314 = (state_32438[(24)]);
var inst_32317 = cljs.core.async.close_BANG_(inst_32314);
var state_32438__$1 = state_32438;
var statearr_32495_34969 = state_32438__$1;
(statearr_32495_34969[(2)] = inst_32317);

(statearr_32495_34969[(1)] = (15));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (22))){
var inst_32337 = (state_32438[(8)]);
var inst_32340 = cljs.core.async.close_BANG_(inst_32337);
var state_32438__$1 = state_32438;
var statearr_32498_34970 = state_32438__$1;
(statearr_32498_34970[(2)] = inst_32340);

(statearr_32498_34970[(1)] = (24));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (36))){
var inst_32385 = (state_32438[(23)]);
var inst_32395 = cljs.core.chunk_first(inst_32385);
var inst_32399 = cljs.core.chunk_rest(inst_32385);
var inst_32400 = cljs.core.count(inst_32395);
var inst_32366 = inst_32399;
var inst_32367 = inst_32395;
var inst_32368 = inst_32400;
var inst_32369 = (0);
var state_32438__$1 = (function (){var statearr_32504 = state_32438;
(statearr_32504[(9)] = inst_32369);

(statearr_32504[(19)] = inst_32368);

(statearr_32504[(10)] = inst_32367);

(statearr_32504[(20)] = inst_32366);

return statearr_32504;
})();
var statearr_32506_34977 = state_32438__$1;
(statearr_32506_34977[(2)] = null);

(statearr_32506_34977[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (41))){
var inst_32385 = (state_32438[(23)]);
var inst_32411 = (state_32438[(2)]);
var inst_32412 = cljs.core.next(inst_32385);
var inst_32366 = inst_32412;
var inst_32367 = null;
var inst_32368 = (0);
var inst_32369 = (0);
var state_32438__$1 = (function (){var statearr_32510 = state_32438;
(statearr_32510[(9)] = inst_32369);

(statearr_32510[(25)] = inst_32411);

(statearr_32510[(19)] = inst_32368);

(statearr_32510[(10)] = inst_32367);

(statearr_32510[(20)] = inst_32366);

return statearr_32510;
})();
var statearr_32511_34978 = state_32438__$1;
(statearr_32511_34978[(2)] = null);

(statearr_32511_34978[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (43))){
var state_32438__$1 = state_32438;
var statearr_32513_34979 = state_32438__$1;
(statearr_32513_34979[(2)] = null);

(statearr_32513_34979[(1)] = (44));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (29))){
var inst_32420 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32515_34980 = state_32438__$1;
(statearr_32515_34980[(2)] = inst_32420);

(statearr_32515_34980[(1)] = (26));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (44))){
var inst_32429 = (state_32438[(2)]);
var state_32438__$1 = (function (){var statearr_32517 = state_32438;
(statearr_32517[(26)] = inst_32429);

return statearr_32517;
})();
var statearr_32518_34981 = state_32438__$1;
(statearr_32518_34981[(2)] = null);

(statearr_32518_34981[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (6))){
var inst_32358 = (state_32438[(27)]);
var inst_32357 = cljs.core.deref(cs);
var inst_32358__$1 = cljs.core.keys(inst_32357);
var inst_32359 = cljs.core.count(inst_32358__$1);
var inst_32360 = cljs.core.reset_BANG_(dctr,inst_32359);
var inst_32365 = cljs.core.seq(inst_32358__$1);
var inst_32366 = inst_32365;
var inst_32367 = null;
var inst_32368 = (0);
var inst_32369 = (0);
var state_32438__$1 = (function (){var statearr_32519 = state_32438;
(statearr_32519[(9)] = inst_32369);

(statearr_32519[(19)] = inst_32368);

(statearr_32519[(27)] = inst_32358__$1);

(statearr_32519[(10)] = inst_32367);

(statearr_32519[(28)] = inst_32360);

(statearr_32519[(20)] = inst_32366);

return statearr_32519;
})();
var statearr_32520_34982 = state_32438__$1;
(statearr_32520_34982[(2)] = null);

(statearr_32520_34982[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (28))){
var inst_32385 = (state_32438[(23)]);
var inst_32366 = (state_32438[(20)]);
var inst_32385__$1 = cljs.core.seq(inst_32366);
var state_32438__$1 = (function (){var statearr_32521 = state_32438;
(statearr_32521[(23)] = inst_32385__$1);

return statearr_32521;
})();
if(inst_32385__$1){
var statearr_32522_34983 = state_32438__$1;
(statearr_32522_34983[(1)] = (33));

} else {
var statearr_32523_34984 = state_32438__$1;
(statearr_32523_34984[(1)] = (34));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (25))){
var inst_32369 = (state_32438[(9)]);
var inst_32368 = (state_32438[(19)]);
var inst_32372 = (inst_32369 < inst_32368);
var inst_32373 = inst_32372;
var state_32438__$1 = state_32438;
if(cljs.core.truth_(inst_32373)){
var statearr_32531_34988 = state_32438__$1;
(statearr_32531_34988[(1)] = (27));

} else {
var statearr_32532_34989 = state_32438__$1;
(statearr_32532_34989[(1)] = (28));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (34))){
var state_32438__$1 = state_32438;
var statearr_32535_34990 = state_32438__$1;
(statearr_32535_34990[(2)] = null);

(statearr_32535_34990[(1)] = (35));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (17))){
var state_32438__$1 = state_32438;
var statearr_32536_34991 = state_32438__$1;
(statearr_32536_34991[(2)] = null);

(statearr_32536_34991[(1)] = (18));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (3))){
var inst_32434 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
return cljs.core.async.impl.ioc_helpers.return_chan(state_32438__$1,inst_32434);
} else {
if((state_val_32439 === (12))){
var inst_32352 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32552_34992 = state_32438__$1;
(statearr_32552_34992[(2)] = inst_32352);

(statearr_32552_34992[(1)] = (9));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (2))){
var state_32438__$1 = state_32438;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_32438__$1,(4),ch);
} else {
if((state_val_32439 === (23))){
var state_32438__$1 = state_32438;
var statearr_32564_34993 = state_32438__$1;
(statearr_32564_34993[(2)] = null);

(statearr_32564_34993[(1)] = (24));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (35))){
var inst_32418 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32572_34997 = state_32438__$1;
(statearr_32572_34997[(2)] = inst_32418);

(statearr_32572_34997[(1)] = (29));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (19))){
var inst_32324 = (state_32438[(7)]);
var inst_32328 = cljs.core.chunk_first(inst_32324);
var inst_32329 = cljs.core.chunk_rest(inst_32324);
var inst_32330 = cljs.core.count(inst_32328);
var inst_32302 = inst_32329;
var inst_32303 = inst_32328;
var inst_32304 = inst_32330;
var inst_32305 = (0);
var state_32438__$1 = (function (){var statearr_32577 = state_32438;
(statearr_32577[(14)] = inst_32303);

(statearr_32577[(15)] = inst_32302);

(statearr_32577[(16)] = inst_32304);

(statearr_32577[(17)] = inst_32305);

return statearr_32577;
})();
var statearr_32579_34998 = state_32438__$1;
(statearr_32579_34998[(2)] = null);

(statearr_32579_34998[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (11))){
var inst_32324 = (state_32438[(7)]);
var inst_32302 = (state_32438[(15)]);
var inst_32324__$1 = cljs.core.seq(inst_32302);
var state_32438__$1 = (function (){var statearr_32583 = state_32438;
(statearr_32583[(7)] = inst_32324__$1);

return statearr_32583;
})();
if(inst_32324__$1){
var statearr_32587_34999 = state_32438__$1;
(statearr_32587_34999[(1)] = (16));

} else {
var statearr_32589_35000 = state_32438__$1;
(statearr_32589_35000[(1)] = (17));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (9))){
var inst_32354 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32594_35001 = state_32438__$1;
(statearr_32594_35001[(2)] = inst_32354);

(statearr_32594_35001[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (5))){
var inst_32300 = cljs.core.deref(cs);
var inst_32301 = cljs.core.seq(inst_32300);
var inst_32302 = inst_32301;
var inst_32303 = null;
var inst_32304 = (0);
var inst_32305 = (0);
var state_32438__$1 = (function (){var statearr_32596 = state_32438;
(statearr_32596[(14)] = inst_32303);

(statearr_32596[(15)] = inst_32302);

(statearr_32596[(16)] = inst_32304);

(statearr_32596[(17)] = inst_32305);

return statearr_32596;
})();
var statearr_32599_35002 = state_32438__$1;
(statearr_32599_35002[(2)] = null);

(statearr_32599_35002[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (14))){
var state_32438__$1 = state_32438;
var statearr_32602_35003 = state_32438__$1;
(statearr_32602_35003[(2)] = null);

(statearr_32602_35003[(1)] = (15));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (45))){
var inst_32426 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32608_35004 = state_32438__$1;
(statearr_32608_35004[(2)] = inst_32426);

(statearr_32608_35004[(1)] = (44));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (26))){
var inst_32358 = (state_32438[(27)]);
var inst_32422 = (state_32438[(2)]);
var inst_32423 = cljs.core.seq(inst_32358);
var state_32438__$1 = (function (){var statearr_32610 = state_32438;
(statearr_32610[(29)] = inst_32422);

return statearr_32610;
})();
if(inst_32423){
var statearr_32611_35005 = state_32438__$1;
(statearr_32611_35005[(1)] = (42));

} else {
var statearr_32612_35006 = state_32438__$1;
(statearr_32612_35006[(1)] = (43));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (16))){
var inst_32324 = (state_32438[(7)]);
var inst_32326 = cljs.core.chunked_seq_QMARK_(inst_32324);
var state_32438__$1 = state_32438;
if(inst_32326){
var statearr_32617_35008 = state_32438__$1;
(statearr_32617_35008[(1)] = (19));

} else {
var statearr_32619_35009 = state_32438__$1;
(statearr_32619_35009[(1)] = (20));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (38))){
var inst_32415 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32620_35010 = state_32438__$1;
(statearr_32620_35010[(2)] = inst_32415);

(statearr_32620_35010[(1)] = (35));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (30))){
var state_32438__$1 = state_32438;
var statearr_32624_35012 = state_32438__$1;
(statearr_32624_35012[(2)] = null);

(statearr_32624_35012[(1)] = (32));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (10))){
var inst_32303 = (state_32438[(14)]);
var inst_32305 = (state_32438[(17)]);
var inst_32313 = cljs.core._nth(inst_32303,inst_32305);
var inst_32314 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32313,(0),null);
var inst_32315 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32313,(1),null);
var state_32438__$1 = (function (){var statearr_32633 = state_32438;
(statearr_32633[(24)] = inst_32314);

return statearr_32633;
})();
if(cljs.core.truth_(inst_32315)){
var statearr_32636_35016 = state_32438__$1;
(statearr_32636_35016[(1)] = (13));

} else {
var statearr_32637_35017 = state_32438__$1;
(statearr_32637_35017[(1)] = (14));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (18))){
var inst_32350 = (state_32438[(2)]);
var state_32438__$1 = state_32438;
var statearr_32638_35018 = state_32438__$1;
(statearr_32638_35018[(2)] = inst_32350);

(statearr_32638_35018[(1)] = (12));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (42))){
var state_32438__$1 = state_32438;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_32438__$1,(45),dchan);
} else {
if((state_val_32439 === (37))){
var inst_32385 = (state_32438[(23)]);
var inst_32292 = (state_32438[(11)]);
var inst_32403 = (state_32438[(22)]);
var inst_32403__$1 = cljs.core.first(inst_32385);
var inst_32404 = cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$3(inst_32403__$1,inst_32292,done);
var state_32438__$1 = (function (){var statearr_32642 = state_32438;
(statearr_32642[(22)] = inst_32403__$1);

return statearr_32642;
})();
if(cljs.core.truth_(inst_32404)){
var statearr_32644_35019 = state_32438__$1;
(statearr_32644_35019[(1)] = (39));

} else {
var statearr_32645_35021 = state_32438__$1;
(statearr_32645_35021[(1)] = (40));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32439 === (8))){
var inst_32304 = (state_32438[(16)]);
var inst_32305 = (state_32438[(17)]);
var inst_32307 = (inst_32305 < inst_32304);
var inst_32308 = inst_32307;
var state_32438__$1 = state_32438;
if(cljs.core.truth_(inst_32308)){
var statearr_32647_35026 = state_32438__$1;
(statearr_32647_35026[(1)] = (10));

} else {
var statearr_32649_35027 = state_32438__$1;
(statearr_32649_35027[(1)] = (11));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$mult_$_state_machine__30544__auto__ = null;
var cljs$core$async$mult_$_state_machine__30544__auto____0 = (function (){
var statearr_32656 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_32656[(0)] = cljs$core$async$mult_$_state_machine__30544__auto__);

(statearr_32656[(1)] = (1));

return statearr_32656;
});
var cljs$core$async$mult_$_state_machine__30544__auto____1 = (function (state_32438){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_32438);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e32658){var ex__30547__auto__ = e32658;
var statearr_32660_35028 = state_32438;
(statearr_32660_35028[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_32438[(4)]))){
var statearr_32661_35029 = state_32438;
(statearr_32661_35029[(1)] = cljs.core.first((state_32438[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35031 = state_32438;
state_32438 = G__35031;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$mult_$_state_machine__30544__auto__ = function(state_32438){
switch(arguments.length){
case 0:
return cljs$core$async$mult_$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$mult_$_state_machine__30544__auto____1.call(this,state_32438);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$mult_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$mult_$_state_machine__30544__auto____0;
cljs$core$async$mult_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$mult_$_state_machine__30544__auto____1;
return cljs$core$async$mult_$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_32666 = f__30908__auto__();
(statearr_32666[(6)] = c__30907__auto___34896);

return statearr_32666;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return m;
});
/**
 * Copies the mult source onto the supplied channel.
 * 
 *   By default the channel will be closed when the source closes,
 *   but can be determined by the close? parameter.
 */
cljs.core.async.tap = (function cljs$core$async$tap(var_args){
var G__32674 = arguments.length;
switch (G__32674) {
case 2:
return cljs.core.async.tap.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.tap.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.tap.cljs$core$IFn$_invoke$arity$2 = (function (mult,ch){
return cljs.core.async.tap.cljs$core$IFn$_invoke$arity$3(mult,ch,true);
}));

(cljs.core.async.tap.cljs$core$IFn$_invoke$arity$3 = (function (mult,ch,close_QMARK_){
cljs.core.async.tap_STAR_(mult,ch,close_QMARK_);

return ch;
}));

(cljs.core.async.tap.cljs$lang$maxFixedArity = 3);

/**
 * Disconnects a target channel from a mult
 */
cljs.core.async.untap = (function cljs$core$async$untap(mult,ch){
return cljs.core.async.untap_STAR_(mult,ch);
});
/**
 * Disconnects all target channels from a mult
 */
cljs.core.async.untap_all = (function cljs$core$async$untap_all(mult){
return cljs.core.async.untap_all_STAR_(mult);
});

/**
 * @interface
 */
cljs.core.async.Mix = function(){};

var cljs$core$async$Mix$admix_STAR_$dyn_35034 = (function (m,ch){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.admix_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5351__auto__.call(null, m,ch));
} else {
var m__5349__auto__ = (cljs.core.async.admix_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5349__auto__.call(null, m,ch));
} else {
throw cljs.core.missing_protocol("Mix.admix*",m);
}
}
});
cljs.core.async.admix_STAR_ = (function cljs$core$async$admix_STAR_(m,ch){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mix$admix_STAR_$arity$2 == null)))))){
return m.cljs$core$async$Mix$admix_STAR_$arity$2(m,ch);
} else {
return cljs$core$async$Mix$admix_STAR_$dyn_35034(m,ch);
}
});

var cljs$core$async$Mix$unmix_STAR_$dyn_35036 = (function (m,ch){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.unmix_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5351__auto__.call(null, m,ch));
} else {
var m__5349__auto__ = (cljs.core.async.unmix_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(m,ch) : m__5349__auto__.call(null, m,ch));
} else {
throw cljs.core.missing_protocol("Mix.unmix*",m);
}
}
});
cljs.core.async.unmix_STAR_ = (function cljs$core$async$unmix_STAR_(m,ch){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mix$unmix_STAR_$arity$2 == null)))))){
return m.cljs$core$async$Mix$unmix_STAR_$arity$2(m,ch);
} else {
return cljs$core$async$Mix$unmix_STAR_$dyn_35036(m,ch);
}
});

var cljs$core$async$Mix$unmix_all_STAR_$dyn_35038 = (function (m){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.unmix_all_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(m) : m__5351__auto__.call(null, m));
} else {
var m__5349__auto__ = (cljs.core.async.unmix_all_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(m) : m__5349__auto__.call(null, m));
} else {
throw cljs.core.missing_protocol("Mix.unmix-all*",m);
}
}
});
cljs.core.async.unmix_all_STAR_ = (function cljs$core$async$unmix_all_STAR_(m){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mix$unmix_all_STAR_$arity$1 == null)))))){
return m.cljs$core$async$Mix$unmix_all_STAR_$arity$1(m);
} else {
return cljs$core$async$Mix$unmix_all_STAR_$dyn_35038(m);
}
});

var cljs$core$async$Mix$toggle_STAR_$dyn_35041 = (function (m,state_map){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.toggle_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(m,state_map) : m__5351__auto__.call(null, m,state_map));
} else {
var m__5349__auto__ = (cljs.core.async.toggle_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(m,state_map) : m__5349__auto__.call(null, m,state_map));
} else {
throw cljs.core.missing_protocol("Mix.toggle*",m);
}
}
});
cljs.core.async.toggle_STAR_ = (function cljs$core$async$toggle_STAR_(m,state_map){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mix$toggle_STAR_$arity$2 == null)))))){
return m.cljs$core$async$Mix$toggle_STAR_$arity$2(m,state_map);
} else {
return cljs$core$async$Mix$toggle_STAR_$dyn_35041(m,state_map);
}
});

var cljs$core$async$Mix$solo_mode_STAR_$dyn_35043 = (function (m,mode){
var x__5350__auto__ = (((m == null))?null:m);
var m__5351__auto__ = (cljs.core.async.solo_mode_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(m,mode) : m__5351__auto__.call(null, m,mode));
} else {
var m__5349__auto__ = (cljs.core.async.solo_mode_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(m,mode) : m__5349__auto__.call(null, m,mode));
} else {
throw cljs.core.missing_protocol("Mix.solo-mode*",m);
}
}
});
cljs.core.async.solo_mode_STAR_ = (function cljs$core$async$solo_mode_STAR_(m,mode){
if((((!((m == null)))) && ((!((m.cljs$core$async$Mix$solo_mode_STAR_$arity$2 == null)))))){
return m.cljs$core$async$Mix$solo_mode_STAR_$arity$2(m,mode);
} else {
return cljs$core$async$Mix$solo_mode_STAR_$dyn_35043(m,mode);
}
});

cljs.core.async.ioc_alts_BANG_ = (function cljs$core$async$ioc_alts_BANG_(var_args){
var args__5732__auto__ = [];
var len__5726__auto___35047 = arguments.length;
var i__5727__auto___35048 = (0);
while(true){
if((i__5727__auto___35048 < len__5726__auto___35047)){
args__5732__auto__.push((arguments[i__5727__auto___35048]));

var G__35049 = (i__5727__auto___35048 + (1));
i__5727__auto___35048 = G__35049;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((3) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((3)),(0),null)):null);
return cljs.core.async.ioc_alts_BANG_.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),argseq__5733__auto__);
});

(cljs.core.async.ioc_alts_BANG_.cljs$core$IFn$_invoke$arity$variadic = (function (state,cont_block,ports,p__32757){
var map__32758 = p__32757;
var map__32758__$1 = cljs.core.__destructure_map(map__32758);
var opts = map__32758__$1;
var statearr_32760_35051 = state;
(statearr_32760_35051[(1)] = cont_block);


var temp__5825__auto__ = cljs.core.async.do_alts((function (val){
var statearr_32763_35052 = state;
(statearr_32763_35052[(2)] = val);


return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state);
}),ports,opts);
if(cljs.core.truth_(temp__5825__auto__)){
var cb = temp__5825__auto__;
var statearr_32765_35053 = state;
(statearr_32765_35053[(2)] = cljs.core.deref(cb));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}));

(cljs.core.async.ioc_alts_BANG_.cljs$lang$maxFixedArity = (3));

/** @this {Function} */
(cljs.core.async.ioc_alts_BANG_.cljs$lang$applyTo = (function (seq32743){
var G__32744 = cljs.core.first(seq32743);
var seq32743__$1 = cljs.core.next(seq32743);
var G__32745 = cljs.core.first(seq32743__$1);
var seq32743__$2 = cljs.core.next(seq32743__$1);
var G__32746 = cljs.core.first(seq32743__$2);
var seq32743__$3 = cljs.core.next(seq32743__$2);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__32744,G__32745,G__32746,seq32743__$3);
}));


/**
* @constructor
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.async.Mix}
 * @implements {cljs.core.async.Mux}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async32793 = (function (change,solo_mode,pick,cs,calc_state,out,changed,solo_modes,attrs,meta32794){
this.change = change;
this.solo_mode = solo_mode;
this.pick = pick;
this.cs = cs;
this.calc_state = calc_state;
this.out = out;
this.changed = changed;
this.solo_modes = solo_modes;
this.attrs = attrs;
this.meta32794 = meta32794;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_32795,meta32794__$1){
var self__ = this;
var _32795__$1 = this;
return (new cljs.core.async.t_cljs$core$async32793(self__.change,self__.solo_mode,self__.pick,self__.cs,self__.calc_state,self__.out,self__.changed,self__.solo_modes,self__.attrs,meta32794__$1));
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_32795){
var self__ = this;
var _32795__$1 = this;
return self__.meta32794;
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mux$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mux$muxch_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return self__.out;
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$admix_STAR_$arity$2 = (function (_,ch){
var self__ = this;
var ___$1 = this;
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(self__.cs,cljs.core.assoc,ch,cljs.core.PersistentArrayMap.EMPTY);

return (self__.changed.cljs$core$IFn$_invoke$arity$0 ? self__.changed.cljs$core$IFn$_invoke$arity$0() : self__.changed.call(null, ));
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$unmix_STAR_$arity$2 = (function (_,ch){
var self__ = this;
var ___$1 = this;
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(self__.cs,cljs.core.dissoc,ch);

return (self__.changed.cljs$core$IFn$_invoke$arity$0 ? self__.changed.cljs$core$IFn$_invoke$arity$0() : self__.changed.call(null, ));
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$unmix_all_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
cljs.core.reset_BANG_(self__.cs,cljs.core.PersistentArrayMap.EMPTY);

return (self__.changed.cljs$core$IFn$_invoke$arity$0 ? self__.changed.cljs$core$IFn$_invoke$arity$0() : self__.changed.call(null, ));
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$toggle_STAR_$arity$2 = (function (_,state_map){
var self__ = this;
var ___$1 = this;
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(self__.cs,cljs.core.partial.cljs$core$IFn$_invoke$arity$2(cljs.core.merge_with,cljs.core.merge),state_map);

return (self__.changed.cljs$core$IFn$_invoke$arity$0 ? self__.changed.cljs$core$IFn$_invoke$arity$0() : self__.changed.call(null, ));
}));

(cljs.core.async.t_cljs$core$async32793.prototype.cljs$core$async$Mix$solo_mode_STAR_$arity$2 = (function (_,mode){
var self__ = this;
var ___$1 = this;
if(cljs.core.truth_((self__.solo_modes.cljs$core$IFn$_invoke$arity$1 ? self__.solo_modes.cljs$core$IFn$_invoke$arity$1(mode) : self__.solo_modes.call(null, mode)))){
} else {
throw (new Error(["Assert failed: ",["mode must be one of: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(self__.solo_modes)].join(''),"\n","(solo-modes mode)"].join('')));
}

cljs.core.reset_BANG_(self__.solo_mode,mode);

return (self__.changed.cljs$core$IFn$_invoke$arity$0 ? self__.changed.cljs$core$IFn$_invoke$arity$0() : self__.changed.call(null, ));
}));

(cljs.core.async.t_cljs$core$async32793.getBasis = (function (){
return new cljs.core.PersistentVector(null, 10, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"change","change",477485025,null),new cljs.core.Symbol(null,"solo-mode","solo-mode",2031788074,null),new cljs.core.Symbol(null,"pick","pick",1300068175,null),new cljs.core.Symbol(null,"cs","cs",-117024463,null),new cljs.core.Symbol(null,"calc-state","calc-state",-349968968,null),new cljs.core.Symbol(null,"out","out",729986010,null),new cljs.core.Symbol(null,"changed","changed",-2083710852,null),new cljs.core.Symbol(null,"solo-modes","solo-modes",882180540,null),new cljs.core.Symbol(null,"attrs","attrs",-450137186,null),new cljs.core.Symbol(null,"meta32794","meta32794",1718518648,null)], null);
}));

(cljs.core.async.t_cljs$core$async32793.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async32793.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async32793");

(cljs.core.async.t_cljs$core$async32793.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async32793");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async32793.
 */
cljs.core.async.__GT_t_cljs$core$async32793 = (function cljs$core$async$__GT_t_cljs$core$async32793(change,solo_mode,pick,cs,calc_state,out,changed,solo_modes,attrs,meta32794){
return (new cljs.core.async.t_cljs$core$async32793(change,solo_mode,pick,cs,calc_state,out,changed,solo_modes,attrs,meta32794));
});


/**
 * Creates and returns a mix of one or more input channels which will
 *   be put on the supplied out channel. Input sources can be added to
 *   the mix with 'admix', and removed with 'unmix'. A mix supports
 *   soloing, muting and pausing multiple inputs atomically using
 *   'toggle', and can solo using either muting or pausing as determined
 *   by 'solo-mode'.
 * 
 *   Each channel can have zero or more boolean modes set via 'toggle':
 * 
 *   :solo - when true, only this (ond other soloed) channel(s) will appear
 *        in the mix output channel. :mute and :pause states of soloed
 *        channels are ignored. If solo-mode is :mute, non-soloed
 *        channels are muted, if :pause, non-soloed channels are
 *        paused.
 * 
 *   :mute - muted channels will have their contents consumed but not included in the mix
 *   :pause - paused channels will not have their contents consumed (and thus also not included in the mix)
 */
cljs.core.async.mix = (function cljs$core$async$mix(out){
var cs = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(cljs.core.PersistentArrayMap.EMPTY);
var solo_modes = new cljs.core.PersistentHashSet(null, new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"pause","pause",-2095325672),null,new cljs.core.Keyword(null,"mute","mute",1151223646),null], null), null);
var attrs = cljs.core.conj.cljs$core$IFn$_invoke$arity$2(solo_modes,new cljs.core.Keyword(null,"solo","solo",-316350075));
var solo_mode = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"mute","mute",1151223646));
var change = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(cljs.core.async.sliding_buffer((1)));
var changed = (function (){
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2(change,true);
});
var pick = (function (attr,chs){
return cljs.core.reduce_kv((function (ret,c,v){
if(cljs.core.truth_((attr.cljs$core$IFn$_invoke$arity$1 ? attr.cljs$core$IFn$_invoke$arity$1(v) : attr.call(null, v)))){
return cljs.core.conj.cljs$core$IFn$_invoke$arity$2(ret,c);
} else {
return ret;
}
}),cljs.core.PersistentHashSet.EMPTY,chs);
});
var calc_state = (function (){
var chs = cljs.core.deref(cs);
var mode = cljs.core.deref(solo_mode);
var solos = pick(new cljs.core.Keyword(null,"solo","solo",-316350075),chs);
var pauses = pick(new cljs.core.Keyword(null,"pause","pause",-2095325672),chs);
return new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"solos","solos",1441458643),solos,new cljs.core.Keyword(null,"mutes","mutes",1068806309),pick(new cljs.core.Keyword(null,"mute","mute",1151223646),chs),new cljs.core.Keyword(null,"reads","reads",-1215067361),cljs.core.conj.cljs$core$IFn$_invoke$arity$2(((((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(mode,new cljs.core.Keyword(null,"pause","pause",-2095325672))) && ((!(cljs.core.empty_QMARK_(solos))))))?cljs.core.vec(solos):cljs.core.vec(cljs.core.remove.cljs$core$IFn$_invoke$arity$2(pauses,cljs.core.keys(chs)))),change)], null);
});
var m = (new cljs.core.async.t_cljs$core$async32793(change,solo_mode,pick,cs,calc_state,out,changed,solo_modes,attrs,cljs.core.PersistentArrayMap.EMPTY));
var c__30907__auto___35065 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_32924){
var state_val_32925 = (state_32924[(1)]);
if((state_val_32925 === (7))){
var inst_32865 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
if(cljs.core.truth_(inst_32865)){
var statearr_32930_35068 = state_32924__$1;
(statearr_32930_35068[(1)] = (8));

} else {
var statearr_32931_35069 = state_32924__$1;
(statearr_32931_35069[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (20))){
var inst_32851 = (state_32924[(7)]);
var state_32924__$1 = state_32924;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_32924__$1,(23),out,inst_32851);
} else {
if((state_val_32925 === (1))){
var inst_32832 = calc_state();
var inst_32833 = cljs.core.__destructure_map(inst_32832);
var inst_32835 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32833,new cljs.core.Keyword(null,"solos","solos",1441458643));
var inst_32836 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32833,new cljs.core.Keyword(null,"mutes","mutes",1068806309));
var inst_32837 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32833,new cljs.core.Keyword(null,"reads","reads",-1215067361));
var inst_32838 = inst_32832;
var state_32924__$1 = (function (){var statearr_32932 = state_32924;
(statearr_32932[(8)] = inst_32836);

(statearr_32932[(9)] = inst_32835);

(statearr_32932[(10)] = inst_32838);

(statearr_32932[(11)] = inst_32837);

return statearr_32932;
})();
var statearr_32933_35070 = state_32924__$1;
(statearr_32933_35070[(2)] = null);

(statearr_32933_35070[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (24))){
var inst_32841 = (state_32924[(12)]);
var inst_32838 = inst_32841;
var state_32924__$1 = (function (){var statearr_32935 = state_32924;
(statearr_32935[(10)] = inst_32838);

return statearr_32935;
})();
var statearr_32936_35071 = state_32924__$1;
(statearr_32936_35071[(2)] = null);

(statearr_32936_35071[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (4))){
var inst_32859 = (state_32924[(13)]);
var inst_32851 = (state_32924[(7)]);
var inst_32849 = (state_32924[(2)]);
var inst_32851__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32849,(0),null);
var inst_32852 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_32849,(1),null);
var inst_32859__$1 = (inst_32851__$1 == null);
var state_32924__$1 = (function (){var statearr_32954 = state_32924;
(statearr_32954[(13)] = inst_32859__$1);

(statearr_32954[(7)] = inst_32851__$1);

(statearr_32954[(14)] = inst_32852);

return statearr_32954;
})();
if(cljs.core.truth_(inst_32859__$1)){
var statearr_32955_35075 = state_32924__$1;
(statearr_32955_35075[(1)] = (5));

} else {
var statearr_32956_35076 = state_32924__$1;
(statearr_32956_35076[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (15))){
var inst_32842 = (state_32924[(15)]);
var inst_32887 = (state_32924[(16)]);
var inst_32887__$1 = cljs.core.empty_QMARK_(inst_32842);
var state_32924__$1 = (function (){var statearr_32958 = state_32924;
(statearr_32958[(16)] = inst_32887__$1);

return statearr_32958;
})();
if(inst_32887__$1){
var statearr_32959_35077 = state_32924__$1;
(statearr_32959_35077[(1)] = (17));

} else {
var statearr_32960_35078 = state_32924__$1;
(statearr_32960_35078[(1)] = (18));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (21))){
var inst_32841 = (state_32924[(12)]);
var inst_32838 = inst_32841;
var state_32924__$1 = (function (){var statearr_32961 = state_32924;
(statearr_32961[(10)] = inst_32838);

return statearr_32961;
})();
var statearr_32966_35079 = state_32924__$1;
(statearr_32966_35079[(2)] = null);

(statearr_32966_35079[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (13))){
var inst_32872 = (state_32924[(2)]);
var inst_32873 = calc_state();
var inst_32838 = inst_32873;
var state_32924__$1 = (function (){var statearr_32970 = state_32924;
(statearr_32970[(10)] = inst_32838);

(statearr_32970[(17)] = inst_32872);

return statearr_32970;
})();
var statearr_32971_35080 = state_32924__$1;
(statearr_32971_35080[(2)] = null);

(statearr_32971_35080[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (22))){
var inst_32914 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
var statearr_32973_35081 = state_32924__$1;
(statearr_32973_35081[(2)] = inst_32914);

(statearr_32973_35081[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (6))){
var inst_32852 = (state_32924[(14)]);
var inst_32863 = cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(inst_32852,change);
var state_32924__$1 = state_32924;
var statearr_32974_35082 = state_32924__$1;
(statearr_32974_35082[(2)] = inst_32863);

(statearr_32974_35082[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (25))){
var state_32924__$1 = state_32924;
var statearr_32976_35083 = state_32924__$1;
(statearr_32976_35083[(2)] = null);

(statearr_32976_35083[(1)] = (26));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (17))){
var inst_32843 = (state_32924[(18)]);
var inst_32852 = (state_32924[(14)]);
var inst_32889 = (inst_32843.cljs$core$IFn$_invoke$arity$1 ? inst_32843.cljs$core$IFn$_invoke$arity$1(inst_32852) : inst_32843.call(null, inst_32852));
var inst_32890 = cljs.core.not(inst_32889);
var state_32924__$1 = state_32924;
var statearr_32977_35084 = state_32924__$1;
(statearr_32977_35084[(2)] = inst_32890);

(statearr_32977_35084[(1)] = (19));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (3))){
var inst_32918 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
return cljs.core.async.impl.ioc_helpers.return_chan(state_32924__$1,inst_32918);
} else {
if((state_val_32925 === (12))){
var state_32924__$1 = state_32924;
var statearr_32978_35085 = state_32924__$1;
(statearr_32978_35085[(2)] = null);

(statearr_32978_35085[(1)] = (13));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (2))){
var inst_32838 = (state_32924[(10)]);
var inst_32841 = (state_32924[(12)]);
var inst_32841__$1 = cljs.core.__destructure_map(inst_32838);
var inst_32842 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32841__$1,new cljs.core.Keyword(null,"solos","solos",1441458643));
var inst_32843 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32841__$1,new cljs.core.Keyword(null,"mutes","mutes",1068806309));
var inst_32844 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_32841__$1,new cljs.core.Keyword(null,"reads","reads",-1215067361));
var state_32924__$1 = (function (){var statearr_32980 = state_32924;
(statearr_32980[(18)] = inst_32843);

(statearr_32980[(15)] = inst_32842);

(statearr_32980[(12)] = inst_32841__$1);

return statearr_32980;
})();
return cljs.core.async.ioc_alts_BANG_(state_32924__$1,(4),inst_32844);
} else {
if((state_val_32925 === (23))){
var inst_32898 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
if(cljs.core.truth_(inst_32898)){
var statearr_32981_35086 = state_32924__$1;
(statearr_32981_35086[(1)] = (24));

} else {
var statearr_32982_35087 = state_32924__$1;
(statearr_32982_35087[(1)] = (25));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (19))){
var inst_32893 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
var statearr_32987_35088 = state_32924__$1;
(statearr_32987_35088[(2)] = inst_32893);

(statearr_32987_35088[(1)] = (16));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (11))){
var inst_32852 = (state_32924[(14)]);
var inst_32869 = cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(cs,cljs.core.dissoc,inst_32852);
var state_32924__$1 = state_32924;
var statearr_32988_35089 = state_32924__$1;
(statearr_32988_35089[(2)] = inst_32869);

(statearr_32988_35089[(1)] = (13));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (9))){
var inst_32842 = (state_32924[(15)]);
var inst_32852 = (state_32924[(14)]);
var inst_32880 = (state_32924[(19)]);
var inst_32880__$1 = (inst_32842.cljs$core$IFn$_invoke$arity$1 ? inst_32842.cljs$core$IFn$_invoke$arity$1(inst_32852) : inst_32842.call(null, inst_32852));
var state_32924__$1 = (function (){var statearr_32999 = state_32924;
(statearr_32999[(19)] = inst_32880__$1);

return statearr_32999;
})();
if(cljs.core.truth_(inst_32880__$1)){
var statearr_33005_35094 = state_32924__$1;
(statearr_33005_35094[(1)] = (14));

} else {
var statearr_33006_35095 = state_32924__$1;
(statearr_33006_35095[(1)] = (15));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (5))){
var inst_32859 = (state_32924[(13)]);
var state_32924__$1 = state_32924;
var statearr_33008_35099 = state_32924__$1;
(statearr_33008_35099[(2)] = inst_32859);

(statearr_33008_35099[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (14))){
var inst_32880 = (state_32924[(19)]);
var state_32924__$1 = state_32924;
var statearr_33009_35100 = state_32924__$1;
(statearr_33009_35100[(2)] = inst_32880);

(statearr_33009_35100[(1)] = (16));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (26))){
var inst_32906 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
var statearr_33010_35101 = state_32924__$1;
(statearr_33010_35101[(2)] = inst_32906);

(statearr_33010_35101[(1)] = (22));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (16))){
var inst_32895 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
if(cljs.core.truth_(inst_32895)){
var statearr_33012_35102 = state_32924__$1;
(statearr_33012_35102[(1)] = (20));

} else {
var statearr_33013_35103 = state_32924__$1;
(statearr_33013_35103[(1)] = (21));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (10))){
var inst_32916 = (state_32924[(2)]);
var state_32924__$1 = state_32924;
var statearr_33014_35104 = state_32924__$1;
(statearr_33014_35104[(2)] = inst_32916);

(statearr_33014_35104[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (18))){
var inst_32887 = (state_32924[(16)]);
var state_32924__$1 = state_32924;
var statearr_33015_35109 = state_32924__$1;
(statearr_33015_35109[(2)] = inst_32887);

(statearr_33015_35109[(1)] = (19));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_32925 === (8))){
var inst_32851 = (state_32924[(7)]);
var inst_32867 = (inst_32851 == null);
var state_32924__$1 = state_32924;
if(cljs.core.truth_(inst_32867)){
var statearr_33016_35110 = state_32924__$1;
(statearr_33016_35110[(1)] = (11));

} else {
var statearr_33018_35111 = state_32924__$1;
(statearr_33018_35111[(1)] = (12));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$mix_$_state_machine__30544__auto__ = null;
var cljs$core$async$mix_$_state_machine__30544__auto____0 = (function (){
var statearr_33036 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_33036[(0)] = cljs$core$async$mix_$_state_machine__30544__auto__);

(statearr_33036[(1)] = (1));

return statearr_33036;
});
var cljs$core$async$mix_$_state_machine__30544__auto____1 = (function (state_32924){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_32924);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33042){var ex__30547__auto__ = e33042;
var statearr_33043_35112 = state_32924;
(statearr_33043_35112[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_32924[(4)]))){
var statearr_33044_35113 = state_32924;
(statearr_33044_35113[(1)] = cljs.core.first((state_32924[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35118 = state_32924;
state_32924 = G__35118;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$mix_$_state_machine__30544__auto__ = function(state_32924){
switch(arguments.length){
case 0:
return cljs$core$async$mix_$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$mix_$_state_machine__30544__auto____1.call(this,state_32924);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$mix_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$mix_$_state_machine__30544__auto____0;
cljs$core$async$mix_$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$mix_$_state_machine__30544__auto____1;
return cljs$core$async$mix_$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33046 = f__30908__auto__();
(statearr_33046[(6)] = c__30907__auto___35065);

return statearr_33046;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return m;
});
/**
 * Adds ch as an input to the mix
 */
cljs.core.async.admix = (function cljs$core$async$admix(mix,ch){
return cljs.core.async.admix_STAR_(mix,ch);
});
/**
 * Removes ch as an input to the mix
 */
cljs.core.async.unmix = (function cljs$core$async$unmix(mix,ch){
return cljs.core.async.unmix_STAR_(mix,ch);
});
/**
 * removes all inputs from the mix
 */
cljs.core.async.unmix_all = (function cljs$core$async$unmix_all(mix){
return cljs.core.async.unmix_all_STAR_(mix);
});
/**
 * Atomically sets the state(s) of one or more channels in a mix. The
 *   state map is a map of channels -> channel-state-map. A
 *   channel-state-map is a map of attrs -> boolean, where attr is one or
 *   more of :mute, :pause or :solo. Any states supplied are merged with
 *   the current state.
 * 
 *   Note that channels can be added to a mix via toggle, which can be
 *   used to add channels in a particular (e.g. paused) state.
 */
cljs.core.async.toggle = (function cljs$core$async$toggle(mix,state_map){
return cljs.core.async.toggle_STAR_(mix,state_map);
});
/**
 * Sets the solo mode of the mix. mode must be one of :mute or :pause
 */
cljs.core.async.solo_mode = (function cljs$core$async$solo_mode(mix,mode){
return cljs.core.async.solo_mode_STAR_(mix,mode);
});

/**
 * @interface
 */
cljs.core.async.Pub = function(){};

var cljs$core$async$Pub$sub_STAR_$dyn_35166 = (function (p,v,ch,close_QMARK_){
var x__5350__auto__ = (((p == null))?null:p);
var m__5351__auto__ = (cljs.core.async.sub_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$4 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$4(p,v,ch,close_QMARK_) : m__5351__auto__.call(null, p,v,ch,close_QMARK_));
} else {
var m__5349__auto__ = (cljs.core.async.sub_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$4 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$4(p,v,ch,close_QMARK_) : m__5349__auto__.call(null, p,v,ch,close_QMARK_));
} else {
throw cljs.core.missing_protocol("Pub.sub*",p);
}
}
});
cljs.core.async.sub_STAR_ = (function cljs$core$async$sub_STAR_(p,v,ch,close_QMARK_){
if((((!((p == null)))) && ((!((p.cljs$core$async$Pub$sub_STAR_$arity$4 == null)))))){
return p.cljs$core$async$Pub$sub_STAR_$arity$4(p,v,ch,close_QMARK_);
} else {
return cljs$core$async$Pub$sub_STAR_$dyn_35166(p,v,ch,close_QMARK_);
}
});

var cljs$core$async$Pub$unsub_STAR_$dyn_35170 = (function (p,v,ch){
var x__5350__auto__ = (((p == null))?null:p);
var m__5351__auto__ = (cljs.core.async.unsub_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$3 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$3(p,v,ch) : m__5351__auto__.call(null, p,v,ch));
} else {
var m__5349__auto__ = (cljs.core.async.unsub_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$3 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$3(p,v,ch) : m__5349__auto__.call(null, p,v,ch));
} else {
throw cljs.core.missing_protocol("Pub.unsub*",p);
}
}
});
cljs.core.async.unsub_STAR_ = (function cljs$core$async$unsub_STAR_(p,v,ch){
if((((!((p == null)))) && ((!((p.cljs$core$async$Pub$unsub_STAR_$arity$3 == null)))))){
return p.cljs$core$async$Pub$unsub_STAR_$arity$3(p,v,ch);
} else {
return cljs$core$async$Pub$unsub_STAR_$dyn_35170(p,v,ch);
}
});

var cljs$core$async$Pub$unsub_all_STAR_$dyn_35171 = (function() {
var G__35172 = null;
var G__35172__1 = (function (p){
var x__5350__auto__ = (((p == null))?null:p);
var m__5351__auto__ = (cljs.core.async.unsub_all_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$1(p) : m__5351__auto__.call(null, p));
} else {
var m__5349__auto__ = (cljs.core.async.unsub_all_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$1 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$1(p) : m__5349__auto__.call(null, p));
} else {
throw cljs.core.missing_protocol("Pub.unsub-all*",p);
}
}
});
var G__35172__2 = (function (p,v){
var x__5350__auto__ = (((p == null))?null:p);
var m__5351__auto__ = (cljs.core.async.unsub_all_STAR_[goog.typeOf(x__5350__auto__)]);
if((!((m__5351__auto__ == null)))){
return (m__5351__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5351__auto__.cljs$core$IFn$_invoke$arity$2(p,v) : m__5351__auto__.call(null, p,v));
} else {
var m__5349__auto__ = (cljs.core.async.unsub_all_STAR_["_"]);
if((!((m__5349__auto__ == null)))){
return (m__5349__auto__.cljs$core$IFn$_invoke$arity$2 ? m__5349__auto__.cljs$core$IFn$_invoke$arity$2(p,v) : m__5349__auto__.call(null, p,v));
} else {
throw cljs.core.missing_protocol("Pub.unsub-all*",p);
}
}
});
G__35172 = function(p,v){
switch(arguments.length){
case 1:
return G__35172__1.call(this,p);
case 2:
return G__35172__2.call(this,p,v);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
G__35172.cljs$core$IFn$_invoke$arity$1 = G__35172__1;
G__35172.cljs$core$IFn$_invoke$arity$2 = G__35172__2;
return G__35172;
})()
;
cljs.core.async.unsub_all_STAR_ = (function cljs$core$async$unsub_all_STAR_(var_args){
var G__33092 = arguments.length;
switch (G__33092) {
case 1:
return cljs.core.async.unsub_all_STAR_.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.unsub_all_STAR_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.unsub_all_STAR_.cljs$core$IFn$_invoke$arity$1 = (function (p){
if((((!((p == null)))) && ((!((p.cljs$core$async$Pub$unsub_all_STAR_$arity$1 == null)))))){
return p.cljs$core$async$Pub$unsub_all_STAR_$arity$1(p);
} else {
return cljs$core$async$Pub$unsub_all_STAR_$dyn_35171(p);
}
}));

(cljs.core.async.unsub_all_STAR_.cljs$core$IFn$_invoke$arity$2 = (function (p,v){
if((((!((p == null)))) && ((!((p.cljs$core$async$Pub$unsub_all_STAR_$arity$2 == null)))))){
return p.cljs$core$async$Pub$unsub_all_STAR_$arity$2(p,v);
} else {
return cljs$core$async$Pub$unsub_all_STAR_$dyn_35171(p,v);
}
}));

(cljs.core.async.unsub_all_STAR_.cljs$lang$maxFixedArity = 2);



/**
* @constructor
 * @implements {cljs.core.async.Pub}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.async.Mux}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async33110 = (function (ch,topic_fn,buf_fn,mults,ensure_mult,meta33111){
this.ch = ch;
this.topic_fn = topic_fn;
this.buf_fn = buf_fn;
this.mults = mults;
this.ensure_mult = ensure_mult;
this.meta33111 = meta33111;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_33112,meta33111__$1){
var self__ = this;
var _33112__$1 = this;
return (new cljs.core.async.t_cljs$core$async33110(self__.ch,self__.topic_fn,self__.buf_fn,self__.mults,self__.ensure_mult,meta33111__$1));
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_33112){
var self__ = this;
var _33112__$1 = this;
return self__.meta33111;
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Mux$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Mux$muxch_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return self__.ch;
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Pub$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Pub$sub_STAR_$arity$4 = (function (p,topic,ch__$1,close_QMARK_){
var self__ = this;
var p__$1 = this;
var m = (self__.ensure_mult.cljs$core$IFn$_invoke$arity$1 ? self__.ensure_mult.cljs$core$IFn$_invoke$arity$1(topic) : self__.ensure_mult.call(null, topic));
return cljs.core.async.tap.cljs$core$IFn$_invoke$arity$3(m,ch__$1,close_QMARK_);
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Pub$unsub_STAR_$arity$3 = (function (p,topic,ch__$1){
var self__ = this;
var p__$1 = this;
var temp__5825__auto__ = cljs.core.get.cljs$core$IFn$_invoke$arity$2(cljs.core.deref(self__.mults),topic);
if(cljs.core.truth_(temp__5825__auto__)){
var m = temp__5825__auto__;
return cljs.core.async.untap(m,ch__$1);
} else {
return null;
}
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Pub$unsub_all_STAR_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.reset_BANG_(self__.mults,cljs.core.PersistentArrayMap.EMPTY);
}));

(cljs.core.async.t_cljs$core$async33110.prototype.cljs$core$async$Pub$unsub_all_STAR_$arity$2 = (function (_,topic){
var self__ = this;
var ___$1 = this;
return cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(self__.mults,cljs.core.dissoc,topic);
}));

(cljs.core.async.t_cljs$core$async33110.getBasis = (function (){
return new cljs.core.PersistentVector(null, 6, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"topic-fn","topic-fn",-862449736,null),new cljs.core.Symbol(null,"buf-fn","buf-fn",-1200281591,null),new cljs.core.Symbol(null,"mults","mults",-461114485,null),new cljs.core.Symbol(null,"ensure-mult","ensure-mult",1796584816,null),new cljs.core.Symbol(null,"meta33111","meta33111",825510462,null)], null);
}));

(cljs.core.async.t_cljs$core$async33110.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async33110.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async33110");

(cljs.core.async.t_cljs$core$async33110.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async33110");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async33110.
 */
cljs.core.async.__GT_t_cljs$core$async33110 = (function cljs$core$async$__GT_t_cljs$core$async33110(ch,topic_fn,buf_fn,mults,ensure_mult,meta33111){
return (new cljs.core.async.t_cljs$core$async33110(ch,topic_fn,buf_fn,mults,ensure_mult,meta33111));
});


/**
 * Creates and returns a pub(lication) of the supplied channel,
 *   partitioned into topics by the topic-fn. topic-fn will be applied to
 *   each value on the channel and the result will determine the 'topic'
 *   on which that value will be put. Channels can be subscribed to
 *   receive copies of topics using 'sub', and unsubscribed using
 *   'unsub'. Each topic will be handled by an internal mult on a
 *   dedicated channel. By default these internal channels are
 *   unbuffered, but a buf-fn can be supplied which, given a topic,
 *   creates a buffer with desired properties.
 * 
 *   Each item is distributed to all subs in parallel and synchronously,
 *   i.e. each sub must accept before the next item is distributed. Use
 *   buffering/windowing to prevent slow subs from holding up the pub.
 * 
 *   Items received when there are no matching subs get dropped.
 * 
 *   Note that if buf-fns are used then each topic is handled
 *   asynchronously, i.e. if a channel is subscribed to more than one
 *   topic it should not expect them to be interleaved identically with
 *   the source.
 */
cljs.core.async.pub = (function cljs$core$async$pub(var_args){
var G__33107 = arguments.length;
switch (G__33107) {
case 2:
return cljs.core.async.pub.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.pub.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.pub.cljs$core$IFn$_invoke$arity$2 = (function (ch,topic_fn){
return cljs.core.async.pub.cljs$core$IFn$_invoke$arity$3(ch,topic_fn,cljs.core.constantly(null));
}));

(cljs.core.async.pub.cljs$core$IFn$_invoke$arity$3 = (function (ch,topic_fn,buf_fn){
var mults = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(cljs.core.PersistentArrayMap.EMPTY);
var ensure_mult = (function (topic){
var or__5002__auto__ = cljs.core.get.cljs$core$IFn$_invoke$arity$2(cljs.core.deref(mults),topic);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return cljs.core.get.cljs$core$IFn$_invoke$arity$2(cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$2(mults,(function (p1__33104_SHARP_){
if(cljs.core.truth_((p1__33104_SHARP_.cljs$core$IFn$_invoke$arity$1 ? p1__33104_SHARP_.cljs$core$IFn$_invoke$arity$1(topic) : p1__33104_SHARP_.call(null, topic)))){
return p1__33104_SHARP_;
} else {
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(p1__33104_SHARP_,topic,cljs.core.async.mult(cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((buf_fn.cljs$core$IFn$_invoke$arity$1 ? buf_fn.cljs$core$IFn$_invoke$arity$1(topic) : buf_fn.call(null, topic)))));
}
})),topic);
}
});
var p = (new cljs.core.async.t_cljs$core$async33110(ch,topic_fn,buf_fn,mults,ensure_mult,cljs.core.PersistentArrayMap.EMPTY));
var c__30907__auto___35190 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33222){
var state_val_33224 = (state_33222[(1)]);
if((state_val_33224 === (7))){
var inst_33192 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33229_35191 = state_33222__$1;
(statearr_33229_35191[(2)] = inst_33192);

(statearr_33229_35191[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (20))){
var state_33222__$1 = state_33222;
var statearr_33230_35195 = state_33222__$1;
(statearr_33230_35195[(2)] = null);

(statearr_33230_35195[(1)] = (21));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (1))){
var state_33222__$1 = state_33222;
var statearr_33232_35200 = state_33222__$1;
(statearr_33232_35200[(2)] = null);

(statearr_33232_35200[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (24))){
var inst_33175 = (state_33222[(7)]);
var inst_33184 = cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$3(mults,cljs.core.dissoc,inst_33175);
var state_33222__$1 = state_33222;
var statearr_33233_35201 = state_33222__$1;
(statearr_33233_35201[(2)] = inst_33184);

(statearr_33233_35201[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (4))){
var inst_33123 = (state_33222[(8)]);
var inst_33123__$1 = (state_33222[(2)]);
var inst_33128 = (inst_33123__$1 == null);
var state_33222__$1 = (function (){var statearr_33235 = state_33222;
(statearr_33235[(8)] = inst_33123__$1);

return statearr_33235;
})();
if(cljs.core.truth_(inst_33128)){
var statearr_33236_35202 = state_33222__$1;
(statearr_33236_35202[(1)] = (5));

} else {
var statearr_33237_35206 = state_33222__$1;
(statearr_33237_35206[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (15))){
var inst_33169 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33238_35211 = state_33222__$1;
(statearr_33238_35211[(2)] = inst_33169);

(statearr_33238_35211[(1)] = (12));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (21))){
var inst_33189 = (state_33222[(2)]);
var state_33222__$1 = (function (){var statearr_33239 = state_33222;
(statearr_33239[(9)] = inst_33189);

return statearr_33239;
})();
var statearr_33240_35212 = state_33222__$1;
(statearr_33240_35212[(2)] = null);

(statearr_33240_35212[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (13))){
var inst_33151 = (state_33222[(10)]);
var inst_33153 = cljs.core.chunked_seq_QMARK_(inst_33151);
var state_33222__$1 = state_33222;
if(inst_33153){
var statearr_33244_35220 = state_33222__$1;
(statearr_33244_35220[(1)] = (16));

} else {
var statearr_33246_35221 = state_33222__$1;
(statearr_33246_35221[(1)] = (17));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (22))){
var inst_33181 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
if(cljs.core.truth_(inst_33181)){
var statearr_33249_35227 = state_33222__$1;
(statearr_33249_35227[(1)] = (23));

} else {
var statearr_33250_35232 = state_33222__$1;
(statearr_33250_35232[(1)] = (24));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (6))){
var inst_33177 = (state_33222[(11)]);
var inst_33123 = (state_33222[(8)]);
var inst_33175 = (state_33222[(7)]);
var inst_33175__$1 = (topic_fn.cljs$core$IFn$_invoke$arity$1 ? topic_fn.cljs$core$IFn$_invoke$arity$1(inst_33123) : topic_fn.call(null, inst_33123));
var inst_33176 = cljs.core.deref(mults);
var inst_33177__$1 = cljs.core.get.cljs$core$IFn$_invoke$arity$2(inst_33176,inst_33175__$1);
var state_33222__$1 = (function (){var statearr_33251 = state_33222;
(statearr_33251[(11)] = inst_33177__$1);

(statearr_33251[(7)] = inst_33175__$1);

return statearr_33251;
})();
if(cljs.core.truth_(inst_33177__$1)){
var statearr_33252_35234 = state_33222__$1;
(statearr_33252_35234[(1)] = (19));

} else {
var statearr_33253_35235 = state_33222__$1;
(statearr_33253_35235[(1)] = (20));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (25))){
var inst_33186 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33254_35236 = state_33222__$1;
(statearr_33254_35236[(2)] = inst_33186);

(statearr_33254_35236[(1)] = (21));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (17))){
var inst_33151 = (state_33222[(10)]);
var inst_33160 = cljs.core.first(inst_33151);
var inst_33161 = cljs.core.async.muxch_STAR_(inst_33160);
var inst_33162 = cljs.core.async.close_BANG_(inst_33161);
var inst_33163 = cljs.core.next(inst_33151);
var inst_33137 = inst_33163;
var inst_33138 = null;
var inst_33139 = (0);
var inst_33140 = (0);
var state_33222__$1 = (function (){var statearr_33255 = state_33222;
(statearr_33255[(12)] = inst_33137);

(statearr_33255[(13)] = inst_33140);

(statearr_33255[(14)] = inst_33139);

(statearr_33255[(15)] = inst_33138);

(statearr_33255[(16)] = inst_33162);

return statearr_33255;
})();
var statearr_33256_35237 = state_33222__$1;
(statearr_33256_35237[(2)] = null);

(statearr_33256_35237[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (3))){
var inst_33207 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33222__$1,inst_33207);
} else {
if((state_val_33224 === (12))){
var inst_33171 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33257_35238 = state_33222__$1;
(statearr_33257_35238[(2)] = inst_33171);

(statearr_33257_35238[(1)] = (9));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (2))){
var state_33222__$1 = state_33222;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33222__$1,(4),ch);
} else {
if((state_val_33224 === (23))){
var state_33222__$1 = state_33222;
var statearr_33258_35242 = state_33222__$1;
(statearr_33258_35242[(2)] = null);

(statearr_33258_35242[(1)] = (25));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (19))){
var inst_33177 = (state_33222[(11)]);
var inst_33123 = (state_33222[(8)]);
var inst_33179 = cljs.core.async.muxch_STAR_(inst_33177);
var state_33222__$1 = state_33222;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33222__$1,(22),inst_33179,inst_33123);
} else {
if((state_val_33224 === (11))){
var inst_33137 = (state_33222[(12)]);
var inst_33151 = (state_33222[(10)]);
var inst_33151__$1 = cljs.core.seq(inst_33137);
var state_33222__$1 = (function (){var statearr_33259 = state_33222;
(statearr_33259[(10)] = inst_33151__$1);

return statearr_33259;
})();
if(inst_33151__$1){
var statearr_33260_35249 = state_33222__$1;
(statearr_33260_35249[(1)] = (13));

} else {
var statearr_33261_35250 = state_33222__$1;
(statearr_33261_35250[(1)] = (14));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (9))){
var inst_33173 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33265_35251 = state_33222__$1;
(statearr_33265_35251[(2)] = inst_33173);

(statearr_33265_35251[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (5))){
var inst_33134 = cljs.core.deref(mults);
var inst_33135 = cljs.core.vals(inst_33134);
var inst_33136 = cljs.core.seq(inst_33135);
var inst_33137 = inst_33136;
var inst_33138 = null;
var inst_33139 = (0);
var inst_33140 = (0);
var state_33222__$1 = (function (){var statearr_33266 = state_33222;
(statearr_33266[(12)] = inst_33137);

(statearr_33266[(13)] = inst_33140);

(statearr_33266[(14)] = inst_33139);

(statearr_33266[(15)] = inst_33138);

return statearr_33266;
})();
var statearr_33267_35253 = state_33222__$1;
(statearr_33267_35253[(2)] = null);

(statearr_33267_35253[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (14))){
var state_33222__$1 = state_33222;
var statearr_33271_35255 = state_33222__$1;
(statearr_33271_35255[(2)] = null);

(statearr_33271_35255[(1)] = (15));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (16))){
var inst_33151 = (state_33222[(10)]);
var inst_33155 = cljs.core.chunk_first(inst_33151);
var inst_33156 = cljs.core.chunk_rest(inst_33151);
var inst_33157 = cljs.core.count(inst_33155);
var inst_33137 = inst_33156;
var inst_33138 = inst_33155;
var inst_33139 = inst_33157;
var inst_33140 = (0);
var state_33222__$1 = (function (){var statearr_33272 = state_33222;
(statearr_33272[(12)] = inst_33137);

(statearr_33272[(13)] = inst_33140);

(statearr_33272[(14)] = inst_33139);

(statearr_33272[(15)] = inst_33138);

return statearr_33272;
})();
var statearr_33273_35257 = state_33222__$1;
(statearr_33273_35257[(2)] = null);

(statearr_33273_35257[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (10))){
var inst_33137 = (state_33222[(12)]);
var inst_33140 = (state_33222[(13)]);
var inst_33139 = (state_33222[(14)]);
var inst_33138 = (state_33222[(15)]);
var inst_33145 = cljs.core._nth(inst_33138,inst_33140);
var inst_33146 = cljs.core.async.muxch_STAR_(inst_33145);
var inst_33147 = cljs.core.async.close_BANG_(inst_33146);
var inst_33148 = (inst_33140 + (1));
var tmp33268 = inst_33137;
var tmp33269 = inst_33139;
var tmp33270 = inst_33138;
var inst_33137__$1 = tmp33268;
var inst_33138__$1 = tmp33270;
var inst_33139__$1 = tmp33269;
var inst_33140__$1 = inst_33148;
var state_33222__$1 = (function (){var statearr_33279 = state_33222;
(statearr_33279[(12)] = inst_33137__$1);

(statearr_33279[(13)] = inst_33140__$1);

(statearr_33279[(14)] = inst_33139__$1);

(statearr_33279[(15)] = inst_33138__$1);

(statearr_33279[(17)] = inst_33147);

return statearr_33279;
})();
var statearr_33280_35261 = state_33222__$1;
(statearr_33280_35261[(2)] = null);

(statearr_33280_35261[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (18))){
var inst_33166 = (state_33222[(2)]);
var state_33222__$1 = state_33222;
var statearr_33281_35263 = state_33222__$1;
(statearr_33281_35263[(2)] = inst_33166);

(statearr_33281_35263[(1)] = (15));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33224 === (8))){
var inst_33140 = (state_33222[(13)]);
var inst_33139 = (state_33222[(14)]);
var inst_33142 = (inst_33140 < inst_33139);
var inst_33143 = inst_33142;
var state_33222__$1 = state_33222;
if(cljs.core.truth_(inst_33143)){
var statearr_33282_35265 = state_33222__$1;
(statearr_33282_35265[(1)] = (10));

} else {
var statearr_33283_35266 = state_33222__$1;
(statearr_33283_35266[(1)] = (11));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33290 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_33290[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33290[(1)] = (1));

return statearr_33290;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33222){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33222);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33291){var ex__30547__auto__ = e33291;
var statearr_33292_35267 = state_33222;
(statearr_33292_35267[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33222[(4)]))){
var statearr_33293_35268 = state_33222;
(statearr_33293_35268[(1)] = cljs.core.first((state_33222[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35269 = state_33222;
state_33222 = G__35269;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33222){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33222);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33316 = f__30908__auto__();
(statearr_33316[(6)] = c__30907__auto___35190);

return statearr_33316;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return p;
}));

(cljs.core.async.pub.cljs$lang$maxFixedArity = 3);

/**
 * Subscribes a channel to a topic of a pub.
 * 
 *   By default the channel will be closed when the source closes,
 *   but can be determined by the close? parameter.
 */
cljs.core.async.sub = (function cljs$core$async$sub(var_args){
var G__33325 = arguments.length;
switch (G__33325) {
case 3:
return cljs.core.async.sub.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
case 4:
return cljs.core.async.sub.cljs$core$IFn$_invoke$arity$4((arguments[(0)]),(arguments[(1)]),(arguments[(2)]),(arguments[(3)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.sub.cljs$core$IFn$_invoke$arity$3 = (function (p,topic,ch){
return cljs.core.async.sub.cljs$core$IFn$_invoke$arity$4(p,topic,ch,true);
}));

(cljs.core.async.sub.cljs$core$IFn$_invoke$arity$4 = (function (p,topic,ch,close_QMARK_){
return cljs.core.async.sub_STAR_(p,topic,ch,close_QMARK_);
}));

(cljs.core.async.sub.cljs$lang$maxFixedArity = 4);

/**
 * Unsubscribes a channel from a topic of a pub
 */
cljs.core.async.unsub = (function cljs$core$async$unsub(p,topic,ch){
return cljs.core.async.unsub_STAR_(p,topic,ch);
});
/**
 * Unsubscribes all channels from a pub, or a topic of a pub
 */
cljs.core.async.unsub_all = (function cljs$core$async$unsub_all(var_args){
var G__33339 = arguments.length;
switch (G__33339) {
case 1:
return cljs.core.async.unsub_all.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.unsub_all.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.unsub_all.cljs$core$IFn$_invoke$arity$1 = (function (p){
return cljs.core.async.unsub_all_STAR_(p);
}));

(cljs.core.async.unsub_all.cljs$core$IFn$_invoke$arity$2 = (function (p,topic){
return cljs.core.async.unsub_all_STAR_(p,topic);
}));

(cljs.core.async.unsub_all.cljs$lang$maxFixedArity = 2);

/**
 * Takes a function and a collection of source channels, and returns a
 *   channel which contains the values produced by applying f to the set
 *   of first items taken from each source channel, followed by applying
 *   f to the set of second items from each channel, until any one of the
 *   channels is closed, at which point the output channel will be
 *   closed. The returned channel will be unbuffered by default, or a
 *   buf-or-n can be supplied
 */
cljs.core.async.map = (function cljs$core$async$map(var_args){
var G__33341 = arguments.length;
switch (G__33341) {
case 2:
return cljs.core.async.map.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.map.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.map.cljs$core$IFn$_invoke$arity$2 = (function (f,chs){
return cljs.core.async.map.cljs$core$IFn$_invoke$arity$3(f,chs,null);
}));

(cljs.core.async.map.cljs$core$IFn$_invoke$arity$3 = (function (f,chs,buf_or_n){
var chs__$1 = cljs.core.vec(chs);
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var cnt = cljs.core.count(chs__$1);
var rets = cljs.core.object_array.cljs$core$IFn$_invoke$arity$1(cnt);
var dchan = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
var dctr = cljs.core.atom.cljs$core$IFn$_invoke$arity$1(null);
var done = cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (i){
return (function (ret){
(rets[i] = ret);

if((cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$2(dctr,cljs.core.dec) === (0))){
return cljs.core.async.put_BANG_.cljs$core$IFn$_invoke$arity$2(dchan,rets.slice((0)));
} else {
return null;
}
});
}),cljs.core.range.cljs$core$IFn$_invoke$arity$1(cnt));
if((cnt === (0))){
cljs.core.async.close_BANG_(out);
} else {
var c__30907__auto___35325 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33396){
var state_val_33397 = (state_33396[(1)]);
if((state_val_33397 === (7))){
var state_33396__$1 = state_33396;
var statearr_33402_35327 = state_33396__$1;
(statearr_33402_35327[(2)] = null);

(statearr_33402_35327[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (1))){
var state_33396__$1 = state_33396;
var statearr_33411_35348 = state_33396__$1;
(statearr_33411_35348[(2)] = null);

(statearr_33411_35348[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (4))){
var inst_33345 = (state_33396[(7)]);
var inst_33349 = (state_33396[(8)]);
var inst_33351 = (inst_33349 < inst_33345);
var state_33396__$1 = state_33396;
if(cljs.core.truth_(inst_33351)){
var statearr_33421_35349 = state_33396__$1;
(statearr_33421_35349[(1)] = (6));

} else {
var statearr_33423_35350 = state_33396__$1;
(statearr_33423_35350[(1)] = (7));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (15))){
var inst_33379 = (state_33396[(9)]);
var inst_33384 = cljs.core.apply.cljs$core$IFn$_invoke$arity$2(f,inst_33379);
var state_33396__$1 = state_33396;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33396__$1,(17),out,inst_33384);
} else {
if((state_val_33397 === (13))){
var inst_33379 = (state_33396[(9)]);
var inst_33379__$1 = (state_33396[(2)]);
var inst_33380 = cljs.core.some(cljs.core.nil_QMARK_,inst_33379__$1);
var state_33396__$1 = (function (){var statearr_33426 = state_33396;
(statearr_33426[(9)] = inst_33379__$1);

return statearr_33426;
})();
if(cljs.core.truth_(inst_33380)){
var statearr_33427_35398 = state_33396__$1;
(statearr_33427_35398[(1)] = (14));

} else {
var statearr_33428_35399 = state_33396__$1;
(statearr_33428_35399[(1)] = (15));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (6))){
var state_33396__$1 = state_33396;
var statearr_33429_35404 = state_33396__$1;
(statearr_33429_35404[(2)] = null);

(statearr_33429_35404[(1)] = (9));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (17))){
var inst_33386 = (state_33396[(2)]);
var state_33396__$1 = (function (){var statearr_33432 = state_33396;
(statearr_33432[(10)] = inst_33386);

return statearr_33432;
})();
var statearr_33437_35407 = state_33396__$1;
(statearr_33437_35407[(2)] = null);

(statearr_33437_35407[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (3))){
var inst_33391 = (state_33396[(2)]);
var state_33396__$1 = state_33396;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33396__$1,inst_33391);
} else {
if((state_val_33397 === (12))){
var _ = (function (){var statearr_33452 = state_33396;
(statearr_33452[(4)] = cljs.core.rest((state_33396[(4)])));

return statearr_33452;
})();
var state_33396__$1 = state_33396;
var ex33431 = (state_33396__$1[(2)]);
var statearr_33456_35409 = state_33396__$1;
(statearr_33456_35409[(5)] = ex33431);


if((ex33431 instanceof Object)){
var statearr_33457_35410 = state_33396__$1;
(statearr_33457_35410[(1)] = (11));

(statearr_33457_35410[(5)] = null);

} else {
throw ex33431;

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (2))){
var inst_33343 = cljs.core.reset_BANG_(dctr,cnt);
var inst_33345 = cnt;
var inst_33349 = (0);
var state_33396__$1 = (function (){var statearr_33464 = state_33396;
(statearr_33464[(7)] = inst_33345);

(statearr_33464[(8)] = inst_33349);

(statearr_33464[(11)] = inst_33343);

return statearr_33464;
})();
var statearr_33465_35411 = state_33396__$1;
(statearr_33465_35411[(2)] = null);

(statearr_33465_35411[(1)] = (4));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (11))){
var inst_33358 = (state_33396[(2)]);
var inst_33359 = cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$2(dctr,cljs.core.dec);
var state_33396__$1 = (function (){var statearr_33466 = state_33396;
(statearr_33466[(12)] = inst_33358);

return statearr_33466;
})();
var statearr_33468_35412 = state_33396__$1;
(statearr_33468_35412[(2)] = inst_33359);

(statearr_33468_35412[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (9))){
var inst_33349 = (state_33396[(8)]);
var _ = (function (){var statearr_33472 = state_33396;
(statearr_33472[(4)] = cljs.core.cons((12),(state_33396[(4)])));

return statearr_33472;
})();
var inst_33365 = (chs__$1.cljs$core$IFn$_invoke$arity$1 ? chs__$1.cljs$core$IFn$_invoke$arity$1(inst_33349) : chs__$1.call(null, inst_33349));
var inst_33366 = (done.cljs$core$IFn$_invoke$arity$1 ? done.cljs$core$IFn$_invoke$arity$1(inst_33349) : done.call(null, inst_33349));
var inst_33367 = cljs.core.async.take_BANG_.cljs$core$IFn$_invoke$arity$2(inst_33365,inst_33366);
var ___$1 = (function (){var statearr_33475 = state_33396;
(statearr_33475[(4)] = cljs.core.rest((state_33396[(4)])));

return statearr_33475;
})();
var state_33396__$1 = state_33396;
var statearr_33476_35430 = state_33396__$1;
(statearr_33476_35430[(2)] = inst_33367);

(statearr_33476_35430[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (5))){
var inst_33377 = (state_33396[(2)]);
var state_33396__$1 = (function (){var statearr_33478 = state_33396;
(statearr_33478[(13)] = inst_33377);

return statearr_33478;
})();
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33396__$1,(13),dchan);
} else {
if((state_val_33397 === (14))){
var inst_33382 = cljs.core.async.close_BANG_(out);
var state_33396__$1 = state_33396;
var statearr_33479_35452 = state_33396__$1;
(statearr_33479_35452[(2)] = inst_33382);

(statearr_33479_35452[(1)] = (16));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (16))){
var inst_33389 = (state_33396[(2)]);
var state_33396__$1 = state_33396;
var statearr_33480_35509 = state_33396__$1;
(statearr_33480_35509[(2)] = inst_33389);

(statearr_33480_35509[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (10))){
var inst_33349 = (state_33396[(8)]);
var inst_33370 = (state_33396[(2)]);
var inst_33371 = (inst_33349 + (1));
var inst_33349__$1 = inst_33371;
var state_33396__$1 = (function (){var statearr_33485 = state_33396;
(statearr_33485[(14)] = inst_33370);

(statearr_33485[(8)] = inst_33349__$1);

return statearr_33485;
})();
var statearr_33487_35568 = state_33396__$1;
(statearr_33487_35568[(2)] = null);

(statearr_33487_35568[(1)] = (4));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33397 === (8))){
var inst_33375 = (state_33396[(2)]);
var state_33396__$1 = state_33396;
var statearr_33488_35569 = state_33396__$1;
(statearr_33488_35569[(2)] = inst_33375);

(statearr_33488_35569[(1)] = (5));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33492 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_33492[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33492[(1)] = (1));

return statearr_33492;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33396){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33396);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33494){var ex__30547__auto__ = e33494;
var statearr_33495_35573 = state_33396;
(statearr_33495_35573[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33396[(4)]))){
var statearr_33498_35575 = state_33396;
(statearr_33498_35575[(1)] = cljs.core.first((state_33396[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35577 = state_33396;
state_33396 = G__35577;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33396){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33396);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33504 = f__30908__auto__();
(statearr_33504[(6)] = c__30907__auto___35325);

return statearr_33504;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

}

return out;
}));

(cljs.core.async.map.cljs$lang$maxFixedArity = 3);

/**
 * Takes a collection of source channels and returns a channel which
 *   contains all values taken from them. The returned channel will be
 *   unbuffered by default, or a buf-or-n can be supplied. The channel
 *   will close after all the source channels have closed.
 */
cljs.core.async.merge = (function cljs$core$async$merge(var_args){
var G__33507 = arguments.length;
switch (G__33507) {
case 1:
return cljs.core.async.merge.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.merge.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.merge.cljs$core$IFn$_invoke$arity$1 = (function (chs){
return cljs.core.async.merge.cljs$core$IFn$_invoke$arity$2(chs,null);
}));

(cljs.core.async.merge.cljs$core$IFn$_invoke$arity$2 = (function (chs,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___35580 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33550){
var state_val_33551 = (state_33550[(1)]);
if((state_val_33551 === (7))){
var inst_33518 = (state_33550[(7)]);
var inst_33519 = (state_33550[(8)]);
var inst_33518__$1 = (state_33550[(2)]);
var inst_33519__$1 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_33518__$1,(0),null);
var inst_33520 = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(inst_33518__$1,(1),null);
var inst_33521 = (inst_33519__$1 == null);
var state_33550__$1 = (function (){var statearr_33556 = state_33550;
(statearr_33556[(7)] = inst_33518__$1);

(statearr_33556[(8)] = inst_33519__$1);

(statearr_33556[(9)] = inst_33520);

return statearr_33556;
})();
if(cljs.core.truth_(inst_33521)){
var statearr_33557_35584 = state_33550__$1;
(statearr_33557_35584[(1)] = (8));

} else {
var statearr_33558_35587 = state_33550__$1;
(statearr_33558_35587[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (1))){
var inst_33508 = cljs.core.vec(chs);
var inst_33509 = inst_33508;
var state_33550__$1 = (function (){var statearr_33559 = state_33550;
(statearr_33559[(10)] = inst_33509);

return statearr_33559;
})();
var statearr_33560_35592 = state_33550__$1;
(statearr_33560_35592[(2)] = null);

(statearr_33560_35592[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (4))){
var inst_33509 = (state_33550[(10)]);
var state_33550__$1 = state_33550;
return cljs.core.async.ioc_alts_BANG_(state_33550__$1,(7),inst_33509);
} else {
if((state_val_33551 === (6))){
var inst_33542 = (state_33550[(2)]);
var state_33550__$1 = state_33550;
var statearr_33561_35595 = state_33550__$1;
(statearr_33561_35595[(2)] = inst_33542);

(statearr_33561_35595[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (3))){
var inst_33544 = (state_33550[(2)]);
var state_33550__$1 = state_33550;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33550__$1,inst_33544);
} else {
if((state_val_33551 === (2))){
var inst_33509 = (state_33550[(10)]);
var inst_33511 = cljs.core.count(inst_33509);
var inst_33512 = (inst_33511 > (0));
var state_33550__$1 = state_33550;
if(cljs.core.truth_(inst_33512)){
var statearr_33564_35599 = state_33550__$1;
(statearr_33564_35599[(1)] = (4));

} else {
var statearr_33565_35601 = state_33550__$1;
(statearr_33565_35601[(1)] = (5));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (11))){
var inst_33509 = (state_33550[(10)]);
var inst_33535 = (state_33550[(2)]);
var tmp33563 = inst_33509;
var inst_33509__$1 = tmp33563;
var state_33550__$1 = (function (){var statearr_33566 = state_33550;
(statearr_33566[(11)] = inst_33535);

(statearr_33566[(10)] = inst_33509__$1);

return statearr_33566;
})();
var statearr_33567_35603 = state_33550__$1;
(statearr_33567_35603[(2)] = null);

(statearr_33567_35603[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (9))){
var inst_33519 = (state_33550[(8)]);
var state_33550__$1 = state_33550;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33550__$1,(11),out,inst_33519);
} else {
if((state_val_33551 === (5))){
var inst_33540 = cljs.core.async.close_BANG_(out);
var state_33550__$1 = state_33550;
var statearr_33574_35612 = state_33550__$1;
(statearr_33574_35612[(2)] = inst_33540);

(statearr_33574_35612[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (10))){
var inst_33538 = (state_33550[(2)]);
var state_33550__$1 = state_33550;
var statearr_33578_35617 = state_33550__$1;
(statearr_33578_35617[(2)] = inst_33538);

(statearr_33578_35617[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33551 === (8))){
var inst_33518 = (state_33550[(7)]);
var inst_33519 = (state_33550[(8)]);
var inst_33509 = (state_33550[(10)]);
var inst_33520 = (state_33550[(9)]);
var inst_33529 = (function (){var cs = inst_33509;
var vec__33514 = inst_33518;
var v = inst_33519;
var c = inst_33520;
return (function (p1__33505_SHARP_){
return cljs.core.not_EQ_.cljs$core$IFn$_invoke$arity$2(c,p1__33505_SHARP_);
});
})();
var inst_33530 = cljs.core.filterv(inst_33529,inst_33509);
var inst_33509__$1 = inst_33530;
var state_33550__$1 = (function (){var statearr_33584 = state_33550;
(statearr_33584[(10)] = inst_33509__$1);

return statearr_33584;
})();
var statearr_33588_35618 = state_33550__$1;
(statearr_33588_35618[(2)] = null);

(statearr_33588_35618[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33589 = [null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_33589[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33589[(1)] = (1));

return statearr_33589;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33550){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33550);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33592){var ex__30547__auto__ = e33592;
var statearr_33593_35619 = state_33550;
(statearr_33593_35619[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33550[(4)]))){
var statearr_33594_35620 = state_33550;
(statearr_33594_35620[(1)] = cljs.core.first((state_33550[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35621 = state_33550;
state_33550 = G__35621;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33550){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33550);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33595 = f__30908__auto__();
(statearr_33595[(6)] = c__30907__auto___35580);

return statearr_33595;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.merge.cljs$lang$maxFixedArity = 2);

/**
 * Returns a channel containing the single (collection) result of the
 *   items taken from the channel conjoined to the supplied
 *   collection. ch must close before into produces a result.
 */
cljs.core.async.into = (function cljs$core$async$into(coll,ch){
return cljs.core.async.reduce(cljs.core.conj,coll,ch);
});
/**
 * Returns a channel that will return, at most, n items from ch. After n items
 * have been returned, or ch has been closed, the return chanel will close.
 * 
 *   The output channel is unbuffered by default, unless buf-or-n is given.
 */
cljs.core.async.take = (function cljs$core$async$take(var_args){
var G__33603 = arguments.length;
switch (G__33603) {
case 2:
return cljs.core.async.take.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.take.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.take.cljs$core$IFn$_invoke$arity$2 = (function (n,ch){
return cljs.core.async.take.cljs$core$IFn$_invoke$arity$3(n,ch,null);
}));

(cljs.core.async.take.cljs$core$IFn$_invoke$arity$3 = (function (n,ch,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___35626 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33630){
var state_val_33631 = (state_33630[(1)]);
if((state_val_33631 === (7))){
var inst_33612 = (state_33630[(7)]);
var inst_33612__$1 = (state_33630[(2)]);
var inst_33613 = (inst_33612__$1 == null);
var inst_33614 = cljs.core.not(inst_33613);
var state_33630__$1 = (function (){var statearr_33632 = state_33630;
(statearr_33632[(7)] = inst_33612__$1);

return statearr_33632;
})();
if(inst_33614){
var statearr_33633_35682 = state_33630__$1;
(statearr_33633_35682[(1)] = (8));

} else {
var statearr_33634_35683 = state_33630__$1;
(statearr_33634_35683[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (1))){
var inst_33607 = (0);
var state_33630__$1 = (function (){var statearr_33635 = state_33630;
(statearr_33635[(8)] = inst_33607);

return statearr_33635;
})();
var statearr_33636_35692 = state_33630__$1;
(statearr_33636_35692[(2)] = null);

(statearr_33636_35692[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (4))){
var state_33630__$1 = state_33630;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33630__$1,(7),ch);
} else {
if((state_val_33631 === (6))){
var inst_33625 = (state_33630[(2)]);
var state_33630__$1 = state_33630;
var statearr_33637_35694 = state_33630__$1;
(statearr_33637_35694[(2)] = inst_33625);

(statearr_33637_35694[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (3))){
var inst_33627 = (state_33630[(2)]);
var inst_33628 = cljs.core.async.close_BANG_(out);
var state_33630__$1 = (function (){var statearr_33638 = state_33630;
(statearr_33638[(9)] = inst_33627);

return statearr_33638;
})();
return cljs.core.async.impl.ioc_helpers.return_chan(state_33630__$1,inst_33628);
} else {
if((state_val_33631 === (2))){
var inst_33607 = (state_33630[(8)]);
var inst_33609 = (inst_33607 < n);
var state_33630__$1 = state_33630;
if(cljs.core.truth_(inst_33609)){
var statearr_33639_35749 = state_33630__$1;
(statearr_33639_35749[(1)] = (4));

} else {
var statearr_33640_35757 = state_33630__$1;
(statearr_33640_35757[(1)] = (5));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (11))){
var inst_33607 = (state_33630[(8)]);
var inst_33617 = (state_33630[(2)]);
var inst_33618 = (inst_33607 + (1));
var inst_33607__$1 = inst_33618;
var state_33630__$1 = (function (){var statearr_33641 = state_33630;
(statearr_33641[(10)] = inst_33617);

(statearr_33641[(8)] = inst_33607__$1);

return statearr_33641;
})();
var statearr_33642_35774 = state_33630__$1;
(statearr_33642_35774[(2)] = null);

(statearr_33642_35774[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (9))){
var state_33630__$1 = state_33630;
var statearr_33643_35777 = state_33630__$1;
(statearr_33643_35777[(2)] = null);

(statearr_33643_35777[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (5))){
var state_33630__$1 = state_33630;
var statearr_33646_35778 = state_33630__$1;
(statearr_33646_35778[(2)] = null);

(statearr_33646_35778[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (10))){
var inst_33622 = (state_33630[(2)]);
var state_33630__$1 = state_33630;
var statearr_33647_35779 = state_33630__$1;
(statearr_33647_35779[(2)] = inst_33622);

(statearr_33647_35779[(1)] = (6));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33631 === (8))){
var inst_33612 = (state_33630[(7)]);
var state_33630__$1 = state_33630;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33630__$1,(11),out,inst_33612);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33648 = [null,null,null,null,null,null,null,null,null,null,null];
(statearr_33648[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33648[(1)] = (1));

return statearr_33648;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33630){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33630);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33650){var ex__30547__auto__ = e33650;
var statearr_33652_35782 = state_33630;
(statearr_33652_35782[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33630[(4)]))){
var statearr_33653_35783 = state_33630;
(statearr_33653_35783[(1)] = cljs.core.first((state_33630[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__35785 = state_33630;
state_33630 = G__35785;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33630){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33630);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33654 = f__30908__auto__();
(statearr_33654[(6)] = c__30907__auto___35626);

return statearr_33654;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.take.cljs$lang$maxFixedArity = 3);


/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Handler}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async33662 = (function (f,ch,meta33657,_,fn1,meta33663){
this.f = f;
this.ch = ch;
this.meta33657 = meta33657;
this._ = _;
this.fn1 = fn1;
this.meta33663 = meta33663;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_33664,meta33663__$1){
var self__ = this;
var _33664__$1 = this;
return (new cljs.core.async.t_cljs$core$async33662(self__.f,self__.ch,self__.meta33657,self__._,self__.fn1,meta33663__$1));
}));

(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_33664){
var self__ = this;
var _33664__$1 = this;
return self__.meta33663;
}));

(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$async$impl$protocols$Handler$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$async$impl$protocols$Handler$active_QMARK_$arity$1 = (function (___$1){
var self__ = this;
var ___$2 = this;
return cljs.core.async.impl.protocols.active_QMARK_(self__.fn1);
}));

(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$async$impl$protocols$Handler$blockable_QMARK_$arity$1 = (function (___$1){
var self__ = this;
var ___$2 = this;
return true;
}));

(cljs.core.async.t_cljs$core$async33662.prototype.cljs$core$async$impl$protocols$Handler$commit$arity$1 = (function (___$1){
var self__ = this;
var ___$2 = this;
var f1 = cljs.core.async.impl.protocols.commit(self__.fn1);
return (function (p1__33655_SHARP_){
var G__33672 = (((p1__33655_SHARP_ == null))?null:(self__.f.cljs$core$IFn$_invoke$arity$1 ? self__.f.cljs$core$IFn$_invoke$arity$1(p1__33655_SHARP_) : self__.f.call(null, p1__33655_SHARP_)));
return (f1.cljs$core$IFn$_invoke$arity$1 ? f1.cljs$core$IFn$_invoke$arity$1(G__33672) : f1.call(null, G__33672));
});
}));

(cljs.core.async.t_cljs$core$async33662.getBasis = (function (){
return new cljs.core.PersistentVector(null, 6, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"f","f",43394975,null),new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"meta33657","meta33657",-1572299145,null),cljs.core.with_meta(new cljs.core.Symbol(null,"_","_",-1201019570,null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"tag","tag",-1290361223),new cljs.core.Symbol("cljs.core.async","t_cljs$core$async33656","cljs.core.async/t_cljs$core$async33656",-1665166633,null)], null)),new cljs.core.Symbol(null,"fn1","fn1",895834444,null),new cljs.core.Symbol(null,"meta33663","meta33663",1370616423,null)], null);
}));

(cljs.core.async.t_cljs$core$async33662.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async33662.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async33662");

(cljs.core.async.t_cljs$core$async33662.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async33662");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async33662.
 */
cljs.core.async.__GT_t_cljs$core$async33662 = (function cljs$core$async$__GT_t_cljs$core$async33662(f,ch,meta33657,_,fn1,meta33663){
return (new cljs.core.async.t_cljs$core$async33662(f,ch,meta33657,_,fn1,meta33663));
});



/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Channel}
 * @implements {cljs.core.async.impl.protocols.WritePort}
 * @implements {cljs.core.async.impl.protocols.ReadPort}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async33656 = (function (f,ch,meta33657){
this.f = f;
this.ch = ch;
this.meta33657 = meta33657;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_33658,meta33657__$1){
var self__ = this;
var _33658__$1 = this;
return (new cljs.core.async.t_cljs$core$async33656(self__.f,self__.ch,meta33657__$1));
}));

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_33658){
var self__ = this;
var _33658__$1 = this;
return self__.meta33657;
}));

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$Channel$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$Channel$close_BANG_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.close_BANG_(self__.ch);
}));

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$Channel$closed_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.closed_QMARK_(self__.ch);
}));

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$ReadPort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$ReadPort$take_BANG_$arity$2 = (function (_,fn1){
var self__ = this;
var ___$1 = this;
var ret = cljs.core.async.impl.protocols.take_BANG_(self__.ch,(new cljs.core.async.t_cljs$core$async33662(self__.f,self__.ch,self__.meta33657,___$1,fn1,cljs.core.PersistentArrayMap.EMPTY)));
if(cljs.core.truth_((function (){var and__5000__auto__ = ret;
if(cljs.core.truth_(and__5000__auto__)){
return (!((cljs.core.deref(ret) == null)));
} else {
return and__5000__auto__;
}
})())){
return cljs.core.async.impl.channels.box((function (){var G__33673 = cljs.core.deref(ret);
return (self__.f.cljs$core$IFn$_invoke$arity$1 ? self__.f.cljs$core$IFn$_invoke$arity$1(G__33673) : self__.f.call(null, G__33673));
})());
} else {
return ret;
}
}));

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$WritePort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33656.prototype.cljs$core$async$impl$protocols$WritePort$put_BANG_$arity$3 = (function (_,val,fn1){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.put_BANG_(self__.ch,val,fn1);
}));

(cljs.core.async.t_cljs$core$async33656.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"f","f",43394975,null),new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"meta33657","meta33657",-1572299145,null)], null);
}));

(cljs.core.async.t_cljs$core$async33656.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async33656.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async33656");

(cljs.core.async.t_cljs$core$async33656.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async33656");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async33656.
 */
cljs.core.async.__GT_t_cljs$core$async33656 = (function cljs$core$async$__GT_t_cljs$core$async33656(f,ch,meta33657){
return (new cljs.core.async.t_cljs$core$async33656(f,ch,meta33657));
});


/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.map_LT_ = (function cljs$core$async$map_LT_(f,ch){
return (new cljs.core.async.t_cljs$core$async33656(f,ch,cljs.core.PersistentArrayMap.EMPTY));
});

/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Channel}
 * @implements {cljs.core.async.impl.protocols.WritePort}
 * @implements {cljs.core.async.impl.protocols.ReadPort}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async33676 = (function (f,ch,meta33677){
this.f = f;
this.ch = ch;
this.meta33677 = meta33677;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_33678,meta33677__$1){
var self__ = this;
var _33678__$1 = this;
return (new cljs.core.async.t_cljs$core$async33676(self__.f,self__.ch,meta33677__$1));
}));

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_33678){
var self__ = this;
var _33678__$1 = this;
return self__.meta33677;
}));

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$Channel$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$Channel$close_BANG_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.close_BANG_(self__.ch);
}));

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$ReadPort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$ReadPort$take_BANG_$arity$2 = (function (_,fn1){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.take_BANG_(self__.ch,fn1);
}));

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$WritePort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33676.prototype.cljs$core$async$impl$protocols$WritePort$put_BANG_$arity$3 = (function (_,val,fn1){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.put_BANG_(self__.ch,(self__.f.cljs$core$IFn$_invoke$arity$1 ? self__.f.cljs$core$IFn$_invoke$arity$1(val) : self__.f.call(null, val)),fn1);
}));

(cljs.core.async.t_cljs$core$async33676.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"f","f",43394975,null),new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"meta33677","meta33677",1037058364,null)], null);
}));

(cljs.core.async.t_cljs$core$async33676.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async33676.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async33676");

(cljs.core.async.t_cljs$core$async33676.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async33676");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async33676.
 */
cljs.core.async.__GT_t_cljs$core$async33676 = (function cljs$core$async$__GT_t_cljs$core$async33676(f,ch,meta33677){
return (new cljs.core.async.t_cljs$core$async33676(f,ch,meta33677));
});


/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.map_GT_ = (function cljs$core$async$map_GT_(f,ch){
return (new cljs.core.async.t_cljs$core$async33676(f,ch,cljs.core.PersistentArrayMap.EMPTY));
});

/**
* @constructor
 * @implements {cljs.core.async.impl.protocols.Channel}
 * @implements {cljs.core.async.impl.protocols.WritePort}
 * @implements {cljs.core.async.impl.protocols.ReadPort}
 * @implements {cljs.core.IMeta}
 * @implements {cljs.core.IWithMeta}
*/
cljs.core.async.t_cljs$core$async33683 = (function (p,ch,meta33684){
this.p = p;
this.ch = ch;
this.meta33684 = meta33684;
this.cljs$lang$protocol_mask$partition0$ = 393216;
this.cljs$lang$protocol_mask$partition1$ = 0;
});
(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$IWithMeta$_with_meta$arity$2 = (function (_33685,meta33684__$1){
var self__ = this;
var _33685__$1 = this;
return (new cljs.core.async.t_cljs$core$async33683(self__.p,self__.ch,meta33684__$1));
}));

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$IMeta$_meta$arity$1 = (function (_33685){
var self__ = this;
var _33685__$1 = this;
return self__.meta33684;
}));

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$Channel$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$Channel$close_BANG_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.close_BANG_(self__.ch);
}));

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$Channel$closed_QMARK_$arity$1 = (function (_){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.closed_QMARK_(self__.ch);
}));

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$ReadPort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$ReadPort$take_BANG_$arity$2 = (function (_,fn1){
var self__ = this;
var ___$1 = this;
return cljs.core.async.impl.protocols.take_BANG_(self__.ch,fn1);
}));

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$WritePort$ = cljs.core.PROTOCOL_SENTINEL);

(cljs.core.async.t_cljs$core$async33683.prototype.cljs$core$async$impl$protocols$WritePort$put_BANG_$arity$3 = (function (_,val,fn1){
var self__ = this;
var ___$1 = this;
if(cljs.core.truth_((self__.p.cljs$core$IFn$_invoke$arity$1 ? self__.p.cljs$core$IFn$_invoke$arity$1(val) : self__.p.call(null, val)))){
return cljs.core.async.impl.protocols.put_BANG_(self__.ch,val,fn1);
} else {
return cljs.core.async.impl.channels.box(cljs.core.not(cljs.core.async.impl.protocols.closed_QMARK_(self__.ch)));
}
}));

(cljs.core.async.t_cljs$core$async33683.getBasis = (function (){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Symbol(null,"p","p",1791580836,null),new cljs.core.Symbol(null,"ch","ch",1085813622,null),new cljs.core.Symbol(null,"meta33684","meta33684",-1607608484,null)], null);
}));

(cljs.core.async.t_cljs$core$async33683.cljs$lang$type = true);

(cljs.core.async.t_cljs$core$async33683.cljs$lang$ctorStr = "cljs.core.async/t_cljs$core$async33683");

(cljs.core.async.t_cljs$core$async33683.cljs$lang$ctorPrWriter = (function (this__5287__auto__,writer__5288__auto__,opt__5289__auto__){
return cljs.core._write(writer__5288__auto__,"cljs.core.async/t_cljs$core$async33683");
}));

/**
 * Positional factory function for cljs.core.async/t_cljs$core$async33683.
 */
cljs.core.async.__GT_t_cljs$core$async33683 = (function cljs$core$async$__GT_t_cljs$core$async33683(p,ch,meta33684){
return (new cljs.core.async.t_cljs$core$async33683(p,ch,meta33684));
});


/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.filter_GT_ = (function cljs$core$async$filter_GT_(p,ch){
return (new cljs.core.async.t_cljs$core$async33683(p,ch,cljs.core.PersistentArrayMap.EMPTY));
});
/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.remove_GT_ = (function cljs$core$async$remove_GT_(p,ch){
return cljs.core.async.filter_GT_(cljs.core.complement(p),ch);
});
/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.filter_LT_ = (function cljs$core$async$filter_LT_(var_args){
var G__33699 = arguments.length;
switch (G__33699) {
case 2:
return cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$2 = (function (p,ch){
return cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$3(p,ch,null);
}));

(cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$3 = (function (p,ch,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___35888 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33724){
var state_val_33725 = (state_33724[(1)]);
if((state_val_33725 === (7))){
var inst_33720 = (state_33724[(2)]);
var state_33724__$1 = state_33724;
var statearr_33726_35914 = state_33724__$1;
(statearr_33726_35914[(2)] = inst_33720);

(statearr_33726_35914[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (1))){
var state_33724__$1 = state_33724;
var statearr_33727_35915 = state_33724__$1;
(statearr_33727_35915[(2)] = null);

(statearr_33727_35915[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (4))){
var inst_33702 = (state_33724[(7)]);
var inst_33702__$1 = (state_33724[(2)]);
var inst_33703 = (inst_33702__$1 == null);
var state_33724__$1 = (function (){var statearr_33728 = state_33724;
(statearr_33728[(7)] = inst_33702__$1);

return statearr_33728;
})();
if(cljs.core.truth_(inst_33703)){
var statearr_33729_35916 = state_33724__$1;
(statearr_33729_35916[(1)] = (5));

} else {
var statearr_33730_35917 = state_33724__$1;
(statearr_33730_35917[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (6))){
var inst_33702 = (state_33724[(7)]);
var inst_33711 = (p.cljs$core$IFn$_invoke$arity$1 ? p.cljs$core$IFn$_invoke$arity$1(inst_33702) : p.call(null, inst_33702));
var state_33724__$1 = state_33724;
if(cljs.core.truth_(inst_33711)){
var statearr_33731_35918 = state_33724__$1;
(statearr_33731_35918[(1)] = (8));

} else {
var statearr_33732_35919 = state_33724__$1;
(statearr_33732_35919[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (3))){
var inst_33722 = (state_33724[(2)]);
var state_33724__$1 = state_33724;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33724__$1,inst_33722);
} else {
if((state_val_33725 === (2))){
var state_33724__$1 = state_33724;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33724__$1,(4),ch);
} else {
if((state_val_33725 === (11))){
var inst_33714 = (state_33724[(2)]);
var state_33724__$1 = state_33724;
var statearr_33736_36075 = state_33724__$1;
(statearr_33736_36075[(2)] = inst_33714);

(statearr_33736_36075[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (9))){
var state_33724__$1 = state_33724;
var statearr_33737_36080 = state_33724__$1;
(statearr_33737_36080[(2)] = null);

(statearr_33737_36080[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (5))){
var inst_33709 = cljs.core.async.close_BANG_(out);
var state_33724__$1 = state_33724;
var statearr_33738_36081 = state_33724__$1;
(statearr_33738_36081[(2)] = inst_33709);

(statearr_33738_36081[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (10))){
var inst_33717 = (state_33724[(2)]);
var state_33724__$1 = (function (){var statearr_33739 = state_33724;
(statearr_33739[(8)] = inst_33717);

return statearr_33739;
})();
var statearr_33740_36125 = state_33724__$1;
(statearr_33740_36125[(2)] = null);

(statearr_33740_36125[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33725 === (8))){
var inst_33702 = (state_33724[(7)]);
var state_33724__$1 = state_33724;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33724__$1,(11),out,inst_33702);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33741 = [null,null,null,null,null,null,null,null,null];
(statearr_33741[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33741[(1)] = (1));

return statearr_33741;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33724){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33724);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33742){var ex__30547__auto__ = e33742;
var statearr_33743_36132 = state_33724;
(statearr_33743_36132[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33724[(4)]))){
var statearr_33744_36133 = state_33724;
(statearr_33744_36133[(1)] = cljs.core.first((state_33724[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__36134 = state_33724;
state_33724 = G__36134;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33724){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33724);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33745 = f__30908__auto__();
(statearr_33745[(6)] = c__30907__auto___35888);

return statearr_33745;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.filter_LT_.cljs$lang$maxFixedArity = 3);

/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.remove_LT_ = (function cljs$core$async$remove_LT_(var_args){
var G__33752 = arguments.length;
switch (G__33752) {
case 2:
return cljs.core.async.remove_LT_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.remove_LT_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.remove_LT_.cljs$core$IFn$_invoke$arity$2 = (function (p,ch){
return cljs.core.async.remove_LT_.cljs$core$IFn$_invoke$arity$3(p,ch,null);
}));

(cljs.core.async.remove_LT_.cljs$core$IFn$_invoke$arity$3 = (function (p,ch,buf_or_n){
return cljs.core.async.filter_LT_.cljs$core$IFn$_invoke$arity$3(cljs.core.complement(p),ch,buf_or_n);
}));

(cljs.core.async.remove_LT_.cljs$lang$maxFixedArity = 3);

cljs.core.async.mapcat_STAR_ = (function cljs$core$async$mapcat_STAR_(f,in$,out){
var c__30907__auto__ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33839){
var state_val_33840 = (state_33839[(1)]);
if((state_val_33840 === (7))){
var inst_33835 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
var statearr_33843_36140 = state_33839__$1;
(statearr_33843_36140[(2)] = inst_33835);

(statearr_33843_36140[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (20))){
var inst_33803 = (state_33839[(7)]);
var inst_33816 = (state_33839[(2)]);
var inst_33817 = cljs.core.next(inst_33803);
var inst_33784 = inst_33817;
var inst_33785 = null;
var inst_33786 = (0);
var inst_33787 = (0);
var state_33839__$1 = (function (){var statearr_33844 = state_33839;
(statearr_33844[(8)] = inst_33816);

(statearr_33844[(9)] = inst_33784);

(statearr_33844[(10)] = inst_33786);

(statearr_33844[(11)] = inst_33787);

(statearr_33844[(12)] = inst_33785);

return statearr_33844;
})();
var statearr_33845_36141 = state_33839__$1;
(statearr_33845_36141[(2)] = null);

(statearr_33845_36141[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (1))){
var state_33839__$1 = state_33839;
var statearr_33846_36142 = state_33839__$1;
(statearr_33846_36142[(2)] = null);

(statearr_33846_36142[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (4))){
var inst_33773 = (state_33839[(13)]);
var inst_33773__$1 = (state_33839[(2)]);
var inst_33774 = (inst_33773__$1 == null);
var state_33839__$1 = (function (){var statearr_33847 = state_33839;
(statearr_33847[(13)] = inst_33773__$1);

return statearr_33847;
})();
if(cljs.core.truth_(inst_33774)){
var statearr_33848_36143 = state_33839__$1;
(statearr_33848_36143[(1)] = (5));

} else {
var statearr_33849_36144 = state_33839__$1;
(statearr_33849_36144[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (15))){
var state_33839__$1 = state_33839;
var statearr_33853_36145 = state_33839__$1;
(statearr_33853_36145[(2)] = null);

(statearr_33853_36145[(1)] = (16));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (21))){
var state_33839__$1 = state_33839;
var statearr_33854_36147 = state_33839__$1;
(statearr_33854_36147[(2)] = null);

(statearr_33854_36147[(1)] = (23));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (13))){
var inst_33784 = (state_33839[(9)]);
var inst_33786 = (state_33839[(10)]);
var inst_33787 = (state_33839[(11)]);
var inst_33785 = (state_33839[(12)]);
var inst_33797 = (state_33839[(2)]);
var inst_33800 = (inst_33787 + (1));
var tmp33850 = inst_33784;
var tmp33851 = inst_33786;
var tmp33852 = inst_33785;
var inst_33784__$1 = tmp33850;
var inst_33785__$1 = tmp33852;
var inst_33786__$1 = tmp33851;
var inst_33787__$1 = inst_33800;
var state_33839__$1 = (function (){var statearr_33855 = state_33839;
(statearr_33855[(9)] = inst_33784__$1);

(statearr_33855[(10)] = inst_33786__$1);

(statearr_33855[(14)] = inst_33797);

(statearr_33855[(11)] = inst_33787__$1);

(statearr_33855[(12)] = inst_33785__$1);

return statearr_33855;
})();
var statearr_33856_36153 = state_33839__$1;
(statearr_33856_36153[(2)] = null);

(statearr_33856_36153[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (22))){
var state_33839__$1 = state_33839;
var statearr_33857_36154 = state_33839__$1;
(statearr_33857_36154[(2)] = null);

(statearr_33857_36154[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (6))){
var inst_33773 = (state_33839[(13)]);
var inst_33782 = (f.cljs$core$IFn$_invoke$arity$1 ? f.cljs$core$IFn$_invoke$arity$1(inst_33773) : f.call(null, inst_33773));
var inst_33783 = cljs.core.seq(inst_33782);
var inst_33784 = inst_33783;
var inst_33785 = null;
var inst_33786 = (0);
var inst_33787 = (0);
var state_33839__$1 = (function (){var statearr_33858 = state_33839;
(statearr_33858[(9)] = inst_33784);

(statearr_33858[(10)] = inst_33786);

(statearr_33858[(11)] = inst_33787);

(statearr_33858[(12)] = inst_33785);

return statearr_33858;
})();
var statearr_33859_36162 = state_33839__$1;
(statearr_33859_36162[(2)] = null);

(statearr_33859_36162[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (17))){
var inst_33803 = (state_33839[(7)]);
var inst_33807 = cljs.core.chunk_first(inst_33803);
var inst_33808 = cljs.core.chunk_rest(inst_33803);
var inst_33809 = cljs.core.count(inst_33807);
var inst_33784 = inst_33808;
var inst_33785 = inst_33807;
var inst_33786 = inst_33809;
var inst_33787 = (0);
var state_33839__$1 = (function (){var statearr_33860 = state_33839;
(statearr_33860[(9)] = inst_33784);

(statearr_33860[(10)] = inst_33786);

(statearr_33860[(11)] = inst_33787);

(statearr_33860[(12)] = inst_33785);

return statearr_33860;
})();
var statearr_33861_36167 = state_33839__$1;
(statearr_33861_36167[(2)] = null);

(statearr_33861_36167[(1)] = (8));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (3))){
var inst_33837 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33839__$1,inst_33837);
} else {
if((state_val_33840 === (12))){
var inst_33825 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
var statearr_33862_36168 = state_33839__$1;
(statearr_33862_36168[(2)] = inst_33825);

(statearr_33862_36168[(1)] = (9));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (2))){
var state_33839__$1 = state_33839;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33839__$1,(4),in$);
} else {
if((state_val_33840 === (23))){
var inst_33833 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
var statearr_33863_36169 = state_33839__$1;
(statearr_33863_36169[(2)] = inst_33833);

(statearr_33863_36169[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (19))){
var inst_33820 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
var statearr_33864_36170 = state_33839__$1;
(statearr_33864_36170[(2)] = inst_33820);

(statearr_33864_36170[(1)] = (16));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (11))){
var inst_33784 = (state_33839[(9)]);
var inst_33803 = (state_33839[(7)]);
var inst_33803__$1 = cljs.core.seq(inst_33784);
var state_33839__$1 = (function (){var statearr_33865 = state_33839;
(statearr_33865[(7)] = inst_33803__$1);

return statearr_33865;
})();
if(inst_33803__$1){
var statearr_33866_36171 = state_33839__$1;
(statearr_33866_36171[(1)] = (14));

} else {
var statearr_33867_36172 = state_33839__$1;
(statearr_33867_36172[(1)] = (15));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (9))){
var inst_33827 = (state_33839[(2)]);
var inst_33828 = cljs.core.async.impl.protocols.closed_QMARK_(out);
var state_33839__$1 = (function (){var statearr_33868 = state_33839;
(statearr_33868[(15)] = inst_33827);

return statearr_33868;
})();
if(cljs.core.truth_(inst_33828)){
var statearr_33869_36176 = state_33839__$1;
(statearr_33869_36176[(1)] = (21));

} else {
var statearr_33870_36177 = state_33839__$1;
(statearr_33870_36177[(1)] = (22));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (5))){
var inst_33776 = cljs.core.async.close_BANG_(out);
var state_33839__$1 = state_33839;
var statearr_33871_36182 = state_33839__$1;
(statearr_33871_36182[(2)] = inst_33776);

(statearr_33871_36182[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (14))){
var inst_33803 = (state_33839[(7)]);
var inst_33805 = cljs.core.chunked_seq_QMARK_(inst_33803);
var state_33839__$1 = state_33839;
if(inst_33805){
var statearr_33872_36188 = state_33839__$1;
(statearr_33872_36188[(1)] = (17));

} else {
var statearr_33873_36189 = state_33839__$1;
(statearr_33873_36189[(1)] = (18));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (16))){
var inst_33823 = (state_33839[(2)]);
var state_33839__$1 = state_33839;
var statearr_33874_36191 = state_33839__$1;
(statearr_33874_36191[(2)] = inst_33823);

(statearr_33874_36191[(1)] = (12));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33840 === (10))){
var inst_33787 = (state_33839[(11)]);
var inst_33785 = (state_33839[(12)]);
var inst_33795 = cljs.core._nth(inst_33785,inst_33787);
var state_33839__$1 = state_33839;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33839__$1,(13),out,inst_33795);
} else {
if((state_val_33840 === (18))){
var inst_33803 = (state_33839[(7)]);
var inst_33812 = cljs.core.first(inst_33803);
var state_33839__$1 = state_33839;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33839__$1,(20),out,inst_33812);
} else {
if((state_val_33840 === (8))){
var inst_33786 = (state_33839[(10)]);
var inst_33787 = (state_33839[(11)]);
var inst_33792 = (inst_33787 < inst_33786);
var inst_33793 = inst_33792;
var state_33839__$1 = state_33839;
if(cljs.core.truth_(inst_33793)){
var statearr_33875_36198 = state_33839__$1;
(statearr_33875_36198[(1)] = (10));

} else {
var statearr_33876_36199 = state_33839__$1;
(statearr_33876_36199[(1)] = (11));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__ = null;
var cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____0 = (function (){
var statearr_33877 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_33877[(0)] = cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__);

(statearr_33877[(1)] = (1));

return statearr_33877;
});
var cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____1 = (function (state_33839){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33839);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33878){var ex__30547__auto__ = e33878;
var statearr_33879_36223 = state_33839;
(statearr_33879_36223[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33839[(4)]))){
var statearr_33880_36225 = state_33839;
(statearr_33880_36225[(1)] = cljs.core.first((state_33839[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__36244 = state_33839;
state_33839 = G__36244;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__ = function(state_33839){
switch(arguments.length){
case 0:
return cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____1.call(this,state_33839);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____0;
cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$mapcat_STAR__$_state_machine__30544__auto____1;
return cljs$core$async$mapcat_STAR__$_state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33881 = f__30908__auto__();
(statearr_33881[(6)] = c__30907__auto__);

return statearr_33881;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));

return c__30907__auto__;
});
/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.mapcat_LT_ = (function cljs$core$async$mapcat_LT_(var_args){
var G__33883 = arguments.length;
switch (G__33883) {
case 2:
return cljs.core.async.mapcat_LT_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.mapcat_LT_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.mapcat_LT_.cljs$core$IFn$_invoke$arity$2 = (function (f,in$){
return cljs.core.async.mapcat_LT_.cljs$core$IFn$_invoke$arity$3(f,in$,null);
}));

(cljs.core.async.mapcat_LT_.cljs$core$IFn$_invoke$arity$3 = (function (f,in$,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
cljs.core.async.mapcat_STAR_(f,in$,out);

return out;
}));

(cljs.core.async.mapcat_LT_.cljs$lang$maxFixedArity = 3);

/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.mapcat_GT_ = (function cljs$core$async$mapcat_GT_(var_args){
var G__33892 = arguments.length;
switch (G__33892) {
case 2:
return cljs.core.async.mapcat_GT_.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.mapcat_GT_.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.mapcat_GT_.cljs$core$IFn$_invoke$arity$2 = (function (f,out){
return cljs.core.async.mapcat_GT_.cljs$core$IFn$_invoke$arity$3(f,out,null);
}));

(cljs.core.async.mapcat_GT_.cljs$core$IFn$_invoke$arity$3 = (function (f,out,buf_or_n){
var in$ = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
cljs.core.async.mapcat_STAR_(f,in$,out);

return in$;
}));

(cljs.core.async.mapcat_GT_.cljs$lang$maxFixedArity = 3);

/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.unique = (function cljs$core$async$unique(var_args){
var G__33901 = arguments.length;
switch (G__33901) {
case 1:
return cljs.core.async.unique.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
case 2:
return cljs.core.async.unique.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.unique.cljs$core$IFn$_invoke$arity$1 = (function (ch){
return cljs.core.async.unique.cljs$core$IFn$_invoke$arity$2(ch,null);
}));

(cljs.core.async.unique.cljs$core$IFn$_invoke$arity$2 = (function (ch,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___36298 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33925){
var state_val_33926 = (state_33925[(1)]);
if((state_val_33926 === (7))){
var inst_33920 = (state_33925[(2)]);
var state_33925__$1 = state_33925;
var statearr_33931_36311 = state_33925__$1;
(statearr_33931_36311[(2)] = inst_33920);

(statearr_33931_36311[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (1))){
var inst_33902 = null;
var state_33925__$1 = (function (){var statearr_33932 = state_33925;
(statearr_33932[(7)] = inst_33902);

return statearr_33932;
})();
var statearr_33933_36312 = state_33925__$1;
(statearr_33933_36312[(2)] = null);

(statearr_33933_36312[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (4))){
var inst_33905 = (state_33925[(8)]);
var inst_33905__$1 = (state_33925[(2)]);
var inst_33906 = (inst_33905__$1 == null);
var inst_33907 = cljs.core.not(inst_33906);
var state_33925__$1 = (function (){var statearr_33934 = state_33925;
(statearr_33934[(8)] = inst_33905__$1);

return statearr_33934;
})();
if(inst_33907){
var statearr_33935_36315 = state_33925__$1;
(statearr_33935_36315[(1)] = (5));

} else {
var statearr_33936_36318 = state_33925__$1;
(statearr_33936_36318[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (6))){
var state_33925__$1 = state_33925;
var statearr_33937_36319 = state_33925__$1;
(statearr_33937_36319[(2)] = null);

(statearr_33937_36319[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (3))){
var inst_33922 = (state_33925[(2)]);
var inst_33923 = cljs.core.async.close_BANG_(out);
var state_33925__$1 = (function (){var statearr_33938 = state_33925;
(statearr_33938[(9)] = inst_33922);

return statearr_33938;
})();
return cljs.core.async.impl.ioc_helpers.return_chan(state_33925__$1,inst_33923);
} else {
if((state_val_33926 === (2))){
var state_33925__$1 = state_33925;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33925__$1,(4),ch);
} else {
if((state_val_33926 === (11))){
var inst_33905 = (state_33925[(8)]);
var inst_33914 = (state_33925[(2)]);
var inst_33902 = inst_33905;
var state_33925__$1 = (function (){var statearr_33939 = state_33925;
(statearr_33939[(10)] = inst_33914);

(statearr_33939[(7)] = inst_33902);

return statearr_33939;
})();
var statearr_33940_36341 = state_33925__$1;
(statearr_33940_36341[(2)] = null);

(statearr_33940_36341[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (9))){
var inst_33905 = (state_33925[(8)]);
var state_33925__$1 = state_33925;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33925__$1,(11),out,inst_33905);
} else {
if((state_val_33926 === (5))){
var inst_33905 = (state_33925[(8)]);
var inst_33902 = (state_33925[(7)]);
var inst_33909 = cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(inst_33905,inst_33902);
var state_33925__$1 = state_33925;
if(inst_33909){
var statearr_33942_36344 = state_33925__$1;
(statearr_33942_36344[(1)] = (8));

} else {
var statearr_33943_36345 = state_33925__$1;
(statearr_33943_36345[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (10))){
var inst_33917 = (state_33925[(2)]);
var state_33925__$1 = state_33925;
var statearr_33945_36355 = state_33925__$1;
(statearr_33945_36355[(2)] = inst_33917);

(statearr_33945_36355[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33926 === (8))){
var inst_33902 = (state_33925[(7)]);
var tmp33941 = inst_33902;
var inst_33902__$1 = tmp33941;
var state_33925__$1 = (function (){var statearr_33946 = state_33925;
(statearr_33946[(7)] = inst_33902__$1);

return statearr_33946;
})();
var statearr_33947_36358 = state_33925__$1;
(statearr_33947_36358[(2)] = null);

(statearr_33947_36358[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_33948 = [null,null,null,null,null,null,null,null,null,null,null];
(statearr_33948[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_33948[(1)] = (1));

return statearr_33948;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33925){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33925);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e33949){var ex__30547__auto__ = e33949;
var statearr_33950_36364 = state_33925;
(statearr_33950_36364[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33925[(4)]))){
var statearr_33951_36372 = state_33925;
(statearr_33951_36372[(1)] = cljs.core.first((state_33925[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__36379 = state_33925;
state_33925 = G__36379;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33925){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33925);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_33952 = f__30908__auto__();
(statearr_33952[(6)] = c__30907__auto___36298);

return statearr_33952;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.unique.cljs$lang$maxFixedArity = 2);

/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.partition = (function cljs$core$async$partition(var_args){
var G__33954 = arguments.length;
switch (G__33954) {
case 2:
return cljs.core.async.partition.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.partition.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.partition.cljs$core$IFn$_invoke$arity$2 = (function (n,ch){
return cljs.core.async.partition.cljs$core$IFn$_invoke$arity$3(n,ch,null);
}));

(cljs.core.async.partition.cljs$core$IFn$_invoke$arity$3 = (function (n,ch,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___36389 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_33992){
var state_val_33993 = (state_33992[(1)]);
if((state_val_33993 === (7))){
var inst_33988 = (state_33992[(2)]);
var state_33992__$1 = state_33992;
var statearr_33994_36390 = state_33992__$1;
(statearr_33994_36390[(2)] = inst_33988);

(statearr_33994_36390[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (1))){
var inst_33955 = (new Array(n));
var inst_33956 = inst_33955;
var inst_33957 = (0);
var state_33992__$1 = (function (){var statearr_33995 = state_33992;
(statearr_33995[(7)] = inst_33957);

(statearr_33995[(8)] = inst_33956);

return statearr_33995;
})();
var statearr_33996_36401 = state_33992__$1;
(statearr_33996_36401[(2)] = null);

(statearr_33996_36401[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (4))){
var inst_33960 = (state_33992[(9)]);
var inst_33960__$1 = (state_33992[(2)]);
var inst_33961 = (inst_33960__$1 == null);
var inst_33962 = cljs.core.not(inst_33961);
var state_33992__$1 = (function (){var statearr_33997 = state_33992;
(statearr_33997[(9)] = inst_33960__$1);

return statearr_33997;
})();
if(inst_33962){
var statearr_33998_36421 = state_33992__$1;
(statearr_33998_36421[(1)] = (5));

} else {
var statearr_33999_36422 = state_33992__$1;
(statearr_33999_36422[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (15))){
var inst_33982 = (state_33992[(2)]);
var state_33992__$1 = state_33992;
var statearr_34000_36435 = state_33992__$1;
(statearr_34000_36435[(2)] = inst_33982);

(statearr_34000_36435[(1)] = (14));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (13))){
var state_33992__$1 = state_33992;
var statearr_34001_36436 = state_33992__$1;
(statearr_34001_36436[(2)] = null);

(statearr_34001_36436[(1)] = (14));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (6))){
var inst_33957 = (state_33992[(7)]);
var inst_33978 = (inst_33957 > (0));
var state_33992__$1 = state_33992;
if(cljs.core.truth_(inst_33978)){
var statearr_34003_36437 = state_33992__$1;
(statearr_34003_36437[(1)] = (12));

} else {
var statearr_34004_36438 = state_33992__$1;
(statearr_34004_36438[(1)] = (13));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (3))){
var inst_33990 = (state_33992[(2)]);
var state_33992__$1 = state_33992;
return cljs.core.async.impl.ioc_helpers.return_chan(state_33992__$1,inst_33990);
} else {
if((state_val_33993 === (12))){
var inst_33956 = (state_33992[(8)]);
var inst_33980 = cljs.core.vec(inst_33956);
var state_33992__$1 = state_33992;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33992__$1,(15),out,inst_33980);
} else {
if((state_val_33993 === (2))){
var state_33992__$1 = state_33992;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_33992__$1,(4),ch);
} else {
if((state_val_33993 === (11))){
var inst_33972 = (state_33992[(2)]);
var inst_33973 = (new Array(n));
var inst_33956 = inst_33973;
var inst_33957 = (0);
var state_33992__$1 = (function (){var statearr_34005 = state_33992;
(statearr_34005[(7)] = inst_33957);

(statearr_34005[(8)] = inst_33956);

(statearr_34005[(10)] = inst_33972);

return statearr_34005;
})();
var statearr_34006_36441 = state_33992__$1;
(statearr_34006_36441[(2)] = null);

(statearr_34006_36441[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (9))){
var inst_33956 = (state_33992[(8)]);
var inst_33970 = cljs.core.vec(inst_33956);
var state_33992__$1 = state_33992;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_33992__$1,(11),out,inst_33970);
} else {
if((state_val_33993 === (5))){
var inst_33960 = (state_33992[(9)]);
var inst_33957 = (state_33992[(7)]);
var inst_33956 = (state_33992[(8)]);
var inst_33965 = (state_33992[(11)]);
var inst_33964 = (inst_33956[inst_33957] = inst_33960);
var inst_33965__$1 = (inst_33957 + (1));
var inst_33966 = (inst_33965__$1 < n);
var state_33992__$1 = (function (){var statearr_34007 = state_33992;
(statearr_34007[(11)] = inst_33965__$1);

(statearr_34007[(12)] = inst_33964);

return statearr_34007;
})();
if(cljs.core.truth_(inst_33966)){
var statearr_34008_36446 = state_33992__$1;
(statearr_34008_36446[(1)] = (8));

} else {
var statearr_34009_36447 = state_33992__$1;
(statearr_34009_36447[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (14))){
var inst_33985 = (state_33992[(2)]);
var inst_33986 = cljs.core.async.close_BANG_(out);
var state_33992__$1 = (function (){var statearr_34011 = state_33992;
(statearr_34011[(13)] = inst_33985);

return statearr_34011;
})();
var statearr_34012_36453 = state_33992__$1;
(statearr_34012_36453[(2)] = inst_33986);

(statearr_34012_36453[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (10))){
var inst_33976 = (state_33992[(2)]);
var state_33992__$1 = state_33992;
var statearr_34016_36456 = state_33992__$1;
(statearr_34016_36456[(2)] = inst_33976);

(statearr_34016_36456[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_33993 === (8))){
var inst_33956 = (state_33992[(8)]);
var inst_33965 = (state_33992[(11)]);
var tmp34010 = inst_33956;
var inst_33956__$1 = tmp34010;
var inst_33957 = inst_33965;
var state_33992__$1 = (function (){var statearr_34025 = state_33992;
(statearr_34025[(7)] = inst_33957);

(statearr_34025[(8)] = inst_33956__$1);

return statearr_34025;
})();
var statearr_34032_36459 = state_33992__$1;
(statearr_34032_36459[(2)] = null);

(statearr_34032_36459[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_34033 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_34033[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_34033[(1)] = (1));

return statearr_34033;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_33992){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_33992);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e34040){var ex__30547__auto__ = e34040;
var statearr_34041_36460 = state_33992;
(statearr_34041_36460[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_33992[(4)]))){
var statearr_34042_36461 = state_33992;
(statearr_34042_36461[(1)] = cljs.core.first((state_33992[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__36462 = state_33992;
state_33992 = G__36462;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_33992){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_33992);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_34043 = f__30908__auto__();
(statearr_34043[(6)] = c__30907__auto___36389);

return statearr_34043;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.partition.cljs$lang$maxFixedArity = 3);

/**
 * Deprecated - this function will be removed. Use transducer instead
 */
cljs.core.async.partition_by = (function cljs$core$async$partition_by(var_args){
var G__34045 = arguments.length;
switch (G__34045) {
case 2:
return cljs.core.async.partition_by.cljs$core$IFn$_invoke$arity$2((arguments[(0)]),(arguments[(1)]));

break;
case 3:
return cljs.core.async.partition_by.cljs$core$IFn$_invoke$arity$3((arguments[(0)]),(arguments[(1)]),(arguments[(2)]));

break;
default:
throw (new Error(["Invalid arity: ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(arguments.length)].join('')));

}
});

(cljs.core.async.partition_by.cljs$core$IFn$_invoke$arity$2 = (function (f,ch){
return cljs.core.async.partition_by.cljs$core$IFn$_invoke$arity$3(f,ch,null);
}));

(cljs.core.async.partition_by.cljs$core$IFn$_invoke$arity$3 = (function (f,ch,buf_or_n){
var out = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1(buf_or_n);
var c__30907__auto___36472 = cljs.core.async.chan.cljs$core$IFn$_invoke$arity$1((1));
cljs.core.async.impl.dispatch.run((function (){
var f__30908__auto__ = (function (){var switch__30543__auto__ = (function (state_34092){
var state_val_34093 = (state_34092[(1)]);
if((state_val_34093 === (7))){
var inst_34088 = (state_34092[(2)]);
var state_34092__$1 = state_34092;
var statearr_34095_36485 = state_34092__$1;
(statearr_34095_36485[(2)] = inst_34088);

(statearr_34095_36485[(1)] = (3));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (1))){
var inst_34046 = [];
var inst_34047 = inst_34046;
var inst_34048 = new cljs.core.Keyword("cljs.core.async","nothing","cljs.core.async/nothing",-69252123);
var state_34092__$1 = (function (){var statearr_34096 = state_34092;
(statearr_34096[(7)] = inst_34047);

(statearr_34096[(8)] = inst_34048);

return statearr_34096;
})();
var statearr_34097_36496 = state_34092__$1;
(statearr_34097_36496[(2)] = null);

(statearr_34097_36496[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (4))){
var inst_34053 = (state_34092[(9)]);
var inst_34053__$1 = (state_34092[(2)]);
var inst_34054 = (inst_34053__$1 == null);
var inst_34055 = cljs.core.not(inst_34054);
var state_34092__$1 = (function (){var statearr_34107 = state_34092;
(statearr_34107[(9)] = inst_34053__$1);

return statearr_34107;
})();
if(inst_34055){
var statearr_34109_36499 = state_34092__$1;
(statearr_34109_36499[(1)] = (5));

} else {
var statearr_34110_36501 = state_34092__$1;
(statearr_34110_36501[(1)] = (6));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (15))){
var inst_34047 = (state_34092[(7)]);
var inst_34080 = cljs.core.vec(inst_34047);
var state_34092__$1 = state_34092;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_34092__$1,(18),out,inst_34080);
} else {
if((state_val_34093 === (13))){
var inst_34075 = (state_34092[(2)]);
var state_34092__$1 = state_34092;
var statearr_34127_36503 = state_34092__$1;
(statearr_34127_36503[(2)] = inst_34075);

(statearr_34127_36503[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (6))){
var inst_34047 = (state_34092[(7)]);
var inst_34077 = inst_34047.length;
var inst_34078 = (inst_34077 > (0));
var state_34092__$1 = state_34092;
if(cljs.core.truth_(inst_34078)){
var statearr_34128_36508 = state_34092__$1;
(statearr_34128_36508[(1)] = (15));

} else {
var statearr_34129_36512 = state_34092__$1;
(statearr_34129_36512[(1)] = (16));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (17))){
var inst_34085 = (state_34092[(2)]);
var inst_34086 = cljs.core.async.close_BANG_(out);
var state_34092__$1 = (function (){var statearr_34131 = state_34092;
(statearr_34131[(10)] = inst_34085);

return statearr_34131;
})();
var statearr_34132_36536 = state_34092__$1;
(statearr_34132_36536[(2)] = inst_34086);

(statearr_34132_36536[(1)] = (7));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (3))){
var inst_34090 = (state_34092[(2)]);
var state_34092__$1 = state_34092;
return cljs.core.async.impl.ioc_helpers.return_chan(state_34092__$1,inst_34090);
} else {
if((state_val_34093 === (12))){
var inst_34047 = (state_34092[(7)]);
var inst_34068 = cljs.core.vec(inst_34047);
var state_34092__$1 = state_34092;
return cljs.core.async.impl.ioc_helpers.put_BANG_(state_34092__$1,(14),out,inst_34068);
} else {
if((state_val_34093 === (2))){
var state_34092__$1 = state_34092;
return cljs.core.async.impl.ioc_helpers.take_BANG_(state_34092__$1,(4),ch);
} else {
if((state_val_34093 === (11))){
var inst_34047 = (state_34092[(7)]);
var inst_34057 = (state_34092[(11)]);
var inst_34053 = (state_34092[(9)]);
var inst_34065 = inst_34047.push(inst_34053);
var tmp34133 = inst_34047;
var inst_34047__$1 = tmp34133;
var inst_34048 = inst_34057;
var state_34092__$1 = (function (){var statearr_34134 = state_34092;
(statearr_34134[(7)] = inst_34047__$1);

(statearr_34134[(8)] = inst_34048);

(statearr_34134[(12)] = inst_34065);

return statearr_34134;
})();
var statearr_34135_36568 = state_34092__$1;
(statearr_34135_36568[(2)] = null);

(statearr_34135_36568[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (9))){
var inst_34048 = (state_34092[(8)]);
var inst_34061 = cljs.core.keyword_identical_QMARK_(inst_34048,new cljs.core.Keyword("cljs.core.async","nothing","cljs.core.async/nothing",-69252123));
var state_34092__$1 = state_34092;
var statearr_34137_36569 = state_34092__$1;
(statearr_34137_36569[(2)] = inst_34061);

(statearr_34137_36569[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (5))){
var inst_34057 = (state_34092[(11)]);
var inst_34048 = (state_34092[(8)]);
var inst_34058 = (state_34092[(13)]);
var inst_34053 = (state_34092[(9)]);
var inst_34057__$1 = (f.cljs$core$IFn$_invoke$arity$1 ? f.cljs$core$IFn$_invoke$arity$1(inst_34053) : f.call(null, inst_34053));
var inst_34058__$1 = cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(inst_34057__$1,inst_34048);
var state_34092__$1 = (function (){var statearr_34139 = state_34092;
(statearr_34139[(11)] = inst_34057__$1);

(statearr_34139[(13)] = inst_34058__$1);

return statearr_34139;
})();
if(inst_34058__$1){
var statearr_34143_36570 = state_34092__$1;
(statearr_34143_36570[(1)] = (8));

} else {
var statearr_34146_36571 = state_34092__$1;
(statearr_34146_36571[(1)] = (9));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (14))){
var inst_34057 = (state_34092[(11)]);
var inst_34053 = (state_34092[(9)]);
var inst_34070 = (state_34092[(2)]);
var inst_34071 = [];
var inst_34072 = inst_34071.push(inst_34053);
var inst_34047 = inst_34071;
var inst_34048 = inst_34057;
var state_34092__$1 = (function (){var statearr_34183 = state_34092;
(statearr_34183[(7)] = inst_34047);

(statearr_34183[(14)] = inst_34072);

(statearr_34183[(15)] = inst_34070);

(statearr_34183[(8)] = inst_34048);

return statearr_34183;
})();
var statearr_34187_36574 = state_34092__$1;
(statearr_34187_36574[(2)] = null);

(statearr_34187_36574[(1)] = (2));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (16))){
var state_34092__$1 = state_34092;
var statearr_34188_36577 = state_34092__$1;
(statearr_34188_36577[(2)] = null);

(statearr_34188_36577[(1)] = (17));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (10))){
var inst_34063 = (state_34092[(2)]);
var state_34092__$1 = state_34092;
if(cljs.core.truth_(inst_34063)){
var statearr_34195_36578 = state_34092__$1;
(statearr_34195_36578[(1)] = (11));

} else {
var statearr_34200_36579 = state_34092__$1;
(statearr_34200_36579[(1)] = (12));

}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (18))){
var inst_34082 = (state_34092[(2)]);
var state_34092__$1 = state_34092;
var statearr_34204_36584 = state_34092__$1;
(statearr_34204_36584[(2)] = inst_34082);

(statearr_34204_36584[(1)] = (17));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
if((state_val_34093 === (8))){
var inst_34058 = (state_34092[(13)]);
var state_34092__$1 = state_34092;
var statearr_34211_36593 = state_34092__$1;
(statearr_34211_36593[(2)] = inst_34058);

(statearr_34211_36593[(1)] = (10));


return new cljs.core.Keyword(null,"recur","recur",-437573268);
} else {
return null;
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
});
return (function() {
var cljs$core$async$state_machine__30544__auto__ = null;
var cljs$core$async$state_machine__30544__auto____0 = (function (){
var statearr_34212 = [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null];
(statearr_34212[(0)] = cljs$core$async$state_machine__30544__auto__);

(statearr_34212[(1)] = (1));

return statearr_34212;
});
var cljs$core$async$state_machine__30544__auto____1 = (function (state_34092){
while(true){
var ret_value__30545__auto__ = (function (){try{while(true){
var result__30546__auto__ = switch__30543__auto__(state_34092);
if(cljs.core.keyword_identical_QMARK_(result__30546__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
continue;
} else {
return result__30546__auto__;
}
break;
}
}catch (e34213){var ex__30547__auto__ = e34213;
var statearr_34214_36613 = state_34092;
(statearr_34214_36613[(2)] = ex__30547__auto__);


if(cljs.core.seq((state_34092[(4)]))){
var statearr_34222_36620 = state_34092;
(statearr_34222_36620[(1)] = cljs.core.first((state_34092[(4)])));

} else {
throw ex__30547__auto__;
}

return new cljs.core.Keyword(null,"recur","recur",-437573268);
}})();
if(cljs.core.keyword_identical_QMARK_(ret_value__30545__auto__,new cljs.core.Keyword(null,"recur","recur",-437573268))){
var G__36621 = state_34092;
state_34092 = G__36621;
continue;
} else {
return ret_value__30545__auto__;
}
break;
}
});
cljs$core$async$state_machine__30544__auto__ = function(state_34092){
switch(arguments.length){
case 0:
return cljs$core$async$state_machine__30544__auto____0.call(this);
case 1:
return cljs$core$async$state_machine__30544__auto____1.call(this,state_34092);
}
throw(new Error('Invalid arity: ' + arguments.length));
};
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$0 = cljs$core$async$state_machine__30544__auto____0;
cljs$core$async$state_machine__30544__auto__.cljs$core$IFn$_invoke$arity$1 = cljs$core$async$state_machine__30544__auto____1;
return cljs$core$async$state_machine__30544__auto__;
})()
})();
var state__30909__auto__ = (function (){var statearr_34226 = f__30908__auto__();
(statearr_34226[(6)] = c__30907__auto___36472);

return statearr_34226;
})();
return cljs.core.async.impl.ioc_helpers.run_state_machine_wrapped(state__30909__auto__);
}));


return out;
}));

(cljs.core.async.partition_by.cljs$lang$maxFixedArity = 3);


//# sourceMappingURL=cljs.core.async.js.map
