import { useState, useRef, useEffect } from 'react';
import { Generaltab }  from '../envConfig/generalTab'
import { newConfig, configuredRegistries, extractPreferredDomain, extractIngressAnnotations }  from '../envConfig/envConfig'
import HelmUI from "helm-react-ui";
import ImageWidget from "../envConfig/imageWidget";
import IngressWidget from "../envConfig/ingressWidget";
import { Controls, DeployStatusPanel } from '../repo/deployStatus';
import SimpleServiceDetail from '../../components/serviceDetail/simpleServiceDetail';
import DeployHandler from '../../deployHandler';
import Confetti from 'react-confetti'
import { Loading } from '../repo/deployStatus';
import { ACTION_TYPE_CLEAR_DEPLOY, ACTION_TYPE_POPUPWINDOWSUCCESS } from "../../redux/redux";
import { v4 as uuidv4 } from 'uuid';
import SealedSecretWidget from "../envConfig/sealedSecretWidget";

export function DeployWizzard(props) {
  const { store, gimletClient } = props
  const { owner, repo, env } = props.match.params;
  const repoName = `${owner}/${repo}`;

  const reduxState = props.store.getState();
  const [templates, setTemplates] = useState()
  const [selectedTemplate, setSelectedTemplate] = useState()
  const [patchedTemplate, setPatchedTemplate] = useState()
  const [configFile, setConfigFile] = useState()
  const [defaultConfigFile, setDefaultConfigFile] = useState()
  const [settings, setSettings] = useState(store.getState().settings);
  const [envConfigs, setEnvConfigs] = useState([])

  const [envs, setEnvs] = useState(store.getState().envs);

  const [runningDeploy, setRunningDeploy] = useState(store.getState().runningDeploy);
  const [runningImageBuild, setRunningImageBuild] = useState(store.getState().runningImageBuild);
  const [imageBuildLogs, setImageBuildLogs] = useState("");
  // eslint-disable-next-line no-unused-vars
  const [logLength, setLogLength] = useState(0); // using this to trigger rerender, array changes did not do it
  const [latestGitopsCommitStatus, setLatestGitopsCommitStatus] = useState()
  const [gitopsCommits, setGitopsCommits] = useState(store.getState().gitopsCommits);

  const [connectedAgents, setConnectedAgents] = useState(store.getState().connectedAgents);
  const [deploying, setDeploying] = useState(false)
  const [deployed, setDeployed] = useState(false)
  const [headBranch, setHeadBranch] = useState("")
  const [latestCommit, setLatestCommit] = useState()
  const [savingConfigInProgress, setSavingConfigInProgress] = useState(false)
  const [renderId, setRenderId] = useState()

  const deployHandler = new DeployHandler(owner, repo, gimletClient, store)

  const [registries, setRegistries] = useState()
  const [preferredDomain, setPreferredDomain] = useState()
  const [ingressAnnotations, setIngressAnnotations] = useState()

  store.subscribe(() => {
    setEnvs(reduxState.envs)
    setConnectedAgents(reduxState.connectedAgents)
    setRunningDeploy(reduxState.runningDeploy)
    if (reduxState.runningDeploy
        && reduxState.runningDeploy.status === "processed"
        && reduxState.runningDeploy.results
        && reduxState.runningDeploy.results.length > 0) {
      setLatestGitopsCommitStatus(reduxState.gitopsCommits.find(c => c.sha === reduxState.runningDeploy.results[0].hash)) 
    } else {
      setLatestGitopsCommitStatus()
    }
    setRunningImageBuild(reduxState.runningImageBuild)
    if (reduxState.runningImageBuild?.trackingId) {
      const logs = store.getState().imageBuildLogs[reduxState.runningImageBuild.trackingId]
      setImageBuildLogs(logs)
      setLogLength(logs?.logLines.length)
    }
    setSettings(reduxState.settings)
    setGitopsCommits(reduxState.gitopsCommits)
  });

  const app = configFile?.app
  let stack = connectedAgents?.[env]?.stacks.find(s => s.service.name === app)
  if (!stack) { // for apps we haven't deployed yet
    stack={service:{name: app}}
  }
  
  const logsEndRef = useRef();
  const topRef = useRef();
  const endRef = useRef();
  const [followLogs, setFollowLogs] = useState(true);

  useEffect(() => {
    gimletClient.getStackConfig(env)
      .then(data => {
        setRegistries(configuredRegistries(data.stackConfig, data.stackDefinition))
        setPreferredDomain(extractPreferredDomain(data.stackConfig, data.stackDefinition))
        setIngressAnnotations(extractIngressAnnotations(data.stackConfig, data.stackDefinition))
      }, () => {/* Generic error handler deals with it */ });

    gimletClient.getDefaultDeploymentTemplates()
    .then(data => {
      setTemplates(data)
      setSelectedTemplate(data[0])
    }, () => {/* Generic error handler deals with it */ });

    gimletClient.getEnvConfigs(owner, repo)
      .then(data => {
        if (data[env]) {
          setEnvConfigs(data[env])
        }
      }, () => {/* Generic error handler deals with it */ });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (followLogs) {
      logsEndRef.current && logsEndRef.current.scrollIntoView({block: "nearest", inline: "nearest"})
    }
  }, [logLength, followLogs]);

  useEffect(() => {
    if (!connectedAgents || !runningDeploy || !latestGitopsCommitStatus) {
      return
    }

    if (latestGitopsCommitStatus.status !== "ReconciliationSucceeded") {
      return
    }

    const env = runningDeploy.env
    const app = runningDeploy.app
    const stack = connectedAgents[env]?.stacks.find(s => s.service.name === app)
    const podStatuses = stack?.deployment?.pods.map(p=>p.status)

    if (!podStatuses) {
      return
    }

    if (!podStatuses.some(c=> c!=='Running') && deploying) {
      endRef.current && endRef.current.scrollIntoView({block: "nearest", inline: "nearest"})
      setDeployed(true)
      setDeploying(false)
    }
  }, [latestGitopsCommitStatus, connectedAgents]);

  useEffect(() => {
    let defaultBranch = 'main'
    gimletClient.getBranches(owner, repo)
    .then(data => {
      for (let branch of data) {
        if (branch === "master") {
          defaultBranch = "master";
        }
      }
    })
    setHeadBranch(defaultBranch)
    gimletClient.getCommits(owner, repo, defaultBranch, "head")
      .then(data => {
        setLatestCommit(data[0])
      }, () => {/* Generic error handler deals with it */
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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

  const validationCallback = (errors) => {
    if (errors) {
      console.log(errors)
    }
  }

  const setValues = (values, nonDefaultValues) => {
    if(nonDefaultValues.ingress?.host === "") {
      delete nonDefaultValues.ingress
    }

    switch (nonDefaultValues.image?.registry) {
      case 'dockerRegistry':
        delete nonDefaultValues.imagePullSecrets
        break
      case 'public':
        delete nonDefaultValues.imagePullSecrets
        break
      default:
        if (nonDefaultValues.image){
          nonDefaultValues = {
            ...nonDefaultValues,
            imagePullSecrets: [`{{ .APP }}-${nonDefaultValues.image.registry?.toLowerCase()}-pullsecret`]
          }
        }
        break
    }

    setConfigFile(prevState => ({
      ...prevState,
      values: nonDefaultValues,
    }));
  }

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

  const setDeploymentTemplate = (template) => {
    setSelectedTemplate(template)
  }

  useEffect(() => {
    if(!selectedTemplate) {
      return
    }

    setPatchedTemplate(patchUIWidgets(selectedTemplate, registries, preferredDomain))
  }, [selectedTemplate, registries, preferredDomain]);

  useEffect(() => {
    if(!selectedTemplate || !preferredDomain || !ingressAnnotations) {
      return
    }

    setConfigFile(newConfig(configFile ? configFile.app : repo, configFile ? configFile.namespace : "default", env, selectedTemplate.reference.chart, repoName, preferredDomain, settings.scmUrl, ingressAnnotations, false))
    setDefaultConfigFile({})
    setRenderId(uuidv4())
  }, [selectedTemplate, preferredDomain, ingressAnnotations]);

  // useEffect(() => {
  //   console.log(configFile)
  // }, [configFile]);

  const saveConfig = () => {
    setSavingConfigInProgress(true)
    gimletClient.saveEnvConfig(owner, repo, env, app, configFile)
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Configuration saved.",
            link: data.link
          }
        });

        setSavingConfigInProgress(false)
        props.history.push(`/repo/${repoName}`);
        window.scrollTo({ top: 0, left: 0 });
      }, err => {
        setSavingConfigInProgress(false)
      })
  }

  if (!patchedTemplate) {
    return <SkeletonLoader />;
  }

  if (!configFile || !defaultConfigFile) {
    return <SkeletonLoader />;
  }

  const onechart = patchedTemplate.reference.chart.name.includes("onechart")
  const staticSite = patchedTemplate.reference.chart.name.includes("static-site")
  const configExists = envConfigs.some(c => c.app === configFile.app)
  const invalidAppName = !configFile.app.match(/^[a-z]([a-z0-9-]{0,51}[a-z0-9])?$/)

  return (
    <div className='text-neutral-900 dark:text-neutral-200' key={renderId}>
    <div className='fixed'>
      {deployed &&
      <Confetti
        recycle={false}
        numberOfPieces={600}
        tweenDuration={15000}
        gravity={0.08}
        initialVelocityY={15}
        width={window.innerWidth}
        height={window.innerHeight}
      />
      }
    </div>
    <div className="w-full bg-white dark:bg-neutral-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 flex items-center">
        <div>
          <h1 className="text-3xl leading-tight text-medium flex-grow">You're almost done.</h1>
          <div className='font-light text-sm pt-2 pb-16'>
            Please follow the steps to configure your application and deploy it.
          </div>
        </div>
      </div>
      <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
    </div>
    <div className={`max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 flex`}>
      <div className="w-80 relative">
        <h2 className='text-lg font-medium flex items-center'>
          <span className={`inline-block h-2 w-2 rounded-full bg-neutral-900 dark:bg-neutral-100 mr-2`} />
          Configure Deployment
        </h2>
        <div className={`absolute h-full left-0 top-0 flex w-2 pt-10 justify-center`}>
          <div className="w-px bg-neutral-400" />
        </div>
      </div>
      <div className="w-full ml-14 space-y-6 pb-8">
        {!(deploying || deployed) &&
        <>
          <Generaltab
            // config={config}
            action="new"
            configFile={configFile}
            setAppName={setAppName}
            setNamespace={setNamespace}
            // deleteApp={deleteApp}
            // toggleDeployPolicy={toggleDeployPolicy}
            // setDeployEvent = {setDeployEvent}
            // setDeployFilter = {setDeployFilter}
            templates = {templates}
            selectedTemplate={patchedTemplate}
            setDeploymentTemplate={setDeploymentTemplate}
            preview={false}
            configExists={configExists}
            invalidAppName={invalidAppName}
          />
          <div className='w-full card p-6 pb-8'>
          <HelmUI
            key={`helmui-container-image`+staticSite}
            schema={patchedTemplate.schema}
            config={[patchedTemplate.uiSchema[0]]}
            fields={customFields}
            values={configFile.values}
            setValues={setValues}
            validate={true}
            validationCallback={validationCallback}
          />
          </div>
          { onechart &&
          <>
          <div className='w-full card p-6 pb-8'>
          <HelmUI
            key={`helmui-envvars`}
            schema={patchedTemplate.schema}
            config={[patchedTemplate.uiSchema[2]]}
            fields={customFields}
            values={configFile.values}
            setValues={setValues}
            validate={true}
            validationCallback={validationCallback}
          />
          </div>
          <div className='w-full card p-6 pb-8'>
          <HelmUI
            key={`helmui-sealedsecrets`}
            schema={patchedTemplate.schema}
            config={[patchedTemplate.uiSchema[3]]}
            fields={customFields}
            values={configFile.values}
            setValues={setValues}
            validate={true}
            validationCallback={validationCallback}
          />
          </div>
          </>
          }
          <div className='w-full card p-6 pb-8'>
          <HelmUI
            key={`helmui-domains`+staticSite}
            schema={patchedTemplate.schema}
            config={[patchedTemplate.uiSchema[1]]}
            fields={customFields}
            values={configFile.values}
            setValues={setValues}
            validate={true}
            validationCallback={validationCallback}
          />
          </div>
          <button
            className={`w-full ${!configExists && !invalidAppName ? 'primaryButton': 'primaryButtonDisabled'}`}
            onClick={() => {
              if (configExists || invalidAppName) {
                return
              }

              store.dispatch({ type: ACTION_TYPE_CLEAR_DEPLOY });
              setDeployed(false)
              setDeploying(true)

              let configFileToDeploy = JSON.parse(JSON.stringify(configFile))

              gimletClient.saveArtifact({
                version: {
                  repositoryName: repoName,
                  sha: latestCommit.sha,
                  created: 1243,
                  branch: headBranch,
                  authorName: latestCommit.authorName,
                  message: latestCommit.message,
                  url: latestCommit.url
                },
                environments: [configFile],
                vars: {
                  APP: configFile.app,
                  SHA: latestCommit.sha
                },
                fake: true
              }).then(data => {
                  deployHandler.deploy({env: env, app: configFile.app, artifactId: data["id"]}, latestCommit.sha, repoName)
              }, () => {/* Generic error handler deals with it */
              });
            }}
            >
            <p className='w-full text-center'>Deploy</p>
          </button>
        </>
        }
        {(deploying || deployed) &&
          <button
            type="button"
            className="w-full secondaryButton"
            onClick={() => {setDeploying(false); setDeployed(false)}}
            >
            <p className='w-full text-center'>Edit Configuration</p>
          </button>
        }
      </div>
    </div>
    <div className={`max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 flex ${!(deploying|deployed) ? "opacity-25 dark:opacity-30" : ""}`}>
      <div className="w-80 relative">
        <h2 className='text-lg font-medium flex items-center'>
          <span className={`inline-block h-2 w-2 rounded-full bg-neutral-900 dark:bg-neutral-100 mr-2`} />
          Deploy
        </h2>
        <div className={`absolute h-full left-0 top-0 flex w-2 pt-10 justify-center`}>
          <div className={`w-px ${deploying || deployed ? 'bg-neutral-400' : 'bg-neutral-200'}`} />
        </div>
      </div>
      <div className="w-full ml-14 space-y-6">
        <div className='mb-8 w-full -mt-4'>
          <Controls topRef={topRef} logsEndRef={logsEndRef} followLogs={followLogs} setFollowLogs={setFollowLogs} disabled={!(deploying||deployed)} />
          <div
            className={`overflow-y-auto overscroll-y-none flex-grow ${!(deploying || deployed) ? "min-h-[20vh] h-[20vh]" : "min-h-[50vh] h-[60vh]"} rounded-lg bg-neutral-900 dark:bg-neutral-800 text-neutral-300 font-mono text-sm p-2 mb-8`}
            onScroll={evt => {
                if ((logsEndRef.current.offsetTop-window.innerHeight-100) > evt.target.scrollTop) {
                  setFollowLogs(false)
                }
              }}
            >
            {(deploying || deployed) && runningDeploy &&
            <DeployStatusPanel
              key="deployLogs"
              runningDeploy={runningDeploy}
              runningImageBuild={runningImageBuild}
              scmUrl={settings.scmUrl}
              envs={envs}
              gitopsCommits={gitopsCommits}
              imageBuildLogs={imageBuildLogs}
              logsEndRef={logsEndRef}
              topRef={topRef}
            />
            }
          </div>
          {(deploying || deployed) && runningDeploy &&
          <div className='mb-8 w-full p-2 card'>
            <SimpleServiceDetail
              newApp={true}
              stack={stack}
              // rolloutHistory={stackRolloutHistory}
              envName={env}
              owner={owner}
              repoName={repoName}
              config={configFile}
              releaseHistorySinceDays={settings.releaseHistorySinceDays}
              gimletClient={gimletClient}
              store={store}
              scmUrl={settings.scmUrl}
              builtInEnv={envs.find(e => e.name === env).builtIn}
              // serviceAlerts={serviceAlerts}
              logsEndRef={logsEndRef}
            />
          </div>
          }
          {deployed &&
          <>
            <h3>Congratulations!</h3>
            <p>You just deployed your application.</p>

            <p className='mt-6'>One last step is to write your configuration to git.</p>
          </>
          }
        </div>
      </div>
    </div>
    <div className={`max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 flex ${!deployed && !deploying ? "opacity-25 dark:opacity-30" : ""}`}>
      <div className="w-80 relative">
        <div className={`absolute h-8 left-0 top-0 flex w-2 -mt-10 justify-center`}>
        <div className={`w-px ${deployed || deploying ? 'bg-neutral-400' : 'bg-neutral-200'}`} />
        </div>
        <h2 className='text-lg font-medium flex items-center'>
          <span className={`inline-block h-2 w-2 rounded-full bg-neutral-900 dark:bg-neutral-100 mr-2`} />
          Write Configuration to Git
        </h2>
      </div>
      <div className="w-full ml-14 space-y-6">
        <button
          onClick={saveConfig}
          disabled={!(deployed || deploying) || savingConfigInProgress}
          className={`w-full ${deployed ? "primaryButton" : deploying ? "secondaryButton" : "primaryButtonDisabled"}`}>
          <p className='w-full flex text-center justify-center'>
            {savingConfigInProgress ? <><Loading />Writing Configuration to Git</> : `Write Configuration to Git ${deploying ? " (even though the app is not deployed yet)" : ""}`}
          </p>
        </button>
      </div>
    </div>
    <p ref={endRef} />
    </div>
  )
}

