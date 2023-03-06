import { Spinner } from "./repositories";

const RefreshRepos = ({ added, deleted, repositoriesRefreshing, installationURL }) => {
  return (
    <div className="p-6 bg-white overflow-hidden shadow rounded-lg space-y-4">
      <div>
        {repositoriesRefreshing ?
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
            <div className="text-sm text-gray-800">
              { added?.length === 0 && deleted?.length === 0 &&
              <p className="mb-4">No new repositories found.</p>
              }
              <p className="text-xs">Missing a repository?<br />You may need to grant Gimlet access to more repositories on the <a
                href={installationURL}
                rel="noreferrer"
                target="_blank"
                className="text-blue-700 hover:text-blue-500 cursor-pointer">Github application installation settings</a> page.</p>
            </div>
          </>
        }
      </div>
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
