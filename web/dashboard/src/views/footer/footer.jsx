import { React, useState, useEffect } from "react";
import { formatDistance } from "date-fns";

const Footer = ({ store }) => {
    let reduxState = store.getState();
    const [gitopsCommits, setGitopsCommits] = useState(reduxState.gitopsCommits);
    const [envs, setEnvs] = useState(reduxState.envs);
    // const [time, setTime] = useState(new Date());
    /*eslint no-unused-vars: ["error", { "varsIgnorePattern": "useEffect" }]*/

    store.subscribe(() => {
        let reduxState = store.getState();
        setGitopsCommits(reduxState.gitopsCommits)
        setEnvs(reduxState.envs)
    });

    // useEffect(() => {
    //     const interval = setInterval(() => setTime(Date.now()), 1000);
    //     return () => {
    //         clearInterval(interval);
    //     };
    // }, []);

    const renderGitopsCommit = (gitopsCommit, idx) => {
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
            <li key={idx} className="flex items-center w-full m-2">
                <p className="font-semibold">{`${gitopsCommit.env.toUpperCase()}:`}</p>
                <ul className="ml-4">
                    <li className="flex items-center cursor-pointer" title={gitopsCommit.statusDesc}>
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
            </li>
        );
    }

    const arrayWithFirstCommitOfEnvs = () => {
        let array = [];
        envs.map((env) => array.push(gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env.name)[0]));
        array.sort((a, b) => b.created - a.created)
        return array;
    };

    if (gitopsCommits.length === 0 ||
        envs.length === 0) {
        return null;
    }

    return (
        (
            <ul>
                <div className="fixed flex justify-center float-left bottom-0 left-0 bg-gray-800 z-50 w-full p-2 text-gray-100">
                    {arrayWithFirstCommitOfEnvs().slice(0, 3).map((gitopsCommit, idx) => renderGitopsCommit(gitopsCommit, idx))}
                </div>
            </ul>)
    );
};

export default Footer;
