import { format, formatDistance } from "date-fns";

const Timeline = ({ alerts, interval = 7 }) => {
  if (!alerts) {
    return null
  }
  
  const endDate = new Date();
  const startDate = new Date(endDate);
  startDate.setDate(endDate.getDate() - interval);

  const dateLabels = [];
  for (let i = 0; i <= interval; i++) {
    const labelDate = new Date(startDate);
    labelDate.setDate(startDate.getDate() + i);
    dateLabels.push(labelDate);
  }

  return (
    <div className="p-2">
      <div className="flex flex-col border border-gray-300 whitespace-no-wrap">
        <div className="flex border-b border-gray-300">
          {dateLabels.map((date, index) => (
            <div key={index} className="flex-1 p-2 text-center">
              {date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
            </div>
          ))}
        </div>
        <div className="relative flex h-8 bg-green-400">
          {alerts.map((alert, index) => {
            // Convert Unix timestamps to milliseconds
            const createdAt = new Date(alert.createdAt * 1000);
            const resolvedAt = new Date(alert.resolvedAt ? (alert.resolvedAt * 1000) : Date.now());
            const startPosition = Math.max(0, (createdAt - startDate) / 86400000); // 86400000 milliseconds in a day
            const endPosition = Math.min(interval, (resolvedAt - startDate) / 86400000);

            const alertStyle = {
              left: `${(startPosition / interval) * 100}%`,
              width: `${((endPosition - startPosition) / interval) * 100}%`,
            };

            return (
              <div
                key={index}
                className="absolute"
                style={alertStyle}
              >
                <Alert
                  endDate={endDate}
                  alert={alert}
                />
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

const Alert = ({ alert, endDate }) => {
  const endDateUnix = (new Date(endDate).getTime() / 1000).toFixed(0)
  const total = (alert.resolvedAt ?? endDateUnix) - alert.createdAt
  let pendingInterval = alert.reachedAt - alert.createdAt
  let firingInterval = (alert.resolvedAt ?? endDateUnix) - alert.reachedAt

  return (
    <div
      title={`${alert.name} reached at
${format(alert.createdAt * 1000, 'h:mm:ss a, MMMM do yyyy')}
${formatDistance(alert.createdAt * 1000, new Date())}`}
      className="flex h-8 bg-slate-300">
      <div
        style={{ width: `${pendingInterval / total * 100}%` }}
        className="bg-yellow-300 transition-all duration-500 ease-out"
      ></div>
      <div
        style={{ width: `${firingInterval / total * 100}%` }}
        className="bg-red-400 transition-all duration-500 ease-out"
      ></div>
    </div>
  )
};

export default Timeline;
