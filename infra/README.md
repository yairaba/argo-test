Installation of ArgoCD installed by FluxCD


flux bootstrap git \
  --url=ssh://git@github.com/binboum/argo-test \
  --branch=main \
  --path=infra \
  --private-key-file=/home/docker/.ssh/id_rsa


