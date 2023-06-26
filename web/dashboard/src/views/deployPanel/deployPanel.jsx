import {Component} from 'react'
import { XIcon } from '@heroicons/react/outline'
import { formatDistance } from "date-fns";
import { ACTION_TYPE_OPEN_DEPLOY_PANEL, ACTION_TYPE_CLOSE_DEPLOY_PANEL } from '../../redux/redux';
import DeployPanelTabs from './deployPanelTabs';
import DeployStatus from "../../components/deployStatus/deployStatus";

const defaultTabs = [
  { name: 'Gitops Status', current: true },
  { name: 'Deploy Status', current: false },
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
      tabs: defaultTabs,
      runningDeploys: reduxState.runningDeploys,
      scmUrl: reduxState.settings.scmUrl,
      imageBuildLogs: reduxState.imageBuildLogs
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        deployPanelOpen: reduxState.deployPanelOpen,
        gitopsCommits: reduxState.gitopsCommits,
        envs: reduxState.envs,
        runningDeploys: reduxState.runningDeploys,
        scmUrl: reduxState.settings.scmUrl,
        tabs: reduxState.runningDeploys.length === 0 ? defaultTabs : [
          { name: 'Gitops Status', current: false },
          { name: 'Deploy Status', current: true },
        ],
        imageBuildLogs: reduxState.imageBuildLogs
      });
    });

    this.switchTab = this.switchTab.bind(this)
  }

  renderLastCommitStatusMessage(lastCommitStatus, lastCommitStatusMessage) {
    if (lastCommitStatus === "Apply failed" || lastCommitStatus === "Trailing") {
        return (
            <p className="truncate">
                {lastCommitStatusMessage}
            </p>);
    }
  }

  renderGitopsCommit(gitopsCommit) {
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
                  onClick={() => this.props.history.push(`/environments/${gitopsCommit.env}/gitops-commits`)}
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

  arrayWithFirstCommitOfEnvs(gitopsCommits, envs) {
      let firstCommitOfEnvs = [];

      for (let env of envs) {
          firstCommitOfEnvs.push(gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env.name)[0]);
      }

      firstCommitOfEnvs = firstCommitOfEnvs.filter(commit => commit !== undefined);

      firstCommitOfEnvs.sort((a, b) => b.created - a.created);

      return firstCommitOfEnvs;
  };

  gitopsStatus(gitopsCommits, envs) {
    if (gitopsCommits.length === 0 ||
      envs.length === 0) {
      return null;
    }

    const firstCommitOfEnvs = this.arrayWithFirstCommitOfEnvs(gitopsCommits, envs)
    if (firstCommitOfEnvs.length === 0) {
        return null;
    }

    return (
      <div className="grid grid-cols-3 left-0 cursor-pointer"
        onClick={() => this.props.store.dispatch({ type: ACTION_TYPE_OPEN_DEPLOY_PANEL })}
      >
          {firstCommitOfEnvs.slice(0, 3).map(gitopsCommit => this.renderGitopsCommit(gitopsCommit))}
      </div>
    )
  }

  deployStatus(runningDeploys, scmUrl, gitopsCommits, envs, imageBuildLogs){
    if (runningDeploys.length === 0) {
      return null;
    }

    const runningDeploy = runningDeploys[0];

    console.log(runningDeploy)

    if (runningDeploy.id) {
      return DeployStatus(runningDeploy, scmUrl, gitopsCommits, envs)
    } else {
      console.log(imageBuildLogs[runningDeploy.buildId])
      return (
        <div>aaa</div>
      );
    }
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
    const {runningDeploys, envs, scmUrl, gitopsCommits, tabs, imageBuildLogs } = this.state;

    if (!this.state.deployPanelOpen) {
      return (
          <div className="fixed bottom-0 left-0 bg-gray-800 z-50 w-full px-6 py-2 text-gray-100">
              {this.gitopsStatus(this.state.gitopsCommits, this.state.envs)}
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
            <div className="px-6 bg-gray-900">
              {DeployPanelTabs(tabs, this.switchTab)}
            </div>
            <div className="mt-4 pb-12 px-6 overflow-y-scroll h-full w-full">
              {tabs[0].current ? this.gitopsStatus(gitopsCommits, envs) : null}
              {tabs[1].current ? this.deployStatus(runningDeploys, scmUrl, gitopsCommits, envs, imageBuildLogs) : null}
            </div>
          </div>
      </div>
    )
  }
}
