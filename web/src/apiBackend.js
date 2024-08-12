import { Component } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_FLUX_EVENTS_RECEIVED,
  ACTION_TYPE_USERS,
  ACTION_TYPE_APPLICATION,
  ACTION_TYPE_ALERTS,
} from "./redux/redux";

export default class APIBackend extends Component {

  componentDidMount() {
    if (this.props.location.pathname.startsWith('/login')) {
      return;
    }

    this.props.gimletClient.getUsers()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_USERS, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getApp()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_APPLICATION, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getEnvs()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_ENVS, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getFluxEvents()
      .then(data => this.props.store.dispatch({ type: ACTION_FLUX_EVENTS_RECEIVED, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getAlerts()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_ALERTS, payload: data }), () => {/* Generic error handler deals with it */
      });
  }

  render() {
    return null;
  }
}
