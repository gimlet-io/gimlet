import { useState } from "react";
import { format, formatDistance } from "date-fns";

const Timeline = ({ alerts }) => {
  const [hours, setHours] = useState([
    { hour: 24, current: false },
    { hour: 6, current: false },
    { hour: 1, current: true }
  ]);
  const selected = "font-semibold"

  if (!alerts) {
    return null;
  }

  const hourHandler = (input) => {
    setHours(hours.map(hour => {
      if (hour.hour === input) {
        return { ...hour, current: true }
      } else {
        return { ...hour, current: false }
      }
    }))
  }

  const currentHour = hours.find(hour => hour.current)
  console.log(currentHour.hour)
  const endDate = new Date();
  const startDate = new Date();
  startDate.setHours(endDate.getHours() - currentHour.hour);

  return (
    <div className="p-2">
      <div className="h-8">
        <div className="flex justify-end divide-x space-x-1 divide-gray-300 text-gray-500 text-xs">
          {hours.map(hour => {
            return (
              <button
                  key={hour.hour}
                  type="button"
                  onClick={() => hourHandler(hour.hour)}
                  className={(hour.current ? selected : "") + ' pl-1'}
                >
                  latest {hour.hour} hours
                </button>
            )
          })}
        </div>
        <div className="relative flex bg-green-400 h-4">
          {alerts.map((alert, index) => {
            const pendingAt = new Date(alert.pendingAt * 1000);
            const resolvedAt = new Date(alert.resolvedAt ? (alert.resolvedAt * 1000) : Date.now());
            const startPosition = Math.max(0, (pendingAt - startDate) / (60 * 60 * 1000));
            const endPosition = Math.min(currentHour.hour, (resolvedAt - startDate) / (60 * 60 * 1000));

            const endDateUnix = (new Date(endDate).getTime() / 1000).toFixed(0)
            const total = (alert.resolvedAt ?? endDateUnix) - alert.pendingAt
            const pendingInterval = alert.firedAt - alert.pendingAt
            const firingInterval = (alert.resolvedAt ?? endDateUnix) - alert.firedAt

            const alertStyle = {
              left: `${(startPosition / currentHour.hour) * 100}%`,
              width: `${((endPosition - startPosition) / currentHour.hour) * 100}%`,
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
    </div>
  );
};

export default Timeline;
