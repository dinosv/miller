// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"miller/parsing/token"
)

const (
	NoState    = -1
	NumStates  = 312
	NumSymbols = 536
)

type Lexer struct {
	src    []byte
	pos    int
	line   int
	column int
}

func NewLexer(src []byte) *Lexer {
	lexer := &Lexer{
		src:    src,
		pos:    0,
		line:   1,
		column: 1,
	}
	return lexer
}

func NewLexerFile(fpath string) (*Lexer, error) {
	src, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return NewLexer(src), nil
}

func (l *Lexer) Scan() (tok *token.Token) {
	tok = new(token.Token)
	if l.pos >= len(l.src) {
		tok.Type = token.EOF
		tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = l.pos, l.line, l.column
		return
	}
	start, startLine, startColumn, end := l.pos, l.line, l.column, 0
	tok.Type = token.INVALID
	state, rune1, size := 0, rune(-1), 0
	for state != -1 {
		if l.pos >= len(l.src) {
			rune1 = -1
		} else {
			rune1, size = utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}

		nextState := -1
		if rune1 != -1 {
			nextState = TransTab[state](rune1)
		}
		state = nextState

		if state != -1 {

			switch rune1 {
			case '\n':
				l.line++
				l.column = 1
			case '\r':
				l.column = 1
			case '\t':
				l.column += 4
			default:
				l.column++
			}

			switch {
			case ActTab[state].Accept != -1:
				tok.Type = ActTab[state].Accept
				end = l.pos
			case ActTab[state].Ignore != "":
				start, startLine, startColumn = l.pos, l.line, l.column
				state = 0
				if start >= len(l.src) {
					tok.Type = token.EOF
				}

			}
		} else {
			if tok.Type == token.INVALID {
				end = l.pos
			}
		}
	}
	if end > start {
		l.pos = end
		tok.Lit = l.src[start:end]
	} else {
		tok.Lit = []byte{}
	}
	tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = start, startLine, startColumn

	return
}

func (l *Lexer) Reset() {
	l.pos = 0
}

