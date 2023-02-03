import { Component } from 'react';
import {
  ACTION_TYPE_AGENTS,
  ACTION_TYPE_ENVS,
  ACTION_TYPE_GIMLETD, ACTION_TYPE_GIT_REPOS,
  ACTION_TYPE_GITOPS_REPO,
  ACTION_TYPE_GITOPS_COMMITS,
  ACTION_TYPE_USER,
  ACTION_TYPE_USERS,
  ACTION_TYPE_APPLICATION,
  ACTION_TYPE_SETTINGS
} from "./redux/redux";

export default class APIBackend extends Component {

  componentDidMount() {
    console.log(this.props.location.pathname);

    if (this.props.location.pathname.startsWith('/login')) {
      return;
    }

    this.props.gimletClient.getGitopsRepo()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GITOPS_REPO, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getAgents()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_AGENTS, payload: data }), () => {/* Generic error handler deals with it */
      });
    this.props.gimletClient.getUser()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_USER, payload: data }), () => {/* Generic error handler deals with it */
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
    this.props.gimletClient.getGitRepos()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GIT_REPOS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getGitopsCommits()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_GITOPS_COMMITS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getSettings()
      .then(data => this.props.store.dispatch({ type: ACTION_TYPE_SETTINGS, payload: data }), () => {/* Generic error handler deals with it */
      });
      this.props.gimletClient.getEvents();
      this.props.gimletClient.getIrregularPods();
  }

  render() {
    return null;
  }
}
