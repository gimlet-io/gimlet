import { DocumentIcon, CogIcon, CloudIcon } from '@heroicons/react/24/solid'
import { format, formatDistance } from "date-fns";

export function CommitEvents(props) {
  const { events, scmUrl, envs } = props

  return (
    <div className="flow-root p-2">
      <ul>
        {events.map((event, eventIdx) => (
          <CommitEvent
            key={eventIdx}
            event={event}
            last={eventIdx !== events.length - 1}
            scmUrl={scmUrl}
            envs={envs}
          />
        ))}
      </ul>
    </div>
  )
}

function CommitEvent(props) {
  const { event, last, scmUrl, envs } = props
  let color = 'bg-green-500'
  let TypeIcon = CloudIcon
  if (event.type === 'artifact') {
    color = 'bg-blue-500'
    TypeIcon = DocumentIcon
  } else if (event.type === 'imageBuild') {
    color = 'bg-purple-500'
    TypeIcon = CogIcon
  }

  return (
    <li key={event.created}>
      <div className="relative pb-8">
        {last &&
          <span className="absolute left-4 top-4 -ml-px h-full w-0.5 bg-neutral-200" aria-hidden="true" />
        }
        <div className="relative flex space-x-3">
          <div>
            <span className={`${color} h-8 w-8 rounded-full flex items-center justify-center ring-8 ring-white`}>
              <TypeIcon className="h-5 w-5 text-white" aria-hidden="true" />
            </span>
          </div>
          <div className='grow'>
            {event.type === 'release' &&
              <ReleaseEventWidget event={event} scmUrl={scmUrl} envs={envs} />
            }
            {event.type === 'artifact' &&
              <ArtifactEventWidget event={event} scmUrl={scmUrl} envs={envs} />
            }
            {event.type === 'imageBuild' &&
              <ImageBuildEventWidget event={event} scmUrl={scmUrl} />
            }
            {/* {event.type === 'rollback' &&
              <p>rollback events are not associated with commit shas, they work with gitops commits</p>
            } */}
          </div>
        </div>
      </div>
    </li>
  )
}

function ReleaseEventWidget(props) {
  const {event, scmUrl, envs} = props

  const exactDate = format(event.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(event.created * 1000, new Date());

  const triggeredBy = event.releaseRequest.triggeredBy
  const env = event.releaseRequest.env
  const builtInEnv = envs.filter(e => e.name ===env).builtIn

  return (
    <div>
    <div className="flex min-w-0 flex-1 justify-between space-x-4">
      <span>
        <img
          className="inline rounded-full overflow-hidden mr-1"
          src={`${scmUrl}/${triggeredBy}.png?size=128`}
          alt={triggeredBy}
          width="20"
          height="20"
        />
        <span className='font-medium'>{triggeredBy}</span>
        <span className='px-1'>deployed to</span>
        <span className='font-medium'>{env}</span>
      </span>
      <span title={exactDate}> {dateLabel} ago</span>
    </div>
      <p className='pl-5'>Status: {event.status}</p>
      <ul>
        {event.results && event.results.map((result, idx) => (
          <li key={idx}>
            <p className={`pl-5 ${result.status === 'failure' ? 'text-red-500' : ''}`}>
              {result.gitopsRef &&
              <span className='font-mono text-sm'> 
                {builtInEnv &&
                  <span>📎 {result.gitopsRef.slice(0, 6)}</span>
                }
                {!builtInEnv &&
                <a
                  href={`${scmUrl}/${result.gitopsRepo}/commit/${result.gitopsRef}`}
                  target="_blank" rel="noopener noreferrer"
                  className='ml-1'
                >
                  📎 {result.gitopsRef.slice(0, 6)}
                </a>
                }
              </span>
              }
              <span className='pl-1'>{result.app}</span>
            </p>
            {result.status === 'failure' &&
            <p className='pl-5 text-red-500'>
              <span>❗</span>
              <span className='pl-1'>{result.statusDesc}</span>
            </p>
            }
          </li>
        ))}
      </ul>
    </div>
  )
}

function ArtifactEventWidget(props) {
  const {event, scmUrl, envs} = props

  const exactDate = format(event.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(event.created * 1000, new Date());

  return (
    <div>
    <div className="flex min-w-0 flex-1 justify-between space-x-4">
      <span>
        <span>Build artifact received</span>
        {event.results &&
        <span> - {event.results.length} {event.results.length === 1 ? 'policy' : 'policies'} triggered</span>
        }
      </span>
      <span title={exactDate}> {dateLabel} ago</span>
    </div>
      <p className='pl-5'>Status: {event.status}</p>
      <ul>
        {event.results?.map((result, idx) => {
          const env = result.env
          const builtInEnv = envs.filter(e => e.name === env).builtIn

          return (
            <li key={idx}>
              <p className={`pl-5 ${result.status === 'failure' ? 'text-red-500' : ''}`}>
                {result.triggeredImageBuildId &&
                  <span>Image Build triggered for</span>
                }
                {result.gitopsRef &&
                <span className='font-mono text-sm'> 
                  {builtInEnv &&
                    <span>📎 {result.gitopsRef.slice(0, 6)}</span>
                  }
                  {!builtInEnv &&
                  <a
                    href={`${scmUrl}/${result.gitopsRepo}/commit/${result.gitopsRef}`}
                    target="_blank" rel="noopener noreferrer"
                    className='ml-1'
                  >
                    📎 {result.gitopsRef.slice(0, 6)}
                  </a>
                  }
                </span>
                }
                <span className='pl-1'>{result.app}</span>
              </p>
              {result.status === 'failure' &&
              <p className='pl-5 text-red-500'>
                <span>❗</span>
                <span className='pl-1'>{result.statusDesc}</span>
              </p>
              }
            </li>
          )
        })}
      </ul>
    </div>
  )
}

function ImageBuildEventWidget(props) {
  const {event, scmUrl} = props

  const exactDate = format(event.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(event.created * 1000, new Date());

  const triggeredBy = event.imageBuildRequest.triggeredBy
  const env = event.imageBuildRequest.env
  const app = event.imageBuildRequest.app
  const dockerfile = event.imageBuildRequest.dockerfile

  return (
    <div>
    <div className="flex min-w-0 flex-1 justify-between space-x-4">
      <span>
        <img
          className="inline rounded-full overflow-hidden mr-1"
          src={`${scmUrl}/${triggeredBy}.png?size=128`}
          alt={triggeredBy}
          width="20"
          height="20"
        />
        <span className='font-medium'>{triggeredBy}</span>
        <span className='px-1'>triggered an image build to</span>
        <span className='font-medium'>{env}/{app}</span>
        <span className='px-1'>using</span>
        <span className='text-sm font-mono'>{dockerfile}</span>
      </span>
      <span title={exactDate}> {dateLabel} ago</span>
    </div>
      <p className='pl-5'>Status: {event.status}</p>
      <ul>
        {event.results?.map((result, idx) => (
          <li key={idx}>
            {result.status === 'failure' &&
            <p className='pl-5 text-red-500'>
              <span>❗</span>
              <span className='pl-1'>{result.statusDesc}</span>
            </p>
            }
            <div className="overflow-y-auto overscroll-none flex-grow h-64 bg-stone-900 text-neutral-300 font-mono text-sm p-2" style={{"whiteSpace": 'pre-line'}}>
              {result.log}
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
