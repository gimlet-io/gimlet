import React, { Component, useEffect } from 'react';
import { useState, useRef } from 'react';
import RepoCard from "../../components/repoCard/repoCard";
import { emptyStateNoMatchingService } from "../pulse/pulse";
import {
  ACTION_TYPE_CHART_UPDATE_PULLREQUESTS,
  ACTION_TYPE_GIT_REPOS
} from "../../redux/redux";
import RefreshRepos from './refreshRepos';
import { renderChartUpdatePullRequests } from '../pulse/pulse';
import { InformationCircleIcon, FilterIcon, XIcon } from '@heroicons/react/solid'
import RefreshButton from '../../components/refreshButton/refreshButton';

export default class Repositories extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    let favoriteRepos = [];
    if (reduxState.user) {
      favoriteRepos = reduxState.user.favoriteRepos;
    }

    this.state = {
      repositories: this.mapToRepositories(reduxState.connectedAgents, reduxState.gitRepos),
      favorites: favoriteRepos,
      search: reduxState.search,
      agents: reduxState.settings.agents,
      application: reduxState.application,
      chartUpdatePullRequests: reduxState.pullRequests.chartUpdates,
      repositoriesLoading: true,
      repositoriesRefreshing: false,
      isOpen: false,
      added: null,
      deleted: null,
      settings: reduxState.settings,
      filters: [],
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      let favoriteRepos = [];
      if (reduxState.user) {
        favoriteRepos = reduxState.user.favoriteRepos;
      }

      this.setState({ repositories: this.mapToRepositories(reduxState.connectedAgents, reduxState.gitRepos) });
      this.setState({ search: reduxState.search });
      this.setState({ agents: reduxState.settings.agents });
      this.setState({ favorites: favoriteRepos });
      this.setState({
        application: reduxState.application,
        settings: reduxState.settings,
      });
      this.setState({ chartUpdatePullRequests: reduxState.pullRequests.chartUpdates });
    });

    this.navigateToRepo = this.navigateToRepo.bind(this);
    this.favoriteHandler = this.favoriteHandler.bind(this);
    this.deleteFilter = this.deleteFilter.bind(this);
    this.addFilter = this.addFilter.bind(this);
    this.filterValueByProperty = this.filterValueByProperty.bind(this);
    this.resetFilters = this.resetFilters.bind(this);
  }

  componentDidMount() {
    if (JSON.parse(localStorage.getItem("filters"))) {
      const storedFilters = JSON.parse(localStorage.getItem("filters"));
      this.setState({ filters: storedFilters });
    }

    this.props.gimletClient.getGitRepos()
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_GIT_REPOS, payload: data
        });
        this.setState({ repositoriesLoading: false });
      }, () => {
        this.setState({ repositoriesLoading: false });
      });
  }

  mapToRepositories(connectedAgents, gitRepos) {
    const repositories = []

    for (const r of gitRepos) {
      const repo = {
        name: r,
        services: [],
      }
      repositories.push(repo)
    }

    if (!connectedAgents) {
      return repositories;
    }

    for (const envName of Object.keys(connectedAgents)) {
      const env = connectedAgents[envName];

      for (const service of env.stacks) {
        repositories.forEach(repo => {
          if (repo.name === service.repo) {
            if (repo.services === undefined) {
              repo.services = [];
            }
  
            repo.services.push(service)
          }
        })
      }
    }

    return repositories;
  }

  favoriteHandler(repo) {
    let favorites = this.state.favorites;
    if (!favorites.includes(repo)) {
      favorites.push(repo);
    } else {
      favorites = favorites.filter(fav => fav !== repo);
    }

    this.props.gimletClient.saveFavoriteRepos(favorites);

    this.setState(prevState => {
      return {
        favorites: favorites
      }
    });
  }

  navigateToRepo(repo) {
    this.props.history.push(`/repo/${repo}`)
  }

  refresh() {
    this.setState({ repositoriesRefreshing: true });
    this.props.gimletClient.refreshRepos()
      .then(data => {
        data.added ? this.setState({ added: data.added }) : this.setState({ added: [] });
        data.deleted ? this.setState({ deleted: data.deleted }) : this.setState({ deleted: [] });
        this.props.store.dispatch({
          type: ACTION_TYPE_GIT_REPOS, payload: data.userRepos
        });
        this.setState({ repositoriesRefreshing: false });
      }, () => {
        this.setState({ repositoriesRefreshing: false });
        /* Generic error handler deals with it */
      });
  }

  chartUpdatePullRequests() {
    this.props.gimletClient.getChartUpdatePullRequests()
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_CHART_UPDATE_PULLREQUESTS,
          payload: data
        })
      }, () => {
        /* Generic error handler deals with it */
      });
  }

  deleteFilter(filter) {
    this.setState(prevState => {
      const deleted = []
      for(const f of prevState.filters){
        if (f.property !== filter.property || f.value !== filter.value){
          deleted.push(f)
        }
      }

      localStorage.setItem("filters", JSON.stringify(deleted))

      return {
        filters: deleted
      }
    });
  }

  addFilter(filter) {
    this.setState(prevState => {
      localStorage.setItem("filters", JSON.stringify([...prevState.filters, filter]));

      return {
        filters: [...prevState.filters, filter]
      }
    });
  }

  filterValueByProperty(property) {
    const filter = this.state.filters.find(f => f.property === property)
    if (!filter) {
      return ""
    }

    return filter.value
  }

  resetFilters() {
    this.setState({filters: []});
    localStorage.removeItem("filters");
  }

  render() {
    const { repositories, favorites, isOpen, settings } = this.state;

    if (!settings.provider || settings.provider === "") {
      return (
        <div>
          <header>
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
              <div>
                <h1 className="text-3xl font-bold leading-tight text-gray-900">Repositories</h1>
                <div className="mt-8">
                  {setupGithubCard(this.props.history)} 
                </div>
              </div>
            </div>
          </header>
        </div>
      )
    }

    const filteredRepositories = filterRepos(repositories, favorites, this.state.filters)
    filteredRepositories.sort((a,b) => a.name - b.name);
    const repoCards = filteredRepositories.map(repo => {
      return (
        <li key={repo.name} className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
          <RepoCard
            name={repo.name}
            services={repo.services}
            navigateToRepo={this.navigateToRepo}
            favorite={favorites.includes(repo.name)}
            favoriteHandler={this.favoriteHandler}
          />
        </li>
      )
    });

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
            <div className="flex justify-between">
              <h1 className="text-3xl font-bold leading-tight text-gray-900">Repositories</h1>
              <RefreshButton
                title="Refresh repositories"
                refreshFunc={() => {
                  this.setState({ isOpen: true });
                  this.refresh();
                  this.chartUpdatePullRequests();
                }}
              />
            </div>
            {isOpen &&
              <div className="mt-8">
                <RefreshRepos
                  added={this.state.added}
                  deleted={this.state.deleted}
                  repositoriesRefreshing={this.state.repositoriesRefreshing}
                  installationURL={this.state.application.installationURL}
                />
              </div>
            }

            <FilterBar
              filters={this.state.filters}
              addFilter={this.addFilter}
              filterValueByProperty={this.filterValueByProperty}
              deleteFilter={this.deleteFilter}
              resetFilters={this.resetFilters}
            />
            {renderChartUpdatePullRequests(this.state.chartUpdatePullRequests)}
          </div>
        </header>
        <main>
          {this.state.repositoriesLoading ?
            <div className='pt-8'><Spinner /></div>
            :
            <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
              <div className="px-4 pt-8 sm:px-0">
                <div>
                  <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">Repositories</h4>
                  <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {repoCards.length > 0 ? repoCards : emptyStateNoMatchingService()}
                  </ul>
                </div>
              </div>
            </div>}
        </main>
      </div>
    )
  }

}

