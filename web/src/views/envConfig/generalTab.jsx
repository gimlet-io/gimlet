import Toggle from '../../components/toggle/toggle';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import Dropdown from '../../components/dropdown/dropdown';
import Error from '../../components/error/error';

export function Generaltab(props) {
  const { config, action, preview, configExists, invalidAppName } = props
  const { configFile, setAppName, setNamespace, deleteApp } = props
  const { toggleDeployPolicy, setDeployEvent, setDeployFilter } = props
  const { templates, selectedTemplate, setDeploymentTemplate } = props
  const deployEvents = ["push", "tag", "pr"]

  return (
    <>
      {preview &&
      <div className='mb-8 w-full card'>
        <div className="p-6 pb-4 items-center">
          <div className="block font-medium">Preview Deploy</div>
          <div className="text-sm mt-2 mb-4 text-neutral-800 dark:text-neutral-400 leading-loose">
            Preview deploys are automatically deployed for Pull Requests. With a unique name on a unique URL.
          </div>
          <div className="max-w-lg flex rounded-md">
            <Toggle
              checked={configFile.preview}
              disabled
            />
          </div>
        </div>
        <div className='learnMoreBox'>
          Learn more about <a href="https://gimlet.io/docs/deployments/preview-deployments" className='learnMoreLink'>Preview Deploys <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
        </div>
      </div>
      }

      <div className='p-6 w-full card'>
        { action !== "new" && action !== 'new-preview' &&
        <div className="mb-4 items-center">
          <label htmlFor="environment" className="block font-medium">
            Environment
          </label>
          <div className='mt-2'>
            <input
              type="text"
              name="environment"
              id="environment"
              disabled
              value={configFile.env}
              className="block w-full input"
            />
          </div>
        </div>
        }
        <div className="mb-4 items-center">
          <label htmlFor="appName" className={`${(!configFile.app || configExists || invalidAppName) ? "text-red-600" : ""} block font-medium`}>
            Name
          </label>
          <div className='mt-2'>
            <input
              type="text"
              name="appName"
              id="appName"
              disabled={action !== "new"}
              value={configFile.app}
              onChange={e => setAppName(e.target.value)}
              className="block w-full input"
            />
          </div>
          {configExists && 
            <div className="mt-2">
              <Error>This application name is already in use.</Error>
            </div>
          }
          {invalidAppName && 
            <div className="mt-2">
              <Error>{`Name must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character, and must not be longer than 53 characters.
                (e.g. 'my-name', or 'abc-123', regex used for validation is '^[a-z]([a-z0-9-]{0,51}[a-z0-9])?$')`}</Error>
            </div>
          }
        </div>
      </div>

      <div className='mt-8 w-full card'>
        <div className="p-6 pb-4 items-center">
          <label htmlFor="environment" className="block font-medium">
            Deployment template
          </label>
          <div className="mt-4 grid grid-cols-1 gap-y-6 sm:grid-cols-3 sm:gap-x-4">
            {templates.map(template => <DeploymentTemplate
              key={templateIdentity(template)}
              template={template}
              selectedTemplate={selectedTemplate}
              setDeploymentTemplate={setDeploymentTemplate}
              disabled={action !== 'new' && action !== 'new-preview'} />
            )}
          </div>
        </div>
        <div className='learnMoreBox'>
          Learn more about <a href="https://gimlet.io/docs/deployment-settings/image-settings" className='learnMoreLink'>Deployment Templates <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
        </div>
      </div>

      <div className='mt-8 w-full card'>
        <div className="p-6 pb-4 items-center">
          <label htmlFor="namespace" className={`${!configFile.namespace ? "text-red-600" : ""} block font-medium`}>
            Namespace
          </label>
          <div className='mt-4'>
            <input
              type="text"
              name="namespace"
              id="namespace"
              value={configFile.namespace}
              onChange={e => setNamespace(e.target.value)}
              className="block w-full filter input"
            />
          </div>
        </div>
        <div className='learnMoreBox'>
          Learn more about <a href="https://gimlet.io/docs/kubernetes-resources/kubernetes-essentials#namespaces" className='learnMoreLink'>Kubernetes Namespaces <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
        </div>
      </div>

      {!preview && action !== "new" &&
      <div className='mt-8 w-full card'>
        <div className="p-6 pb-4 items-center">
          <div className="block font-medium">Automatic Deploys</div>
          <div className="text-sm mt-2 mb-4 text-neutral-800 dark:text-neutral-400 leading-loose">
            You can automate releases to your staging or production environment.
          </div>
          <div className="max-w-lg flex rounded-md">
            <Toggle
              checked={configFile.deploy !== undefined}
              onChange={e => toggleDeployPolicy()}
            />
          </div>
          {configFile.deploy &&
            <div className="space-y-4 pt-4">
              <div className="items-center">
                <label htmlFor="deployEvent" className="mr-4 mb-2 block text-sm font-medium">
                  Deploy event
                </label>
                <Dropdown
                  items={deployEvents}
                  value={configFile.deploy.event}
                  changeHandler={setDeployEvent}
                  onCard={true}
                />
              </div>
              <div className="items-center">
                <label htmlFor="deployFilterInput" className="mr-4 block text-sm font-medium">
                  {`${configFile.deploy.event === "tag" ? "Tag" : "Branch"} filter`}
                </label>
                <input
                  key={configFile.deploy.event}
                  type="text"
                  name="deployFilterInput"
                  id="deployFilterInput"
                  value={configFile.deploy.event === "tag" ? configFile.deploy.tag : configFile.deploy.branch}
                  onChange={e => { setDeployFilter(e.target.value) }}
                  className="input mt-2"
                />
                <ul className="list-none text-sm text-neutral-500 dark:text-neutral-400 mt-2">
                  {configFile.deploy.event === "tag" ?
                    <>
                      <li>
                        Deploy on tag patterns.
                      </li>
                      <li>
                        Use glob patterns like <code>`v1.*`</code> or negated conditions like <code>`!v2.*`</code>.
                      </li>
                    </>
                    :
                    <>
                      <li>
                        Deploy on branch name patterns.
                      </li>
                      <li>
                        Use glob patterns like <code>`feature/*`</code> or negated conditions like <code>`!main`</code>.
                      </li>
                    </>}
                </ul>
              </div>
            </div>
          }
        </div>
        <div className='learnMoreBox'>
          Learn more about <a href="https://gimlet.io/docs/deployments/automated-deployments" className='learnMoreLink'>Automatic Deploys <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
        </div>
      </div>
      }

      {action !== "new" && !preview &&
        <div className='mt-8 w-full redCard'>
          <div className="p-6 pb-4 items-center">
          <label htmlFor="environment" className="block font-medium">
            Delete Deployment Config
          </label>
          <p className='text-sm text-neutral-800 dark:text-neutral-400 mt-4'>Once you delete the {config} deployment config, automatic deploys stop
            and you will no longer be able to deploy new versions manually either.<br />
            <br />
            Existing deployed instances of this config will remain deployed, and you need to delete them manually.
          </p>
          </div>
          <div className='flex items-center w-full learnMoreRed'>
            <div className='flex-grow'>
            </div>
            <button
              type="button"
              className="destructiveButton"
              onClick={() => {
                // eslint-disable-next-line no-restricted-globals
                confirm(`Are you sure you want to delete the ${config} deployment configuration? (deployed app instances of this configuration will remain deployed, and you can delete them later)`) &&
                  deleteApp()
              }}
            >Delete</button>
          </div>
        </div>
      }
    </>
  )
}

