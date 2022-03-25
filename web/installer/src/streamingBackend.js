import React, { Component } from 'react'

export default class StreamingBackend extends Component {
    constructor(props) {
        super(props)

        this.state = {
            URL: this.props.client.URL().replace('https', 'wss').replace('http', 'ws') + '/ws',
            ws: null
        }

        this.connect = this.connect.bind(this)
        this.check = this.check.bind(this)
    }

    componentDidMount() {
        this.connect()
    }

    connect() {
        const ws = new WebSocket(this.state.URL)
        ws.onopen = () => {
            console.log('connected websocket main component')
            this.setState({ ws: ws })
        }
        ws.onclose = e => {
            console.log(`Reconnecting`, e.reason);
            this.connect();
        }
        ws.onerror = err => {
            console.error('Socket encountered error: ',
                err.message,
                'Closing socket'
            )
            ws.close()
        }
        console.log(ws)
    };

    check() {
        const { ws } = this.state
        if (!ws || ws.readyState === WebSocket.CLOSED) this.connect()
    };

    render() {
        return null
    }
}
