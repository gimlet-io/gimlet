import { useEffect, useState } from 'react';
import posthog from 'posthog-js';

const Posthog = ({ store }) => {
  const reduxState = store.getState();
  const [login, setLogin] = useState(reduxState.user.login)
  const [posthogFeatureFlag, setPosthogFeatureFlag] = useState(reduxState.settings.posthogFeatureFlag)
  const [posthogApiKey, setPosthogApiKey] = useState(reduxState.settings.posthogApiKey)
  const [posthogIdentifyUser, setPosthogIdentifyUser] = useState(reduxState.settings.posthogIdentifyUser)

  store.subscribe(() => {
    const reduxState = store.getState();
    setLogin(reduxState.user.login)
    setPosthogFeatureFlag(reduxState.settings.posthogFeatureFlag)
    setPosthogApiKey(reduxState.settings.posthogApiKey)
    setPosthogIdentifyUser(reduxState.settings.posthogIdentifyUser)
  });

  useEffect(() => {
    if (posthogFeatureFlag === undefined) {
      return;
    }
    if ( posthogIdentifyUser === undefined) {
      return;
    }
    if(!posthogApiKey) {
      return;
    }
    if (!posthogFeatureFlag) {
      return;
    }

    // TODO
    // hide sensitives from autocapture events
    // (clicked p with text "owner/repo"
    // url/screen http://localhost:9000/repo/owner/repo/envs/staging-0/config/onechart-4e96/new
    // Pageleave http://localhost:9000/repo/owner/repo/envs/staging-0/config/onechart-570a/new)
    if (!posthog.__loaded) {
      posthog.init(posthogApiKey, {
        api_host: 'https://eu.posthog.com',
        disable_session_recording: false, // TODO disable on open source usage
        session_recording: {
          maskTextSelector: "*" // Masks all text elements
        },
        autocapture: {
          url_allowlist: [/\/repositories/, /\/environments/, /\/settings/, /\/profile/], // strings or RegExps
          element_allowlist: ['a', 'button', 'form', 'input', 'select', 'textarea', 'label'], // DOM elements from this list ['a', 'button', 'form', 'input', 'select', 'textarea', 'label']
        },
      });
      console.log(posthog.config)
    }

    if (posthog.__loaded && posthogIdentifyUser && login) {
      posthog.identify(login)
    }
  }, [posthogFeatureFlag, posthogApiKey, posthogIdentifyUser, login]);

  return null;
}

export default Posthog;
