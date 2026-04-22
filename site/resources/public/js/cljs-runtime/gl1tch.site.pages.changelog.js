goog.provide('gl1tch.site.pages.changelog');
gl1tch.site.pages.changelog.type_colors = new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"feat","feat",1051430276),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"fix","fix",-1031773329),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"danger","danger",-624338030)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"danger","danger",-624338030))),new cljs.core.Keyword(null,"docs","docs",-1974280502),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent-2","accent-2",-1373850285)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent-2","accent-2",-1373850285))),new cljs.core.Keyword(null,"refactor","refactor",-1110737559),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))], null);
gl1tch.site.pages.changelog.page = (function gl1tch$site$pages$changelog$page(){
return new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div.wrap","div.wrap",1832950772),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"padding-top","padding-top",1929675955),"80px"], null)], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h1","h1",-1896887462),"Changelog"], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-top","margin-top",392161226),"40px"], null)], null),(function (){var iter__5480__auto__ = (function gl1tch$site$pages$changelog$page_$_iter__41752(s__41753){
return (new cljs.core.LazySeq(null,(function (){
var s__41753__$1 = s__41753;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__41753__$1);
if(temp__5825__auto__){
var s__41753__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__41753__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__41753__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__41755 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__41754 = (0);
while(true){
if((i__41754 < size__5479__auto__)){
var entry = cljs.core._nth(c__5478__auto__,i__41754);
cljs.core.chunk_append(b__41755,cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"48px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h2","h2",-372662728),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),"1.1rem",new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"16px"], null)], null),new cljs.core.Keyword(null,"date","date",-1463434462).cljs$core$IFn$_invoke$arity$1(entry)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"flex-direction","flex-direction",364609438),new cljs.core.Keyword(null,"column","column",2078222095),new cljs.core.Keyword(null,"gap","gap",80255254),"8px"], null)], null),(function (){var iter__5480__auto__ = ((function (i__41754,entry,c__5478__auto__,size__5479__auto__,b__41755,s__41753__$2,temp__5825__auto__){
return (function gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41764(s__41765){
return (new cljs.core.LazySeq(null,((function (i__41754,entry,c__5478__auto__,size__5479__auto__,b__41755,s__41753__$2,temp__5825__auto__){
return (function (){
var s__41765__$1 = s__41765;
while(true){
var temp__5825__auto____$1 = cljs.core.seq(s__41765__$1);
if(temp__5825__auto____$1){
var s__41765__$2 = temp__5825__auto____$1;
if(cljs.core.chunked_seq_QMARK_(s__41765__$2)){
var c__5478__auto____$1 = cljs.core.chunk_first(s__41765__$2);
var size__5479__auto____$1 = cljs.core.count(c__5478__auto____$1);
var b__41767 = cljs.core.chunk_buffer(size__5479__auto____$1);
if((function (){var i__41766 = (0);
while(true){
if((i__41766 < size__5479__auto____$1)){
var vec__41768 = cljs.core._nth(c__5478__auto____$1,i__41766);
var i = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41768,(0),null);
var item = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41768,(1),null);
cljs.core.chunk_append(b__41767,cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"gap","gap",80255254),"12px",new cljs.core.Keyword(null,"align-items","align-items",-267946462),new cljs.core.Keyword(null,"baseline","baseline",1151033280)], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xxs","xxs",-16722349)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xxs","xxs",-16722349))),new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-transform","text-transform",1685000676),new cljs.core.Keyword(null,"uppercase","uppercase",2080890922),new cljs.core.Keyword(null,"letter-spacing","letter-spacing",-948993767),"0.1em",new cljs.core.Keyword(null,"color","color",1011675173),cljs.core.get.cljs$core$IFn$_invoke$arity$3(gl1tch.site.pages.changelog.type_colors,new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"min-width","min-width",1926193728),"60px"], null)], null),cljs.core.name(new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item))], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg","fg",-101797208)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg","fg",-101797208)))], null)], null),new cljs.core.Keyword(null,"summary","summary",380847952).cljs$core$IFn$_invoke$arity$1(item)], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),i], null)));

var G__41810 = (i__41766 + (1));
i__41766 = G__41810;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41767),gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41764(cljs.core.chunk_rest(s__41765__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41767),null);
}
} else {
var vec__41771 = cljs.core.first(s__41765__$2);
var i = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41771,(0),null);
var item = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41771,(1),null);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"gap","gap",80255254),"12px",new cljs.core.Keyword(null,"align-items","align-items",-267946462),new cljs.core.Keyword(null,"baseline","baseline",1151033280)], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xxs","xxs",-16722349)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xxs","xxs",-16722349))),new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-transform","text-transform",1685000676),new cljs.core.Keyword(null,"uppercase","uppercase",2080890922),new cljs.core.Keyword(null,"letter-spacing","letter-spacing",-948993767),"0.1em",new cljs.core.Keyword(null,"color","color",1011675173),cljs.core.get.cljs$core$IFn$_invoke$arity$3(gl1tch.site.pages.changelog.type_colors,new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"min-width","min-width",1926193728),"60px"], null)], null),cljs.core.name(new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item))], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg","fg",-101797208)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg","fg",-101797208)))], null)], null),new cljs.core.Keyword(null,"summary","summary",380847952).cljs$core$IFn$_invoke$arity$1(item)], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),i], null)),gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41764(cljs.core.rest(s__41765__$2)));
}
} else {
return null;
}
break;
}
});})(i__41754,entry,c__5478__auto__,size__5479__auto__,b__41755,s__41753__$2,temp__5825__auto__))
,null,null));
});})(i__41754,entry,c__5478__auto__,size__5479__auto__,b__41755,s__41753__$2,temp__5825__auto__))
;
return iter__5480__auto__(cljs.core.map_indexed.cljs$core$IFn$_invoke$arity$2(cljs.core.vector,new cljs.core.Keyword(null,"entries","entries",-86943161).cljs$core$IFn$_invoke$arity$1(entry)));
})()], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"date","date",-1463434462).cljs$core$IFn$_invoke$arity$1(entry)], null)));

