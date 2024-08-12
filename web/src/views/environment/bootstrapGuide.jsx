import CopiableCodeSnippet from '../envConfig/copiableCodeSnippet';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

export default function ConnectCluster(props) {
  const { envName, token } = props;
  const url = window.location.protocol + "//" + window.location.host

  return (
    <div className="w-full card">
      <div className="p-6 pb-2 items-center">
        <label htmlFor="environment" className="block font-medium text-2xl">Connect your cluster</label>
        <div className="mt-4 text-sm">
          <p className="font-medium">Prerequisite - Install Gimlet CLI</p>
          <CopiableCodeSnippet
            copiable
            code={`curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.27.0/gimlet-$(uname)-$(uname -m)" -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet`}
          />
          <p className="font-medium">Connect your cluster</p>
          <CopiableCodeSnippet
            copiable
            code={`gimlet environment connect \\
  --env ${envName} \\
  --server ${url} \\
  --token ${token}`}
          />
        </div>
      </div>
      <div className='learnMoreBox'>
        Learn more about <a href="https://gimlet-documentation-home-page-revamp-emxxuioo.gimlet.app/docs/kubernetes-resources" className='learnMoreLink'>Kubernetes Clusters<ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
      </div>
    </div>
  )
};
