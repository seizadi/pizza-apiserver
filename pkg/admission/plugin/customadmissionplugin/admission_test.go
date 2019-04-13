/*
Copyright 2017 The Kubernetes Authors.

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

package customadmissionplugin_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	clienttesting "k8s.io/client-go/testing"
	"github.com/programming-kubernetes/custom-apiserver/pkg/admission/plugin/customadmissionplugin"
	"github.com/programming-kubernetes/custom-apiserver/pkg/admission/custominitializer"
	"github.com/programming-kubernetes/custom-apiserver/pkg/apis/custom"
	"github.com/programming-kubernetes/custom-apiserver/pkg/generated/clientset/internalversion/fake"
	informers "github.com/programming-kubernetes/custom-apiserver/pkg/generated/informers/internalversion"
)

// TestBanfluderAdmissionPlugin tests various test cases against
// ban flunder admission plugin
func TestBanflunderAdmissionPlugin(t *testing.T) {
	var scenarios = []struct {
		informersOutput        custom.PolicyList
		admissionInput         custom.Flunder
		admissionInputKind     schema.GroupVersionKind
		admissionInputResource schema.GroupVersionResource
		admissionMustFail      bool
	}{
		// scenario 1:
		// a flunder with a name that appears on a list of disallowed flunders must be banned
		{
			informersOutput: custom.PolicyList{
				Items: []custom.Policy{
					{DisallowedFlunders: []string{"badname"}},
				},
			},
			admissionInput: custom.Flunder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "badname",
					Namespace: "",
				},
			},
			admissionInputKind:     custom.Kind("Flunder").WithVersion("version"),
			admissionInputResource: custom.Resource("flunders").WithVersion("version"),
			admissionMustFail:      true,
		},
		// scenario 2:
		// a flunder with a name that does not appear on a list of disallowed flunders must be admitted
		{
			informersOutput: custom.PolicyList{
				Items: []custom.Policy{
					{DisallowedFlunders: []string{"badname"}},
				},
			},
			admissionInput: custom.Flunder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "goodname",
					Namespace: "",
				},
			},
			admissionInputKind:     custom.Kind("Flunder").WithVersion("version"),
			admissionInputResource: custom.Resource("flunders").WithVersion("version"),
			admissionMustFail:      false,
		},
		// scenario 3:
		// a flunder with a name that appears on a list of disallowed flunders would be banned
		// but the kind passed in is not a flunder thus the whole request is accepted
		{
			informersOutput: custom.PolicyList{
				Items: []custom.Policy{
					{DisallowedFlunders: []string{"badname"}},
				},
			},
			admissionInput: custom.Flunder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "badname",
					Namespace: "",
				},
			},
			admissionInputKind:     custom.Kind("NotFlunder").WithVersion("version"),
			admissionInputResource: custom.Resource("notflunders").WithVersion("version"),
			admissionMustFail:      false,
		},
	}

	for index, scenario := range scenarios {
		func() {
			// prepare
			cs := &fake.Clientset{}
			cs.AddReactor("list", "policies", func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, &scenario.informersOutput, nil
			})
			informersFactory := informers.NewSharedInformerFactory(cs, 5*time.Minute)

			target, err := customadmissionplugin.New()
			if err != nil {
				t.Fatalf("scenario %d: failed to create customadmissionplugin due to = %v", index, err)
			}

			targetInitializer := custominitializer.New(informersFactory)
			targetInitializer.Initialize(target)

			err = admission.ValidateInitialization(target)
			if err != nil {
				t.Fatalf("scenario %d: failed to initialize customadmissionplugin due to =%v", index, err)
			}

			stop := make(chan struct{})
			defer close(stop)
			informersFactory.Start(stop)
			informersFactory.WaitForCacheSync(stop)

			// act
			err = target.Admit(admission.NewAttributesRecord(
				&scenario.admissionInput,
				nil,
				scenario.admissionInputKind,
				scenario.admissionInput.ObjectMeta.Namespace,
				"",
				scenario.admissionInputResource,
				"",
				admission.Create,
				false,
				nil),
			)

			// validate
			if scenario.admissionMustFail && err == nil {
				t.Errorf("scenario %d: expected an error but got nothing", index)
			}

			if !scenario.admissionMustFail && err != nil {
				t.Errorf("scenario %d: customadmissionplugin returned unexpected error = %v", index, err)
			}
		}()
	}
}
