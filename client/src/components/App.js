import React, { Component } from 'react'
import EmailForm from '../containers/EmailForm'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap/dist/css/bootstrap-theme.css'
import './App.css'

class App extends Component {
  render() {
    return (
      <div>
        <nav className="navbar navbar-inverse navbar-fixed-top">
          <div className="container">
            <div className="navbar-header">
              <a href="#" className="navbar-brand">Gomail</a>
            </div>
            <div id="navbar" className="navbar-collapse collapse">
              <ul className="nav navbar-nav">
                <li className="active">
                  <a href="#">Send</a>
                </li>
              </ul>
            </div>
          </div>
        </nav>
        <div className="container">
          <div className="alert alert-info">
            <strong>DISCLAIMER:</strong> Please use emails ending with <strong>shanab.me</strong> for the field "From Email". Sending from an email under a domain other than <strong>shanab.me</strong> will probably not work, and might also have a bad impact on <strong>shanab.me</strong>'s domain reputation.
          </div>
          <EmailForm />
        </div>
      </div>
    );
  }
}

export default App
