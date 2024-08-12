import * as eventHandlers from './eventHandlers/eventHandlers';
import * as podEventHandlers from './eventHandlers/podEventHandlers';
import * as deploymentEventHandlers from './eventHandlers/deploymentEventHandlers';
import * as ingressEventHandlers from './eventHandlers/ingressEventHandlers';

export const ACTION_TYPE_STREAMING = 'streaming';
export const ACTION_TYPE_ENVS = 'envs';
export const ACTION_FLUX_EVENTS_RECEIVED = 'fluxEventsReceived';
export const ACTION_TYPE_STACK_CONFIG = 'stackConfig';
export const ACTION_TYPE_USER = 'user';
export const ACTION_TYPE_USERS = 'users';
export const ACTION_TYPE_APPLICATION = 'application';
export const ACTION_TYPE_SEARCH = 'search';
export const ACTION_TYPE_ROLLOUT_HISTORY = 'rolloutHistory';
export const ACTION_TYPE_BRANCHES = 'branches';
export const ACTION_TYPE_ENVCONFIGS = 'envConfigs';
export const ACTION_TYPE_ADD_ENVCONFIG = 'addEnvConfig';
export const ACTION_TYPE_REPO_METAS = "repoMetas";
export const ACTION_TYPE_DEPLOY = 'deploy';
export const ACTION_TYPE_CLEAR_DEPLOY = 'deploy';
export const ACTION_TYPE_DEPLOY_STATUS = 'deployStatus';
export const ACTION_TYPE_IMAGEBUILD = 'imageBuild';
export const ACTION_TYPE_IMAGEBUILD_STATUS = 'imageBuildStatus';
export const ACTION_TYPE_GITOPS_REPO = 'gitopsRepo';
export const ACTION_TYPE_GIT_REPOS = 'gitRepos';
export const ACTION_TYPE_POPUPWINDOWPROGRESS = 'popupWindowProgress';
export const ACTION_TYPE_POPUPWINDOWERROR = 'popupWindowError';
export const ACTION_TYPE_POPUPWINDOWERRORLIST = 'popupWindowErrorList';
export const ACTION_TYPE_ENVUPDATED = 'envUpdated';
export const ACTION_TYPE_SETTINGS = 'settings';
export const ACTION_TYPE_CLEAR_PODLOGS = 'clearPodLogs'
export const ACTION_TYPE_CLEAR_DETAILS = 'clearDetails'
export const ACTION_TYPE_ALERTS = 'alerts'

export const ACTION_TYPE_POPUPWINDOWSUCCESS = 'popupWindowSaved';
export const ACTION_TYPE_POPUPWINDOWRESET = 'popupWindowReset';

export const ACTION_TYPE_ENVSPINNEDOUT = 'envSpinnedOut';

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

export const EVENT_ALERT_PENDING = 'alertPending';
export const EVENT_ALERT_FIRED = 'alertFired';
export const EVENT_ALERT_RESOLVED = 'alertResolved'

export const EVENT_DEPLOYMENT_CREATED = 'deploymentCreated';
export const EVENT_DEPLOYMENT_UPDATED = 'deploymentUpdated';
export const EVENT_DEPLOYMENT_DELETED = 'deploymentDeleted';

export const EVENT_INGRESS_CREATED = 'ingressCreated';
export const EVENT_INGRESS_UPDATED = 'ingressUpdated';
export const EVENT_INGRESS_DELETED = 'ingressDeleted';

export const EVENT_IMAGE_BUILD_LOG_EVENT = 'imageBuildLogEvent';

export const EVENT_FLUX_STATE_UPDATED_EVENT = 'fluxStateUpdatedEvent';
export const EVENT_FLUX_EVENTS_UPDATED_EVENT = 'fluxK8sEventsUpdatedEvent';

export const EVENT_DEPLOYMENT_DETAILS_EVENT = 'deploymentDetailsEvent';
export const EVENT_POD_DETAILS_EVENT = 'podDetailsEvent';

export const EVENT_TYPE_COMMITEVENT = 'commitEvent';

export const initialState = {
  settings: {},
  connectedAgents: {},
  fluxState: {},
  search: { filter: '' },
  rolloutHistory: {},
  commits: {},
  branches: {},
  repoRefreshQueue: [],
  gitRepos: [],
  envConfigs: {},
  application: {},
  repoMetas: {},
  fileInfos: [],
  envs: [],
  gitopsCommits: [],
  alerts: {},
  popupWindow: {
    visible: false,
    finished: false,
    isError: false,
    header: "",
    message: "",
    link: "",
    errorList: null
  },
  podLogs: {},
  details: {},
  textColors: {},
  imageBuildLogs: {},
  users: [],
  commitEvents: {},
};

