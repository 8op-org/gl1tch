goog.provide('gl1tch.site.diagrams.core');
gl1tch.site.diagrams.core.node_w = (120);
gl1tch.site.diagrams.core.node_h = (40);
gl1tch.site.diagrams.core.gap_x = (60);
gl1tch.site.diagrams.core.gap_y = (30);
gl1tch.site.diagrams.core.padding = (20);
/**
 * Topological sort into layers. Nodes with no incoming edges go first.
 */
gl1tch.site.diagrams.core.assign_layers = (function gl1tch$site$diagrams$core$assign_layers(nodes,edges){
var ids = cljs.core.set(cljs.core.map.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"id","id",-1388402092),nodes));
var in_map = cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (m,p__21233){
var vec__21234 = p__21233;
var from = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21234,(0),null);
var to = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21234,(1),null);
return cljs.core.update.cljs$core$IFn$_invoke$arity$4(m,to,cljs.core.fnil.cljs$core$IFn$_invoke$arity$2(cljs.core.conj,cljs.core.PersistentHashSet.EMPTY),from);
}),cljs.core.PersistentArrayMap.EMPTY,edges);
var remaining = ids;
var layers = cljs.core.PersistentVector.EMPTY;
while(true){
if(cljs.core.empty_QMARK_(remaining)){
return layers;
} else {
var ready = cljs.core.filterv(((function (remaining,layers,ids,in_map){
return (function (p1__21232_SHARP_){
return cljs.core.every_QMARK_(((function (remaining,layers,ids,in_map){
return (function (dep){
return cljs.core.not((remaining.cljs$core$IFn$_invoke$arity$1 ? remaining.cljs$core$IFn$_invoke$arity$1(dep) : remaining.call(null, dep)));
});})(remaining,layers,ids,in_map))
,cljs.core.get.cljs$core$IFn$_invoke$arity$3(in_map,p1__21232_SHARP_,cljs.core.PersistentHashSet.EMPTY));
});})(remaining,layers,ids,in_map))
,remaining);
if(cljs.core.empty_QMARK_(ready)){
return cljs.core.conj.cljs$core$IFn$_invoke$arity$2(layers,cljs.core.vec(remaining));
} else {
var G__21299 = cljs.core.apply.cljs$core$IFn$_invoke$arity$3(cljs.core.disj,remaining,ready);
var G__21300 = cljs.core.conj.cljs$core$IFn$_invoke$arity$2(layers,ready);
remaining = G__21299;
layers = G__21300;
continue;
}
}
break;
}
});
gl1tch.site.diagrams.core.node_center = (function gl1tch$site$diagrams$core$node_center(node){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.Keyword(null,"x","x",2099068185).cljs$core$IFn$_invoke$arity$1(node) + (new cljs.core.Keyword(null,"w","w",354169001).cljs$core$IFn$_invoke$arity$1(node) / (2))),(new cljs.core.Keyword(null,"y","y",-1757859776).cljs$core$IFn$_invoke$arity$1(node) + (new cljs.core.Keyword(null,"h","h",1109658740).cljs$core$IFn$_invoke$arity$1(node) / (2)))], null);
});
gl1tch.site.diagrams.core.compute_edge_path = (function gl1tch$site$diagrams$core$compute_edge_path(from_node,to_node,direction){
var vec__21252 = gl1tch.site.diagrams.core.node_center(from_node);
var fx = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21252,(0),null);
var fy = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21252,(1),null);
var vec__21255 = gl1tch.site.diagrams.core.node_center(to_node);
var tx = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21255,(0),null);
var ty = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21255,(1),null);
if(cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(direction,new cljs.core.Keyword(null,"lr","lr",445647393))){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(new cljs.core.Keyword(null,"x","x",2099068185).cljs$core$IFn$_invoke$arity$1(from_node) + new cljs.core.Keyword(null,"w","w",354169001).cljs$core$IFn$_invoke$arity$1(from_node)),fy], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"x","x",2099068185).cljs$core$IFn$_invoke$arity$1(to_node),ty], null)], null);
} else {
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [fx,(new cljs.core.Keyword(null,"y","y",-1757859776).cljs$core$IFn$_invoke$arity$1(from_node) + new cljs.core.Keyword(null,"h","h",1109658740).cljs$core$IFn$_invoke$arity$1(from_node))], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [tx,new cljs.core.Keyword(null,"y","y",-1757859776).cljs$core$IFn$_invoke$arity$1(to_node)], null)], null);
}
});
/**
 * Takes a flowchart spec {:direction :lr/:tb :nodes [...] :edges [...]}
 * Returns {:nodes [...with :x :y :w :h...] :edges [...with :path...]
 *          :width :height}
 */
