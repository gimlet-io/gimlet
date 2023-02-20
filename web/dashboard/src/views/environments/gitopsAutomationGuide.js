import React from 'react'
import CopiableCodeSnippet from '../envConfig/copiableCodeSnippet';

const GitopsAutomationGuide = () => {
    return (
        <div className="rounded-md bg-blue-50 p-4 mb-4">
            <ul className="text-sm text-blue-700 space-y-2">
                <span className="text-lg font-bold text-blue-800">Verify the gitops automation</span>
                <li>Check Flux's custom resources on the cluster to verify the gitops automation.</li>
                <li>Flux uses the <code>`gitrepository`</code> custom resource to point to git repository locations and credentials. Flux's source controller periodically checks the content of the git repositories, and you can validate their status as follows:</li>
                <CopiableCodeSnippet
                    color="blue"
                    code={`kubectl get gitrepositories -A
NAMESPACE     NAME                                       URL                                                        AGE    READY   STATUS
flux-system   gitops-repo-gimlet-bootstraping-tutorial   ssh://git@github.com/gimlet/gimlet-bootstraping-tutorial   125m   True    stored artifact for revision 'main/f4a2a676bbcc04f38120b24463ca1c66cc099ab4'`}
                />
                <li>
                    If the git repositories are in ready state, validate the <code>`kustomization`</code> custom resources. These resources point to a path in a git repository to apply yamls from. If they are in ready state, you can be sure the Flux applied your latest manifests.
                </li>
                <CopiableCodeSnippet
                    color="blue"
                    code={`kubectl get kustomizations -A 
NAMESPACE     NAME                                                    AGE    READY   STATUS
flux-system   gitops-repo-gimlet-bootstraping-tutorial                127m   True    Applied revision: main/f4a2a676bbcc04f38120b24463ca1c66cc099ab4
flux-system   gitops-repo-gimlet-bootstraping-tutorial-dependencies   127m   True    Applied revision: main/f4a2a676bbcc04f38120b24463ca1c66cc099ab4`}
                />
                <li>
                    Now that the gitops automation is in place, every manifest you put in the gitops repositories will be applied on the cluster by the gitops controller.
                </li>
                <li className="text-lg font-bold text-blue-800 pt-8">Need to debug Flux?</li>
                <li>If <code>`kustomizations`</code> or <code>`gitrepositories`</code> are not in ready state, you get an error message in their status.</li>
                <li>If you need to further debug their behavior, you can check Flux logs in the <code>`flux-system`</code> namespace.</li>
                <CopiableCodeSnippet
                    color="blue"
                    code={`kubectl logs -f deploy/kustomize-controller -n flux-system
kubectl logs -f deploy/source-controller -n flux-system`}
                />
            </ul>
        </div>)
};

export default GitopsAutomationGuide;
