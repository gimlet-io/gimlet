import { useState } from 'react'
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

export default function GithubIntegration(props) {
    return (
      <div className='text-neutral-900 dark:text-neutral-200'>
      <div className="w-full bg-white dark:bg-neutral-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow">Github Integration</h1>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center">
          <div className='font-light text-sm pt-2 pb-12'>
            To deploy an application, integrate with Github first.
          </div>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 mt-16 flex">
        <Installer />
      </div>
    </div>
    )
}

const Installer = () => {
  const [org, setOrg] = useState(null);

  let url = window.location.href;
  url = url[url.length - 1] === '/' ? url.slice(0, -1) : url; // strip trailing slash

  const manifest = JSON.stringify({
    "name": "Gimlet",
    "url": "https://gimlet.io",
    "redirect_url": url + '/created',
    "callback_url": url + '/auth',
    "setup_url": url + '/installed',
    "request_oauth_on_install": false,
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
    <div className="p-6 w-full card">
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
            <li>First, you need to create this Github Application.</li>
            <li>You will own this application so you are not giving access to any third party or the makers of Gimlet.</li>
            <li>If you want to deploy your company's repositories, provide your company's Github Organization name in this box:
              <input id="org" name="org"
                value={org}
                onChange={e => setOrg(e.target.value)}
                className="block w-full filter input"
                type="text" />
            </li>
          </ul>

          <input type="submit" value="Create Github app & Integrate Gimlet"
            className="mt-8 primaryButton w-full py-4">
          </input>
        </form>
      </div>
    </div>
  );
};
