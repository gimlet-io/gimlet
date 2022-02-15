export function deploymentCreated(state, event) {
  const env = event.env;
  const namespace = event.subject.split('/')[0];
  const deploymentName = event.subject.split('/')[1];

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
      return;
    }

    if (stack.deployment === undefined) {
      stacks[stackID].deployment = {
        name: deploymentName,
        namespace: namespace,
        pods: [],
        sha: event.sha
      };
    }
  });

  return state
}

export function deploymentUpdated(state, event) {
  const env = event.env;

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
      return;
    }

    if (stack.deployment && stack.deployment.namespace + '/' + stack.deployment.name === event.subject) {
      stacks[stackID].deployment.sha = event.sha;
      stacks[stackID].deployment.commitMessage = event.commitMessage;
    }
  });

  return state
}

export function deploymentDeleted(state, event) {
  const env = event.env;

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (stack.deployment && stack.deployment.namespace + '/' + stack.deployment.name === event.subject) {
      delete stacks[stackID].deployment;
    }
  });

  return state
}
