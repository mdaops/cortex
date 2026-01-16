package main

import (
	"context"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/mdaops/cortex/configurations/pkg/resources"
)

type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
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

	tenantName, err := oxr.Resource.GetString("spec.name")
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "spec.name is required"))
		return rsp, nil
	}

	description, _ := oxr.Resource.GetString("spec.description")
	sourceRepos, _ := oxr.Resource.GetStringArray("spec.sourceRepos")

	project, err := resources.NewTenantProject(tenantName, description, sourceRepos)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "spec.sourceRepos is required"))
		return rsp, nil
	}

	desiredTyped := map[resource.Name]any{
		"namespace": resources.NewTenantNamespace(tenantName),
		"project":   project,
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	for name, obj := range desiredTyped {
		c := composed.New()
		if err := resources.ConvertViaJSON(c, obj); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
			return rsp, nil
		}
		desired[name] = &resource.DesiredComposed{Resource: c}
	}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	log.Info("Composed tenant resources", "tenant", tenantName)
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}
