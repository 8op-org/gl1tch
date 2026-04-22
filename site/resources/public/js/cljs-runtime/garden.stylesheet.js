goog.provide('garden.stylesheet');
/**
 * Create a rule function for the given selector. The `selector`
 *   argument must be valid selector (ie. a keyword, string, or symbol).
 *   Additional arguments may consist of extra selectors or
 *   declarations.
 * 
 *   The returned function accepts any number of arguments which represent
 *   the rule's children.
 * 
 *   Ex.
 *    (let [text-field (rule "[type="text"])]
 *     (text-field {:border ["1px" :solid "black"]}))
 *    ;; => ["[type="text"] {:boder ["1px" :solid "black"]}]
 */
garden.stylesheet.rule = (function garden$stylesheet$rule(var_args){
var args__5732__auto__ = [];
var len__5726__auto___41422 = arguments.length;
var i__5727__auto___41423 = (0);
while(true){
if((i__5727__auto___41423 < len__5726__auto___41422)){
args__5732__auto__.push((arguments[i__5727__auto___41423]));

var G__41424 = (i__5727__auto___41423 + (1));
i__5727__auto___41423 = G__41424;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.stylesheet.rule.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.stylesheet.rule.cljs$core$IFn$_invoke$arity$variadic = (function (selector,more){
if((!((((selector instanceof cljs.core.Keyword)) || (((typeof selector === 'string') || ((selector instanceof cljs.core.Symbol)))))))){
throw cljs.core.ex_info.cljs$core$IFn$_invoke$arity$2("Selector must be either a keyword, string, or symbol.",cljs.core.PersistentArrayMap.EMPTY);
} else {
return (function() { 
var G__41430__delegate = function (children){
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.apply.cljs$core$IFn$_invoke$arity$3(cljs.core.vector,selector,more),children);
};
var G__41430 = function (var_args){
var children = null;
if (arguments.length > 0) {
var G__41431__i = 0, G__41431__a = new Array(arguments.length -  0);
while (G__41431__i < G__41431__a.length) {G__41431__a[G__41431__i] = arguments[G__41431__i + 0]; ++G__41431__i;}
  children = new cljs.core.IndexedSeq(G__41431__a,0,null);
} 
return G__41430__delegate.call(this,children);};
G__41430.cljs$lang$maxFixedArity = 0;
G__41430.cljs$lang$applyTo = (function (arglist__41432){
var children = cljs.core.seq(arglist__41432);
return G__41430__delegate(children);
});
G__41430.cljs$core$IFn$_invoke$arity$variadic = G__41430__delegate;
return G__41430;
})()
;
}
}));

(garden.stylesheet.rule.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.stylesheet.rule.cljs$lang$applyTo = (function (seq41408){
var G__41409 = cljs.core.first(seq41408);
var seq41408__$1 = cljs.core.next(seq41408);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__41409,seq41408__$1);
}));

garden.stylesheet.cssfn = (function garden$stylesheet$cssfn(fn_name){
return (function() { 
var G__41433__delegate = function (args){
return (new garden.types.CSSFunction(fn_name,args,null,null,null));
};
var G__41433 = function (var_args){
var args = null;
if (arguments.length > 0) {
var G__41434__i = 0, G__41434__a = new Array(arguments.length -  0);
while (G__41434__i < G__41434__a.length) {G__41434__a[G__41434__i] = arguments[G__41434__i + 0]; ++G__41434__i;}
  args = new cljs.core.IndexedSeq(G__41434__a,0,null);
} 
return G__41433__delegate.call(this,args);};
G__41433.cljs$lang$maxFixedArity = 0;
G__41433.cljs$lang$applyTo = (function (arglist__41435){
var args = cljs.core.seq(arglist__41435);
return G__41433__delegate(args);
});
G__41433.cljs$core$IFn$_invoke$arity$variadic = G__41433__delegate;
return G__41433;
})()
;
});
garden.stylesheet.at_rule = (function garden$stylesheet$at_rule(identifier,value){
return (new garden.types.CSSAtRule(identifier,value,null,null,null));
});
/**
 * Create a CSS @font-face rule.
 */
garden.stylesheet.at_font_face = (function garden$stylesheet$at_font_face(var_args){
var args__5732__auto__ = [];
var len__5726__auto___41436 = arguments.length;
var i__5727__auto___41437 = (0);
while(true){
if((i__5727__auto___41437 < len__5726__auto___41436)){
args__5732__auto__.push((arguments[i__5727__auto___41437]));

var G__41438 = (i__5727__auto___41437 + (1));
i__5727__auto___41437 = G__41438;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((0) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((0)),(0),null)):null);
return garden.stylesheet.at_font_face.cljs$core$IFn$_invoke$arity$variadic(argseq__5733__auto__);
});

(garden.stylesheet.at_font_face.cljs$core$IFn$_invoke$arity$variadic = (function (font_properties){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, ["@font-face",font_properties], null);
}));

