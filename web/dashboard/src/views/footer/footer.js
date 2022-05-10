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

    renderGitopsCommit(gitopsCommit) {
        if (gitopsCommit === undefined) {
            return null
        }

        const dateLabel = formatDistance(gitopsCommit.created * 1000, new Date());
        let color = "yellow";
        let lastCommitStatus = "Trailing";

        if (gitopsCommit.status.includes("NotReady")) {
            lastCommitStatus = "Applying";
        } else if (gitopsCommit.status.includes("Succeeded")) {
            color = "green";
            lastCommitStatus = "Applied";
        } else if (gitopsCommit.status.includes("Failed")) {
            color = "red";
            lastCommitStatus = "Apply failed";
        }

        return (
            <div className="flex items-center w-full truncate">
                <p className="font-semibold">{`${gitopsCommit.env.toUpperCase()}`}:</p>
                <div className="w-72 ml-2 cursor-pointer truncate text-sm"
                    title={gitopsCommit.statusDesc}>
                    <span
                        onClick={() => this.props.history.push(`/environments/${gitopsCommit.env}/gitops-commits`)}>
                        <span className={(color === "yellow" && "animate-pulse") + ` h-4 w-4 rounded-full mx-1 relative top-1 inline-block bg-${color}-400`} />
                        {lastCommitStatus}
                        <span className="ml-1">
                            {dateLabel} ago <span className="font-mono">{gitopsCommit.sha.slice(0, 6)}</span>
                        </span>
                    </span>
                    {lastCommitStatus.includes("failed")
                        &&
                        <p class="w-64 truncate">
                            {gitopsCommit.statusDesc}
                        </p>}
                    {lastCommitStatus === "Trailing:" &&
                        <p>
                            Flux is trailing
                        </p>
                    }
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
            <div className="fixed flex justify-center float-left bottom-0 left-0 bg-gray-800 z-50 w-full px-4 py-2 text-gray-100">
                {firstCommitOfEnvs.slice(0, 3).map(gitopsCommit => this.renderGitopsCommit(gitopsCommit))}
            </div>)
    }
}
