package pull

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func Pull() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	schemaDir := path.Join(usr.HomeDir, ".ytt-lint", "schema", "k8s")

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return err
	}

	crdClientset, err := apiextensionsclientset.NewForConfig(config)

	crds, err := crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, crd := range crds.Items {
		kind := crd.Spec.Names.Kind
		group := crd.Spec.Group

		for _, version := range crd.Spec.Versions {
			if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
				fmt.Printf("%s version %s of group %s does not contain a schema and will be skipped.\n", kind, version.Name, group)
				continue
			}
			schema := version.Schema.OpenAPIV3Schema
			dirname := path.Join(schemaDir, group, version.Name)
			filename := path.Join(dirname, strings.ToLower(kind)+".json")

			err = os.MkdirAll(dirname, os.ModePerm)
			if err != nil {
				return err
			}

			data, err := json.Marshal(schema)
			if err != nil {
				return err
			}

			fmt.Printf("Writing schema for %s version %s of group %s to %s\n", kind, version.Name, group, filename)
			err = ioutil.WriteFile(filename, data, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
