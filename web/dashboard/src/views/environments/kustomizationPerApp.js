import React from 'react'
import { Switch } from '@headlessui/react'

const KustomizationPerApp = ({ kustomizationPerApp, setKustomizationPerApp }) => {
    return (<div className="text-gray-700">
        <div className="flex mt-4">
            <div className="font-medium self-center">Kustomization per app</div>
            <div className="max-w-lg flex rounded-md ml-4">
                <Switch
                    checked={kustomizationPerApp}
                    onChange={setKustomizationPerApp}
                    className={(
                        kustomizationPerApp ? "bg-indigo-600" : "bg-gray-200") +
                        " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
                    }
                >
                    <span className="sr-only">Use setting</span>
                    <span
                        aria-hidden="true"
                        className={(
                            kustomizationPerApp ? "translate-x-5" : "translate-x-0") +
                            " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                        }
                    />
                </Switch>
            </div>
        </div>
        <div className="text-sm text-gray-500 leading-loose">To apply only the flux/ folder in gitops. Separate kustomization objects must be created to apply other folders. Used in application git repos.</div>
    </div>)
};

export default KustomizationPerApp;
