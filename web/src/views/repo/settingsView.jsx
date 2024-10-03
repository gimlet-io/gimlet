import { useState, useEffect, useRef } from 'react';
import Toggle from '../../components/toggle/toggle';
import { useParams, useLocation } from 'react-router-dom'
import { toast } from 'react-toastify';
import { InProgress, Success, Error } from '../../popUpWindow';

export function RepoSettingsView(props) {
  const { gimletClient } = props;
  const [pullRequestPolicyLoaded, setPullRequestPolicyLoaded] = useState()
  const [pullRequestPolicy, setPullRequestPolicy] = useState()
  const [defaultpullRequestPolicy, setDefaultPullRequestPolicy] = useState()

  const { owner, repo } = useParams();
  const location = useLocation()
  const progressToastId = useRef(null);

  useEffect(() => {
    gimletClient.repoPullRequestsPolicy(owner, repo)
      .then(data => {
        setPullRequestPolicyLoaded(true)
        setPullRequestPolicy(data.pullRequestPolicy)
        setDefaultPullRequestPolicy(data.pullRequestPolicy)
      }, () => {
        setPullRequestPolicyLoaded(true)
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const saveRepoPullRequestPolicy = () => {
    progressToastId.current = toast(<InProgress header="Saving..." />, { autoClose: false });

    gimletClient.saveRepoPullRequestPolicy(owner, repo, pullRequestPolicy)
      .then(() => {
        toast.update(progressToastId.current, {
          render: <Success header="Saved" />,
          className: "bg-green-50 shadow-lg p-2",
          bodyClassName: "p-2",
          autoClose: 3000,
        });
        setDefaultPullRequestPolicy(pullRequestPolicy)
      }, (err) => {
        toast.update(progressToastId.current, {
          render: <Error header="Error" message={err.statusText} />,
          className: "bg-red-50 shadow-lg p-2",
          bodyClassName: "p-2",
          progressClassName: "!bg-red-200",
          autoClose: 5000
        });
      });
  }

  if (!pullRequestPolicyLoaded) {
    return <SkeletonLoader />
  }

  const hasChange = pullRequestPolicy !== defaultpullRequestPolicy
  const navigation = [
    { name: "General", href: "/general" },
  ]
  let selectedNavigation = navigation.find(i => location.pathname.endsWith(i.href))
  if (!selectedNavigation) {
    selectedNavigation = navigation[0]
  }

  return (
    <>
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 pb-8 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow py-0.5">Repository settings</h1>
          <button
            type="button"
            disabled={!hasChange}
            className={(hasChange ? 'primaryButton' : 'primaryButtonDisabled') + ` px-4`}
            onClick={saveRepoPullRequestPolicy}
          >
            Save
          </button>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-64">
        <div className="sticky top-0 h-96">
          <nav aria-label="Sidebar">
            <ul className="w-56">
              {navigation.map((item) => (
                <li key={item.name}>
                  <button
                    className='font-medium group flex w-full gap-x-3 p-2 pl-3 text-sm leading-6 rounded-md hover:bg-neutral-200 dark:hover:bg-neutral-600'
                  >
                    {item.name}
                  </button>
                </li>
              ))}
            </ul>
          </nav>
        </div>
        <div className="w-full ml-14">
          {(!selectedNavigation || selectedNavigation?.name === "General") &&
            <General
              pullRequestPolicy={pullRequestPolicy}
              setPullRequestPolicy={setPullRequestPolicy}
            />
          }
        </div>
      </div>
    </>
  )
}

function General(props) {
  const { pullRequestPolicy, setPullRequestPolicy } = props;

  return (
    <div className="w-full card">
      <div className="p-6 pb-4 items-center">
        <label htmlFor="environment" className="block font-medium">Open Pull Request For Configuration Changes</label>
        <div className="my-4">
          <p className="max-w-4xl text-sm text-neutral-800 dark:text-neutral-400">
            Enabling this option for configuration changes to be made as a Pull Request.
          </p>
          <div className="my-4">
            <Toggle
              checked={pullRequestPolicy}
              onChange={setPullRequestPolicy}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

function SkeletonLoader(props) {
  return (
    <>
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 pb-8 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow py-0.5">Repository settings</h1>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-64 animate-pulse">
        <div className="sticky h-96 top-56">
          <div className="w-56 p-4 pl-3 space-y-6">
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
          </div>
        </div>
        <div className="w-full ml-14">
          <div role="status" className="flex items-center justify-center h-56 bg-neutral-300 dark:bg-neutral-500 rounded-lg">
            <span className="sr-only">Loading...</span>
          </div>
        </div>
      </div>
    </>
  )
}
