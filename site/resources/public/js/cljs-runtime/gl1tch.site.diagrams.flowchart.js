goog.provide('gl1tch.site.diagrams.flowchart');
/**
 * Parses inline diagram children from content EDN into a flowchart spec.
 * Input attrs: {:type :flowchart :direction :lr}
 * Input children: [[:node :a "Start"] [:node :b "End"] [:edge :a :b]]
 * Returns: {:direction :lr :nodes [...] :edges [...]}
 */
gl1tch.site.diagrams.flowchart.parse_inline = (function gl1tch$site$diagrams$flowchart$parse_inline(attrs,children){
var nodes = cljs.core.filterv((function (p1__41719_SHARP_){
return cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"node","node",581201198),cljs.core.first(p1__41719_SHARP_));
}),children);
var edges = cljs.core.filterv((function (p1__41720_SHARP_){
return cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"edge","edge",919909153),cljs.core.first(p1__41720_SHARP_));
}),children);
return new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"direction","direction",-633359395),(function (){var or__5002__auto__ = new cljs.core.Keyword(null,"direction","direction",-633359395).cljs$core$IFn$_invoke$arity$1(attrs);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return new cljs.core.Keyword(null,"lr","lr",445647393);
}
})(),new cljs.core.Keyword(null,"nodes","nodes",-2099585805),cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (p__41727){
var vec__41728 = p__41727;
var _ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41728,(0),null);
var id = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41728,(1),null);
var label = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41728,(2),null);
return new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"id","id",-1388402092),id,new cljs.core.Keyword(null,"label","label",1718410804),label], null);
}),nodes),new cljs.core.Keyword(null,"edges","edges",-694791395),cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (p__41731){
var vec__41732 = p__41731;
var _ = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41732,(0),null);
var from = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41732,(1),null);
var to = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41732,(2),null);
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [from,to], null);
}),edges)], null);
});
/**
 * Renders a flowchart spec as hiccup SVG.
 */
gl1tch.site.diagrams.flowchart.render = (function gl1tch$site$diagrams$flowchart$render(spec){
var map__41735 = gl1tch.site.diagrams.core.compute(spec);
var map__41735__$1 = cljs.core.__destructure_map(map__41735);
var nodes = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41735__$1,new cljs.core.Keyword(null,"nodes","nodes",-2099585805));
var edges = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41735__$1,new cljs.core.Keyword(null,"edges","edges",-694791395));
var width = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41735__$1,new cljs.core.Keyword(null,"width","width",-384071477));
var height = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41735__$1,new cljs.core.Keyword(null,"height","height",1025178622));
return new cljs.core.PersistentVector(null, 5, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"svg","svg",856789142),new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"width","width",-384071477),width,new cljs.core.Keyword(null,"height","height",1025178622),height,new cljs.core.Keyword(null,"viewBox","viewBox",-469489477),["0 0 ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(width)," ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(height)].join(''),new cljs.core.Keyword(null,"xmlns","xmlns",-1862095571),"http://www.w3.org/2000/svg",new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"block","block",664686210),new cljs.core.Keyword(null,"margin","margin",-995903681),"16px 0"], null)], null),new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrowhead_marker], null),(function (){var iter__5480__auto__ = (function gl1tch$site$diagrams$flowchart$render_$_iter__41736(s__41737){
return (new cljs.core.LazySeq(null,(function (){
var s__41737__$1 = s__41737;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__41737__$1);
if(temp__5825__auto__){
var s__41737__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__41737__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__41737__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__41739 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__41738 = (0);
while(true){
if((i__41738 < size__5479__auto__)){
var node = cljs.core._nth(c__5478__auto__,i__41738);
cljs.core.chunk_append(b__41739,cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.box,node], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"id","id",-1388402092).cljs$core$IFn$_invoke$arity$1(node)], null)));

var G__41746 = (i__41738 + (1));
i__41738 = G__41746;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41739),gl1tch$site$diagrams$flowchart$render_$_iter__41736(cljs.core.chunk_rest(s__41737__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41739),null);
}
} else {
var node = cljs.core.first(s__41737__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.box,node], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"id","id",-1388402092).cljs$core$IFn$_invoke$arity$1(node)], null)),gl1tch$site$diagrams$flowchart$render_$_iter__41736(cljs.core.rest(s__41737__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(nodes);
})(),(function (){var iter__5480__auto__ = (function gl1tch$site$diagrams$flowchart$render_$_iter__41740(s__41741){
return (new cljs.core.LazySeq(null,(function (){
var s__41741__$1 = s__41741;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__41741__$1);
if(temp__5825__auto__){
var s__41741__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__41741__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__41741__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__41743 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__41742 = (0);
while(true){
if((i__41742 < size__5479__auto__)){
var edge = cljs.core._nth(c__5478__auto__,i__41742);
cljs.core.chunk_append(b__41743,cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrow,edge], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"from","from",1815293044).cljs$core$IFn$_invoke$arity$1(edge)),"-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"to","to",192099007).cljs$core$IFn$_invoke$arity$1(edge))].join('')], null)));

var G__41747 = (i__41742 + (1));
i__41742 = G__41747;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41743),gl1tch$site$diagrams$flowchart$render_$_iter__41740(cljs.core.chunk_rest(s__41741__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41743),null);
}
} else {
var edge = cljs.core.first(s__41741__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.svg.arrow,edge], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"from","from",1815293044).cljs$core$IFn$_invoke$arity$1(edge)),"-",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"to","to",192099007).cljs$core$IFn$_invoke$arity$1(edge))].join('')], null)),gl1tch$site$diagrams$flowchart$render_$_iter__41740(cljs.core.rest(s__41741__$2)));
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
