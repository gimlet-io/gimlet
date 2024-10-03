import { useRef, useState, useEffect } from 'react';
import HelmUI from "helm-react-ui";
import ReactDiffViewer from "react-diff-viewer-continued";
import yaml from "js-yaml";
import posthog from "posthog-js"
import ImageWidget from "./imageWidget";
import SealedSecretWidget from "./sealedSecretWidget";
import * as Diff from "diff";
import { Generaltab, templateIdentity } from './generalTab';
import { Modal } from '../../components/modal'
import { ArrowTopRightOnSquareIcon, FolderIcon } from '@heroicons/react/24/solid';
import IngressWidget from "../envConfig/ingressWidget";
import { useNavigate, useLocation, useParams } from 'react-router-dom'
import { toast } from 'react-toastify';
import { InProgress, Success, Error } from '../../popUpWindow';

export function EnvConfig(props) {
  const { store, gimletClient } = props
  const { owner, repo, env, config, action } = useParams();
  let navigate = useNavigate()
  let location = useLocation()
  const preview = action === "new-preview" || action === "edit-preview"
  const repoName = `${owner}/${repo}`;

  const reduxState = props.store.getState();
  const [scmUrl, setScmUrl] = useState(reduxState.settings.scmUrl)
  const [fileInfos, setFileInfos] = useState(reduxState.fileInfos)
  const [configFile, setConfigFile] = useState()
  const [savedConfigFile, setSavedConfigFile] = useState()
  const [templates, setTemplates] = useState()
  const [selectedTemplate, setSelectedTemplate] = useState()
  const [patchedTemplate, setPatchedTemplate] = useState()
  const [errors, setErrors] = useState()
  const [navigation, setNavigation] = useState([])
  const [showModal, setShowModal] = useState(false)

  const [stackConfigDerivedValues, setStackConfigDerivedValues] = useState()
  const [templateLoadError, setTemplateLoadError] = useState(false)

  const progressToastId = useRef(null);

  store.subscribe(() => {
    const reduxState = store.getState()
    setScmUrl(reduxState.settings.scmUrl)
  })

  useEffect(() => {
    gimletClient.getStackConfig(env)
      .then(data => {
        setStackConfigDerivedValues({
          "registries": configuredRegistries(data.stackConfig, data.stackDefinition),
          "preferredDomain": extractPreferredDomain(data.stackConfig, data.stackDefinition),
          "ingressAnnotations": extractIngressAnnotations(data.stackConfig, data.stackDefinition),
        }
        )
      }, () => {/* Generic error handler deals with it */ });

    gimletClient.getRepoMetas(owner, repo)
      .then(data => {
        setFileInfos(data.fileInfos)
      }, () => {/* Generic error handler deals with it */ });

    if (action === "new-preview") {
      gimletClient.getDefaultDeploymentTemplates()
        .then(data => {
          setTemplates(data)
          setSelectedTemplate(data[0])
        }, () => {/* Generic error handler deals with it */ });
    } else {
      gimletClient.getDefaultDeploymentTemplates()
        .then(defaultTemplates => {
          gimletClient.getDeploymentTemplates(owner, repo, env, encodeURIComponent(config))
            .then(appTemplate => {
              const templates = [...defaultTemplates]
              const existingTemplate = defaultTemplates.find(d => templateIdentity(d) === templateIdentity(appTemplate[0]))
              if (!existingTemplate) {
                templates.push(appTemplate[0])
              }
              setTemplates(templates)
              setSelectedTemplate(existingTemplate ? existingTemplate : appTemplate[0])
            }, () => {
              setTemplateLoadError(true)
            });
        }, () => {/* Generic error handler deals with it */ });
      gimletClient.getEnvConfigs(owner, repo)
        .then(envConfigs => {
          if (envConfigs[env]) {
            const configFileContentFromEnvConfigs = envConfigs[env].find(c => c.app === config)
            let deepCopied = JSON.parse(JSON.stringify(configFileContentFromEnvConfigs))
            setConfigFile(configFileContentFromEnvConfigs)
            setSavedConfigFile(deepCopied)
          }
        }, () => {/* Generic error handler deals with it */
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!selectedTemplate || !stackConfigDerivedValues) {
      return
    }

    setPatchedTemplate(patchUIWidgets(selectedTemplate, stackConfigDerivedValues.registries, stackConfigDerivedValues.preferredDomain))
  }, [selectedTemplate, stackConfigDerivedValues]);

  useEffect(() => {
    if (configFile && configFile.values.ingress) {
      if (configFile.values.ingress.protectWithOauthProxy && stackConfigDerivedValues) {
        setConfigFile(prevState => ({
          ...prevState,
          values: {
            ...prevState.values,
            ingress: {
              ...prevState.values.ingress,
              annotations: {
                ...prevState.values.ingress.annotations,
                "nginx.ingress.kubernetes.io/auth-url": "https://auth" + stackConfigDerivedValues.preferredDomain + "/oauth2/auth",
                "nginx.ingress.kubernetes.io/auth-signin": "https://auth" + stackConfigDerivedValues.preferredDomain + "/oauth2/start?rd=/redirect/$http_host$escaped_request_uri",
              }
            }
          },
        }))
      }

      if (!configFile.values.ingress.protectWithOauthProxy && configFile.values.ingress.annotations) {
        let copiedConfigFile = Object.assign({}, configFile)
        delete copiedConfigFile.values.ingress.annotations["nginx.ingress.kubernetes.io/auth-url"]
        delete copiedConfigFile.values.ingress.annotations["nginx.ingress.kubernetes.io/auth-signin"]
        setConfigFile(copiedConfigFile)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [configFile && configFile.values.ingress && configFile.values.ingress.protectWithOauthProxy, stackConfigDerivedValues])

  const setAppName = (appName) => {
    setConfigFile(prevState => ({
      ...prevState,
      app: appName,
    }))
  }

  const setNamespace = (namespace) => {
    setConfigFile(prevState => ({
      ...prevState,
      namespace: namespace,
    }))
  }

  const setDeployFilter = (filter) => {
    setConfigFile(prevState => {
      if (prevState.deploy.event === "tag") {
        return {
          ...prevState,
          deploy: {
            ...prevState.deploy,
            tag: filter
          },
        }
      }

      return {
        ...prevState,
        deploy: {
          ...prevState.deploy,
          branch: filter
        },
      }
    });
  }

  const setDeployEvent = (deployEvent) => {
    setConfigFile(prevState => ({
      ...prevState,
      deploy: {
        event: deployEvent
      },
    }));
  }

  const toggleDeployPolicy = () => {
    setConfigFile(prevState => ({
      ...prevState,
      deploy: prevState.deploy ? undefined : { event: "push" },
    }));
  }

  const validationCallback = (errors) => {
    if (errors) {
      errors.forEach(e => toast.warn(e.message), { autoClose: 7000, className: 'font-xs font-mono' })
      setErrors(errors);
    } else {
      setErrors(undefined);
    }
  }

  const setValues = (values, nonDefaultValues) => {
    nonDefaultValues = handlePullSecret(nonDefaultValues)
    setConfigFile(prevState => ({
      ...prevState,
      values: nonDefaultValues,
    }));
  }

  const save = () => {
    if (errors) {
      errors.forEach(e => toast.warn(e.message), {
        className: "bg-gray-50 shadow-lg p-2",
        autoClose: 7000,
      })
      return
    }

    progressToastId.current = toast(<InProgress header="Saving..."/>, { autoClose: false });

    gimletClient.saveEnvConfig(owner, repo, env, encodeURIComponent(config), configFile)
      .then((data) => {
        toast.update(progressToastId.current, {
          render: <Success header="Configuration Saved" message={<div className='pb-4'>Deploy it on <span className='underline cursor-pointer' onClick={()=>navigate(`/repo/${owner}/${repo}/commits`)}>Commits view</span></div>} link={data.link}/>,
          className: "bg-green-50 shadow-lg p-2",
          bodyClassName: "p-2",
        });

        if (preview) {
          navigate(`/repo/${repoName}/previews`);
        } else {
          navigate(`/repo/${repoName}`);
        }
        window.scrollTo({ top: 0, left: 0 });
      }, err => {
        toast.update(progressToastId.current, {
          render: <Error header="Error" message={err.data?.message ?? err.statusText}/>,
          className: "bg-red-50 shadow-lg p-2",
          bodyClassName: "p-2",
          progressClassName: "!bg-red-200",
          autoClose: 5000
        });
      })
  }

  const deleteApp = () => {
    progressToastId.current = toast(<InProgress header="Deleting..."/>, { autoClose: false });

    gimletClient.deleteEnvConfig(owner, repo, env, config)
      .then((data) => {
        toast.update(progressToastId.current, {
          render: <Success header="Configuration deleted" link={data.link}/>,
          className: "bg-green-50 shadow-lg p-2",
          bodyClassName: "p-2",
        });

        navigate(`/repo/${repoName}`);
        window.scrollTo({ top: 0, left: 0 });
      }, err => {
        toast.update(progressToastId.current, {
          render: <Error header="Error" message={err.data?.message ?? err.statusText}/>,
          className: "bg-red-50 shadow-lg p-2",
          bodyClassName: "p-2",
          progressClassName: "!bg-red-200",
          autoClose: 5000
        });
      })
  }

  useEffect(() => {
    if (!selectedTemplate || !stackConfigDerivedValues) {
      return
    }

    if (action === "new-preview") {
      setConfigFile(
        newConfig(
          configFile ? configFile.app : config,
          configFile ? configFile.namespace : "default",
          env,
          selectedTemplate.reference.chart,
          repoName,
          stackConfigDerivedValues.preferredDomain,
          scmUrl,
          stackConfigDerivedValues.ingressAnnotations,
          true
        )
      )
      setSavedConfigFile({})
    }

    setNavigation(translateToNavigation(selectedTemplate))
  }, [selectedTemplate, stackConfigDerivedValues]);


  const setDeploymentTemplate = (template) => {
    setSelectedTemplate(template)
  }

  if (!configFile || !savedConfigFile) {
    return <SkeletonLoader preview={preview} />;
  }

  const fileInfo = fileInfos.find(f => f.envName === configFile.env && f.appName === configFile.app)

  if (!patchedTemplate) {
    if (templateLoadError) {
      return <TemplateLoadError preview={preview} configFile={configFile} fileInfo={fileInfo} scmUrl={scmUrl} repoName={repoName} />
    } else {
      return <SkeletonLoader preview={preview} />;
    }
  }

  const customFields = {
    imageWidget: ImageWidget,
    sealedSecretWidget: (props) => <SealedSecretWidget
      {...props}
      gimletClient={gimletClient}
      store={store}
      env={env}
    />,
    ingressWidget: IngressWidget
  }

  const configFileString = JSON.stringify(configFile)
  const savedConfigFileString = JSON.stringify(savedConfigFile)
  const hasChange = configFileString !== savedConfigFileString

  const diffStat = Diff.diffChars(savedConfigFileString, configFileString);
  const addedStat = diffStat.find(stat => stat.added)?.count
  const removedStat = diffStat.find(stat => stat.removed)?.count
  const addedLines = addedStat ? addedStat : 0
  const removedLines = removedStat ? removedStat : 0

  let selectedNavigation = navigation.find(i => location.pathname.endsWith(i.href))
  if (!selectedNavigation) {
    selectedNavigation = navigation[0]
  }

  const canSave = hasChange && configFile.namespace && configFile.app

  return (
    <>
      {showModal &&
        <Modal closeHandler={() => setShowModal(false)}>
          <ReactDiffViewer
            oldValue={yaml.dump(savedConfigFile)}
            newValue={yaml.dump(configFile)}
            splitView={false}
            showDiffOnly={false}
            useDarkTheme={document.documentElement.classList.contains('dark')}
            styles={{
              diffContainer: {
                overflowX: "auto",
                display: "block",
                height: "100%",
                "& pre": { whiteSpace: "pre" }
              },
            }} />
        </Modal>
      }
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow">{preview ? 'Preview Config' : 'Deployment Config'}</h1>
          {hasChange &&
            <span className="mr-8 text-sm bg-neutral-300 dark:bg-neutral-600 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-600 dark:text-neutral-300 ml-2 px-1 rounded-md cursor-pointer"
              onClick={() => setShowModal(true)}
            >
              <span>Review changes (</span>
              <span className="font-mono text-teal-500">+{addedLines}</span>
              <span className="font-mono ml-1 text-red-500">-{removedLines}</span>
              <span>)</span>
            </span>
          }
          <button
            type="button"
            disabled={!canSave}
            className={`${canSave ? 'primaryButton' : 'primaryButtonDisabled'} px-4`}
            onClick={() => {
              posthog?.capture('Env config save pushed')
              save()
            }}
          >
            Save
          </button>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-2 pb-4 flex items-center">
          {fileInfo &&
            <a className="externalLink flex space-x-1 text-sm font-mono font-thin items-center text-neutral-500" href={`${scmUrl}/${repoName}/blob/${fileInfo.branch}/.gimlet/${encodeURIComponent(fileInfo.fileName)}`} target="_blank" rel="noopener noreferrer">
              <FolderIcon className="externalLinkIcon" aria-hidden="true" />
              <span>{`.gimlet/${fileInfo.fileName}`}</span>
              <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" />
            </a>
          }
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-56">
        <div className="sticky top-0 h-96 top-56">
          <SideBar
            navigation={navigation}
            selected={selectedNavigation}
          />
        </div>
        <div className="w-full ml-14">
          {(!selectedNavigation || selectedNavigation?.name === "General") &&
            <Generaltab
              config={config}
              action={action}
              configFile={configFile}
              setAppName={setAppName}
              setNamespace={setNamespace}
              deleteApp={deleteApp}
              toggleDeployPolicy={toggleDeployPolicy}
              setDeployEvent={setDeployEvent}
              setDeployFilter={setDeployFilter}
              templates={templates}
              selectedTemplate={patchedTemplate}
              setDeploymentTemplate={setDeploymentTemplate}
              preview={preview}
            />
          }
          {selectedNavigation && selectedNavigation.name !== "General" &&
            <>
              <div className='w-full card p-6 pb-8'>
                <HelmUI
                  key={`helmui-${selectedNavigation.name}`}
                  schema={patchedTemplate.schema}
                  config={[patchedTemplate.uiSchema[selectedNavigation.uiSchemaOrder]]}
                  fields={customFields}
                  values={configFile.values}
                  setValues={setValues}
                  validate={true}
                  validationCallback={validationCallback}
                />
              </div>
              {selectedNavigation.name === "Container Image" &&
                <div className='-mt-2 learnMoreBox'>
                  Learn more about <a href="https://gimlet.io/docs/deployment-settings/image-settings" className='learnMoreLink'>Container Build Settings<ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                </div>
              }
              {selectedNavigation.name === "Domain" &&
                <div className='-mt-2 learnMoreBox'>
                  Learn more about <a href="https://gimlet.io/docs/deployment-settings/dns" className='learnMoreLink'>Setting Domain Names <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                </div>
              }
              {selectedNavigation.name === "Secrets" &&
                <div className='-mt-2 learnMoreBox'>
                  Learn more about <a href="https://gimlet.io/docs/deployment-settings/secrets" className='learnMoreLink'>Encrypted Secrets <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                </div>
              }
              {selectedNavigation.name === "Resources" &&
                <div className='-mt-2 learnMoreBox'>
                  Learn more about <a href="https://gimlet.io/docs/deployment-settings/resource-usage" className='learnMoreLink'>Resource Usage <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                </div>
              }
              {selectedNavigation.name === "Volumes" &&
                <div className='-mt-2 learnMoreBox'>
                  Learn more about <a href="https://gimlet.io/docs/deployment-settings/volumes" className='learnMoreLink'>Volumes <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                </div>
              }
            </>
          }
        </div>
      </div>
    </>
  );
}

export function SideBar(props) {
  const { navigation, selected } = props;

  const location = useLocation()
  const navigate = useNavigate()

  return (
    <nav aria-label="Sidebar">
      <ul className="w-56">
        {navigation.map((item) => (
          <li key={item.name}>
            <button
              className={`${item.name === selected.name ? 'font-medium' : 'text-neutral-600 dark:text-neutral-400'} group flex w-full gap-x-3 p-2 pl-3 text-sm leading-6 rounded-md hover:bg-neutral-200 dark:hover:bg-neutral-600 font-light`}
              onClick={() => navigate(location.pathname.replace(selected.href, "") + item.href)}
            >
              {item.name}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );
};

function TemplateLoadError(props) {
  const { preview, configFile, fileInfo, scmUrl, repoName } = props
  return (
    <>
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow">{preview ? 'Preview Config' : 'Deployment Config'}</h1>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-2 pb-4 flex items-center">
          {fileInfo &&
            <a className="externalLink flex space-x-1 text-sm font-mono font-thin items-center text-neutral-500" href={`${scmUrl}/${repoName}/blob/main/.gimlet/${encodeURIComponent(fileInfo.fileName)}`} target="_blank" rel="noopener noreferrer">
              <FolderIcon className="externalLinkIcon" aria-hidden="true" />
              <span>{`.gimlet/${fileInfo.fileName}`}</span>
              <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" />
            </a>
          }
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-56">
        <div className='w-full card p-4 mt-4'>
          <div className='items-center border-dashed border border-neutral-200 dark:border-neutral-700 rounded-md p-4 py-16'>
            <h3 className="mt-2 text-sm font-semibold text-center text-red-500">Template Load Error</h3>
            <p className="mt-1 text-sm text-neutral-500 text-center">The used custom deployment <pre>{configFile.chart.name}</pre> template is not following conventions that Gimlet requires.</p>
            <p className="mt-1 text-sm text-neutral-500 text-center underline"><a href="https://gimlet.io/docs/deployment-settings/custom-template" target="_blank" rel="noreferrer" className='externalLink'>Learn more about template conventions<ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a></p>
          </div>
        </div>
      </div>
    </>
  )
}

function SkeletonLoader(props) {
  const { preview } = props
  return (
    <>
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-12 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow py-0.5">{preview ? 'Preview Config' : 'Deployment Config'}</h1>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-56 animate-pulse">
        <div className="sticky h-96 top-56">
          <div className="w-56 p-4 pl-3 space-y-6">
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
          </div>
        </div>
        <div className="w-full ml-14">
          <div role="status" className="flex items-center justify-center h-96 bg-neutral-300 dark:bg-neutral-500 rounded-lg">
            <span className="sr-only">Loading...</span>
          </div>
        </div>
      </div>
    </>
  )
}

export const configuredRegistries = (stackConfig, stackDefinition) => {
  const config = stackConfig.config;
  const registryComponents = stackDefinition.components.filter(c => c.category === "registry")
  const configuredRegistries = registryComponents
    .filter(c => Object.keys(config).includes(c.variable) && config[c.variable])
  const decoratedConfiguredRegistries = configuredRegistries.map(r => {
    const schema = typeof r.schema === 'object'
      ? r.schema
      : JSON.parse(r.schema)
    return {
      "name": config[r.variable].displayName ?? r.name,
      "logo": r.logo,
      "variable": r.variable,
      "login": config[r.variable].credentials?.login,
      "url": config[r.variable].credentials?.url ?? schema.properties.credentials?.properties.url?.default,
    }
  })
  decoratedConfiguredRegistries.unshift({ name: "Public", variable: "public" })
  return decoratedConfiguredRegistries
}

export const extractPreferredDomain = (stackConfig, stackDefinition) => {
  const config = stackConfig.config;
  const registryComponents = stackDefinition.components.filter(c => c.category === "ingress")
  const configuredIngresses = registryComponents
    .filter(c => Object.keys(config).includes(c.variable) && config[c.variable])
  if (configuredIngresses.length > 0) {
    return config[configuredIngresses[0].variable].host
  } else {
    return ""
  }
}

export const extractIngressAnnotations = (stackConfig, stackDefinition) => {
  const config = stackConfig.config;
  const registryComponents = stackDefinition.components.filter(c => c.category === "ingress")
  const configuredIngresses = registryComponents
    .filter(c => Object.keys(config).includes(c.variable) && config[c.variable] && config[c.variable].enabled)

  if (configuredIngresses.length > 0) {
    const definition = configuredIngresses[0]
    const values = config[configuredIngresses[0].variable]

    if (values.ingressAnnotations && Object.keys(values.ingressAnnotations).length > 0) {
      return values.ingressAnnotations
    } else if (definition.variable === "nginx") {
      return {
        "cert-manager.io/cluster-issuer": "letsencrypt",
        "kubernetes.io/ingress.class": "nginx"
      }
    } else {
      return {}
    }
  } else {
    return {}
  }
}

function translateToNavigation(template) {
  const navigation = template.uiSchema.map((elem, idx) => ({ name: elem.metaData.name, href: ref(elem.metaData.name), uiSchemaOrder: idx }))
  navigation.unshift({ name: "General", href: "/general" })
  return navigation
}

function ref(name) {
  return "/" + name.replaceAll(" ", "-").toLowerCase()
}

export function robustName(str) {
  var regex = /[^a-zA-Z0-9_]/g;
  var replacedStr = str.replace(regex, '-');
  replacedStr = replacedStr.endsWith("-") ? replacedStr.slice(0, -1) : replacedStr
  replacedStr = replacedStr.length > 63 ? replacedStr.slice(0, 63) : replacedStr
  return replacedStr.toLowerCase()
}

export function newConfig(configName, namespace, env, chartRef, repoName, preferredDomain, scmUrl, ingressAnnotations, preview) {
  const config = {
    app: configName,
    namespace: namespace,
    env: env,
    chart: chartRef,
    values: {
      gitRepository: repoName,
      gitSha: "{{ .SHA }}"
    },
  }

  const oneChart = chartRef.name.includes("onechart")
  const staticSite = chartRef.name.includes("static-site")

  if (oneChart && !staticSite) {
    config.values.image = {
      repository: "nginx",
      tag: "1.27",
      strategy: "static",
      registry: "public",
    }
    config.values.resources = {
      ignoreLimits: true,
    }
  }

  if (staticSite) {
    config.values.gitCloneUrl = `${scmUrl}/${repoName}.git`
  }

  if (preview) {
    config.preview = true
  }

  if (preferredDomain) {
    let sanitizedRepoName = robustName(repoName)
    sanitizedRepoName = sanitizedRepoName.length > 55 ? sanitizedRepoName.slice(0, 55) : sanitizedRepoName

    config.values.ingress = {
      host: `${sanitizedRepoName}${preferredDomain}`,
      tlsEnabled: true,
      annotations: {
        ...ingressAnnotations
      }
    }
  }

  return config
}

const patchUIWidgets = (chart, registries, preferredDomain) => {
  if (!chart.reference.chart.name.includes("onechart")) {
    return chart
  }

  if (!chart.uiSchema[0].uiSchema["#/properties/image"]) {
    chart.uiSchema[0].uiSchema = {
      ...chart.uiSchema[0].uiSchema,
      "#/properties/image": {
        "ui:field": "imageWidget",
        'ui:options': {
          registries: registries,
        },
      },
    }
  }

  if (chart.uiSchema[1].uiSchema["#/properties/ingress"] && preferredDomain) {
    chart.uiSchema[1].uiSchema["#/properties/ingress"] = {
      "host": {
        "ui:field": "ingressWidget",
        'ui:options': {
          preferredDomain: preferredDomain,
        }
      }
    }
  }

  if (chart.uiSchema[1].uiSchema["#/properties/ingress"]) {
    chart.uiSchema[1].uiSchema["#/properties/ingress"] = {
      ...chart.uiSchema[1].uiSchema["#/properties/ingress"],
      "nginxBasicAuth": {
        "ui:order": ["user", "password"]
      },
      "ui:order": ["host", "tlsEnabled", "protectWithOauthProxy", "nginxBasicAuth", "annotations"]
    }
  }

  if (chart.uiSchema.length >= 3) {
    chart.uiSchema[3].uiSchema = {
      ...chart.uiSchema[3].uiSchema,
      "#/properties/sealedSecrets": {
        "additionalProperties": {
          "ui:field": "sealedSecretWidget"
        }
      },
    }
  }

  if (chart.schema.properties.ingress) {
    chart.schema.properties.ingress.properties.protectWithOauthProxy = {
      "$id": "#/properties/ingress/properties/protectWithOauthProxy",
      "type": "boolean",
      "title": "Protect With OauthProxy",
      "description": "",
      "default": false
    }
  }

  return chart
}

export function handlePullSecret(nonDefaultValues) {
  switch (nonDefaultValues.image?.registry) {
    case 'dockerRegistry':
      delete nonDefaultValues.imagePullSecrets
      break
    case 'public':
      delete nonDefaultValues.imagePullSecrets
      break
    default:
      if (nonDefaultValues.image) {
        nonDefaultValues = {
          ...nonDefaultValues,
          imagePullSecrets: [`{{ .APP }}-${nonDefaultValues.image.registry?.toLowerCase()}-pullsecret`]
        }
      }
      break
  }
  return nonDefaultValues
}

export default EnvConfig;
