package modelparam

import (
	"fmt"
	"strings"
)

type PathSegment struct {
	Name     string
	Index    *int
	Wildcard bool
}

type CompiledPath struct {
	Original    string
	GJSON       string
	Segments    []PathSegment
	HasWildcard bool
}

func CompilePath(path string) (CompiledPath, error) {
	if len(path) < 3 || len(path) > 512 || path[0] != '$' {
		return CompiledPath{}, fmt.Errorf("invalid JSONPath")
	}
	if strings.Contains(path, "..") || strings.ContainsAny(path, "?:,") {
		return CompiledPath{}, fmt.Errorf("unsupported JSONPath expression")
	}
	var segments []PathSegment
	for i := 1; i < len(path); {
		if path[i] == '.' {
			i++
			start := i
			for i < len(path) && path[i] != '.' && path[i] != '[' {
				i++
			}
			if start == i {
				return CompiledPath{}, fmt.Errorf("empty JSONPath segment")
			}
			segments = append(segments, PathSegment{Name: path[start:i]})
			continue
		}
		if path[i] != '[' {
			return CompiledPath{}, fmt.Errorf("invalid JSONPath segment")
		}
		end := strings.IndexByte(path[i:], ']')
		if end < 0 {
			return CompiledPath{}, fmt.Errorf("unterminated JSONPath bracket")
		}
		end += i
		part := path[i+1 : end]
		switch {
		case part == "*":
			segments = append(segments, PathSegment{Wildcard: true})
		case strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") && len(part) > 2:
			segments = append(segments, PathSegment{Name: part[1 : len(part)-1]})
		default:
			var n int
			if _, err := fmt.Sscanf(part, "%d", &n); err != nil || n < 0 || fmt.Sprintf("%d", n) != part {
				return CompiledPath{}, fmt.Errorf("invalid JSONPath array index")
			}
			segments = append(segments, PathSegment{Index: &n})
		}
		i = end + 1
	}
	if len(segments) == 0 || len(segments) > 32 {
		return CompiledPath{}, fmt.Errorf("invalid JSONPath depth")
	}
	var parts []string
	for _, segment := range segments {
		switch {
		case segment.Wildcard:
			parts = append(parts, "#")
		case segment.Index != nil:
			parts = append(parts, fmt.Sprintf("%d", *segment.Index))
		default:
			parts = append(parts, escapeGJSONSegment(segment.Name))
		}
	}
	return CompiledPath{Original: path, GJSON: strings.Join(parts, "."), Segments: segments, HasWildcard: len(parts) > 0 && strings.Contains(path, "[*]")}, nil
}

func escapeGJSONSegment(name string) string {
	return strings.NewReplacer("\\", "\\\\", ".", "\\.", "#", "\\#").Replace(name)
}
