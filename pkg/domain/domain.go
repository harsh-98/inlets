package domain

import (
	"net/http"
	"strings"
	"os/exec"
	"os"
	"fmt"

	"github.com/harsh-98/inlets/pkg/transport"
)

func getDomainCmd() string{
	DOMAINCMD, exists := os.LookupEnv("DOMAINCMD")

    if exists {
	return DOMAINCMD
	}
	return ""
}

func RegisterDomain(req *http.Request){
	upstreams := req.Header[http.CanonicalHeaderKey(transport.UpstreamHeader)]

	for _, upstream := range upstreams {
		parts := strings.SplitN(upstream, "=", 2)
		if len(parts) != 2 {
			continue
		}
		args:= fmt.Sprintf(getDomainCmd(), parts[0])
		command:= strings.Split(args, " ")

		cmd := exec.Command(command[0], command[1:]...)
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Errorf("%s", err)
		}
		// do something with output
		fmt.Printf("%s\n", stdoutStderr)
	}
}