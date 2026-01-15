package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"CreateTenantResources": {
			reason: "The function should create a namespace and ArgoCD project for the tenant",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.synapse.io/v1alpha1",
								"kind": "XTenant",
								"metadata": {
									"name": "finance"
								},
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
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"namespace": {Resource: resource.MustStructJSON(`{
								"apiVersion": "kubernetes.crossplane.io/v1alpha2",
								"kind": "Object",
								"metadata": {
									"name": "finance-namespace"
								},
								"spec": {
									"forProvider": {
										"manifest": {
											"apiVersion": "v1",
											"kind": "Namespace",
											"metadata": {
												"name": "finance",
												"labels": {
													"platform.synapse.io/tenant": "finance"
												}
											}
										}
									},
									"providerConfigRef": {
										"name": "default"
									}
								}
							}`)},
							"project": {Resource: resource.MustStructJSON(`{
								"apiVersion": "projects.argocd.crossplane.io/v1alpha1",
								"kind": "Project",
								"metadata": {
									"name": "finance-project"
								},
								"spec": {
									"forProvider": {
										"metadata": {
											"name": "finance",
											"namespace": "argo-system"
										},
										"description": "Finance team workloads",
										"sourceRepos": ["https://github.com/org/finance-apps.git"],
										"destinations": [
											{
												"namespace": "finance",
												"server": "https://kubernetes.default.svc"
											},
											{
												"namespace": "finance-*",
												"server": "https://kubernetes.default.svc"
											}
										]
									},
									"providerConfigRef": {
										"name": "default"
									}
								}
							}`)},
						},
					},
					Conditions: []*fnv1.Condition{
						{
							Type:   "FunctionSuccess",
							Status: fnv1.Status_STATUS_CONDITION_TRUE,
							Reason: "Success",
							Target: fnv1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
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

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform(), protocmp.IgnoreFields(&fnv1.RunFunctionResponse{}, "meta")); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