(garden.stylesheet.at_font_face.cljs$lang$maxFixedArity = (0));

/** @this {Function} */
(garden.stylesheet.at_font_face.cljs$lang$applyTo = (function (seq41410){
var self__5712__auto__ = this;
return self__5712__auto__.cljs$core$IFn$_invoke$arity$variadic(cljs.core.seq(seq41410));
}));

/**
 * Create a CSS @import rule.
 */
garden.stylesheet.at_import = (function garden$stylesheet$at_import(var_args){
var G__41415 = arguments.length;
switch (G__41415) {
case 1:
return garden.stylesheet.at_import.cljs$core$IFn$_invoke$arity$1((arguments[(0)]));

break;
default:
var args_arr__5751__auto__ = [];
var len__5726__auto___41440 = arguments.length;
var i__5727__auto___41441 = (0);
while(true){
if((i__5727__auto___41441 < len__5726__auto___41440)){
args_arr__5751__auto__.push((arguments[i__5727__auto___41441]));

var G__41442 = (i__5727__auto___41441 + (1));
i__5727__auto___41441 = G__41442;
continue;
} else {
}
break;
}

var argseq__5752__auto__ = ((((1) < args_arr__5751__auto__.length))?(new cljs.core.IndexedSeq(args_arr__5751__auto__.slice((1)),(0),null)):null);
return garden.stylesheet.at_import.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5752__auto__);

}
});

(garden.stylesheet.at_import.cljs$core$IFn$_invoke$arity$1 = (function (url){
return garden.stylesheet.at_rule(new cljs.core.Keyword(null,"import","import",-1399500709),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"url","url",276297046),url,new cljs.core.Keyword(null,"media-queries","media-queries",-1563277678),null], null));
}));

(garden.stylesheet.at_import.cljs$core$IFn$_invoke$arity$variadic = (function (url,media_queries){
return garden.stylesheet.at_rule(new cljs.core.Keyword(null,"import","import",-1399500709),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"url","url",276297046),url,new cljs.core.Keyword(null,"media-queries","media-queries",-1563277678),media_queries], null));
}));

/** @this {Function} */
(garden.stylesheet.at_import.cljs$lang$applyTo = (function (seq41413){
var G__41414 = cljs.core.first(seq41413);
var seq41413__$1 = cljs.core.next(seq41413);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__41414,seq41413__$1);
}));

(garden.stylesheet.at_import.cljs$lang$maxFixedArity = (1));

/**
 * Create a CSS @media rule.
 */
garden.stylesheet.at_media = (function garden$stylesheet$at_media(var_args){
var args__5732__auto__ = [];
var len__5726__auto___41443 = arguments.length;
var i__5727__auto___41444 = (0);
while(true){
if((i__5727__auto___41444 < len__5726__auto___41443)){
args__5732__auto__.push((arguments[i__5727__auto___41444]));

var G__41445 = (i__5727__auto___41444 + (1));
i__5727__auto___41444 = G__41445;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.stylesheet.at_media.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.stylesheet.at_media.cljs$core$IFn$_invoke$arity$variadic = (function (media_queries,rules){
return garden.stylesheet.at_rule(new cljs.core.Keyword(null,"media","media",-1066138403),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"media-queries","media-queries",-1563277678),media_queries,new cljs.core.Keyword(null,"rules","rules",1198912366),rules], null));
}));

