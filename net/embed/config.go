package embed

import "net/url"

const(

	DefaultName                  = "default"

	DefaultListenPeerURLs = "http://localhost:2380"
	DefaultListenClientURLs = "http://localhost:2379"
)

type Config struct {
	Name string `json:"name"`
	LPUrls,LCUrls []url.URL
}

func NewConfig() *Config {
	lpurl,_ :=url.Parse(DefaultListenPeerURLs)
	lcurl,_ := url.Parse(DefaultListenClientURLs)
	cfg := &Config{
		Name: DefaultName,
		LPUrls:[]url.URL{*lpurl},
		LCUrls: []url.URL{*lcurl},
	}
	return cfg
}
