import { ExclamationCircleIcon } from '@heroicons/react/24/solid'

export default function Error(props) {
  const { children } = props;

  return (
    <div className="flex items-center text-red-600 text-sm space-x-1">
      <ExclamationCircleIcon className="h-5 w-5 inline fill-current" aria-hidden="true" />
      <p>{children}</p>
    </div>
  )
}
