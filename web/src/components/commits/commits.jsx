import { format, formatDistance } from "date-fns";
import React, { Component, useState, useEffect, useRef, Fragment } from "react";
import { Transition } from '@headlessui/react'
import DeployWidget from "../deployWidget/deployWidget";
import InfiniteScroll from "react-infinite-scroll-component";
import { Modal, SkeletonLoader } from "../modal";
import { CommitEvents } from "./commitEvents";
import { EventWidget } from "./eventWidget"

export function Commits(props) {
  const { commits, envs, connectedAgents, deployHandler, owner, repo, gimletClient, fetchNextCommitsWidgets, scmUrl, } = props

  const [isScrollButtonActive, setIsScrollButtonActive] = useState(false)
  const repoName = `${owner}/${repo}`
  const commitsRef = useRef();

  useEffect(() => {
    window.addEventListener('scroll', () => {
      window.scrollY > commitsRef?.current?.offsetTop ? setIsScrollButtonActive(true) : setIsScrollButtonActive(false);
    });
  }, []);

  const handleClickScroll = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth',
    });
  }

  if (!commits) {
    return null;
  }

  const envNames = envs.map(env => env["name"]);
  for (let env of envs) {
    env.isOnline = isOnline(connectedAgents, env.name)
  }

  const commitWidgets = commits.map((commit, idx, ar) =>
    <CommitWidget
      key={idx}
      owner={owner}
      repo={repo}
      repoName={repoName}
      commit={commit}
      last={idx === ar.length - 1}
      idx={idx}
      commitsRef={commitsRef}
      envNames={envNames}
      scmUrl={scmUrl}
      connectedAgents={connectedAgents}
      deployHandler={deployHandler}
      gimletClient={gimletClient}
      envs={envs}
    />
  )

  return (
    <div className="flow-root">
      <InfiniteScroll
        dataLength={commitWidgets.length}
        next={fetchNextCommitsWidgets}
        style={{ overflow: 'visible' }}
        hasMore={true}
      >
        <ul className="-mb-4">
          {commitWidgets}
        </ul>
      </InfiniteScroll>
      <Transition
        show={isScrollButtonActive}
        as={Fragment}
        enter="transform ease-out duration-300 transition"
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave="transition ease-in duration-100"
        leaveFrom="opacity-100"
        leaveTo="opacity-0"
      >
        <div className='fixed inset-10 flex items-end px-4 py-6 pointer-events-none'>
          <button onClick={handleClickScroll} className='my-8 ml-auto px-5 py-2 bg-blue-500 hover:bg-blue-400 transition ease-in-out duration-150 text-white text-sm font-bold tracking-wide rounded-full focus:outline-none pointer-events-auto'>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="2.5" stroke="currentColor" className="w-6 h-6">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 10.5L12 3m0 0l7.5 7.5M12 3v18" />
            </svg>
          </button>
        </div>
      </Transition>
    </div>
  )
}

function isOnline(onlineEnvs, singleEnv) {
  return Object.keys(onlineEnvs)
    .map(env => onlineEnvs[env])
    .some(onlineEnv => {
      return onlineEnv.name === singleEnv.name
    })
};

