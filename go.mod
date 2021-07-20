module github.com/nacos-group/nacos-k8s-sync

go 1.15

require (
	github.com/hashicorp/go-multierror v1.1.0
	github.com/nacos-group/nacos-sdk-go v1.0.7-0.20210312023737-9edc707e7511
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
)
