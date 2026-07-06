package v5

import "encoding/json"

// The generated *OperationFilterInput types (generated.go) lack `omitempty` on their
// fields. This is a genqlient limitation with shared struct types on a large schema
// (https://github.com/Khan/genqlient/issues/149, #260): a per-operation
// `# @genqlient(omitempty: true)` directive can be silently overridden when the same
// generated struct is also referenced by another operation that lacks the directive.
//
// Left as-is, every unset comparator (eq, neq, gt, ...) serializes as an explicit
// `null` instead of being omitted. HotChocolate treats an explicitly-provided `null`
// as its own predicate - e.g. `eq: null` becomes "field IS NULL" - so combined with a
// real value on a sibling field (e.g. `gte: "2026-06-18"`), the resulting filter is
// self-contradictory ("IS NULL AND >= X") and never matches anything.
//
// These overrides restore the intended "unset field is omitted" behavior for the
// filter types our own code (asset.go, device_queries.go) actually populates
// partially. They live outside generated.go so they survive regeneration.

func (v ArchitectureOperationFilterInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Eq  *Architecture  `json:"eq,omitempty"`
		Neq *Architecture  `json:"neq,omitempty"`
		In  []Architecture `json:"in,omitempty"`
		Nin []Architecture `json:"nin,omitempty"`
	}{v.Eq, v.Neq, v.In, v.Nin})
}

func (v ComparableGuidOperationFilterInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Eq   *string  `json:"eq,omitempty"`
		Neq  *string  `json:"neq,omitempty"`
		In   []string `json:"in,omitempty"`
		Nin  []string `json:"nin,omitempty"`
		Gt   *string  `json:"gt,omitempty"`
		Ngt  *string  `json:"ngt,omitempty"`
		Gte  *string  `json:"gte,omitempty"`
		Ngte *string  `json:"ngte,omitempty"`
		Lt   *string  `json:"lt,omitempty"`
		Nlt  *string  `json:"nlt,omitempty"`
		Lte  *string  `json:"lte,omitempty"`
		Nlte *string  `json:"nlte,omitempty"`
	}{v.Eq, v.Neq, v.In, v.Nin, v.Gt, v.Ngt, v.Gte, v.Ngte, v.Lt, v.Nlt, v.Lte, v.Nlte})
}

func (v ComparableNullableOfDateTimeOperationFilterInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Eq   *string   `json:"eq,omitempty"`
		Neq  *string   `json:"neq,omitempty"`
		In   []*string `json:"in,omitempty"`
		Nin  []*string `json:"nin,omitempty"`
		Gt   *string   `json:"gt,omitempty"`
		Ngt  *string   `json:"ngt,omitempty"`
		Gte  *string   `json:"gte,omitempty"`
		Ngte *string   `json:"ngte,omitempty"`
		Lt   *string   `json:"lt,omitempty"`
		Nlt  *string   `json:"nlt,omitempty"`
		Lte  *string   `json:"lte,omitempty"`
		Nlte *string   `json:"nlte,omitempty"`
	}{v.Eq, v.Neq, v.In, v.Nin, v.Gt, v.Ngt, v.Gte, v.Ngte, v.Lt, v.Nlt, v.Lte, v.Nlte})
}

func (v StringOperationFilterInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		And         []*StringOperationFilterInput `json:"and,omitempty"`
		Or          []*StringOperationFilterInput `json:"or,omitempty"`
		Eq          *string                       `json:"eq,omitempty"`
		Neq         *string                       `json:"neq,omitempty"`
		Contains    *string                       `json:"contains,omitempty"`
		Ncontains   *string                       `json:"ncontains,omitempty"`
		In          []*string                     `json:"in,omitempty"`
		Nin         []*string                     `json:"nin,omitempty"`
		StartsWith  *string                       `json:"startsWith,omitempty"`
		NstartsWith *string                       `json:"nstartsWith,omitempty"`
		EndsWith    *string                       `json:"endsWith,omitempty"`
		NendsWith   *string                       `json:"nendsWith,omitempty"`
	}{v.And, v.Or, v.Eq, v.Neq, v.Contains, v.Ncontains, v.In, v.Nin, v.StartsWith, v.NstartsWith, v.EndsWith, v.NendsWith})
}
