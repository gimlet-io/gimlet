const axios = require('axios').default;

export default class GimletClient {
  constructor(onError) {
    this.onError = onError
  }

  URL = () => this.url;

  getApp = () => this.get('/api/app');

  getUser = () => this.get('/api/user');

  getUsers = () => this.get('/api/users');

  deleteUser = (userName) => this.postWithAxios("/api/deleteUser", JSON.stringify(userName))

  saveUser = (userName) => this.postWithAxios("/api/saveUser", JSON.stringify(userName));

  getAgents = () => this.get('/api/agents');

  getEnvs = () => this.get('/api/envs');

  getGitRepos = () => this.get('/api/gitRepos');

  refreshRepos = () => this.get('/api/refreshRepos');

  getGitopsCommits = () => this.getWithAxios("/api/gitopsCommits");

  getSettings = () => this.getWithAxios("/api/settings");

  getChartUpdatePullRequests = () => this.getWithAxios("/api/chartUpdatePullRequests");
  
  getRolloutHistoryPerApp = (owner, name, env, app) => this.get(`/api/releases?env=${env}&app=${app}&git-repo=${owner}/${name}&limit=10&reverse=true`);

  getReleases = (env, limit) => this.getWithAxios(`/api/releases?env=${env}&limit=${limit}&reverse=true`);

  getCommits = (owner, name, branch, fromHash) => this.get(`/api/repo/${owner}/${name}/commits?branch=${branch}&fromHash=${fromHash}`);

  getBranches = (owner, name) => this.get(`/api/repo/${owner}/${name}/branches`);

  getChartSchema = (owner, name, env) => this.get(`/api/repo/${owner}/${name}/env/${env}/chartSchema`);

  getEnvConfigs = (owner, name) => this.getWithAxios(`/api/repo/${owner}/${name}/envConfigs`);

  saveEnvConfig = (owner, name, env, configName, values, namespace, chart, appName, useDeployPolicy, deployBranch, deployTag, deployEvent) => this.postWithAxios(`/api/repo/${owner}/${name}/env/${env}/config/${configName}`, JSON.stringify({ values, namespace, chart, appName, useDeployPolicy, deployBranch, deployTag, deployEvent }));

  getRepoMetas = (owner, name) => this.getWithAxios(`/api/repo/${owner}/${name}/metas`);

  getPullRequests = (owner, name) => this.getWithAxios(`/api/repo/${owner}/${name}/pullRequests`);

  getPullRequestsFromInfraRepo = () => this.getWithAxios(`/api/infraRepoPullRequests`);

  podLogsRequest = (namespace, serviceName) => this.getWithAxios(`/api/podLogs?namespace=${namespace}&serviceName=${serviceName}`);

  stopPodlogsRequest = (namespace, serviceName) => this.getWithAxios(`/api/stopPodLogs?namespace=${namespace}&serviceName=${serviceName}`);

  getAlerts = () => this.getWithAxios("/api/alerts");
  
  bootstrapGitops = (envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo) => this.postWithAxios('/api/bootstrapGitops', JSON.stringify({ envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo }));

  saveEnvToDB = (envName) => this.postWithAxios("/api/saveEnvToDB", JSON.stringify(envName));

  deleteEnvFromDB = (envName) => this.postWithAxios("/api/deleteEnvFromDB", JSON.stringify(envName));

  deploy = (artifactId, env, app, tenant) => this.post('/api/releases', JSON.stringify({ env, app, artifactId, tenant }));

  rollback = (env, app, rollbackTo) => this.post(`/api/rollback?env=${env}&app=${app}&sha=${rollbackTo}`);

  getDeployStatus = (trackingId) => this.get(`/api/eventReleaseTrack?id=${trackingId}`);

  saveFavoriteRepos = (favoriteRepos) => this.post('/api/saveFavoriteRepos', JSON.stringify({ favoriteRepos }));

  saveFavoriteServices = (favoriteServices) => this.post('/api/saveFavoriteServices', JSON.stringify({ favoriteServices }));

  saveInfrastructureComponents = (env, infrastructureComponents) => this.postWithAxios('/api/environments', JSON.stringify({ env, infrastructureComponents }));

  getWithAxios = async (path) => {
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

  postWithAxios = async (path, body) => {
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

  get = (path) => fetch(path, {
    credentials: 'include'
  })
    .then(response => {
      if (!response.ok && window !== undefined) {
        return Promise.reject({ status: response.status, statusText: response.statusText, path });
      }
      return response.json();
    })
    .catch((error) => {
      this.onError(error);
      throw error;
    });

  post = (path, body) => fetch(path, {
    method: 'post',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body
  })
    .then(response => {
      if (!response.ok && window !== undefined) {
        return Promise.reject({ status: response.status, statusText: response.statusText, path });
      }
      return response.json();
    })
    .catch((error) => {
      this.onError(error);
      throw error;
    });

  postWithoutCreds = (path, body) => fetch(path, {
    method: 'post',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body
  })
    .then(response => {
      if (!response.ok && window !== undefined) {
        return Promise.reject({ status: response.status, statusText: response.statusText, path });
      }
      return response.json();
    })
    .catch((error) => {
      this.onError(error);
      throw error;
    })
}
