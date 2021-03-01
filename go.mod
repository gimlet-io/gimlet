module github.com/gimlet-io/gimlet-cli

go 1.16

require (
	github.com/bitnami-labs/sealed-secrets v0.13.1
	github.com/enescakir/emoji v1.0.0
	github.com/fatih/color v1.7.0
	github.com/fluxcd/flux2 v0.7.7
	github.com/fluxcd/pkg/ssh v0.0.5
	github.com/franela/goblin v0.0.0-20200105215937-c9ffbefa60db
	github.com/gimlet-io/gimletd v0.0.0-20210301134851-e3199f7eb3d1
	github.com/go-chi/chi v1.5.1
	github.com/go-chi/cors v1.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/mdaverde/jsonpath v0.0.0-20180315003411-f4ae4b6f36b5
	github.com/rvflash/elapsed v0.2.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whilp/git-urls v1.0.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// https://github.com/helm/helm/issues/9354
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