/*
Lexer symbols:
0: '"'
1: '"'
2: '0'
3: 'x'
4: '0'
5: 'b'
6: '.'
7: '.'
8: '-'
9: '.'
10: '.'
11: '-'
12: '.'
13: '.'
14: '-'
15: 'M'
16: '_'
17: 'P'
18: 'I'
19: 'M'
20: '_'
21: 'E'
22: 'I'
23: 'P'
24: 'S'
25: 'I'
26: 'F'
27: 'S'
28: 'I'
29: 'R'
30: 'S'
31: 'O'
32: 'P'
33: 'S'
34: 'O'
35: 'F'
36: 'S'
37: 'O'
38: 'R'
39: 'S'
40: 'O'
41: 'F'
42: 'L'
43: 'A'
44: 'T'
45: 'S'
46: 'E'
47: 'P'
48: 'N'
49: 'F'
50: 'N'
51: 'R'
52: 'F'
53: 'N'
54: 'R'
55: 'F'
56: 'I'
57: 'L'
58: 'E'
59: 'N'
60: 'A'
61: 'M'
62: 'E'
63: 'F'
64: 'I'
65: 'L'
66: 'E'
67: 'N'
68: 'U'
69: 'M'
70: 'E'
71: 'N'
72: 'V'
73: 'b'
74: 'e'
75: 'g'
76: 'i'
77: 'n'
78: 'd'
79: 'o'
80: 'e'
81: 'l'
82: 'i'
83: 'f'
84: 'e'
85: 'l'
86: 's'
87: 'e'
88: 'e'
89: 'n'
90: 'd'
91: 'f'
92: 'i'
93: 'l'
94: 't'
95: 'e'
96: 'r'
97: 'f'
98: 'o'
99: 'r'
100: 'i'
101: 'f'
102: 'i'
103: 'n'
104: 'w'
105: 'h'
106: 'i'
107: 'l'
108: 'e'
109: 'b'
110: 'r'
111: 'e'
112: 'a'
113: 'k'
114: 'c'
115: 'o'
116: 'n'
117: 't'
118: 'i'
119: 'n'
120: 'u'
121: 'e'
122: 'r'
123: 'e'
124: 't'
125: 'u'
126: 'r'
127: 'n'
128: 'f'
129: 'u'
130: 'n'
131: 'c'
132: 's'
133: 'u'
134: 'b'
135: 'r'
136: 'c'
137: 'a'
138: 'l'
139: 'l'
140: 'a'
141: 'r'
142: 'r'
143: 'b'
144: 'o'
145: 'o'
146: 'l'
147: 'f'
148: 'l'
149: 'o'
150: 'a'
151: 't'
152: 'i'
153: 'n'
154: 't'
155: 'm'
156: 'a'
157: 'p'
158: 'n'
159: 'u'
160: 'm'
161: 's'
162: 't'
163: 'r'
164: 'v'
165: 'a'
166: 'r'
167: 'u'
168: 'n'
169: 's'
170: 'e'
171: 't'
172: 'd'
173: 'u'
174: 'm'
175: 'p'
176: 'e'
177: 'd'
178: 'u'
179: 'm'
180: 'p'
181: 'e'
182: 'm'
183: 'i'
184: 't'
185: 'e'
186: 'm'
187: 'i'
188: 't'
189: 'p'
190: 'e'
191: 'm'
192: 'i'
193: 't'
194: 'f'
195: 'e'
196: 'p'
197: 'r'
198: 'i'
199: 'n'
200: 't'
201: 'e'
202: 'p'
203: 'r'
204: 'i'
205: 'n'
206: 't'
207: 'n'
208: 'p'
209: 'r'
210: 'i'
211: 'n'
212: 't'
213: 'p'
214: 'r'
215: 'i'
216: 'n'
217: 't'
218: 'n'
219: 't'
220: 'e'
221: 'e'
222: 's'
223: 't'
224: 'd'
225: 'o'
226: 'u'
227: 't'
228: 's'
229: 't'
230: 'd'
231: 'e'
232: 'r'
233: 'r'
234: '$'
235: '$'
236: '{'
237: '}'
238: '$'
239: '*'
240: '@'
241: '@'
242: '{'
243: '}'
244: '@'
245: '*'
246: 'a'
247: 'l'
248: 'l'
249: '%'
250: '%'
251: '%'
252: 'p'
253: 'a'
254: 'n'
255: 'i'
256: 'c'
257: '%'
258: '%'
259: '%'
260: ';'
261: '{'
262: '}'
263: '='
264: '>'
265: '>'
266: '>'
267: '|'
268: ','
269: '('
270: ')'
271: '$'
272: '['
273: ']'
274: '$'
275: '['
276: '['
277: '$'
278: '['
279: '['
280: '['
281: '@'
282: '['
283: '|'
284: '|'
285: '='
286: '^'
287: '^'
288: '='
289: '&'
290: '&'
291: '='
292: '?'
293: '?'
294: '='
295: '?'
296: '?'
297: '?'
298: '='
299: '|'
300: '='
301: '&'
302: '='
303: '^'
304: '='
305: '<'
306: '<'
307: '='
308: '>'
309: '>'
310: '='
311: '>'
312: '>'
313: '>'
314: '='
315: '+'
316: '='
317: '.'
318: '='
319: '-'
320: '='
321: '*'
322: '='
323: '/'
324: '='
325: '/'
326: '/'
327: '='
328: '%'
329: '='
330: '*'
331: '*'
332: '='
333: '?'
334: ':'
335: '|'
336: '|'
337: '^'
338: '^'
339: '&'
340: '&'
341: '?'
342: '?'
343: '?'
344: '?'
345: '?'
346: '='
347: '~'
348: '!'
349: '='
350: '~'
351: '='
352: '='
353: '!'
354: '='
355: '>'
356: '='
357: '<'
358: '<'
359: '='
360: '^'
361: '&'
362: '<'
363: '<'
364: '>'
365: '>'
366: '>'
367: '+'
368: '-'
369: '.'
370: '+'
371: '.'
372: '-'
373: '.'
374: '*'
375: '/'
376: '/'
377: '/'
378: '%'
379: '.'
380: '*'
381: '.'
382: '/'
383: '.'
384: '/'
385: '/'
386: '!'
387: '~'
388: '*'
389: '*'
390: '['
391: '['
392: '['
393: '['
394: '['
395: '['
396: '_'
397: ' '
398: '!'
399: '#'
400: '$'
401: '%'
402: '&'
403: '''
404: '\'
405: '('
406: ')'
407: '*'
408: '+'
409: ','
410: '-'
411: '.'
412: '/'
413: ':'
414: ';'
415: '<'
416: '='
417: '>'
418: '?'
419: '@'
420: '['
421: ']'
422: '^'
423: '_'
424: '`'
425: '{'
426: '|'
427: '}'
428: '~'
429: '\'
430: '\'
431: '\'
432: '"'
433: '\'
434: '['
435: '\'
436: ']'
437: '\'
438: 'b'
439: '\'
440: 'f'
441: '\'
442: 'n'
443: '\'
444: 'r'
445: '\'
446: 't'
447: '\'
448: 'x'
449: '\'
450: '0'
451: '\'
452: '1'
453: '\'
454: '2'
455: '\'
456: '3'
457: '\'
458: '4'
459: '\'
460: '5'
461: '\'
462: '6'
463: '\'
464: '7'
465: '\'
466: '8'
467: '\'
468: '9'
469: 'e'
470: 'E'
471: 't'
472: 'r'
473: 'u'
474: 'e'
475: 'f'
476: 'a'
477: 'l'
478: 's'
479: 'e'
480: ' '
481: '!'
482: '#'
483: '$'
484: '%'
485: '&'
486: '''
487: '\'
488: '('
489: ')'
490: '*'
491: '+'
492: ','
493: '-'
494: '.'
495: '/'
496: ':'
497: ';'
498: '<'
499: '='
500: '>'
501: '?'
502: '@'
503: '['
504: ']'
505: '^'
506: '_'
507: '`'
508: '|'
509: '~'
510: '\'
511: '{'
512: '\'
513: '}'
514: ' '
515: '\t'
516: '\n'
517: '\r'
518: '#'
519: '\n'
520: 'a'-'z'
521: 'A'-'Z'
522: '0'-'9'
523: '0'-'9'
524: 'a'-'f'
525: 'A'-'F'
526: '0'-'1'
527: 'A'-'Z'
528: 'a'-'z'
529: '0'-'9'
530: \u0100-\U0010ffff
531: 'A'-'Z'
532: 'a'-'z'
533: '0'-'9'
534: \u0100-\U0010ffff
535: .
*/
