import React from 'react'
import { Switch } from '@headlessui/react'

const SeparateEnvironments = ({ repoPerEnv, setRepoPerEnv, infraRepo, appsRepo, setInfraRepo, setAppsRepo, envName }) => {
    return (
        <div className="text-gray-700">
            <div className="flex mt-4">
                <div className="font-medium self-center">Separate environments by git repositories</div>
                <div className="max-w-lg flex rounded-md ml-4">
                    <Switch
                        checked={repoPerEnv}
                        onChange={setRepoPerEnv}
                        className={(
                            repoPerEnv ? "bg-indigo-600" : "bg-gray-200") +
                            " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
                        }
                    >
                        <span className="sr-only">Use setting</span>
                        <span
                            aria-hidden="true"
                            className={(
                                repoPerEnv ? "translate-x-5" : "translate-x-0") +
                                " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                            }
                        />
                    </Switch>
                </div>
            </div>
            {repoPerEnv ?
                <div className="text-sm text-gray-500 leading-loose">Manifests will be placed in environment specific repositories</div>
                :
                <div className="text-sm text-gray-500 leading-loose">
                    {`Manifests of this environment will be placed in the ${envName} folder of the shared git repositories:`} <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{infraRepo}</span> and <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{appsRepo}</span>
                </div>}
            <div className="flex mt-4">
                <div className="font-medium self-center">Infrastructure git repository</div>
                <div className="max-w-lg flex rounded-md ml-4">
                    <div className="max-w-lg w-full lg:max-w-xs">
                        <input id="infra" name="infra"
                            className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                            type="text"
                            value={infraRepo}
                            onChange={e => setInfraRepo(e.target.value)}
                        />
                    </div>
                </div>
            </div>
            {repoPerEnv ?
                <div className="text-sm text-gray-500 leading-loose">Infrastructure manifests will be placed in the root of the specified repository</div>
                :
                <div className="text-sm text-gray-500 leading-loose">{`Infrastructure manifests will be placed in ${envName} folder of the specified repository`}</div>}
            <div className="flex mt-4">
                <div className="font-medium self-center">Application git repository</div>
                <div className="max-w-lg flex rounded-md ml-4">
                    <div className="max-w-lg w-full lg:max-w-xs">
                        <input id="apps" name="apps"
                            className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                            type="text"
                            value={appsRepo}
                            onChange={e => setAppsRepo(e.target.value)}
                        />
                    </div>
                </div>
            </div>
            {repoPerEnv ?
                <div className="text-sm text-gray-500 leading-loose">Application manifests will be placed in the root of the specified repository</div>
                :
                <div className="text-sm text-gray-500 leading-loose">{`Application manifests will be placed in ${envName} folder of the specified repository`}</div>}
        </div>
    );
};

export default SeparateEnvironments;
