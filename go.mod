module github.com/getlantern/lantern-shadowsocks

require (
	github.com/getlantern/fdcount v0.0.0-20210503151800-5decd65b3731
	github.com/getlantern/grtrack v0.0.0-20160824195228-cbf67d3fa0fd
	github.com/getlantern/transports v0.0.0-00010101000000-000000000000
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/prometheus/client_golang v1.7.1
	github.com/shadowsocks/go-shadowsocks2 v0.1.4-0.20201002022019-75d43273f5a5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/getlantern/transports => ../transports

// TODO: this kind of sucks, but seems unavoidable... unless transports just uses getlantern/utls?
replace github.com/refraction-networking/utls => github.com/getlantern/utls v0.0.0-20200903013459-0c02248f7ce1

go 1.14
