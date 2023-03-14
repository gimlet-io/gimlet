import React, { Component } from 'react';
import RepoCard from "../../components/repoCard/repoCard";
import { emptyStateNoMatchingService } from "../pulse/pulse";
import { ACTION_TYPE_GIT_REPOS } from "../../redux/redux";
import RefreshRepos from './refreshRepos';

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
      repositoriesLoading: true,
      repositoriesRefreshing: false,
      isOpen: false,
      added: null,
      deleted: null,
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
      this.setState({ application: reduxState.application });
    });

    this.navigateToRepo = this.navigateToRepo.bind(this);
    this.favoriteHandler = this.favoriteHandler.bind(this);
  }

  componentDidMount() {
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
    const repositories = {}

    for (const r of gitRepos) {
      if (repositories[r] === undefined) {
        repositories[r] = [];
      }
    }

    if (!connectedAgents) {
      return repositories;
    }

    for (const envName of Object.keys(connectedAgents)) {
      const env = connectedAgents[envName];

      for (const service of env.stacks) {
        if (repositories[service.repo] === undefined) {
          repositories[service.repo] = [];
        }

        repositories[service.repo].push(service);
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

  render() {
    const { repositories, search, favorites, isOpen } = this.state;

    let filteredRepositories = {};
    for (const repoName of Object.keys(repositories)) {
      filteredRepositories[repoName] = repositories[repoName];
      if (search.filter !== '') {
        filteredRepositories[repoName] = filteredRepositories[repoName].filter((service) => {
          return service.service.name.includes(search.filter) ||
            (service.deployment !== undefined && service.deployment.name.includes(search.filter)) ||
            (service.ingresses !== undefined && service.ingresses.filter((ingress) => ingress.url.includes(search.filter)).length > 0)
        })
        if (filteredRepositories[repoName].length === 0 && !repoName.includes(search.filter)) {
          delete filteredRepositories[repoName];
        }
      }
    }

    const filteredRepoNames = Object.keys(filteredRepositories);
    filteredRepoNames.sort();
    const repoCards = filteredRepoNames.map(repoName => {
      return (
        <li key={repoName} className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
          <RepoCard
            name={repoName}
            services={filteredRepositories[repoName]}
            navigateToRepo={this.navigateToRepo}
            favorite={favorites.includes(repoName)}
            favoriteHandler={this.favoriteHandler}
          />
        </li>
      )
    });

    const filteredFavorites = filteredRepoNames.filter(repo => favorites.includes(repo))
    const favoriteRepoCards = filteredFavorites.map(repoName => {
      return (
        <li key={repoName} className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
          <RepoCard
            name={repoName}
            services={filteredRepositories[repoName]}
            navigateToRepo={this.navigateToRepo}
            favorite={favorites.includes(repoName)}
            favoriteHandler={this.favoriteHandler}
          />
        </li>
      )
    });

    const emptyState = search.filter !== '' ? emptyStateNoMatchingService() : null;

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Repositories</h1>
            <div className="space-y-2">
              <button className="flex text-xs text-gray-700 hover:text-blue-700 mt-2"
                onClick={() => {
                  this.setState({ isOpen: true });
                  this.refresh();
                }}
              >
                Refresh repositories
              </button>
              {isOpen &&
                <RefreshRepos
                  added={this.state.added}
                  deleted={this.state.deleted}
                  repositoriesRefreshing={this.state.repositoriesRefreshing}
                  installationURL={this.state.application.installationURL}
                />}
            </div>
          </div>
        </header>
        <main>
          {this.state.repositoriesLoading ?
            <Spinner />
            :
            <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
              {favorites.length > 0 &&
                <div className="px-4 pt-8 sm:px-0">
                  <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">Favorites</h4>
                  <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {favoriteRepoCards}
                  </ul>
                </div>
              }
              <div className="px-4 pt-8 sm:px-0">
                <div>
                  {favorites.length > 0 &&
                    <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">Repositories</h4>
                  }
                  <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {repoCards.length > 0 ? repoCards : emptyState}
                  </ul>
                </div>

              </div>
            </div>}
        </main>
      </div>
    )
  }

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
