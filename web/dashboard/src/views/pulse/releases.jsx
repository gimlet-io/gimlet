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

        releaseStatuses.forEach((rollout, idx, arr) => {
            const exactDate = format(rollout.created * 1000, 'h:mm:ss a, MMMM do yyyy');
            const dateLabel = formatDistance(rollout.created * 1000, new Date());

            renderReleaseStatuses.unshift(rolloutWidget(idx, arr, exactDate, dateLabel, undefined, undefined, undefined, undefined, rollout))
        })

        return (
            <div className="mb-12">
                <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">{env}</h4>
                {releaseStatuses.length > 0 ?
                    <div className="bg-white p-4 rounded">
                        <div className="flow-root">
                            <ul>
                                {renderReleaseStatuses}
                            </ul>
                        </div>
                    </div>
                    :
                    <div className="text-xs text-gray-800">No releases in the past {releaseHistorySinceDays} days.</div>}
            </div>
        );
    }
};
