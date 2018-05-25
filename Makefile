.PHONY: clean codegen codegen-verify build istio-setup kubectl-apply

# This is ONE of the generated files (alongside everything in pkg/client)
# that serves as make dependency tracking
GENERATED_SOURCE = pkg/apis/spike.local/v1alpha1/zz_generated.deepcopy.go

GO_SOURCES = $(shell find pkg/apis -type f -name '*.go' ! -path $(GENERATED_SOURCE))

codegen: $(GENERATED_SOURCE)

$(GENERATED_SOURCE): $(GO_SOURCES) hack/vendor vendor
	hack/vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/scothis/stream-spike/pkg/client \
      github.com/scothis/stream-spike/pkg/apis \
      "spike.local:v1alpha1 config.istio.io:v1alpha2" \
      --go-header-file  hack/boilerplate.go.txt
	hack/vendor/k8s.io/code-generator/generate-internal-groups.sh defaulter \
      github.com/scothis/stream-spike/pkg/client \
      '' \
      github.com/scothis/stream-spike/pkg/apis \
      "spike.local:v1alpha1 config.istio.io:v1alpha2" \
      --go-header-file  hack/boilerplate.go.txt

codegen-verify: hack/vendor vendor
	hack/vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/scothis/stream-spike/pkg/client \
      github.com/scothis/stream-spike/pkg/apis \
      "spike.local:v1alpha1 config.istio.io:v1alpha2" \
      --go-header-file  hack/boilerplate.go.txt \
      --verify-only

clean:
	rm -fR pkg/client
	rm -f $(GENERATED_SOURCE)

vendor: glide.lock
	glide install -v --force

glide.lock: glide.yaml
	glide up -v --force

hack/vendor: hack/glide.lock
	# Note the absence of -v
	cd hack && glide install

hack/glide.lock: hack/glide.yaml
	# Note the absence of -v
	cd hack && glide up

istio-setup:
	kubectl apply -f istio-0.7.1/istio-auth.yaml
	./istio-0.7.1/webhook-create-signed-cert.sh \
		--service istio-sidecar-injector \
		--namespace istio-system \
		--secret sidecar-injector-certs
	kubectl apply -f istio-0.7.1/istio-sidecar-injector-configmap-release.yaml
	cat istio-0.7.1/istio-sidecar-injector.yaml | \
		./istio-0.7.1/webhook-patch-ca-bundle.sh | \
		kubectl apply -f -
	kubectl label namespace default istio-injection=enabled

kubectl-apply:
	kubectl apply -f config/rbac.yaml
	kubectl apply -f config/stream-resource.yaml
	kubectl apply -f config/subscription-resource.yaml
	ko apply -f config/controller-deployment.yaml
