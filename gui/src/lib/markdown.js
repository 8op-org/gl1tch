import 'highlight.js/styles/github-dark.min.css';
import { marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js/lib/core'
import json from 'highlight.js/lib/languages/json'
import bash from 'highlight.js/lib/languages/bash'
import yaml from 'highlight.js/lib/languages/yaml'
import python from 'highlight.js/lib/languages/python'
import go from 'highlight.js/lib/languages/go'

hljs.registerLanguage('json', json)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('python', python)
hljs.registerLanguage('go', go)

hljs.registerLanguage('glitch', () => ({
  name: 'glitch',
  contains: [
    { scope: 'comment', begin: ';;', end: '$' },
    { scope: 'string', begin: '```', end: '```', contains: [{ scope: 'template-variable', begin: '\\{\\{', end: '\\}\\}' }] },
    { scope: 'string', begin: '"', end: '"', contains: [{ begin: '\\\\.' }] },
    { scope: 'attr', begin: ':[a-zA-Z_][a-zA-Z0-9_-]*' },
    { scope: 'number', begin: '\\b[0-9]+\\b' },
    { scope: 'keyword', begin: '(?<=\\()\\s*(workflow|step|run|llm|phase|gate|retry|timeout|par|when|each|def|plugin|env)\\b' },
    { scope: 'title.function', begin: '(?<=\\()\\s*[a-zA-Z_][a-zA-Z0-9_-]*' },
    { scope: 'punctuation', begin: '[()]' },
  ],
}))

marked.use(markedHighlight({
  langPrefix: 'hljs language-',
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  },
}))

export function renderMarkdown(src) {
  return marked.parse(src)
}
