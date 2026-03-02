package admission

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bonial-oss/pod-image-swap-webhook/pkg/config"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TestPodImageHandler tests the admission handler end-to-end.
//
// For each test case, it creates the handler as a standalone webhook using the
// *config.Config from the test case and then calls the handler with an
// AdmissionReview request which contains the provided Pod manifest.
//
// Finally, it makes assertions about the result and patch operations returned
// in the AdmissionReview response.
func TestPodImageHandler(t *testing.T) {
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			webhookConfig := testCase.config
			if webhookConfig == nil {
				webhookConfig = &config.Config{}
			}

			handler := NewPodImageHandler(webhookConfig, runtime.NewScheme())

			httpHandler, err := admission.StandaloneWebhook(
				&admission.Webhook{Handler: handler},
				admission.StandaloneOptions{},
			)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			req := buildAdmissionRequest(t, testCase.pod)
			httpHandler.ServeHTTP(rr, req)
			resp := parseAdmissionResponse(t, rr.Body)

			assert.Contains(t, resp.Result.Message, testCase.expectedMessage)
			assert.Equal(t, int32(testCase.expectedCode), resp.Result.Code)

			if testCase.expectedPatches != nil {
				var patches []jsonpatch.JsonPatchOperation

				err = json.Unmarshal(resp.Patch, &patches)
				require.NoError(t, err)

				assert.ElementsMatch(t, testCase.expectedPatches, patches)
			}
		})
	}
}

func buildAdmissionRequest(t *testing.T, pod *corev1.Pod) *http.Request {
	raw, err := json.Marshal(pod)
	require.NoError(t, err)

	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: raw,
			},
		},
	}

	buf, err := json.Marshal(review)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(buf))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	return req
}

func parseAdmissionResponse(t *testing.T, r io.Reader) *admissionv1.AdmissionResponse {
	var review admissionv1.AdmissionReview

	err := json.NewDecoder(r).Decode(&review)
	require.NoError(t, err)

	return review.Response
}

// TestNormalizeImage tests the image normalization logic which expands images
// from dockerhub to their full form (`docker.io/<namespace>/<image-name>`).
func TestNormalizeImage(t *testing.T) {
	assert.Equal(t, "docker.io/library/nginx:latest", normalizeImage("nginx:latest"))
	assert.Equal(t, "docker.io/goharbor/harbor-core:v2.4.2", normalizeImage("goharbor/harbor-core:v2.4.2"))
	assert.Equal(t, "docker.io/goharbor/harbor-core:v2.4.2", normalizeImage("docker.io/goharbor/harbor-core:v2.4.2"))
	assert.Equal(t, "k8s.gcr.io/ingress-nginx/controller:v0.48.1", normalizeImage("k8s.gcr.io/ingress-nginx/controller:v0.48.1"))
	assert.Equal(t, "docker.io/library/docker:20.10.12-dind-alpine3.15", normalizeImage("docker:20.10.12-dind-alpine3.15"))
}
