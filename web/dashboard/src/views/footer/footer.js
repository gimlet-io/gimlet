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

        const color = gitopsCommit.status.includes("Succeeded") ?
            "green"
            :
            gitopsCommit.status.includes("Failed") ?
                "red"
                :
                "yellow";

        const lastCommitStatus = gitopsCommit.status.includes("Succeeded") ?
            "Applied:"
            :
            gitopsCommit.status.includes("NotReady") ?
                "Applying:"
                :
                gitopsCommit.status.includes("Failed") ?
                    "Apply failed:"
                    :
                    "Trailing:";

        return (
            <div className="flex items-center w-full m-2">
                <p className="font-semibold">{`${gitopsCommit.env.toUpperCase()}:`}</p>
                <ul className="ml-4">
                    <li className="flex items-center cursor-pointer"
                        onClick={() => {
                            window.location.href = `/environments/${gitopsCommit.env}/gitops-commits`
                            return true
                        }}>
                        {lastCommitStatus}
                        <span className={(color === "yellow" && "animate-pulse") + ` h1 rounded-full p-2 mx-1 bg-${color}-400`} />
                        {`${dateLabel} ago ${gitopsCommit.sha && gitopsCommit.sha.slice(0, 6)}`}
                    </li>
                    {lastCommitStatus.includes("failed")
                        &&
                        <li>{gitopsCommit.statusDesc}</li>}
                    {lastCommitStatus === "Trailing:" &&
                        <li>Flux is trailing</li>}
                </ul>
            </div>
        );
    }

    arrayWithFirstCommitOfEnvs() {
        let array = [];
        this.state.envs.map((env) => array.push(this.state.gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env.name)[0]));
        array.sort((a, b) => b.created - a.created)
        return array;
    };

    render() {
        return (
            <div className="fixed flex justify-center float-left bottom-0 left-0 bg-gray-800 z-50 w-full p-2 text-gray-100">
                {this.arrayWithFirstCommitOfEnvs().slice(0, 3).map(gitopsCommit => this.renderGitopsCommit(gitopsCommit))}
            </div>)
    }
}
