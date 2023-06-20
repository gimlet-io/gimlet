import { useState } from 'react'
import { XIcon } from '@heroicons/react/outline'

export default function DeployPanel() {
  const [open, setOpen] = useState(true)

  return (
    <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true">
        <div className="fixed inset-x-0 bottom-0 h-2/5 z-40 bg-gray-800 text-gray-100">
          <div className="absolute top-0 right-0 p-4">
            <button type="button" className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
              <span className="sr-only">Close panel</span>
              <XIcon className="h-5 w-5" aria-hidden="true"/>
            </button>
          </div>
          <div className="mt-12 pb-20 px-6 overflow-y-scroll h-full w-full">
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello</div>
            <div>Hello1</div>
            <div>Hello1</div>
            <div>Hello1</div>
            <div>Hello1</div>
            <div>Hello2</div>
            <div>Hello2</div>
            <div>Hello2</div>
            <div>Hello2</div>
            <div>Hello3</div>
            <div>Hello3</div>
            <div>Hello3</div>
            <div>Hello3</div>
          </div>
        </div>
    </div>
  )
}
