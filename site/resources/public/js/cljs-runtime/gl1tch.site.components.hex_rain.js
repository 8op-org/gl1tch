goog.provide('gl1tch.site.components.hex_rain');
gl1tch.site.components.hex_rain.col_w = (28);
gl1tch.site.components.hex_rain.row_h = (18);
gl1tch.site.components.hex_rain.hex_byte = (function gl1tch$site$components$hex_rain$hex_byte(){
return Math.floor((Math.random() * (256))).toString((16)).padStart((2),"0").toUpperCase();
});
gl1tch.site.components.hex_rain.setup = (function gl1tch$site$components$hex_rain$setup(canvas){
var w = window.innerWidth;
var h = window.innerHeight;
var cols = (Math.ceil((w / gl1tch.site.components.hex_rain.col_w)) + (2));
var rows = (Math.ceil((h / gl1tch.site.components.hex_rain.row_h)) + (3));
var grid_h = (rows * (4));
(canvas.width = w);

(canvas.height = h);

return new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"cols","cols",-1914801295),cols,new cljs.core.Keyword(null,"rows","rows",850049680),rows,new cljs.core.Keyword(null,"offsets","offsets",727555611),cljs.core.vec(cljs.core.repeatedly.cljs$core$IFn$_invoke$arity$2(cols,(function (){
return ((Math.random() * rows) * gl1tch.site.components.hex_rain.row_h);
}))),new cljs.core.Keyword(null,"speeds","speeds",-1805136846),cljs.core.vec(cljs.core.repeatedly.cljs$core$IFn$_invoke$arity$2(cols,(function (){
return (0.25 + (Math.random() * 0.45));
}))),new cljs.core.Keyword(null,"grid","grid",402978600),cljs.core.vec(cljs.core.repeatedly.cljs$core$IFn$_invoke$arity$2(grid_h,(function (){
return cljs.core.vec(cljs.core.repeatedly.cljs$core$IFn$_invoke$arity$2(cols,gl1tch.site.components.hex_rain.hex_byte));
})))], null);
});
gl1tch.site.components.hex_rain.hex_rain = (function gl1tch$site$components$hex_rain$hex_rain(){
var state = reagent.core.atom.cljs$core$IFn$_invoke$arity$1(null);
var canvas_ref = reagent.core.atom.cljs$core$IFn$_invoke$arity$1(null);
var raf_id = reagent.core.atom.cljs$core$IFn$_invoke$arity$1(null);
return reagent.core.create_class.cljs$core$IFn$_invoke$arity$1(new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"component-did-mount","component-did-mount",-1126910518),(function (_){
var temp__5825__auto__ = cljs.core.deref(canvas_ref);
if(cljs.core.truth_(temp__5825__auto__)){
var canvas = temp__5825__auto__;
cljs.core.reset_BANG_(state,gl1tch.site.components.hex_rain.setup(canvas));

var ctx = canvas.getContext("2d");
var draw = (function gl1tch$site$components$hex_rain$hex_rain_$_draw(){
var map__41489 = cljs.core.deref(state);
var map__41489__$1 = cljs.core.__destructure_map(map__41489);
var cols = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41489__$1,new cljs.core.Keyword(null,"cols","cols",-1914801295));
var rows = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41489__$1,new cljs.core.Keyword(null,"rows","rows",850049680));
var offsets = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41489__$1,new cljs.core.Keyword(null,"offsets","offsets",727555611));
var speeds = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41489__$1,new cljs.core.Keyword(null,"speeds","speeds",-1805136846));
var grid = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41489__$1,new cljs.core.Keyword(null,"grid","grid",402978600));
ctx.clearRect((0),(0),canvas.width,canvas.height);

(ctx.font = "12px 'JetBrains Mono', monospace");

(ctx.fillStyle = "rgba(135, 135, 175, 0.07)");

var new_offsets_41517 = cljs.core.mapv.cljs$core$IFn$_invoke$arity$2((function (col){
var off = cljs.core.mod((cljs.core.nth.cljs$core$IFn$_invoke$arity$2(offsets,col) + cljs.core.nth.cljs$core$IFn$_invoke$arity$2(speeds,col)),(cljs.core.count(grid) * gl1tch.site.components.hex_rain.row_h));
var start_row = Math.floor((off / gl1tch.site.components.hex_rain.row_h));
var sub_px = cljs.core.mod(off,gl1tch.site.components.hex_rain.row_h);
var seq__41493_41520 = cljs.core.seq(cljs.core.range.cljs$core$IFn$_invoke$arity$1((rows + (2))));
var chunk__41494_41521 = null;
var count__41495_41522 = (0);
var i__41496_41523 = (0);
while(true){
if((i__41496_41523 < count__41495_41522)){
var row_41524 = chunk__41494_41521.cljs$core$IIndexed$_nth$arity$2(null, i__41496_41523);
var y_41525 = ((row_41524 * gl1tch.site.components.hex_rain.row_h) - sub_px);
var grid_row_41526 = cljs.core.mod((start_row + row_41524),cljs.core.count(grid));
ctx.fillText(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(grid,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [grid_row_41526,col], null)),(col * gl1tch.site.components.hex_rain.col_w),y_41525);


var G__41528 = seq__41493_41520;
var G__41529 = chunk__41494_41521;
var G__41530 = count__41495_41522;
var G__41531 = (i__41496_41523 + (1));
seq__41493_41520 = G__41528;
chunk__41494_41521 = G__41529;
count__41495_41522 = G__41530;
i__41496_41523 = G__41531;
continue;
} else {
var temp__5825__auto___41532__$1 = cljs.core.seq(seq__41493_41520);
if(temp__5825__auto___41532__$1){
var seq__41493_41533__$1 = temp__5825__auto___41532__$1;
if(cljs.core.chunked_seq_QMARK_(seq__41493_41533__$1)){
var c__5525__auto___41534 = cljs.core.chunk_first(seq__41493_41533__$1);
var G__41535 = cljs.core.chunk_rest(seq__41493_41533__$1);
var G__41536 = c__5525__auto___41534;
var G__41537 = cljs.core.count(c__5525__auto___41534);
var G__41538 = (0);
seq__41493_41520 = G__41535;
chunk__41494_41521 = G__41536;
count__41495_41522 = G__41537;
i__41496_41523 = G__41538;
continue;
} else {
var row_41539 = cljs.core.first(seq__41493_41533__$1);
var y_41540 = ((row_41539 * gl1tch.site.components.hex_rain.row_h) - sub_px);
var grid_row_41541 = cljs.core.mod((start_row + row_41539),cljs.core.count(grid));
ctx.fillText(cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(grid,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [grid_row_41541,col], null)),(col * gl1tch.site.components.hex_rain.col_w),y_41540);


var G__41543 = cljs.core.next(seq__41493_41533__$1);
var G__41544 = null;
var G__41545 = (0);
var G__41546 = (0);
seq__41493_41520 = G__41543;
chunk__41494_41521 = G__41544;
count__41495_41522 = G__41545;
i__41496_41523 = G__41546;
continue;
}
} else {
}
}
break;
}

return off;
}),cljs.core.range.cljs$core$IFn$_invoke$arity$1(cols));
cljs.core.swap_BANG_.cljs$core$IFn$_invoke$arity$4(state,cljs.core.assoc,new cljs.core.Keyword(null,"offsets","offsets",727555611),new_offsets_41517);