export function rootReducer(state = initialState, action) {
  switch (action.type) {
    case ACTION_TYPE_STREAMING:
      return processStreamingEvent(state, action.payload)
    case ACTION_TYPE_GIT_REPOS:
      return eventHandlers.gitRepos(state, action.payload);
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
    case ACTION_TYPE_ENVS:
      return eventHandlers.envs(state, action.payload)
    case ACTION_FLUX_EVENTS_RECEIVED:
        return eventHandlers.fluxEventsReceived(state, action.payload)
    case ACTION_TYPE_STACK_CONFIG:
      return eventHandlers.stackConfig(state, action.payload)
    case ACTION_TYPE_USER:
      return eventHandlers.user(state, action.payload)
    case ACTION_TYPE_USERS:
      return eventHandlers.users(state, action.payload)
    case ACTION_TYPE_APPLICATION:
      return eventHandlers.application(state, action.payload)
    case ACTION_TYPE_SEARCH:
      return eventHandlers.search(state, action.payload)
    case ACTION_TYPE_ROLLOUT_HISTORY:
      return eventHandlers.rolloutHistory(state, action.payload)
    case ACTION_TYPE_BRANCHES:
      return eventHandlers.branches(state, action.payload)
    case ACTION_TYPE_ALERTS:
      return eventHandlers.alerts(state, action.payload)
    case ACTION_TYPE_ENVCONFIGS:
      return eventHandlers.envConfigs(state, action.payload)
    case ACTION_TYPE_ADD_ENVCONFIG:
      return eventHandlers.addEnvConfig(state, action.payload)
    case ACTION_TYPE_REPO_METAS:
      return eventHandlers.repoMetas(state, action.payload)
    case ACTION_TYPE_SETTINGS:
        return eventHandlers.settings(state, action.payload)
    case ACTION_TYPE_DEPLOY:
      return eventHandlers.deploy(state, action.payload)
    case ACTION_TYPE_CLEAR_DEPLOY:
      delete state.runningDeploy
      return state
    case ACTION_TYPE_DEPLOY_STATUS:
      return eventHandlers.deployStatus(state, action.payload)
    case ACTION_TYPE_IMAGEBUILD:
      return eventHandlers.imageBuild(state, action.payload)
    case ACTION_TYPE_IMAGEBUILD_STATUS:
      return eventHandlers.imageBuildStatus(state, action.payload)
    case ACTION_TYPE_CLEAR_PODLOGS:
       return podEventHandlers.clearPodLogs(state, action.payload)
    case ACTION_TYPE_CLEAR_DETAILS:
      return eventHandlers.clearDetails(state, action.payload)
    case ACTION_TYPE_ENVSPINNEDOUT:
      return eventHandlers.envSpinnedOut(state, action.payload)
    default:
      if (action.type && action.type.startsWith('@@redux/INIT')) { // Ignoring Redux ActionTypes.INIT action
        return state
      }
      console.log('Could not process redux event: ' + JSON.stringify(action));
      return state;
  }
}

function processStreamingEvent(state, event) {
  // console.log(event.event);

  switch (event.event) {
    case EVENT_AGENT_CONNECTED:
      return eventHandlers.agentConnected(state, event);
    case EVENT_AGENT_DISCONNECTED:
      return eventHandlers.agentDisconnected(state, event);
    case EVENT_ENVS_UPDATED:
      return eventHandlers.agentEnvsUpdated(state, event.envs);
    case EVENT_ALERT_PENDING:
      return eventHandlers.alertPending(state, event.alert);
    case EVENT_ALERT_FIRED:
      return eventHandlers.alertFired(state, event.alert);
    case EVENT_ALERT_RESOLVED:
      return eventHandlers.alertResolved(state, event.alert);
    case EVENT_POD_CREATED:
      return podEventHandlers.podCreated(state, event);
    case EVENT_POD_UPDATED:
      return podEventHandlers.podUpdated(state, event);
    case EVENT_POD_DELETED:
      return podEventHandlers.podDeleted(state, event);
    case EVENT_POD_LOGS:
      return podEventHandlers.podLogs(state, event);
    case EVENT_IMAGE_BUILD_LOG_EVENT:
      return deploymentEventHandlers.imageBuildLogs(state, event);
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
    case EVENT_FLUX_STATE_UPDATED_EVENT:
      return eventHandlers.fluxStateUpdated(state, event);
    case EVENT_FLUX_EVENTS_UPDATED_EVENT:
      return eventHandlers.fluxEventsUpdated(state, event);
    case EVENT_DEPLOYMENT_DETAILS_EVENT:
      return eventHandlers.deploymentDetails(state, event);
    case EVENT_POD_DETAILS_EVENT:
      return eventHandlers.podDetails(state, event);
    case EVENT_TYPE_COMMITEVENT:
      return eventHandlers.commitEvent(state, event)
    default:
      if (event.type && event.type.startsWith('@@redux/INIT')) { // Ignoring Redux ActionTypes.INIT action
        return state
      }
      console.log('Could not process streaming event: ' + JSON.stringify(event));
      return state;
  }
}
