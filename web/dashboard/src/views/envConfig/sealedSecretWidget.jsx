import { Component } from "react";

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
    return () => {
      this.props.onChange("toSeal: " + this.state.value)
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
          <button className="m-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-12"
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
