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

package custominitializer_test

import (
	"testing"
	"time"

	"k8s.io/apiserver/pkg/admission"
	"github.com/programming-kubernetes/custom-apiserver/pkg/admission/custominitializer"
	"github.com/programming-kubernetes/custom-apiserver/pkg/generated/clientset/internalversion/fake"
	informers "github.com/programming-kubernetes/custom-apiserver/pkg/generated/informers/internalversion"
)

// TestWantsInternalWardleInformerFactory ensures that the informer factory is injected
// when the WantsInternalWardleInformerFactory interface is implemented by a plugin.
func TestWantsInternalWardleInformerFactory(t *testing.T) {
	cs := &fake.Clientset{}
	sf := informers.NewSharedInformerFactory(cs, time.Duration(1)*time.Second)
	target := custominitializer.New(sf)

	wantWardleInformerFactory := &wantInternalWardleInformerFactory{}
	target.Initialize(wantWardleInformerFactory)
	if wantWardleInformerFactory.sf != sf {
		t.Errorf("expected informer factory to be initialized")
	}
}

// wantInternalWardleInformerFactory is a test stub that fulfills the WantsInternalWardleInformerFactory interface
type wantInternalWardleInformerFactory struct {
	sf informers.SharedInformerFactory
}

func (self *wantInternalWardleInformerFactory) SetInternalWardleInformerFactory(sf informers.SharedInformerFactory) {
	self.sf = sf
}
func (self *wantInternalWardleInformerFactory) Admit(a admission.Attributes) error { return nil }
func (self *wantInternalWardleInformerFactory) Handles(o admission.Operation) bool { return false }
func (self *wantInternalWardleInformerFactory) ValidateInitialization() error      { return nil }

var _ admission.Interface = &wantInternalWardleInformerFactory{}
var _ custominitializer.WantsInternalWardleInformerFactory = &wantInternalWardleInformerFactory{}
