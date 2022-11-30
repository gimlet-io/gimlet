import * as eventHandlers from './eventHandlers/eventHandlers';
import * as podEventHandlers from './eventHandlers/podEventHandlers';
import * as deploymentEventHandlers from './eventHandlers/deploymentEventHandlers';
import * as ingressEventHandlers from './eventHandlers/ingressEventHandlers';

export const ACTION_TYPE_STREAMING = 'streaming';
export const ACTION_TYPE_ENVS = 'envs';
export const ACTION_TYPE_USER = 'user';
export const ACTION_TYPE_USERS = 'users';
export const ACTION_TYPE_APPLICATION = 'application';
export const ACTION_TYPE_GIMLETD = 'gimletd';
export const ACTION_TYPE_CHARTSCHEMA = 'chartSchema';
export const ACTION_TYPE_SEARCH = 'search';
export const ACTION_TYPE_ROLLOUT_HISTORY = 'rolloutHistory';
export const ACTION_TYPE_RELEASE_STATUSES = 'releaseStatuses';
export const ACTION_TYPE_COMMITS = 'commits';
export const ACTION_TYPE_UPDATE_COMMITS = 'updateCommits';
export const ACTION_TYPE_BRANCHES = 'branches';
export const ACTION_TYPE_ENVCONFIGS = 'envConfigs';
export const ACTION_TYPE_SAVE_ENV_PULLREQUEST = 'updateEnvsPullRequestsOnSave';
export const ACTION_TYPE_SAVE_REPO_PULLREQUEST = 'updateReposPullRequestsOnSave';
export const ACTION_TYPE_ENV_PULLREQUESTS = 'envsPullRequestListUpdated';
export const ACTION_TYPE_REPO_PULLREQUESTS = 'reposPullRequestListUpdated';
export const ACTION_TYPE_ADD_ENVCONFIG = 'addEnvConfig';
export const ACTION_TYPE_REPO_METAS = "repoMetas";
export const ACTION_TYPE_DEPLOY = 'deploy';
export const ACTION_TYPE_DEPLOY_STATUS = 'deployStatus';
export const ACTION_TYPE_CLEAR_DEPLOY_STATUS = 'clearDeployStatus';
export const ACTION_TYPE_GITOPS_REPO = 'gitopsRepo';
export const ACTION_TYPE_GITOPS_COMMITS = 'gitopsCommits';
export const ACTION_TYPE_GIT_REPOS = 'gitRepos';
export const ACTION_TYPE_AGENTS = 'agents';
export const ACTION_TYPE_POPUPWINDOWPROGRESS = 'popupWindowProgress';
export const ACTION_TYPE_POPUPWINDOWERROR = 'popupWindowError';
export const ACTION_TYPE_POPUPWINDOWERRORLIST = 'popupWindowErrorList';
export const ACTION_TYPE_ENVUPDATED = 'envUpdated';
export const ACTION_TYPE_SETTINGS = 'settings';

export const ACTION_TYPE_POPUPWINDOWSUCCESS = 'popupWindowSaved';
export const ACTION_TYPE_POPUPWINDOWRESET = 'popupWindowReset';

export const ACTION_TYPE_OVERLAY = 'overlay';
export const ACTION_TYPE_OVERLAYRESET = 'overlayReset';


export const EVENT_AGENT_CONNECTED = 'agentConnected';
export const EVENT_AGENT_DISCONNECTED = 'agentDisconnected';
export const EVENT_ENVS_UPDATED = 'envsUpdated';
export const EVENT_STALE_REPO_DATA = 'staleRepoData';
export const EVENT_GITOPS_COMMIT_EVENT = 'gitopsCommit';
export const EVENT_COMMIT_STATUS_UPDATED = 'commitStatusUpdated';

export const EVENT_POD_CREATED = 'podCreated';
export const EVENT_POD_UPDATED = 'podUpdated';
export const EVENT_POD_DELETED = 'podDeleted';
export const EVENT_POD_LOGS = 'podLogs';

export const EVENT_DEPLOYMENT_CREATED = 'deploymentCreated';
export const EVENT_DEPLOYMENT_UPDATED = 'deploymentUpdated';
export const EVENT_DEPLOYMENT_DELETED = 'deploymentDeleted';

export const EVENT_INGRESS_CREATED = 'ingressCreated';
export const EVENT_INGRESS_UPDATED = 'ingressUpdated';
export const EVENT_INGRESS_DELETED = 'ingressDeleted';

export const initialState = {
  settings: {
    agents: []
  },
  connectedAgents: {},
  search: { filter: '' },
  rolloutHistory: {},
  releaseStatuses: {},
  commits: {},
  branches: {},
  pullRequests: {},
  runningDeploys: [],
  repoRefreshQueue: [],
  gitRepos: [],
  defaultChart: undefined,
  envConfigs: {},
  application: {},
  repoMetas: {},
  fileInfos: [],
  envs: [],
  gitopsCommits: [],
  popupWindow: {
    visible: false,
    finished: false,
    isError: false,
    header: "",
    message: "",
    link: "",
    errorList: null
  },
  overlay: {
    visible: false,
  },
  podLogs: "",
  users: []
};

