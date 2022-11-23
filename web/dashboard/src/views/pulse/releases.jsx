import React, { Component } from 'react';
import { format, formatDistance } from "date-fns";
import { CheckIcon, XIcon, FlagIcon } from '@heroicons/react/solid';
import { ACTION_TYPE_RELEASE_STATUSES } from '../../redux/redux';

export default class Releases extends Component {
    constructor(props) {
        super(props);

        let { env } = this.props;
        let reduxState = this.props.store.getState();
        this.state = {
            releaseStatuses: reduxState.releaseStatuses[env]
        }

        this.props.store.subscribe(() => {
            let reduxState = this.props.store.getState();

            this.setState({ releaseStatuses: reduxState.releaseStatuses[env] });
        });
    }

    componentDidMount() {
        let { env } = this.props;
        this.props.gimletClient.getReleaseStatuses(env)
            .then(data => {
                this.props.store.dispatch({
                    type: ACTION_TYPE_RELEASE_STATUSES,
                    payload: {
                        envName: env,
                        data: data,
                    }
                });
            }, () => {/* Generic error handler deals with it */
            })
    }

    render() {
        let { releaseStatuses } = this.state;

        if (!releaseStatuses) {
            return null;
        }

        return (
            <div>
                <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">{this.props.env}</h4>
                {releaseStatuses.length > 0 ?
                    <ul className="-mb-8">
                        {releaseStatuses.map((releaseStatus, releaseStatusIdx) => {
                            const exactDate = format(releaseStatus.created * 1000, 'h:mm:ss a, MMMM do yyyy')
                            const dateLabel = formatDistance(releaseStatus.created * 1000, new Date());
                            const gitopsCommit = this.props.gitopsCommitsByEnv.find(gitopsCommit => gitopsCommit.sha === releaseStatus.gitopsRef);
                            
                            let color = "bg-yellow-200";
                            let Icon = FlagIcon;
                    
                            if (gitopsCommit.status.includes("Succeeded")) {
                                color = "bg-green-500";
                                Icon = CheckIcon;
                            } else if (gitopsCommit.status.includes("Failed")) {
                                color = "bg-red-500";
                                Icon = XIcon;
                            }
                            
                            return (
                                <li key={releaseStatusIdx}>
                                    <div className="relative pb-8">
                                        {releaseStatusIdx !== releaseStatuses.length - 1 ? (
                                            <span className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true" />
                                        ) : null}
                                        <div className="relative flex space-x-3">
                                            <div>
                                                <span
                                                    className={color + ' h-8 w-8 rounded-full flex items-center justify-center ring-8 ring-white'}>
                                                    <Icon className="h-5 w-5 text-white" aria-hidden="true" />
                                                </span>
                                            </div>
                                            <div className="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
                                                <div>
                                                    <p className="text-sm text-gray-500">
                                                        {`${releaseStatus.triggeredBy} `}
                                                    </p>
                                                    <p className="ml-2 text-sm text-gray-500">
                                                        <span className="font-medium text-gray-900">
                                                            {`${releaseStatus.app} -> ${releaseStatus.env} - ${releaseStatus.version.repositoryName}`}
                                                        </span>
                                                    </p>
                                                    <p className="ml-2 text-sm text-gray-500">
                                                        <a href={`https://github.com/${releaseStatus.gitopsRepo}/commit/${releaseStatus.gitopsRef}`} className="font-medium text-gray-900">
                                                            {`${releaseStatus.gitopsRef.slice(0, 6)} - ${releaseStatus.version.message}`}
                                                        </a>
                                                    </p>
                                                </div>
                                                <div className="whitespace-nowrap text-right text-sm text-gray-500">
                                                    <p
                                                        className="ml-1"
                                                        title={exactDate}
                                                        target="_blank"
                                                        rel="noopener noreferrer">
                                                        {dateLabel} ago
                                                    </p>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </li>)
                        })}
                    </ul>
                    :
                    <p className="text-xs text-gray-800">{`There are no deploys yet in ${this.props.env}.`}</p>}
            </div>
        );
    }
};
