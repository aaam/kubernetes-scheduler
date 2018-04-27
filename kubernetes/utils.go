package kubernetes

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"os/user"
)

func (api KubernetesCoreV1Api) currentTLSInfo() (clientCert tls.Certificate, serverCaCert *x509.CertPool) {
	return api.clientCert, api.serverCaCert
}

// Parses the cert data and generates
func (api *KubernetesCoreV1Api) loadTLSInfo() {
	var currentContextUser string
	var currentContextCluster string
	var certData []byte
	var keyData []byte
	var caCertData []byte

	// Load current context information
	for _, context := range api.config.Contexts {
		if context.Name == api.config.CurrentContext {
			currentContextUser = context.Data.User
			currentContextCluster = context.Data.Cluster
		}
	}

	// Get cert and key data from the user
	for _, user := range api.config.Users {
		if user.Name == currentContextUser {
			certData, keyData = user.Data.ClientCertificateData, user.Data.ClientKeyData
		}
	}

	// Get CA Cert data from current cluster
	for _, cluster := range api.config.Clusters {
		if cluster.Name == currentContextCluster {
			caCertData = cluster.Data.CertificateAuthorityData
		}
	}

	certificate, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertData)

	api.clientCert = certificate
	api.serverCaCert = caCertPool
}

func (api KubernetesCoreV1Api) currentApiUrlEndpoint() string {
	for _, context := range api.config.Contexts {
		if context.Name == api.config.CurrentContext {
			for _, cluster := range api.config.Clusters {
				if cluster.Name == context.Data.Cluster {
					return cluster.Data.Server
				}
			}
		}
	}
	log.Panic("kubernetes: current API Url endpoint couldn't be determined, checkout if the configuration is correct")
	return ""
}

func getKubeConfigFileDefaultLocation() string {
	kubeConf, isSet := os.LookupEnv("KUBECONFIG")
	if isSet && kubeConf != "" {
		return kubeConf
	}

	usr, err := user.Current()
	if err != nil {
		log.Panic(err)
	}
	return usr.HomeDir + "/.kube/config"
}
