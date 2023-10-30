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

  getGitRepos = () => this.get('/api/gitRepos');

  refreshRepos = () => this.get('/api/refreshRepos');

  getGitopsCommits = () => this.get("/api/gitopsCommits");

  getSettings = () => this.get("/api/settings");

  getChartUpdatePullRequests = () => this.get("/api/chartUpdatePullRequests");
  
  getRolloutHistoryPerApp = (owner, name, env, app) => this.get(`/api/releases?env=${env}&app=${app}&git-repo=${owner}/${name}&limit=10&reverse=true`);

  getReleases = (env, limit) => this.get(`/api/releases?env=${env}&limit=${limit}&reverse=true`);

  getCommits = (owner, name, branch, fromHash) => this.get(`/api/repo/${owner}/${name}/commits?branch=${branch}&fromHash=${fromHash}`);

  triggerCommitSync = (owner, name) => this.get(`/api/repo/${owner}/${name}/triggerCommitSync`);

  getBranches = (owner, name) => this.get(`/api/repo/${owner}/${name}/branches`);

  getDefaultDeploymentTemplates = () => this.get(`/api/defaultDeploymentTemplates`);

  getDeploymentTemplates = (owner, name, env, config) => this.get(`/api/repo/${owner}/${name}/env/${env}/config/${config}/deploymentTemplates`);

  getEnvConfigs = (owner, name) => this.get(`/api/repo/${owner}/${name}/envConfigs`);

  saveEnvConfig = (owner, name, env, configName, values, namespace, chart, appName, useDeployPolicy, deployBranch, deployTag, deployEvent) => this.post(`/api/repo/${owner}/${name}/env/${env}/config/${configName}`, JSON.stringify({ values, namespace, chart, appName, useDeployPolicy, deployBranch, deployTag, deployEvent }));

  deleteEnvConfig = (owner, name, env, configName) => this.post(`/api/repo/${owner}/${name}/env/${env}/config/${configName}/delete`);

  getRepoMetas = (owner, name) => this.get(`/api/repo/${owner}/${name}/metas`);

  getPullRequests = (owner, name) => this.get(`/api/repo/${owner}/${name}/pullRequests`);

  getPullRequestsFromInfraRepo = () => this.get(`/api/infraRepoPullRequests`);

  getGitopsUpdatePullRequests = () => this.get("/api/gitopsUpdatePullRequests");

  podLogsRequest = (namespace, serviceName) => this.get(`/api/podLogs?namespace=${namespace}&serviceName=${serviceName}`);

  stopPodlogsRequest = (namespace, serviceName) => this.get(`/api/stopPodLogs?namespace=${namespace}&serviceName=${serviceName}`);

  deploymentDetailsRequest = (namespace, serviceName) => this.get(`/api/deploymentDetails?namespace=${namespace}&serviceName=${serviceName}`);

  getAlerts = () => this.get("/api/alerts");
  
  bootstrapGitops = (envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo) => this.post('/api/bootstrapGitops', JSON.stringify({ envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo }));

  saveEnvToDB = (envName) => this.post("/api/saveEnvToDB", JSON.stringify(envName));

  spinOutBuiltInEnv = () => this.post("/api/spinOutBuiltInEnv");

  deleteEnvFromDB = (envName) => this.post("/api/deleteEnvFromDB", JSON.stringify(envName));

  deleteAppInstance = (env, app) => this.post(`/api/delete?env=${env}&app=${app}`,)

  deploy = (artifactId, env, app, tenant) => this.post('/api/releases', JSON.stringify({ env, app, artifactId, tenant }));

  rollback = (env, app, rollbackTo) => this.post(`/api/rollback?env=${env}&app=${app}&sha=${rollbackTo}`);

  getDeployStatus = (trackingId) => this.get(`/api/eventReleaseTrack?id=${trackingId}`);

  saveFavoriteRepos = (favoriteRepos) => this.post('/api/saveFavoriteRepos', JSON.stringify({ favoriteRepos }));

  saveFavoriteServices = (favoriteServices) => this.post('/api/saveFavoriteServices', JSON.stringify({ favoriteServices }));

  saveInfrastructureComponents = (env, infrastructureComponents) => this.post('/api/environments', JSON.stringify({ env, infrastructureComponents }));

  get = async (path) => {
    try {
      const { data } = await axios.get(path, {
        credentials: 'include'
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
}
