import { Component } from "react";

class ImageWidget extends Component {
  constructor(props) {
    super(props);

    const { repository, tag, dockerfile } = props.formData
    const strategy = this.extractStrategyFromValue(repository, tag, dockerfile)

    this.state = {
      strategy: strategy,
      ...props.formData,
    };
  }

  extractStrategyFromValue(repository, tag, dockerfile) {
    const hasVariable = repository.includes("{{") || tag.includes("{{")
    const pointsToBuiltInRegistry = repository.includes("127.0.0.1:32447")
    const hasDockerfile = dockerfile && dockerfile !== ""

    if (!hasVariable) {
      return "static"
    } else {
      if (!pointsToBuiltInRegistry) {
        return "dynamic"
      } else {
        if (hasDockerfile) {
          return "dockerfile"
        } else {
          return "buildpacks"
        }
      }
    }
  }

  defaults(strategy) {
    let repository = ""
    let tag = ""
    let dockerfile = ""

    switch (strategy) {
      case 'dynamic':
        repository = "ghcr.io/your-company/your-repo"
        tag = "{{ .SHA }}"
        dockerfile = ""
        break;
      case 'dockerfile':
        repository = "127.0.0.1:32447/{{ .APP }}"
        tag = "{{ .SHA }}"
        dockerfile = "Dockerfile"
        break;
      case 'buildpacks':
        repository = "127.0.0.1:32447/{{ .APP }}"
        tag = "{{ .SHA }}"
        dockerfile = ""
        break;
      default:
        repository = "nginx"
        tag = "1.19.3"
        dockerfile = ""
    }

    return {
      repository: repository,
      tag: tag,
      dockerfile: dockerfile,
    }
  }

  onChange(name) {
    return (event) => {
      this.setState(
        {
          [name]: event.target.value,
        },
        () => this.props.onChange({"repository": this.state.repository, "tag": this.state.tag})
      );
    };
  }

  render() {
    const { strategy, repository, tag, dockerfile } = this.state;
    return (
      <>
      <div className="form-group field field-object">
        <fieldset id="root">
          <legend id="root__title">Image</legend>
          <p id="root__description" className="field-description">The image to deploy</p>
          <div className="mt-4 grid grid-cols-1 gap-y-6 sm:grid-cols-4 sm:gap-x-4 px-2">
            <div 
              className={`relative flex cursor-pointer rounded-lg border bg-white p-4 shadow-sm focus:outline-none ${strategy === "static" ? "border-indigo-600" : ""}`}
              onClick={(e) => this.setState({strategy: "static", ...this.defaults("static")}, () => this.props.onChange({"repository": this.state.repository, "tag": this.state.tag}))}
              >
              <span className="flex flex-1">
                <span className="flex flex-col">
                  <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">Static image tag</span>
                  <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-gray-500 select-none">If you want to deploy a specific version of an existing image</span>
                </span>
              </span>
              <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-indigo-600 ${strategy === "static" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
              </svg>
            </div>

            <div
              className={`relative flex cursor-pointer rounded-lg border bg-white p-4 shadow-sm focus:outline-none ${strategy === "dynamic" ? "border-indigo-600" : ""}`}
              onClick={(e) => this.setState({strategy: "dynamic", ...this.defaults("dynamic")}, () => this.props.onChange({"repository": this.state.repository, "tag": this.state.tag}))}
              >
              <span className="flex flex-1">
                <span className="flex flex-col">
                  <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">Dynamic image tag</span>
                  <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-gray-500 select-none">If CI builds an image and tags it with the git hash, tag or other dynamic identifier</span>
                </span>
              </span>
              <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-indigo-600 ${strategy === "dynamic" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
              </svg>
            </div>

            <div
              className={`relative pr-8 flex cursor-pointer rounded-lg border bg-white p-4 shadow-sm focus:outline-none ${strategy === "buildpacks" ? "border-indigo-600" : ""}`}
              onClick={(e) => this.setState({strategy: "buildpacks", ...this.defaults("buildpacks")}, () => this.props.onChange({"repository": this.state.repository, "tag": this.state.tag}))}
              >
              <span className="flex flex-1">
                <span className="flex flex-col">
                  <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">Automatic image building</span>
                  <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-gray-500 select-none">If you want Gimlet to build an image from source code</span>
                </span>
              </span>
              <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-indigo-600 ${strategy === "buildpacks" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
              </svg>
            </div>

            <div
              className={`relative pr-8 flex cursor-pointer rounded-lg border bg-white p-4 shadow-sm focus:outline-none ${strategy === "dockerfile" ? "border-indigo-600" : ""}`}
              onClick={(e) => this.setState({strategy: "dockerfile", ...this.defaults("dockerfile")}, () => this.props.onChange({"repository": this.state.repository, "tag": this.state.tag, "dockerfile": this.state.dockerfile}))}
              >
              <span className="flex flex-1">
                <span className="flex flex-col">
                  <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">Using a Dockerfile</span>
                  <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-gray-500 select-none">If there is a Dockerfile in your source code and want Gimlet to build it</span>
                </span>
              </span>
              <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-indigo-600 ${strategy === "dockerfile" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
              </svg>
            </div>

          </div>
          <div className="form-group field field-string">
            <label className="control-label" htmlFor="root_repository">Repository<span className="required">*</span></label>
            <input className="form-control" id="root_repository" label="Repository" required="" placeholder="" type="text" list="examples_root_repository" value={repository} onChange={this.onChange('repository')} />
          </div>
          <div className="form-group field field-string">
            <label className="control-label" htmlFor="root_tag">Tag<span className="required">*</span></label>
            <input className="form-control" id="root_tag" label="Tag" required="" placeholder="" type="text" list="examples_root_tag" value={tag}  onChange={this.onChange('tag')}/>
          </div>
          { strategy === "dockerfile" &&
          <div className="form-group field field-string">
            <label className="control-label" htmlFor="root_tag">Dockerfile<span className="required">*</span></label>
            <input className="form-control" id="root_tag" label="Dockerfile" required="" placeholder="" type="text" list="examples_root_tag" value={dockerfile}  onChange={this.onChange('dockerfile')}/>
          </div>
          }
        </fieldset>
      </div>
      </>
    );
  }
}

export default ImageWidget;