export function rootReducer(state = initialState, action) {
  switch (action.type) {
    case ACTION_TYPE_STREAMING:
      return processStreamingEvent(state, action.payload)
    case ACTION_TYPE_GITOPS_REPO:
      return eventHandlers.gitopsRepo(state, action.payload);
    case ACTION_TYPE_GIT_REPOS:
      return eventHandlers.gitRepos(state, action.payload);
    case ACTION_TYPE_AGENTS:
      return eventHandlers.agents(state, action.payload);
    case ACTION_TYPE_POPUPWINDOWPROGRESS:
      return eventHandlers.popupWindowProgress(state, action.payload);
    case ACTION_TYPE_POPUPWINDOWERROR:
      return eventHandlers.popupWindowError(state, action.payload);
    case ACTION_TYPE_POPUPWINDOWERRORLIST:
      return eventHandlers.popupWindowErrorList(state, action.payload);
    case ACTION_TYPE_POPUPWINDOWSUCCESS:
      return eventHandlers.popupWindowSuccess(state, action.payload);
    case ACTION_TYPE_POPUPWINDOWRESET:
      return eventHandlers.popupWindowReset(state);
    case ACTION_TYPE_OVERLAY:
      return eventHandlers.overlay(state);
    case ACTION_TYPE_OVERLAYRESET:
      return eventHandlers.overlayReset(state);
    case ACTION_TYPE_ENVS:
      return eventHandlers.envsUpdated(state, action.payload)
    case ACTION_TYPE_GITOPS_COMMITS:
      return eventHandlers.gitopsCommits(state, action.payload)
    case ACTION_TYPE_USER:
      return eventHandlers.user(state, action.payload)
    case ACTION_TYPE_USERS:
      return eventHandlers.users(state, action.payload)
    case ACTION_TYPE_APPLICATION:
      return eventHandlers.application(state, action.payload)
    case ACTION_TYPE_GIMLETD:
      return eventHandlers.gimletd(state, action.payload)
    case ACTION_TYPE_CHARTSCHEMA:
      return eventHandlers.schemas(state, action.payload)
    case ACTION_TYPE_SEARCH:
      return eventHandlers.search(state, action.payload)
    case ACTION_TYPE_ROLLOUT_HISTORY:
      return eventHandlers.rolloutHistory(state, action.payload)
    case ACTION_TYPE_RELEASE_STATUSES:
      return eventHandlers.releaseStatuses(state, action.payload)
    case ACTION_TYPE_COMMITS:
      return eventHandlers.commits(state, action.payload)
      case ACTION_TYPE_UPDATE_COMMITS:
        return eventHandlers.updateCommits(state, action.payload)
    case ACTION_TYPE_BRANCHES:
      return eventHandlers.branches(state, action.payload)
    case ACTION_TYPE_ENVCONFIGS:
      return eventHandlers.envConfigs(state, action.payload)
    case ACTION_TYPE_ADD_ENVCONFIG:
      return eventHandlers.addEnvConfig(state, action.payload)
    case ACTION_TYPE_ENV_PULLREQUESTS:
      return eventHandlers.envPullRequests(state, action.payload)
    case ACTION_TYPE_REPO_PULLREQUESTS:
      return eventHandlers.repoPullRequests(state, action.payload)
    case ACTION_TYPE_SAVE_ENV_PULLREQUEST:
      return eventHandlers.saveEnvPullRequest(state, action.payload)
    case ACTION_TYPE_SAVE_REPO_PULLREQUEST:
      return eventHandlers.saveRepoPullRequest(state, action.payload)
    case ACTION_TYPE_REPO_METAS:
      return eventHandlers.repoMetas(state, action.payload)
      case ACTION_TYPE_SETTINGS:
        return eventHandlers.settings(state, action.payload)
    case ACTION_TYPE_DEPLOY:
      return eventHandlers.deploy(state, action.payload)
    case ACTION_TYPE_DEPLOY_STATUS:
      return eventHandlers.deployStatus(state, action.payload)
    case ACTION_TYPE_CLEAR_DEPLOY_STATUS:
      return eventHandlers.clearDeployStatus(state)
    case ACTION_TYPE_ENVUPDATED:
      return eventHandlers.envStackUpdated(state, action.name, action.payload)
    default:
      console.log('Could not process redux event: ' + JSON.stringify(action));
      return state;
  }
}

function processStreamingEvent(state, event) {
  console.log(event.event);

  switch (event.event) {
    case EVENT_AGENT_CONNECTED:
      return eventHandlers.agentConnected(state, event);
    case EVENT_AGENT_DISCONNECTED:
      return eventHandlers.agentDisconnected(state, event);
    case EVENT_ENVS_UPDATED:
      return eventHandlers.agentEnvsUpdated(state, event.envs);
    case EVENT_POD_CREATED:
      return podEventHandlers.podCreated(state, event);
    case EVENT_POD_UPDATED:
      return podEventHandlers.podUpdated(state, event);
    case EVENT_POD_DELETED:
      return podEventHandlers.podDeleted(state, event);
    case EVENT_POD_LOGS:
      return podEventHandlers.podLogs(state, event);
    case EVENT_DEPLOYMENT_CREATED:
      return deploymentEventHandlers.deploymentCreated(state, event);
    case EVENT_DEPLOYMENT_UPDATED:
      return deploymentEventHandlers.deploymentUpdated(state, event);
    case EVENT_DEPLOYMENT_DELETED:
      return deploymentEventHandlers.deploymentDeleted(state, event);
    case EVENT_INGRESS_CREATED:
      return ingressEventHandlers.ingressCreated(state, event);
    case EVENT_INGRESS_UPDATED:
      return ingressEventHandlers.ingressUpdated(state, event);
    case EVENT_INGRESS_DELETED:
      return ingressEventHandlers.ingressDeleted(state, event);
    case EVENT_STALE_REPO_DATA:
      return eventHandlers.staleRepoData(state, event);
    case EVENT_GITOPS_COMMIT_EVENT:
      return eventHandlers.updateGitopsCommits(state, event);
    case EVENT_COMMIT_STATUS_UPDATED:
      return eventHandlers.updateCommitStatus(state, event);
    default:
      console.log('Could not process streaming event: ' + JSON.stringify(event));
      return state;
  }
}