export const patchUIWidgets = (chart, registries, preferredDomain) => {
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
      "#/properties/replicas": {
        "ui:widget": "hidden"
      }
    }
  }

  if (chart.uiSchema[1].uiSchema && preferredDomain) {
    chart.uiSchema[1].uiSchema["#/properties/ingress"] = {
      "host": {
        "ui:field": "ingressWidget",
        'ui:options': {
          preferredDomain: preferredDomain,
        }
      }
    }
  }

  if (chart.uiSchema[1].uiSchema) {
    chart.uiSchema[1].uiSchema["#/properties/ingress"] = {
      ...chart.uiSchema[1].uiSchema["#/properties/ingress"],
      "nginxBasicAuth": {
        "ui:widget": "hidden"
      },
      "annotations": {
        "ui:widget": "hidden"
      },
      "tlsEnabled": {
        "ui:widget": "hidden"
      },
      "ui:order": ["host", "tlsEnabled", "nginxBasicAuth", "annotations"]
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

  return chart
}

const SkeletonLoader = () => {
  return (
    <div className='text-neutral-900 dark:text-neutral-200'>
      <div className="w-full bg-white dark:bg-neutral-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 flex items-center">
          <div>
            <h1 className="text-3xl leading-tight text-medium flex-grow">You're almost done.</h1>
            <div className='font-light text-sm pt-2 pb-16'>
              Please follow the steps to configure your application and deploy it.
            </div>
          </div>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 flex animate-pulse">
        <div className="w-80 relative">
          <div className="h-4 my-2 ml-1 bg-neutral-300 dark:bg-neutral-500 rounded w-48"></div>
        </div>
        <div className="w-full ml-14 space-y-8">
          <div role="status" className="flex items-center justify-center h-32 bg-neutral-300 dark:bg-neutral-500 rounded-lg">
            <span className="sr-only">Loading...</span>
          </div>
          <div role="status" className="flex items-center justify-center h-64 bg-neutral-300 dark:bg-neutral-500 rounded-lg">
            <span className="sr-only">Loading...</span>
          </div>
        </div>
      </div>
    </div>
  );
}
