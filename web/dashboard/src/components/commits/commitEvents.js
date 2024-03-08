import { DocumentIcon, CogIcon, CloudIcon } from '@heroicons/react/solid'
import { EventWidget } from './eventWidget';
import { format, formatDistance } from "date-fns";

export function CommitEvents(props) {
  const { events, scmUrl, envs } = props

  return (
    <div>
      <div className="flow-root">
        <ul className="-mb-8">
          {events.map((event, eventIdx) => (
            <CommitEvent
              key={event.created}
              event={event}
              last={eventIdx !== events.length - 1}
              scmUrl={scmUrl}
              envs={envs}
            />
          ))}
        </ul>
      </div>
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
          <span className="absolute left-4 top-4 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true" />
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
            {event.type !== 'release' &&
            <>
            <div className="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
              <EventWidget event={event} />
            </div>
            <div>
              <ResultsWidget event={event} />
            </div>
            </>
            }
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
      <ul>
        {event.results.map((result, idx) => (
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

function ResultsWidget(props) {
  const { event } = props

  if (event.type === 'artifact' && event.results) {
    return (
      <ul>
        {event.results.map((result, idx) => (
          <li key={idx}>
            <Result 
              result={result}
              builtInEnv={false}
              scmUrl={"https://github.com"}
            />
          </li>
        ))}
      </ul>
    )
  }

  if (event.type === 'imageBuild' && event.results) {
    return (
      <ul>
        {event.results.map((result, idx) => (
          <li key={idx}>
            <span>{result.triggeredBy}??</span>
            <span>{result.status}</span>
            {/* <div>{result.log}</div> */}
          </li>
        ))}
      </ul>
    )
  }

  return (
    null
  )
}

function Result(props) {
  const { scmUrl, builtInEnv, result } = props

  return (
    <>
      <p className={`${result.status === 'failure' ? 'text-red-500' : ''}`}>
        <span>
          {result.triggeredBy !== 'policy' &&
          <img
            className="inline rounded-full overflow-hidden"
            src={`${scmUrl}/${result.triggeredBy}.png?size=128`}
            alt={result.triggeredBy}
            width="20"
            height="20"
          />
          }
          {result.triggeredBy === 'policy' &&
            <>Policy</>
          }
        </span>
        <span className='pl-1'>triggered {result.env}/{result.app}</span>
        {result.gitopsRef &&
        <span className='font-mono text-sm pl-4'> 
          {builtInEnv &&
            <span>📎{result.gitopsRef.slice(0, 6)}</span>
          }
          {!builtInEnv &&
          <a
            href={`${scmUrl}/${result.gitopsRepo}/commit/${result.gitopsRef}`}
            target="_blank" rel="noopener noreferrer"
            className='ml-1'
          >
            📎{result.gitopsRef.slice(0, 6)}
          </a>
          }
        </span>
        }
      </p>
      {result.status === 'failure' &&
      <p className='text-red-500'>
        <span>❗</span>
        <span className='pl-1'>{result.statusDesc}</span>
      </p>
      }
    </>
  );
}
