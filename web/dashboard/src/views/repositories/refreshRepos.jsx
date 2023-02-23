import { useState } from "react";
import { ACTION_TYPE_GIT_REPOS } from "../../redux/redux";
import { Spinner } from "./repositories";

const RefreshRepos = ({ gimletClient, store }) => {
  const [added, setAdded] = useState(null)
  const [deleted, setDeleted] = useState(null)
  const [reposLoading, setReposLoading] = useState(false)

  const refresh = () => {
    setReposLoading(true);
    gimletClient.refreshRepos()
      .then(data => {
        data.added ? setAdded(data.added) : setAdded([]);
        data.deleted ? setDeleted(data.deleted) : setDeleted([]);
        store.dispatch({
          type: ACTION_TYPE_GIT_REPOS, payload: data.userRepos
        });
        setReposLoading(false);
      }, () => {
        setReposLoading(false);
        /* Generic error handler deals with it */
      });
  }

  return (
    <div className="p-6 bg-white overflow-hidden shadow rounded-lg space-y-4">
      <button
        onClick={() => refresh()}
        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        Refresh repos
      </button>
      {reposLoading ?
        <Spinner />
        :
        <>
          <RenderRepos
            color="green"
            repos={added}
          />
          <RenderRepos
            color="red"
            repos={deleted}
          />
          {
            added?.length === 0 && deleted?.length === 0 &&
            <p className="text-sm text-gray-800">Currently there are no created or deleted repositories.</p>
          }
        </>
      }
    </div >)
};

const RenderRepos = ({ repos, color }) => {
  if (!repos) {
    return null
  }

  const colors = {
    green: "text-green-600",
    red: "text-red-600"
  }

  return (
    <ul className={`px-6 text-sm list-disc ${colors[color]} font-bold`}>
      {repos.map(repo => <li key={repo}>{repo}</li>)}
    </ul>
  )
};

export default RefreshRepos;
