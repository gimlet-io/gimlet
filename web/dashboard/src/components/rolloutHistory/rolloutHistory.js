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
    let { env, app, appRolloutHistory, rollback, releaseHistorySinceDays, scmUrl } = this.props;

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

    appRolloutHistory.forEach((rollout, idx, arr) => {
      const exactDate = format(rollout.created * 1000, 'h:mm:ss a, MMMM do yyyy')
      const dateLabel = formatDistance(rollout.created * 1000, new Date());

      const showDate = previousDateLabel !== dateLabel
      previousDateLabel = dateLabel;

      let color = rollout.rolledBack ? 'bg-grey-300' : 'bg-yellow-100';
      if (rollout.gitopsCommitStatus.includes("Succeeded") && !rollout.rolledBack) {
        color = "bg-green-100";
      } else if (rollout.gitopsCommitStatus.includes("Failed") && !rollout.rolledBack) {
        color = "bg-red-300";
      }

      let border = showDate ? 'lg:border-l' : '';

      const currentlyReleased = rollout.gitopsRef === currentlyReleasedRef

      markers.push(marker(rollout, border, color, showDate, dateLabel, exactDate, this.toggle))
      rollouts.unshift(rolloutWidget(idx, arr, exactDate, dateLabel, rollback, env, app, currentlyReleased, rollout, scmUrl))
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
          <div className="flow-root">
            <ul>
              {rollouts}
            </ul>
          </div>
        }
      </div>
    )
  }
}

function Commit(props) {
  const { version, isReleaseStatus, gitopsRepo, scmUrl } = props;

  const exactDate = format(version.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(version.created * 1000, new Date());

  return (
    <div className="md:flex text-xs">
      <div className="md:flex-initial space-y-0.5 mt-2">
      {isReleaseStatus && <div className="font-semibold text-gray-700"> <Emoji text={gitopsRepo} /></div>}
        <div className="text-gray-700">{version.message && <Emoji text={version.message} />}</div>
        <div className="flex">
          {version.author &&
            <img
              className="rounded-sm overflow-hidden mr-1"
              src={`${scmUrl}/${version.author}.png?size=128`}
              alt={version.authorName}
              width="20"
              height="20"
            />
          }
          <div className="text-gray-600">
            <span className="font-semibold">{version.authorName}</span>
            <a
              className="ml-1"
              title={exactDate}
              href={`${scmUrl}/${version.repositoryName}/commit/${version.sha}`}
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

export function rolloutWidget(idx, arr, exactDate, dateLabel, rollback, env, app, currentlyReleased, rollout, scmUrl) {
  const exactGitopsCommitCreatedDate = format(rollout.gitopsCommitCreated * 1000, 'h:mm:ss a, MMMM do yyyy')
  let gitopsCommitCreatedDateLabel = formatDistance(rollout.gitopsCommitCreated * 1000, new Date());

  let rounding = "";
  let status = rollout.gitopsCommitStatus.includes("NotReady") ? "Applying" : "Trailing";
  let ringColor = 'ring-yellow-300';
  let bgColor = 'bg-yellow-100';
  let hoverBgColor = 'hover:bg-yellow-200';

  if (rollout.rolledBack) {
    ringColor = 'ring-grey-300';
    bgColor = 'bg-grey-100';
    hoverBgColor = 'hover:bg-grey-200';
    status = "Rolled back";
  } else if (rollout.gitopsCommitStatus.includes("Succeeded")) {
    ringColor = "ring-green-300";
    bgColor = 'bg-green-100';
    hoverBgColor = 'hover:bg-green-200';
    status = "Applied";
  } else if (rollout.gitopsCommitStatus.includes("Failed")) {
    ringColor = "ring-red-300";
    bgColor = 'bg-red-100';
    hoverBgColor = 'hover:bg-red-200';
    status = "Apply failed";
  }

  if (idx === 0) {
    rounding = "rounded-b"
  } else if (idx === arr.length - 1) {
    rounding = "rounded-t"
  }

  return (
    <li key={rollout.gitopsRef}
      className={`${hoverBgColor} ${bgColor} p-4 ${rounding}`}
    >
      <div className="relative pb-4">
        {idx !== 0 &&
          <span className="absolute top-8 left-4 -ml-px h-full w-0.5 bg-gray-300" aria-hidden="true"></span>
        }
        <div className="relative flex items-start space-x-3">
          <div className="relative">
            <img
              className={`h-8 w-8 rounded-full bg-gray-400 flex items-center justify-center ring-4 ${ringColor}`}
              src={`${scmUrl}/${rollout.triggeredBy}.png?size=128`}
              onError={(e) => { e.target.src = defaultProfilePicture }}
              alt={rollout.triggeredBy} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-sm space-y-0.5">
              <p className="font-semibold text-gray-700">{rollout.triggeredBy}</p>
              <p className="text-gray-700">
                <span>Released</span>
                <a
                  className="ml-1"
                  title={exactDate}
                  href={`${scmUrl}/${rollout.gitopsRepo}/commit/${rollout.gitopsRef}`}
                  target="_blank"
                  rel="noopener noreferrer">
                  {dateLabel} ago
                </a>
              </p>
              <div className="text-gray-600">
                <span title={exactGitopsCommitCreatedDate} >
                  {status} {gitopsCommitCreatedDateLabel} ago, {!rollout.gitopsCommitStatusDesc ? "commit is not applied yet." : rollout.gitopsCommitStatusDesc}
                </span>
              </div>
            </div>
            <div className="mt-2 ml-4">
              <Commit
              isReleaseStatus={rollback === undefined}
              gitopsRepo={rollout.gitopsRepo}
              version={rollout.version}
              scmUrl={scmUrl}
              />
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
