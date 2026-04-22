goog.provide('gl1tch.site.core');
gl1tch.site.core.current_page = (function gl1tch$site$core$current_page(){
var match = cljs.core.deref(gl1tch.site.state.current_match);
var name = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(match,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"data","data",-232669377),new cljs.core.Keyword(null,"name","name",1843675177)], null));
var slug = cljs.core.get_in.cljs$core$IFn$_invoke$arity$2(match,new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"path-params","path-params",-48130597),new cljs.core.Keyword(null,"slug","slug",2029314850)], null));
return new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div.app","div.app",-99849286),new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.components.nav.nav], null),new cljs.core.PersistentVector(null, 3, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"main","main",-2117802661),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"style","style",-496642736),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"padding-top","padding-top",1929675955),"52px"], null)], null),(function (){var G__42622 = name;
var G__42622__$1 = (((G__42622 instanceof cljs.core.Keyword))?G__42622.fqn:null);
switch (G__42622__$1) {
case "home":
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.home.page], null);

break;
case "docs/index":
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.docs.index], null);

break;
case "docs/page":
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.docs.page,slug], null);

break;
case "labs/index":
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.labs.index], null);

break;
case "labs/page":
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.labs.page,slug], null);

break;
case "changelog":
return new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.pages.changelog.page], null);

break;
default:
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"div.wrap","div.wrap",1832950772),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"h1","h1",-1896887462),"404"], null)], null);

}
})()], null)], null);
});
gl1tch.site.core.init = (function gl1tch$site$core$init(){
gl1tch.site.styles.global.inject_BANG_();

gl1tch.site.router.start_BANG_();

return reagent.dom.render.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.core.current_page], null),document.getElementById("app"));
});
goog.exportSymbol('gl1tch.site.core.init', gl1tch.site.core.init);
gl1tch.site.core.reload = (function gl1tch$site$core$reload(){
gl1tch.site.styles.global.inject_BANG_();

return reagent.dom.render.cljs$core$IFn$_invoke$arity$2(new cljs.core.PersistentVector(null, 1, 5, cljs.core.PersistentVector.EMPTY_NODE, [gl1tch.site.core.current_page], null),document.getElementById("app"));
});

//# sourceMappingURL=gl1tch.site.core.js.map
