/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	notifv1 "github.com/fluxcd/notification-controller/api/v1"
	notifv1beta3 "github.com/fluxcd/notification-controller/api/v1beta3"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"

	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
)

func Generate(options Options) (*manifestgen.Manifest, error) {
	gvk := sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	gitRepository := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			Ignore: &options.GimletPath,
			URL:    options.URL,
			Interval: metav1.Duration{
				Duration: options.Interval,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: options.Branch,
			},
			SecretRef: &meta.LocalObjectReference{
				Name: options.Secret,
			},
			RecurseSubmodules: options.RecurseSubmodules,
		},
	}

	gitData, err := yaml.Marshal(gitRepository)
	if err != nil {
		return nil, err
	}

	gvk = kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	var kustomizationDependencies kustomizev1.Kustomization
	if options.GenerateDependencies {
		kustomizationDependencies = kustomizev1.Kustomization{
			TypeMeta: metav1.TypeMeta{
				Kind:       gvk.Kind,
				APIVersion: gvk.GroupVersion().String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", options.Name, "dependencies"),
				Namespace: options.Namespace,
			},
			Spec: kustomizev1.KustomizationSpec{
				Interval: metav1.Duration{
					Duration: 24 * time.Hour,
				},
				Path:  DependenciesPath(options.DependenciesPath),
				Prune: true,
				SourceRef: kustomizev1.CrossNamespaceSourceReference{
					Kind: sourcev1.GitRepositoryKind,
					Name: options.Name,
				},
			},
		}
	}

	ksDepData, err := yaml.Marshal(kustomizationDependencies)
	if err != nil {
		return nil, err
	}

	kustomization := kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{
				Duration: 24 * time.Hour,
			},
			Path:  fmt.Sprintf("./%s", strings.TrimPrefix(options.TargetPath, "./")),
			Prune: true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
				Name: options.Name,
			},
		},
	}

	if options.GenerateDependencies {
		kustomization.Spec.DependsOn = []meta.NamespacedObjectReference{
			{
				Name: fmt.Sprintf("%s-%s", options.Name, "dependencies"),
			},
		}
	}

	ksData, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf("---\n%s---\n%s", resourceToString(gitData), resourceToString(ksData))
	if options.GenerateDependencies {
		content += fmt.Sprintf("---\n%s", resourceToString(ksDepData))
	}

	return &manifestgen.Manifest{
		Path:    path.Join(options.TargetPath, options.Namespace, options.ManifestFile),
		Content: content,
	}, nil
}

func GenerateProviderAndAlert(
	envName string,
	gimletdUrl string,
	token string,
	targetPath string,
	kustomizationName string,
	notificationsName string,
	fileName string) (*manifestgen.Manifest, error) {
	namespace := "flux-system"
	gvk := notifv1beta3.GroupVersion.WithKind(notifv1beta3.ProviderKind)
	provider := notifv1beta3.Provider{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      notificationsName,
			Namespace: namespace,
		},
		Spec: notifv1beta3.ProviderSpec{
			Type:    "generic",
			Address: fmt.Sprintf("%s/api/flux-events?access_token=%s&env=%s", gimletdUrl, token, envName),
		},
	}

	gvk = notifv1beta3.GroupVersion.WithKind(notifv1beta3.AlertKind)
	kk := kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	alert := notifv1beta3.Alert{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      notificationsName,
			Namespace: namespace,
		},
		Spec: notifv1beta3.AlertSpec{
			ProviderRef: meta.LocalObjectReference{
				Name: notificationsName,
			},
			EventSeverity: "info",
			EventSources: []notifv1.CrossNamespaceObjectReference{
				{
					Kind:      kk.Kind,
					Namespace: namespace,
					Name:      kustomizationName,
				},
			},
			Suspend: false,
		},
	}

	providerData, err := yaml.Marshal(provider)
	if err != nil {
		return nil, err
	}

	alertData, err := yaml.Marshal(alert)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(targetPath, namespace, fileName),
		Content: fmt.Sprintf("%s---\n%s", resourceToString(providerData), resourceToString(alertData)),
	}, nil
}

func GenerateKustomizationForApp(
	app string,
	env string,
	kustomizationName string,
	sourceName string,
	singleEnv bool,
) (*manifestgen.Manifest, error) {
	filePath := filepath.Join(env, "flux")
	kustomizationPath := filepath.Join(env, app)
	if singleEnv {
		filePath = "flux"
		kustomizationPath = app
	}
	gvk := kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	kustomization := kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kustomizationName,
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{
				Duration: 24 * time.Hour,
			},
			Path:  fmt.Sprintf("./%s", strings.TrimPrefix(kustomizationPath, "./")),
			Prune: true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
				Name: sourceName,
			},
		},
	}

	ksData, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(filePath, fmt.Sprintf("kustomization-%s.yaml", app)),
		Content: fmt.Sprintf("---\n%s", resourceToString(ksData)),
	}, nil
}

func GenerateIngress(
	app string,
	port int32,
	namespace string,
	host string,
	targetPath string,
	httpPath string,
) (*manifestgen.Manifest, error) {
	nginx := "nginx"
	var pathType networkingv1.PathType
	pathType = "Prefix"

	ingress := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", app),
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"nginx.ingress.kubernetes.io/configuration-snippet": `sub_filter '</body>' '
	<div class="bg-transparent bottom-0 md:px-0 fixed z-[2147483647] left-0 md:left-[calc(50%-390px)]">
		<iframe class="h-48 min-h-[initial] max-h-[initial] translate-[initial] bg-transparent border-0 block w-screen md:w-[780px]"
			id="github-iframe"
			title="Gimlet Drawer"
			src=""
		>
		</iframe>
	<script src="https://cdn.tailwindcss.com"></script>
	<script>
		fetch("https://api.github.com/repos/dzsak/deploying-a-static-site-with-netlify-sample/contents/gimlet-preview.html?ref=v0.0.1-rc.19").then(function(t){return t.json()})
			.then(function(t){(iframe=document.getElementById("github-iframe")).src="data:text/html;base64,"+encodeURIComponent(t.content)});
	</script>
</body>';
proxy_set_header Accept-Encoding "";`,
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &nginx,
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: app,
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
									Path:     httpPath,
									PathType: &pathType,
								},
							},
						},
					},
				},
			},
		},
	}

	ingressData, err := yaml.Marshal(ingress)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(targetPath, namespace, fmt.Sprintf("ingress-%s.yaml", app)),
		Content: fmt.Sprintf("---\n%s", resourceToString(ingressData)),
	}, nil
}

func GenerateConfigMap(
	configMapName string,
	namespace string,
	data map[string]string,
) (*manifestgen.Manifest, error) {
	if len(data) == 0 {
		return nil, nil
	}

	immutable := false
	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data:      data,
		Immutable: &immutable,
	}

	yamlString, err := yaml.Marshal(configMap)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(".", fmt.Sprintf("configmap-%s-%s.yaml", configMapName, namespace)),
		Content: fmt.Sprintf("---\n%s", resourceToString(yamlString)),
	}, nil
}

func resourceToString(data []byte) string {
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)
	return string(data)
}

func DependenciesPath(targetPath string) string {
	if targetPath == "" {
		return fmt.Sprintf("./%sdependencies", strings.TrimPrefix(targetPath, "./"))
	} else {
		return fmt.Sprintf("./%s/dependencies", strings.TrimPrefix(targetPath, "./"))
	}
}
