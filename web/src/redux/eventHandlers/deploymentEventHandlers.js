import {produce} from 'immer';

export function deploymentCreated(state, event) {
  const env = event.env;
  const namespace = event.subject.split('/')[0];
  const deploymentName = event.subject.split('/')[1];

  if (state.connectedAgents[env] === undefined) {
    return state;
  }

  state.connectedAgents = produce(state.connectedAgents, draft => {
    if (!draft[env].stacks) {
      draft[env].stacks = []
    }

    draft[env].stacks.forEach((stack, stackID, stacks) => {
      if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
        return;
      }

      if (stack.deployment === undefined) {
        stacks[stackID].deployment = {
          name: deploymentName,
          namespace: namespace,
          pods: [],
          branch: event.branch,
          sha: event.sha
        };
      }
    });
  });
  return state
}

export function deploymentUpdated(state, event) {
  const env = event.env;

  if (state.connectedAgents[env] === undefined) {
    return state;
  }

  state.connectedAgents = produce(state.connectedAgents, draft => {
    draft[env].stacks.forEach((stack, stackID, stacks) => {
      if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
        return;
      }

      if (stack.deployment && stack.deployment.namespace + '/' + stack.deployment.name === event.subject) {
        stacks[stackID].deployment.sha = event.sha;
        stacks[stackID].deployment.branch = event.branch;
        stacks[stackID].deployment.commitMessage = event.commitMessage;
      }
    });
  });
  return state
}

export function deploymentDeleted(state, event) {
  const env = event.env;

  if (state.connectedAgents[env] === undefined) {
    return state;
  }

  state.connectedAgents = produce(state.connectedAgents, draft => {
    draft[env].stacks.forEach((stack, stackID, stacks) => {
      if (stack.deployment && stack.deployment.namespace + '/' + stack.deployment.name === event.subject) {
        delete stacks[stackID].deployment;
      }
    });
  });
  return state
}

export function imageBuildLogs(state, event) {
  if (!state.imageBuildLogs[event.buildId]) {
    state.imageBuildLogs[event.buildId] = {
      status: event.status,
      logLines: [],
    };
  }

  state.imageBuildLogs[event.buildId].status = event.status;

  if (event.logLine) {
    state.imageBuildLogs[event.buildId].logLines.push(event.logLine);
  }
  return state;
}
