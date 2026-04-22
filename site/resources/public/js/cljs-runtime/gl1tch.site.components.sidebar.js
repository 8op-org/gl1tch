goog.provide('gl1tch.site.components.sidebar');
gl1tch.site.components.sidebar.sidebar = (function gl1tch$site$components$sidebar$sidebar(){
var match = cljs.core.deref(gl1tch.site.state.current_match);
var current_slug = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(match,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"path-params","path-params",-48130597),new cljs.core.Keyword(null,"slug","slug",2029314850)], null));
var docs = gl1tch.site.content.docs_index();
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"nav","nav",719540477),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"position","position",-2011731912),new cljs.core.Keyword(null,"sticky","sticky",-2121213869),new cljs.core.Keyword(null,"top","top",-1856271961),"80px",new cljs.core.Keyword(null,"max-height","max-height",-612563804),"calc(100vh - 100px)",new cljs.core.Keyword(null,"overflow-y","overflow-y",-1436589285),new cljs.core.Keyword(null,"auto","auto",-566279492)], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"flex-direction","flex-direction",364609438),new cljs.core.Keyword(null,"column","column",2078222095),new cljs.core.Keyword(null,"gap","gap",80255254),"2px"], null)], null),(function (){var iter__5480__auto__ = (function gl1tch$site$components$sidebar$sidebar_$_iter__41620(s__41621){
return (new cljs.core.LazySeq(null,(function (){
var s__41621__$1 = s__41621;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__41621__$1);
if(temp__5825__auto__){
var s__41621__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__41621__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__41621__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__41623 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__41622 = (0);
while(true){
if((i__41622 < size__5479__auto__)){
var doc = cljs.core._nth(c__5478__auto__,i__41622);
cljs.core.chunk_append(b__41623,cljs.core.with_meta(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"href","href",-793805698),reitit.frontend.easy.href.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword("docs","page","docs/page",848096842),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"slug","slug",2029314850),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xs","xs",649443341)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xs","xs",649443341))),new cljs.core.Keyword(null,"color","color",1011675173),((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(current_slug,new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)))?(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))):(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478),new cljs.core.Keyword(null,"padding","padding",1660304693),"5px 12px",new cljs.core.Keyword(null,"border-radius","border-radius",419594011),gl1tch.site.styles.theme.radius,new cljs.core.Keyword(null,"border-left","border-left",-1150760178),["2px solid ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(current_slug,new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)))?(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))):"transparent"))].join('')], null)], null),new cljs.core.Keyword(null,"title","title",636505583).cljs$core$IFn$_invoke$arity$1(doc)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)));

var G__41642 = (i__41622 + (1));
i__41622 = G__41642;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41623),gl1tch$site$components$sidebar$sidebar_$_iter__41620(cljs.core.chunk_rest(s__41621__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41623),null);
}
} else {
var doc = cljs.core.first(s__41621__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"href","href",-793805698),reitit.frontend.easy.href.cljs$core$IFn$_invoke$arity$2(new cljs.core.Keyword("docs","page","docs/page",848096842),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"slug","slug",2029314850),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xs","xs",649443341)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xs","xs",649443341))),new cljs.core.Keyword(null,"color","color",1011675173),((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(current_slug,new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)))?(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))):(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478),new cljs.core.Keyword(null,"padding","padding",1660304693),"5px 12px",new cljs.core.Keyword(null,"border-radius","border-radius",419594011),gl1tch.site.styles.theme.radius,new cljs.core.Keyword(null,"border-left","border-left",-1150760178),["2px solid ",cljs.core.str.cljs$core$IFn$_invoke$arity$1(((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(current_slug,new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)))?(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))):"transparent"))].join('')], null)], null),new cljs.core.Keyword(null,"title","title",636505583).cljs$core$IFn$_invoke$arity$1(doc)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)),gl1tch$site$components$sidebar$sidebar_$_iter__41620(cljs.core.rest(s__41621__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(docs);
})()], null)], null);
});

//# sourceMappingURL=gl1tch.site.components.sidebar.js.map
