import React, { useEffect } from 'react';
import {XIcon} from '@heroicons/react/outline'

export function Modal(props) {
  const { closeHandler, children } = props;

  useEffect(() => {
    document.body.style.overflow = 'hidden';
    document.body.style.paddingRight = '15px';
    return () => { document.body.style.overflow = 'unset'; document.body.style.paddingRight = '0px' }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div
      className="fixed flex inset-0 z-40 bg-stone-800 bg-opacity-25"
      onClick={closeHandler}
    >
      <div className="flex self-center items-center justify-center w-full h-full p-8">
        <div className="transform flex flex-col overflow-hidden bg-gray-200 rounded-xl h-[90vh] max-h-full w-[90vw] pt-8"
          onClick={e => e.stopPropagation()}
        >
          <div className="absolute top-0 right-0 p-1.5">
            <button
              className="rounded-md inline-flex text-gray-800 hover:bg-gray-300 focus:outline-none"
              onClick={closeHandler}
            >
              <span className="sr-only">Close</span>
              <XIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div className="h-full relative overflow-y-auto p-4 bg-stone-100 rounded-b-lg font-normal">
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
      <div className="h-2 bg-slate-700 rounded w-1/5"></div>
      <div className="h-2 bg-slate-700 rounded w-2/5"></div>
      <div className="h-2 bg-slate-700 rounded w-3/5"></div>
      <div className="h-2 bg-slate-700 rounded w-4/5"></div>
      <div className="h-2 bg-slate-700 rounded w-4/5"></div>
      <div className="h-2 bg-slate-700 rounded w-3/5"></div>
      <div className="h-2 bg-slate-700 rounded w-2/5"></div>
      <div className="h-2 bg-slate-700 rounded w-1/5"></div>
      <div className="h-2 bg-slate-700 rounded w-2/5"></div>
      <div className="h-2 bg-slate-700 rounded w-2/5"></div>
      <div className="h-2 bg-slate-700 rounded w-1/5"></div>
      <div className="h-2 bg-slate-700 rounded w-1/5"></div>
      <div className="h-2 bg-slate-700 rounded w-1/5"></div>
      <div className="h-2 bg-slate-700 rounded w-1/6"></div>
      <div className="h-2 bg-slate-700 rounded w-2/5"></div>
      <div className="h-2 bg-slate-700 rounded w-3/5"></div>
    </div>
  )
}
