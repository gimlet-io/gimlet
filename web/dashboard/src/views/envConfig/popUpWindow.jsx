import { CheckIcon, ExclamationCircleIcon } from '@heroicons/react/solid'

const PopUpWindow = ({ hasAPIResponded, isError, errorMessage, isTimedOut }) => {

    const backgroundColor = () => {
        return !hasAPIResponded ? "bg-gray-600" : isError || isTimedOut ? "bg-red-600" : "bg-green-600"
    }

    const savingText = () => {
        if (!hasAPIResponded) {
            return (
                <>
                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none"
                        viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path className="opacity-75" fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                </>
            )
        } else if (isTimedOut) {
            return (
                <>
                    <ExclamationCircleIcon className="mr-1 h-5 w-5" />
                    <div>Saving failed: The process has timed out.</div>
                </>
            )
        } else if (isError) {
            return (
                <>
                    <ExclamationCircleIcon className="mr-1 h-5 w-5" />
                    <div>Something went wrong: {errorMessage}.</div>
                </>
            )
        } else {
            return (
                <>
                    <CheckIcon className="mr-1 h-5 w-5" />
                    <div>Config saved succesfully!</div>
                </>
            )
        }
    }

    return (
        <div
            className="fixed inset-0 flex px-4 py-6 pointer-events-none sm:p-6 w-full flex-col items-end space-y-4">
            <div
                className={backgroundColor() + ` max-w-lg w-full text-gray-100 text-sm shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden`}>
                <div className="flex p-4">
                    {savingText()}
                </div>
            </div>
        </div>
    );
};

export default PopUpWindow;