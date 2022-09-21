import React from 'react'

const BootstrapGuide = ({ envName, notificationsFileName, repoPath, repoPerEnv, publicKey, secretFileName, gitopsRepoFileName, controllerGenerated }) => {
    const repoName = parseRepoName(repoPath);
    let type = "";

    if (repoPath.includes("apps")) {
        type = "apps";
    } else if (repoPath.includes("infra")) {
        type = "infra";
    }

    const renderBootstrapGuideText = (controllerGenerated) => {
        return (
            <>
                <li>👉 Clone the Gitops repository</li>
                <ul className="list-none text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">
                    <li>git clone git@github.com:{repoPath}.git</li>
                    <li>cd {repoName}</li>
                </ul>

                <li>👉 Add the following deploy key to your Git provider to the <a href={`https://github.com/${repoPath}/settings/keys`} rel="noreferrer" target="_blank" className="font-medium hover:text-blue-900">{repoName}</a> repository</li>
                <li className="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">{publicKey}</li>
                <li>( Don't know how to do it?
                    <a
                        target="_blank"
                        rel="noreferrer"
                        className="hover:text-blue-900 mx-1 hover:underline"
                        href="https://gimlet.io/docs/make-kubernetes-an-application-platform-with-gimlet-stack/#authorize-flux-to-fetch-your-gitops-repository">
                        click here
                    </a>)
                </li>
                <li>👉 Apply the gitops manifests on the cluster to start the gitops loop:</li>
                <ul className="list-none text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">
                    {controllerGenerated &&
                    <>
                    <li>{repoPerEnv ? `kubectl apply -f flux/flux.yaml` : `kubectl apply -f ${envName}/flux/flux.yaml`}</li>
                    <li>kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io</li>
                    <li>kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io</li>
                    </>
                    }
                    <li>{repoPerEnv ? `kubectl apply -f flux/${secretFileName}` : `kubectl apply -f ${envName}/flux/${secretFileName}`}</li>
                    <li>{repoPerEnv ? `kubectl apply -f flux/${gitopsRepoFileName}` : `kubectl apply -f ${envName}/flux/${gitopsRepoFileName}`}</li>
                    {notificationsFileName && (<li>{repoPerEnv ? `kubectl apply -f flux/${notificationsFileName}` : `kubectl apply -f ${envName}/flux/${notificationsFileName}`}</li>)}
                </ul>

            </>)
    };

    return (
        <div className="rounded-md bg-blue-50 p-4 mb-4 overflow-hidden">
            <ul className="break-all text-sm text-blue-700 space-y-2">
                <span className="text-lg font-bold text-blue-800">Gitops {type}</span>
                {renderBootstrapGuideText(controllerGenerated)}
            </ul>
        </div>
    );
};

const parseRepoName = (repo) => {
    return repo.split("/")[1];
};

export default BootstrapGuide;