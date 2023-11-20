import { useEffect, useState } from "react";
// eslint-disable-next-line import/no-webpack-loader-syntax
import gimletHeader from "!file-loader!./gimletHeader.svg";
import axios from "axios";

const LoginPage = () => {
  const [token, setToken] = useState("");
  const [provider, setProvider] = useState("");
  const [termsOfServiceFeatureFlag, setTermsOfServiceFeatureFlag] = useState(false);
  let redirect = localStorage.getItem('redirect');
  if (!redirect) {
    redirect = "/"
  }

  useEffect(() => {
    getFlags().then(data => {
      setProvider(data.provider);
      setTermsOfServiceFeatureFlag(data.termsOfServiceFeatureFlag);
    }).catch(err => {
      console.error(`Error: ${err}`);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getFlags]);

  return (
    <div className="fixed flex inset-0 bg-gray-100">
      <div className="max-w-7xl m-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          <div className="md:flex md:items-center md:justify-between">
            <div className="flex-1 min-w-0">
              <div className="sm:mx-auto sm:max-w-md py-8 px-4 bg-white shadow-md sm:px-10">
                <div className="my-8">
                  <img className="h-16 mx-auto" src={gimletHeader} alt="gimlet-logo" />
                  <div className="my-16 text-base font-medium text-gray-700">
                    {loginButton(provider, redirect)}
                    {provider === "" &&
                      <form action="/admin-key-auth" method="post">
                        <div className="space-y-8">
                          <div className="space-y-1">
                            <label htmlFor="token" className="text-gray-700 mr-4 block text-sm font-medium">
                              Admin key
                            </label>
                            <input
                              type="text"
                              name="token"
                              id="token"
                              value={token}
                              onChange={e => setToken(e.target.value)}
                              className="block w-full p-3 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                            />
                          </div>
                          <button
                            type="submit"
                            disabled={token === ""}
                            className={(token === "" && "cursor-not-allowed") + " inline-flex items-center justify-center w-full font-medium px-20 py-3 rounded border border-gray-300 bg-gray-700 hover:bg-gray-600 shadow-sm text-base text-gray-100"}
                          >
                            Sign in with admin key
                          </button>
                        </div>
                      </form>
                    }
                  </div>
                  {termsOfServiceFeatureFlag && <div className="text-center font-light text-gray-700 flex flex-wrap justify-center">
                    By logging in, you're agreeing to our
                    <a href="https://gimlet.io/tos" className="hover:underline" target="_blank" rel="noopener noreferrer">Terms of Service</a>.
                  </div>}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const getFlags = async () => {
  try {
    const resp = await axios.get('/flags');
    return await resp.data;
  } catch (err) {
    // Handle Error Here
    console.error(`Error: ${err}`);
  }
};

const loginButton = (provider, redirect) => {
  if (provider === "GitHub") {
    return (
      <button
        onClick={() => {
          window.location.replace(`/auth?appState=https://${window.location.hostname}/auth%26redirect=${redirect}`);
        }}
        className="inline-flex items-center justify-center w-full font-medium px-20 py-3 rounded border border-gray-300 hover:bg-gray-50 shadow-sm"
      >
        <svg width="32" height="32" className="mr-2" viewBox="0 -3 256 256" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid">
          <path d="M128.001 0C57.317 0 0 57.307 0 128.001c0 56.554 36.676 104.535 87.535 121.46 6.397 1.185 8.746-2.777 8.746-6.158 0-3.052-.12-13.135-.174-23.83-35.61 7.742-43.124-15.103-43.124-15.103-5.823-14.795-14.213-18.73-14.213-18.73-11.613-7.944.876-7.78.876-7.78 12.853.902 19.621 13.19 19.621 13.19 11.417 19.568 29.945 13.911 37.249 10.64 1.149-8.272 4.466-13.92 8.127-17.116-28.431-3.236-58.318-14.212-58.318-63.258 0-13.975 5-25.394 13.188-34.358-1.329-3.224-5.71-16.242 1.24-33.874 0 0 10.749-3.44 35.21 13.121 10.21-2.836 21.16-4.258 32.038-4.307 10.878.049 21.837 1.47 32.066 4.307 24.431-16.56 35.165-13.12 35.165-13.12 6.967 17.63 2.584 30.65 1.255 33.873 8.207 8.964 13.173 20.383 13.173 34.358 0 49.163-29.944 59.988-58.447 63.157 4.591 3.972 8.682 11.762 8.682 23.704 0 17.126-.148 30.91-.148 35.126 0 3.407 2.304 7.398 8.792 6.14C219.37 232.5 256 184.537 256 128.002 256 57.307 198.691 0 128.001 0Zm-80.06 182.34c-.282.636-1.283.827-2.194.39-.929-.417-1.45-1.284-1.15-1.922.276-.655 1.279-.838 2.205-.399.93.418 1.46 1.293 1.139 1.931Zm6.296 5.618c-.61.566-1.804.303-2.614-.591-.837-.892-.994-2.086-.375-2.66.63-.566 1.787-.301 2.626.591.838.903 1 2.088.363 2.66Zm4.32 7.188c-.785.545-2.067.034-2.86-1.104-.784-1.138-.784-2.503.017-3.05.795-.547 2.058-.055 2.861 1.075.782 1.157.782 2.522-.019 3.08Zm7.304 8.325c-.701.774-2.196.566-3.29-.49-1.119-1.032-1.43-2.496-.726-3.27.71-.776 2.213-.558 3.315.49 1.11 1.03 1.45 2.505.701 3.27Zm9.442 2.81c-.31 1.003-1.75 1.459-3.199 1.033-1.448-.439-2.395-1.613-2.103-2.626.301-1.01 1.747-1.484 3.207-1.028 1.446.436 2.396 1.602 2.095 2.622Zm10.744 1.193c.036 1.055-1.193 1.93-2.715 1.95-1.53.034-2.769-.82-2.786-1.86 0-1.065 1.202-1.932 2.733-1.958 1.522-.03 2.768.818 2.768 1.868Zm10.555-.405c.182 1.03-.875 2.088-2.387 2.37-1.485.271-2.861-.365-3.05-1.386-.184-1.056.893-2.114 2.376-2.387 1.514-.263 2.868.356 3.061 1.403Z" fill="#161614" />
        </svg>
        Sign in with GitHub
      </button>
    )
  } else if (provider === "GitLab") {
    return (
      <button
        onClick={() => {
          window.location.replace(`/auth?appState=https://${window.location.hostname}/auth%26redirect=${redirect}`);
        }}
        className="inline-flex items-center justify-center w-full font-medium px-20 py-3 rounded border border-gray-300 hover:bg-gray-50 shadow-sm"
      >
        <svg width="32" height="32" className="mr-2" viewBox="0 -10 256 256" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid">
          <path d="m128.075 236.075 47.104-144.97H80.97l47.104 144.97Z" fill="#E24329" />
          <path d="M128.075 236.074 80.97 91.104H14.956l113.119 144.97Z" fill="#FC6D26" />
          <path d="M14.956 91.104.642 135.16a9.752 9.752 0 0 0 3.542 10.903l123.891 90.012-113.12-144.97Z" fill="#FCA326" />
          <path d="M14.956 91.105H80.97L52.601 3.79c-1.46-4.493-7.816-4.492-9.275 0l-28.37 87.315Z" fill="#E24329" />
          <path d="m128.075 236.074 47.104-144.97h66.015l-113.12 144.97Z" fill="#FC6D26" />
          <path d="m241.194 91.104 14.314 44.056a9.752 9.752 0 0 1-3.543 10.903l-123.89 90.012 113.119-144.97Z" fill="#FCA326" />
          <path d="M241.194 91.105h-66.015l28.37-87.315c1.46-4.493 7.816-4.492 9.275 0l28.37 87.315Z" fill="#E24329" />
        </svg>
        Sign in with GitLab
      </button>
    )
  }
}

export default LoginPage;
