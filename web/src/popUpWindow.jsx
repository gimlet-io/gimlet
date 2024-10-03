import { CheckIcon, XCircleIcon } from '@heroicons/react/24/solid'

export function InProgress(props) {
  return (
    <div className="flex z-50">
      <div className="flex-shrink-0">
        <svg className="animate-spin h-5 w-5 text-black" xmlns="http://www.w3.org/2000/svg" fill="none"
          viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
          <path className="opacity-75" fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-medium text-neutral-800">{props.header}</h3>
      </div>
    </div>
  )
}

export function Success(props) {
  return (
    <div className="flex z-50">
      <div className="flex-shrink-0">
        <CheckIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-semibold text-green-800">{props.header}</h3>
        <div className="mt-2 text-sm text-green-700">
          <div>{props.message}</div>
          <a
            href={props.link}
            rel="noreferrer"
            target="_blank"
            className="text-sm text-green-700 hover:text-green-800 underline">
            {props.link}
          </a>
        </div>
      </div>
    </div>
  )
}

export function Error(props) {
  return (
    <div className="flex z-50">
      <div className="flex-shrink-0">
        <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-medium text-red-800">{props.header}</h3>
        <div className="mt-2 text-sm text-red-700">
          <p>{props.message}</p>
        </div>
      </div>
    </div>
  )
}
