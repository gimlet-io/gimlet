import React from 'react'
import CopiableCodeSnippet from '../envConfig/copiableCodeSnippet';

const BootstrapGuide = ({ envName, host, token }) => {
  const renderBootstrapGuideText = () => {
    return (
      <>
        <li>👉 Set the API Key</li>
        <CopiableCodeSnippet
          copiable
          color="blue"
          code={`mkdir -p ~/.gimlet

cat << EOF > ~/.gimlet/config
export GIMLET_SERVER=${host}
export GIMLET_TOKEN=${token}
EOF

source ~/.gimlet/config`}
        />

        <li>👉 Apply the gitops manifests on the cluster to start the gitops loop:</li>
        <CopiableCodeSnippet
          copiable
          color="blue"
          code={`gimlet environment connect --env ${envName}`}
        />

      </>)
  };

  return (
    <div className="rounded-md bg-blue-50 p-4 mb-4 overflow-hidden">
      <ul className="break-all text-sm text-blue-700 space-y-2">
        <span className="text-lg font-bold text-blue-800">Connect environment</span>
        {renderBootstrapGuideText()}
      </ul>
    </div>
  );
};

export default BootstrapGuide;
