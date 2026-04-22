goog.provide('gl1tch.site.components.toc');
gl1tch.site.components.toc.heading__GT_id = (function gl1tch$site$components$toc$heading__GT_id(heading){
return clojure.string.replace(clojure.string.replace(clojure.string.lower_case(heading),/[^a-z0-9]+/,"-"),/^-|-$/,"");
});
gl1tch.site.components.toc.scroll_to = (function gl1tch$site$components$toc$scroll_to(id){
return (function (e){
e.preventDefault();

var temp__5825__auto__ = document.getElementById(id);
if(cljs.core.truth_(temp__5825__auto__)){
var el = temp__5825__auto__;
return el.scrollIntoView(({"behavior": "smooth", "block": "start"}));
} else {
return null;
}
});
});
gl1tch.site.components.toc.toc = (function gl1tch$site$components$toc$toc(sections){
if(cljs.core.seq(sections)){
return new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"nav","nav",719540477),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"position","position",-2011731912),new cljs.core.Keyword(null,"sticky","sticky",-2121213869),new cljs.core.Keyword(null,"top","top",-1856271961),"80px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xxs","xxs",-16722349)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xxs","xxs",-16722349))),new cljs.core.Keyword(null,"text-transform","text-transform",1685000676),new cljs.core.Keyword(null,"uppercase","uppercase",2080890922),new cljs.core.Keyword(null,"letter-spacing","letter-spacing",-948993767),"0.2em",new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"16px",new cljs.core.Keyword(null,"opacity","opacity",397153780),0.6], null)], null),"On this page"], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"flex-direction","flex-direction",364609438),new cljs.core.Keyword(null,"column","column",2078222095),new cljs.core.Keyword(null,"gap","gap",80255254),"8px"], null)], null),(function (){var iter__5480__auto__ = (function gl1tch$site$components$toc$toc_$_iter__42627(s__42628){
return (new cljs.core.LazySeq(null,(function (){
var s__42628__$1 = s__42628;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__42628__$1);
if(temp__5825__auto__){
var s__42628__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__42628__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__42628__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__42630 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__42629 = (0);
while(true){
if((i__42629 < size__5479__auto__)){
var map__42631 = cljs.core._nth(c__5478__auto__,i__42629);
var map__42631__$1 = cljs.core.__destructure_map(map__42631);
var heading = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42631__$1,new cljs.core.Keyword(null,"heading","heading",-1312171873));
var level = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42631__$1,new cljs.core.Keyword(null,"level","level",1290497552));
cljs.core.chunk_append(b__42630,cljs.core.with_meta(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"href","href",-793805698),"javascript:void(0)",new cljs.core.Keyword(null,"on-click","on-click",1632826543),gl1tch.site.components.toc.scroll_to(gl1tch.site.components.toc.heading__GT_id(heading)),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xs","xs",649443341)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xs","xs",649443341))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818))),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478),new cljs.core.Keyword(null,"padding-left","padding-left",-1180879053),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(((level - (2)) * (12))),"px"].join(''),new cljs.core.Keyword(null,"border-left","border-left",-1150760178),"1px solid transparent",new cljs.core.Keyword(null,"cursor","cursor",1011937484),new cljs.core.Keyword(null,"pointer","pointer",85071187)], null)], null),heading], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),heading], null)));

var G__42633 = (i__42629 + (1));
i__42629 = G__42633;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__42630),gl1tch$site$components$toc$toc_$_iter__42627(cljs.core.chunk_rest(s__42628__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__42630),null);
}
} else {
var map__42632 = cljs.core.first(s__42628__$2);
var map__42632__$1 = cljs.core.__destructure_map(map__42632);
var heading = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42632__$1,new cljs.core.Keyword(null,"heading","heading",-1312171873));
var level = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42632__$1,new cljs.core.Keyword(null,"level","level",1290497552));
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"href","href",-793805698),"javascript:void(0)",new cljs.core.Keyword(null,"on-click","on-click",1632826543),gl1tch.site.components.toc.scroll_to(gl1tch.site.components.toc.heading__GT_id(heading)),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xs","xs",649443341)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xs","xs",649443341))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818))),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478),new cljs.core.Keyword(null,"padding-left","padding-left",-1180879053),[cljs.core.str.cljs$core$IFn$_invoke$arity$1(((level - (2)) * (12))),"px"].join(''),new cljs.core.Keyword(null,"border-left","border-left",-1150760178),"1px solid transparent",new cljs.core.Keyword(null,"cursor","cursor",1011937484),new cljs.core.Keyword(null,"pointer","pointer",85071187)], null)], null),heading], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),heading], null)),gl1tch$site$components$toc$toc_$_iter__42627(cljs.core.rest(s__42628__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(sections);
})()], null)], null);
} else {
return null;
}
});

//# sourceMappingURL=gl1tch.site.components.toc.js.map
