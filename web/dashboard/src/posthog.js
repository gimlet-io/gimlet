import { useEffect, useState } from 'react';
import posthog from 'posthog-js';

// TODO posthog.reset() on logout

const Posthog = ({ store }) => {
  const reduxState = store.getState();
  const [login, setLogin] = useState(reduxState.user.login)
  const [posthogApiKey, setPosthogApiKey] = useState(reduxState.settings.posthogApiKey)
  const [posthogInitiated, setPosthogInitiated] = useState(false)

  store.subscribe(() => {
    const reduxState = store.getState();
    setLogin(reduxState.user.login)
    setPosthogApiKey(reduxState.settings.posthogApiKey)
  });

  useEffect(() => {
    if (posthogApiKey && login && !posthogInitiated) {
      posthog.init(posthogApiKey, { api_host: 'https://eu.posthog.com' })
      posthog.identify(login)
      setPosthogInitiated(true)
    }
  }, [posthogApiKey, login, posthogInitiated]);

  return null;
}

export const posthogCapture = (text) => {
  posthog.capture(text)
}

export default Posthog;
