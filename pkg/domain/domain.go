package domain

import (
	"net/http"
	"strings"
	"os/exec"
	"fmt"

	"github.com/harsh-98/inlets/pkg/transport"
)
const (
	InletsHeader   = "x-inlets-id"
	UpstreamHeader = "x-inlets-upstream"
)
func RegisterDomain(req *http.Request){
	upstreams := req.Header[http.CanonicalHeaderKey(transport.UpstreamHeader)]

	for _, upstream := range upstreams {
		parts := strings.SplitN(upstream, "=", 2)
		if len(parts) != 2 {
			continue
		}

		cmd := exec.Command("az", fmt.Sprintf("network dns record-set a add-record -g Blockchain -z tunzal.ml -n %s  -a 52.187.64.208", parts[0]))
		err := cmd.Run()
		if err != nil {
			fmt.Errorf("cmd.Run() failed with %s\n", err)
		}
	}
}