import Toggle from "../../components/toggle/toggle";

const KustomizationPerApp = ({ kustomizationPerApp, setKustomizationPerApp }) => {
  return (<div className="text-neutral-700">
    <div className="flex mt-4">
      <div className="font-medium self-center">Kustomization per app</div>
      <div className="max-w-lg flex rounded-md ml-4">
        <Toggle
          checked={kustomizationPerApp}
          onChange={setKustomizationPerApp}
        />
      </div>
    </div>
    <div className="text-sm text-neutral-500 leading-loose">Enable it for each application to have a separate deployment pipeline. This is a more robust setup, but generates potentially hundreds of kustomization files. One per application.</div>
  </div>)
};

export default KustomizationPerApp;
