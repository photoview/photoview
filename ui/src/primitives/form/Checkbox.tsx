import React from 'react'
import styled from 'styled-components'

const CheckboxLabelWrapper = styled.label`
  & input[type='checkbox'] {
    appearance: none;
    outline: none;
  }

  & input[type='checkbox'] + span {
    cursor: pointer;
  }

  & input[type='checkbox'] + span::before {
    content: '';
    background-repeat: no-repeat;
    background-position: center;
    width: 12px;
    height: 12px;
    background-color: #f9fafb;
    border: 1px solid #aaa;
    display: inline-block;
    border-radius: 3px;
    cursor: pointer;
    margin-right: 8px;
    margin-bottom: -1px;
  }

  & input[type='checkbox']:checked + span::before {
    background-color: #0072ff;
    border-color: #109dff;
    background-image: url("data:image/svg+xml,%3C%3Fxml version='1.0' encoding='UTF-8'%3F%3E%3Csvg width='7px' height='4px' viewBox='0 0 7 4' version='1.1' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'%3E%3Ctitle%3EPath 5%3C/title%3E%3Cg stroke='none' stroke-width='1.5' fill='none' fill-rule='evenodd'%3E%3Cpolyline stroke='white' points='1 1.93487039 2.77174339 3.70661378 6.47835718 2.27373675e-13'%3E%3C/polyline%3E%3C/g%3E%3C/svg%3E");
  }

  & input[type='checkbox']:focus + span::before {
    outline: 2px solid #86c6ff;
    border-color: #109dff;
  }

  & input[type='checkbox']:disabled + span::before {
    border-color: #ccc;
    background-color: #fefefe;
    cursor: initial;
  }

  & input[type='checkbox']:disabled:checked + span::before {
    background-color: #eee;
    background-image: url("data:image/svg+xml,%3C%3Fxml version='1.0' encoding='UTF-8'%3F%3E%3Csvg width='7px' height='4px' viewBox='0 0 7 4' version='1.1' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'%3E%3Ctitle%3EPath 5%3C/title%3E%3Cg stroke='none' stroke-width='1.5' fill='none' fill-rule='evenodd'%3E%3Cpolyline stroke='%23444444' points='1 1.93487039 2.77174339 3.70661378 6.47835718 2.27373675e-13'%3E%3C/polyline%3E%3C/g%3E%3C/svg%3E");
  }

  & input[type='checkbox']:disabled + span {
    color: #555;
    cursor: initial;
  }
`

type CheckboxProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label: string
  className?: string
}

const Checkbox = ({ label, className, ...props }: CheckboxProps) => {
  return (
    <CheckboxLabelWrapper className={className}>
      <input type="checkbox" {...props} />
      <span>{label}</span>
    </CheckboxLabelWrapper>
  )
}

export default Checkbox
