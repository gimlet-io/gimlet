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
        console.log(data);
        data.added ? setAdded(data.added) : setAdded([]);
        data.deleted ? setDeleted(data.deleted) : setDeleted([]);
        store.dispatch({
          type: ACTION_TYPE_GIT_REPOS, payload: data.repos
        });
        setReposLoading(false);
      }, () => {
        setReposLoading(false);
        /* Generic error handler deals with it */
      });
  }

  return (
    <div className="p-6 bg-white overflow-hidden shadow rounded-lg">
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
            <div>There are no repository changes.</div>
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
    <ul className={`mt-8 mx-8 grid grid-cols-3 list-disc gap-4 ${colors[color]} font-medium`}>
      {repos.map(repo => <li key={repo}>{repo}</li>)}
    </ul>
  )
};

export default RefreshRepos;
