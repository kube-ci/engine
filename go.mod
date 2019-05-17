module github.com/kube-ci/engine

go 1.12

require (
	cloud.google.com/go v0.39.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.5.0 // indirect
	github.com/Azure/go-autorest v12.0.0+incompatible // indirect
	github.com/appscode/go v0.0.0-20190424183524-60025f1135c9
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/drone/envsubst v0.0.0-00010101000000-000000000000
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/go-openapi/jsonpointer v0.19.0 // indirect
	github.com/go-openapi/jsonreference v0.19.0 // indirect
	github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190516165734-b3a23cc94cc5 // indirect
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/philopon/go-toposort v0.0.0-20170620085441-9be86dbd762f
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3 // indirect
	github.com/prometheus/procfs v0.0.0-20190517135640-51af30a78b0e // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/tamalsaha/go-oneliners v0.0.0-20190126213733-6d24eabef827
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20190514140710-3ec191127204 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190516110030-61b9204099cb // indirect
	gomodules.xyz/cert v1.0.0
	google.golang.org/appengine v1.6.0 // indirect
	google.golang.org/genproto v0.0.0-20190516172635-bb713bdc0e52 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	k8s.io/api v0.0.0-20190515023547-db5a9d1c40eb
	k8s.io/apiextensions-apiserver v0.0.0-20190515024537-2fd0e9006049
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
	k8s.io/apiserver v0.0.0-20190515064100-fc28ef5782df
	k8s.io/cli-runtime v0.0.0-20190515024640-178667528169 // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider v0.0.0-20190515224602-984e89dd9575 // indirect
	k8s.io/component-base v0.0.0-20190515024022-2354f2393ad4 // indirect
	k8s.io/kube-aggregator v0.0.0-20190515024249-81a6edcf70be
	k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22
	k8s.io/kubernetes v1.14.2
	kmodules.xyz/client-go v0.0.0-20190515205239-a16030cc2e50
	kmodules.xyz/openshift v0.0.0-20190508141315-99ec9fc946bf // indirect
	kmodules.xyz/webhook-runtime v0.0.0-20190508093950-b721b4eba5e5
)

replace (
	github.com/drone/envsubst => github.com/appscode/envsubst v1.0.2-0.20180924062550-040bfb31793a
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	gopkg.in/robfig/cron.v2 => github.com/appscode/cron v0.0.0-20170717094345-ca60c6d796d4
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
)

replace k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed

replace k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4

replace k8s.io/klog => k8s.io/klog v0.3.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213

replace k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
