package modelparam

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type compiledRule struct {
	rule       Rule
	path       CompiledPath
	conditions []compiledCondition
	min, max   *decimal.Decimal
	pattern    *regexp.Regexp
}

type compiledCondition struct {
	path CompiledPath
	rule Condition
}

type Snapshot struct{ rules []compiledRule }

func ValidateDefault(input Input) (*Violation, error) {
	registry := DefaultRegistry()
	if registry == nil {
		return nil, nil
	}
	snapshot := registry.Snapshot()
	if snapshot == nil {
		return nil, nil
	}
	return snapshot.Validate(input)
}

func Compile(rules []Rule) (*Snapshot, error) {
	compiled := make([]compiledRule, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if rule.Model.Pattern == "" {
			return nil, fmt.Errorf("rule %d has empty model pattern", rule.ID)
		}
		if rule.Model.Match == "" {
			rule.Model.Match = MatchExact
		}
		if rule.Model.Match != MatchExact && rule.Model.Match != MatchPrefix && rule.Model.Match != MatchGlob {
			return nil, fmt.Errorf("rule %d has invalid model match", rule.ID)
		}
		if rule.Quantifier != "" && rule.Quantifier != QuantifierAll && rule.Quantifier != QuantifierAny {
			return nil, fmt.Errorf("rule %d has invalid quantifier", rule.ID)
		}
		for _, protocol := range rule.Protocols {
			switch protocol {
			case ProtocolOpenAIChat, ProtocolOpenAIResponses, ProtocolOpenAIImage, ProtocolOpenAIAudio, ProtocolOpenAIEmbedding, ProtocolClaudeMessages, ProtocolGeminiGenerate, ProtocolRerank:
			default:
				return nil, fmt.Errorf("rule %d has invalid protocol %q", rule.ID, protocol)
			}
		}
		path, err := CompilePath(rule.Path)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", rule.ID, err)
		}
		if err := validateConstraint(rule.Constraint); err != nil {
			return nil, fmt.Errorf("rule %d: %w", rule.ID, err)
		}
		cr := compiledRule{rule: rule, path: path}
		if rule.Constraint.Operator == OperatorRange {
			if rule.Constraint.Min != "" {
				v, err := decimal.NewFromString(rule.Constraint.Min)
				if err != nil {
					return nil, fmt.Errorf("rule %d invalid min", rule.ID)
				}
				cr.min = &v
			}
			if rule.Constraint.Max != "" {
				v, err := decimal.NewFromString(rule.Constraint.Max)
				if err != nil {
					return nil, fmt.Errorf("rule %d invalid max", rule.ID)
				}
				cr.max = &v
			}
			if cr.min != nil && cr.max != nil && cr.min.GreaterThan(*cr.max) {
				return nil, fmt.Errorf("rule %d range min exceeds max", rule.ID)
			}
		}
		if rule.Constraint.Operator == OperatorMatches {
			value, ok := rule.Constraint.Value.(string)
			if !ok || value == "" {
				return nil, fmt.Errorf("rule %d invalid regex", rule.ID)
			}
			cr.pattern, err = regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("rule %d invalid regex: %w", rule.ID, err)
			}
		}
		for _, condition := range rule.When {
			p, err := CompilePath(condition.Path)
			if err != nil {
				return nil, fmt.Errorf("rule %d condition: %w", rule.ID, err)
			}
			if condition.Operator != OperatorRequired && condition.Operator != OperatorForbidden && condition.Operator != OperatorEquals && condition.Operator != OperatorNotEquals && condition.Operator != OperatorIn && condition.Operator != OperatorNotIn {
				return nil, fmt.Errorf("rule %d has invalid condition operator", rule.ID)
			}
			cr.conditions = append(cr.conditions, compiledCondition{path: p, rule: condition})
		}
		compiled = append(compiled, cr)
	}
	sort.SliceStable(compiled, func(i, j int) bool {
		if compiled[i].rule.Priority != compiled[j].rule.Priority {
			return compiled[i].rule.Priority > compiled[j].rule.Priority
		}
		return compiled[i].rule.ID < compiled[j].rule.ID
	})
	return &Snapshot{rules: compiled}, nil
}

func (s *Snapshot) Validate(input Input) (*Violation, error) {
	if !gjson.ValidBytes(input.Body) {
		return nil, ErrInvalidRequestJSON
	}
	for _, rule := range s.rules {
		if !modelMatches(rule.rule.Model, input.Model) || !protocolMatches(rule.rule.Protocols, input.Protocol) {
			continue
		}
		if !conditionsMatch(input.Body, rule.conditions) {
			continue
		}
		results := resolve(input.Body, rule.path)
		if ok, parameter := check(rule, results); !ok {
			return violation(rule.rule, parameter), nil
		}
	}
	return nil, nil
}

type resolved struct {
	result    gjson.Result
	parameter string
	exists    bool
}

