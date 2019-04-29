package grok

/*
Apache License Version 2.0, January 2004
see: https://github.com/vjeantet/grok/blob/master/LICENSE
*/

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	canonical = regexp.MustCompile(`%{(\w+(?::\w+(?::\w+)?)?)}`)
	normal    = regexp.MustCompile(`%{([\w-.]+(?::[\w-.]+(?::[\w-.]+)?)?)}`)
	symbolic  = regexp.MustCompile(`\W`)
)

// A Config structure is used to configure a Grok parser.
type Config struct {
	NamedCapturesOnly   bool
	SkipDefaultPatterns bool
	RemoveEmptyValues   bool
	Patterns            map[string]string
}

// Grok object us used to load patterns and deconstruct strings using those
// patterns.
type Grok struct {
	rawPattern       map[string]string
	config           *Config
	aliases          map[string]string
	compiledPatterns map[string]*gRegexp
	patterns         map[string]*gPattern
	patternsGuard    *sync.RWMutex
	compiledGuard    *sync.RWMutex
}

type gPattern struct {
	expression string
	typeInfo   semanticTypes
}

type gRegexp struct {
	regexp   *regexp.Regexp
	typeInfo semanticTypes
}

type semanticTypes map[string]string

// New returns a Grok object.
func New() (*Grok, error) {
	return NewWithConfig(&Config{SkipDefaultPatterns: false})
}

// NewWithConfig returns a Grok object that is configured to behave according
// to the supplied Config structure.
func NewWithConfig(config *Config) (*Grok, error) {
	g := &Grok{
		config:           config,
		aliases:          map[string]string{},
		compiledPatterns: map[string]*gRegexp{},
		patterns:         map[string]*gPattern{},
		rawPattern:       map[string]string{},
		patternsGuard:    new(sync.RWMutex),
		compiledGuard:    new(sync.RWMutex),
	}

	if !config.SkipDefaultPatterns {
		g.AddPatternsFromMap(default_patterns)
	}

	if err := g.AddPatternsFromMap(config.Patterns); err != nil {
		return nil, err
	}

	return g, nil
}

// AddPattern adds a new pattern to the list of loaded patterns.
func (g *Grok) addPattern(name, pattern string) error {
	dnPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	if err != nil {
		return err
	}

	g.patterns[name] = &gPattern{expression: dnPattern, typeInfo: ti}
	return nil
}

// AddPattern adds a named pattern to grok
func (g *Grok) AddPattern(name, pattern string) error {
	g.patternsGuard.Lock()
	defer g.patternsGuard.Unlock()

	g.rawPattern[name] = pattern
	g.buildPatterns()
	return nil
}

// AddPatternsFromMap loads a map of named patterns
func (g *Grok) AddPatternsFromMap(m map[string]string) error {
	g.patternsGuard.Lock()
	defer g.patternsGuard.Unlock()

	for name, pattern := range m {
		g.rawPattern[name] = pattern
	}
	return g.buildPatterns()
}

// AddPatternsFromMap adds new patterns from the specified map to the list of
// loaded patterns.
func (g *Grok) addPatternsFromMap(m map[string]string) error {
	patternDeps := graph{}
	for k, v := range m {
		keys := []string{}
		for _, key := range canonical.FindAllStringSubmatch(v, -1) {
			names := strings.Split(key[1], ":")
			syntax := names[0]
			if g.patterns[syntax] == nil {
				if _, ok := m[syntax]; !ok {
					return fmt.Errorf("no pattern found for %%{%s}", syntax)
				}
			}
			keys = append(keys, syntax)
		}
		patternDeps[k] = keys
	}
	order, _ := sortGraph(patternDeps)
	for _, key := range reverseList(order) {
		g.addPattern(key, m[key])
	}

	return nil
}

// Match returns true if the specified text matches the pattern.
func (g *Grok) Match(pattern, text string) (bool, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return false, err
	}

	if ok := gr.regexp.MatchString(text); !ok {
		return false, nil
	}

	return true, nil
}

// compiledParse parses the specified text and returns a map with the results.
func (g *Grok) compiledParse(gr *gRegexp, text string) (map[string]string, error) {
	captures := make(map[string]string)
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				if g.config.RemoveEmptyValues && match[i] == "" {
					continue
				}
				name = g.nameToAlias(name)
				captures[name] = match[i]
			}
		}
	}

	return captures, nil
}

// Parse the specified text and return a map with the results.
func (g *Grok) Parse(pattern, text string) (map[string]string, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}

	return g.compiledParse(gr, text)
}

