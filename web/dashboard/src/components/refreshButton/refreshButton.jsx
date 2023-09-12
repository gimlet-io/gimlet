const RefreshButton = ({ refreshFunc, title = "Refresh" }) => {
  return (
    <button
      onClick={refreshFunc}
      className="h-11 bg-gray-500 hover:bg-gray-400 focus:outline-none focus:shadow-outline-indigo gap-x-1.5 px-4 py-2.5 inline-flex items-center border border-gray-300 text-sm leading-6 font-medium rounded-md text-white transition ease-in-out duration-150">
      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      {title}
    </button>
  )
};

export default RefreshButton;
