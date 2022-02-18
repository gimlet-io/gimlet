const axios = require('axios').default;

export default class GimletClient {
  constructor(onError) {
    this.onError = onError
  }

  URL = () => this.url;

  getApp = () => this.get('/api/app');

  getUser = () => this.get('/api/user');

  getGitopsRepo = () => this.get('/api/gitopsRepo');

  getAgents = () => this.get('/api/agents');

  getEnvs = () => this.get('/api/envs');

  getGitRepos = () => this.get('/api/gitRepos');

  getChartSchema = () => this.get('/api/chartSchema');

  getGimletD = () => this.get('/api/gimletd');

  getRolloutHistory = (owner, name) => this.get(`/api/repo/${owner}/${name}/rolloutHistory`);

  getCommits = (owner, name, branch) => this.get(`/api/repo/${owner}/${name}/commits?branch=${branch}`);

  getBranches = (owner, name) => this.get(`/api/repo/${owner}/${name}/branches`);

  getEnvConfigs = (owner, name) => this.getWithAxios(`/api/repo/${owner}/${name}/envConfigs`);

  saveEnvConfig = (owner, name, env, configName, config) => this.postWithAxios(`/api/repo/${owner}/${name}/env/${env}/config/${configName}`, JSON.stringify(config));

  saveEnvToDB = (envName) => this.postWithAxios("/api/saveEnvToDB", JSON.stringify(envName));

  deleteEnvFromDB = (envName) => this.postWithAxios("/api/deleteEnvFromDB", JSON.stringify(envName));

  deploy = (artifactId, env, app) => this.post('/api/deploy', JSON.stringify({ env, app, artifactId }));

  rollback = (env, app, rollbackTo) => this.post('/api/rollback', JSON.stringify({ env, app, targetSHA: rollbackTo }));

  getDeployStatus = (trackingId) => this.get(`/api/deployStatus?trackingId=${trackingId}`);

  saveFavoriteRepos = (favoriteRepos) => this.post('/api/saveFavoriteRepos', JSON.stringify({ favoriteRepos }));

  saveFavoriteServices = (favoriteServices) => this.post('/api/saveFavoriteServices', JSON.stringify({ favoriteServices }));

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
