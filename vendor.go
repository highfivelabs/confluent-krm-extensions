// Copyright 2026 High Five Labs
// SPDX-License-Identifier: Apache-2.0

package extensions

import (
	"net/http"
	"strings"

	"github.com/highfivelabs/krm-rest-runtime/rest"
)

// Vendor is how the manager reaches Confluent Cloud.
var Vendor = &rest.Vendor{
	Name:           "confluent",
	DefaultBaseURL: "https://api.confluent.cloud",
	CredentialKeys: []string{"apiKey", "apiSecret"},
	Authorize:      func(r *http.Request, c map[string]string) { r.SetBasicAuth(c["apiKey"], c["apiSecret"]) },
	// Reading or deleting an API key that does not exist (deleted, or never
	// made) answers 403 forbidden_access instead of 404. Scoped to
	// /iam/v2/api-keys/{id} so real 403s elsewhere stay errors.
	NotFoundReq: func(method, path string, status int, body []byte) bool {
		if status != http.StatusForbidden || (method != http.MethodGet && method != http.MethodDelete) {
			return false
		}
		p, _, _ := strings.Cut(path, "?")
		return strings.HasPrefix(p, "/iam/v2/api-keys/") && len(p) > len("/iam/v2/api-keys/") &&
			strings.Contains(string(body), "forbidden_access")
	},
}
