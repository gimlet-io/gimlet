import { useState } from 'react';

const StepOne = ({ getContext }) => {
  const [org, setOrg] = useState(null);
  const [gitlabUrl, setGitlabUrl] = useState("https://gitlab.com");

  let url = window.location.href;
  url = url[url.length - 1] === '/' ? url.slice(0, -1) : url; // strip trailing slash

  const manifest = JSON.stringify({
    "name": "Gimlet",
    "url": url,
    "callback_url": url + '/auth',
    "hook_attributes": {
      "url": url + '/hook'
    },
    "redirect_url": url + '/created',
    "setup_url": url + '/installed',
    "public": false,
    "default_permissions": {
      "administration": "write",
      "checks": "read",
      "contents": "write",
      "pull_requests": "write",
      "repository_hooks": "write",
      "statuses": "read",
      "members": "read"
    },
    "default_events": [
      "create",
      "push",
      "delete",
      "status",
      "check_run"
    ]
  })

  return (
    <div className="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate mb-12">Gimlet Installer</h2>
          </div>
        </div>
        <nav aria-label="Progress">
          <ol className="border border-gray-300 rounded-md divide-y divide-gray-300 md:flex md:divide-y-0">
            <li className="relative md:flex-1 md:flex">
              {/* <!-- Current Step --> */}
              <div className="px-6 py-4 flex items-center text-sm font-medium select-none cursor-default" aria-current="step">
                <span
                  className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-indigo-600 rounded-full">
                  <span className="text-indigo-600">01</span>
                </span>
                <span className="ml-4 text-sm font-medium text-indigo-600">Integrate with Source Code Manager</span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                    strokeLinejoin="round" />
                </svg>
              </div>
            </li>

            <li className="relative md:flex-1 md:flex">
              {/* <!-- Upcoming Step --> */}
              <div className="group flex items-center select-none cursor-default">
                <span className="px-6 py-4 flex items-center text-sm font-medium">
                  <span
                    className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-gray-300 rounded-full">
                    <span className="text-gray-500">02</span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500">Prepare gitops repository</span>
                </span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                    strokeLinejoin="round" />
                </svg>
              </div>
            </li>

            <li className="relative md:flex-1 md:flex">
              {/* <!-- Upcoming Step --> */}
              <div className="group flex items-center select-none cursor-default">
                <span className="px-6 py-4 flex items-center text-sm font-medium">
                  <span
                    className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-gray-300 rounded-full">
                    <span className="text-gray-500">03</span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500">Bootstrap gitops automation</span>
                </span>
              </div>
            </li>
          </ol>
        </nav>

        <div className="mt-8 font-mono text-sm">
          <div>
          <h2 className="text-xl font-sans font-bold leading-7 text-gray-900 sm:text-2xl sm:truncate mb-4">Github</h2>
          <p className="">Gimlet Dashboard uses a Github Application to gain access to your source code.</p>
          <p className="">You must create this application first.</p>

          <p className="pt-8">Please note that this application will be owned by you, thus you don't give access to any third party or the makers of Gimlet.</p>

          <form action={org ? `https://github.com/organizations/${org}/settings/apps/new` : "https://github.com/settings/apps/new"} method="post">
            <input type="hidden" name="manifest" id="manifest" value={manifest}></input><br />
            <div className="text-gray-700">
              <div className="flex mt-4">
                <div className="font-medium self-center">Github Organization</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="org" name="org"
                      value={org}
                      onChange={e => setOrg(e.target.value)}
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
              <div className="text-xs text-gray-500">Leave it empty to have your personal Github account own the Github Application. Leaving empty is best for Gimlet evaluation.</div>
            </div>
            <input type="submit" value="Create Github app"
              className="mt-8 cursor-pointer font-sans inline-flex items-center px-4 py-2 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            </input>
          </form>
          </div>
          <div className="my-8">
            <div className="inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-gray-300" />
            </div>
          </div>
          <div className="mb-64">
            <h2 className="text-xl font-sans font-bold leading-7 text-gray-900 sm:text-2xl sm:truncate mb-4">Gitlab.com</h2>
            <p className="">Gimlet uses personal or group access tokens to integrate with Gitlab. 
            Gimlet will also need an OAuth application to handle OAuth based authentication.</p>
            <form action="/gitlabInit" method="post">
            <div className="text-gray-700">
            <div className="flex mt-4">
                <div className="font-medium self-center">Gitlab URL</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="gitlabUrl" name="gitlabUrl"
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      value={gitlabUrl}
                      onChange={e => setGitlabUrl(e.target.value)}
                      type="text" />
                  </div>
                </div>
              </div>
              <div className="flex mt-4">
                <div className="font-medium self-center">Personal/Group Access Token</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="token" name="token"
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
              <div className="text-xs text-gray-500">
                Personal access  tokens can be created on the <a className="text-blue-500" href="https://gitlab.com/-/profile/personal_access_tokens" target="_blank" rel="noreferrer">https://gitlab.com/-/profile/personal_access_tokens</a> page,
                group access tokens on <a className="text-blue-500" href="https://gitlab.com/groups/$your-group/-/settings/access_tokens" target="_blank" rel="noreferrer">https://gitlab.com/groups/$your-group/-/settings/access_tokens</a>.
                Grant `api` and `write_repository` access. For Group tokens, grant an `Owner` role too.</div>
              <div className="flex mt-4">
                <div className="font-medium self-center">Gitlab Application ID</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="appId" name="appId"
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
              <div className="flex mt-4">
                <div className="font-medium self-center">Gitlab Application Secret</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="appSecret" name="appSecret"
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
              <div className="text-xs text-gray-500">Applications can be created on the <a className="text-blue-500" href="https://gitlab.com/-/profile/applications" target="_blank" rel="noreferrer">https://gitlab.com/-/profile/applications</a> page,
              group applications on <a className="text-blue-500" href="https://gitlab.com/groups/$your-group/-/settings/applications" target="_blank" rel="noreferrer">https://gitlab.com/groups/$your-group/-/settings/applications</a>.
                Grant `api` access. The redirect URL is: {window.location.href}auth</div>
            </div>
            <input type="submit" value="Integrate Gimlet"
              className="mt-8 cursor-pointer font-sans inline-flex items-center px-4 py-2 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            </input>
          </form>
          </div>
        </div>

      </div>
    </div>
  );
};

export default StepOne;
