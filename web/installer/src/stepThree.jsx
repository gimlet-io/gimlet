import { useState, useEffect } from 'react';
import { BootstrapGuide } from 'shared-components';
import StreamingBackend from './streamingBackend';
import GimletCLIClient from './client';

const StepThree = ({ getContext }) => {
    const [context, setContext] = useState(null);

    const client = new GimletCLIClient()
    client.onError = (response) => {
        console.log(response)
        console.log(`${response.status}: ${response.statusText} on ${response.path}`)
    }

    useEffect(() => {
        getContext().then(data => setContext(data))
            .catch(err => {
                console.error(`Error: ${err}`);
            });
    }, [getContext]);

    if (!context) {
        return null;
    }

    return (
        <>
            <StreamingBackend client={client} />
            <div className="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="max-w-4xl mx-auto">
                    <div className="md:flex md:items-center md:justify-between">
                        <div className="flex-1 min-w-0">
                            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate mb-12">Gimlet Installer</h2>
                        </div>
                    </div>
                    <nav aria-label="Progress">
                        <ol className="border border-gray-300 rounded-md divide-y divide-gray-300 md:flex md:divide-y-0">
                            <li className="relative md:flex-1 md:flex">
                                {/* <!-- Completed Step --> */}
                                <div className="group flex items-center w-full select-none cursor-default">
                                    <span className="px-6 py-4 flex items-center text-sm font-medium">
                                        <span className="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full ">
                                            {/* <!-- Heroicon name: solid/check --> */}
                                            <svg className="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                                                fill="currentColor" aria-hidden="true">
                                                <path fillRule="evenodd"
                                                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                                    clipRule="evenodd" />
                                            </svg>
                                        </span>
                                        <span className="ml-4 text-sm font-medium text-gray-900">Create Github Application</span>
                                    </span>
                                </div>

                                {/* <!-- Arrow separator for lg screens and up --> */}
                                <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                                    <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                                        <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                                            strokeLinejoin="round" />
                                    </svg>
                                </div>
                            </li>

                            <li className="relative md:flex-1 md:flex">
                                {/* <!-- Completed Step --> */}
                                <div className="group flex items-center w-full select-none cursor-default">
                                    <span className="px-6 py-4 flex items-center text-sm font-medium">
                                        <span className="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full ">
                                            {/* <!-- Heroicon name: solid/check --> */}
                                            <svg className="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                                                fill="currentColor" aria-hidden="true">
                                                <path fillRule="evenodd"
                                                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                                    clipRule="evenodd" />
                                            </svg>
                                        </span>
                                        <span className="ml-4 text-sm font-medium text-gray-900">Prepare gitops repository</span>
                                    </span>
                                </div>

                                {/* <!-- Arrow separator for lg screens and up --> */}
                                <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                                    <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                                        <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                                            strokeLinejoin="round" />
                                    </svg>
                                </div>
                            </li>

                            <li className="relative md:flex-1 md:flex">
                                {/* <!-- Current Step --> */}
                                <div className="px-6 py-4 flex items-center text-sm font-medium select-none cursor-default" aria-current="step">
                                    <span
                                        className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-indigo-600 rounded-full">
                                        <span className="text-indigo-600">03</span>
                                    </span>
                                    <span className="ml-4 text-sm font-medium text-indigo-600">Bootstrap gitops automation</span>
                                </div>
                            </li>
                        </ol>
                    </nav>

                    {context.appId === "" ?
                        (<div className="rounded-md bg-red-50 p-4 my-8">
                            <div className="flex">
                                <div className="flex-shrink-0">
                                    {/* <!-- Heroicon name: solid/x-circle --> */}
                                    <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                        aria-hidden="true">
                                        <path fillRule="evenodd"
                                            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                                            clipRule="evenodd" />
                                    </svg>
                                </div>
                                <div className="ml-3">
                                    <h3 className="text-sm font-medium text-red-800">A Github Application was not yet created by the installer. Go
                                        to <a href="/" className="font-bold">Step One</a> to create it.</h3>
                                </div>
                            </div>
                        </div>)
                        :
                        (<div className="text-sm">
                            <h3 className="text-2xl font-bold pt-16">Your gitops repositories are now prepared</h3>
                            <div className="pt-4">
                                <p>
                                    Go checkout the repo for your infrastructure components: <br />
                                    👉 <a href={`https://github.com/${context.infraRepo}`} className="text-blue-600" rel="noreferrer" target="_blank">https://github.com/{context.infraRepo}</a>
                                </p>
                                <p className="mt-2">
                                    Don't forget to check the repo for your own applications: <br />
                                    👉 <a href={`https://github.com/${context.appsRepo}`} className="text-blue-600" rel="noreferrer" target="_blank">https://github.com/{context.appsRepo}</a>
                                </p>
                            </div>
                            <h3 className="text-2xl font-bold pt-16">Kick off the gitops sync loop with the following steps</h3>
                            <BootstrapGuide
                                envName={context.envName}
                                repoPath={context.infraRepo}
                                repoPerEnv={context.repoPerEnv}
                                publicKey={context.infraPublicKey}
                                secretFileName={context.infraSecretFileName}
                                gitopsRepoFileName={context.infraGitopsRepoFileName}
                                isNewRepo={context.isNewInfraRepo}
                            />
                            <BootstrapGuide
                                envName={context.envName}
                                repoPath={context.appsRepo}
                                repoPerEnv={context.repoPerEnv}
                                publicKey={context.appsPublicKey}
                                secretFileName={context.appsSecretFileName}
                                gitopsRepoFileName={context.appsGitopsRepoFileName}
                                isNewRepo={context.isNewAppsRepo}
                                notificationsFileName={context.notificationsFileName}
                            />
                            <div className="rounded-md bg-blue-50 p-4 mb-4 overflow-hidden">
                                <ul className="break-all text-sm text-blue-700 space-y-2">
                                    <li>👉 Add the following deploy key to your Git provider to the <a href={`https://github.com/${context.appsRepo}`} rel="noreferrer" target="_blank" className="font-medium hover:text-blue-900">{context.appsRepo}</a> repository</li>
                                    <li className="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">{context.gimletdPublicKey}</li>
                                </ul>
                            </div>
                            <div className='text-gray-900 mt-16 mb-32'>
                                <h2 className=''>Happy Gitopsing🎊</h2>
                                <h2 className='mt-16 font-bold'>Now you can close this browser tab, and return to the Terminal to finalize the install.</h2>
                            </div>
                        </div>)}
                </div>
            </div>
        </>
    );
};

export default StepThree;