var G__41815 = (i__41754 + (1));
i__41754 = G__41815;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41755),gl1tch$site$pages$changelog$page_$_iter__41752(cljs.core.chunk_rest(s__41753__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41755),null);
}
} else {
var entry = cljs.core.first(s__41753__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"48px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h2","h2",-372662728),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),"1.1rem",new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"16px"], null)], null),new cljs.core.Keyword(null,"date","date",-1463434462).cljs$core$IFn$_invoke$arity$1(entry)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"flex-direction","flex-direction",364609438),new cljs.core.Keyword(null,"column","column",2078222095),new cljs.core.Keyword(null,"gap","gap",80255254),"8px"], null)], null),(function (){var iter__5480__auto__ = ((function (entry,s__41753__$2,temp__5825__auto__){
return (function gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41776(s__41777){
return (new cljs.core.LazySeq(null,(function (){
var s__41777__$1 = s__41777;
while(true){
var temp__5825__auto____$1 = cljs.core.seq(s__41777__$1);
if(temp__5825__auto____$1){
var s__41777__$2 = temp__5825__auto____$1;
if(cljs.core.chunked_seq_QMARK_(s__41777__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__41777__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__41779 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__41778 = (0);
while(true){
if((i__41778 < size__5479__auto__)){
var vec__41780 = cljs.core._nth(c__5478__auto__,i__41778);
var i = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41780,(0),null);
var item = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41780,(1),null);
cljs.core.chunk_append(b__41779,cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"gap","gap",80255254),"12px",new cljs.core.Keyword(null,"align-items","align-items",-267946462),new cljs.core.Keyword(null,"baseline","baseline",1151033280)], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xxs","xxs",-16722349)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xxs","xxs",-16722349))),new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-transform","text-transform",1685000676),new cljs.core.Keyword(null,"uppercase","uppercase",2080890922),new cljs.core.Keyword(null,"letter-spacing","letter-spacing",-948993767),"0.1em",new cljs.core.Keyword(null,"color","color",1011675173),cljs.core.get.cljs$core$IFn$_invoke$arity$3(gl1tch.site.pages.changelog.type_colors,new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"min-width","min-width",1926193728),"60px"], null)], null),cljs.core.name(new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item))], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg","fg",-101797208)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg","fg",-101797208)))], null)], null),new cljs.core.Keyword(null,"summary","summary",380847952).cljs$core$IFn$_invoke$arity$1(item)], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),i], null)));

var G__41820 = (i__41778 + (1));
i__41778 = G__41820;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__41779),gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41776(cljs.core.chunk_rest(s__41777__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__41779),null);
}
} else {
var vec__41796 = cljs.core.first(s__41777__$2);
var i = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41796,(0),null);
var item = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__41796,(1),null);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 3, [new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"flex","flex",-1425124628),new cljs.core.Keyword(null,"gap","gap",80255254),"12px",new cljs.core.Keyword(null,"align-items","align-items",-267946462),new cljs.core.Keyword(null,"baseline","baseline",1151033280)], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"xxs","xxs",-16722349)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"xxs","xxs",-16722349))),new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-transform","text-transform",1685000676),new cljs.core.Keyword(null,"uppercase","uppercase",2080890922),new cljs.core.Keyword(null,"letter-spacing","letter-spacing",-948993767),"0.1em",new cljs.core.Keyword(null,"color","color",1011675173),cljs.core.get.cljs$core$IFn$_invoke$arity$3(gl1tch.site.pages.changelog.type_colors,new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)))),new cljs.core.Keyword(null,"min-width","min-width",1926193728),"60px"], null)], null),cljs.core.name(new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(item))], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"span","span",1394872991),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065))),new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg","fg",-101797208)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg","fg",-101797208)))], null)], null),new cljs.core.Keyword(null,"summary","summary",380847952).cljs$core$IFn$_invoke$arity$1(item)], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),i], null)),gl1tch$site$pages$changelog$page_$_iter__41752_$_iter__41776(cljs.core.rest(s__41777__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});})(entry,s__41753__$2,temp__5825__auto__))
;
return iter__5480__auto__(cljs.core.map_indexed.cljs$core$IFn$_invoke$arity$2(cljs.core.vector,new cljs.core.Keyword(null,"entries","entries",-86943161).cljs$core$IFn$_invoke$arity$1(entry)));
})()], null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"date","date",-1463434462).cljs$core$IFn$_invoke$arity$1(entry)], null)),gl1tch$site$pages$changelog$page_$_iter__41752(cljs.core.rest(s__41753__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(gl1tch.site.content.changelog_entries());
})()], null)], null);
});

//# sourceMappingURL=gl1tch.site.pages.changelog.js.map
