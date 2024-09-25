import {useEffect} from "react";
import {ACTION_TYPE_STREAMING} from "./redux/redux"
import { useLocation } from 'react-router-dom'

let URL = '';
if (typeof window !== 'undefined') {
  let protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
  URL = protocol + '://' + window.location.hostname;

  let port = window.location.port
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    port = 9000
  }
  if (port && port !== '') {
    URL = URL + ':' + port
  }
}

export default function StreamingBackend(props) {
  const location = useLocation()
  
  const onOpen = () => {
    console.log('connected');
  };

  const onClose = (evt) => {
    console.log('disconnected: ' + evt.code + ': ' + evt.reason);

    setTimeout(() => {
      console.log("Connecting to " + URL + '/ws/')
      const ws = new WebSocket(URL + '/ws/');
      ws.onopen = onOpen;
      ws.onmessage = onMessage;
      ws.onclose = onClose;
    }, 100);
  }

  const onMessage = (evt) => {
    evt.data.split('\n').forEach((line) => {
      const message = JSON.parse(line);
      // console.log(line)
      props.store.dispatch({type: ACTION_TYPE_STREAMING, payload: message});
    });
  }

  useEffect(() => {
    if (location.pathname === '/login') {
      return;
    }

    console.log("Connecting to " + URL + '/ws/')
    const ws = new WebSocket(URL + '/ws/');
    ws.onopen = onOpen;
    ws.onmessage = onMessage;
    ws.onclose = onClose;
  })

  return null;
}
