package docker

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	client "docker.io/go-docker"
	"docker.io/go-docker/api"
	"github.com/polyverse/play-with-polyverse/pwd/types"
	"github.com/polyverse/play-with-polyverse/router"
)

type FactoryApi interface {
	GetForSession(session *types.Session) (DockerApi, error)
	GetForInstance(instance *types.Instance) (DockerApi, error)
}

func NewClient(instance *types.Instance, proxyHost string) (*client.Client, error) {
	var host string
	var durl string

	host = router.EncodeHost(instance.SessionId, instance.RoutableIP, router.HostOpts{EncodedPort: 2375})

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost: 5,
	}

	transport.Proxy = http.ProxyURL(&url.URL{Host: proxyHost})
	durl = fmt.Sprintf("http://%s", host)

	cli := &http.Client{
		Transport: transport,
	}

	dc, err := client.NewClient(durl, api.DefaultVersion, cli, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to DinD docker daemon: %v", err)
	}

	return dc, nil
}
