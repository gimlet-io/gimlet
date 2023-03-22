import React from 'react'

const GitopsAutomationGuide = () => {
    return (
        <div className="rounded-md bg-blue-50 p-4 mb-4">
            <ul className="text-sm text-blue-700 space-y-2">
                <span className="text-lg font-bold text-blue-800">Verify the gitops automation</span>
                <li>Check Flux's custom resources on the cluster to verify the gitops automation.</li>
                <li>Flux uses the <code>`gitrepository`</code> custom resource to point to git repository locations and credentials. Flux's source controller periodically checks the content of the git repositories, and you can validate their status as follows:</li>
                <code className="block whitespace-pre overflow-x-scroll font-mono text-xs my-4 p-2 bg-blue-100 text-blue-500 rounded">
                    {`kubectl get gitrepositories -A`}
                </code>
                <li>
                    If the git repositories are in ready state, validate the <code>`kustomization`</code> custom resources. These resources point to a path in a git repository to apply yamls from. If they are in ready state, you can be sure the Flux applied your latest manifests.
                </li>
                <code className="block whitespace-pre overflow-x-scroll font-mono text-xs my-4 p-2 bg-blue-100 text-blue-500 rounded">
                    {`kubectl get kustomizations -A`}
                </code>
                <li>
                    Now that the gitops automation is in place, every manifest you put in the gitops repositories will be applied on the cluster by the gitops controller.
                </li>
                <li className="text-lg font-bold text-blue-800">Need to debug Flux?</li>
                <li>If <code>`kustomizations`</code> or <code>`gitrepositories`</code> are not in ready state, you can see the error message if you run `kubectl describe` on them.</li>
                <li>If you need to further debug their behavior, you can check Flux logs in the <code>`flux-system`</code> namespace.</li>
                <code className="block whitespace-pre overflow-x-scroll font-mono text-xs my-4 p-2 bg-blue-100 text-blue-500 rounded">
                    {`kubectl logs -f deploy/kustomize-controller -n flux-system
kubectl logs -f deploy/source-controller -n flux-system`}
                </code>
            </ul>
        </div>)
};

export default GitopsAutomationGuide;
