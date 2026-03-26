module btaskee-quiz

go 1.23

require (
	github.com/RussellLuo/timingwheel v0.0.0-20220218152713-54845bda3108
	github.com/WuKongIM/crypto v0.0.0-20240416072338-b872b70b395f
	github.com/gin-gonic/gin v1.8.2
	github.com/gobwas/ws v1.2.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/lni/goutils v1.4.0
	github.com/panjf2000/ants/v2 v2.11.0
	github.com/panjf2000/gnet/v2 v2.7.1
	github.com/prometheus/client_golang v1.19.1
	github.com/prometheus/common v0.55.0
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/sendgrid/rest v2.6.9+incompatible
	github.com/stretchr/testify v1.10.0
	go.etcd.io/etcd/pkg/v3 v3.5.17
	go.etcd.io/raft/v3 v3.6.0-beta.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/prometheus v0.46.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/sdk/metric v1.28.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.26.0
	google.golang.org/protobuf v1.36.3
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.19.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.25.0
	golang.org/x/text v0.17.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude k8s.io/client-go v8.0.0+incompatible

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.14.5

// replace github.com/WuKongIM/WuKongIMGoSDK => ../../WuKongIMGoSDK
// replace github.com/WuKongIM/WuKongIMGoProto => ../../WuKongIMGoProto

replace github.com/WuKongIM/WuKongIM => ./
