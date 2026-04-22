goog.provide('gl1tch.site.components.code');
var module$node_modules$highlight_DOT_js$lib$core=shadow.js.require("module$node_modules$highlight_DOT_js$lib$core", {});
var module$node_modules$highlight_DOT_js$lib$languages$bash=shadow.js.require("module$node_modules$highlight_DOT_js$lib$languages$bash", {});
var module$node_modules$highlight_DOT_js$lib$languages$clojure=shadow.js.require("module$node_modules$highlight_DOT_js$lib$languages$clojure", {});
var module$node_modules$highlight_DOT_js$lib$languages$json=shadow.js.require("module$node_modules$highlight_DOT_js$lib$languages$json", {});
var module$node_modules$highlight_DOT_js$lib$languages$yaml=shadow.js.require("module$node_modules$highlight_DOT_js$lib$languages$yaml", {});
if((typeof gl1tch !== 'undefined') && (typeof gl1tch.site !== 'undefined') && (typeof gl1tch.site.components !== 'undefined') && (typeof gl1tch.site.components.code !== 'undefined') && (typeof gl1tch.site.components.code._register !== 'undefined')){
} else {
gl1tch.site.components.code._register = (function (){
module$node_modules$highlight_DOT_js$lib$core.registerLanguage("bash",module$node_modules$highlight_DOT_js$lib$languages$bash);

module$node_modules$highlight_DOT_js$lib$core.registerLanguage("clojure",module$node_modules$highlight_DOT_js$lib$languages$clojure);

module$node_modules$highlight_DOT_js$lib$core.registerLanguage("json",module$node_modules$highlight_DOT_js$lib$languages$json);

module$node_modules$highlight_DOT_js$lib$core.registerLanguage("yaml",module$node_modules$highlight_DOT_js$lib$languages$yaml);

return true;
})()
;
}
gl1tch.site.components.code.highlight = (function gl1tch$site$components$code$highlight(code,lang){
if(cljs.core.truth_((function (){var and__5000__auto__ = lang;
if(cljs.core.truth_(and__5000__auto__)){
return module$node_modules$highlight_DOT_js$lib$core.getLanguage(lang);
} else {
return and__5000__auto__;
}
})())){
return module$node_modules$highlight_DOT_js$lib$core.highlight(code,({"language": lang})).value;
} else {
return code;
}
});
/**
 * Renders a syntax-highlighted code block.
 * Usage in hiccup EDN: [:code {:lang "bash"} "echo hello"]
 * Note: dangerouslySetInnerHTML is safe here — content comes from
 * Highlight.js processing our own EDN, never user input.
 */
gl1tch.site.components.code.code_block = (function gl1tch$site$components$code$code_block(p__41613,code){
var map__41614 = p__41613;
var map__41614__$1 = cljs.core.__destructure_map(map__41614);
var lang = cljs.core.get.cljs$core$IFn$_invoke$arity$2(map__41614__$1,new cljs.core.Keyword(null,"lang","lang",-1819677104));
var highlighted = gl1tch.site.components.code.highlight(code,(function (){var or__5002__auto__ = lang;
if(cljs.core.truth_(or__5002__auto__)){
return or__5002__auto__;
} else {
return "plaintext";
}
})());
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"pre.code","pre.code",2043838796),new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"code","code",1586293142),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"dangerouslySetInnerHTML","dangerouslySetInnerHTML",-554971138),new cljs.core.PersistentArrayMap(null, 1, [new cljs.core.Keyword(null,"__html","__html",674048345),highlighted], null)], null)], null)], null);
});
/**
 * Renders inline code. Usage: [:code "some-thing"]
 */
gl1tch.site.components.code.inline_code = (function gl1tch$site$components$code$inline_code(text){
return new cljs.core.PersistentVector(null, 2, 5, cljs.core.PersistentVector.EMPTY_NODE, [new cljs.core.Keyword(null,"code","code",1586293142),text], null);
});

//# sourceMappingURL=gl1tch.site.components.code.js.map
