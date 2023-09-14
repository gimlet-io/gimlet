import { useState } from 'react'

const Installer = () => {
  const [org, setOrg] = useState(null);

  let url = window.location.href;
  url = url[url.length - 1] === '/' ? url.slice(0, -1) : url; // strip trailing slash

  const manifest = JSON.stringify({
    "name": "Gimlet",
    "url": "https://gimlet.io",
    "redirect_url": url + '/created',
    "callback_url": url + '/installed',
    "request_oauth_on_install": true,
    "hook_attributes": {
      "url": "https://nosuchapp.gimlet.io/nosuchthing/hook",
    },
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
    <div className="mx-2">
      <div>
        <p>
          Gimlet uses
          <a href="https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps" target="_blank" rel="noopener noreferrer" className="ml-1">
            Github Apps
            <svg xmlns="http://www.w3.org/2000/svg"
              className="inline fill-current text-gray-500 hover:text-gray-700 mr-1" width="12" height="12"
              viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none" />
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
            </svg>
          </a>
          to integrate with Github.
        </p>

        <form action={org ? `https://github.com/organizations/${org}/settings/apps/new` : "https://github.com/settings/apps/new"} method="post">
          <input type="hidden" name="manifest" id="manifest" value={manifest}></input><br />

          <ul className="pl-8 list-disc">
            <li>When you integrate with Github you create a Github Apps application.</li>
            <li>This application will be owned by your personal Github account, thus you don't give access to any third party or the makers of Gimlet.</li>
            <li>(optional) In case you want to integrate your company's Github repositories, provide your company's Github Organization:
              <input id="org" name="org"
                value={org}
                onChange={e => setOrg(e.target.value)}
                className="ml-2 p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                type="text" />
            </li>
            <li>Github Apps allow fine-grained integrations, so you will be able to pick the repositories that you want to integrate with Gimlet.</li>
          </ul>

          <input type="submit" value="Create Github app & Integrate Gimlet"
            className="mt-8 cursor-pointer font-sans inline-flex items-center px-4 py-2 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
          </input>
        </form>
      </div>
    </div>
  );
};

export default Installer;
