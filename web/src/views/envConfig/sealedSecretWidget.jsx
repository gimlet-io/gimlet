import { Component } from "react";
import { toast } from 'react-toastify';
import { Error } from '../../popUpWindow';

class SealedSecretWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      value: props.formData === "New Value" ? "" : props.formData,
      sealed: props.formData === "New Value" ? false : true,
    };
  }

  componentDidUpdate(prevProps) {
    if (prevProps.formData !== this.props.formData) {
      this.setState({
        value: this.props.formData,
        sealed: true,
      });
    }
  }

  onChange() {
    return (event) => {
      this.setState({
        value: event.target.value,
      });
    };
  }

  seal() {
    const { gimletClient, env } = this.props;
    return () => {
      gimletClient.seal(env, this.state.value)
        .then(data => {
          this.props.onChange(data)
        }, (err) => {
          toast(<Error header="Failed to encrypt" message={`${err.statusText}, is the environment connected?`} />, {
            className: "bg-red-50 shadow-lg p-2",
            bodyClassName: "p-2",
            progressClassName: "!bg-red-200",
            autoClose: 3000
          });
        });
    };
  }

  render() {
    return (
      <>
        <div className="form-group field field-string flex">
          { this.state.sealed &&
          <>
          <textarea disabled rows="5" className="form-control bg-neutral-400" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository" value={this.state.value} onChange={this.onChange()} />
          </>
          }
          { !this.state.sealed &&
          <>
          <textarea rows="5" className="form-control" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository" value={this.state.value} onChange={this.onChange()} />
          <button disabled={this.state.value === ""} className={(this.state.value === "" ? "primaryButtonDisabled" : "primaryButton") + " m-2 px-4 h-12"}
            onClick={this.seal()}
          >
            Encrypt
          </button>
          </>
          }
        </div>
      </>
    );
  }
}

export default SealedSecretWidget;
