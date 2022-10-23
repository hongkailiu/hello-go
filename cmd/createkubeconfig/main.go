package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/test-infra/prow/logrusutil"
	"os"
	"path/filepath"
	"strings"
)

type options struct {
	logLevel       string
	kubeconfig     string
	token          string
	tokenFile      string
	server         string
	cluster        string
	serviceAccount string
	namespace      string
}

func gatherOptions() (options, error) {
	o := options{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.logLevel, "log-level", "info", "Level at which to log output.")
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file to use.")
	fs.StringVar(&o.token, "token", "", "The token to use in the kubeconfig file.")
	fs.StringVar(&o.tokenFile, "token-file", "", "Path to the token file to use in the kubeconfig file. If it is a file name containing no path separator, it will be saved under the same directory of the kubeconfig file")
	fs.StringVar(&o.server, "server", "", "The Kubernetes API server.")
	fs.StringVar(&o.serviceAccount, "service-account", "", "The service account.")
	fs.StringVar(&o.namespace, "namespace", "ci", "The namespace of the service account.")
	fs.StringVar(&o.cluster, "cluster", "", "The cluster name.")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return o, fmt.Errorf("failed to parse flags: %w", err)
	}
	return o, nil
}

func validateOptions(o options) error {
	_, err := log.ParseLevel(o.logLevel)
	if err != nil {
		return fmt.Errorf("invalid --log-level: %w", err)
	}
	if o.token == "" && o.tokenFile == "" {
		return fmt.Errorf("either --token or --token-file must be specified")
	}
	if o.serviceAccount == "" {
		return fmt.Errorf("--service-account must be specified")
	}
	if o.namespace == "" {
		return fmt.Errorf("--namespace must be specified")
	}
	if o.cluster == "" {
		return fmt.Errorf("--cluster must be specified")
	}
	if o.server == "" {
		return fmt.Errorf("--server must be specified")
	}

	if o.kubeconfig == "" {
		return fmt.Errorf("--kubeconfig must be specified")
	}
	return nil
}

func createKubeconfig(server, cluster, namespace, token, tokenFile string) clientcmdapi.Config {
	authInfo := &clientcmdapi.AuthInfo{}
	if tokenFile != "" {
		authInfo.TokenFile = tokenFile
	} else {
		authInfo.Token = token
	}

	return clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			cluster: {
				Server: server,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			cluster: {
				Cluster:   cluster,
				Namespace: namespace,
				AuthInfo:  cluster,
			},
		},
		CurrentContext: cluster,
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			cluster: authInfo,
		},
	}
}

func main() {
	o, err := gatherOptions()
	if err != nil {
		logrus.WithError(err).Fatal("failed to gather options")
	}
	if err := validateOptions(o); err != nil {
		log.WithError(err).Fatal("invalid options")
	}

	level, _ := log.ParseLevel(o.logLevel)
	log.SetLevel(level)
	logrusutil.ComponentInit()

	if err := clientcmd.WriteToFile(createKubeconfig(o.server, o.cluster, o.namespace, o.token, o.tokenFile), o.kubeconfig); err != nil {
		log.WithField("kubeconfig", o.kubeconfig).WithError(err).Fatal("failed to write to the kubeconfig file")
	}
	if o.tokenFile != "" && o.token != "" {
		path := o.tokenFile
		if strings.IndexByte(o.tokenFile, filepath.Separator) == -1 {
			path = filepath.Join(filepath.Dir(o.kubeconfig), o.tokenFile)
		}
		if err := ioutil.WriteFile(path, []byte(o.token), 0600); err != nil {
			log.WithField("tokenFile", o.tokenFile).WithError(err).Fatal("failed to write to the token file")
		}
	}
}