gl1tch.site.diagrams.core.compute = (function gl1tch$site$diagrams$core$compute(p__21267){
var map__21268 = p__21267;
var map__21268__$1 = cljs.core.__destructure_map(map__21268);
var direction = cljs.core.get.cljs$core$IFn$_invoke$arity$3(map__21268__$1,new cljs.core.Keyword(null,"direction","direction",-633359395),new cljs.core.Keyword(null,"lr","lr",445647393));
var nodes = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21268__$1,new cljs.core.Keyword(null,"nodes","nodes",-2099585805));
var edges = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__21268__$1,new cljs.core.Keyword(null,"edges","edges",-694791395));
var layers = gl1tch.site.diagrams.core.assign_layers(nodes,edges);
var node_map = cljs.core.into.cljs$core$IFn$_invoke$arity$2(cljs.core.PersistentArrayMap.EMPTY,cljs.core.map.cljs$core$IFn$_invoke$arity$2(cljs.core.juxt.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword(null,"id","id",-1388402092),cljs.core.identity),nodes));
var positioned = cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (acc,p__21271){
var vec__21272 = p__21271;
var layer_idx = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21272,(0),null);
var layer = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21272,(1),null);
return cljs.core.reduce.cljs$core$IFn$_invoke$arity$3((function (acc2,p__21275){
var vec__21276 = p__21275;
var pos_in_layer = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21276,(0),null);
var node_id = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21276,(1),null);
var node = cljs.core.get.cljs$core$IFn$_invoke$arity$2(node_map,node_id);
var vec__21279 = ((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(direction,new cljs.core.Keyword(null,"lr","lr",445647393)))?new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(gl1tch.site.diagrams.core.padding + (layer_idx * (gl1tch.site.diagrams.core.node_w + gl1tch.site.diagrams.core.gap_x))),(gl1tch.site.diagrams.core.padding + (pos_in_layer * (gl1tch.site.diagrams.core.node_h + gl1tch.site.diagrams.core.gap_y)))], null):new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [(gl1tch.site.diagrams.core.padding + (pos_in_layer * (gl1tch.site.diagrams.core.node_w + gl1tch.site.diagrams.core.gap_x))),(gl1tch.site.diagrams.core.padding + (layer_idx * (gl1tch.site.diagrams.core.node_h + gl1tch.site.diagrams.core.gap_y)))], null));
var x = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21279,(0),null);
var y = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21279,(1),null);
return cljs.core.assoc.cljs$core$IFn$_invoke$arity$3(acc2,node_id,cljs.core.merge.cljs$core$IFn$_invoke$arity$variadic(cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([node,new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"x","x",2099068185),x,new cljs.core.Keyword(null,"y","y",-1757859776),y,new cljs.core.Keyword(null,"w","w",354169001),gl1tch.site.diagrams.core.node_w,new cljs.core.Keyword(null,"h","h",1109658740),gl1tch.site.diagrams.core.node_h], null)], 0)));
}),acc,cljs.core.map_indexed.cljs$core$IFn$_invoke$arity$2(cljs.core.vector,layer));
}),cljs.core.PersistentArrayMap.EMPTY,cljs.core.map_indexed.cljs$core$IFn$_invoke$arity$2(cljs.core.vector,layers));
var edge_results = cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (p__21292){
var vec__21293 = p__21292;
var from = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21293,(0),null);
var to = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__21293,(1),null);
return new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"from","from",1815293044),from,new cljs.core.Keyword(null,"to","to",192099007),to,new cljs.core.Keyword(null,"path","path",-188191168),gl1tch.site.diagrams.core.compute_edge_path((positioned.cljs$core$IFn$_invoke$arity$1 ? positioned.cljs$core$IFn$_invoke$arity$1(from) : positioned.call(null, from)),(positioned.cljs$core$IFn$_invoke$arity$1 ? positioned.cljs$core$IFn$_invoke$arity$1(to) : positioned.call(null, to)),direction)], null);
}),edges);
var all_nodes = cljs.core.mapv.cljs$core$IFn$_invoke$arity$2(positioned,cljs.core.mapcat.cljs$core$IFn$_invoke$arity$variadic(cljs.core.identity,cljs.core.prim_seq.cljs$core$IFn$_invoke$arity$2([layers], 0)));
var max_x = cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.max,cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p1__21261_SHARP_){
return (new cljs.core.Keyword(null,"x","x",2099068185).cljs$core$IFn$_invoke$arity$1(p1__21261_SHARP_) + new cljs.core.Keyword(null,"w","w",354169001).cljs$core$IFn$_invoke$arity$1(p1__21261_SHARP_));
}),all_nodes));
var max_y = cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.max,cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p1__21262_SHARP_){
return (new cljs.core.Keyword(null,"y","y",-1757859776).cljs$core$IFn$_invoke$arity$1(p1__21262_SHARP_) + new cljs.core.Keyword(null,"h","h",1109658740).cljs$core$IFn$_invoke$arity$1(p1__21262_SHARP_));
}),all_nodes));
return new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"nodes","nodes",-2099585805),all_nodes,new cljs.core.Keyword(null,"edges","edges",-694791395),edge_results,new cljs.core.Keyword(null,"width","width",-384071477),(max_x + gl1tch.site.diagrams.core.padding),new cljs.core.Keyword(null,"height","height",1025178622),(max_y + gl1tch.site.diagrams.core.padding)], null);
});

//# sourceMappingURL=gl1tch.site.diagrams.core.js.map
