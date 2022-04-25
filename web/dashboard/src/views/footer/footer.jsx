import { React, useState, useEffect } from "react";
import { formatDistance } from "date-fns";

const Footer = ({ store }) => {
    let reduxState = store.getState();
    const [recentGitopsCommit, setRecentGitopsCommit] = useState(reduxState.recentGitopsCommit);
    const [gitopsCommits, setGitopsCommits] = useState(reduxState.gitopsCommits);
    const [time, setTime] = useState(new Date());
      /*eslint no-unused-vars: ["error", { "varsIgnorePattern": "time" }]*/

    store.subscribe(() => {
        let reduxState = store.getState();
        setRecentGitopsCommit(reduxState.recentGitopsCommit)
        setGitopsCommits(reduxState.gitopsCommits)
    });

    useEffect(() => {
        const interval = setInterval(() => setTime(Date.now()), 1000);
        return () => {
          clearInterval(interval);
        };
      }, []);

    const renderGitopsCommit = (gitopsCommit) => {
        const dateLabel = formatDistance(gitopsCommit.created * 1000, new Date());
        const color = gitopsCommit.status.includes("Succeeded") ?
            "green"
            :
            gitopsCommit.status.includes("Failed") ?
                "red"
                :
                "yellow";

        return (
            <li className="flex items-center cursor-pointer" title={gitopsCommit.statusDesc}><span className={(color === "yellow" && "animate-pulse") + ` h1 rounded-full p-2 mr-1 bg-${color}-400`} />{`${dateLabel} ago`}</li>
        );
    }

    const gitopsCommit = recentGitopsCommit ?? gitopsCommits[0];
    const lastAppliedCommit = gitopsCommits.find(gitopsCommit => gitopsCommit.status.includes("Succeeded"));

    if (!gitopsCommit) {
        return null;
    }

    return (
        (<div className="fixed flex justify-center float-left bottom-0 left-0 bg-gray-800 z-50 w-full p-2 text-gray-100">
            <div className="flex items-center w-full m-2">
                <p className="font-semibold">STAGING:</p>
                <ul className="ml-4">
                    {renderGitopsCommit(gitopsCommit)}
                    <li>Last applied: {lastAppliedCommit.sha.slice(0, 6)}</li>
                </ul>
            </div>
            <div className="flex items-center w-full m-2">
                <p className="font-semibold">PROD:</p>
                <ul className="ml-4">
                    {renderGitopsCommit(gitopsCommit)}
                    <li>Last applied: {lastAppliedCommit.sha.slice(0, 6)}</li>
                </ul>
            </div>
        </div>)
    );
};

export default Footer;
