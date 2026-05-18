package render

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var sourcePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

type Renderer struct {
	resolvers map[string]Resolver
}

func NewRenderer(resolvers ...Resolver) *Renderer {
	r := &Renderer{resolvers: make(map[string]Resolver)}
	for _, resolver := range resolvers {
		if resolver == nil {
			continue
		}
		r.resolvers[resolver.Source()] = resolver
	}
	return r
}

func (r *Renderer) Render(raw string) (string, error) {
	var (
		builder strings.Builder
		errs    = make(map[string]struct{})
	)

	for i := 0; i < len(raw); {
		switch {
		case strings.HasPrefix(raw[i:], `$${`):
			builder.WriteString(`${`)
			i += len(`$${`)
			continue
		case strings.HasPrefix(raw[i:], `\${`):
			builder.WriteString(`${`)
			i += len(`\${`)
			continue
		case strings.HasPrefix(raw[i:], `${`):
			end := strings.IndexByte(raw[i+2:], '}')
			if end < 0 {
				addError(errs, "invalid placeholder: missing closing }")
				builder.WriteByte(raw[i])
				i++
				continue
			}

			placeholder := raw[i : i+2+end+1]
			body := raw[i+2 : i+2+end]
			if !strings.Contains(body, ":") {
				builder.WriteString(placeholder)
				i += len(placeholder)
				continue
			}

			source, key, ok := strings.Cut(body, ":")
			if !ok || source == "" || key == "" || !sourcePattern.MatchString(source) {
				addError(errs, fmt.Sprintf("invalid placeholder %s", placeholder))
				builder.WriteString(placeholder)
				i += len(placeholder)
				continue
			}

			resolver, ok := r.resolvers[source]
			if !ok {
				addError(errs, fmt.Sprintf("unsupported placeholder source %q", source))
				builder.WriteString(placeholder)
				i += len(placeholder)
				continue
			}

			value, found, err := resolver.Resolve(key)
			if err != nil {
				addError(errs, fmt.Sprintf("resolve %s:%s: %v", source, key, err))
				builder.WriteString(placeholder)
				i += len(placeholder)
				continue
			}
			if !found {
				addError(errs, fmt.Sprintf("missing value for %s:%s", source, key))
				builder.WriteString(placeholder)
				i += len(placeholder)
				continue
			}

			builder.WriteString(value)
			i += len(placeholder)
			continue
		default:
			builder.WriteByte(raw[i])
			i++
		}
	}

	if len(errs) > 0 {
		return "", joinErrors(errs)
	}
	return builder.String(), nil
}

func addError(errs map[string]struct{}, message string) {
	errs[message] = struct{}{}
}

func joinErrors(errs map[string]struct{}) error {
	messages := make([]string, 0, len(errs))
	for message := range errs {
		messages = append(messages, message)
	}
	sort.Strings(messages)
	return errors.New(strings.Join(messages, "; "))
}
