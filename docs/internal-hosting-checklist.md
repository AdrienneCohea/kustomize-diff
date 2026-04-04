# Internal Hosting Checklist

Checklist for hosting `kustomize-diff` on a self-hosted GitHub Enterprise instance with AWS ECR.

## 1. Fork/mirror the repository
- [ ] Create the repo on your internal GHE instance (`scm.example.com`)
- [ ] Update `go.mod` module path to your internal GHE hostname

## 2. Container registry (AWS ECR)
- [ ] Create an ECR repository for `kustomize-diff`
- [ ] Update `action.yml` line 20 to reference your ECR image URL
- [ ] Update `.github/workflows/release.yml` to push to ECR instead of `ghcr.io`
- [ ] Configure IRSA or Pod Identity on your runner pods with an IAM role that allows push to the ECR repo
- [ ] Confirm `amazon-ecr-credential-helper` is configured in the runner pod Docker config — if other workflows already pull from ECR, this is likely already done

## 3. Dockerfile build dependencies
- [ ] Ensure runner pods can reach public Docker Hub (for `golang:1.23-alpine`, `alpine:3.21`) — or mirror those base images to ECR
- [ ] Ensure runner pods can reach `github.com/kubernetes-sigs/kustomize` at build time — or pre-download the binary and `COPY` it into the image instead

## 4. Actions dependencies
The release workflow references third-party Actions (`actions/checkout@v4`, `docker/login-action@v3`, `docker/build-push-action@v6`). Confirm these are available on your GHE instance via one of:
- [ ] GitHub Connect (syncs from github.com)
- [ ] Manually mirrored to your GHE instance
- [ ] Replaced with inline `run:` steps
