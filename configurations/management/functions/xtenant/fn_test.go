package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}

	cases := map[string]struct {
		reason          string
		args            args
		wantErr         bool
		wantFatal       bool
		wantResourceCnt int
		validateFn      func(t *testing.T, rsp *fnv1.RunFunctionResponse)
	}{
		"CreateTenantResources": {
			reason: "The function should create a namespace and ArgoCD project for the tenant",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.synapse.io/v1alpha1",
								"kind": "Tenant",
								"metadata": {"name": "finance"},
								"spec": {
									"name": "finance",
									"description": "Finance team workloads",
									"sourceRepos": ["https://github.com/org/finance-apps.git"]
								}
							}`),
						},
					},
				},
			},
			wantResourceCnt: 2,
			validateFn: func(t *testing.T, rsp *fnv1.RunFunctionResponse) {
				t.Helper()
				ns := rsp.GetDesired().GetResources()["namespace"]
				if ns == nil {
					t.Fatal("expected namespace resource")
				}
				nsData := structToMap(t, ns.GetResource())
				assertEqual(t, "kubernetes.crossplane.io/v1alpha2", getNestedString(nsData, "apiVersion"))
				assertEqual(t, "Object", getNestedString(nsData, "kind"))
				assertEqual(t, "finance-namespace", getNestedString(nsData, "metadata", "name"))
				assertEqual(t, "finance", getNestedString(nsData, "spec", "forProvider", "manifest", "metadata", "name"))
				assertEqual(t, "finance", getNestedString(nsData, "spec", "forProvider", "manifest", "metadata", "labels", "platform.synapse.io/tenant"))

				proj := rsp.GetDesired().GetResources()["project"]
				if proj == nil {
					t.Fatal("expected project resource")
				}
				projData := structToMap(t, proj.GetResource())
				assertEqual(t, "projects.argocd.crossplane.io/v1alpha1", getNestedString(projData, "apiVersion"))
				assertEqual(t, "Project", getNestedString(projData, "kind"))
				assertEqual(t, "finance-project", getNestedString(projData, "metadata", "name"))
				assertEqual(t, "Finance team workloads", getNestedString(projData, "spec", "forProvider", "description"))

				dests := getNestedSlice(projData, "spec", "forProvider", "destinations")
				if len(dests) != 2 {
					t.Fatalf("expected 2 destinations, got %d", len(dests))
				}
			},
		},
		"DefaultDescription": {
			reason: "The function should use default description when not provided",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.synapse.io/v1alpha1",
								"kind": "Tenant",
								"metadata": {"name": "minimal"},
								"spec": {
									"name": "minimal",
									"sourceRepos": ["https://github.com/org/repo.git"]
								}
							}`),
						},
					},
				},
			},
			wantResourceCnt: 2,
			validateFn: func(t *testing.T, rsp *fnv1.RunFunctionResponse) {
				t.Helper()
				proj := rsp.GetDesired().GetResources()["project"]
				projData := structToMap(t, proj.GetResource())
				assertEqual(t, "minimal tenant workloads", getNestedString(projData, "spec", "forProvider", "description"))
			},
		},
		"MissingSourceRepos": {
			reason:    "The function should return fatal when spec.sourceRepos is missing",
			wantFatal: true,
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.synapse.io/v1alpha1",
								"kind": "Tenant",
								"metadata": {"name": "no-repos"},
								"spec": {"name": "no-repos"}
							}`),
						},
					},
				},
			},
		},
		"MissingTenantName": {
			reason:    "The function should return fatal when spec.name is missing",
			wantFatal: true,
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.synapse.io/v1alpha1",
								"kind": "Tenant",
								"metadata": {"name": "invalid"},
								"spec": {}
							}`),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.wantErr, err != nil, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}

			if tc.wantFatal {
				if len(rsp.GetResults()) == 0 {
					t.Fatal("expected fatal result")
				}
				if rsp.GetResults()[0].GetSeverity() != fnv1.Severity_SEVERITY_FATAL {
					t.Errorf("expected SEVERITY_FATAL, got %v", rsp.GetResults()[0].GetSeverity())
				}
				return
			}

			if tc.wantResourceCnt > 0 {
				gotCnt := len(rsp.GetDesired().GetResources())
				if gotCnt != tc.wantResourceCnt {
					t.Errorf("expected %d resources, got %d", tc.wantResourceCnt, gotCnt)
				}
			}

			if tc.validateFn != nil {
				tc.validateFn(t, rsp)
			}

			if len(rsp.GetConditions()) == 0 {
				t.Error("expected at least one condition")
			}
		})
	}
}

func structToMap(t *testing.T, s *structpb.Struct) map[string]any {
	t.Helper()
	bs, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal struct: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(bs, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}
	return m
}

func getNestedString(m map[string]any, keys ...string) string {
	val := getNestedValue(m, keys...)
	if val == nil {
		return ""
	}
	s, _ := val.(string)
	return s
}

func getNestedSlice(m map[string]any, keys ...string) []any {
	val := getNestedValue(m, keys...)
	if val == nil {
		return nil
	}
	s, _ := val.([]any)
	return s
}

func getNestedValue(m map[string]any, keys ...string) any {
	if len(keys) == 0 {
		return m
	}
	val, ok := m[keys[0]]
	if !ok {
		return nil
	}
	if len(keys) == 1 {
		return val
	}
	nested, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	return getNestedValue(nested, keys[1:]...)
}

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}