const filterRepos = (repos, favorites, filters) => {
  let filteredRepositories = repos;
  filters.forEach(filter => {
    switch (filter.property) {
      case 'Repository':
        filteredRepositories = filteredRepositories.filter(repo => repo.name.includes(filter.value))
        break;
      case 'Service':
        filteredRepositories = filteredRepositories.filter(repo => {
          return repo.services.length !== 0 && repo.services.some(service => service.service.name.includes(filter.value));
        })
        break;
      case 'Namespace':
        filteredRepositories = filteredRepositories.filter(repo => {
          return repo.services.length !== 0 && repo.services.some(service => service.service.namespace.includes(filter.value));
        })
        break;
      case 'Owner':
        filteredRepositories = filteredRepositories.filter(repo => {
          return repo.services.length !== 0 && repo.services.some(service => {
            return service.osca && service.osca.owner.includes(filter.value)
          });
        })
        break;
      case 'Starred':
        filteredRepositories = filteredRepositories.filter(repo => favorites.includes(repo.name))
        break;
      case 'Domain':
        filteredRepositories = filteredRepositories.filter(repo => {
          return repo.services.length !== 0 && repo.services.some(service => service.ingresses !== undefined && service.ingresses.some(ingress => ingress.url.includes(filter.value)));
        })
        break;
      default:
    }
  })
  return filteredRepositories;
}

