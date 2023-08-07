import React from 'react'
import CopiableCodeSnippet from '../envConfig/copiableCodeSnippet';

const BootstrapGuide = ({ envName, token }) => {
  const renderBootstrapGuideText = () => {
    const url = window.location.protocol + "//" + window.location.host;

    return (
      <>
      <li>👉 Need a cluster? <a className='underline' target='_blank' rel="noopener noreferrer" href="https://gimlet.io/blog/running-kubernetes-on-your-laptop-with-k3d">We recommend using k3d</a> on your laptop if you are evaluating Gimlet.
      <br/>But any Kubernetes cluster will do. Skip this step if you have one.<br />
      <br/>Optional - Run the following commands to get a containerized Kubernetes cluster on your laptop:
                <CopiableCodeSnippet
          copiable
          color="blue"
          code={`curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create gimlet-cluster --k3s-arg "--disable=traefik@server:0"`}
        />
                </li>
        <li>👉 Install Gimlet CLI</li>
        <CopiableCodeSnippet
          copiable
          color="blue"
          code={`curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.23.4/gimlet-$(uname)-$(uname -m)" -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet`}
        />

        <li>👉 Connect your cluster</li>
        <CopiableCodeSnippet
          copiable
          color="blue"
          code={`gimlet environment connect \\
  --env ${envName} \\
  --server ${url} \\
  --token ${token}`}
        />

      </>)
  };

  return (
    <div className="p-4 mb-4 overflow-hidden">
      <ul className="break-all text-sm text-blue-700 space-y-2">
        {renderBootstrapGuideText()}
      </ul>
    </div>
  );
};

export default BootstrapGuide;