export const CommitWidget = (props) => {
  const { owner, repo, commit, last, idx, commitsRef, envNames, scmUrl, connectedAgents, deployHandler, gimletClient, envs } = props

  const [showModal, setShowModal] = useState(false)
  const [events, setEvents] = useState()

  const exactDate = format(commit.created_at * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(commit.created_at * 1000, new Date());

  const loadEvents = () => {
    setShowModal(true)
    gimletClient.getCommitEvents(owner, repo, commit.sha)
      .then(data => setEvents(data), () => {/* Generic error handler deals with it */});
  }

  return (
    <>
      {showModal &&
        <Modal closeHandler={() => setShowModal(false)}>
          <div className="p-4">
            {events ?
              <CommitEvents events={events} scmUrl={scmUrl} envs={envs} />
              :
              <SkeletonLoader />
            }
          </div>
        </Modal>
      }
      <li key={commit.sha}>
        {idx === 10 &&
          <div ref={commitsRef} />
        }
        <div className="relative pl-2 py-4 rounded">
          {!last &&
            <span className="absolute top-4 left-6 -ml-px h-full w-0.5 bg-neutral-200 dark:bg-neutral-500" aria-hidden="true"></span>
          }
          <div className="relative flex items-start space-x-3">
            <div className="relative">
              <img
                className={`h-8 w-8 rounded-full bg-neutral-400 flex items-center justify-center ring-4 ring-neutral-100 dark:ring-neutral-500`}
                src={`${commit.author_pic}&s=60`}
                alt={commit.author} />
            </div>
            <div className="min-w-0 flex-1">
              <div>
                <div className="text-sm">
                  <p className="font-medium text-baseline text-black dark:text-white">
                    <a
                      className="font-mono text-xs pr-1"
                      href={commit.url}
                      target="_blank"
                      rel="noopener noreferrer">
                        {commit.sha.substring(0,8)}
                    </a>
                    {commit.message}
                    <span className="commitStatus">
                      {
                        commit.status && commit.status.statuses &&
                        commit.status.statuses.map(status => (
                          <a key={status.context} href={status.targetURL} target="_blank" rel="noopener noreferrer"
                            title={status.context}>
                            <StatusIcon status={status} />
                          </a>
                        ))
                      }
                    </span>
                  </p>
                </div>
                <p className="mt-0.5 text-xs text-neutral-600 dark:text-neutral-300">
                  <a
                    href={`${scmUrl}/${commit.author}`}
                    target="_blank"
                    rel="noopener noreferrer">
                    {commit.authorName}
                  </a>
                  <span className="ml-1">committed</span>
                  <a
                    className="ml-1"
                    title={exactDate}
                    href={commit.url}
                    target="_blank"
                    rel="noopener noreferrer">
                    {dateLabel} ago
                  </a>
                </p>
                {commit.lastEvent &&
                <p className="mt-0.5 text-xs text-neutral-600 dark:text-neutral-300">
                  <EventWidget event={commit.lastEvent} />
                  <span
                    className="px-1 py-1 rounded bg-neutral-300 hover:bg-neutral-200 dark:bg-neutral-500 dark:hover:bg-neutral-600 ml-1 cursor-pointer"
                    onClick={() => loadEvents()}>
                    View build events
                  </span>
                </p>
                }
              </div>
            </div>
            <div className="space-x-2">
              {connectedAgents &&
              <ReleaseBadges
                sha={commit.sha}
                connectedAgents={connectedAgents}
              />
              }
              {deployHandler &&
              <DeployWidget
                deployTargets={filterDeployTargets(commit.deployTargets, envNames)}
                deployHandler={deployHandler}
                sha={commit.sha}
                repo={`${owner}/${repo}`}
              />
              }
            </div>
          </div>
        </div>
      </li>
    </>
  )
}

const filterDeployTargets = (deployTargets, envs) => {
  if (!deployTargets || !envs) {
    return undefined;
  }

  const filteredTargets = deployTargets.filter(deployTarget => envs.includes(deployTarget.env));

  if (filteredTargets.length === 0) {
    return undefined;
  }

  return filteredTargets;
};

export default Commits;

class StatusIcon extends Component {
  render() {
    const { status } = this.props;

    switch (status.state) {
      case 'SUCCESS':
      case 'COMPLETED':
      case 'NEUTRAL':
        return (
          <svg className="inline fill-current text-green-300 ml-1" viewBox="0 0 12 16" version="1.1" width="15"
            height="20"
            role="img"
          >
            <title>{status.context}</title>
            <path fillRule="evenodd" d="M12 5l-8 8-4-4 1.5-1.5L4 10l6.5-6.5L12 5z" />
          </svg>
        );
      case 'PENDING':
      case 'IN_PROGRESS':
      case 'QUEUED':
        return (
          <svg className="inline fill-current text-yellow-400 ml-1" viewBox="0 0 8 16" version="1.1" width="10"
            height="20"
            role="img"
          >
            <title>{status.context}</title>
            <path fillRule="evenodd" d="M0 8c0-2.2 1.8-4 4-4s4 1.8 4 4-1.8 4-4 4-4-1.8-4-4z" />
          </svg>
        );
      default:
        return (
          <svg className="inline fill-current text-red-400 ml-1" viewBox="0 0 12 16" version="1.1" width="15"
            height="20"
            role="img"
          >
            <title>{status.context}</title>
            <path fillRule="evenodd"
              d="M7.48 8l3.75 3.75-1.48 1.48L6 9.48l-3.75 3.75-1.48-1.48L4.52 8 .77 4.25l1.48-1.48L6 6.52l3.75-3.75 1.48 1.48L7.48 8z" />
          </svg>
        )
    }
  }
}

class ReleaseBadges extends Component {
  render() {
    const { sha, connectedAgents } = this.props;

    let current = [];
    for (let envName of Object.keys(connectedAgents)) {
      const env = connectedAgents[envName];
      for (let stack of env.stacks) {
        if (stack.deployment &&
          stack.deployment.sha === sha) {
          current.push({
            env: envName,
            app: stack.service.name
          })
        }
      }
    }

    let releaseBadges = current.map((release) => (
      <span key={`${release.app}-${release.env}`} className="badge">
        {release.app} on {release.env}
      </span>
    ))

    return (
      <div className="max-w-sm break-all inline-block text-sm space-x-2">
        {releaseBadges}
      </div>
    )
  }
}
