module github.com/dbsystel/prometheus-config-controller

go 1.12

require (
	github.com/dbsystel/kube-controller-dbsystel-go-common v0.0.0-20190307121541-2d8f1275b8b2
	github.com/go-kit/kit v0.8.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/prometheus v0.0.0-20190424153033-d3245f150225
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.0.0-20181213150558-05914d821849
)
