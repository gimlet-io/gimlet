import React from "react";
import {EVENT_AGENT_CONNECTED, EVENT_AGENT_DISCONNECTED, initialState, rootReducer} from './redux';

function deepCopy(toCopy) {
  return JSON.parse(JSON.stringify(toCopy));
}

test('should process agent connected event', () => {
  const agentConnected = {
    type: 'streaming',
    payload: {
      event: EVENT_AGENT_CONNECTED,
      agent: {
        name: 'agent-a',
        namespace: '',
      }
    }
  };

  let reduced = rootReducer(initialState, agentConnected);

  expect(reduced.settings.agents.length).toEqual(1);
  expect(reduced.settings.agents[0]).toEqual({name: 'agent-a', namespace: ''});
});

test('should process agent disconnected event', () => {
  const agentDisconnected = {
    type: 'streaming',
    payload: {
      event: EVENT_AGENT_DISCONNECTED,
      agent: {
        name: 'agent-a',
        namespace: '',
      }
    }
  };

  const state = deepCopy(initialState)
  state.settings.agents.push({name: 'agent-a', namespace: ''})
  let reduced = rootReducer(state, agentDisconnected);

  expect(reduced.settings.agents.length).toEqual(0);
});
