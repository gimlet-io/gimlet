import { Component } from 'react';
import { formatDistance } from "date-fns";

export default class Footer extends Component {
    constructor(props) {
        super(props);
        let reduxState = this.props.store.getState();

        this.state = {
            gitopsCommits: reduxState.gitopsCommits,
            envs: reduxState.envs
        };
        this.props.store.subscribe(() => {
            let reduxState = this.props.store.getState();

            this.setState({
                gitopsCommits: reduxState.gitopsCommits,
                envs: reduxState.envs
            });
        });
    }

    renderLastCommitStatusMessage(lastCommitStatus, lastCommitStatusMessage) {
        if (lastCommitStatus === "Apply failed" || lastCommitStatus === "Trailing") {
            return (
                <p className="truncate">
                    {lastCommitStatusMessage}
                </p>);
        }
    }

    renderGitopsCommit(gitopsCommit) {
        if (gitopsCommit === undefined) {
            return null
        }

        if (gitopsCommit.sha === undefined) {
            return null
        }

        const dateLabel = formatDistance(gitopsCommit.created * 1000, new Date());
        let color = "bg-yellow-400";
        let lastCommitStatus = "Trailing";
        let lastCommitStatusMessage = "Flux is trailing";

        if (gitopsCommit.status.includes("NotReady")) {
            lastCommitStatus = "Applying";
        } else if (gitopsCommit.status.includes("Succeeded")) {
            color = "bg-green-400";
            lastCommitStatus = "Applied";
        } else if (gitopsCommit.status.includes("Failed")) {
            color = "bg-red-400";
            lastCommitStatus = "Apply failed";
            lastCommitStatusMessage = gitopsCommit.statusDesc;
        }

        return (
            <div className="flex items-center align-middle justify-center w-full truncate" key={gitopsCommit.sha}>
                <p className="font-semibold">{`${gitopsCommit.env.toUpperCase()}`}:</p>
                <div className="w-72 ml-2 cursor-pointer truncate text-sm"
                    onClick={() => this.props.history.push(`/environments/${gitopsCommit.env}/gitops-commits`)}
                    title={gitopsCommit.statusDesc}>
                    <span>
                        <span className={(color === "bg-yellow-400" && "animate-pulse") + ` h-4 w-4 rounded-full mr-1 relative top-1 inline-block ${color}`} />
                        {lastCommitStatus}
                        <span className="ml-1">
                            {dateLabel} ago <span className="font-mono">{gitopsCommit.sha?.slice(0, 6)}</span>
                        </span>
                    </span>
                    {this.renderLastCommitStatusMessage(lastCommitStatus, lastCommitStatusMessage)}
                </div>
            </div>
        );
    }

    arrayWithFirstCommitOfEnvs() {
        let firstCommitOfEnvs = [];

        for (let env of this.state.envs) {
            firstCommitOfEnvs.push(this.state.gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env.name)[0]);
        }

        firstCommitOfEnvs = firstCommitOfEnvs.filter(commit => commit !== undefined);

        firstCommitOfEnvs.sort((a, b) => b.created - a.created);

        return firstCommitOfEnvs;
    };

    render() {
        if (this.state.gitopsCommits.length === 0 ||
            this.state.envs.length === 0) {
            return null;
        }

        const firstCommitOfEnvs = this.arrayWithFirstCommitOfEnvs()
        if (firstCommitOfEnvs.length === 0) {
            return null;
        }

        return (
            <div className="grid grid-cols-3 fixed bottom-0 left-0 bg-gray-800 z-50 w-full px-4 py-2 text-gray-100">
                {firstCommitOfEnvs.slice(0, 3).map(gitopsCommit => this.renderGitopsCommit(gitopsCommit))}
            </div>)
    }
}
