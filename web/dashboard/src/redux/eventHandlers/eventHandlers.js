export function agentConnected(state, event) {
  state.settings.agents.push(event.agent);
  return state;
}

export function agentDisconnected(state, event) {
  state.settings.agents = state.settings.agents.filter(agent => agent.name !== event.agent.name);
  return state;
}

export function gitopsRepo(state, event) {
  state.settings.gitopsRepo = event.gitopsRepo;
  return state;
}

export function gitRepos(state, event) {
  state.gitRepos = event;
  return state;
}

export function agents(state, event) {
  state.settings.agents = event.agents;
  return state;
}

export function envsUpdated(state, envs) {
  envs.forEach((env) => {
    state.envs[env.name] = env;
  });
  return state;
}

export function user(state, user) {
  state.user = user;
  return state;
}

export function application(state, application) {
  state.application = { name: application.appName, appSettingsURL: application.appSettingsURL, installationURL: application.installationURL };
  return state;
}

export function gimletd(state, gimletd) {
  state.gimletd = gimletd;
  return state;
}

export function schemas(state, schemas) {
  state.chartSchema = schemas.chartSchema;
  state.chartUISchema = schemas.uiSchema;
  return state;
}

export function search(state, search) {
  state.search = search;
  return state;
}

export function rolloutHistory(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  state.rolloutHistory[repo] = payload.releases;
  return state;
}

export function commits(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  state.commits[repo] = payload.commits;
  return state;
}

export function branches(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  state.branches[repo] = payload.branches;
  return state;
}

export function envConfigs(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  state.envConfigs[repo] = payload.envConfigs;
  return state;
}

export function deploy(state, payload) {
  state.runningDeploys = [payload];
  return state;
}

export function deployStatus(state, payload) {
  for (let runningDeploy of state.runningDeploys) {
    if (runningDeploy.sha === payload.sha &&
      runningDeploy.env === payload.env &&
      runningDeploy.app === payload.app) {
      runningDeploy = payload;
    }
  }

  return state;
}

export function clearDeployStatus(state) {
  state.runningDeploys = [];
  return state;
}

export function staleRepoData(state, event) {
  state.repoRefreshQueue.push(event.repo);
  return state;
}
