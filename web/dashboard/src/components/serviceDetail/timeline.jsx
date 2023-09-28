import { useState } from "react";
import { format, formatDistance } from "date-fns";

const Timeline = ({ alerts }) => {
  const [hours, setHours] = useState(1)

  if (!alerts) {
    return null;
  }

  const endDate = new Date();
  const startDate = new Date();
  startDate.setHours(endDate.getHours() - hours);

  return (
    <div className="p-2 h-8 mb-4">
      <div className="flex justify-end divide-x space-x-2 divide-gray-300 text-gray-500 text-xs">
        <button
          onClick={() => setHours(24)}>
          last 24 hours
        </button>
        <button className="pl-2"
          onClick={() => setHours(6)}>
          last 6 hours
        </button>
        <button className="pl-2"
          onClick={() => setHours(1)}>
          last hour
        </button>
      </div>
      <div className="relative flex bg-green-400 h-4">
        {alerts.map((alert, index) => {
          const pendingAt = new Date(alert.pendingAt * 1000);
          const resolvedAt = new Date(alert.resolvedAt ? (alert.resolvedAt * 1000) : Date.now());
          const startPosition = Math.max(0, (pendingAt - startDate) / (60 * 60 * 1000));
          const endPosition = Math.min(hours, (resolvedAt - startDate) / (60 * 60 * 1000));

          const endDateUnix = (new Date(endDate).getTime() / 1000).toFixed(0)
          const total = (alert.resolvedAt ?? endDateUnix) - alert.pendingAt
          const pendingInterval = alert.firedAt - alert.pendingAt
          const firingInterval = (alert.resolvedAt ?? endDateUnix) - alert.firedAt

          const alertStyle = {
            left: `${(startPosition / hours) * 100}%`,
            width: `${((endPosition - startPosition) / hours) * 100}%`,
          };

          let title = ""
          if (alert.firedAt) {
            title = `${alert.objectName} fired at
${format(alert.firedAt * 1000, 'h:mm:ss a, MMMM do yyyy')}
${formatDistance(alert.firedAt * 1000, new Date())} ago`
          }

          return (
            <div
              key={index}
              className="absolute"
              style={alertStyle}
            >
              <div
                title={title}
                className="flex h-4 bg-slate-300">
                <div
                  style={{ width: `${pendingInterval / total * 100}%` }}
                  className="bg-yellow-400 transition-all duration-500 ease-out"
                ></div>
                <div
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