return cljs.core.reset_BANG_(raf_id,requestAnimationFrame(gl1tch$site$components$hex_rain$hex_rain_$_draw));
});
return draw();
} else {
return null;
}
}),new cljs.core.Keyword(null,"component-will-unmount","component-will-unmount",-2058314698),(function (_){
if(cljs.core.truth_(cljs.core.deref(raf_id))){
return cancelAnimationFrame(cljs.core.deref(raf_id));
} else {
return null;
}
}),new cljs.core.Keyword(null,"reagent-render","reagent-render",-985383853),(function (){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"canvas","canvas",-1798817489),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"ref","ref",1289896967),(function (p1__41477_SHARP_){
return cljs.core.reset_BANG_(canvas_ref,p1__41477_SHARP_);
}),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 8, [new cljs.core.Keyword(null,"position","position",-2011731912),new cljs.core.Keyword(null,"fixed","fixed",-562004358),new cljs.core.Keyword(null,"top","top",-1856271961),(0),new cljs.core.Keyword(null,"left","left",-399115937),(0),new cljs.core.Keyword(null,"width","width",-384071477),"100%",new cljs.core.Keyword(null,"height","height",1025178622),"100%",new cljs.core.Keyword(null,"z-index","z-index",1892827090),(0),new cljs.core.Keyword(null,"pointer-events","pointer-events",-1053858853),new cljs.core.Keyword(null,"none","none",1333468478),new cljs.core.Keyword(null,"opacity","opacity",397153780),0.04], null)], null)], null);
})], null));
});

//# sourceMappingURL=gl1tch.site.components.hex_rain.js.map
