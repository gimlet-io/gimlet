import React, { Component, useEffect } from 'react';
import './app.css';
import Nav from "./components/nav/nav";
import StreamingBackend from "./streamingBackend";
import { createStore } from 'redux';
import { rootReducer } from './redux/redux';
import { BrowserRouter as Router, Route, Routes, useNavigate } from "react-router-dom";
import GimletClient from "./client/client";
import Repositories from "./views/repositories/repositories";
import APIBackend from "./apiBackend";
import Profile from "./views/profile/profile";
import Settings from "./views/settings/settings";
import Repo from "./views/repo/repo";
import { CommitView } from "./views/repo/commitView";
import { RepoSettingsView } from "./views/repo/settingsView";
import { PreviewView } from "./views/repo/previewView";
import LoginPage from './views/login/loginPage';
import EnvConfig from './views/envConfig/envConfig'
import { DeployWizzard } from './views/deployWizzard/deployWizzard'
import RepositoryWizard from './views/repositoryWizard/repositoryWizard';
import Environments from './views/environments/environments'
import Environment from './views/environment/environment';
import PopUpWindow from './popUpWindow';
import Footer from './views/footer/footer';
import {
  ACTION_TYPE_USER,
  ACTION_TYPE_SETTINGS,
} from "./redux/redux";
import Posthog from './posthog';
import './style.css'
import GithubIntegration from './views/githubIntegration';

export default class App extends Component {
  constructor(props) {
    super(props);

    const store = createStore(rootReducer);
    const gimletClient = new GimletClient(
      (response) => {
        if (response.status === 401) {
          if (!window.location.pathname.includes("/login")) {
            localStorage.setItem('redirect', window.location.pathname);
            window.location.replace("/login");
          }
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

  componentDidMount() {
    this.state.gimletClient.getUser()
      .then(data => {
        this.state.store.dispatch({ type: ACTION_TYPE_USER, payload: data });
        this.setState({
          userLoaded: true,
          authenticated: true
        });
        this.state.gimletClient.getSettings()
          .then(data => {
            this.state.store.dispatch({ type: ACTION_TYPE_SETTINGS, payload: data });
            this.setState({ settings: data });
          });
      }, () => {
        this.setState({
          userLoaded: true,
        });
      });
  }

  render() {
    const { store, gimletClient } = this.state;

    if (!this.state.userLoaded) {
      return (<div>loading</div>)
    }

    if (!this.state.authenticated) {
      return (
        <Router>
          <div className="min-h-screen bg-neutral-100 dark:bg-neutral-900 pb-20">
            <div className="py-10">
              <Routes>
                <Route path="/login"  element={ <LoginPage />} />
              </Routes>
            </div>
          </div>
        </Router>
      )
    }

    if (!this.state.settings) {
      return (<div>loading</div>)
    }

    if(!this.state.settings.provider || this.state.settings.provider === ""){
      return (
        <Router>
          <div className="min-h-screen bg-neutral-100 dark:bg-neutral-900 pb-20">
            <GithubIntegration store={store} gimletClient={gimletClient} />
          </div>
        </Router>
      )
    }

    return (
      <Router>
        <StreamingBackend store={store} />
        <APIBackend store={store} gimletClient={gimletClient} />
        <PopUpWindow store={store} />
        <Posthog store={store} />

        <div className="min-h-screen bg-neutral-100 dark:bg-neutral-900 pb-20">
          <Footer store={store} gimletClient={gimletClient} />
          <div className="">
            <Routes>
              <Route exact path="/" element={<RedirectToRepositories />} />

              <Route path="/repositories" element= {
                <>
                  <Nav store={store} />
                  <Repositories store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/environments" element= {
                <>
                  <Nav store={store} />
                  <Environments store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/env/:env/:tab?" element= {
                <>
                  <Nav store={store} />
                  <Environment store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/cli" element= {
                <>
                  <Nav store={store} />
                  <Profile store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/settings" element= {
                <>
                  <Nav store={store} />
                  <Settings store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/login" element= {
                <>
                  <Nav store={store} />
                  <LoginPage />
                </>
              } />

              <Route path="/repo/:owner/:repo/envs/:env/config/:config/:action?/:nav?" element= {
                <>
                  <Nav store={store} />
                  <EnvConfig store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/repo/:owner/:repo/envs/:env/deploy" element= {
                <>
                  <Nav store={store} />
                  <DeployWizzard store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/import-repositories" element= {
                <>
                  <Nav store={store} />
                  <RepositoryWizard store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/repo/:owner/:repo/commits" element= {
                <>
                  <Nav store={store} />
                  <CommitView store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/repo/:owner/:repo/settings/:nav?" element= {
                <>
                  <Nav store={store} />
                  <RepoSettingsView store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/repo/:owner/:repo/previews/:deployment?" element= {
                <>
                  <Nav store={store} />
                  <PreviewView store={store} gimletClient={gimletClient} />
                </>
              } />

              <Route path="/repo/:owner/:repo/:environment?/:deployment?" element= {
                <>
                  <Nav store={store} />
                  <Repo store={store} gimletClient={gimletClient} />
                </>
              } />

            </Routes>
          </div>
        </div>
      </Router>
    )
  }
}

const RedirectToRepositories = () => {
  const navigate = useNavigate();

  useEffect(() => {
    navigate('/repositories');
  });

  return null;
};
