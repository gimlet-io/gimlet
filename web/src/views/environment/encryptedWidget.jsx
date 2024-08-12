import { useState } from "react";
import {
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWERROR
} from "../../redux/redux";

export function EncryptedWidget(props) {
  const { formData, onChange, gimletClient, store, env } = props;
  const [plainTextValue, setPlainTextValue] = useState("")
  const [encryptedValue, setEncryptedValue] = useState(formData)

  const encrypt = () => {
    return () => {
      gimletClient.seal(env, plainTextValue)
        .then(data => {
          onChange(data)
        }, () => {
          store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
              header: "Error",
              message: "Failed to encrypt."
            }
          });
          resetPopupWindowAfterThreeSeconds(store)
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
            <textarea rows="8" className="form-control" id="root_repository" required="" placeholder="" type="text" list="examples_root_repository"
              value={plainTextValue} onChange={e => setPlainTextValue(e.target.value)} />
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

function resetPopupWindowAfterThreeSeconds(store) {
  setTimeout(() => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWRESET
    });
  }, 3000);
};

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
