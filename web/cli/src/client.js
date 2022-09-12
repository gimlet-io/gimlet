export default class GimletCLIClient {
  constructor () {
    if (typeof window !== 'undefined') {
      let port = window.location.port
      this.url = window.location.protocol + '//' + window.location.hostname
      if (port && port !== '') {
        this.url = this.url + ':' + port
      }
    }
  }

  URL () {
    return this.url
  }

  saveValues(values) {
    this.post('/saveValues', JSON.stringify(values));
  }

  post (path, body) {
    fetch(this.url + path, {
      method: 'post',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json'
      },
      body
    })
      .then(response => {
        if (!response.ok && window !== undefined) {
          return Promise.reject({ status: response.status, statusText: response.statusText, path })
        }
        return response.json()
      })
      .catch((error) => {
        this.onError(error)
        throw error
      })
  }
}


