import { format, formatDistance } from "date-fns";
import { Component } from "react";
import Emoji from "react-emoji-render";
import { nanoid } from 'nanoid';
import defaultProfilePicture from "../../views/profile/defaultProfilePicture.png"

export class RolloutHistory extends Component {
  constructor(props) {
    super(props);

    this.state = {
      open: false
    }

    this.toggle = this.toggle.bind(this);
  }

  toggle() {
    this.setState(prevState => ({
      open: !prevState.open
    }));
  }

  render() {
    let { env, app, appRolloutHistory, rollback, releaseHistorySinceDays } = this.props;

    const { open } = this.state;

    if (!appRolloutHistory) {
      return null;
    }

    let previousDateLabel = ''
    const markers = [];
    const rollouts = [];

    let currentlyReleasedRef;
    let allPreviousCommitsAreRollbacks = true;

    const releasesCount = appRolloutHistory.length;
    for (let i = releasesCount-1; i >= 0; i--) {
      const rollout = appRolloutHistory[i];
      const currentlyReleased = !rollout.rolledBack && allPreviousCommitsAreRollbacks;
      if (currentlyReleased) {
        currentlyReleasedRef = rollout.gitopsRef;
        allPreviousCommitsAreRollbacks = false;
      }
    }

    appRolloutHistory.forEach((rollout, idx) => {
      const exactDate = format(rollout.created * 1000, 'h:mm:ss a, MMMM do yyyy')
      const dateLabel = formatDistance(rollout.created * 1000, new Date());

      const showDate = previousDateLabel !== dateLabel
      previousDateLabel = dateLabel;

      let color = rollout.rolledBack ? 'bg-grey-300' : 'bg-yellow-100';
      let ringColor = rollout.rolledBack ? 'ring-grey-400' : 'ring-yellow-300';
      let bgColor = rollout.rolledBack ? 'bg-grey-100' : 'bg-yellow-100';
      let hoverBgColor = rollout.rolledBack ? 'hover:bg-grey-200' : 'hover:bg-yellow-200';
      if (rollout.gitopsCommitStatus.includes("Succeeded") && !rollout.rolledBack) {
        color = "bg-green-100";
        ringColor = "ring-green-300";
        bgColor = 'bg-green-100';
        hoverBgColor = 'hover:bg-green-200';
      } else if (rollout.gitopsCommitStatus.includes("Failed") && !rollout.rolledBack) {
        color = "bg-red-300";
        ringColor = "ring-red-400";
        bgColor = 'bg-red-100';
        hoverBgColor = 'hover:bg-red-200';
      }

      let border = showDate ? 'lg:border-l' : '';

      const currentlyReleased = rollout.gitopsRef === currentlyReleasedRef

      markers.push(marker(rollout, border, color, showDate, dateLabel, exactDate, this.toggle))
      rollouts.unshift(rolloutWidget(idx, bgColor, hoverBgColor, ringColor, exactDate, dateLabel, rollback, env, app, currentlyReleased, rollout))
    })

    if (releaseHistorySinceDays && releasesCount === 0) {
      return (
        <div className="text-xs text-gray-500 py-2">
          No releases in the past {releaseHistorySinceDays} days.
        </div>)
    }

    return (
      <div className="">
        <div className="grid grid-cols-10 p-2">
          {markers}
        </div>
        {open &&
          <div className="bg-yellow-50 rounded">
            <div className="flow-root">
              <ul className="-mb-4 p-2">
                {rollouts}
              </ul>
            </div>
          </div>
        }
      </div>
    )
  }
}

