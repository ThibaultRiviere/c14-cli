package client

import (
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/online-net/c14-cli/pkg/api"
	"github.com/online-net/c14-cli/pkg/api/auth"
	"github.com/online-net/c14-cli/pkg/utils/ssh"
	"github.com/online-net/c14-cli/pkg/version"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
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

func GetsftpCred(c *api.OnlineAPI, archive string) (sftpCred sshUtils.Credentials, err error) {
	var (
		safe        api.OnlineGetSafe
		bucket      api.OnlineGetBucket
		uuidArchive string
	)

	if safe, uuidArchive, err = c.FindSafeUUIDFromArchive(archive, true); err != nil {
		if safe, uuidArchive, err = c.FindSafeUUIDFromArchive(archive, false); err != nil {
			return
		}
	}
	if bucket, err = c.GetBucket(safe.UUIDRef, uuidArchive); err != nil {
		return
	}

	fmt.Println(strings.Split(bucket.Credentials[0].URI, "@")[1])
	sftpCred.Host = strings.Split(bucket.Credentials[0].URI, "@")[1]
	sftpCred.Password = bucket.Credentials[0].Password
	sftpCred.User = bucket.Credentials[0].Login
	return
}

func GetsftpConn(sftpCred sshUtils.Credentials) (conn *sftp.Client, err error) {
	conn, err = sftpCred.NewSFTPClient()
	return
}

func UploadAFile(c *sftp.Client, reader io.ReadCloser, file string, size int64, padding int) (err error) {
	log.Debugf("Upload %s -> /buffer/%s", file, file)

	var (
		buff   = make([]byte, 1<<23)
		nr, nw int
		w      *sftp.File
	)
	if w, err = c.Create(fmt.Sprintf("/buffer/%s", file)); err != nil {
		return
	}
	defer w.Close()
	if size == 0 {
		log.Warnf("upload %s is empty", file)
		return
	}
	sf := streamformatter.NewStreamFormatter()
	progressBarOutput := sf.NewProgressOutput(os.Stdout, true)
	rc := progress.NewProgressReader(reader, progressBarOutput, size, "", fmt.Sprintf("%-*s", padding, file))
	defer rc.Close()
	for {
		nr, err = rc.Read(buff)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if nw, err = w.Write(buff[:nr]); err != nil {
			return
		}
		if nw != nr {
			err = errors.Errorf("Error during write")
			return
		}
	}
	return
}