export function templateIdentity(template) {
  return "tpl-"+
    template.reference.chart.name+
    template.reference.chart.repository+
    template.reference.chart.version
}

export function DeploymentTemplate(props) {
  const { template, selectedTemplate, setDeploymentTemplate, disabled } = props

  const key = templateIdentity(template)
  const selected = key === templateIdentity(selectedTemplate)

  const customChart = template.reference.title === ""
  const title = customChart ? "Custom" : template.reference.title

  return (
    <div
      key={key}
      className={`relative flex ${disabled ? "" : "cursor-pointer"} ${disabled && !selected ? "opacity-30" : ""} rounded-lg bg-white dark:bg-neutral-100 p-4 focus:outline-none text-neutral-500 border dark:border-2 ${selected ? "border-indigo-600" : "border-neutral-200"}`}
      onClick={() => !disabled && setDeploymentTemplate(template)}
      >
        <span className="flex flex-1">
          <span className="flex flex-col">
            <span className="block text-sm font-medium text-neutral-900 dark:text-neutral-800 select-none capitalize">{title}</span>
            <span className="mt-1 flex items-center text-sm text-neutral-500 select-none">{template.reference.description}</span>
            <span className="mt-3 flex items-center text-xs font-mono text-neutral-500 select-none">{customChart ? template.reference.chart.name : ""}</span>
          </span>
        </span>
        <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-indigo-600 ${selected ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
        </svg>
    </div>
  )
}
