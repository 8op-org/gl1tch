goog.provide('gl1tch.site.diagrams.flowchart');
/**
 * Parses inline diagram children from content EDN into a flowchart spec.
 * Input attrs: {:type :flowchart :direction :lr}
 * Input children: [[:node :a "Start"] [:node :b "End"] [:edge :a :b]]
 * Returns: {:direction :lr :nodes [...] :edges [...]}
 */
gl1tch.site.diagrams.flowchart.parse_inline = (function gl1tch$site$diagrams$flowchart$parse_inline(attrs,children){
var nodes = cljs.core.filterv((function (p1__21321_SHARP_){
return cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"node","node",581201198),cljs.core.first(p1__21321_SHARP_));
}),children);
var edges = cljs.core.filterv((function (p1__21322_SHARP_){
return cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"edge","edge",919909153),cljs.core.first(p1__21322_SHARP_));
}),children);
return new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"direction","direction",-633359395),(function (){var or__5002__auto__ = new cljs.core.Keyword(null,"direction","direction",-633359395).cljs$core$IFn$_invoke$arity$1(attrs);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return new cljs.core.Keyword(null,"lr","lr",445647393);
}
})(),new cljs.core.Keyword(null,"nodes","nodes",-2099585805),cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (p__21324){
var vec__21326 = p__21324;
var _ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21326,(0),null);
var id = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21326,(1),null);
var label = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21326,(2),null);
return new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"id","id",-1388402092),id,new cljs.core.Keyword(null,"label","label",1718410804),label], null);
}),nodes),new cljs.core.Keyword(null,"edges","edges",-694791395),cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (p__21329){
var vec__21330 = p__21329;
var _ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21330,(0),null);
var from = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21330,(1),null);
var to = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21330,(2),null);
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [from,to], null);
}),edges)], null);
});
/**
 * Renders a flowchart spec as hiccup SVG.
 */
gl1tch.site.diagrams.flowchart.render = (function gl1tch$site$diagrams$flowchart$render(spec){
var map__21333 = gl1tch.site.diagrams.core.compute(spec);
var map__21333__$1 = cljs.core.__destructure_map(map__21333);
var nodes = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21333__$1,new cljs.core.Keyword(null,"nodes","nodes",-2099585805));
var edges = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21333__$1,new cljs.core.Keyword(null,"edges","edges",-694791395));
var width = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21333__$1,new cljs.core.Keyword(null,"width","width",-384071477));
var height = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21333__$1,new cljs.core.Keyword(null,"height","height",1025178622));
return new cljs.core.PersistentVector(null, 5, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"svg","svg",856789142),new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"width","width",-384071477),width,new cljs.core.Keyword(null,"height","height",1025178622),height,new cljs.core.Keyword(null,"viewBox","viewBox",-469489477),["0 0 ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(width)," ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(height)].join(''),new cljs.core.Keyword(null,"xmlns","xmlns",-1862095571),"http://www.w3.org/2000/svg",new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"block","block",664686210),new cljs.core.Keyword(null,"margin","margin",-995903681),"16px 0"], null)], null),new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrowhead_marker], null),(function (){var iter__5480__auto__ = (function gl1tch$site$diagrams$flowchart$render_$_iter__21334(s__21335){
return (new cljs.core.LazySeq(null,(function (){
var s__21335__$1 = s__21335;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__21335__$1);
if(temp__5825__auto__){
var s__21335__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__21335__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__21335__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__21337 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__21336 = (0);
while(true){
if((i__21336 < size__5479__auto__)){
var node = cljs.core._nth(c__5478__auto__,i__21336);
cljs.core.chunk_append(b__21337,cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.box,node], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"id","id",-1388402092).cljs$core$IFn$_invoke$arity$1(node)], null)));

var G__21346 = (i__21336 + (1));
i__21336 = G__21346;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__21337),gl1tch$site$diagrams$flowchart$render_$_iter__21334(cljs.core.chunk_rest(s__21335__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__21337),null);
}
} else {
var node = cljs.core.first(s__21335__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.box,node], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"id","id",-1388402092).cljs$core$IFn$_invoke$arity$1(node)], null)),gl1tch$site$diagrams$flowchart$render_$_iter__21334(cljs.core.rest(s__21335__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(nodes);
})(),(function (){var iter__5480__auto__ = (function gl1tch$site$diagrams$flowchart$render_$_iter__21338(s__21339){
return (new cljs.core.LazySeq(null,(function (){
var s__21339__$1 = s__21339;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__21339__$1);
if(temp__5825__auto__){
var s__21339__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__21339__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__21339__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__21341 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__21340 = (0);
while(true){
if((i__21340 < size__5479__auto__)){
var edge = cljs.core._nth(c__5478__auto__,i__21340);
cljs.core.chunk_append(b__21341,cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrow,edge], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"from","from",1815293044).cljs$core$IFn$_invoke$arity$1(edge)),"-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"to","to",192099007).cljs$core$IFn$_invoke$arity$1(edge))].join('')], null)));

var G__21349 = (i__21340 + (1));
i__21340 = G__21349;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__21341),gl1tch$site$diagrams$flowchart$render_$_iter__21338(cljs.core.chunk_rest(s__21339__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__21341),null);
}
} else {
var edge = cljs.core.first(s__21339__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrow,edge], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"from","from",1815293044).cljs$core$IFn$_invoke$arity$1(edge)),"-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"to","to",192099007).cljs$core$IFn$_invoke$arity$1(edge))].join('')], null)),gl1tch$site$diagrams$flowchart$render_$_iter__21338(cljs.core.rest(s__21339__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(edges);
})()], null);
});

//# sourceMappingURL=gl1tch.site.diagrams.flowchart.js.map
