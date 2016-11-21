import React, { Component } from 'react'

class FieldWithErrors extends Component {
  render() {
    const { id, label, errors } = this.props
    const error = errors ? errors[id] : null
    let formGroupClass = 'form-group'
    let errorSpan
    if (error) {
      formGroupClass += ' has-error'
      errorSpan = <span className="help-block">{error}</span>
    }
    return (
      <div className={formGroupClass}>
        <label className="control-label col-sm-2" htmlFor={id}>{label}</label>
        <div className="col-sm-10">
          {this.props.children}
          {errorSpan}
        </div>
      </div>
    )
  }
}

export default FieldWithErrors
