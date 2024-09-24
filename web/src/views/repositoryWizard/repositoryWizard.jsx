import { useState, useEffect } from 'react'
import { ArrowRightIcon } from '@heroicons/react/20/solid'
import { ACTION_TYPE_GIT_REPOS } from '../../redux/redux';
import { Loading } from '../repo/deployStatus';
import { useHistory } from 'react-router-dom'

export default function RepositoryWizard(props) {
  const history = useHistory()
  return (
    <div className='text-neutral-900 dark:text-neutral-200'>
      <div className="w-full bg-white dark:bg-neutral-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow">Import Git Repository</h1>
          <button
            type="button"
            className='secondaryButton'
            onClick={() => history.push("/repositories")}
          >
            I am done importing
          </button>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center">
          <div className='font-light text-sm pt-2 pb-12'>
            To deploy an application, import its Git Repository first.
          </div>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 flex">
        <ImportWizard {...props} />
      </div>
    </div>
  )
}

export function ImportWizard(props) {
  const { gimletClient, store } = props;
  let reduxState = store.getState();

  const [application, setApplication] = useState(reduxState.application)
  const [user, setUser] = useState(reduxState.user)
  const [importedRepos, setImportedRepos] = useState(reduxState.gitRepos)
  const [repos, setRepos] = useState()
  const [searchTerm, setSearchTerm] = useState("")

  store.subscribe(() => {
    let reduxState = store.getState();
    setApplication(reduxState.application)
    setUser(reduxState.user)
    setImportedRepos(reduxState.gitRepos)
  });

  useEffect(() => {
    if (importedRepos.length === 0) {
      gimletClient.getGitRepos()
        .then(repos => {
          store.dispatch({
            type: ACTION_TYPE_GIT_REPOS, payload: repos
          })
        }, () => {/* Generic error handler deals with it */ });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setRepos()
    //Instead of sending that request on every keypress, we can wait a little bit until the user stops typing, and then send the entire value in one go.
    const delayDebounceFn = setTimeout(() => {
      gimletClient.searchRepo(searchTerm, controller.signal)
        .then(data => {
          setRepos(data)
        }, () => {/* Generic error handler deals with it */ });
    }, 500)

    return () => {
      controller.abort()
      clearTimeout(delayDebounceFn)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchTerm])

  return (
    <div className="p-6 w-full h-[556px] card">
      <div>
        <label htmlFor="filter" className="sr-only">
          Search
        </label>
        <div className="w-full">
          <div className="relative">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3">
              <svg data-testid="geist-icon" className="filterIcon" strokeLinejoin="round" viewBox="0 0 16 16">
                <path fillRule="evenodd" clipRule="evenodd" d="M1.5 6.5C1.5 3.73858 3.73858 1.5 6.5 1.5C9.26142 1.5 11.5 3.73858 11.5 6.5C11.5 9.26142 9.26142 11.5 6.5 11.5C3.73858 11.5 1.5 9.26142 1.5 6.5ZM6.5 0C2.91015 0 0 2.91015 0 6.5C0 10.0899 2.91015 13 6.5 13C8.02469 13 9.42677 12.475 10.5353 11.596L13.9697 15.0303L14.5 15.5607L15.5607 14.5L15.0303 13.9697L11.596 10.5353C12.475 9.42677 13 8.02469 13 6.5C13 2.91015 10.0899 0 6.5 0Z" fill="currentColor"></path>
              </svg>
            </div>
            <input
              onChange={e => setSearchTerm(e.target.value)}
              type="text"
              name="filter"
              id="filter"
              className="filter"
              placeholder="Search..."
            />
          </div>
        </div>
      </div>
      {repos ?
        <>
          <div className='relative w-full bg-white dark:bg-neutral-800 rounded-md ring-1 ring-inset ring-neutral-200 dark:ring-neutral-700 mt-10'>
            {repos.length === 0 && <NoRepos searchTerm={searchTerm} user={user.login} installationUrl={application.installationURL} />}
            <div className='divide-y dark:divide-neutral-700'>
              {
                repos.slice(0, 5).map((repo, idx) => {
                  const imported = importedRepos.includes(repo);
                  return (
                    <li key={idx} className="flex items-center justify-between space-x-3 p-4">
                      <div className="flex min-w-0 flex-1 items-center space-x-3">
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium text-neutral-900 dark:text-neutral-300">{repo}</p>
                        </div>
                      </div>
                      <ImportButton {...props} repo={repo} imported={imported} />
                    </li>)
                })
              }
            </div>
          </div>
          <div className="flex items-center justify-between mt-4">
            {repos.length !== 0 &&
              <div className='text-sm text-neutral-600 dark:text-neutral-400 py-4'>
                Missing a Git repository? <a href={application.installationURL} rel="noreferrer" target="_blank" className='text-blue-500'>Adjust Github App Permission<ArrowRightIcon className="size-4 inline" aria-hidden="true" /></a>
              </div>
            }
          </div>
        </>
        :
        <div className="flex w-full h-full items-center justify-center">
          <label htmlFor="label-title" className="text-neutral-700 dark:text-neutral-300 text-sm">Searching...</label>
        </div>
      }
    </div>
  )
}

function ImportButton(props) {
  const { repo, imported } = props;
  const { gimletClient, store } = props;

  const [importing, setImporting] = useState(false)

  const importRepo = (name) => {
    setImporting(true)
    gimletClient.importRepo(name)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_GIT_REPOS, payload: data
        });
        setImporting(false)
      }, () => {
        setImporting(false)
        /* Generic error handler deals with it */
      });
  }

  return (
    <div className="flex-shrink-0">
      {imported ?
        <button className="cursor-default bg-neutral-200 dark:bg-neutral-700 inline-flex items-center px-3 py-2 border border-transparent font-normal text-base font-sans rounded-md text-neutral-400" >
          Imported
        </button>
        :
        <button
          disabled={importing}
          onClick={() => importRepo(repo)}
          className={`${importing ? 'primaryButtonDisabled' : 'primaryButton'} px-3`}
        >
          {importing ? <><Loading />Importing</> : 'Import'}
        </button>
      }
    </div>
  )
}

