import React, { Component } from 'react';
import { format, formatDistance } from "date-fns";
import { ACTION_TYPE_RELEASE_STATUSES } from '../../redux/redux';
import { rolloutWidget } from '../../components/rolloutHistory/rolloutHistory';

export default class Releases extends Component {
    componentDidMount() {
        let { gimletClient, store, env } = this.props;
        gimletClient.getReleaseStatuses(env, 3)
            .then(data => {
                store.dispatch({
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
        let { releaseStatuses, releaseHistorySinceDays, env } = this.props;

        if (!releaseStatuses) {
            return null;
        }

        let renderReleaseStatuses = [];

        releaseStatuses.forEach((rollout, idx) => {
            const exactDate = format(rollout.created * 1000, 'h:mm:ss a, MMMM do yyyy');
            const dateLabel = formatDistance(rollout.created * 1000, new Date());

            let ringColor = rollout.rolledBack ? 'ring-grey-400' : 'ring-yellow-300';
            let bgColor = rollout.rolledBack ? 'bg-grey-100' : 'bg-yellow-100';
            let hoverBgColor = rollout.rolledBack ? 'hover:bg-grey-200' : 'hover:bg-yellow-200';
            if (rollout.gitopsCommitStatus.includes("Succeeded") && !rollout.rolledBack) {
                ringColor = "ring-green-200";
                bgColor = 'bg-green-100';
                hoverBgColor = 'hover:bg-green-200';
            } else if (rollout.gitopsCommitStatus.includes("Failed") && !rollout.rolledBack) {
                ringColor = "ring-red-400";
                bgColor = 'bg-red-100';
                hoverBgColor = 'hover:bg-red-200';
            }

            renderReleaseStatuses.unshift(rolloutWidget(idx, bgColor, hoverBgColor, ringColor, exactDate, dateLabel, undefined, undefined, undefined, undefined, rollout))
        })

        return (
            <div className="mb-12">
                <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">{env}</h4>
                {releaseStatuses.length > 0 ?
                    <div className="bg-white p-4 rounded">
                        <div className="bg-yellow-50 rounded">
                            <div className="flow-root">
                                <ul className="-mb-4 p-2">
                                    {renderReleaseStatuses}
                                </ul>
                            </div>
                        </div>
                    </div>
                    :
                    <div className="text-xs text-gray-800">No releases in the past {releaseHistorySinceDays} days.</div>}
            </div>
        );
    }
};
