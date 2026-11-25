// Copyright 2026 High Five Labs
// SPDX-License-Identifier: Apache-2.0

package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/highfivelabs/krm-rest-runtime/descriptor"
	apifake "github.com/highfivelabs/krm-rest-runtime/fake"
	"github.com/highfivelabs/krm-rest-runtime/reconciler"
)

// Confluent never reads a webhook integration's URL back (the live run of
// 2026-11-20: target = {kind: Webhook}). The URL is write-only, so a read
// without it is not drift (HFL-675).
func TestWebhookIntegrationURLIsNotDrift(t *testing.T) {
	f, err := descriptor.Parse(Types)
	if err != nil {
		t.Fatal(err)
	}
	var k *descriptor.Kind
	for i := range f.Kinds {
		if f.Kinds[i].Kind == "NotificationsIntegration" {
			k = &f.Kinds[i]
		}
	}
	if k == nil {
		t.Fatal("NotificationsIntegration not in types.yaml")
	}
	api := apifake.New(f)
	// As Confluent does, whatever the descriptor says: reads leave out
	// target.url.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		api.ServeHTTP(rec, r)
		body := rec.Body.Bytes()
		if r.Method == http.MethodGet {
			var m map[string]any
			if json.Unmarshal(body, &m) == nil {
				if tg, ok := m["target"].(map[string]any); ok {
					delete(tg, "url")
				}
				body, _ = json.Marshal(m)
			}
		}
		for h, v := range rec.Header() {
			w.Header()[h] = v
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(bytes.TrimSpace(body))
	}))
	defer srv.Close()

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	mapper := meta.NewDefaultRESTMapper(nil)
	mapper.Add(k.GVK(), meta.RESTScopeNamespace)
	mapper.Add(f.ConnectionGVK(), meta.RESTScopeNamespace)
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Secret"), meta.RESTScopeNamespace)
	creds := map[string][]byte{}
	for _, key := range Vendor.CredentialKeys {
		creds[key] = []byte("test")
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "team-a", Name: "creds"}, Data: creds}
	conn := &unstructured.Unstructured{}
	conn.SetGroupVersionKind(f.ConnectionGVK())
	conn.SetNamespace("team-a")
	conn.SetName("default")
	_ = unstructured.SetNestedField(conn.Object, srv.URL, "spec", "baseURL")
	_ = unstructured.SetNestedField(conn.Object, "creds", "spec", "secretRef", "name")
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(k.GVK())
	obj.SetNamespace("team-a")
	obj.SetName("hook")
	obj.Object["spec"] = map[string]any{"display_name": "hook", "target": map[string]any{"kind": "Webhook", "url": "https://example.com/hook"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(secret, conn, obj).WithStatusSubresource(obj).Build()
	r := &reconciler.Reconciler{Client: c, Kind: k, Vendor: Vendor, Connection: f.ConnectionGVK(), Kinds: reconciler.KindsByName(f)}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "team-a", Name: "hook"}}

	for i, want := range []string{"Created", "Unchanged"} {
		if _, err := r.Reconcile(context.Background(), req); err != nil {
			t.Fatalf("reconcile %d: %v", i+1, err)
		}
		got := &unstructured.Unstructured{}
		got.SetGroupVersionKind(k.GVK())
		if err := c.Get(context.Background(), req.NamespacedName, got); err != nil {
			t.Fatal(err)
		}
		conds, _, _ := unstructured.NestedSlice(got.Object, "status", "conditions")
		cond, _ := conds[0].(map[string]any)
		if cond["status"] != "True" || cond["reason"] != want {
			t.Fatalf("reconcile %d: Ready %v %v (%v), want True %s", i+1, cond["status"], cond["reason"], cond["message"], want)
		}
	}
}
