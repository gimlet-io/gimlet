import {Component} from "react";
import {ACTION_TYPE_STREAMING} from "./redux/redux"

let URL = '';
if (typeof window !== 'undefined') {
  let protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
  URL = protocol + '://' + window.location.hostname;
}

export default class StreamingBackend extends Component {
  componentDidMount() {
    if (this.props.location.pathname === '/login') {
      return;
    }

    this.ws = new WebSocket(URL + '/ws/');
    this.ws.onopen = this.onOpen;
    this.ws.onmessage = this.onMessage;
    this.ws.onclose = this.onClose;

    this.onClose = this.onClose.bind(this);
  }

  render() {
    return null;
  }

  onOpen = () => {
    console.log('connected');
  };

  onClose = (evt) => {
    console.log('disconnected: ' + evt.code + ': ' + evt.reason);
    const ws = new WebSocket(URL + '/ws/');
    ws.onopen = this.onOpen;
    ws.onmessage = this.onMessage;
    ws.onclose = this.onClose;
    this.setState({
      ws
    });
  }

  onMessage = (evt) => {
    evt.data.split('\n').forEach((line) => {
      const message = JSON.parse(line);
      this.props.store.dispatch({type: ACTION_TYPE_STREAMING, payload: message});
    });
  }
}
