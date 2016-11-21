import axios from 'axios'

export const SUBMIT_EMAIL_REQUEST = 'SUBMIT_EMAIL_REQUEST'
export const SUBMIT_EMAIL_RESPONSE = 'SUBMIT_EMAIL_RESPONSE'
export const SUBMIT_EMAIL_ERROR = 'SUBMIT_EMAIL_ERROR'

const API = axios.create({
  baseURL: 'http://gomail-api.shanab.me',
  timeout: 1000,
  headers: { 'Content-Type': 'application/json' }
})

export const submitEmailRequest = () => ({
  type: SUBMIT_EMAIL_REQUEST,
})

export const submitEmailResponse = payload => ({
  type: SUBMIT_EMAIL_RESPONSE,
  payload
})

export const submitEmailError = payload => ({
  type: SUBMIT_EMAIL_ERROR,
  payload
})

const normalizeErrors = (err) => {
  if (err.response && err.response.data && err.response.data.errors) {
    return err.response.data.errors
  } else {
    return { base: err.message }
  }
}

export const submitEmail = payload => dispatch => {
  dispatch(submitEmailRequest())

  const success = (response) => dispatch(submitEmailResponse({ email: response.data.email }))
  const failure = (err) => dispatch(submitEmailError({ errors: normalizeErrors(err) }))
  return API.post('/email/send', { email: payload })
    .then(success, failure)
}
