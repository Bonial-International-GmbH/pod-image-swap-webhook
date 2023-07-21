package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Bonial-International-GmbH/pod-image-swap-webhook/pkg/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var logger = log.Log.WithName("admission")

// PodImageHandler is an admission handler that mutates Pod container images
// based on replacement rules.
type PodImageHandler struct {
	config  *config.Config
	decoder *admission.Decoder
}

// NewPodImageHandler creates a new *PodImageHandler which mutates Pod
// container images according to the provided configuration.
func NewPodImageHandler(config *config.Config, scheme *runtime.Scheme) *PodImageHandler {
	return &PodImageHandler{
		config:  config,
		decoder: admission.NewDecoder(scheme),
	}
}

// Handle implements admission.Handler.
func (h *PodImageHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	var pod corev1.Pod

	err := h.decoder.Decode(req, &pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	pod = h.patchPod(pod)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		// This should never happen, but we handle it anyways.
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (h *PodImageHandler) patchPod(pod corev1.Pod) corev1.Pod {
	logger.V(1).Info("patching pod", "namespace", pod.Namespace, "name", pod.Name)

	pod.Spec.InitContainers = h.patchContainers(pod.Spec.InitContainers)
	pod.Spec.Containers = h.patchContainers(pod.Spec.Containers)

	return pod
}

func (h *PodImageHandler) patchContainers(containers []corev1.Container) []corev1.Container {
	for i, container := range containers {
		containers[i] = h.patchContainer(container)
	}

	return containers
}

func (h *PodImageHandler) patchContainer(container corev1.Container) corev1.Container {
	image := normalizeImage(container.Image)

	for _, rule := range h.config.Exclude {
		if strings.HasPrefix(image, rule.Prefix) {
			logger.Info("image excluded from replacement via config, not patching", "image", image)

			return container
		}
	}

	for _, rule := range h.config.Replace {
		if strings.HasPrefix(image, rule.Prefix) {
			replacedImage := strings.Replace(image, rule.Prefix, rule.Replacement, 1)
			container.Image = replacedImage

			logger.Info("patching container image", "from", image, "to", replacedImage)

			return container
		}
	}

	return container
}

// normalizeImage normalizes images from dockerhub to their long form.
// Dockerhub images without namespace are prefixed with `docker.io/library`,
// namespaced images are prefixed with `docker.io/`. Images that already start
// with `docker.io/` or any other registry domain are left untouched.
//
// Examples:
//
//	nginx:latest => docker.io/library/nginx:latest
//	goharbor/harbor-core:v2.4.2 => docker.io/goharbor/harbor-core:v2.4.2
func normalizeImage(image string) string {
	parts := strings.Split(image, "/")
	if len(parts) == 1 {
		// Image without namespace from dockerhub.
		return fmt.Sprintf("docker.io/library/%s", parts[0])
	}

	if strings.Contains(parts[0], ".") {
		// Image starts with a registry domain, no normalization needed.
		return image
	}

	// Namespaced image from dockerhub.
	return fmt.Sprintf("docker.io/%s", strings.Join(parts, "/"))
}
