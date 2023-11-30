import React, {Component} from 'react'
import { XIcon } from '@heroicons/react/outline'
import { formatDistance } from "date-fns";
import { ACTION_TYPE_OPEN_DEPLOY_PANEL, ACTION_TYPE_CLOSE_DEPLOY_PANEL } from '../../redux/redux';
import DeployPanelTabs from './deployPanelTabs';
import { DeployStatus, deployHeader, Loading, ImageBuild } from "../../components/deployStatus/deployStatus";

const defaultTabs = [
  { name: 'Gitops Status', current: true },
  { name: 'Deploy Status', current: false },
]

const deployTabOpen = [
  { name: 'Gitops Status', current: false },
  { name: 'Deploy Status', current: true },
]

export default class DeployPanel extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      deployPanelOpen: reduxState.deployPanelOpen,
      gitopsCommits: reduxState.gitopsCommits,
      envs: reduxState.envs,
      connectedAgents: reduxState.connectedAgents,
      tabs: defaultTabs,
      runningDeploys: reduxState.runningDeploys,
      scmUrl: reduxState.settings.scmUrl,
      imageBuildLogs: reduxState.imageBuildLogs,
      runningDeployId: "",
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();
      this.setState((prevState) => {
        let runningDeployId = "";
        if (reduxState.runningDeploys.length !== 0) {
          runningDeployId = reduxState.runningDeploys[0].trackingId
        }

        return {
          deployPanelOpen: reduxState.deployPanelOpen,
          gitopsCommits: reduxState.gitopsCommits,
          envs: reduxState.envs,
          connectedAgents: reduxState.connectedAgents,
          runningDeploys: reduxState.runningDeploys,
          scmUrl: reduxState.settings.scmUrl,
          tabs: prevState.runningDeployId !== runningDeployId ? deployTabOpen : prevState.tabs,
          imageBuildLogs: reduxState.imageBuildLogs,
          runningDeployId: runningDeployId
        }
      });

      if (this.logsEndRef.current) {
        this.logsEndRef.current.scrollIntoView();
      }
    });

    this.switchTab = this.switchTab.bind(this)
    this.logsEndRef = React.createRef();
  }

  renderLastCommitStatusMessage(lastCommitStatus, lastCommitStatusMessage) {
    if (lastCommitStatus === "Apply failed" || lastCommitStatus === "Trailing") {
        return (
            <p className="truncate">
                {lastCommitStatusMessage}
            </p>);
    }
  }

  renderGitopsCommit(gitopsCommit, navigationHistory) {
      if (gitopsCommit === undefined) {
          return null
      }

      if (gitopsCommit.sha === undefined) {
          return null
      }

      const dateLabel = formatDistance(gitopsCommit.created * 1000, new Date());
      let color = "bg-yellow-400";
      let lastCommitStatus = "Trailing";
      let lastCommitStatusMessage = "Flux is trailing";

      if (gitopsCommit.status.includes("NotReady")) {
          lastCommitStatus = "Applying";
      } else if (gitopsCommit.status.includes("Succeeded")) {
          color = "bg-green-400";
          lastCommitStatus = "Applied";
      } else if (gitopsCommit.status.includes("Failed")) {
          color = "bg-red-400";
          lastCommitStatus = "Apply failed";
          lastCommitStatusMessage = gitopsCommit.statusDesc;
      }

      return (
          <div className="w-full truncate" key={gitopsCommit.sha}>
              <p className="font-semibold">{`${gitopsCommit.env.toUpperCase()}`}</p>
              <div className="w-72 cursor-pointer truncate text-sm"
                  onClick={() => navigationHistory.push(`/env/${gitopsCommit.env}/gitops-commits`)}
                  title={gitopsCommit.statusDesc}>
                  <span>
                      <span className={(color === "bg-yellow-400" && "animate-pulse") + ` h-4 w-4 rounded-full mr-1 relative top-1 inline-block ${color}`} />
                      {lastCommitStatus}
                      <span className="ml-1">
                          {dateLabel} ago <span className="font-mono">{gitopsCommit.sha?.slice(0, 6)}</span>
                      </span>
                  </span>
                  {this.renderLastCommitStatusMessage(lastCommitStatus, lastCommitStatusMessage)}
              </div>
          </div>
      );
  }

  renderEnvState(env, state, compact) {
    if (!state || !state.fluxState) {
      return (
        <div className="w-full truncate" key={env.name}>
            <p className="font-semibold">
              {`${env.name.toUpperCase()}`}
              <span title="Disconnected">
                <svg className="text-red-400 inline fill-current ml-1" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
                  <path
                    d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                  />
                </svg>
              </span>
            </p>
            {!compact &&
            <p
              className="cursor-pointer"
              onClick={() => this.props.history.push(`/env/${env.name}`)}
            >Please connect this environment.</p>
            }
        </div>
      )
    }

    return compact ? this.compactEnvState(env, state) : this.extendedEnvState(env, state)
  }

  compactEnvState(env, state) {
    const kustomizationWidgets = state.fluxState.kustomizations.map(kustomization => {
      let name = ""
      if (kustomization.name.endsWith("-infra")){
        name = "infra"
      } else if (kustomization.name.endsWith("-apps")){
        name = "apps"
      } else {
        return null
      }

      let color = "bg-yellow-400";
      let status = "applying";
      let statusDesc = "";

      if (kustomization.status.includes("Succeeded")) {
          color = "bg-green-400";
          status = "Applied";
          statusDesc = kustomization.statusDesc.replace('main@sha1:', '').replace("Applied revision: ", "").substring(0, 8)
      } else if (kustomization.status.includes("Failed")) {
          color = "bg-red-400";
          status = "failed";
          statusDesc = kustomization.statusDesc;
      }

      const desc = kustomization.statusDesc.replace('main@sha1:', '')
      const title = kustomization.status + " at " + new Date(kustomization.lastTransitionTime*1000) + "\n" + desc
      const dateLabel = formatDistance(kustomization.lastTransitionTime * 1000, new Date());

      return (
        <div key={kustomization.namespace + "/" + kustomization.name} title={title}>
          <p>
            <span className={(color === "bg-yellow-400" && "animate-pulse") + ` h-4 w-4 rounded-full mr-1 relative top-1 inline-block ${color}`} />
            {name}: {status} {statusDesc.substring(0, 27)}{statusDesc.length>27 ? "..." : ""} {dateLabel} ago
          </p>
        </div>
      )
    });

    return (
      <div className="w-full truncate" key={env.name}>
          <p className="font-semibold">{`${env.name.toUpperCase()}`}</p>
          <div className="ml-2">
            <div>{kustomizationWidgets}</div>
          </div>
      </div>
    );
  }

  extendedEnvState(env, state) {
    const gitrepositoryWidgets = state.fluxState.gitRepositories.map(repository => {
      let color = "text-yellow-400";

      if (repository.status.includes("Succeeded")) {
          color = "text-green-400";
      } else if (repository.status.includes("Failed")) {
          color = "text-red-400";
      }

      // const desc = repository.statusDesc.replace('main@sha1:', '')
      const title = repository.status + " at " + new Date(repository.lastTransitionTime*1000) + "\n"// + desc
      const dateLabel = formatDistance(repository.lastTransitionTime * 1000, new Date());
      const nameAndNamespace = repository.namespace + "/" + repository.name;

      return (
        <div key={nameAndNamespace} title={title}>
          <p>
              <span>
                <svg className={color + " inline fill-current h-5 w-5 mr-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
                  <path
                    d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                  />
                </svg>
              </span>
            <span className="font-bold">{nameAndNamespace}</span>: "{repository.statusDesc}" {dateLabel} ago
          </p>
        </div>
      )
    });

    const kustomizationWidgets = state.fluxState.kustomizations.map(kustomization => {
      let color = "bg-yellow-400";

      if (kustomization.status.includes("Succeeded")) {
          color = "bg-green-400";
      } else if (kustomization.status.includes("Failed")) {
          color = "bg-red-400";
      }

      const desc = kustomization.statusDesc.replace('main@sha1:', '')
      const title = kustomization.status + " at " + new Date(kustomization.lastTransitionTime*1000) + "\n" + desc
      const dateLabel = formatDistance(kustomization.lastTransitionTime * 1000, new Date());
      const nameAndNamespace = kustomization.namespace + "/" + kustomization.name;

      return (
        <div key={nameAndNamespace} title={title}>
          <p>
            <span className={(color === "bg-yellow-400" && "animate-pulse") + ` h-4 w-4 rounded-full mr-1 relative top-1 inline-block ${color}`} />
            <span className="font-bold">{nameAndNamespace}</span>: "{kustomization.statusDesc}" {dateLabel} ago
          </p>
        </div>
      )
    });

    const helmReleasesWidgets = !state.fluxState.helmReleases ? [] : state.fluxState.helmReleases.map(helmRelease => {
      let color = "bg-yellow-400";

      if (helmRelease.status.includes("Succeeded")) {
        color = "bg-green-400";
      } else if (helmRelease.status.includes("Failed")) {
        color = "bg-red-400";
      }

      const title = helmRelease.status + " at " + new Date(helmRelease.lastTransitionTime * 1000) + "\n" + helmRelease.statusDesc
      const dateLabel = formatDistance(helmRelease.lastTransitionTime * 1000, new Date());
      const nameAndNamespace = helmRelease.namespace + "/" + helmRelease.name;

      return (
        <div key={nameAndNamespace} title={title}>
          <p>
            <span className={(color === "bg-yellow-400" && "animate-pulse") + ` h-4 w-4 rounded-full mr-1 relative top-1 inline-block ${color}`} />
            <span className="font-bold">{nameAndNamespace}</span>: "{helmRelease.statusDesc}" {dateLabel} ago
          </p>
        </div>
      )
    });

    return (
        <div className="w-full truncate text-lg" key={env.name}>
            <p className="font-semibold">{`${env.name.toUpperCase()}`}</p>
            <div className="ml-2">
              <h3 className="mt-2 text-lg">Git Repositories:</h3>
              <div className="ml-2">{gitrepositoryWidgets}</div>
              <h3 className="mt-4 text-lg">Kustomizations:</h3>
              <div className="ml-2">{kustomizationWidgets}</div>
              <h3 className="mt-4 text-lg">HelmReleases:</h3>
              <div className="ml-2">{helmReleasesWidgets}</div>
            </div>
        </div>
    );
  }

  gitopsStatus(envs, connectedAgents, compact) {
    const envWidgets = envs.slice(0, 3).map(env => this.renderEnvState(env, connectedAgents[env.name], compact));

    return (
      <div className={compact ? "grid grid-cols-3 cursor-pointer" : "space-y-8"}
        onClick={() => this.props.store.dispatch({ type: ACTION_TYPE_OPEN_DEPLOY_PANEL })}
      >
          {envWidgets}
      </div>
    )
  }

  deployStatus(runningDeploys, scmUrl, gitopsCommits, envs, imageBuildLogs, logsEndRef){
    if (runningDeploys.length === 0) {
      return null;
    }

    const runningDeploy = runningDeploys[0];

    const loading = (
      <div className="p-2">
        <Loading/>
      </div>
    )

    let imageBuildWidget = null
    let deployStatusWidget = null

    if (runningDeploy.trackingId) {
      // console.log(runningDeploy)
      deployStatusWidget = DeployStatus(runningDeploy, scmUrl, gitopsCommits, envs)
    }
    if (runningDeploy.type === "imageBuild") {
      let trackingId = runningDeploy.trackingId
      if (runningDeploy.imageBuildTrackingId) {
        trackingId = runningDeploy.imageBuildTrackingId
      }

      imageBuildWidget = ImageBuild(imageBuildLogs[trackingId], logsEndRef);
    }

    const deployHeaderWidget = deployHeader(scmUrl, runningDeploy)

    return (
      <>
        {deployHeaderWidget}
        {imageBuildWidget}
        {deployStatusWidget}
        {deployStatusWidget == null && imageBuildWidget == null ? loading : null}
      </>
    );
  }

  switchTab(tab) {
    let gitopsStatus = true;
    let deployStatus = false;

    if (tab === "Deploy Status") {
      gitopsStatus = false;
      deployStatus = true;
    }

    this.setState({
      tabs: [
        { name: 'Gitops Status', current: gitopsStatus },
        { name: 'Deploy Status', current: deployStatus },
      ]
    });
  }

  render() {
    const {runningDeploys, envs, scmUrl, gitopsCommits, tabs, imageBuildLogs, connectedAgents } = this.state;

    if (!this.state.deployPanelOpen) {
      return (
          <div className="fixed bottom-0 left-0 bg-gray-800 z-50 w-full px-6 py-2 text-gray-100">
              {this.gitopsStatus(envs, connectedAgents, true)}
          </div>
      );
    }

    return (
      <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true">
          <div className="fixed inset-x-0 bottom-0 h-2/5 z-40 bg-gray-800 text-gray-300">
            <div className="absolute top-0 right-0 p-4">
              <button 
                onClick={() => this.props.store.dispatch({ type: ACTION_TYPE_CLOSE_DEPLOY_PANEL })}
                type="button" className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
                <span className="sr-only">Close panel</span>
                <XIcon className="h-5 w-5" aria-hidden="true"/>
              </button>
            </div>
            <div className="px-6 pt-4 bg-gray-900">
              {DeployPanelTabs(tabs, this.switchTab)}
            </div>
            <div className="pt-4 pb-24 px-6 overflow-y-scroll h-full w-full">
              {tabs[0].current ? this.gitopsStatus(envs, connectedAgents, false) : null}
              {tabs[1].current ? this.deployStatus(runningDeploys, scmUrl, gitopsCommits, envs, imageBuildLogs, this.logsEndRef) : null}
            </div>
          </div>
      </div>
    )
  }
}
