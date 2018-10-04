package config

import (
	"flag"
	"regexp"

	"github.com/gorilla/securecookie"

	"golang.org/x/oauth2"
)

const (
	PWDHostnameRegex      = "[0-9]{1,3}-[0-9]{1,3}-[0-9]{1,3}-[0-9]{1,3}"
	PortRegex             = "[0-9]{1,5}"
	AliasnameRegex        = "[0-9|a-z|A-Z|-]*"
	AliasSessionRegex     = "[0-9|a-z|A-Z]{8}"
	AliasGroupRegex       = "(" + AliasnameRegex + ")-(" + AliasSessionRegex + ")"
	PWDHostPortGroupRegex = "^.*ip(" + PWDHostnameRegex + ")(?:-?(" + PortRegex + "))?(?:\\..*)?$"
	AliasPortGroupRegex   = "^.*pwd" + AliasGroupRegex + "(?:-?(" + PortRegex + "))?\\..*$"
)

var (
	NameFilter  = regexp.MustCompile(PWDHostPortGroupRegex)
	AliasFilter = regexp.MustCompile(AliasPortGroupRegex)

	PortNumber, PWDContainerName, L2ContainerName, L2Subdomain, HashKey, L2RouterIP,
	DindVolumeSize, CookieHashKey, CookieBlockKey,
	DefaultDinDImage, DefaultSessionDuration, S3Bucket, AdminToken, SessionsFile string
	MaxLoadAvg   float64
	SecureCookie *securecookie.SecureCookie

	PlaygroundDomain string

	SegmentId string

	// TODO move this to a sync map so it can be updated on demand when the configuration for a playground changes
	Providers = map[string]map[string]*oauth2.Config{}
)

func ParseFlags() {
	flag.StringVar(&PortNumber, "port", "3000", "Port number")
	flag.StringVar(&SessionsFile, "sessions-file", "./pwd/sessions", "File to store sessions in.")
	flag.StringVar(&PWDContainerName, "name", "pwd", "Container name used to run PWD (used to be able to connect it to the networks it creates)")
	flag.StringVar(&L2ContainerName, "l2", "l2", "Container name used to run L2 Router")
	flag.StringVar(&L2RouterIP, "l2-ip", "", "Host IP address for L2 router ping response")
	flag.StringVar(&L2Subdomain, "l2-subdomain", "direct", "Subdomain to the L2 Router")
	flag.StringVar(&HashKey, "hash_key", "salmonrosado", "Hash key to use for cookies")
	flag.StringVar(&DindVolumeSize, "dind-volume-size", "5G", "Dind volume folder size")
	flag.Float64Var(&MaxLoadAvg, "maxload", 100, "Maximum allowed load average before failing ping requests")
	flag.StringVar(&CookieHashKey, "cookie-hash-key", "", "Hash key to use to validate cookies")
	flag.StringVar(&CookieBlockKey, "cookie-block-key", "", "Block key to use to encrypt cookies")
	flag.StringVar(&DefaultDinDImage, "default-dind-image", "franela/dind", "Default DinD image to use if not specified otherwise")
	flag.StringVar(&DefaultSessionDuration, "default-session-duration", "30m", "Default session duration if not specified otherwise")
	flag.StringVar(&S3Bucket, "s3bucket", "", "AWS S3 Bucket to write state to.")
	flag.StringVar(&PlaygroundDomain, "playground-domain", "localhost", "Domain to use for the playground")
	flag.StringVar(&AdminToken, "admin-token", "foobar", "Admin token")

	flag.StringVar(&SegmentId, "segment-id", "", "Segment id to post metrics")

	flag.Parse()

	SecureCookie = securecookie.New([]byte(CookieHashKey), []byte(CookieBlockKey))

}
