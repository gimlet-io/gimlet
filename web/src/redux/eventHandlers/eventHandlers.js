import {produce} from 'immer';

export function agentConnected(state, event) {
  state.connectedAgents = produce(state.connectedAgents, draft => {
    draft[event.agent.name] = {name: event.agent.name, stacks: []}
  });
  return state
}

export function agentDisconnected(state, event) {
  state.connectedAgents = produce(state.connectedAgents, draft => {
    delete draft[event.agent.name]
  });
  return state
}

export function gitRepos(state, event) {
  state.gitRepos = event;
  return state;
}

export function popupWindowProgress(state, payload) {
  state.popupWindow.visible = true;
  state.popupWindow.header = payload.header;
  return state;
}

export function popupWindowError(state, payload) {
  state.popupWindow.visible = true;
  state.popupWindow.finished = true;
  state.popupWindow.isError = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.message = payload.message;
  return state;
}

export function popupWindowErrorList(state, payload) {
  state.popupWindow.visible = true;
  state.popupWindow.finished = true;
  state.popupWindow.isError = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.errorList = payload.errorList;
  return state;
}

export function popupWindowSuccess(state, payload) {
  state.popupWindow.visible = true;
  state.popupWindow.finished = true;
  state.popupWindow.header = payload.header;
  state.popupWindow.message = payload.message;
  state.popupWindow.link = payload.link;
  return state;
}

export function popupWindowReset(state) {
  state.popupWindow.visible = false;
  state.popupWindow.isError = false;
  state.popupWindow.finished = false;
  state.popupWindow.header = "";
  state.popupWindow.message = "";
  state.popupWindow.link = "";
  state.popupWindow.errorList = null;
  return state;
}

export function envs(state, allEnvs) {
  state.connectedAgents = produce(state.connectedAgents, draft => {
    allEnvs.connectedAgents.forEach((agent) => {
      draft[agent.name] = agent;
      if (!draft[agent.name].stacks) {
        draft[agent.name].stacks = []
      }
    });
  });

  state.fluxState = produce(state.fluxState, draft => {
    allEnvs.connectedAgents.forEach((agent) => {
      draft[agent.name] = agent.fluxState;
    });
  });

  allEnvs.envs.forEach(env => {
    if (!env.pullRequests) {
      state.envs.forEach(stateEnv => {
        if (env.name === stateEnv.name) {
          env.pullRequests = stateEnv.pullRequests
        }
      });
    }
  });
  state.envs = allEnvs.envs

  return state
}

export function fluxEventsReceived(state, payload) {
  state.fluxEvents = payload
  return state
}

export function stackConfig(state, env) {
  state.envs.forEach(stateEnv => {
    if (env.name === stateEnv.name) {
      stateEnv.stackConfig = env.stackConfig;
      stateEnv.stackDefinition = env.stackDefinition;
    }
  });

  return state;
}

export function envSpinnedOut(state, env) {
  state.envs = produce(state.envs, draft => {
    for (let e of draft) {
      if (e.name === env.name) {
        e.appsRepo = env.appsRepo
        e.infraRepo = env.infraRepo
        e.builtIn = env.builtIn
        break
      }
    }
  })
  return state;
}

export function agentEnvsUpdated(state, connectedAgents) {
  state.connectedAgents = produce(state.connectedAgents, draft => {
    connectedAgents.forEach((agent) => {
      draft[agent.name] = agent;
      if (!draft[agent.name].stacks) {
        draft[agent.name].stacks = []
      }
    });
  });
  return state
}

