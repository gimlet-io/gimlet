import {
  EVENT_AGENT_CONNECTED,
  EVENT_AGENT_DISCONNECTED,
  ACTION_TYPE_STREAMING,
  ACTION_TYPE_REPO_PULLREQUESTS,
  ACTION_TYPE_SAVE_REPO_PULLREQUEST,
  ACTION_TYPE_ENV_PULLREQUESTS,
  ACTION_TYPE_SAVE_ENV_PULLREQUEST,
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
  state.settings.agents.push({name: 'agent-a', namespace: ''})
  let reduced = rootReducer(state, agentDisconnected);

  expect(reduced.settings.agents.length).toEqual(0);
});

test('should set PR list', () => {
  const repo = "owner/repo"
  const prListUpdated = {
    type: ACTION_TYPE_REPO_PULLREQUESTS,
    payload: {
      data: {
        "staging": [
          {
            "sha": "abc123",
            "link": "http://doesnotexist"
          }
        ]
      },
      repoName: repo
    }
  };

  let reduced = rootReducer(initialState, prListUpdated);

  expect(Object.keys(reduced.pullRequests.configChanges[repo]["staging"]).length).toEqual(1);
  expect(reduced.pullRequests.configChanges[repo]["staging"][0].link).toEqual("http://doesnotexist");
});

test('should save single repo PR', () => {
  const repo = "owner/repo"
  const prSaved = {
    type: ACTION_TYPE_SAVE_REPO_PULLREQUEST,
    payload: {
      repoName: repo,
      envName: 'staging',
      createdPr: {
        "sha": "def456",
        "link": "http://alsodoesnotexist"
      }
    }
  };

  const state = deepCopy(initialState)
  state.pullRequests.configChanges[repo] = {
    "staging": [
      {
        "sha": "abc123",
        "link": "http://doesnotexist"
      }
    ]
  }
  let reduced = rootReducer(state, prSaved);

  expect(Object.keys(reduced.pullRequests.configChanges[repo]["staging"]).length).toEqual(2);
  expect(reduced.pullRequests.configChanges[repo]["staging"][1].link).toEqual("http://alsodoesnotexist");
});

test('should set infra PR list', () => {
  const prListUpdated = {
    type: ACTION_TYPE_ENV_PULLREQUESTS,
    payload: {
      "staging": [
        {
          "sha": "abc123",
          "link": "http://doesnotexist"
        }
      ]
    }
  };

  let reduced = rootReducer(initialState, prListUpdated);

  expect(reduced.envs[0].name).toEqual("staging");
  expect(reduced.envs[0].pullRequests.length).toEqual(1);
});

test('should save single infra PR', () => {
  const prCreated = {
    type: ACTION_TYPE_SAVE_ENV_PULLREQUEST,
    payload: {
      "envName": "staging",
      "createdPr": {
        "sha": "def456",
        "link": "http://alsodoesnotexist"
      },
      "stackConfig": {}
    }
  };

  const state = deepCopy(initialState)
  state.envs = [
    {
      name: "staging",
      pullRequests: [
        {
          "sha": "abc123",
          "link": "http://doesnotexist"
        }
      ],
      anotherField: "checking if it doesn't get lost"
    }
  ]
  
  let reduced = rootReducer(state, prCreated);

  expect(reduced.envs[0].name).toEqual("staging");
  expect(reduced.envs[0].anotherField).toEqual("checking if it doesn't get lost");
  expect(reduced.envs[0].pullRequests.length).toEqual(2);
  expect(reduced.envs[0].pullRequests[1].link).toEqual("http://alsodoesnotexist");
});

