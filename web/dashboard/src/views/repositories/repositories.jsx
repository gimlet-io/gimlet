import React, {Component} from 'react';
import RepoCard from "../../components/repoCard/repoCard";
import {emptyStateNoAgents, emptyStateNoMatchingService} from "../services/services";

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
      repositories: this.mapToRepositories(reduxState.envs, reduxState.gitRepos),
      favorites: favoriteRepos,
      search: reduxState.search,
      agents: reduxState.settings.agents
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      let favoriteRepos = [];
      if (reduxState.user) {
        favoriteRepos = reduxState.user.favoriteRepos;
      }

      this.setState({repositories: this.mapToRepositories(reduxState.envs, reduxState.gitRepos)});
      this.setState({search: reduxState.search});
      this.setState({agents: reduxState.settings.agents});
      this.setState({favorites: favoriteRepos});
    });

    this.navigateToRepo = this.navigateToRepo.bind(this);
    this.favoriteHandler = this.favoriteHandler.bind(this);
  }

  mapToRepositories(envs, gitRepos) {
    const repositories = {}

    for (const r of gitRepos) {
      if (repositories[r] === undefined) {
        repositories[r] = [];
      }
    }

    for (const envName of Object.keys(envs)) {
      const env = envs[envName];

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

  render() {
    const {repositories, search, agents, favorites} = this.state;

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

    const emptyState = search.filter !== '' ?
      emptyStateNoMatchingService()
      :
      (<p className="text-xs text-gray-800">No services</p>);

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Repositories</h1>
          </div>
        </header>
        <main>
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
              {agents.length === 0 && emptyStateNoAgents()}
              
              <div>
                {favorites.length > 0 &&
                <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">Repositories</h4>
                }
                <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                  {repoCards.length > 0 ? repoCards : emptyState}
                </ul>
              </div>
              
            </div>
          </div>
        </main>
      </div>
    )
  }

}
