import { format } from "date-fns";

function EnvironmentCard({ env, navigateToEnv, isOnline, trial }) {
  const exactDate = format(env.expiry * 1000, 'h:mm:ss a, MMMM do yyyy')

  return (
    <li className="card h-24 cursor-pointer"
      onClick={() => navigateToEnv(env.name)}
      >
      <div className="w-full flex items-center justify-between p-6 space-x-6 ">
        <div className="flex-1 truncate">
          <div className="flex">
            <div className="flex">
              <p className="text-sm font-bold capitalize">{env.name}</p>
              {env.builtIn &&
              <span className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-neutral-800 dark:text-neutral-400 text-xs font-medium bg-neutral-100 dark:bg-neutral-700 rounded-full">
                built-in
              </span>
              }
              {env.ephemeral &&
              <span
                className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-neutral-800 dark:text-neutral-400 text-xs font-medium bg-neutral-100 dark:bg-neutral-700 rounded-full"
                title={`Environment will be disabled at ${exactDate}`}
                >
                  trial
              </span>
              }
              <span className={`flex-shrink-0 inline-block px-2 py-0.5 mx-1 ${isOnline ? 'text-teal-800 dark:text-teal-400' : 'text-red-700 dark:text-red-300'} text-xs font-medium ${isOnline ? 'bg-teal-100 dark:bg-teal-700' : 'bg-red-200 dark:bg-red-700'} rounded-full`}>
                {isOnline ? 'connected' : 'disconnected'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </li>
  )
}

export default EnvironmentCard;
