import React from 'react'
import CopiableCodeSnippet from '../envConfig/copiableCodeSnippet';

const BootstrapGuide = ({ envName, token }) => {
  const renderBootstrapGuideText = () => {
    const url = window.location.protocol + "//" + window.location.host;

    return (
      <>
        <li>👉 Install Gimlet CLI</li>
        <CopiableCodeSnippet
          copiable
          color="blue"
          code={`curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.23.2/gimlet-$(uname)-$(uname -m)" -o gimlet
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
    <div className="rounded-md bg-blue-50 p-4 mb-4 overflow-hidden">
      <ul className="break-all text-sm text-blue-700 space-y-2">
        {renderBootstrapGuideText()}
      </ul>
    </div>
  );
};

export default BootstrapGuide;