const setupGithubCard = (history) => {
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
            onClick={() => {history.push("/settings");return true}}
          >
            Click to integrate Github on the Settings page.
          </button>
        </div>
      </div>
    </div>
    </div>
  );
}

const FilterBar = (props) => {
  return (
    <div className="w-full pt-8">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 flex items-center pl-3">
          <FilterIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
          {props.filters.map(filter => (
            <Filter key={filter.property+filter.value} filter={filter} deleteFilter={props.deleteFilter} />
          ))}
          <FilterInput addFilter={props.addFilter} filterValueByProperty={props.filterValueByProperty} />
        </div>
        <div className="block w-full rounded-md border-0 bg-white py-1.5 pl-10 pr-3 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6">
          &nbsp;
        </div>
        <div className="absolute inset-y-0 right-0 flex items-center p-1">
          <button onClick={props.resetFilters} className="py-1 px-2 bg-gray-200 text-gray-400 rounded-full text-sm">reset</button>
        </div>
      </div>
    </div>
  )
}

const Filter = (props) => {
  const { filter } = props;
  return (
    <span className="ml-1 text-blue-50 bg-blue-600 rounded-full pl-3 pr-1" aria-hidden="true">
      <span>{filter.property}</span>{filter.property !== "Starred" && <span>: {filter.value}</span>}
      <span className="ml-1 px-1 bg-blue-400 rounded-full ">
        <XIcon className="cursor-pointer inline h-3 w-3" aria-hidden="true" onClick={() => props.deleteFilter(filter)}/>
      </span>
    </span>
  )
}

const FilterInput = (props) => {
  const [active, setActive] = useState(false)
  const [property, setProperty] = useState("")
  const [value, setValue] = useState("")
  const properties=["Repository", "Service", "Namespace", "Owner", "Starred", "Domain"]
  const { addFilter, filterValueByProperty } = props;
	const inputRef = useRef(null);

  const reset = () => {
    setActive(false)
    setProperty("")
    setValue("")
  }

  useEffect(() => {
    if (property !== "") {
      inputRef.current.focus();
    }  
  });

  return (
    <span className="relative w-48 ml-2">
      <span className="items-center flex">
        {property !== "" &&
          <span>{property}: </span>
        }
        <input
          ref={inputRef}
          key={property}
          className={`${property ? "ml-10" : "" }block border-0 border-t border-b border-gray-300 pt-1.5 pb-1 px-1 text-gray-900 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6`}
          placeholder='Enter Filter'
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onFocus={() => {setActive(true)}}
          onBlur={() => {
            setTimeout(() => {
              setActive(false);
              if (value !== "") {
                if (property === "") {
                  addFilter({property: "Repository", value: value})
                } else {
                  addFilter({property, value})
                }
                reset()
              } else {
                if (property !== "") {
                  reset()
                }
              }
            }, 200);}
          }
          onKeyUp={(e) => {
            if (e.keyCode === 13){
              setActive(false)
              if (property === "") {
                addFilter({property: "Repository", value: value})
              } else {
                addFilter({property, value})
              }
              reset()
            }
            if (e.keyCode === 27){
              reset()
              // inputRef.current.blur();
            }
          }}
          type="search"
        />
      </span>
      {active && property === "" &&
      <div className="z-10 absolute bg-blue-100 w-48 p-2 text-blue-800">
        <ul className="">
          {properties.map(p => {
            if (filterValueByProperty(p) !== "") {
              return null;
            }

            return (
              <li
                key={p}
                className="cursor-pointer hover:bg-blue-200"
                onClick={() => {
                  if (p === "Starred") {
                    addFilter({property: p, value: "true"})
                    return
                  }

                  setProperty(p);
                  setActive(false);
                  }}>
                {p}
              </li>
          )})}
        </ul>
      </div>
      }
    </span>
  )
}

export const Spinner = () => {
  return (
    <div className="max-w-7xl grid place-items-center mx-auto px-4 sm:px-6 lg:px-8">
      <div role="status">
        <svg className="inline w-16 h-16 text-gray-200 animate-spin dark:text-gray-200 fill-blue-600" viewBox="0 0 100 101" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z" fill="currentColor" />
          <path d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z" fill="currentFill" />
        </svg>
        <span className="sr-only">Loading...</span>
      </div>
    </div>
  )
}
