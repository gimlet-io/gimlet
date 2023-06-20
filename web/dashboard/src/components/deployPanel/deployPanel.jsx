import { Fragment, useState } from 'react'
import { Dialog, Transition } from '@headlessui/react'
import { XIcon } from '@heroicons/react/outline'

export default function DeployPanel() {
  const [open, setOpen] = useState(true)

  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog as="div" className="relative" onClose={setOpen}>
        <div className="fixed inset-x-0 bottom-0" />

        <div className="fixed inset-x-0 bottom-0 overflow-hidden h-full z-0">
          <div className="absolute inset-x-0 bottom-0 overflow-hidden h-1/2 z-10">
            <div className="pointer-events-none fixed inset-x-0 right-0 flex h-1/2">
              <Transition.Child
                as={Fragment}
                enter="transform transition ease-in-out duration-500 sm:duration-700"
                enterFrom="translate-y-full"
                enterTo="translate-y-0"
                leave="transform transition ease-in-out duration-500 sm:duration-700"
                leaveFrom="translate-y-0"
                leaveTo="translate-y-full"
              >
                <Dialog.Panel className="pointer-events-auto w-full">
                  <div className="flex flex-col bg-gray-800 text-gray-100 py-4 shadow-xl h-full">
                    <div className="px-4 sm:px-6">
                      <div className="flex items-start justify-between">
                        <Dialog.Title className="text-base font-semibold leading-6">
                        </Dialog.Title>
                        <div className="ml-3 flex h-7 items-center">
                          <button
                            type="button"
                            className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                            onClick={() => setOpen(false)}
                          >
                            <span className="sr-only">Close panel</span>
                            <XIcon className="h-6 w-6 fill-current" aria-hidden="true" />
                          </button>
                        </div>
                      </div>
                    </div>
                    <div className="relative mt-6 flex-1 px-4 sm:px-6 overflow-y-scroll h-full">
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
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  )
}
