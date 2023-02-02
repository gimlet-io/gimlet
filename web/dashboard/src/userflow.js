import { useEffect, useState } from 'react';
import userflow from 'userflow.js'

const Userflow = ({ store }) => {
  let reduxState = store.getState();
  const [user, setUser] = useState(reduxState.user)
  const [settings, setSettings] = useState(reduxState.settings)
  const [userflowInitiated, setUserflowInitiated] = useState(false)

  store.subscribe(() => {
    setUser(reduxState.user)
    setSettings(reduxState.settings)
  });

  useEffect(() => {
    if (settings && settings.userflowToken
        && user
        && !userflowInitiated) {
      console.log("userflow init")
      console.log(settings.userflowToken)
      userflow.init(settings.userflowToken)
      userflow.identify(user.login)
      setUserflowInitiated(true)
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  return null;
}

export default Userflow;
