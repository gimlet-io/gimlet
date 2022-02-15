export function ingressCreated(state, event) {
  const env = event.env;
  const namespace = event.subject.split('/')[0];
  const ingressName = event.subject.split('/')[1];

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
      return;
    }

    if (stack.ingresses === undefined) {
      stack.ingresses = [];
    }

    stack.ingresses.push({
      name: ingressName,
      namespace: namespace,
      url: event.url
    });
  });

  return state
}

export function ingressUpdated(state, event) {
  const env = event.env;

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (stack.service.namespace + '/' + stack.service.name !== event.svc) {
      return;
    }

    if (stack.ingresses === undefined) {
      return;
    }

    for (let i of stack.ingresses) {
      if (i.namespace + '/' + i.name === event.subject) {
        i.url = event.url;
      }
    };
  });

  return state
}

export function ingressDeleted(state, event) {
  const env = event.env;

  if (state.envs[env] === undefined) {
    return state;
  }

  state.envs[env].stacks.forEach((stack, stackID, stacks) => {
    if (!stack.ingresses) {
      return;
    }

    let filtered = stack.ingresses.filter((ingress) => ingress.namespace + '/' + ingress.name !== event.subject);
    stacks[stackID].ingresses = filtered;
  });

  return state
}
