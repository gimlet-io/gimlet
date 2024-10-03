import { useState } from "react";
import { toast } from 'react-toastify';
import { Error } from '../../popUpWindow';

export function RegistryWidget(props) {
  const { formData, onChange, gimletClient, store, env } = props;
  const [url, setURL] = useState(formData ? formData.url : "")
  const [login, setLogin] = useState(formData ? formData.login : "")
  const [token, setToken] = useState("")
  const encryptedDockerconfigjson = formData ? formData.encryptedDockerconfigjson : ""

  const encrypt = () => {
    const configjson = {
      "auths": {
        [url]: {
          "auth": btoa(`${login}:${token}`)
        }
      }
    }

    return () => {
      gimletClient.seal(env, JSON.stringify(configjson))
        .then(data => {
          onChange({ "login": login, "url": url, "encryptedDockerconfigjson": data })
        }, () => {
          toast(<Error header="Failed to encrypt" message={`${err.statusText}, is the environment connected?`} />, {
            className: "bg-red-50 shadow-lg p-2",
            bodyClassName: "p-2",
            progressClassName: "!bg-red-200",
            autoClose: 3000
          });
        });
    };
  }

  const reset = () => {
    return () => {
      onChange({})
    };
  }

  const disabled = url === "" || login === "" || token === "";

  return (
    <fieldset className="pb-4">
      <p className="control-label pb-1">
        {props.schema.title}
      </p>
      {encrypted(encryptedDockerconfigjson) ?
        <>
          <Registry url={url} login={login} resetFunc={reset()} />
        </>
        :
        <>
          <div>
            <p className="control-label">URL</p>
            <input className="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
              value={url} onChange={e => setURL(e.target.value)} />
          </div>
          <div>
            <p className="control-label">Login</p>
            <input className="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
              value={login} onChange={e => setLogin(e.target.value)} />
          </div>
          <div>
            <p className="control-label">Token</p>
            <input className="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
              value={token} onChange={e => setToken(e.target.value)} />
          </div>
        </>
      }
      {!encrypted(encryptedDockerconfigjson) &&
        <div className='flex justify-end pt-2'>
          <button
            onClick={encrypt()}
            disabled={disabled}
            className={`${disabled ? 'primaryButtonDisabled': 'primaryButton'} px-4`}
          >Encrypt</button>
        </div>
      }
    </fieldset>
  )
}

function encrypted(configJson) {
  return configJson && configJson !== ""
}

function Registry(props) {
  const { url, login, resetFunc } = props;
  return (
    <div className="card flex items-center space-x-2 rounded-lg ring-1 ring-inset ring-neutral-200 p-2">
      <svg className="size-6" strokeLinejoin="round" viewBox="0 0 16 16"><path fillRule="evenodd" clipRule="evenodd" d="M8 0.154663L8.34601 0.334591L14.596 3.58459L15 3.79466V4.25V11.75V12.2053L14.596 12.4154L8.34601 15.6654L8 15.8453L7.65399 15.6654L1.40399 12.4154L1 12.2053V11.75V4.25V3.79466L1.40399 3.58459L7.65399 0.334591L8 0.154663ZM2.5 11.2947V5.44058L7.25 7.81559V13.7647L2.5 11.2947ZM8.75 13.7647L13.5 11.2947V5.44056L8.75 7.81556V13.7647ZM8 1.84534L12.5766 4.22519L7.99998 6.51352L3.42335 4.2252L8 1.84534Z" fill="currentColor"></path>
      </svg>
      <p target="_blank" rel="noreferrer" className="flex items-center flex-grow text-sm">
        <span>{login}</span> @ <span>{url}</span>
        {/* <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" /> */}
      </p>
      <button
        onClick={resetFunc}
        className="primaryButton !py-0 px-2">
        Reset
      </button>
    </div>
  )
}
