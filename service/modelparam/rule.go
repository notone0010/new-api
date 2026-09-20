package modelparam

type Protocol string

const (
	ProtocolOpenAIChat      Protocol = "openai_chat"
	ProtocolOpenAIResponses Protocol = "openai_responses"
	ProtocolOpenAIImage     Protocol = "openai_image"
	ProtocolOpenAIAudio     Protocol = "openai_audio"
	ProtocolOpenAIEmbedding Protocol = "openai_embedding"
	ProtocolClaudeMessages  Protocol = "claude_messages"
	ProtocolGeminiGenerate  Protocol = "gemini_generate"
	ProtocolRerank          Protocol = "rerank"
)

type MatchType string

const (
	MatchExact  MatchType = "exact"
	MatchPrefix MatchType = "prefix"
	MatchGlob   MatchType = "glob"
)

type Operator string

const (
	OperatorRequired  Operator = "required"
	OperatorForbidden Operator = "forbidden"
	OperatorNotNull   Operator = "not_null"
	OperatorType      Operator = "type"
	OperatorEquals    Operator = "equals"
	OperatorNotEquals Operator = "not_equals"
	OperatorIn        Operator = "in"
	OperatorNotIn     Operator = "not_in"
	OperatorRange     Operator = "range"
	OperatorMinLength Operator = "min_length"
	OperatorMaxLength Operator = "max_length"
	OperatorMinItems  Operator = "min_items"
	OperatorMaxItems  Operator = "max_items"
	OperatorMatches   Operator = "matches"
)

type Quantifier string

const (
	QuantifierAll Quantifier = "all"
	QuantifierAny Quantifier = "any"
)

type ModelSelector struct {
	Match   MatchType `json:"match"`
	Pattern string    `json:"pattern"`
}

type Constraint struct {
	Operator     Operator `json:"operator"`
	Value        any      `json:"value,omitempty"`
	Values       []any    `json:"values,omitempty"`
	Min          string   `json:"min,omitempty"`
	Max          string   `json:"max,omitempty"`
	MinInclusive *bool    `json:"min_inclusive,omitempty"`
	MaxInclusive *bool    `json:"max_inclusive,omitempty"`
}

type Condition struct {
	Path     string   `json:"path"`
	Operator Operator `json:"operator"`
	Value    any      `json:"value,omitempty"`
	Values   []any    `json:"values,omitempty"`
}

type ErrorDefinition struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Rule struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Enabled    bool            `json:"enabled"`
	Priority   int             `json:"priority"`
	Model      ModelSelector   `json:"model"`
	Protocols  []Protocol      `json:"protocols"`
	Path       string          `json:"path"`
	Quantifier Quantifier      `json:"quantifier,omitempty"`
	When       []Condition     `json:"when,omitempty"`
	Constraint Constraint      `json:"constraint"`
	Error      ErrorDefinition `json:"error"`
}

type Input struct {
	Model    string
	Protocol Protocol
	Body     []byte
}

type Violation struct {
	RuleID    int64
	Path      string
	Parameter string
	Code      string
	Message   string
}
