import {
  ACTION_TYPE_DEPLOY,
  ACTION_TYPE_DEPLOY_STATUS,
  ACTION_TYPE_IMAGEBUILD,
  ACTION_TYPE_IMAGEBUILD_STATUS,
  ACTION_TYPE_ROLLOUT_HISTORY,
} from "./redux/redux";

export default class DeployHandler {
  constructor(owner, repo, gimletClient, store) {
    this.owner = owner
    this.repo = repo
    this.gimletClient = gimletClient
    this.store = store
  }

  deploy = (target, sha, repo) => {
    this.gimletClient.deploy(target.artifactId, target.env, target.app)
      .then(data => {
        const trackingId = data.id
        if (data.type === 'imageBuild') {
          this.store.dispatch({
            type: ACTION_TYPE_IMAGEBUILD, payload: {
              trackingId: trackingId
            }
          });
          this.store.dispatch({
            type: ACTION_TYPE_DEPLOY, payload: {
              repo: repo,
              env: target.env,
              app: target.app,
              sha: sha
            }
          });
          setTimeout(() => {
            this.checkImageBuildStatus(trackingId);
          }, 2000);
        } else {
          this.store.dispatch({ type: ACTION_TYPE_IMAGEBUILD, payload: undefined });
          this.store.dispatch({
            type: ACTION_TYPE_DEPLOY, payload: {
              trackingId: trackingId,
              repo: repo,
              env: target.env,
              app: target.app,
              sha: sha
            }
          });
          setTimeout(() => {
            this.checkDeployStatus(trackingId);
          }, 1000);
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  checkImageBuildStatus = (trackingId) => {
    this.gimletClient.getDeployStatus(trackingId)
      .then(data => {
        const triggeredDeployRequestID = data.results && data.results.length > 0 ? data.results[0].triggeredDeployRequestID : undefined
        this.store.dispatch({
          type: ACTION_TYPE_IMAGEBUILD_STATUS, payload: {
            triggeredDeployRequestID: triggeredDeployRequestID,
            status: data.status,
            statusDesc: data.statusDesc,
            results: data.results,
          }
        });

        if (data.status === "new") {
          setTimeout(() => {
            this.checkImageBuildStatus(trackingId);
          }, 2000);
        }

        if (data.type === "imageBuild" && data.status === "success") {
          const triggeredReleaseId = data.results[0].triggeredDeployRequestID
          this.checkDeployStatus(triggeredReleaseId);
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  checkDeployStatus(trackingId) {
    this.gimletClient.getDeployStatus(trackingId)
      .then(data => {
        this.store.dispatch({
          type: ACTION_TYPE_DEPLOY_STATUS, payload: {
            trackingId: trackingId,
            status: data.status,
            statusDesc: data.statusDesc,
            results: data.results,
          }
        });

        if (data.status === "new") {
          setTimeout(() => {
            this.checkDeployStatus(trackingId);
          }, 1000);
        }

        if (data.status === "processed") {
          let gitopsCommitsApplied = true;
          const numberOfResults = data.results.length;

          if (numberOfResults > 0) {
            const latestGitopsHashMetadata = data.results[0];
            if (latestGitopsHashMetadata.gitopsCommitStatus === "N/A") { // poll until all gitops writes are applied
              gitopsCommitsApplied = false;
              setTimeout(() => {
                this.checkDeployStatus(trackingId);
              }, 1000);
            }
          }
          if (gitopsCommitsApplied) {
            for (const result of data.results) {
              setTimeout(() => {
                this.gimletClient.getRolloutHistoryPerApp(this.owner, this.repo, result.env, result.app)
                  .then(data => {
                    this.store.dispatch({
                      type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
                        owner: this.owner,
                        repo: this.repo,
                        env: result.env,
                        app: result.app,
                        releases: data,
                      }
                    });
                  }, () => {/* Generic error handler deals with it */ }
                  );
              }, 300);
            }
          }
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  rollback(env, app, rollbackTo, e) {
    const target = {rollback: true, app: app, env: env};
    this.gimletClient.rollback(env, app, rollbackTo)
      .then(data => {
        const trackingId = data.id;
        this.store.dispatch({
          type: ACTION_TYPE_DEPLOY, payload: {
            rollback: true,
            trackingId: trackingId,
            // repo: repo,
            env: target.env,
            app: target.app,
          }
        });
        setTimeout(() => {
          this.checkDeployStatus(trackingId);
        }, 1000);
      }, () => {/* Generic error handler deals with it */
      });
  }
}