module helm-rename

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	helm.sh/helm/v3 v3.5.4
	k8s.io/apimachinery v0.20.4
	k8s.io/cli-runtime v0.20.4
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
