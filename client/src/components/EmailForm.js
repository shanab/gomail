import React, { Component } from 'react'
import FieldWithErrors from './FieldWithErrors'

class EmailForm extends Component {
  constructor(props) {
    super(props)
    this.state = { fromName: '', fromEmail: '', toName: '', toEmail: '', subject: '', body: '' }

    // Bindings
    this.handleChange = this.handleChange.bind(this)
    this.handleSubmit = this.handleSubmit.bind(this)
  }

  handleChange(event) {
    this.setState({ [event.target.name]: event.target.value })
  }

  handleSubmit() {
    const { submitEmail } = this.props
    const { fromName, fromEmail, toName, toEmail, subject, body } = this.state
    submitEmail({ fromName, fromEmail, toName, toEmail, subject, body })
  }

  renderBaseError() {
    const { errors } = this.props
    return errors && errors['base'] ?
      (
        <div className="alert alert-danger">
          {errors['base']}
        </div>
      ) :
      null
  }

  render() {
    const { errors } = this.props
    return (
      <div className="send-email">
        {this.renderBaseError()}
        <div className="form-horizontal">
          <FieldWithErrors errors={errors} id="fromName" label="From Name">
            <input type="text" className="form-control" name="fromName" id="fromName" value={this.state.fromName} onChange={this.handleChange} />
          </FieldWithErrors>
          <FieldWithErrors errors={errors} id="fromEmail" label="From Email">
            <input type="email" className="form-control" name="fromEmail" id="fromEmail" value={this.state.fromEmail} onChange={this.handleChange} />
          </FieldWithErrors>
          <FieldWithErrors errors={errors} id="toName" label="To Name">
            <input type="text" className="form-control" name="toName" id="toName" value={this.state.toName} onChange={this.handleChange} />
          </FieldWithErrors>
          <FieldWithErrors errors={errors} id="toEmail" label="To Email">
            <input type="email" className="form-control" name="toEmail" id="toEmail" value={this.state.toEmail} onChange={this.handleChange} />
          </FieldWithErrors>
          <FieldWithErrors errors={errors} id="subject" label="Subject">
            <input type="text" className="form-control" name="subject" id="subject" value={this.state.subject} onChange={this.handleChange} />
          </FieldWithErrors>
          <FieldWithErrors errors={errors} id="body" label="Body">
            <textarea name="body" id="body" rows="15" className="form-control"></textarea>
          </FieldWithErrors>
          <div className="form-group">
            <div className="col-sm-offset-2 col-sm-10">
              <button className="btn btn-default" onClick={this.handleSubmit}>Submit</button>
            </div>
          </div>
        </div>
      </div>
    )
  }
}

export default EmailForm
