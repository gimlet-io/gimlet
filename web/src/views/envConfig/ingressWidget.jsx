import { useState, useEffect } from 'react';

export default function IngressWidget(props) {
  const preferredDomain = props.uiSchema["ui:options"]?.preferredDomain
  const [host, setHost] = useState(props.formData)

  useEffect(() => {
    props.onChange(host)
  }, [host]);

  return (
  <>
    <label className="control-label" htmlFor="root_tag">Host<span className="required"></span></label>
    {(!preferredDomain || !host.endsWith(preferredDomain)) && 
      <input className="form-control" id="root_tag" label="Host" required="" placeholder="" type="text" list="examples_root_tag" value={host} onChange={e=>setHost(e.target.value)}/>
    }
    {host.endsWith(preferredDomain) && 
    <div className="mt-2 inline-flex rounded-md shadow-sm">
      <input
        type="text"
        name="company-website"
        id="company-website"
        className="block w-96 min-w-0 flex-1 rounded-none rounded-l-md border-0 py-1.5 text-neutral-700 ring-1 ring-inset ring-neutral-200 dark:ring-neutral-700 placeholder:text-neutral-300 dark:placeholder:text-neutral-600 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
        placeholder="myapp"
        value={host.slice(0, -preferredDomain.length)}
        onChange={e=>setHost(e.target.value+preferredDomain)}
      />
      <span className="inline-flex items-center rounded-r-md border border-l-0 border-neutral-200 dark:border-neutral-700 pl-1 pr-3 bg-white dark:bg-neutral-900 text-neutral-700 dark:text-neutral-300 sm:text-sm">
        {preferredDomain}
      </span>
    </div>
    }
  </>
  );
}
