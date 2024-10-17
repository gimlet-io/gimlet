import { useState } from "react";
import { toast } from 'react-toastify';
import { Error } from '../../popUpWindow';

export function EncryptedWidget(props) {
  const { formData, onChange, gimletClient, store, env, singleLine } = props;
  const [plainTextValue, setPlainTextValue] = useState("")
  const encryptedValue = formData

  const encrypt = () => {
    return () => {
      gimletClient.seal(env, plainTextValue)
        .then(data => {
          onChange(data)
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

  const reset = () => {
    return () => {
      onChange(undefined)
    };
  }

  const disabled = plainTextValue === "";

  return (
    <div className="field">
      <p className="control-label pb-1">
        {props.schema.title}
      </p>
      {encryptedValue
        ? <Encrypted resetFunc={reset()} />
        : <>
            {singleLine &&
            <input className="form-control" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository"
            value={plainTextValue} onChange={e => setPlainTextValue(e.target.value)} />
            }
            {!singleLine &&
            <textarea rows="8" className="form-control" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository"
              value={plainTextValue} onChange={e => setPlainTextValue(e.target.value)} />
            }
            <div className='flex justify-end pt-2'>
              <button
                onClick={encrypt()}
                disabled={disabled}
                className={`${disabled ? 'primaryButtonDisabled': 'primaryButton'} px-4`}
              >Encrypt</button>
            </div>
          </>
      }
    </div>
  )
}

function Encrypted(props) {
  const { resetFunc } = props;
  return (
    <div className="card flex items-center rounded-lg ring-1 ring-inset ring-neutral-200 p-2">
      <p target="_blank" rel="noreferrer" className="flex items-center flex-grow text-sm">
        <span>***encrypted***</span>
      </p>
      <button
        onClick={resetFunc}
        className="primaryButton !py-0 px-2">
        Reset
      </button>
    </div>
  )
}
