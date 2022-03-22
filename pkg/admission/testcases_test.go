package admission

import (
	"net/http"

	"github.com/Bonial-International-GmbH/pod-image-swap-webhook/pkg/config"
	"gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
)

type testCase struct {
	name            string
	config          *config.Config
	pod             *corev1.Pod
	expectedCode    int
	expectedPatches []jsonpatch.JsonPatchOperation
	expectedMessage string
}

var testCases = []testCase{
	{
		name:            "empty pod input is invalid",
		expectedCode:    http.StatusBadRequest,
		expectedMessage: "there is no content to decode",
	},
	{
		name: "empty config does not cause pod mutations",
		pod: &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "nginx:latest"},
				},
			},
		},
		expectedCode: http.StatusOK,
	},
	{
		name: "replace images from dockerhub",
		config: &config.Config{
			Replace: []config.ReplacementRule{
				{
					Prefix:      "docker.io",
					Replacement: "registry.example.com/docker.io",
				},
			},
		},
		pod: &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "goharbor/harbor-core:v2.4.2"},
				},
				InitContainers: []corev1.Container{
					{Image: "busybox:latest"},
				},
			},
		},
		expectedCode: http.StatusOK,
		expectedPatches: []jsonpatch.JsonPatchOperation{
			{
				Operation: "replace",
				Path:      "/spec/initContainers/0/image",
				Value:     "registry.example.com/docker.io/library/busybox:latest",
			},
			{
				Operation: "replace",
				Path:      "/spec/containers/0/image",
				Value:     "registry.example.com/docker.io/goharbor/harbor-core:v2.4.2",
			},
		},
	},
}
