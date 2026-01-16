// Package resources provides builders for Crossplane managed resources.
package resources

import "encoding/json"

// ConvertViaJSON converts a typed struct to an unstructured target via JSON marshaling.
// This is the recommended pattern for converting typed provider resources to
// composed.Resource in Crossplane composition functions.
func ConvertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
