/*
Copyright 2016 The Kubernetes Authors.

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

package bootstrap

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	api "k8s.io/kubernetes/pkg/apis/core"
)

func init() {
	spew.Config.DisableMethods = true
}

func newTokenCleaner() (*TokenCleaner, *fake.Clientset, error) {
	options := DefaultTokenCleanerOptions()
	cl := fake.NewSimpleClientset()
	tcc, err := NewTokenCleaner(cl, options)
	if err != nil {
		return nil, nil, err
	}
	return tcc, cl, nil
}

func TestCleanerNoExpiration(t *testing.T) {
	cleaner, cl, err := newTokenCleaner()
	if err != nil {
		t.Fatalf("error creating TokenCleaner: %v", err)
	}

	secret := newTokenSecret("tokenID", "tokenSecret")
	cleaner.secrets.Add(secret)

	cleaner.evalSecret(secret)

	expected := []core.Action{}

	verifyActions(t, expected, cl.Actions())
}

func TestCleanerExpired(t *testing.T) {
	cleaner, cl, err := newTokenCleaner()
	if err != nil {
		t.Fatalf("error creating TokenCleaner: %v", err)
	}

	secret := newTokenSecret("tokenID", "tokenSecret")
	addSecretExpiration(secret, timeString(-time.Hour))
	cleaner.secrets.Add(secret)

	cleaner.evalSecret(secret)

	expected := []core.Action{
		core.NewDeleteAction(
			schema.GroupVersionResource{Version: "v1", Resource: "secrets"},
			api.NamespaceSystem,
			secret.ObjectMeta.Name),
	}

	verifyActions(t, expected, cl.Actions())
}

func TestCleanerNotExpired(t *testing.T) {
	cleaner, cl, err := newTokenCleaner()
	if err != nil {
		t.Fatalf("error creating TokenCleaner: %v", err)
	}

	secret := newTokenSecret("tokenID", "tokenSecret")
	addSecretExpiration(secret, timeString(time.Hour))
	cleaner.secrets.Add(secret)

	cleaner.evalSecret(secret)

	expected := []core.Action{}

	verifyActions(t, expected, cl.Actions())
}