import { Switch } from '@headlessui/react'

export default function Toggle(props) {
  const { checked, onChange, disabled } = props;

  return (
    <Switch
      disabled={disabled}
      checked={checked}
      onChange={onChange}
      className={`${checked ? 'bg-blue-500 dark:bg-blue-800' : 'bg-neutral-200 dark:bg-neutral-500'} disabled:bg-neutral-200 dark:disabled:bg-neutral-500 disabled:cursor-default relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200`}
    >
      <span className="sr-only">Use setting</span>
      <span
        aria-hidden="true"
        className={`${checked ? 'translate-x-5' : 'translate-x-0'} pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200`}
      />
    </Switch>
  )
}
