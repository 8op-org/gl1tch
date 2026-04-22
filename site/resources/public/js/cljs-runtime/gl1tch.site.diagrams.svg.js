goog.provide('gl1tch.site.diagrams.svg');
gl1tch.site.diagrams.svg.box = (function gl1tch$site$diagrams$svg$box(p__41681){
var map__41682 = p__41681;
var map__41682__$1 = cljs.core.__destructure_map(map__41682);
var x = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"x","x",2099068185));
var y = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"y","y",-1757859776));
var w = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"w","w",354169001));
var h = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"h","h",1109658740));
var label = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"label","label",1718410804));
var style = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41682__$1,new cljs.core.Keyword(null,"style","style",-496642736));
var fill = (function (){var or__5002__auto__ = new cljs.core.Keyword(null,"fill","fill",883462889).cljs$core$IFn$_invoke$arity$1(style);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"bg-alt","bg-alt",1108747320)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"bg-alt","bg-alt",1108747320)));
}
})();
var stroke = (function (){var or__5002__auto__ = new cljs.core.Keyword(null,"stroke","stroke",1741823555).cljs$core$IFn$_invoke$arity$1(style);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468)));
}
})();
var text_fill = (function (){var or__5002__auto__ = new cljs.core.Keyword(null,"color","color",1011675173).cljs$core$IFn$_invoke$arity$1(style);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg","fg",-101797208)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg","fg",-101797208)));
}
})();
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"g","g",1738089905),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"rect","rect",-108902628),cljs.core.PersistentHashMap.fromArrays([new cljs.core.Keyword(null,"y","y",-1757859776),new cljs.core.Keyword(null,"rx","rx",1627208482),new cljs.core.Keyword(null,"stroke","stroke",1741823555),new cljs.core.Keyword(null,"fill","fill",883462889),new cljs.core.Keyword(null,"width","width",-384071477),new cljs.core.Keyword(null,"stroke-width","stroke-width",716836435),new cljs.core.Keyword(null,"x","x",2099068185),new cljs.core.Keyword(null,"ry","ry",-334598563),new cljs.core.Keyword(null,"height","height",1025178622)],[y,(4),stroke,fill,w,1.5,x,(4),h])], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"text","text",-1790561697),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"x","x",2099068185),(x + (w / (2))),new cljs.core.Keyword(null,"y","y",-1757859776),((y + (h / (2))) + (5)),new cljs.core.Keyword(null,"text-anchor","text-anchor",585613696),"middle",new cljs.core.Keyword(null,"fill","fill",883462889),text_fill,new cljs.core.Keyword(null,"font-family","font-family",-667419874),(gl1tch.site.styles.theme.fonts.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.fonts.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"mono","mono",-1777958350)) : gl1tch.site.styles.theme.fonts.call(null, new cljs.core.Keyword(null,"mono","mono",-1777958350))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(12)], null),label], null)], null);
});
gl1tch.site.diagrams.svg.arrow = (function gl1tch$site$diagrams$svg$arrow(p__41686){
var map__41687 = p__41686;
var map__41687__$1 = cljs.core.__destructure_map(map__41687);
var path = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41687__$1,new cljs.core.Keyword(null,"path","path",-188191168));
var style = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41687__$1,new cljs.core.Keyword(null,"style","style",-496642736));
var stroke = (function (){var or__5002__auto__ = new cljs.core.Keyword(null,"stroke","stroke",1741823555).cljs$core$IFn$_invoke$arity$1(style);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)));
}
})();
var points = cljs.core.apply.cljs$core$IFn$_invoke$arity$2(cljs.core.str,cljs.core.interpose.cljs$core$IFn$_invoke$arity$2(" ",cljs.core.map.cljs$core$IFn$_invoke$arity$2((function (p__41689){
var vec__41690 = p__41689;
var px = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41690,(0),null);
var py = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41690,(1),null);
return [cljs.core.str.cljs$core$IFn$_invoke$arity$1(px),",",cljs.core.str.cljs$core$IFn$_invoke$arity$1(py)].join('');
}),path)));
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"polyline","polyline",-1731551044),new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"points","points",-1486596883),points,new cljs.core.Keyword(null,"fill","fill",883462889),"none",new cljs.core.Keyword(null,"stroke","stroke",1741823555),stroke,new cljs.core.Keyword(null,"stroke-width","stroke-width",716836435),1.5,new cljs.core.Keyword(null,"marker-end","marker-end",341488703),"url(#arrowhead)"], null)], null);
});
gl1tch.site.diagrams.svg.arrowhead_marker = (function gl1tch$site$diagrams$svg$arrowhead_marker(){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"defs","defs",1398449717),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"marker","marker",865118313),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"id","id",-1388402092),"arrowhead",new cljs.core.Keyword(null,"markerWidth","markerWidth",-568766230),(10),new cljs.core.Keyword(null,"markerHeight","markerHeight",-1744163958),(7),new cljs.core.Keyword(null,"refX","refX",1265839261),(10),new cljs.core.Keyword(null,"refY","refY",113675749),3.5,new cljs.core.Keyword(null,"orient","orient",1933743565),"auto"], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"polygon","polygon",837053759),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"points","points",-1486596883),"0 0, 10 3.5, 0 7",new cljs.core.Keyword(null,"fill","fill",883462889),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))], null)], null)], null)], null);
});
gl1tch.site.diagrams.svg.label = (function gl1tch$site$diagrams$svg$label(p__41701){
var map__41702 = p__41701;
var map__41702__$1 = cljs.core.__destructure_map(map__41702);
var x = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41702__$1,new cljs.core.Keyword(null,"x","x",2099068185));
var y = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41702__$1,new cljs.core.Keyword(null,"y","y",-1757859776));
var text = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41702__$1,new cljs.core.Keyword(null,"text","text",-1790561697));
var style = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41702__$1,new cljs.core.Keyword(null,"style","style",-496642736));
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"text","text",-1790561697),new cljs.core.PersistentArrayMap(null, 5, [new cljs.core.Keyword(null,"x","x",2099068185),x,new cljs.core.Keyword(null,"y","y",-1757859776),y,new cljs.core.Keyword(null,"fill","fill",883462889),(function (){var or__5002__auto__ = new cljs.core.Keyword(null,"color","color",1011675173).cljs$core$IFn$_invoke$arity$1(style);
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return (gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)));
}
})(),new cljs.core.Keyword(null,"font-family","font-family",-667419874),(gl1tch.site.styles.theme.fonts.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.fonts.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"mono","mono",-1777958350)) : gl1tch.site.styles.theme.fonts.call(null, new cljs.core.Keyword(null,"mono","mono",-1777958350))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(11)], null),text], null);
});

//# sourceMappingURL=gl1tch.site.diagrams.svg.js.map
