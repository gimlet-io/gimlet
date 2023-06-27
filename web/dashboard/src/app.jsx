import React, { Component } from 'react';
import './app.css';
import Nav from "./components/nav/nav";
import StreamingBackend from "./streamingBackend";
import { createStore } from 'redux';
import { rootReducer } from './redux/redux';
import { BrowserRouter as Router, Redirect, Route, Switch, withRouter } from "react-router-dom";
import GimletClient from "./client/client";
import Repositories from "./views/repositories/repositories";
import APIBackend from "./apiBackend";
import Profile from "./views/profile/profile";
import Settings from "./views/settings/settings";
import Pulse from "./views/pulse/pulse";
import Repo from "./views/repo/repo";
import LoginPage from './views/login/loginPage';
import EnvConfig from './views/envConfig/envConfig'
import Environments from './views/environments/environments'
import PopUpWindow from './popUpWindow';
import Userflow from './userflow';
import DeployPanel from './views/deployPanel/deployPanel';

export default class App extends Component {
  constructor(props) {
    super(props);

    const store = createStore(rootReducer);
    const gimletClient = new GimletClient(
      (response) => {
        if (response.status === 401) {
          window.location.replace("/login");
        } else {
          console.log(`${response.status}: ${response.statusText} on ${response.path}`);
        }
      }
    );

    this.state = {
      store: store,
      gimletClient: gimletClient
    }
  }

  render() {
    const { store, gimletClient } = this.state;

    const NavBar = withRouter(props => <Nav {...props} store={store} />);
    const APIBackendWithLocation = withRouter(
      props => <APIBackend {...props} store={store} gimletClient={gimletClient} />
    );
    const StreamingBackendWithLocation = withRouter(props => <StreamingBackend {...props} store={store} />);
    const RepoWithRouting = withRouter(props => <Repo {...props} store={store} gimletClient={gimletClient} />);
    const PulseWithRouting = withRouter(props => <Pulse {...props} store={store} gimletClient={gimletClient} />);
    const RepositoriesWithRouting = withRouter(props => <Repositories {...props} store={store} gimletClient={gimletClient} />);
    const EnvironmentsWithRouting = withRouter(props => <Environments {...props} store={store} gimletClient={gimletClient} />);
    const ChartUIWithRouting = withRouter(props => <EnvConfig {...props} store={store} gimletClient={gimletClient} />);
    const PopUpWindowWithLocation = withRouter(props => <PopUpWindow {...props} store={store} />);
    const ProfileWithRouting = withRouter(props => <Profile {...props} store={store} gimletClient={gimletClient} />);
    const SettingsWithRouting = withRouter(props => <Settings {...props} store={store} gimletClient={gimletClient} />);

    return (
      <Router>
        <StreamingBackendWithLocation />
        <APIBackendWithLocation />
        <PopUpWindowWithLocation />
        <Userflow store={store} />

        <Route exact path="/">
          <Redirect to="/repositories" />
        </Route>

        <div className="min-h-screen bg-gray-100 pb-20">
          <NavBar />
          <DeployPanel store={store} />
          <div className="py-10">
            <Switch>
              <Route path="/pulse">
                <PulseWithRouting />
              </Route>

              <Route path="/repositories">
                <RepositoriesWithRouting />
              </Route>

              <Route path="/environments/:environment?/:tab?">
                <EnvironmentsWithRouting />
              </Route>

              <Route path="/profile">
                <ProfileWithRouting store={store} />
              </Route>

              <Route path="/settings">
                <SettingsWithRouting store={store} />
              </Route>

              <Route path="/login">
                <LoginPage />
              </Route>

              <Route path="/repo/:owner/:repo/envs/:env/config/:config/:action?">
                <ChartUIWithRouting />
              </Route>

              <Route path="/repo/:owner/:repo/:environment?/:deployment?">
                <RepoWithRouting store={store} />
              </Route>

            </Switch>
          </div>
        </div>
      </Router>
    )
  }
}
