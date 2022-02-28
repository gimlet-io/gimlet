import React from 'react';

const Button = (props) => {
  return (
    <button className={`btn btn--${props.kind} CTA`}
      data-id={props.id}
      type={props.type}
      name={props.name}
      value={props.value}
      disabled={props.disabled}
      onClick={props.handleClick}>
      <h4>{props.label} hello9</h4>
    </button>
  )
}

export default Button;
