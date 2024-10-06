import { useState, useEffect } from 'react';

export default function ImageWidget(props) {
  const registries = [...props.uiSchema["ui:options"]?.registries]
  const [image, setImage] = useState(props.formData)

  const setRepository = (repository) => {
    setImage({
      ...image,
      "repository": repository,
    })
  }

  const setTag = (tag) => {
    setImage({
      ...image,
      "tag": tag,
    })
  }

  const setDockerfile = (dockerfile) => {
    setImage({
      ...image,
      "dockerfile": dockerfile,
    })
  }

  const setContext = (context) => {
    setImage({
      ...image,
      "context": context,
    })
  }

  const setStrategy = (strategy) => {
    let registry = {}
    registry = registries.find(r => r.variable === "customRegistry")
    if (!registry) {
      registry = registries.find(r => r.variable === "containerizedRegistry")
    }

    switch (strategy) {
      case 'dynamic':
        setImage({
          ...image,
          "strategy": strategy,
          "registry": "",
          "repository": "your-company/your-repo",
          "tag": "{{ .SHA }}",
          "dockerfile": ""
        })
        break;
      case 'dockerfile':
        setImage({
          ...image,
          "strategy": strategy,
          "registry": registry.variable,
          "repository": registry.url+"/{{ .APP }}",
          "tag": "{{ .SHA }}",
          "context": ".",
          "dockerfile": "Dockerfile"
        })
        break;
      case 'buildpacks':
        setImage({
          ...image,
          "strategy": strategy,
          "registry": registry.variable,
          "repository": registry.url+"/{{ .APP }}",
          "tag": "{{ .SHA }}",
          "dockerfile": ""
        })
        break;
      default:
        setImage({
          ...image,
          "strategy": strategy,
          "registry": "public",
          "repository": "nginx",
          "tag": "1.27",
          "dockerfile": ""
        })
    }
  }

  useEffect(() => {
    props.onChange(image)
  }, [image]);

  const setRegistry = (registry) => {
    if (!registries) {
      return
    }

    const selectedRegistry = registries.find(r => r.variable === registry)
    if (selectedRegistry.variable === "public") {
      setImage({
        ...image,
        "registry": registry,
      })
    } else {
      const login = selectedRegistry.login ?? "your-company"
      let repository = ""
      switch(selectedRegistry.variable) {
        case "containerizedRegistry":
          repository = `${selectedRegistry.url}/{{ .APP }}`
          break
        case "dockerhubRegistry":
          repository = `${login}/{{ .APP }}`
          break
        default:
          repository = `${selectedRegistry.url}/${login}/{{ .APP }}`
      }
      setImage({
        ...image,
        "registry": registry,
        "repository": repository
      })
    }
  }

  return (
    <div className="form-group field field-object">
      <fieldset id="root">
        <legend id="root__title">Container Image</legend>
        <p id="root__description" className="field-description">Choose a container image building strategy and specify the image location.</p>
        <div className="my-8 grid grid-cols-1 gap-y-6 sm:grid-cols-4 sm:gap-x-4">
          <div 
            className={`relative flex cursor-pointer rounded-lg border dark:border-2 bg-white dark:bg-neutral-100 p-4 shadow-sm focus:outline-none ${image.strategy === "static" ? "border-blue-600" : ""}`}
            onClick={(e) => setStrategy("static")}
            >
            <span className="flex flex-1">
              <span className="flex flex-col">
                <span id="project-type-0-label" className="block text-sm font-medium text-neutral-900 select-none">Static image tag</span>
                <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-neutral-500 select-none">If you want to deploy a specific version of an existing image</span>
              </span>
            </span>
            <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-blue-600 ${image.strategy === "static" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
            </svg>
          </div>

          <div
            className={`relative flex cursor-pointer rounded-lg border dark:border-2 bg-white dark:bg-neutral-100 p-4 shadow-sm focus:outline-none ${image.strategy === "dynamic" ? "border-blue-600" : ""}`}
            onClick={(e) => setStrategy("dynamic")}
            >
            <span className="flex flex-1">
              <span className="flex flex-col">
                <span id="project-type-0-label" className="block text-sm font-medium text-neutral-900 select-none">Build with CI</span>
                <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-neutral-500 select-none">If CI builds an image and tags it with the git hash, tag or other dynamic identifier</span>
              </span>
            </span>
            <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-blue-600 ${image.strategy === "dynamic" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
            </svg>
          </div>

          <div
            className={`relative pr-8 flex cursor-pointer rounded-lg border dark:border-2 bg-white dark:bg-neutral-100 p-4 shadow-sm focus:outline-none ${image.strategy === "buildpacks" ? "border-blue-600" : ""}`}
            onClick={(e) => setStrategy("buildpacks")}
            >
            <span className="flex flex-1">
              <span className="flex flex-col">
                <span id="project-type-0-label" className="block text-sm font-medium text-neutral-900 select-none">Build with Buildpacks</span>
                <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-neutral-500 select-none">If you want Gimlet to build an image from source code</span>
              </span>
            </span>
            <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-blue-600 ${image.strategy === "buildpacks" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
            </svg>
          </div>

          <div
            className={`relative pr-8 flex cursor-pointer rounded-lg border dark:border-2 bg-white dark:bg-neutral-100 p-4 shadow-sm focus:outline-none ${image.strategy === "dockerfile" ? "border-blue-600" : ""}`}
            onClick={(e) => setStrategy("dockerfile")}
            >
            <span className="flex flex-1">
              <span className="flex flex-col">
                <span id="project-type-0-label" className="block text-sm font-medium text-neutral-900 select-none">Using a Dockerfile</span>
                <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm text-neutral-500 select-none">If there is a Dockerfile in your source code and want Gimlet to build it</span>
              </span>
            </span>
            <svg className={`absolute top-0 right-0 m-4 h-5 w-5 text-blue-600 ${image.strategy === "dockerfile" ? "" : "hidden"}`} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
            </svg>
          </div>

        </div>
        <div className="form-group field">
          <label className="control-label" htmlFor="root_tag">Registry<span className="required"></span></label>
          <select id="registry" className="form-control" value={image.registry} 
          onChange={e => setRegistry(e.target.value)}
          >
            {registries.map(r => <option key={r.variable} value={r.variable}>{r.name}</option>)}
          </select>
        </div>
        <div className="form-group field field-string">
          <label className="control-label" htmlFor="root_repository">Repository<span className="required"></span></label>
          <input className="form-control" id="root_repository" label="Repository" required="" placeholder="" type="text" list="examples_root_repository" value={image.repository} onChange={e=>setRepository(e.target.value)} />
        </div>
        <div className="form-group field field-string">
          <label className="control-label" htmlFor="root_tag">Tag<span className="required"></span></label>
          <input className="form-control max-w-64" id="root_tag" label="Tag" required="" placeholder="" type="text" list="examples_root_tag" value={image.tag}  onChange={e=>setTag(e.target.value)}/>
        </div>
        { image.strategy === "dockerfile" &&
        <>
          <div className="form-group field field-string">
            <label className="control-label" htmlFor="root_tag">Context<span className="required"></span></label>
            <input className="form-control max-w-64" id="root_tag" label="Context" required="" placeholder="" type="text" list="examples_root_tag" value={image.context}  onChange={e=>setContext(e.target.value)}/>
            <p className="help-block">Case-sensitive relative path from the git repository root (signaled as '.') to the project root. Change it for monorepos, like 'backend/'.</p>
          </div>
          <div className="form-group field field-string">
            <label className="control-label" htmlFor="root_tag">Dockerfile<span className="required"></span></label>
            <input className="form-control max-w-64" id="root_tag" label="Dockerfile" required="" placeholder="" type="text" list="examples_root_tag" value={image.dockerfile}  onChange={e=>setDockerfile(e.target.value)}/>
            <p className="help-block">Case-sensitive relative path from the project root to the Dockerfile, like 'backend/Dockerfile'</p>
          </div>
        </>
        }
      </fieldset>
    </div>
  );
}
