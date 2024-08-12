import { format, formatDistance } from "date-fns";

export function EventWidget(props) {
  const { event } = props

  const exactDate = format(event.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(event.created * 1000, new Date());

  let label = ""
  let color = ""

  switch (event.type) {
    case 'artifact':
      switch(event.status) {
        case 'error':
          label = 'Failed to process'
          color = 'text-red-500'
          break;
        case 'new':
          label = spinnerWithStatus('Processing...')
          break;
        case 'processed':
          if (event.results) {
            if (event.results.length === 1) {
              if (event.results[0].triggeredImageBuildId) {
                label = `Image build triggered`
              } else {
                let errorCount = 0;
                let successCount = 0;
                let hasError = false
                event.results.forEach((result) => {
                  if (result.status === 'failure') {
                    errorCount++
                  } else {
                    successCount++
                  }
                });
                if (errorCount > 0) {
                  label = `Deployment policy error`
                  color = 'text-red-500'
                } else {
                  label = `${successCount} Deployment ${successCount > 1 ? 'policies' : 'policy'} triggered`
                  color = 'text-teal-400'
                }
              }
            } else {
              label = `${event.results.length} deployments triggered`
              color = 'text-teal-400'
            }
          } else {
            label = "No policy triggered"
          }
          break;
      }
      break;
    case 'release':
      if (event.status === 'new') {
        label = spinnerWithStatus('Releasing...')
      } else if (event.status === 'processed') {
        label = 'Released'
        color = 'text-teal-400'
      } else {
        label = 'Release failure'
        color = 'text-red-500'
      }
      break
    case 'imageBuild':
      if (event.status === 'new') {
        label = spinnerWithStatus('Building image...')
      } else if (event.status === 'success') {
        label = 'Image built'
      } else {
        label = 'Image build error'
        color = 'text-red-500'
      }
      break;
    case 'rollback':
      if (event.status === 'new') {
        label = spinnerWithStatus('Rolling back...')
      } else if (event.status === 'processed') {
        label = 'Rolled back'
        color = 'text-teal-400'
      } else {
        label = 'Rollback error'
        color = 'text-red-500'
      }
      break
    default:
      label = ""
      color = ""
  }

  return (
    <>
        <span className={`uppercase font-bold text-sm ${color}`}>{label}</span>
        <span title={exactDate}> {dateLabel} ago</span>
    </>
  )
}

function spinnerWithStatus(status) {
  return (
    <>
      <svg className="animate-spin h-6 w-6 text-black dark:text-white inline mr-1 mb-1" xmlns="http://www.w3.org/2000/svg" fill="none"
          viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
          <path className="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <span className={`uppercase font-bold text-sm text-blue-500`}>{status}</span>
    </>
  )
}
