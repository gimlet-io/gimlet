import { ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/solid';
import React, { memo, Component } from 'react';
import { Summary } from "./capacitor/Summary"
import { ExpandedFooter } from './capacitor/ExpandedFooter';

const Footer = memo(class Footer extends Component {
  constructor(props) {
    super(props);
    const reduxState = this.props.store.getState();

    this.state = {
      fluxEvents: reduxState.fluxEvents,
      selectedTab: "Kustomizations",
      targetReference: {objectNs: "", objectName: "", objectKind: ""},
      selectedEnv: defaultEnvName(reduxState.connectedAgents),
      connectedAgents: reduxState.connectedAgents,
      gitopsCommits: reduxState.gitopsCommits,
      open: false,
    };

    this.props.store.subscribe(() => {
      const reduxState = this.props.store.getState();
      this.setState((prevState) => ({
          fluxEvents: reduxState.fluxEvents,
          connectedAgents: reduxState.connectedAgents,
          gitopsCommits: reduxState.gitopsCommits,
          selectedEnv: !prevState.selectedEnv ? defaultEnvName(reduxState.connectedAgents) : prevState.selectedEnv
        }))
    });

    this.handleToggle = this.handleToggle.bind(this)
    this.handleNavigationSelect = this.handleNavigationSelect.bind(this)
    this.setSelectedEnv = this.setSelectedEnv.bind(this)
  }

  handleToggle() {
    this.setState(prevState => ({
      open: !prevState.open
    }));
  }

  setSelectedEnv(selectedEnv) {
    this.setState({
      selectedEnv: selectedEnv
    })
  }

  handleNavigationSelect(selectedNav, objectNs, objectName, objectKind) {
    this.setState({
      selectedTab: selectedNav,
      targetReference: {objectNs, objectName, objectKind}
    })
  }

  render() {
    const { gimletClient, store } = this.props;
    const { connectedAgents, fluxEvents, selectedTab, targetReference, open, selectedEnv } = this.state;

    if (!connectedAgents || Object.keys(connectedAgents).length === 0) {
      return null
    }

    const fluxState = connectedAgents[selectedEnv].fluxState;
    const fluxK8sEvents = fluxEvents ? fluxEvents[selectedEnv] : [];

    if (!fluxState) {
      return null
    }

    let sources = []
    if (fluxState.ociRepositories) {
      sources.push(...fluxState.ociRepositories)
      sources.push(...fluxState.gitRepositories)
      sources.push(...fluxState.buckets)
    }
    sources = [...sources].sort((a, b) => a.metadata.name.localeCompare(b.metadata.name));

    return (
      <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true" className={`fixed inset-x-0 bottom-0 bg-neutral-200 z-40 border-t border-neutral-300 ${open ? 'h-4/5' : ''}`}>
        <div className={`flex justify-between w-full ${open ? '' : 'h-full'}`}>
          {!open &&
          <div className='h-auto w-full cursor-pointer px-16 py-4 flex gap-x-12' onClick={this.handleToggle} >
            <CollapsedFooter connectedAgents={connectedAgents} />
          </div>
          }
          {open &&
            <FooterNav connectedAgents={connectedAgents} selectedEnv={selectedEnv} setSelectedEnv={this.setSelectedEnv} />
          }
          <div className='px-4 py-2'>
            <button
              onClick={this.handleToggle}
              type="button" className="ml-1 rounded-md hover:bg-white hover:text-black text-neutral-700 p-1">
              <span className="sr-only">{open ? 'Close panel' : 'Open panel'}</span>
              {open ? <ArrowDownIcon className="h-5 w-5" aria-hidden="true" /> : <ArrowUpIcon className="h-5 w-5" aria-hidden="true" />}
            </button>
          </div>
        </div>
        {open &&
          <div className='no-doc-scroll h-full overscroll-contain'>
            <div className="w-full h-full overscroll-contain">
              <ExpandedFooter
                client={gimletClient}
                handleNavigationSelect={this.handleNavigationSelect}
                targetReference={targetReference}
                fluxState={fluxState}
                fluxEvents={fluxK8sEvents}
                sources={sources}
                selected={selectedTab}
                store={store}
              />
            </div>
          </div>
        }
      </div>
    )
  }
})

function CollapsedFooter(props) {
  const { connectedAgents } = props

  return (
    <div className="grid grid-cols-6 space-x-4">
      {Object.keys(connectedAgents).slice(0,6).map(envName => {
        const fluxState = connectedAgents[envName].fluxState;

        if (!fluxState) {
          return null
        }

        return (
          <div className="w-full truncate" key={envName}>
            <p className="font-semibold text-neutral-700">
              {`${envName.toUpperCase()}`}:
              <span className='ml-1'>
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

export function FooterNav(props) {
  const { connectedAgents, selectedEnv, setSelectedEnv } = props;

  return (
    <nav className="flex space-x-8 px-6 pt-4 mb-2" aria-label="Tabs">
      {Object.keys(connectedAgents).map((env) => (
        <span
          key={env}
          onClick={() => { setSelectedEnv(env); return false }}
          className={`${env === selectedEnv ? 'border-indigo-500' : 'border-transparent hover:border-gray-300'} whitespace-nowrap border-b-2 pb-2 px-1 text-neutral-700 font-semibold cursor-pointer`}
          aria-current={env === selectedEnv ? 'page' : undefined}
        >
          {env.toUpperCase()}
        </span>
      ))}
    </nav>
  )
}

function defaultEnvName(connectedAgents) {
  if (!connectedAgents || Object.keys(connectedAgents).length === 0) {
    return undefined
  }

  return Object.keys(connectedAgents)[0]
}

export default Footer;
