import uuid from 'uuid'

function generateID() {
  return uuid().substr(-12)
}

export default generateID
