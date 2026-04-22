goog.provide('gl1tch.site.pages.docs');
gl1tch.site.pages.docs.heading__GT_id = (function gl1tch$site$pages$docs$heading__GT_id(heading){
return clojure.string.replace(clojure.string.replace(clojure.string.lower_case(heading),/[^a-z0-9]+/,"-"),/^-|-$/,"");
});
/**
 * Renders a hiccup body vector, resolving :code and :diagram tags.
 */
gl1tch.site.pages.docs.render_hiccup = (function gl1tch$site$pages$docs$render_hiccup(element){
if(typeof element === 'string'){
return element;
} else {
if((element == null)){
return null;
} else {
if(cljs.core.vector_QMARK_(element)){
var vec__42670 = element;
var seq__42671 = cljs.core.seq(vec__42670);
var first__42672 = cljs.core.first(seq__42671);
var seq__42671__$1 = cljs.core.next(seq__42671);
var tag = first__42672;
var rest = seq__42671__$1;
var vec__42673 = ((cljs.core.map_QMARK_(cljs.core.first(rest)))?new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [cljs.core.first(rest),cljs.core.vec(cljs.core.next(rest))], null):new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [null,cljs.core.vec(rest)], null));
var attrs = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__42673,(0),null);
var children = cljs.core.nth.cljs$core$IFn$_invoke$arity$3(vec__42673,(1),null);
var G__42676 = tag;
var G__42676__$1 = (((G__42676 instanceof cljs.core.Keyword))?G__42676.fqn:null);
switch (G__42676__$1) {
case "p":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"p","p",151049309)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "code":
if(cljs.core.truth_(attrs)){
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.components.code.code_block,attrs,cljs.core.first(children)], null);
} else {
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.components.code.inline_code,cljs.core.first(children)], null);
}

break;
case "diagram":
var G__42677 = new cljs.core.Keyword(null,"type","type",1174270348).cljs$core$IFn$_invoke$arity$1(attrs);
var G__42677__$1 = (((G__42677 instanceof cljs.core.Keyword))?G__42677.fqn:null);
switch (G__42677__$1) {
case "flowchart":
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.diagrams.flowchart.render,gl1tch.site.diagrams.flowchart.parse_inline(attrs,children)], null);

break;
default:
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),"Unknown diagram type"], null);

}

break;
case "a":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),attrs], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "strong":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"strong","strong",269529000)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "em":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"em","em",707813035)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "ul":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ul","ul",-1349521403)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "ol":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"ol","ol",932524051)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
case "li":
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"li","li",723558921)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));

