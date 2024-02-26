import { Component } from 'react';
import {
  ACTION_TYPE_AGENTS,
  ACTION_TYPE_ENVS,
  ACTION_FLUX_EVENTS_RECEIVED,
  ACTION_TYPE_GIT_REPOS,
  ACTION_TYPE_GITOPS_COMMITS,
  ACTION_TYPE_USERS,
  ACTION_TYPE_APPLICATION,
  ACTION_TYPE_SETTINGS,
  ACTION_TYPE_ALERTS,
  ACTION_TYPE_CHART_UPDATE_PULLREQUESTS,
  ACTION_TYPE_GITOPS_UPDATE_PULLREQUESTS
} from "./redux/redux";

export default class APIBackend extends Component {

  componentDidMount() {
    if (this.props.location.pathname.startsWith('/login')) {
      return;
    }

    this.props.gimletClient.getAgents()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_AGENTS, payload: data }), () => {/* Generic error handler deals with it */
      });
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
    this.props.gimletClient.getGitRepos()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GIT_REPOS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getGitopsCommits()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GITOPS_COMMITS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getSettings()
      .then(data => {
        this.props.store.dispatch({ type: ACTION_TYPE_SETTINGS, payload: data });
      }, () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getChartUpdatePullRequests()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_CHART_UPDATE_PULLREQUESTS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getGitopsUpdatePullRequests()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GITOPS_UPDATE_PULLREQUESTS, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getAlerts()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_ALERTS, payload: data }), () => {/* Generic error handler deals with it */
      });
  }

  render() {
    return null;
  }
}
