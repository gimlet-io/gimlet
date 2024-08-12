import { useState } from 'react'
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

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
          <a href="https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps" target="_blank" rel="noopener noreferrer" className="externalLink ml-1">
            Github Apps
            <ArrowTopRightOnSquareIcon className="externalLinkIcon mr-1" aria-hidden="true" />
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
                className="ml-2 p-2 border border-neutral-300 rounded-md leading-5 bg-white placeholder-neutral-500 focus:outline-none focus:placeholder-neutral-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
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
