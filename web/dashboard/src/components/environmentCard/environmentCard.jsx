function EnvironmentCard({ name, builtIn, navigateToEnv, isOnline }) {
  return (
    <li className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
      <div className="w-full flex items-center justify-between p-6 space-x-6 cursor-pointer"
        onClick={() => navigateToEnv(name)}>
        <div className="flex-1 truncate">
          <div className="flex justify-between">
            <div className="flex">
              <p className="text-sm font-bold capitalize">{name}</p>
              {builtIn &&
                <span
                  className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-gray-800 text-xs font-medium bg-gray-100 rounded-full">
                  built-in
                </span>
              }
            </div>
            <svg className={(isOnline ? "text-green-400" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
              <path
                d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
              />
            </svg>
          </div>
          <div className="p-2">
          </div>
        </div>
      </div>
    </li>
  )
}

export default EnvironmentCard;
