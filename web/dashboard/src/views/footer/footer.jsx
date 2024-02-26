import { ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/solid';
import React, { memo, Component } from 'react';
import { Summary } from "./capacitor/Summary"
import GitopsStatus from './gitopsStatus';
import {
  ACTION_TYPE_OPEN_DEPLOY_PANEL,
  ACTION_TYPE_CLOSE_DEPLOY_PANEL
} from '../../redux/redux';

const Footer = memo(class Footer extends Component {
  constructor(props) {
    super(props);
    let reduxState = this.props.store.getState();

    this.state = {
      fluxEvents: reduxState.fluxEvents,
      selectedTab: "Kustomizations",
      targetReference: {objectNs: "", objectName: "", objectKind: ""},
      connectedAgents: reduxState.connectedAgents,
      gitopsCommits: reduxState.gitopsCommits,
      deployPanelOpen: reduxState.deployPanelOpen,
    };

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();
      this.setState({
          fluxEvents: reduxState.fluxEvents,
          connectedAgents: reduxState.connectedAgents,
          gitopsCommits: reduxState.gitopsCommits,
          deployPanelOpen: reduxState.deployPanelOpen,
        })
    });

    this.handleToggle = this.handleToggle.bind(this)
    this.handleNavigationSelect = this.handleNavigationSelect.bind(this)
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

  render() {
    const { gimletClient, store } = this.props;
    const { connectedAgents, fluxEvents, selectedTab, targetReference, deployPanelOpen } = this.state;

    if (!connectedAgents || Object.keys(connectedAgents).length === 0) {
      return null
    }

    return (
      <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true" className={`fixed inset-x-0 bottom-0 bg-neutral-200 z-40 border-t border-neutral-300 ${deployPanelOpen ? 'h-4/5' : ''}`}>
        <div className={`flex justify-between w-full ${deployPanelOpen ? '' : 'h-full'}`}>
          <div
            className='h-auto w-full cursor-pointer px-16 py-4 flex gap-x-12'
            onClick={this.handleToggle} >
            {!deployPanelOpen &&
              <CollapsedFooter connectedAgents={connectedAgents} />
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
            <GitopsStatus connectedAgents={connectedAgents} fluxEvents={fluxEvents} handleNavigationSelect={this.handleNavigationSelect} selectedTab={selectedTab} gimletClient={gimletClient} store={store} targetReference={targetReference} />
            {/* {tabs[1].current ? <DeployStatusTab runningDeploys={runningDeploys} scmUrl={scmUrl} gitopsCommits={gitopsCommits} envs={envs} imageBuildLogs={imageBuildLogs} logsEndRef={this.logsEndRef} /> : null} */}
          </div>
        }
      </div>
    )
  }
})

function CollapsedFooter(props) {
  const { connectedAgents } = props

  return (
    <div className="grid grid-cols-3">
      {Object.keys(connectedAgents).slice(0,3).map(envName => {
        const fluxState = connectedAgents[envName].fluxState;

        if (!fluxState) {
          return null
        }

        return (
          <div className="w-full truncate" key={envName}>
            <p className="font-semibold text-neutral-700">
              {`${envName.toUpperCase()}`}:
              <span className='ml-2'>
                <Summary resources={fluxState.gitRepositories} label="SOURCES" simple={true} />
                <Summary resources={fluxState.kustomizations} label="KUSTOMIZATIONS" simple={true}  />
                <Summary resources={fluxState.helmReleases} label="HELM-RELEASES" simple={true}  />
              </span>
            </p>
          </div>
        )
      })}
    </div>
  )
}

export default Footer;
