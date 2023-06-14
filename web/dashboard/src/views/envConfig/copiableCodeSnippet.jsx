import { copyToClipboard } from "../profile/profile";
import { useState } from 'react';

const CopiableCodeSnippet = ({ code, copiable, color = 'gray' }) => {
  const [isCopied, setIsCopied] = useState(false);

  const textColorClasses = {
    gray: "text-gray-500",
    blue: "text-blue-500",
  };
  const bgColorClasses = {
    gray: "bg-gray-50",
    blue: "bg-blue-100"
  };

  const handleCopyClick = () => {
    setIsCopied(true);

    setTimeout(() => {
      setIsCopied(false);
    }, 2000);
  };

  return (
    <div className="relative">
      <code className={`block whitespace-pre overflow-x-scroll font-mono text-xs my-4 p-2 ${bgColorClasses[color]} ${textColorClasses[color]} rounded`}>
        {copiable &&
          <svg
            onClick={() => {
              copyToClipboard(code)
              handleCopyClick()
            }}
            xmlns="http://www.w3.org/2000/svg" className="cursor-pointer float-right h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>}
        {code}
      </code>
      {isCopied && (
        <div className="absolute top-8 right-0">
          <div className="p-2 bg-indigo-600 select-none text-white inline-block rounded">
            Copied!
          </div>
        </div>
      )}
    </div>
  )
}

export default CopiableCodeSnippet;
