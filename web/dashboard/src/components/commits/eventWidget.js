import { format, formatDistance } from "date-fns";

export function EventWidget(props) {
  const { event } = props

  const exactDate = format(event.created * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(event.created * 1000, new Date());

  let label = ""
  switch (event.type) {
    case 'artifact':
      label = event.status === 'error'
        ? 'Build artifact failed to process'
        : 'Build artifact received';
      if (event.status === 'processed' && event.results) {
        label = label + ' - ' + event.results.length + ' policy triggered'
      }
      break;
    case 'release':
      if (event.status === 'new') {
        label = 'Release triggered'
      } else if (event.status === 'processed') {
        label = 'Released'
      } else {
        label = 'Release failure'
      }
      break
    case 'imageBuild':
      if (event.status === 'new') {
        label = 'Image build requested'
      } else if (event.status === 'success') {
        label = 'Image built'
      } else {
        label = 'Image build error'
      }
      break;
    case 'rollback':
      if (event.status === 'new') {
        label = 'Rollback initiated'
      } else if (event.status === 'processed') {
        label = 'Rolled back'
      } else {
        label = 'Rollback error'
      }
      break
    default:
      label = ""
  }

  return (
    <>
      <span className="">{label}</span><span title={exactDate}> {dateLabel} ago</span>
    </>
  )
}
