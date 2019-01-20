package regex

import (
	"regexp"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/processors"
)

type Regex struct {
	Tags       []converter
	Fields     []converter
	regexCache map[string]*regexp.Regexp
}

type converter struct {
	Key         string
	Pattern     string
	Replacement string
	ResultKey   string
}

func NewRegex() *Regex {
	return &Regex{
		regexCache: make(map[string]*regexp.Regexp),
	}
}

func (r *Regex) Apply(in ...telex.Metric) []telex.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if value, ok := metric.GetTag(converter.Key); ok {
				if key, newValue := r.convert(converter, value); newValue != "" {
					metric.AddTag(key, newValue)
				}
			}
		}

		for _, converter := range r.Fields {
			if value, ok := metric.GetField(converter.Key); ok {
				switch value := value.(type) {
				case string:
					if key, newValue := r.convert(converter, value); newValue != "" {
						metric.AddField(key, newValue)
					}
				}
			}
		}
	}

	return in
}

func (r *Regex) convert(c converter, src string) (string, string) {
	regex, compiled := r.regexCache[c.Pattern]
	if !compiled {
		regex = regexp.MustCompile(c.Pattern)
		r.regexCache[c.Pattern] = regex
	}

	value := ""
	if c.ResultKey == "" || regex.MatchString(src) {
		value = regex.ReplaceAllString(src, c.Replacement)
	}

	if c.ResultKey != "" {
		return c.ResultKey, value
	}

	return c.Key, value
}

func init() {
	processors.Add("regex", func() telex.Processor {
		return NewRegex()
	})
}
