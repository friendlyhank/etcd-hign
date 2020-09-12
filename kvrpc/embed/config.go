package embed

import(
	"net/url"
)

type Config struct {
	MaxRequestBytes   uint  `json:"max-request-bytes"` //grpc最大请求书

	LPUrls, LCUrls []url.URL
	APUrls, ACUrls []url.URL
}