const variables = [
    { variable: 'REPO', type: 'String', value: 'The owner and repository name.', },
    { variable: 'OWNER', type: 'String', value: "The repository owner's name.", },
    { variable: 'BRANCH', type: 'String', value: 'The name of the Git branch currently being built.', },
    { variable: 'TAG', type: '	String', value: 'The name of the git tag, if the current build is tagged.', },
    { variable: 'SHA', type: 'String', value: 'The commit SHA that triggered the workflow.', },
    { variable: 'ACTOR', type: 'String', value: 'The name of the person or app that initiated the workflow.', },
    { variable: 'EVENT', type: 'String', value: 'The name of the event that triggered the workflow. ', },
    { variable: 'JOB', type: 'String', value: 'A unique identifier for the current job.', },
]

export default function EnvVarsTable() {
    return (
        <div>
            <div className="sm:flex sm:items-center">
                <p className="text-sm text-gray-500">
                    A list of all the variables used by CI.
                </p>
            </div>
            <div className="mt-4 flex flex-col">
                <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
                    <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                        <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                            <table className="min-w-full divide-y divide-gray-300">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">
                                            Variable
                                        </th>
                                        <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                                            Type
                                        </th>
                                        <th scope="col" className="px-3 py-3.5 text-center text-sm font-semibold text-gray-900">
                                            Value
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white">
                                    {variables.map((variable, variableIdx) => (
                                        <tr key={variable.variable} className={variableIdx % 2 === 0 ? undefined : 'bg-gray-50'}>
                                            <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                                                {variable.variable}
                                            </td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{variable.type}</td>
                                            <td className="relative whitespace-nowrap px-3 py-4 text-center text-sm text-gray-500">{variable.value}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}