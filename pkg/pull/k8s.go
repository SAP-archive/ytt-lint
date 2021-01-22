package pull

import (
	"fmt"

	"github.com/SAP/ytt-lint/pkg/importer"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func Pull(kubeconfigPath, contextName string) error {
	importer, err := importer.NewImporter()
	if err != nil {
		return err
	}

	var kubeConfig clientcmd.ClientConfig
	if kubeconfigPath != "" {
		c, err := clientcmd.LoadFromFile(kubeconfigPath)
		if err != nil {
			return fmt.Errorf("clientcmd.LoadFromFile(%s): %w", kubeconfigPath, err)
		}
		kubeConfig = clientcmd.NewNonInteractiveClientConfig(*c, contextName, &clientcmd.ConfigOverrides{}, nil)
	} else {
		kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	}
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("kubeConfig.ClientConfig(): %w", err)
	}

	crdClientset, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("apiextensionsclientset.NewForConfig(): %w", err)
	}

	crds, err := crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(): %w", err)
		}
		fmt.Println("Error for ApiextensionsV1. Falling back to ApiextensionsV1beta1.")
		crds, err := crdClientset.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(): %w", err)
		}

		for _, crd := range crds.Items {
			err = importer.ImportV1Beta1(crd)
			if err != nil {
				return err
			}
		}

	} else {
		for _, crd := range crds.Items {
			err = importer.ImportV1(crd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
