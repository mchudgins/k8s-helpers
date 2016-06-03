package clientConfig

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/ghodss/yaml"
	client "k8s.io/kubernetes/pkg/client/restclient"
)

type k8sClusterDetails struct {
	Insecure bool   `json:"insecure-skip-tls-verify"`
	Server   string `json:"server"`
}

type k8sCluster struct {
	Name    string            `json:"name"`
	Details k8sClusterDetails `json:"cluster"`
}

type k8sContextDetails struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	User      string `json:"user"`
}

type k8sContext struct {
	Name    string            `json:"name"`
	Context k8sContextDetails `json:"context"`
}

type k8sUserDetails struct {
	Token string `json:"token"`
}

type k8sUser struct {
	Name string         `json:"name"`
	User k8sUserDetails `json:"user"`
}

type k8sConfig struct {
	Kind           string       `json:"kind"`
	APIVersion     string       `json:"apiVersion"`
	Clusters       []k8sCluster `json:"clusters"`
	Contexts       []k8sContext `json:"contexts"`
	CurrentContext string       `json:"current-context"`
	Users          []k8sUser    `json:"users"`
}

// NewConfig returns a k8s client.Config based on the user's ${HOME}/.kube/config
func NewConfig() (*client.Config, error) {
	k8sConfig := loadKubeConfig()
	ctx, _ := k8sConfig.ActiveContext()
	cluster, _ := k8sConfig.FindCluster(ctx.Context.Cluster)
	user, _ := k8sConfig.FindUser(ctx.Context.User)

	config := &client.Config{
		Host:        cluster.Details.Server,
		BearerToken: user.User.Token,
		Insecure:    cluster.Details.Insecure,
	}

	return config, nil
}

func (c *k8sConfig) ActiveContext() (*k8sContext, error) {
	if len(c.CurrentContext) == 0 {
		return nil, fmt.Errorf("no active context")
	}

	return c.FindContext(c.CurrentContext)
}

func (c *k8sConfig) FindContext(name string) (*k8sContext, error) {
	for _, ctx := range c.Contexts {
		if name == ctx.Name {
			return &ctx, nil
		}
	}
	return nil, fmt.Errorf("unable to find context %s", name)
}

func (c *k8sConfig) FindCluster(name string) (*k8sCluster, error) {
	for _, cluster := range c.Clusters {
		if name == cluster.Name {
			return &cluster, nil
		}
	}
	return nil, fmt.Errorf("unable to find cluster %s", name)
}

func (c *k8sConfig) FindUser(name string) (*k8sUser, error) {
	for _, user := range c.Users {
		if name == user.Name {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("unable to find cluster %s", name)
}

func (c *k8sConfig) SkipTLSVerify(cluster string) (bool, error) {
	clstr, err := c.FindCluster(cluster)
	if err != nil {
		return false, err
	}

	return clstr.Details.Insecure, nil
}

func loadKubeConfig() *k8sConfig {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// load the .kube/config file
	log.Printf("home:  %s", user.HomeDir)
	cfg, err := os.Open(user.HomeDir + "/.kube/config")
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	info, err := cfg.Stat()
	if err != nil {
		log.Fatal(err)
		os.Exit(3)
	}

	buf := make([]byte, info.Size(), info.Size())
	_, err = cfg.Read(buf)
	if err != nil {
		log.Fatal(err)
		os.Exit(4)
	}

	k8sConfig := k8sConfig{}
	err = yaml.Unmarshal(buf, &k8sConfig)
	if err != nil {
		log.Fatal(err)
		os.Exit(5)
	}

	return &k8sConfig
}
