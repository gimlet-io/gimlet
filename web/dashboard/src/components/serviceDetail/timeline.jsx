//mocked data
const alertsMock = [
  {
    type: "imagePullBackOffThreshold",
    name: "default/react-7ajs9w-dummy-1",
    deploymentName: "default/react-7ajs9w",
    status: "Resolved",
    createdAt: 1695283200,
    reachedAt: 1695283260,
    resolvedAt: 1695290460,
  },
  {
    type: "imagePullBackOffThreshold",
    name: "default/react-7ajs9w-dummy-2",
    deploymentName: "default/react-7ajs9w",
    status: "Resolved",
    createdAt: 1695031200,
    reachedAt: 1695031260,
    resolvedAt: 1695060060,
  },
  {
    type: "imagePullBackOffThreshold",
    name: "default/react-7ajs9w-dummy-3",
    deploymentName: "default/react-7ajs9w",
    status: "Resolved",
    createdAt: 1694851200,
    reachedAt: 1694851320,
    resolvedAt: 1694862120,
  },
];

const Timeline = ({ alerts = alertsMock, interval = 7 }) => {
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
    <div className="flex flex-col border border-gray-300 whitespace-no-wrap mt-2">
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
          const resolvedAt = new Date(alert.resolvedAt * 1000);

          if (createdAt < startDate) {
            // Skip alerts that are not within the last 7 days (TODO skip fetching alerts from db that older than 7 day)
            return null;
          }

          const startPosition = Math.max(0, (createdAt - startDate) / 86400000); // 86400000 milliseconds in a day
          const endPosition = Math.min(interval, (resolvedAt - startDate) / 86400000);
          const total = alert.resolvedAt - alert.createdAt
          const pendingInterval = alert.reachedAt - alert.createdAt
          const firingInterval = alert.resolvedAt - alert.reachedAt

          const alertStyle = {
            left: `${(startPosition / interval) * 100}%`,
            width: `${((endPosition - startPosition) / interval) * 100}%`,
          };

          return (
            <div
              key={index}
              className="absolute text-white bg-green text-xs overflow-hidden whitespace-nowrap"
              style={alertStyle}
            >
              <div className="mb-4 flex h-8 overflow-hidden bg-gray-100 text-xs">
                <div
                  //           title={`${format(alert.createdAt * 1000, 'h:mm:ss a, MMMM do yyyy')}
                  // ${formatDistance(alert.createdAt * 1000, new Date())}`}
                  style={{ width: `${pendingInterval / total * 100}%` }}
                  className="bg-yellow-300 transition-all duration-500 ease-out"
                ></div>
                <div
                  //           title={`${format(alert.reachedAt * 1000, 'h:mm:ss a, MMMM do yyyy')}
                  // ${formatDistance(alert.reachedAt * 1000, new Date())}`}
                  style={{ width: `${firingInterval / total * 100}%` }}
                  className="bg-red-400 transition-all duration-500 ease-out"
                ></div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default Timeline;
