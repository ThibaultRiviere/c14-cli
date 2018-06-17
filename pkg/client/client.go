package client

import (
	"golang.org/x/oauth2"
	"os"

	"github.com/online-net/c14-cli/pkg/api"
	"github.com/online-net/c14-cli/pkg/api/auth"
	"github.com/online-net/c14-cli/pkg/version"
)

func InitAPI() (cli *api.OnlineAPI, err error) {
	var (
		c            *auth.Credentials
		privateToken string
	)

	if privateToken = os.Getenv("C14_PRIVATE_TOKEN"); privateToken != "" {
		c = &auth.Credentials{
			AccessToken: privateToken,
		}
	} else {
		if c, err = auth.GetCredentials(); err != nil {
			return
		}
	}
	cli = api.NewC14API(oauth2.NewClient(oauth2.NoContext, c), version.UserAgent, true)
	return
}
