const DeleteButton = ({ envName, deleteFunc }) => {
  return (
    <button
      onClick={() => {
        // eslint-disable-next-line no-restricted-globals
        confirm(`Are you sure you want to delete ${envName}?`) &&
          deleteFunc()
      }}
      className="h-10 bg-red-500 hover:bg-red-400 focus:outline-none focus:shadow-outline-indigo gap-x-1.5 px-4 py-2.5 inline-flex items-center border border-gray-300 text-sm leading-6 font-medium rounded-md text-white transition ease-in-out duration-150">
      <svg xmlns="http://www.w3.org/2000/svg" className="cursor-pointer inline h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
      </svg>
      Delete
    </button>
  )
}

export default DeleteButton;
