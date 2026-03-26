module btaskee-quiz

go 1.23.0

require (
	github.com/RussellLuo/timingwheel v0.0.0-20220218152713-54845bda3108
	github.com/WuKongIM/crypto v0.0.0-20240416072338-b872b70b395f
	github.com/gobwas/ws v1.2.1
	github.com/gorilla/websocket v1.5.3
	github.com/lni/goutils v1.4.0
	github.com/panjf2000/ants/v2 v2.11.0
	github.com/panjf2000/gnet/v2 v2.7.1
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.18.0
	github.com/sasha-s/go-deadlock v0.3.1
	go.etcd.io/etcd/pkg/v3 v3.5.17
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/sync v0.13.0 // indirect

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.35.0
)

exclude k8s.io/client-go v8.0.0+incompatible

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.14.5

// replace github.com/WuKongIM/WuKongIMGoSDK => ../../WuKongIMGoSDK
// replace github.com/WuKongIM/WuKongIMGoProto => ../../WuKongIMGoProto

replace github.com/WuKongIM/WuKongIM => ./
