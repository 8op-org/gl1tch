import { StreamLanguage } from '@codemirror/language';
import { tags } from '@lezer/highlight';

const glitchStreamParser = {
  startState() {
    return { inTripleBacktick: false };
  },

  token(stream, state) {
    // Triple-backtick strings
    if (state.inTripleBacktick) {
      if (stream.match('```')) { state.inTripleBacktick = false; return 'string'; }
      // Template variables inside triple-backtick
      if (stream.match(/\{\{[^}]*\}\}/)) return 'variableName';
      stream.next();
      return 'string';
    }
    if (stream.match('```')) { state.inTripleBacktick = true; return 'string'; }

    // Line comments: ;;
    if (stream.match(/;;.*/)) return 'lineComment';

    // Strings
    if (stream.match(/"(?:[^"\\]|\\.)*"/)) return 'string';

    // Keyword arguments :foo
    if (stream.match(/:[a-zA-Z_][a-zA-Z0-9_-]*/)) return 'attributeName';

    // Numbers
    if (stream.match(/\b[0-9]+\b/)) return 'number';

    // Open paren followed by form name
    if (stream.match('(')) {
      // Peek at following word
      const m = stream.match(/\s*(workflow|step|run|llm|phase|gate|retry|timeout|par|when|each|def|plugin|env)\b/, false);
      return 'paren';
    }
    if (stream.match(')')) return 'paren';

    // Keywords (form names right after open paren — handled via lookahead after paren consumed above)
    if (stream.match(/\b(workflow|step|run|llm|phase|gate|retry|timeout|par|when|each|def|plugin|env)\b/)) return 'keyword';

    // Identifiers (function-like names)
    if (stream.match(/[a-zA-Z_][a-zA-Z0-9_-]*/)) return 'variableName.definition';

    // Skip whitespace
    stream.next();
    return null;
  },
};

export const glitchLanguage = StreamLanguage.define(glitchStreamParser);
