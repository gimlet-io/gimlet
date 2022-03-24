import { BootstrapGuide } from 'shared-components';

const StepThree = ({ appId, infraRepo, appsRepo, bootstrapMessage }) => {

    return (
        <div class="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="max-w-4xl mx-auto">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate mb-12">Gimlet Installer</h2>
                    </div>
                </div>
                <nav aria-label="Progress">
                    <ol class="border border-gray-300 rounded-md divide-y divide-gray-300 md:flex md:divide-y-0">
                        <li class="relative md:flex-1 md:flex">
                            {/* <!-- Completed Step --> */}
                            <div class="group flex items-center w-full select-none cursor-default">
                                <span class="px-6 py-4 flex items-center text-sm font-medium">
                                    <span class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full ">
                                        {/* <!-- Heroicon name: solid/check --> */}
                                        <svg class="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                                            fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd"
                                                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                                clip-rule="evenodd" />
                                        </svg>
                                    </span>
                                    <span class="ml-4 text-sm font-medium text-gray-900">Create Github Application</span>
                                </span>
                            </div>

                            {/* <!-- Arrow separator for lg screens and up --> */}
                            <div class="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                                <svg class="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                                    <path d="M0 -2L20 40L0 82" vector-effect="non-scaling-stroke" stroke="currentcolor"
                                        stroke-linejoin="round" />
                                </svg>
                            </div>
                        </li>

                        <li class="relative md:flex-1 md:flex">
                            {/* <!-- Completed Step --> */}
                            <div class="group flex items-center w-full select-none cursor-default">
                                <span class="px-6 py-4 flex items-center text-sm font-medium">
                                    <span class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full ">
                                        {/* <!-- Heroicon name: solid/check --> */}
                                        <svg class="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                                            fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd"
                                                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                                clip-rule="evenodd" />
                                        </svg>
                                    </span>
                                    <span class="ml-4 text-sm font-medium text-gray-900">Prepare gitops repository</span>
                                </span>
                            </div>

                            {/* <!-- Arrow separator for lg screens and up --> */}
                            <div class="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                                <svg class="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                                    <path d="M0 -2L20 40L0 82" vector-effect="non-scaling-stroke" stroke="currentcolor"
                                        stroke-linejoin="round" />
                                </svg>
                            </div>
                        </li>

                        <li class="relative md:flex-1 md:flex">
                            {/* <!-- Current Step --> */}
                            <div class="px-6 py-4 flex items-center text-sm font-medium select-none cursor-default" aria-current="step">
                                <span
                                    class="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-indigo-600 rounded-full">
                                    <span class="text-indigo-600">03</span>
                                </span>
                                <span class="ml-4 text-sm font-medium text-indigo-600">Bootstrap gitops automation</span>
                            </div>
                        </li>
                    </ol>
                </nav>

                {appId === "" &&
                    <div class="rounded-md bg-red-50 p-4 my-8">
                        <div class="flex">
                            <div class="flex-shrink-0">
                                {/* <!-- Heroicon name: solid/x-circle --> */}
                                <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                    aria-hidden="true">
                                    <path fill-rule="evenodd"
                                        d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                                        clip-rule="evenodd" />
                                </svg>
                            </div>
                            <div class="ml-3">
                                <h3 class="text-sm font-medium text-red-800">A Github Application was not yet created by the installer. Go
                                    to <a href="/" class="font-bold">Step One</a> to create it.</h3>
                            </div>
                        </div>
                    </div>
                }
                <div class="text-sm">
                    <h3 class="text-2xl font-bold pt-16">Your gitops repositories are now prepared</h3>
                    <div class="pt-4">
                        <p>
                            Go checkout the repo for your infrastructure components: <br />
                            👉 <a href={`https://github.com/${infraRepo}`} class="text-blue-600" rel="noreferrer" target="_blank">https://github.com/{infraRepo}</a>
                        </p>
                        <p class="mt-2">
                            Don't forget to check the repo for your own applications: <br />
                            👉 <a href={`https://github.com/${appsRepo}`} class="text-blue-600" rel="noreferrer" target="_blank">https://github.com/{appsRepo}</a>
                        </p>
                    </div>
                    <h3 class="text-2xl font-bold pt-16">Kick off the gitops sync loop with the following steps</h3>
                    <BootstrapGuide
                        envName={bootstrapMessage.envName}
                        repoPath={bootstrapMessage.infraRepo}
                        repoPerEnv={bootstrapMessage.repoPerEnv}
                        publicKey={bootstrapMessage.infraPublicKey}
                        secretFileName={bootstrapMessage.infraSecretFileName}
                        gitopsRepoFileName={bootstrapMessage.infraGitopsRepoFileName}
                        isNewRepo={bootstrapMessage.isNewInfraRepo}
                    />
                    <BootstrapGuide
                        envName={bootstrapMessage.envName}
                        repoPath={bootstrapMessage.appsRepo}
                        repoPerEnv={bootstrapMessage.repoPerEnv}
                        publicKey={bootstrapMessage.appsPublicKey}
                        secretFileName={bootstrapMessage.appsSecretFileName}
                        gitopsRepoFileName={bootstrapMessage.appsGitopsRepoFileName}
                        isNewRepo={bootstrapMessage.isNewAppsRepo}
                    />
                </div>
                <div class="p-0 flow-root my-8">
                    <span class="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                        <button
                            class="bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700 inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150"
                            // eslint-disable-next-line no-restricted-globals
                            onClick={close()}>
                            I am done
                        </button>
                    </span>
                </div>
            </div>
        </div>

    );
};

export default StepThree;
