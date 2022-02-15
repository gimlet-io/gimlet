import {format, formatDistance} from "date-fns";
import React, {Component} from "react";
import DeployWidget from "../deployWidget/deployWidget";

export class Commits extends Component {
  render() {
    const {commits, envs, rolloutHistory, deployHandler, repo} = this.props;

    if (!commits) {
      return null;
    }

    const commitWidgets = [];

    commits.forEach((commit, idx, ar) => {
      const exactDate = format(commit.created_at * 1000, 'h:mm:ss a, MMMM do yyyy')
      const dateLabel = formatDistance(commit.created_at * 1000, new Date());
      let ringColor = 'ring-gray-100';

      commitWidgets.push(
        <li key={idx}>
          <div className="relative pl-2 py-4 hover:bg-gray-100 rounded">
            {idx !== ar.length - 1 &&
            <span className="absolute top-4 left-6 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true"></span>
            }
            <div className="relative flex items-start space-x-3">
              <div className="relative">
                <img
                  className={`h-8 w-8 rounded-full bg-gray-400 flex items-center justify-center ring-4 ${ringColor}`}
                  src={`${commit.author_pic}&s=60`}
                  alt={commit.author}/>
              </div>
              <div className="min-w-0 flex-1">
                <div>
                  <div className="text-sm">
                    <p href="#" className="font-semibold text-gray-800">{commit.message}
                      <span>
                        {
                          commit.status && commit.status.statuses &&
                          commit.status.statuses.map(status => (
                            <a key={status.createdAt} href={status.targetURL} target="_blank" rel="noopener noreferrer"
                               title={status.context}>
                              <StatusIcon status={status}/>
                            </a>
                          ))
                        }
                      </span>
                    </p>
                  </div>
                  <p className="mt-0.5 text-xs text-gray-800">
                    <a
                      className="font-semibold"
                      href={`https://github.com/${commit.author}`}
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
                </div>
                <div className="mt-2 text-sm text-gray-700">
                  <div className="ml-2 md:ml-4">

                  </div>
                </div>
              </div>
              <div className="pr-4">
                <ReleaseBadges
                  sha={commit.sha}
                  envs={envs}
                  rolloutHistory={rolloutHistory}
                />
                <DeployWidget
                  deployTargets={commit.deployTargets}
                  deployHandler={deployHandler}
                  sha={commit.sha}
                  repo={repo}
                />
              </div>
            </div>
          </div>
        </li>
      )
    })

    return (
      <div className="flow-root">
        <ul className="-mb-4">
          {commitWidgets}
        </ul>
      </div>
    )
  }
}

class StatusIcon extends Component {
  render() {
    const {status} = this.props;

    switch (status.state) {
      case 'SUCCESS':
      case 'COMPLETED':
        return (
          <svg className="inline fill-current text-green-400 ml-1" viewBox="0 0 12 16" version="1.1" width="15"
               height="20"
               role="img"
          >
            <title>{status.context}</title>
            <path fillRule="evenodd" d="M12 5l-8 8-4-4 1.5-1.5L4 10l6.5-6.5L12 5z"/>
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
            <path fillRule="evenodd" d="M0 8c0-2.2 1.8-4 4-4s4 1.8 4 4-1.8 4-4 4-4-1.8-4-4z"/>
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
                  d="M7.48 8l3.75 3.75-1.48 1.48L6 9.48l-3.75 3.75-1.48-1.48L4.52 8 .77 4.25l1.48-1.48L6 6.52l3.75-3.75 1.48 1.48L7.48 8z"/>
          </svg>
        )
    }
  }
}

class ReleaseBadges extends Component {
  render() {
    const {sha, envs, rolloutHistory} = this.props;

    let current = [];
    for (let envName of Object.keys(envs)) {
      const env = envs[envName];
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

    let recent = [];
    if (rolloutHistory) {
      rolloutHistory.forEach(env => {
        if (env.apps) {
          env.apps.forEach(app => {
            app.releases.forEach(release => {
              if (release.version.sha === sha) {
                const exists = recent.find(element => element.env === env.name && element.app === app.name);
                const isCurrent = current.find(element => element.env === env.name && element.app === app.name);
                if (!exists && !isCurrent) {
                  recent.push({
                    env: env.name,
                    app: app.name
                  })
                }
              }
            })
          })
        }
      })
    }

    let recentBadges = recent.map((release) => (
      <span key={`${release.app}-${release.env}`}
            className="inline-flex items-center px-2.5 py-0.5 rounded-md font-medium bg-gray-100 text-gray-800 mr-2"
      >
        was recently {release.app} on {release.env}
      </span>
    ))

    let releaseBadges = current.map((release) => (
      <span key={`${release.app}-${release.env}`}
            className="inline-flex items-center px-2.5 py-0.5 rounded-md font-medium bg-pink-100 text-pink-800 mr-2"
      >
        {release.app} on {release.env}
      </span>
    ))

    return (
      <div className="max-w-sm break-all inline-block text-sm">
        {recentBadges}
        {releaseBadges}
      </div>
    )
  }
}
