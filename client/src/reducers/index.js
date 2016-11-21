import { SUBMIT_EMAIL_REQUEST, SUBMIT_EMAIL_RESPONSE, SUBMIT_EMAIL_ERROR } from '../actions'

const initialState = {
  errors: null,
  messageId: null,
  isSubmittingEmail: false,
}
const emails = (state = initialState, action) => {
  switch (action.type) {
    case SUBMIT_EMAIL_REQUEST:
      return {
        ...state,
        isSubmittingEmail: true,
        errors: null
      }
    case SUBMIT_EMAIL_RESPONSE:
      return {
        ...state,
        isSubmittingEmail: false,
        errors: null,
        messageId: action.payload.messageId
      }
    case SUBMIT_EMAIL_ERROR:
      return {
        ...state,
        isSubmittingEmail: false,
        errors: action.payload.errors
      }
    default:
      return state
  }
}

export default emails
