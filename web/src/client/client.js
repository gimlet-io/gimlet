import axios from 'axios';

export default class GimletClient {
  constructor(onError) {
    this.onError = onError
  }

  URL = () => this.url;

  getApp = () => this.get('/api/app');

  getUser = () => this.get('/api/user');

  getUsers = () => this.get('/api/users');

  deleteUser = (userName) => this.post("/api/deleteUser", JSON.stringify(userName))

  saveUser = (userName) => this.post("/api/saveUser", JSON.stringify(userName));

  getAgents = () => this.get('/api/agents');

  getEnvs = () => this.get('/api/envs');

  getFluxEvents = () => this.get('/api/fluxEvents');

  getGitRepos = () => this.get('/api/gitRepos');

  searchRepo = (searchTerm, signal) => this.get(`/api/searchRepo?query=${searchTerm}`, signal);

  importRepo = (repoName) => this.get(`/api/importRepo?name=${repoName}`);

  getGitopsCommits = () => this.get("/api/gitopsCommits");

  getSettings = () => this.get("/api/settings");

  getChartUpdatePullRequests = () => this.get("/api/chartUpdatePullRequests");
  
  getRolloutHistoryPerApp = (owner, name, env, app) => this.get(`/api/releases?env=${env}&app=${app}&git-repo=${owner}/${name}&limit=10&reverse=true`);

  getReleases = (env, limit) => this.get(`/api/releases?env=${env}&limit=${limit}&reverse=true`);

  getCommits = (owner, name, branch, fromHash) => this.get(`/api/repo/${owner}/${name}/commits?branch=${branch}&fromHash=${fromHash}`);

  getCommitEvents = (owner, name, hash) => this.get(`/api/repo/${owner}/${name}/commits/${hash}/events`);

  triggerCommitSync = (owner, name) => this.get(`/api/repo/${owner}/${name}/triggerCommitSync`);

  getBranches = (owner, name) => this.get(`/api/repo/${owner}/${name}/branches`);

  getDefaultDeploymentTemplates = () => this.get(`/api/defaultDeploymentTemplates`);

  getDeploymentTemplates = (owner, name, env, config) => this.get(`/api/repo/${owner}/${name}/env/${env}/config/${config}/deploymentTemplates`);

  getEnvConfigs = (owner, name) => this.get(`/api/repo/${owner}/${name}/envConfigs`);

  saveEnvConfig = (owner, name, env, configName, configFile) => this.post(`/api/repo/${owner}/${name}/env/${env}/config/${configName}`, JSON.stringify({ ...configFile }));

  deleteEnvConfig = (owner, name, env, configName) => this.post(`/api/repo/${owner}/${name}/env/${env}/config/${configName}/delete`);

  getRepoMetas = (owner, name) => this.get(`/api/repo/${owner}/${name}/metas`);

  getConfigChangePullRequestsPerConfig = (owner, name, env, config) => this.get(`/api/repo/${owner}/${name}/env/${env}/config/${config}/configChangePullRequestsPerConfig`);

  getPullRequests = (owner, name) => this.get(`/api/repo/${owner}/${name}/pullRequests`);

  repoPullRequestsPolicy = (owner, repo) => this.get(`/api/repo/${owner}/${repo}/pullRequestPolicy`);

  saveRepoPullRequestPolicy = (owner, repo, pullRequestPolicy) => this.post(`/api/repo/${owner}/${repo}/saveRepoPullRequestPolicy`, JSON.stringify({ pullRequestPolicy }));

  getPullRequestsFromInfraRepo = (env) => this.get(`/api/infraRepoPullRequests?env=${env}`);

  getGitopsUpdatePullRequests = (env) => this.get(`/api/gitopsUpdatePullRequests?env=${env}`);

  podLogsRequest = (namespace, deployment, pod) => this.get(`/api/podLogs?namespace=${namespace}&deploymentName=${deployment}&podName=${pod}`);

  stopPodlogsRequest = (namespace, deployment) => this.get(`/api/stopPodLogs?namespace=${namespace}&deploymentName=${deployment}`);

  deploymentDetailsRequest = (namespace, name) => this.get(`/api/deploymentDetails?namespace=${namespace}&name=${name}`);

  podDetailsRequest = (namespace, name) => this.get(`/api/podDetails?namespace=${namespace}&name=${name}`);

  reconcile = (resource, namespace, name) => this.post(`/api/reconcile?resource=${resource}&namespace=${namespace}&name=${name}`);

  getAlerts = () => this.get("/api/alerts");

  stopEnv = () => this.get("/accounts/api/stopEnv");
  
  bootstrapGitops = (envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo) => this.post('/api/bootstrapGitops', JSON.stringify({ envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo }));

  saveEnvToDB = (envName) => this.post("/api/saveEnvToDB", JSON.stringify(envName));

  spinOutBuiltInEnv = () => this.post("/api/spinOutBuiltInEnv");

  deleteEnvFromDB = (envName) => this.post("/api/deleteEnvFromDB", JSON.stringify(envName));

  deleteAppInstance = (env, app) => this.post(`/api/delete?env=${env}&app=${app}`,)

  deploy = (artifactId, env, app) => this.post('/api/releases', JSON.stringify({ env, app, artifactId }));

  rollback = (env, app, rollbackTo) => this.post(`/api/rollback?env=${env}&app=${app}&sha=${rollbackTo}`);

  saveArtifact = (artifact) => this.post("/api/artifact", JSON.stringify(artifact));

  getDeployStatus = (trackingId) => this.get(`/api/eventReleaseTrack?id=${trackingId}`);

  saveFavoriteRepos = (favoriteRepos) => this.post('/api/saveFavoriteRepos', JSON.stringify({ favoriteRepos }));

  saveFavoriteServices = (favoriteServices) => this.post('/api/saveFavoriteServices', JSON.stringify({ favoriteServices }));

  saveInfrastructureComponents = (env, infrastructureComponents) => this.post('/api/environments', JSON.stringify({ env, infrastructureComponents }));

  seal = (env, secret) => this.plainPost(`/api/env/${env}/seal`, secret);

  getStackConfig = (env) =>  this.get(`/api/env/${env}/stackConfig`);

  silenceAlert = (object, until) => this.post(`/api/silenceAlert?object=${object}&&until=${until}`);

  restartDeploymentRequest = (namespace, name) =>this.post(`/api/restartDeployment?namespace=${namespace}&name=${name}`);

  getDependencyCatalog = () =>  this.get(`/api/dependencyCatalog`);

  get = async (path, signal) => {
    try {
      const { data } = await axios.get(path, {
        credentials: 'include',
        signal: signal
      });
      return data;
    } catch (error) {
      this.onError(error.response);
      throw error.response;
    }
  }

  post = async (path, body) => {
    try {
      const { data } = await axios
        .post(path, body, {
          credentials: 'include',
          headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
          },
        });
      return data;
    } catch (error) {
      this.onError(error.response);
      throw error.response;
    }
  }

  plainPost = async (path, body) => {
    try {
      const { data } = await axios
        .post(path, body, {
          credentials: 'include',
          headers: {
            Accept: "plain/text",
            "Content-Type": "plain/text",
          },
        });
      return data;
    } catch (error) {
      this.onError(error.response);
      throw error.response;
    }
  }
}
