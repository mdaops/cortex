package main

import (
	"context"

	"github.com/mdaops/cortex/configurations/pkg/composer"
	"github.com/mdaops/cortex/configurations/pkg/resources"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}

func composeBase(c *composer.Composer) error {
	name := c.GetString("spec.name")
	if err := c.Err(); err != nil {
		return err
	}

	c.Add("namespace", resources.NewTenantNamespace(name))

	project, err := resources.NewTenantProject(name, c.GetString("spec.description"), c.GetStringArray("spec.sourceRepos"))
	c.ClearErrs()
	if err != nil {
		return errors.Wrap(err, "spec.sourceRepos is required")
	}
	c.Add("project", project)
	return nil
}

func composeArgoWorkflows(c *composer.Composer) {
	if !c.GetBool("spec.features.argoWorkflows.enabled") {
		c.ClearErrs()
		return
	}
	name := c.GetString("spec.name")
	c.ClearErrs()

	c.Add("argo-workflow-sa", resources.NewServiceAccount(resources.ServiceAccountConfig{
		Name:      "argo-workflow",
		Namespace: name,
	}))
	c.Add("argo-workflow-role", resources.NewRole(resources.RoleConfig{
		Name:      "argo-workflow",
		Namespace: name,
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"workflowtaskresults"},
			Verbs:     []string{"create", "patch"},
		}},
	}))
	c.Add("argo-workflow-rolebinding", resources.NewRoleBinding(resources.RoleBindingConfig{
		Name:               "argo-workflow",
		Namespace:          name,
		RoleName:           "argo-workflow",
		ServiceAccountName: "argo-workflow",
	}))
}

func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	rsp := response.To(req, response.DefaultTTL)

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composite resource"))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-name", oxr.Resource.GetName(),
		"xr-kind", oxr.Resource.GetKind(),
	)

	c := composer.New(oxr)

	if err := composeBase(c); err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	composeArgoWorkflows(c)

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	for name, obj := range c.Desired() {
		res := composed.New()
		if err := resources.ConvertViaJSON(res, obj); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
			return rsp, nil
		}
		desired[name] = &resource.DesiredComposed{Resource: res}
	}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	log.Info("Composed tenant resources", "tenant", c.GetString("spec.name"))
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}