func resolve(body []byte, path CompiledPath) []resolved {
	if !path.HasWildcard {
		r := gjson.GetBytes(body, path.GJSON)
		if !r.Exists() {
			return nil
		}
		return []resolved{{result: r, parameter: path.Original[2:], exists: true}}
	}
	return resolveSegments(gjson.ParseBytes(body), path.Segments, "")
}

func resolveSegments(node gjson.Result, segments []PathSegment, prefix string) []resolved {
	if len(segments) == 0 {
		return []resolved{{result: node, parameter: prefix, exists: true}}
	}
	segment := segments[0]
	rest := segments[1:]
	if segment.Wildcard {
		if node.Type != gjson.JSON || !strings.HasPrefix(strings.TrimSpace(node.Raw), "[") {
			return nil
		}
		var out []resolved
		node.ForEach(func(key, value gjson.Result) bool {
			p := fmt.Sprintf("%s[%s]", prefix, key.String())
			out = append(out, resolveSegments(value, rest, p)...)
			return true
		})
		return out
	}
	var child gjson.Result
	var p string
	if segment.Index != nil {
		child = node.Get(fmt.Sprintf("%d", *segment.Index))
		p = fmt.Sprintf("%s[%d]", prefix, *segment.Index)
	} else {
		child = node.Get(segment.Name)
		p = prefix + "." + segment.Name
	}
	if !child.Exists() {
		return []resolved{{parameter: renderMissingPath(p, rest), exists: false}}
	}
	p = strings.TrimPrefix(p, ".")
	return resolveSegments(child, rest, p)
}

func renderMissingPath(prefix string, segments []PathSegment) string {
	for _, segment := range segments {
		switch {
		case segment.Index != nil:
			prefix = fmt.Sprintf("%s[%d]", prefix, *segment.Index)
		case segment.Wildcard:
			prefix += "[*]"
		default:
			prefix += "." + segment.Name
		}
	}
	return strings.TrimPrefix(prefix, ".")
}

