import { XMarkIcon } from '@heroicons/react/20/solid';
import React, { useEffect } from 'react';

export function Modal(props) {
  const { closeHandler, children } = props;
  let {w, h} = props

  useEffect(() => {
    document.body.style.overflow = 'hidden';
    document.body.style.paddingRight = '15px';
    return () => { document.body.style.overflow = 'unset'; document.body.style.paddingRight = '0px' }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!w) {
    w = "w-3/5"
  }

  if (!h) {
    h = "h-4/5"
  }

  return (
    <div
      className="fixed flex inset-0 z-40 bg-neutral-800 bg-opacity-25 dark:bg-neutral-500 dark:bg-opacity-75"
      onClick={closeHandler}
    >
      <div className="flex self-center items-center justify-center w-full h-full p-8">
        <div className={`transform flex flex-col overflow-hidden bg-neutral-200 dark:bg-neutral-600 rounded-xl max-h-full ${w} ${h} pt-8`}
          onClick={e => e.stopPropagation()}
        >
          <div className="absolute top-0 right-0 p-1.5">
            <button
              className="rounded-md inline-flex text-neutral-800 hover:bg-neutral-300 dark:text-neutral-200 dark:hover:text-neutral-500 focus:outline-none"
              onClick={closeHandler}
            >
              <span className="sr-only">Close</span>
              <XMarkIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div className="h-full relative overflow-y-auto bg-neutral-100 dark:bg-neutral-900 rounded-b-lg font-normal">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}

export const SkeletonLoader = () => {
  return (
    <div className="w-full max-w-4xl animate-pulse space-y-3">
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-4/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-4/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/6"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
    </div>
  )
}
