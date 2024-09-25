import { useState, useEffect } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET,
} from "../../redux/redux";
import EnvironmentCard from '../../components/environmentCard/environmentCard';
import { SkeletonLoader } from '../../../src/views/repositories/repositories';
import { useNavigate } from 'react-router-dom'

export default function Environments(props) {
  const { store, gimletClient } = props
  const navigate = useNavigate()

  const reduxState = store.getState();
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [envs, setEnvs] = useState(reduxState.envs)
  const [input, setInput] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const [saveButtonTriggered, setSaveButtonTriggered] = useState(false)
  const [settings, setSettings] = useState(reduxState.settings)
  const [environmentsLoading, setEnvironmentsLoading] = useState(true)

  store.subscribe(() => {
    const reduxState = store.getState();
    setConnectedAgents(reduxState.connectedAgents)
    setEnvs(reduxState.envs)
    setSettings(reduxState.settings)
  });

  useEffect(() => {
    gimletClient.getEnvs()
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ENVS, payload: data
        });
        setEnvironmentsLoading(false);
      }, () => {
      setEnvironmentsLoading(false);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const sortingByName = (envs) => {
    const envsCopy = [...envs]
    return envsCopy.sort((a, b) => a.name.localeCompare(b.name));
  }

  const isOnline = (onlineEnvs, singleEnv) => {
    return Object.keys(onlineEnvs)
      .map(env => onlineEnvs[env])
      .some(onlineEnv => {
        return onlineEnv.name === singleEnv.name
      })
  };

  const refreshEnvs = () => {
    gimletClient.getEnvs()
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ENVS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  const save = () => {
    if (envs.some(env => env.name === input)) {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
          header: "Error",
          message: "Environment already exists"
        }
      });
      return
    }

    setIsOpen(false);
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving..."
      }
    });
    setSaveButtonTriggered(true);
    gimletClient.saveEnvToDB(input)
      .then(() => {
        setEnvs([...envs, {name: input, infraRepo: "", appsRepo: "", expiry: 0}])
        setInput("")
        setSaveButtonTriggered(false);
        refreshEnvs();
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWRESET
        });
      }, err => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
    })
  }

  if (!envs) {
    return null;
  }

  if (!connectedAgents) {
    return null;
  }

  const sortedEnvs = sortingByName(envs);

  return (
    <div>
      <header>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow">Environments</h1>
          <button type="button" className={`${!settings.trial ? 'primaryButton' : 'primaryButtonDisabled'} px-4`}
            onClick={() => !settings.trial && setIsOpen(true)}
            title={`${!settings.trial ? '' : 'Upgrade Gimlet to create environments'}`}
            >
            Create
          </button>
        </div>
      </header>
      <main>
        <div className="max-w-7xl mx-auto sm:px-6 lg:px-8 mt-8">
          <div className="px-4 sm:px-0">
            {isOpen &&
              <div className="card mb-4 p-4 flex space-x-2 items-center">
                <input
                  onChange={e => setInput(e.target.value)}
                  className="input" id="environment" type="text" value={input} placeholder="Staging" />
                <div className="p-0 flow-root space-x-1">
                  <span className="inline-flex rounded-md shadow-sm gap-x-1 float-right">
                    <button
                      disabled={input === "" || saveButtonTriggered}
                      onClick={() => save()}
                      className={(input === "" || saveButtonTriggered ? "primaryButtonDisabled" : "primaryButton") + " px-4"}>
                      Save
                    </button>
                    <button
                      disabled={input === "" || saveButtonTriggered}
                      onClick={() => setIsOpen(false)}
                      className='border-blue-500 dark:border-blue-700 text-blue-500 dark:text-blue-700 border hover:border-blue-400 dark:hover:border-blue-800 hover:text-blue-400 dark:hover:text-blue-800 cursor-pointer inline-flex items-center px-6 py-2 text-base leading-6 font-medium rounded-md transition ease-in-out duration-150'>
                      Cancel
                    </button>
                  </span>
                </div>
              </div>
            }
            <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {
                environmentsLoading ?
                  <SkeletonLoader />
                  :
                  <>
                    {
                      sortedEnvs.map(env => (<EnvironmentCard
                        key={env.name}
                        env={env}
                        navigateToEnv={() => navigate(`/env/${env.name}`)}
                        isOnline={isOnline(connectedAgents, env)}
                        trial={settings.trial}
                      />))
                    }
                  </>
              }
            </ul>
          </div>
        </div>
      </main>
    </div>
  )
}
