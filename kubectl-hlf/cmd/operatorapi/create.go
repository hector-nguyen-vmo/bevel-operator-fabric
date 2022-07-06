package operatorapi

import (
	"context"
	"fmt"
	"github.com/kfsoftware/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	Name           string
	StorageClass   string
	Capacity       string
	NS             string
	Image          string
	Version        string
	IngressGateway string
	IngressPort    int
	Hosts          []string
	Output         bool
	TLSSecretName  string
}

func (o Options) Validate() error {
	return nil
}

type createCmd struct {
	out     io.Writer
	errOut  io.Writer
	apiOpts Options
}

func (c *createCmd) validate() error {
	return c.apiOpts.Validate()
}
func (c *createCmd) run() error {
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	hosts := []v1alpha1.IngressHost{}
	for _, host := range c.apiOpts.Hosts {
		hosts = append(hosts, v1alpha1.IngressHost{
			Paths: []v1alpha1.IngressPath{
				{
					Path:     "/",
					PathType: "Prefix",
				},
			},
			Host: host,
		})
	}

	ingress := v1alpha1.Ingress{
		Enabled:   true,
		ClassName: "istio",
		Annotations: map[string]string{
			"kubernetes.io/ingress.class": "istio",
		},
		TLS:   []v1beta1.IngressTLS{},
		Hosts: hosts,
	}
	fabricAPI := &v1alpha1.FabricOperatorAPI{
		TypeMeta: v1.TypeMeta{
			Kind:       "FabricOperatorAPI",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.apiOpts.Name,
			Namespace: c.apiOpts.NS,
		},
		Spec: v1alpha1.FabricOperatorAPISpec{
			Resources: &corev1.ResourceRequirements{
				Limits:   nil,
				Requests: nil,
			},
			Image:            c.apiOpts.Image,
			Tag:              c.apiOpts.Version,
			ImagePullPolicy:  "Always",
			Tolerations:      []corev1.Toleration{},
			Replicas:         1,
			Env:              []corev1.EnvVar{},
			ImagePullSecrets: []corev1.LocalObjectReference{},
			Affinity:         &corev1.Affinity{},
			Ingress:          ingress,
		},
		Status: v1alpha1.FabricOperatorAPIStatus{},
	}
	if c.apiOpts.Output {
		ot, err := helpers.MarshallWithoutStatus(&fabricAPI)
		if err != nil {
			return err
		}
		fmt.Println(string(ot))
	} else {
		ctx := context.Background()
		_, err = oclient.HlfV1alpha1().FabricOperatorAPIs(c.apiOpts.NS).Create(
			ctx,
			fabricAPI,
			v1.CreateOptions{},
		)
		if err != nil {
			return err
		}
		log.Infof("Operator API %s created on namespace %s", fabricAPI.Name, fabricAPI.Namespace)
	}
	return nil
}
func newCreateOperatorAPICmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := createCmd{out: out, errOut: errOut}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Operator API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run()
		},
	}
	f := cmd.Flags()
	f.StringVar(&c.apiOpts.Name, "name", "", "Name of the Operator API to create")
	f.StringVar(&c.apiOpts.Capacity, "capacity", "1Gi", "Total raw capacity of Operator API in this zone, e.g. 16Ti")
	f.StringVarP(&c.apiOpts.NS, "namespace", "n", helpers.DefaultNamespace, "Namespace scope for this request")
	f.StringVarP(&c.apiOpts.StorageClass, "storage-class", "s", helpers.DefaultStorageclass, "Storage class for this Operator API")
	f.StringVarP(&c.apiOpts.Image, "image", "", helpers.DefaultOperationsOperatorAPIImage, "Image of the Operator API")
	f.StringVarP(&c.apiOpts.Version, "version", "", helpers.DefaultOperationsOperatorAPIVersion, "Version of the Operator API")
	f.StringVarP(&c.apiOpts.TLSSecretName, "tls-secret-name", "", "", "TLS Secret for the Operator API")
	f.StringVarP(&c.apiOpts.IngressGateway, "istio-ingressgateway", "", "ingressgateway", "Istio ingress gateway name")
	f.IntVarP(&c.apiOpts.IngressPort, "istio-port", "", 443, "Istio ingress port")
	f.StringArrayVarP(&c.apiOpts.Hosts, "hosts", "", []string{}, "External hosts")
	f.BoolVarP(&c.apiOpts.Output, "output", "o", false, "Output in yaml")
	return cmd
}
