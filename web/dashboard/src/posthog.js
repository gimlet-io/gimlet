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
    if (!posthogFeatureFlag) {
      return;
    }
    if (posthogIdentifyUser === undefined) {
      return;
    }

    if (!posthog.__loaded && posthogApiKey) {
      posthog.init(posthogApiKey, {
        api_host: 'https://eu.posthog.com',
        disable_session_recording: !posthogIdentifyUser, // TODO disable on open source usage
        enable_recording_console_log: false,
        session_recording: {
          maskTextSelector: "*" // Masks all text elements
        },
        capture_pageview: false,
        capture_pageleave: false,
        autocapture: {
          url_allowlist: [/\/repositories/, /\/environments/, /\/settings/, /\/profile/],
          element_allowlist: ['a', 'button', 'form', 'input', 'select', 'textarea', 'label'],
        },
      });
    }

    if (posthogIdentifyUser && posthog.__loaded && login) {
      posthog.identify(login)
    }
  }, [posthogFeatureFlag, posthogApiKey, posthogIdentifyUser, login]);

  return null;
}

export default Posthog;
