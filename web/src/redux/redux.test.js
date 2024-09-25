import { expect, test } from 'vitest'
import {
  EVENT_AGENT_CONNECTED,
  EVENT_AGENT_DISCONNECTED,
  ACTION_TYPE_STREAMING,
  initialState,
  rootReducer
} from './redux';

function deepCopy(toCopy) {
  return JSON.parse(JSON.stringify(toCopy));
}

test('should process agent connected event', () => {
  const agentConnected = {
    type: ACTION_TYPE_STREAMING,
    payload: {
      event: EVENT_AGENT_CONNECTED,
      agent: {
        name: 'agent-a',
      }
    }
  };

  const reduced = rootReducer(initialState, agentConnected);

  expect(Object.keys(reduced.connectedAgents).length).toEqual(1);
  expect(reduced.connectedAgents['agent-a']).toEqual({name: 'agent-a', stacks: []});
});

test('should process agent disconnected event', () => {
  const agentDisconnected = {
    type: ACTION_TYPE_STREAMING,
    payload: {
      event: EVENT_AGENT_DISCONNECTED,
      agent: {
        name: 'agent-a',
        namespace: '',
      }
    }
  };

  const state = deepCopy(initialState)
  state.connectedAgents['agent-a'] = {name: 'agent-a', namespace: ''}
  const reduced = rootReducer(state, agentDisconnected);

  expect(Object.keys(reduced.connectedAgents).length).toEqual(0);
});
