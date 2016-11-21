import { connect } from 'react-redux'
import { submitEmail } from '../actions'
import EmailForm from '../components/EmailForm'

const mapStateToProps = (state) => ({
  errors: state.errors,
  // Not really using messageId nor isSubmittingEmail for now
  messageId: state.messageId,
  isSubmittingEmail: state.isSubmittingEmail,
})

const mapDispatchToProps = {
  submitEmail
}

export default connect(mapStateToProps, mapDispatchToProps)(EmailForm)