(garden.stylesheet.at_media.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.stylesheet.at_media.cljs$lang$applyTo = (function (seq41416){
var G__41417 = cljs.core.first(seq41416);
var seq41416__$1 = cljs.core.next(seq41416);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__41417,seq41416__$1);
}));

garden.stylesheet.at_supports = (function garden$stylesheet$at_supports(var_args){
var args__5732__auto__ = [];
var len__5726__auto___41446 = arguments.length;
var i__5727__auto___41447 = (0);
while(true){
if((i__5727__auto___41447 < len__5726__auto___41446)){
args__5732__auto__.push((arguments[i__5727__auto___41447]));

var G__41448 = (i__5727__auto___41447 + (1));
i__5727__auto___41447 = G__41448;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.stylesheet.at_supports.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.stylesheet.at_supports.cljs$core$IFn$_invoke$arity$variadic = (function (feature_queries,rules){

return garden.stylesheet.at_rule(new cljs.core.Keyword(null,"feature","feature",27242652),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"feature-queries","feature-queries",-1340998408),feature_queries,new cljs.core.Keyword(null,"rules","rules",1198912366),rules], null));
}));

(garden.stylesheet.at_supports.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.stylesheet.at_supports.cljs$lang$applyTo = (function (seq41418){
var G__41419 = cljs.core.first(seq41418);
var seq41418__$1 = cljs.core.next(seq41418);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__41419,seq41418__$1);
}));

/**
 * Create a CSS @keyframes rule.
 */
garden.stylesheet.at_keyframes = (function garden$stylesheet$at_keyframes(var_args){
var args__5732__auto__ = [];
var len__5726__auto___41449 = arguments.length;
var i__5727__auto___41450 = (0);
while(true){
if((i__5727__auto___41450 < len__5726__auto___41449)){
args__5732__auto__.push((arguments[i__5727__auto___41450]));

var G__41451 = (i__5727__auto___41450 + (1));
i__5727__auto___41450 = G__41451;
continue;
} else {
}
break;
}

var argseq__5733__auto__ = ((((1) < args__5732__auto__.length))?(new cljs.core.IndexedSeq(args__5732__auto__.slice((1)),(0),null)):null);
return garden.stylesheet.at_keyframes.cljs$core$IFn$_invoke$arity$variadic((arguments[(0)]),argseq__5733__auto__);
});

(garden.stylesheet.at_keyframes.cljs$core$IFn$_invoke$arity$variadic = (function (identifier,frames){
return garden.stylesheet.at_rule(new cljs.core.Keyword(null,"keyframes","keyframes",-1437976012),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"identifier","identifier",-805503498),identifier,new cljs.core.Keyword(null,"frames","frames",1765687497),frames], null));
}));

(garden.stylesheet.at_keyframes.cljs$lang$maxFixedArity = (1));

/** @this {Function} */
(garden.stylesheet.at_keyframes.cljs$lang$applyTo = (function (seq41420){
var G__41421 = cljs.core.first(seq41420);
var seq41420__$1 = cljs.core.next(seq41420);
var self__5711__auto__ = this;
return self__5711__auto__.cljs$core$IFn$_invoke$arity$variadic(G__41421,seq41420__$1);
}));

/**
 * Create a color from RGB values.
 */
garden.stylesheet.rgb = (function garden$stylesheet$rgb(r,g,b){
return garden.color.rgb.cljs$core$IFn$_invoke$arity$1(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [r,g,b], null));
});
/**
 * Create a color from HSL values.
 */
garden.stylesheet.hsl = (function garden$stylesheet$hsl(h,s,l){
return garden.color.hsl.cljs$core$IFn$_invoke$arity$1(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [h,s,l], null));
});

//# sourceMappingURL=garden.stylesheet.js.map
