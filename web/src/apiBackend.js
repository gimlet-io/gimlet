import { useEffect } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_FLUX_EVENTS_RECEIVED,
  ACTION_TYPE_USERS,
  ACTION_TYPE_APPLICATION,
  ACTION_TYPE_ALERTS,
} from "./redux/redux";
import { useLocation } from 'react-router-dom';

export default function APIBackend(props) {
  const { gimletClient, store } = props
  const location = useLocation()

  useEffect(() => {
    if (location.pathname.startsWith('/login')) {
      return;
    }

    gimletClient.getUsers()
      .then(data => store.dispatch({ type: ACTION_TYPE_USERS, payload: data }), () => {/* Generic error handler deals with it */
      });
    gimletClient.getApp()
      .then(data => store.dispatch({ type: ACTION_TYPE_APPLICATION, payload: data }), () => {/* Generic error handler deals with it */
      });
    gimletClient.getEnvs()
      .then(data => store.dispatch({ type: ACTION_TYPE_ENVS, payload: data }), () => {/* Generic error handler deals with it */
      });
    gimletClient.getFluxEvents()
      .then(data => store.dispatch({ type: ACTION_FLUX_EVENTS_RECEIVED, payload: data }), () => {/* Generic error handler deals with it */
      });
    gimletClient.getAlerts()
      .then(data => store.dispatch({ type: ACTION_TYPE_ALERTS, payload: data }), () => {/* Generic error handler deals with it */
      });
  })

  return null;
}