break;
default:
if(cljs.core.truth_(attrs)){
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [tag,attrs], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));
} else {
return cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [tag], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,children));
}

}
} else {
return cljs.core.str.cljs$core$IFn$_invoke$arity$1(element);

}
}
}
});
gl1tch.site.pages.docs.render_section = (function gl1tch$site$pages$docs$render_section(p__42678){
var map__42679 = p__42678;
var map__42679__$1 = cljs.core.__destructure_map(map__42679);
var heading = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42679__$1,new cljs.core.Keyword(null,"heading","heading",-1312171873));
var level = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42679__$1,new cljs.core.Keyword(null,"level","level",1290497552));
var body = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__42679__$1,new cljs.core.Keyword(null,"body","body",-2049205669));
return new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"section","section",-300141526),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"id","id",-1388402092),gl1tch.site.pages.docs.heading__GT_id(heading),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-top","margin-top",392161226),((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(level,(2)))?"56px":"36px")], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [cljs.core.keyword.cljs$core$IFn$_invoke$arity$1(["h",cljs.core.str.cljs$core$IFn$_invoke$arity$1(level)].join('')),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"color","color",1011675173),((cljs.core._EQ_.cljs$core$IFn$_invoke$arity$2(level,(2)))?(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))):(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent-2","accent-2",-1373850285)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent-2","accent-2",-1373850285)))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"12px"], null)], null),heading], null),cljs.core.into.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632)], null),cljs.core.map.cljs$core$IFn$_invoke$arity$2(gl1tch.site.pages.docs.render_hiccup,body))], null);
});
gl1tch.site.pages.docs.page = (function gl1tch$site$pages$docs$page(slug){
var temp__5823__auto__ = gl1tch.site.content.doc_by_slug(slug);
if(cljs.core.truth_(temp__5823__auto__)){
var doc = temp__5823__auto__;
return new cljs.core.PersistentVector(null, 5, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 6, [new cljs.core.Keyword(null,"max-width","max-width",-1939924051),"1280px",new cljs.core.Keyword(null,"margin","margin",-995903681),"0 auto",new cljs.core.Keyword(null,"padding","padding",1660304693),"80px 40px 0",new cljs.core.Keyword(null,"display","display",242065432),new cljs.core.Keyword(null,"grid","grid",402978600),new cljs.core.Keyword(null,"grid-template-columns","grid-template-columns",-594112133),"220px 1fr 200px",new cljs.core.Keyword(null,"gap","gap",80255254),"0 48px"], null)], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.components.sidebar.sidebar], null)], null),new cljs.core.PersistentVector(null, 5, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"article","article",-21685045),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"min-width","min-width",1926193728),(0),new cljs.core.Keyword(null,"max-width","max-width",-1939924051),"720px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h1","h1",-1896887462),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"h2","h2",-372662728)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"h2","h2",-372662728))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"12px"], null)], null),new cljs.core.Keyword(null,"title","title",636505583).cljs$core$IFn$_invoke$arity$1(doc)], null),(cljs.core.truth_(new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc))?new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"p","p",151049309),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818))),new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"40px"], null)], null),new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc)], null):null),(function (){var iter__5480__auto__ = (function gl1tch$site$pages$docs$page_$_iter__42680(s__42681){
return (new cljs.core.LazySeq(null,(function (){
var s__42681__$1 = s__42681;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__42681__$1);
if(temp__5825__auto__){
var s__42681__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__42681__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__42681__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__42683 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__42682 = (0);
while(true){
if((i__42682 < size__5479__auto__)){
var section = cljs.core._nth(c__5478__auto__,i__42682);
cljs.core.chunk_append(b__42683,cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.docs.render_section,section], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"heading","heading",-1312171873).cljs$core$IFn$_invoke$arity$1(section)], null)));

var G__42690 = (i__42682 + (1));
i__42682 = G__42690;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__42683),gl1tch$site$pages$docs$page_$_iter__42680(cljs.core.chunk_rest(s__42681__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__42683),null);
}
} else {
var section = cljs.core.first(s__42681__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.docs.render_section,section], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"heading","heading",-1312171873).cljs$core$IFn$_invoke$arity$1(section)], null)),gl1tch$site$pages$docs$page_$_iter__42680(cljs.core.rest(s__42681__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(new cljs.core.Keyword(null,"sections","sections",-886710106).cljs$core$IFn$_invoke$arity$1(doc));
})()], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.components.toc.toc,gl1tch.site.content.extract_toc(doc)], null)], null)], null);
} else {
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div.wrap","div.wrap",1832950772),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h1","h1",-1896887462),"Doc not found"], null)], null);
}
});
gl1tch.site.pages.docs.index = (function gl1tch$site$pages$docs$index(){
return new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div.wrap","div.wrap",1832950772),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"padding-top","padding-top",1929675955),"80px"], null)], null),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h1","h1",-1896887462),"Documentation"], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-top","margin-top",392161226),"40px"], null)], null),(function (){var iter__5480__auto__ = (function gl1tch$site$pages$docs$index_$_iter__42684(s__42685){
return (new cljs.core.LazySeq(null,(function (){
var s__42685__$1 = s__42685;
while(true){
var temp__5825__auto__ = cljs.core.seq(s__42685__$1);
if(temp__5825__auto__){
var s__42685__$2 = temp__5825__auto__;
if(cljs.core.chunked_seq_QMARK_(s__42685__$2)){
var c__5478__auto__ = cljs.core.chunk_first(s__42685__$2);
var size__5479__auto__ = cljs.core.count(c__5478__auto__);
var b__42687 = cljs.core.chunk_buffer(size__5479__auto__);
if((function (){var i__42686 = (0);
while(true){
if((i__42686 < size__5479__auto__)){
var doc = cljs.core._nth(c__5478__auto__,i__42686);
cljs.core.chunk_append(b__42687,cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"24px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"href","href",-793805698),["#/docs/",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc))].join(''),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),"1.05rem",new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478)], null)], null),new cljs.core.Keyword(null,"title","title",636505583).cljs$core$IFn$_invoke$arity$1(doc)], null),(cljs.core.truth_(new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc))?new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"p","p",151049309),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065)))], null)], null),new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc)], null):null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)));

var G__42691 = (i__42686 + (1));
i__42686 = G__42691;
continue;
} else {
return true;
}
break;
}
})()){
return cljs.core.chunk_cons(cljs.core.chunk(b__42687),gl1tch$site$pages$docs$index_$_iter__42684(cljs.core.chunk_rest(s__42685__$2)));
} else {
return cljs.core.chunk_cons(cljs.core.chunk(b__42687),null);
}
} else {
var doc = cljs.core.first(s__42685__$2);
return cljs.core.cons(cljs.core.with_meta(new cljs.core.PersistentVector(null, 4, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div","div",1057191632),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"margin-bottom","margin-bottom",388334941),"24px"], null)], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"a","a",-2123407586),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"href","href",-793805698),["#/docs/",cljs.core.str.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc))].join(''),new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 4, [new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"accent","accent",-1826298468)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"accent","accent",-1826298468))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),"1.05rem",new cljs.core.Keyword(null,"font-weight","font-weight",2085804583),(700),new cljs.core.Keyword(null,"text-decoration","text-decoration",1836813207),new cljs.core.Keyword(null,"none","none",1333468478)], null)], null),new cljs.core.Keyword(null,"title","title",636505583).cljs$core$IFn$_invoke$arity$1(doc)], null),(cljs.core.truth_(new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc))?new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"p","p",151049309),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 2, [new cljs.core.Keyword(null,"color","color",1011675173),(gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.colors.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818)) : gl1tch.site.styles.theme.colors.call(null, new cljs.core.Keyword(null,"fg-dim","fg-dim",1664513818))),new cljs.core.Keyword(null,"font-size","font-size",-1847940346),(gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1 ? gl1tch.site.styles.theme.sizes.cljs$core$IFn$_invoke$arity$1(new cljs.core.Keyword(null,"sm","sm",-1402575065)) : gl1tch.site.styles.theme.sizes.call(null, new cljs.core.Keyword(null,"sm","sm",-1402575065)))], null)], null),new cljs.core.Keyword(null,"description","description",-1428560544).cljs$core$IFn$_invoke$arity$1(doc)], null):null)], null),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"key","key",-1516042587),new cljs.core.Keyword(null,"slug","slug",2029314850).cljs$core$IFn$_invoke$arity$1(doc)], null)),gl1tch$site$pages$docs$index_$_iter__42684(cljs.core.rest(s__42685__$2)));
}
} else {
return null;
}
break;
}
}),null,null));
});
return iter__5480__auto__(gl1tch.site.content.docs_index());
})()], null)], null);
});

//# sourceMappingURL=gl1tch.site.pages.docs.js.map
