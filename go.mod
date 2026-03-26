module btaskee-quiz

go 1.23

require (
	github.com/RussellLuo/timingwheel v0.0.0-20220218152713-54845bda3108
	github.com/WuKongIM/crypto v0.0.0-20240416072338-b872b70b395f
	github.com/gin-gonic/gin v1.8.2
	github.com/gobwas/ws v1.2.1
	github.com/google/uuid v1.6.0
	github.com/lni/goutils v1.4.0
	github.com/panjf2000/ants/v2 v2.11.0
	github.com/panjf2000/gnet/v2 v2.7.1
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/sendgrid/rest v2.6.9+incompatible
	go.etcd.io/etcd/pkg/v3 v3.5.17
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.26.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.19.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/net v0.28.0 // indirect
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
)

exclude k8s.io/client-go v8.0.0+incompatible

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.14.5

// replace github.com/WuKongIM/WuKongIMGoSDK => ../../WuKongIMGoSDK
// replace github.com/WuKongIM/WuKongIMGoProto => ../../WuKongIMGoProto

replace github.com/WuKongIM/WuKongIM => ./