function NoRepos(props) {
  const { searchTerm, user, installationUrl } = props;

  const gitUser = <span className="justify-center items-center font-medium text-neutral-700 dark:text-neutral-300"><GithubLogo className="text-neutral-900 dark:text-white size-3 mr-0.5 inline fill-current" />{user}</span>;

  return (
    <div className="w-full card p-16">
      <div className="mx-auto text-center py-12 font-light">
        <label htmlFor="label-title" className="text-neutral-900 dark:text-neutral-300 font-medium text-sm">No Results Found</label>
        <div className="text-sm text-neutral-600 dark:text-neutral-400 mt-2">Your search for "{searchTerm}" in {gitUser} did not return any results.</div>
        <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-2">Make sure to grant Gimlet access to the Git repositories you’d like to import.</p>
        <a
          href={installationUrl}
          rel="noreferrer"
          target="_blank"
          className="primaryButton mt-6 px-40 py-2"
        >
          <GithubLogo className="size-7 mr-2" />
          Configure Github App
        </a>
      </div>
    </div>
  )
}

function GithubLogo(props) {
  const { className } = props;
  return (
    <svg data-testid="geist-icon" className={className} strokeLinejoin="round" viewBox="0 0 16 16"><g clipPath="url(#clip0_872_3147)">
      <path fillRule="evenodd" clipRule="evenodd" d="M8 0C3.58 0 0 3.57879 0 7.99729C0 11.5361 2.29 14.5251 5.47 15.5847C5.87 15.6547 6.02 15.4148 6.02 15.2049C6.02 15.0149 6.01 14.3851 6.01 13.7154C4 14.0852 3.48 13.2255 3.32 12.7757C3.23 12.5458 2.84 11.836 2.5 11.6461C2.22 11.4961 1.82 11.1262 2.49 11.1162C3.12 11.1062 3.57 11.696 3.72 11.936C4.44 13.1455 5.59 12.8057 6.05 12.5957C6.12 12.0759 6.33 11.726 6.56 11.5261C4.78 11.3262 2.92 10.6364 2.92 7.57743C2.92 6.70773 3.23 5.98797 3.74 5.42816C3.66 5.22823 3.38 4.40851 3.82 3.30888C3.82 3.30888 4.49 3.09895 6.02 4.1286C6.66 3.94866 7.34 3.85869 8.02 3.85869C8.7 3.85869 9.38 3.94866 10.02 4.1286C11.55 3.08895 12.22 3.30888 12.22 3.30888C12.66 4.40851 12.38 5.22823 12.3 5.42816C12.81 5.98797 13.12 6.69773 13.12 7.57743C13.12 10.6464 11.25 11.3262 9.47 11.5261C9.76 11.776 10.01 12.2558 10.01 13.0056C10.01 14.0752 10 14.9349 10 15.2049C10 15.4148 10.15 15.6647 10.55 15.5847C12.1381 15.0488 13.5182 14.0284 14.4958 12.6673C15.4735 11.3062 15.9996 9.67293 16 7.99729C16 3.57879 12.42 0 8 0Z" fill="currentColor"></path>
    </g>
      <defs>
        <clipPath id="clip0_872_3147">
          <rect width="16" height="16" fill="white"></rect>
        </clipPath>
      </defs></svg>
  )
}
