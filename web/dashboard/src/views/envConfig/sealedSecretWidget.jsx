import { Component } from "react";
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET
} from "../../redux/redux";

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

  resetPopupWindowAfterThreeSeconds() {
    const { store } = this.props;
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  seal() {
    const { gimletClient, store, env } = this.props;
    return () => {
      gimletClient.seal(env, this.state.value)
        .then(data => {
          this.props.onChange(data)
        }, () => {
          store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
              header: "Error",
              message: "Failed to seal."
            }
          });
          this.resetPopupWindowAfterThreeSeconds()
        });
    };
  }

  render() {
    return (
      <>
        <div className="form-group field field-string flex">
          { this.state.sealed &&
          <>
          <textarea disabled rows="5" className="form-control bg-gray-400" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository" value={this.state.value} onChange={this.onChange()} />
          </>
          }
          { !this.state.sealed &&
          <>
          <textarea rows="5" className="form-control" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository" value={this.state.value} onChange={this.onChange()} />
          <button disabled={this.state.value === ""} className={(this.state.value === "" ? "bg-gray-500" : "bg-blue-500 hover:bg-blue-700") + " m-2 text-white font-bold py-2 px-4 rounded h-12"}
            onClick={this.seal()}
          >
            Seal
          </button>
          </>
          }
        </div>
      </>
    );
  }
}

export default SealedSecretWidget;