function Commit(props) {
  const { version } = props;

  const exactDate = format(version.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(version.created * 1000, new Date());

  return (
    <div className="md:flex text-xs text-gray-500">
      <div className="md:flex-initial">
        <span className="font-semibold leading-none">{version.message && <Emoji text={version.message} />}</span>
        <div className="flex mt-1">
          {version.author &&
            <img
              className="rounded-sm overflow-hidden mr-1"
              src={`https://github.com/${version.author}.png?size=128`}
              alt={version.authorName}
              width="20"
              height="20"
            />
          }
          <div>
            <span className="font-semibold">{version.authorName}</span>
            <a
              className="ml-1"
              title={exactDate}
              href={`https://github.com/${version.repositoryName}/commit/${version.sha}`}
              target="_blank"
              rel="noopener noreferrer">
              comitted {dateLabel} ago
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}

function marker(rollout, border, color, showDate, dateLabel, exactDate, toggle) {
  const title = `[${rollout.version.sha.slice(0, 6)}] ${truncate(rollout.version.message)}

Deployed by ${rollout.triggeredBy}

at ${exactDate}`;

  return (
    <div key={nanoid()} className={`h-8 ${border} cursor-pointer`} title={title} onClick={() => toggle()}>
      <div className={`h-2 ml-1 md:mx-1 ${color} rounded`}></div>
      {showDate &&
        <div className="hidden lg:block mx-2 mt-2 text-xs text-gray-400">
          <span>{dateLabel} ago</span>
        </div>
      }
    </div>
  )
}

export function rolloutWidget(idx, bgColor, hoverBgColor, ringColor, exactDate, dateLabel, rollback, env, app, currentlyReleased, rollout) {
  const exactGitopsCommitCreatedDate = format(rollout.gitopsCommitCreated * 1000, 'h:mm:ss a, MMMM do yyyy')
  let gitopsCommitCreatedDateLabel = formatDistance(rollout.gitopsCommitCreated * 1000, new Date());
  gitopsCommitCreatedDateLabel = gitopsCommitCreatedDateLabel.charAt(0).toUpperCase() + gitopsCommitCreatedDateLabel.slice(1);

  return (
    <li key={rollout.gitopsRef}
      className={`${hoverBgColor} ${bgColor} p-4 rounded`}
    >
      <div className="relative pb-4">
        {idx !== 0 &&
          <span className="absolute top-8 left-4 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true"></span>
        }
        <div className="relative flex items-start space-x-3">
          <div className="relative">
            <img
              className={`h-8 w-8 rounded-full bg-gray-400 flex items-center justify-center ring-4 ${ringColor}`}
              src={`https://github.com/${rollout.triggeredBy}.png?size=128`}
              onError={(e) => { e.target.src = defaultProfilePicture }}
              alt={rollout.triggeredBy} />
          </div>
          <div className="min-w-0 flex-1">
            <div>
              <div className="text-sm">
                <p href="#" className="font-medium text-gray-900">{rollout.triggeredBy}</p>
              </div>
              <p className="mt-0.5 text-sm text-gray-500">
                <span>Released</span>
                <a
                  className="ml-1"
                  title={exactDate}
                  href={`https://github.com/${rollout.gitopsRepo}/commit/${rollout.gitopsRef}`}
                  target="_blank"
                  rel="noopener noreferrer">
                  {dateLabel} ago
                </a>
              </p>
              <div className="mt-0.5 text-sm text-gray-500">
                <span title={exactGitopsCommitCreatedDate} >
                  {gitopsCommitCreatedDateLabel} ago {!rollout.gitopsCommitStatusDesc ? "commit is not applied yet." : rollout.gitopsCommitStatusDesc}
                </span>
              </div>
            </div>
            <div className="mt-2 text-sm text-gray-700">
              <div className="ml-4 md:ml-4">
                <Commit version={rollout.version} />
              </div>
            </div>
          </div>
          {rollback &&
            <div>
              {!currentlyReleased && !rollout.rolledBack &&
                <button
                  type="button"
                  onClick={(e) => {
                    // eslint-disable-next-line no-restricted-globals
                    confirm('Are you sure you want to roll back?') &&
                      rollback(env, app, rollout.gitopsRef, e);
                  }}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Rollback to this version
                </button>
              }
              {rollout.rolledBack &&
                <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-red-100 text-red-800">
                  Rolled back
                </span>
              }
              {currentlyReleased &&
                <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-green-100 text-green-800">
                  Current version
                </span>
              }
            </div>}
        </div>
      </div>
    </li>
  )
}

function truncate(input) {
  if (!input) {
    return input
  }
  if (input.length > 30) {
    return input.substring(0, 30) + '...';
  }
  return input;
}
