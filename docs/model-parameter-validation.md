# Configurable model parameter validation

The `service/modelparam` package validates an inbound JSON request before token estimation and billing. Rules match the client model and protocol, then apply a restricted JSONPath constraint. Request JSON is queried with `gjson`; the validator is read-only. JSON mutation, if added later as a separate normalization phase, must use `sjson`.

## Rule shape

```json
{
  "version": 1,
  "rules": [
    {
      "id": 1001,
      "enabled": true,
      "priority": 100,
      "model": {"match": "glob", "pattern": "gpt-5*"},
      "protocols": ["openai_chat"],
      "path": "$.temperature",
      "constraint": {"operator": "range", "min": "0", "max": "1"},
      "error": {
        "code": "model_parameter_out_of_range",
        "message": "temperature must be between 0 and 1"
      }
    }
  ]
}
```

Supported model selectors are `exact`, `prefix`, and trailing-star `glob`. Supported protocols are `openai_chat`, `openai_responses`, `claude_messages`, and `gemini_generate`.

Supported JSONPath syntax is `$`, object fields, numeric array indexes, quoted bracket fields, and one or more `[*]` array wildcards. Recursive descent, filters, slices, modifiers, and scripts are rejected when the configuration is loaded.

Supported operators are `required`, `forbidden`, `not_null`, `type`, `equals`, `not_equals`, `in`, `not_in`, `range`, `min_length`, `max_length`, `min_items`, `max_items`, and `matches`.

Invalid client JSON and rule violations are returned as non-retryable HTTP 400 errors. Invalid rule configurations are rejected before publication, leaving the previous registry snapshot active.

The persisted system option key is `ModelParameterRules`. It can be updated through the existing admin option endpoint; validation occurs before persistence, and option-map synchronization reloads the compiled snapshot automatically.
