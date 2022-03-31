module github.com/gimlet-io/gimlet-cli

go 1.16

require (
	cuelang.org/go v0.4.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/MichaelMure/go-term-markdown v0.1.4
	github.com/bitnami-labs/sealed-secrets v0.13.1
	github.com/blang/semver/v4 v4.0.0
	github.com/btubbs/datetime v0.1.1
	github.com/bwmarrin/discordgo v0.23.2
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/enescakir/emoji v1.0.0
	github.com/epiclabs-io/diff3 v0.0.0-20181217103619-05282cece609
	github.com/fatih/color v1.9.0
	github.com/fluxcd/flux2 v0.24.0
	github.com/fluxcd/kustomize-controller/api v0.18.1
	github.com/fluxcd/pkg/apis/meta v0.10.2
	github.com/fluxcd/pkg/runtime v0.12.3
	github.com/fluxcd/pkg/ssh v0.3.2
	github.com/fluxcd/source-controller v0.21.2
	github.com/fluxcd/source-controller/api v0.21.2
	github.com/franela/goblin v0.0.0-20200105215937-c9ffbefa60db
	github.com/gimlet-io/go-scm v1.7.1-0.20211007095331-cab5866f4eee
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/cors v1.2.0
	github.com/go-chi/jwtauth/v5 v5.0.2
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/golang-jwt/jwt/v4 v4.4.1
	github.com/google/go-github/v37 v37.0.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/laszlocph/go-login v1.0.4-0.20200901120411-b6d05e420c8a
	github.com/lib/pq v1.10.4
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mdaverde/jsonpath v0.0.0-20180315003411-f4ae4b6f36b5
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/otiai10/copy v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/russross/meddler v1.0.1
	github.com/rvflash/elapsed v0.2.0
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20220115235240-a14260e6f8a2
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whilp/git-urls v1.0.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/net v0.0.0-20211215060638-4ddde0e984e9
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sys v0.0.0-20220318055525-2edf467146b5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	helm.sh/helm/v3 v3.7.2
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.1
	sigs.k8s.io/kustomize/api v0.10.1
	sigs.k8s.io/yaml v1.3.0
)

//replace github.com/go-git/go-git/v5 => github.com/juliens/go-git/v5 v5.4.3-0.20210820144752-1cb831023bcc
replace github.com/go-git/go-git/v5 => github.com/gimlet-io/go-git/v5 v5.2.1-0.20210917081253-a2ab483ba818