export function updateGitopsCommits(state, event) {
  let isPresent = false;

  state.gitopsCommits = produce(state.gitopsCommits, draft => {
    draft.forEach(gitopsCommit => {
      if (gitopsCommit.sha !== event.gitopsCommit.sha) {
        return
      }
      gitopsCommit.created = event.gitopsCommit.created;
      gitopsCommit.sha = event.gitopsCommit.sha;
      gitopsCommit.status = event.gitopsCommit.status;
      gitopsCommit.statusDesc = event.gitopsCommit.statusDesc;
      gitopsCommit.env = event.gitopsCommit.env;
      isPresent = true;
    });

    if (!isPresent) {
      draft.unshift(event.gitopsCommit);
    }  
  });

  for (const repo in state.rolloutHistory) {
    const repoObj = state.rolloutHistory[repo];
    for (const env in repoObj) {
      const envObj = repoObj[env];
      for (const appName in envObj) {
        envObj[appName].forEach(app => {
          if (app.gitopsRef === event.gitopsCommit.sha) {
            app.gitopsCommitStatus = event.gitopsCommit.status
            app.gitopsCommitStatusDesc = event.gitopsCommit.statusDesc
            app.gitopsCommitCreated = event.gitopsCommit.created
          }
        })
      }
    }
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
  state.application = {
    name: application.appName,
    appSettingsURL: application.appSettingsURL,
    installationURL: application.installationURL,
    dashboardVersion: application.dashboardVersion,
  };
  return state;
}

export function search(state, search) {
  state.search = search;
  return state;
}

export function rolloutHistory(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  const env = payload.env;
  const app = payload.app;
  const releases = payload.releases;

  if (!state.rolloutHistory[repo]) {
    state.rolloutHistory[repo] = {}
  }
  if (!state.rolloutHistory[repo][env]) {
    state.rolloutHistory[repo][env] = {}
  }

  state.rolloutHistory[repo][env][app] = releases
  return state;
}

export function updateCommits(state, payload) {
  const repo = `${payload.owner}/${payload.repo}`;
  const uniqueCommits = payload.commits.filter(commit => (
    !state.commits[repo].some(existingCommit => existingCommit.sha === commit.sha)
  ));

  state.commits[repo] = state.commits[repo].concat(uniqueCommits);
  return state;
}

export function fluxStateUpdated(state, event) {
  state.fluxState = produce(state.fluxState, draft => {
    draft[event.envName] = event.fluxState;
  });
  return state
}

export function fluxEventsUpdated(state, event) {
  if (state.fluxEvents[event.envName] === undefined) {
    return state;
  }

  state.fluxEvents = produce(state.fluxEvents, draft => {
    draft[event.envName] = event.fluxEvents;
  });
  return state
}

export function deploymentDetails(state, event) {
  state.details[event.deployment] = event.details;
  return state;
}

export function podDetails(state, event) {
  state.details[event.pod] = event.details;
  return state;
}

export function clearDetails(state) {
  state.details = {};
  return state;
}

export function updateCommitStatus(state, event) {
  const repo = `${event.owner}/${event.repo}`;

  if (!state.commits[repo]) {
    state.commits[repo] = [];
  }

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

export function alerts(state, alerts) {
  alerts.forEach(alert => {
    if (!state.alerts[alert.deploymentName]) {
      state.alerts[alert.deploymentName] = []
    }
    state.alerts[alert.deploymentName].push(alert)
  });
  return state;
}

export function alertPending(state, alert) {
  if (!state.alerts[alert.deploymentName]) {
    state.alerts[alert.deploymentName] = []
  }
  state.alerts[alert.deploymentName].push(alert);

  return state;
}

export function alertFired(state, alert) {
  if (!state.alerts[alert.deploymentName]) {
    state.alerts[alert.deploymentName] = []
  }
  state.alerts[alert.deploymentName].push(alert);

  return state;
}

export function alertResolved(state, alert) {
  if (state.alerts[alert.deploymentName] === undefined) {
    return state;
  }

  state.alerts[alert.deploymentName].forEach(a => {
    if (a.objectName === alert.objectName) {
      a.status = alert.status;
      a.resolvedAt = alert.resolvedAt;
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

export function settings(state, payload) {
  state.settings = produce(state.settings, draft => {
    draft.releaseHistorySinceDays = payload.releaseHistorySinceDays;
    draft.posthogFeatureFlag = payload.posthogFeatureFlag;
    draft.posthogApiKey = payload.posthogApiKey;
    draft.posthogIdentifyUser = payload.posthogIdentifyUser;
    draft.scmUrl = payload.scmUrl;
    draft.host = payload.host;
    draft.provider = payload.provider;
    draft.trial = payload.trial;
    draft.instance = payload.instance;
    draft.licensed = payload.licensed;
  })
  return state;
}

export function deploy(state, payload) {
  state.runningDeploy = payload;
  return state;
}

export function deployStatus(state, payload) {
  state.runningDeploy = produce(state.runningDeploy, draft => {
    draft.trackingId = payload.trackingId
    draft.status = payload.status
    draft.statusDesc = payload.statusDesc
    draft.results = payload.results
  })
  return state;
}

export function imageBuild(state, payload) {
  state.runningImageBuild = payload;
  return state;
}

export function imageBuildStatus(state, payload) {
  state.runningImageBuild.status = payload.status
  state.runningImageBuild.statusDesc = payload.statusDesc
  state.runningImageBuild.results = payload.results
  return state;
}

export function staleRepoData(state, event) {
  state.repoRefreshQueue.push(event.repo);
  return state;
}

export function commitEvent(state, event) {
  const sha = event.commitEvent.sha
  const eventId = event.commitEvent.id

  state.commitEvents = produce(state.commitEvents, draft => {
    if (!draft[sha]) {
      draft[sha] = {
        [eventId]: event.commitEvent,
        updated: Date.now()
      }
    } else {
      draft[sha] = {
        ...draft[sha],
        [eventId]: event.commitEvent,
        updated: Date.now()
      }
    }

    const anHourAgo = Date.now() - ONE_HOUR
    const recentShas = Object.keys(draft).filter(sha => draft[sha].updated > anHourAgo)
    Object.keys(draft).forEach(sha => {
      if (!recentShas.includes(sha)){
        console.log("filtering commit event: " + draft[sha])
        delete draft[sha]
      }
    })
  });

  return state;
}

const ONE_HOUR = 60 * 60 * 1000 /* ms */
