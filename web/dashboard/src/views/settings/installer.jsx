import { useState } from 'react'

const Installer = () => {
    const [org, setOrg] = useState(null);
    const [gitlabUrl, setGitlabUrl] = useState("https://gitlab.com");

    let url = window.location.href;
    url = url[url.length - 1] === '/' ? url.slice(0, -1) : url; // strip trailing slash

    const manifest = JSON.stringify({
        "name": "Gimlet Dashboard",
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
        <div className="mx-2">
            <div className="font-mono text-sm">
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
                <div className="mb-4">
                    <h2 className="text-xl font-sans font-bold leading-7 text-gray-900 sm:text-2xl sm:truncate mb-4">Gitlab.com</h2>
                    <p className="">Gimlet Dashboard uses personal or group access tokens to integrate with Gitlab.
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
    );
};

export default Installer;