// ParseTyped returns a inteface{} map with typed captured fields based on provided pattern over the text
func (g *Grok) ParseTyped(pattern string, text string) (map[string]interface{}, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}
	match := gr.regexp.FindStringSubmatch(text)
	captures := make(map[string]interface{})
	if len(match) > 0 {
		for i, segmentName := range gr.regexp.SubexpNames() {
			if len(segmentName) != 0 {
				if g.config.RemoveEmptyValues == true && match[i] == "" {
					continue
				}
				name := g.nameToAlias(segmentName)
				if segmentType, ok := gr.typeInfo[segmentName]; ok {
					switch segmentType {
					case "int":
						captures[name], _ = strconv.Atoi(match[i])
					case "float":
						captures[name], _ = strconv.ParseFloat(match[i], 64)
					default:
						return nil, fmt.Errorf("ERROR the value %s cannot be converted to %s", match[i], segmentType)
					}
				} else {
					captures[name] = match[i]
				}
			}

		}
	}

	return captures, nil
}

// ParseToMultiMap parses the specified text and returns a map with the
// results. Values are stored in an string slice, so values from captures with
// the same name don't get overridden.
func (g *Grok) ParseToMultiMap(pattern, text string) (map[string][]string, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}

	captures := make(map[string][]string)
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				if g.config.RemoveEmptyValues == true && match[i] == "" {
					continue
				}
				name = g.nameToAlias(name)
				captures[name] = append(captures[name], match[i])
			}
		}
	}

	return captures, nil
}

func (g *Grok) buildPatterns() error {
	g.patterns = map[string]*gPattern{}
	return g.addPatternsFromMap(g.rawPattern)
}

func (g *Grok) compile(pattern string) (*gRegexp, error) {
	g.compiledGuard.RLock()
	gr, ok := g.compiledPatterns[pattern]
	g.compiledGuard.RUnlock()

	if ok {
		return gr, nil
	}

	g.patternsGuard.RLock()
	newPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	g.patternsGuard.RUnlock()
	if err != nil {
		return nil, err
	}

	compiledRegex, err := regexp.Compile(newPattern)
	if err != nil {
		return nil, err
	}
	gr = &gRegexp{regexp: compiledRegex, typeInfo: ti}

	g.compiledGuard.Lock()
	g.compiledPatterns[pattern] = gr
	g.compiledGuard.Unlock()

	return gr, nil
}

func (g *Grok) denormalizePattern(pattern string, storedPatterns map[string]*gPattern) (string, semanticTypes, error) {
	ti := semanticTypes{}
	for _, values := range normal.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		syntax, semantic, alias := names[0], names[0], names[0]
		if len(names) > 1 {
			semantic = names[1]
			alias = g.aliasizePatternName(semantic)
		}

		// Add type cast information only if type set, and not string
		if len(names) == 3 {
			if names[2] != "string" {
				ti[semantic] = names[2]
			}
		}

		storedPattern, ok := storedPatterns[syntax]
		if !ok {
			return "", ti, fmt.Errorf("no pattern found for %%{%s}", syntax)
		}

		var buffer bytes.Buffer
		if !g.config.NamedCapturesOnly || (g.config.NamedCapturesOnly && len(names) > 1) {
			buffer.WriteString("(?P<")
			buffer.WriteString(alias)
			buffer.WriteString(">")
			buffer.WriteString(storedPattern.expression)
			buffer.WriteString(")")
		} else {
			buffer.WriteString("(")
			buffer.WriteString(storedPattern.expression)
			buffer.WriteString(")")
		}

		//Merge type Informations
		for k, v := range storedPattern.typeInfo {
			//Lastest type information is the one to keep in memory
			if _, ok := ti[k]; !ok {
				ti[k] = v
			}
		}

		pattern = strings.Replace(pattern, values[0], buffer.String(), -1)
	}

	return pattern, ti, nil

}

func (g *Grok) aliasizePatternName(name string) string {
	alias := symbolic.ReplaceAllString(name, "_")
	g.aliases[alias] = name
	return alias
}

func (g *Grok) nameToAlias(name string) string {
	alias, ok := g.aliases[name]
	if ok {
		return alias
	}
	return name
}

// ParseStream will match the given pattern on a line by line basis from the reader
// and apply the results to the process function
func (g *Grok) ParseStream(reader *bufio.Reader, pattern string, process func(map[string]string) error) error {
	gr, err := g.compile(pattern)
	if err != nil {
		return err
	}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		values, err := g.compiledParse(gr, line)
		if err != nil {
			return err
		}
		if err = process(values); err != nil {
			return err
		}
	}
}
