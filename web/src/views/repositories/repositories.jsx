import RepoCard from "../../components/repoCard/repoCard";
import { useState, useEffect } from 'react';
import {
  ACTION_TYPE_GIT_REPOS
} from "../../redux/redux";
import { InformationCircleIcon } from '@heroicons/react/20/solid'
import FilterBar from '../filterBar/filterBar';
import { useNavigate } from 'react-router-dom'

export default function Repositories (props) {
  let navigate = useNavigate()
  const { store, gimletClient } = props;
  const reduxState = store.getState();

  const [favorites, setFavorites] = useState(reduxState.user?.favoriteRepos)
  const [repositories, setRepositories] = useState(reduxState.gitRepos)
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [chartUpdatePullRequests, setChartUpdatePullRequests] = useState()
  const [repositoriesLoading, setRepositoriesLoading] = useState(false)
  const [settings, setSettings] = useState(reduxState.settings)
  const [filters, setFilters] = useState(JSON.parse(localStorage.getItem("filters")) ?? [])

  store.subscribe(() => {
    const reduxState = store.getState()
    setFavorites(reduxState.user?.favoriteRepos)
    setRepositories(reduxState.gitRepos)
    setConnectedAgents(reduxState.connectedAgents)
    setSettings(reduxState.settings)
  })

  useEffect(() => {
    if (repositories.length === 0) {
      setRepositoriesLoading(true)
      gimletClient.getGitRepos()
        .then(repos => {
          store.dispatch({
            type: ACTION_TYPE_GIT_REPOS, payload: repos
          })
          setRepositoriesLoading(false)

          if (repos.length === 0) {
            navigate("/import-repositories")
          }
        }, () => {
          setRepositoriesLoading(false)
        });
    }
    gimletClient.getChartUpdatePullRequests()
      .then(data => {
        setChartUpdatePullRequests(data)
      }, () => {/* Generic error handler deals with it */
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    localStorage.setItem("filters", JSON.stringify(filters));
  }, [filters]);

  const favoriteHandler = (repo) => {
    let deepCopied = JSON.parse(JSON.stringify(favorites))

    if (!favorites.includes(repo)) {
      deepCopied.push(repo);
    } else {
      deepCopied = deepCopied.filter(fav => fav !== repo);
    }

    setFavorites(deepCopied)
    gimletClient.saveFavoriteRepos(deepCopied);
  }

  if (!settings.provider || settings.provider === "") {
    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
            <div>
              <div className="mt-8">
                <SetupGithubCard />
              </div>
            </div>
          </div>
        </header>
      </div>
    )
  }

  const services = servicesPerRepo(connectedAgents, repositories)
  const filteredRepositories = filterRepos(repositories, services, favorites, filters)
  filteredRepositories.sort((a,b) => a.name - b.name);

  return (
    <div>
      <header>
        <div className="max-w-7xl mx-auto pt-24 px-4 sm:px-6 lg:px-8">
          <div className='flex items-center space-x-4'>
            <FilterBar
              properties={["Repository", "Service", "Namespace", "Owner", "Starred", "Domain"]}
              filters={filters}
              change={setFilters}
            />
            <button
              onClick={() => navigate("/import-repositories")}
              className="primaryButton px-8">
              Import
            </button>
          </div>
          {renderChartUpdatePullRequests(chartUpdatePullRequests)}
        </div>
      </header>
      <main>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div>
            <h4 className="text-lg font-base capitalize leading-tight mt-8 mb-5 pl-1">Repositories</h4>
            <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              { repositoriesLoading && <SkeletonLoader /> }
              { !repositoriesLoading && filteredRepositories.length === 0 && <EmptyStateNoMatchingService /> }
              { !repositoriesLoading && filteredRepositories.map(repo =>
                <li key={repo} className="card">
                  <RepoCard
                    name={repo}
                    services={services[repo]}
                    navigateToRepo={() => navigate(`/repo/${repo}`)}
                    favorite={favorites.includes(repo)}
                    favoriteHandler={favoriteHandler}
                  />
                </li>
              )}
            </ul>
          </div>
        </div>
      </main>
    </div>
  )
}

const filterRepos = (repos, services, favorites, filters) => {
  let filteredRepositories = [...repos];
  filters.forEach(filter => {
    switch (filter.property) {
      case 'Repository':
        filteredRepositories = filteredRepositories.filter(repo => repo.includes(filter.value))
        break;
      case 'Service':
        filteredRepositories = filteredRepositories.filter(repo => {
          return services[repo].length !== 0 && services[repo].some(service => service.service.name.includes(filter.value));
        })
        break;
      case 'Namespace':
        filteredRepositories = filteredRepositories.filter(repo => {
          return services[repo].length !== 0 && services[repo].some(service => service.service.namespace.includes(filter.value));
        })
        break;
      case 'Owner':
        filteredRepositories = filteredRepositories.filter(repo => {
          return services[repo].length !== 0 && services[repo].some(service => {
            return service.osca && service.osca.owner.includes(filter.value)
          });
        })
        break;
      case 'Starred':
        filteredRepositories = filteredRepositories.filter(repo => favorites.includes(repo))
        break;
      case 'Domain':
        filteredRepositories = filteredRepositories.filter(repo => {
          return services[repo].length !== 0 && services[repo].some(service => service.ingresses !== undefined && service.ingresses.some(ingress => ingress.url.includes(filter.value)));
        })
        break;
      default:
    }
  })
  return filteredRepositories;
}

const SetupGithubCard = () => {
  const navigate = useNavigate()
  return (
    <div className="rounded-md bg-blue-50 p-4 mb-4">
    <div className="flex">
      <div className="flex-shrink-0">
        <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-medium text-blue-800">Integrate Github</h3>
        <div className="mt-2 text-sm text-blue-700">
          This view will load your git repositories once you integrated Github.<br />
          <button
            className="font-medium"
            onClick={() => {navigate("/settings");return true}}
          >
            Click to integrate Github on the Settings page.
          </button>
        </div>
      </div>
    </div>
    </div>
  );
}

export const SkeletonLoader = () => {
  return (
    <>
      <li className="animate-pulse card">
        <div className="w-full flex items-center justify-between p-6 space-x-6">
          <div className="flex-1">
            <div className="w-full max-w-4xl py-1 animate-pulse">
              <div className="h-2 bg-neutral-300 dark:bg-neutral-500 py-1 animate-pulse rounded w-2/5"></div>
            </div>
            <div className="p-2 space-y-2"></div>
          </div>
        </div>
      </li>
      <li className="animate-pulse card">
        <div className="w-full flex items-center justify-between p-6 space-x-6">
          <div className="flex-1">
            <div className="w-full max-w-4xl py-1 animate-pulse">
              <div className="h-2 bg-neutral-300 dark:bg-neutral-500 py-1 animate-pulse rounded w-2/5"></div>
            </div>
            <div className="p-2 space-y-2"></div>
          </div>
        </div>
      </li>
    </>
  )
}

function servicesPerRepo(connectedAgents, gitRepos) {
  if (!connectedAgents) {
    return {};
  }

  const perRepo = {}
  gitRepos.forEach(repo => {
    const services = Object.values(connectedAgents)
      .flatMap(a => a.stacks ? a.stacks : [])
      .filter(stack => stack.repo === repo)
    perRepo[repo] = services
  })

  return perRepo
}

function EmptyStateNoMatchingService (props) {
  return (
    <p className="text-base text-neutral-800 dark:text-neutral-300">No service matches the search</p>
  )
}

function renderChartUpdatePullRequests(chartUpdatePullRequests) {
  if (!chartUpdatePullRequests || JSON.stringify(chartUpdatePullRequests) === "{}") {
    return null
  }

  const prList = [];
  for (const [repoName, pullRequests] of Object.entries(chartUpdatePullRequests)) {
    pullRequests.forEach(p => {
      prList.push(
        <li key={p.sha}>
          <a href={p.link} target="_blank" rel="noopener noreferrer">
            <span className="font-medium">{repoName}</span>: {p.title}
          </a>
        </li>)
    })
  }

  return (
    <div className="rounded-md bg-blue-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
        </div>
        <div className="ml-3 flex-1 text-blue-700 md:flex md:justify-between">
          <div className="text-xs flex flex-col">
            <span className="font-semibold text-sm">Helm chart version updates:</span>
            <ul className="list-disc list-inside text-xs ml-2">
              {prList}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
