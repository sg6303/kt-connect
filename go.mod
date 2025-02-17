module github.com/alibaba/kt-connect

go 1.15

require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/cilium/ipam v0.0.0-20201106170308-4184bc4bf9d6
	github.com/deckarep/golang-set v1.7.1
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gin-gonic/gin v1.7.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/websocket v1.4.1
	github.com/kubernetes/dashboard v1.10.1
	github.com/linfan/socks4 v0.2.3-2
	github.com/miekg/dns v1.1.31
	github.com/mitchellh/go-ps v1.0.0
	github.com/rs/zerolog v1.23.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
	istio.io/api v0.0.0-20200221025927-228308df3f1b
	istio.io/client-go v0.0.0-20200221055756-736d3076b458
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
