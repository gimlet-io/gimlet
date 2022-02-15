const EnvironmentCard = ({ singleEnv }) => {
    return (
        <div className='bg-white overflow-hidden shadow rounded-lg my-4 w-fullpx-4 py-5 sm:px-6 focus:outline-none'>
            <div className='inline-grid'>
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                    {singleEnv.name}
                </h3>
            </div>
        </div>
    )
}

export default EnvironmentCard;
