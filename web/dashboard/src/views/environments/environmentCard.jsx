const EnvironmentCard = ({ isOnline, singleEnv, deleteEnv }) => {
    return (
        <div className="bg-white overflow-hidden shadow rounded-lg my-4 w-full px-4 py-5 sm:px-6 focus:outline-none">
            <div className="flex justify-between">
                <div className="inline-flex">
                    <h3 className="text-lg leading-6 font-medium text-gray-900 pr-1">
                        {singleEnv.name}
                    </h3>
                    <span title={isOnline ? "Connected" : "Disconnected"}>
                        <svg className={(isOnline ? "text-teal-400" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
                            <path
                                d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                            />
                        </svg>
                    </span>
                </div>
                {!isOnline &&
                    <div className="inline-flex">
                        <a className="cursor-pointer text-gray-500 hover:text-gray-700 mr-2"
                            target="_blank"
                            rel="noreferrer"
                            href="https://gimlet.io/docs/installing-gimlet-agent">
                            Install agent
                        </a>
                        <svg xmlns="http://www.w3.org/2000/svg" onClick={deleteEnv} className="cursor-pointer inline text-red-400 hover:text-red-600 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </div>
                }
            </div>
        </div>
    )
};

export default EnvironmentCard;
