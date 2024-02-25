import { ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/solid';
import React, { memo, Component } from 'react';
import { Summary } from "./fluxState"
import GitopsStatus from './gitopsStatus';
import DeployPanelTabs from '../deployPanel/deployPanelTabs';
import { DeployStatusTab } from '../../components/deployStatus/deployStatus';
import {
  ACTION_TYPE_OPEN_DEPLOY_PANEL,
  ACTION_TYPE_CLOSE_DEPLOY_PANEL
} from '../../redux/redux';

const defaultTabs = [
  { name: 'Gitops Status', current: true },
  { name: 'Deploy Status', current: false },
]

const deployTabOpen = [
  { name: 'Gitops Status', current: false },
  { name: 'Deploy Status', current: true },
]

const Footer = memo(class Footer extends Component {
  constructor(props) {
    super(props);
    let reduxState = this.props.store.getState();

    this.state = {
      fluxStates: reduxState.fluxState,
      fluxEvents: reduxState.fluxEvents,
      selectedTab: "Kustomizations",
      targetReference: {
        objectNs: "",
        objectName: "",
        objectKind: "",
      },
      tabs: defaultTabs,
      runningDeploys: reduxState.runningDeploys,
      scmUrl: reduxState.settings.scmUrl,
      envs: reduxState.envs,
      gitopsCommits: reduxState.gitopsCommits,
      deployPanelOpen: reduxState.deployPanelOpen,
      imageBuildLogs: reduxState.imageBuildLogs,
      runningDeployId: "",
    };

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();
      this.setState((prevState) => {
        let runningDeployId = "";
        if (reduxState.runningDeploys.length !== 0) {
          runningDeployId = reduxState.runningDeploys[0].trackingId
        }

        return {
          fluxStates: reduxState.fluxState,
          fluxEvents: reduxState.fluxEvents,
          runningDeploys: reduxState.runningDeploys,
          scmUrl: reduxState.settings.scmUrl,
          envs: reduxState.envs,
          tabs: prevState.runningDeployId !== runningDeployId ? deployTabOpen : prevState.tabs,
          imageBuildLogs: reduxState.imageBuildLogs,
          runningDeployId: runningDeployId,
          gitopsCommits: reduxState.gitopsCommits,
          deployPanelOpen: reduxState.deployPanelOpen,
        }
      })

      if (this.logsEndRef.current) {
        this.logsEndRef.current.scrollIntoView();
      }
    });

    this.handleToggle = this.handleToggle.bind(this)
    this.handleNavigationSelect = this.handleNavigationSelect.bind(this)
    this.switchTab = this.switchTab.bind(this)
    this.logsEndRef = React.createRef();
  }

  handleToggle() {
    if (this.state.deployPanelOpen) {
      this.props.store.dispatch({ type: ACTION_TYPE_CLOSE_DEPLOY_PANEL })
    } else {
      this.props.store.dispatch({ type: ACTION_TYPE_OPEN_DEPLOY_PANEL })
    }
  }

  handleNavigationSelect(selectedNav, objectNs, objectName, objectKind) {
    this.setState({
      selectedTab: selectedNav,
      targetReference: {objectNs, objectName, objectKind}
    })
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
    const { gimletClient, store } = this.props;
    const { fluxStates, fluxEvents, selectedTab, targetReference, tabs, runningDeploys, envs, scmUrl, gitopsCommits, imageBuildLogs, deployPanelOpen } = this.state;

    if (!fluxStates || Object.keys(fluxStates).length === 0) {
      return null
    }

    return (
      <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true" className={`fixed inset-x-0 bottom-0 bg-neutral-200 z-40 border-t border-neutral-300 ${deployPanelOpen ? 'h-4/5' : ''}`}>
        <div className={`flex justify-between w-full ${deployPanelOpen ? '' : 'h-full'}`}>
          <div
            className='h-auto w-full cursor-pointer px-16 py-4 flex gap-x-12'
            onClick={this.handleToggle} >
            {!deployPanelOpen &&
              <div className="grid grid-cols-3">
                {Object.keys(fluxStates).slice(0, 3).map(env => {
                  const fluxState = fluxStates[env];

                  if (!fluxState) {
                    return null
                  }

                  return (
                    <div className="w-full truncate" key={env}>
                      <p className="font-semibold text-neutral-700">{`${env.toUpperCase()}`}</p>
                      <div className="ml-2">
                        <Summary resources={fluxState.gitRepositories} label="SOURCES" />
                        <Summary resources={fluxState.kustomizations} label="KUSTOMIZATIONS" />
                        <Summary resources={fluxState.helmReleases} label="HELM-RELEASES" />
                      </div>
                    </div>
                  )
                })}
              </div>
            }
          </div>
          <div className='px-4 py-2'>
            <button
              onClick={this.handleToggle}
              type="button" className="ml-1 rounded-md hover:bg-white hover:text-black text-neutral-700 p-1">
              <span className="sr-only">{deployPanelOpen ? 'Close panel' : 'Open panel'}</span>
              {deployPanelOpen ? <ArrowDownIcon className="h-5 w-5" aria-hidden="true" /> : <ArrowUpIcon className="h-5 w-5" aria-hidden="true" />}
            </button>
          </div>
        </div>
        {deployPanelOpen &&
          <div className='no-doc-scroll h-full overscroll-contain'>
            <div className="px-6">
              {DeployPanelTabs(tabs, this.switchTab)}
            </div>
            {tabs[0].current ? <GitopsStatus fluxStates={fluxStates} fluxEvents={fluxEvents} handleNavigationSelect={this.handleNavigationSelect} selectedTab={selectedTab} gimletClient={gimletClient} store={store} targetReference={targetReference} /> : null}
            {tabs[1].current ? <DeployStatusTab runningDeploys={runningDeploys} scmUrl={scmUrl} gitopsCommits={gitopsCommits} envs={envs} imageBuildLogs={imageBuildLogs} logsEndRef={this.logsEndRef} /> : null}
          </div>
        }
      </div>
    )
  }
})

export default Footer;
