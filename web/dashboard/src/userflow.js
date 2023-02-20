import { useEffect, useState } from 'react';
import userflow from 'userflow.js'

const Userflow = ({ store }) => {
  const reduxState = store.getState();
  const [login, setLogin] = useState(reduxState.user?.login)
  const [userflowToken, setUserflowToken] = useState(reduxState.settings.userflowToken)
  const [userflowInitiated, setUserflowInitiated] = useState(false)

  store.subscribe(() => {
    const reduxState = store.getState();
    setLogin(reduxState.user?.login)
    setUserflowToken(reduxState.settings.userflowToken)
  });

  useEffect(() => {
    if (userflowToken && login && !userflowInitiated) {
      userflow.init(userflowToken)
      userflow.identify(window.location.hostname + "/" + login)
      setUserflowInitiated(true)
    }
  }, [userflowToken, login, userflowInitiated]);

  return null;
}

export default Userflow;
