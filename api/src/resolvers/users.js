import jwt from 'jsonwebtoken'
import uuid from 'uuid'

const mutation = {
  async authorizeUser(root, args, ctx, info) {
    let { username, password } = args

    let session = ctx.driver.session()

    let result = await session.run(
      'MATCH (usr:User {username: {username}, password: {password} }) RETURN usr.id',
      { username, password }
    )

    if (result.records.length == 0) {
      return {
        success: false,
        status: 'Username or password was invalid',
        token: null,
      }
    }

    const record = result.records[0]

    const userId = record.get('usr.id')

    const token = jwt.sign({ id: userId }, process.env.JWT_SECRET)

    return {
      success: true,
      status: 'Authorized',
      token,
    }
  },
  async registerUser(root, args, ctx, info) {
    let { username, password } = args

    let session = ctx.driver.session()
    let findResult = await session.run(
      'MATCH (usr:User {username: {username} }) RETURN usr',
      { username }
    )

    if (findResult.records.length > 0) {
      return {
        success: false,
        status: 'Username is already taken',
        token: null,
      }
    }

    const registerResult = await session.run(
      'CREATE (n:User { username: {username}, password: {password}, id: {id} }) return n.id',
      { username, password, id: uuid() }
    )

    let id = registerResult.records[0].get('n.id')

    const token = jwt.sign({ id }, process.env.JWT_SECRET)

    session.close()

    return {
      success: true,
      status: 'User created',
      token: token,
    }
  },
}

export default {
  mutation,
}
