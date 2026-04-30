/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourceclaimtemplate

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/kubernetes/fake"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/apis/resource"
)

var apiVersions = []string{"v1beta1", "v1beta2", "v1"}

func TestDeclarativeValidate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidate(t, apiVersion)
		})
	}
}

func testDeclarativeValidate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceclaimtemplates",
	})
	fakeClient := fake.NewClientset()
	nsClient := fakeClient.CoreV1().Namespaces()
	Strategy := NewStrategy(nsClient)

	testCases := map[string]struct {
		input        resource.ResourceClaimTemplate
		expectedErrs field.ErrorList
	}{
		"valid": {
			input: mkValidResourceClaimTemplate(),
		},
		"invalid requests, too many": {
			input: mkValidResourceClaimTemplate(tweakDevicesRequests(33)),
			expectedErrs: field.ErrorList{
				field.TooMany(field.NewPath("spec", "spec", "devices", "requests"), 33, 32).WithOrigin("maxItems").MarkAlpha(),
			},
		},
		"invalid requests, duplicate name": {
			input: mkValidResourceClaimTemplate(tweakAddDeviceRequest(mkDeviceRequest("req-0"))),
			expectedErrs: field.ErrorList{
				field.Duplicate(field.NewPath("spec", "spec", "devices", "requests").Index(1), "req-0").MarkAlpha(),
			},
		},
		// TODO: Add more test cases
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			apitesting.VerifyValidationEquivalence(t, ctx, &tc.input, Strategy, tc.expectedErrs)
		})
	}
}

func TestDeclarativeValidateUpdate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidateUpdate(t, apiVersion)
		})
	}
}

func testDeclarativeValidateUpdate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceclaimtemplates",
	})
	fakeClient := fake.NewClientset()
	nsClient := fakeClient.CoreV1().Namespaces()
	Strategy := NewStrategy(nsClient)

	testCases := map[string]struct {
		old          resource.ResourceClaimTemplate
		update       resource.ResourceClaimTemplate
		expectedErrs field.ErrorList
	}{
		"valid update (no spec change)": {
			old:    mkValidResourceClaimTemplate(),
			update: mkValidResourceClaimTemplate(),
		},
		// TODO: Add more test cases
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.old.ResourceVersion = "1"
			tc.update.ResourceVersion = "2"
			apitesting.VerifyUpdateValidationEquivalence(t, ctx, &tc.update, &tc.old, Strategy, tc.expectedErrs)
		})
	}
}

func mkValidResourceClaimTemplate(tweaks ...func(rct *resource.ResourceClaimTemplate)) resource.ResourceClaimTemplate {
	rct := resource.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valid-claim-template",
			Namespace: "default",
		},
		Spec: resource.ResourceClaimTemplateSpec{
			Spec: resource.ResourceClaimSpec{
				Devices: resource.DeviceClaim{
					Requests: []resource.DeviceRequest{
						mkDeviceRequest("req-0"),
					},
				},
			},
		},
	}
	for _, tweak := range tweaks {
		tweak(&rct)
	}
	return rct
}

func mkDeviceRequest(name string) resource.DeviceRequest {
	return resource.DeviceRequest{
		Name: name,
		Exactly: &resource.ExactDeviceRequest{
			DeviceClassName: "class",
			AllocationMode:  resource.DeviceAllocationModeExactCount,
			Count:           1,
		},
	}
}

func tweakDevicesRequests(items int) func(*resource.ResourceClaimTemplate) {
	return func(rct *resource.ResourceClaimTemplate) {
		// The first request already exists in the valid template
		for i := 1; i < items; i++ {
			rct.Spec.Spec.Devices.Requests = append(rct.Spec.Spec.Devices.Requests, mkDeviceRequest(fmt.Sprintf("req-%d", i)))
		}
	}
}

func tweakAddDeviceRequest(req resource.DeviceRequest) func(*resource.ResourceClaimTemplate) {
	return func(rct *resource.ResourceClaimTemplate) {
		rct.Spec.Spec.Devices.Requests = append(rct.Spec.Spec.Devices.Requests, req)
	}
}