func modelMatches(selector ModelSelector, model string) bool {
	switch selector.Match {
	case MatchExact:
		return model == selector.Pattern
	case MatchPrefix:
		return strings.HasPrefix(model, selector.Pattern)
	case MatchGlob:
		if strings.HasSuffix(selector.Pattern, "*") {
			return strings.HasPrefix(model, strings.TrimSuffix(selector.Pattern, "*"))
		}
		return model == selector.Pattern
	}
	return false
}
func protocolMatches(protocols []Protocol, protocol Protocol) bool {
	if len(protocols) == 0 {
		return true
	}
	for _, p := range protocols {
		if p == protocol {
			return true
		}
	}
	return false
}
func conditionsMatch(body []byte, conditions []compiledCondition) bool {
	for _, c := range conditions {
		values := resolve(body, c.path)
		switch c.rule.Operator {
		case OperatorRequired:
			if len(values) == 0 {
				return false
			}
		case OperatorForbidden:
			if len(values) > 0 {
				return false
			}
		case OperatorEquals, OperatorNotEquals, OperatorIn, OperatorNotIn:
			matched := false
			for _, value := range values {
				candidates := c.rule.Values
				if c.rule.Operator == OperatorEquals || c.rule.Operator == OperatorNotEquals {
					candidates = []any{c.rule.Value}
				}
				for _, candidate := range candidates {
					if sameJSONValue(value.result, candidate) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if (c.rule.Operator == OperatorEquals || c.rule.Operator == OperatorIn) != matched {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func check(rule compiledRule, values []resolved) (bool, string) {
	op := rule.rule.Constraint.Operator
	if op == OperatorRequired {
		if len(values) == 0 {
			return false, rule.path.Original[2:]
		}
		for _, value := range values {
			if !value.exists {
				return false, value.parameter
			}
		}
		return true, ""
	}
	if op == OperatorForbidden {
		for _, value := range values {
			if value.exists {
				return false, value.parameter
			}
		}
		return true, ""
	}
	if len(values) == 0 {
		return true, ""
	}
	for _, value := range values {
		if !value.exists {
			if rule.rule.Quantifier == QuantifierAny {
				continue
			}
			return false, value.parameter
		}
		valid := checkValue(rule, value.result)
		if rule.rule.Quantifier == QuantifierAny && valid {
			return true, ""
		}
		if rule.rule.Quantifier != QuantifierAny && !valid {
			return false, value.parameter
		}
	}
	if rule.rule.Quantifier == QuantifierAny {
		return false, rule.path.Original[2:]
	}
	return true, ""
}

func checkValue(rule compiledRule, value gjson.Result) bool {
	op := rule.rule.Constraint.Operator
	if op == OperatorNotNull {
		return value.Type != gjson.Null
	}
	if op == OperatorType {
		want, _ := rule.rule.Constraint.Value.(string)
		return typeName(value) == want
	}
	if op == OperatorRange {
		if value.Type != gjson.Number {
			return false
		}
		n, err := decimal.NewFromString(value.Raw)
		if err != nil {
			return false
		}
		if rule.min != nil && n.LessThan(*rule.min) {
			return false
		}
		if rule.min != nil && rule.rule.Constraint.MinInclusive != nil && !*rule.rule.Constraint.MinInclusive && n.Equal(*rule.min) {
			return false
		}
		if rule.max != nil && n.GreaterThan(*rule.max) {
			return false
		}
		if rule.max != nil && rule.rule.Constraint.MaxInclusive != nil && !*rule.rule.Constraint.MaxInclusive && n.Equal(*rule.max) {
			return false
		}
		return true
	}
	if op == OperatorEquals || op == OperatorNotEquals || op == OperatorIn || op == OperatorNotIn {
		equal := false
		candidates := rule.rule.Constraint.Values
		if op == OperatorEquals || op == OperatorNotEquals {
			candidates = []any{rule.rule.Constraint.Value}
		}
		for _, candidate := range candidates {
			if sameJSONValue(value, candidate) {
				equal = true
				break
			}
		}
		if op == OperatorEquals || op == OperatorIn {
			return equal
		}
		return !equal
	}
	if op == OperatorMinLength {
		return value.Type == gjson.String && len([]rune(value.Str)) >= intValue(rule.rule.Constraint.Value)
	}
	if op == OperatorMaxLength {
		return value.Type == gjson.String && len([]rune(value.Str)) <= intValue(rule.rule.Constraint.Value)
	}
	if op == OperatorMinItems || op == OperatorMaxItems {
		if value.Type != gjson.JSON || !strings.HasPrefix(strings.TrimSpace(value.Raw), "[") {
			return false
		}
		want := intValue(rule.rule.Constraint.Value)
		count := len(value.Array())
		if op == OperatorMinItems {
			return count >= want
		}
		return count <= want
	}
	if op == OperatorMatches {
		return value.Type == gjson.String && rule.pattern.MatchString(value.Str)
	}
	return true
}
func typeName(v gjson.Result) string {
	switch v.Type {
	case gjson.String:
		return "string"
	case gjson.Number:
		if strings.ContainsAny(v.Raw, ".eE") {
			return "number"
		}
		return "integer"
	case gjson.True, gjson.False:
		return "boolean"
	case gjson.Null:
		return "null"
	case gjson.JSON:
		if strings.HasPrefix(strings.TrimSpace(v.Raw), "[") {
			return "array"
		}
		return "object"
	}
	return ""
}
func intValue(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
func isIntegerValue(v any) bool {
	switch n := v.(type) {
	case int:
		return true
	case float64:
		return n == float64(int(n))
	default:
		return false
	}
}
func violation(rule Rule, parameter string) *Violation {
	code := rule.Error.Code
	if code == "" {
		code = "model_parameter_invalid"
	}
	message := rule.Error.Message
	if message == "" {
		message = "model parameter is invalid"
	}
	return &Violation{RuleID: rule.ID, Path: rule.Path, Parameter: parameter, Code: code, Message: message}
}

func validateConstraint(c Constraint) error {
	switch c.Operator {
	case OperatorRequired, OperatorForbidden, OperatorNotNull:
	case OperatorType:
		value, ok := c.Value.(string)
		allowed := map[string]bool{"string": true, "number": true, "integer": true, "boolean": true, "object": true, "array": true, "null": true}
		if !ok || !allowed[value] {
			return fmt.Errorf("type constraint requires a string value")
		}
	case OperatorEquals, OperatorNotEquals:
		if c.Value == nil {
			return fmt.Errorf("equality constraint requires a value")
		}
	case OperatorIn, OperatorNotIn:
		if len(c.Values) == 0 {
			return fmt.Errorf("set constraint requires values")
		}
	case OperatorRange:
		if c.Min == "" && c.Max == "" {
			return fmt.Errorf("range constraint requires min or max")
		}
	case OperatorMinLength, OperatorMaxLength, OperatorMinItems, OperatorMaxItems:
		if c.Value == nil || intValue(c.Value) < 0 || !isIntegerValue(c.Value) {
			return fmt.Errorf("size constraint requires a value")
		}
	case OperatorMatches:
		if _, ok := c.Value.(string); !ok {
			return fmt.Errorf("matches constraint requires a string value")
		}
	default:
		return fmt.Errorf("unsupported operator %q", c.Operator)
	}
	return nil
}

func sameJSONValue(result gjson.Result, value any) bool {
	switch v := value.(type) {
	case string:
		return result.Type == gjson.String && result.Str == v
	case bool:
		return (v && result.Type == gjson.True) || (!v && result.Type == gjson.False)
	case float64:
		if result.Type != gjson.Number {
			return false
		}
		actual, err := decimal.NewFromString(result.Raw)
		if err != nil {
			return false
		}
		expected, err := decimal.NewFromString(strconv.FormatFloat(v, 'g', -1, 64))
		return err == nil && actual.Equal(expected)
	case int:
		if result.Type != gjson.Number {
			return false
		}
		actual, err := decimal.NewFromString(result.Raw)
		expected := decimal.NewFromInt(int64(v))
		return err == nil && actual.Equal(expected)
	default:
		return false
	}
}
