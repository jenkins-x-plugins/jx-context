module github.com/jenkins-x-plugins/jx-context

replace (
	k8s.io/api => k8s.io/api v0.20.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.6
	k8s.io/client-go => k8s.io/client-go v0.20.6
)

go 1.16

require (
	github.com/jenkins-x/jx-helpers/v3 v3.0.130
	github.com/jenkins-x/jx-kube-client/v3 v3.0.2
	github.com/jenkins-x/jx-logging/v3 v3.0.6
	github.com/spf13/cobra v1.2.1
	k8s.io/client-go v11.0.0+incompatible
)
