package plugin

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"os"
	"regexp"
	"runtime"
	"strings"
)

func loadGeositeFromFile(fp string, crm map[string]struct{}) (map[string][]matcher, error) {
	defer quickGC()()
	file, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	defer func() {
		runtime.GC()
		runtime.GC()
	}()
	geosite, err := loadGeosite(file, crm)
	if err != nil {
		return nil, err
	}
	return geosite, nil
}

func loadGeosite(b []byte, crm map[string]struct{}) (map[string][]matcher, error) {
	var dl GeoSiteList
	err := proto.Unmarshal(b, &dl)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]matcher, len(crm))
	ccm := make(map[string]map[string]*regexp.Regexp)
	for cr := range crm {
		before, after, _ := strings.Cut(cr, ":")
		cm, ok := ccm[before]
		if !ok {
			cm = make(map[string]*regexp.Regexp)
			ccm[before] = cm
		}
		if after != "" {
			compile, err := regexp.Compile(after)
			if err != nil {
				return nil, err
			}
			cm[cr] = compile
		} else {
			cm[cr] = nil
		}
	}
	for _, site := range dl.Entry {
		cm, ok := ccm[site.CountryCode]
		if !ok {
			continue
		}
		for _, domain := range site.Domain {
			if domain.Value == "" {
				continue
			}
			dmatcher, err := getDomainTypeMatcher(domain.Type, domain.Value)
			if err != nil {
				continue
			}
			for k, r := range cm {
				if r == nil {
					m[k] = append(m[k], dmatcher)
					continue
				}
				for _, attribute := range domain.Attribute {
					if r.MatchString(attribute.GetKey()) {
						m[k] = append(m[k], dmatcher)
						break
					}
				}
			}
		}
	}
	return m, nil
}

func getDomainTypeMatcher(t Domain_Type, str string) (matcher, error) {
	switch t {
	case Domain_Plain:
		m := substrMatcher(str)
		return m, nil
	case Domain_Regex:
		compile, err := regexp.Compile(str)
		if err != nil {
			return nil, err
		}
		return &regexMatcher{r: compile}, nil
	case Domain_Domain:
		m := domainMatcher(str)
		return m, nil
	case Domain_Full:
		m := fullMatcher(str)
		return m, nil
	default:
		return nil, fmt.Errorf("invalid domain type: %s", t.String())
	}
}

type matcher interface {
	match(string) bool
	String() string
}

type fullMatcher string

func (m fullMatcher) String() string {
	return string(m)
}

func (m fullMatcher) match(s string) bool {
	return string(m) == s
}

type substrMatcher string

func (m substrMatcher) String() string {
	return string(m)
}

func (m substrMatcher) match(s string) bool {
	return strings.Contains(s, string(m))
}

type domainMatcher string

func (m domainMatcher) String() string {
	return string(m)
}

func (m domainMatcher) match(s string) bool {
	pattern := string(m)
	if !strings.HasSuffix(s, pattern) {
		return false
	}
	return len(s) == len(pattern) || s[len(s)-len(pattern)-1] == '.'
}

type regexMatcher struct {
	r *regexp.Regexp
}

func (m *regexMatcher) String() string {
	return m.r.String()
}

func (m *regexMatcher) match(s string) bool {
	return m.r.MatchString(s)
}
