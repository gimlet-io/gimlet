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

export function popupWindowOpened(state, payload) {
  state.popupWindow.visible = true;
  state.popupWindow.header = payload.header;
  return state;
}

export function popupWindowError(state, payload) {
  state.popupWindow.finished = true;
  state.popupWindow.isError = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.message = payload.message;
  return state;
}

export function popupWindowErrorList(state, payload) {
  state.popupWindow.finished = true;
  state.popupWindow.isError = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.errorList = payload.errorList;
  return state;
}

export function popupWindowSuccess(state, payload) {
  state.popupWindow.finished = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.message = payload.message;
  return state;
}

export function popupWindowReset(state) {
  state.popupWindow.visible = false;
  state.popupWindow.isError = false;
  state.popupWindow.finished = false;
  state.popupWindow.header = "";
  state.popupWindow.message = "";
  state.popupWindow.errorList = null;
  return state;
}

export function envsUpdated(state, allEnvs) {
  allEnvs.connectedAgents.forEach((agent) => {
    state.connectedAgents[agent.name] = agent;
  });
  state.envs = allEnvs.envs;
  return state;
}

export function envStackUpdated(state, envName, payload) {
  state.envs = state.envs.map((env) => {
    if (env.name === envName) {
      env.stackConfig = payload;
    }

    return env;
  });

  return state;
}

export function agentEnvsUpdated(state, connectedAgents) {
  connectedAgents.forEach((agent) => {
    state.connectedAgents[agent.name] = agent;
  });
  return state;
}

export function gitopsCommits(state, gitopsCommits) {
  state.gitopsCommits = gitopsCommits;
  return state;
}

export function updateGitopsCommits(state, event) {
  let isPresent = false;

  state.gitopsCommits.forEach(gitopsCommit => {
    if (gitopsCommit.sha === event.gitopsCommit.sha) {
      gitopsCommit.created = event.gitopsCommit.created;
      gitopsCommit.sha = event.gitopsCommit.sha;
      gitopsCommit.status = event.gitopsCommit.status;
      gitopsCommit.statusDesc = event.gitopsCommit.statusDesc;
      gitopsCommit.env = event.gitopsCommit.env;
      isPresent = true;
    };
  });

  if (!isPresent) {
    state.gitopsCommits.unshift(event.gitopsCommit);
  }

  return state;
}

export function user(state, user) {
  state.user = user;
  return state;
}

export function users(state, users) {
  state.users = users;
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

export function updateCommitStatus(state, event) {
  const repo = `${event.owner}/${event.repo}`;

  state.commits[repo].forEach(commit => {
    if (commit.sha === event.sha) {
      Object.assign(commit.status, event.commitStatus);

      if (event.deployTargets.length > 0) {
        commit.deployTargets = [...event.deployTargets];
      }
    }
  });

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

export function addEnvConfig(state, payload) {
  if (!state.envConfigs[payload.repo][payload.env]) {
    state.envConfigs[payload.repo][payload.env] = []
  }

  state.envConfigs[payload.repo][payload.env].push(
    payload.envConfig
  )
  return state;
}

export function repoMetas(state, payload) {
  state.repoMetas = payload.repoMetas;
  state.fileInfos = payload.repoMetas.fileInfos;
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
