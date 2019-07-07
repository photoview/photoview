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
    let result = await session.run(
      'MATCH (usr:User {username: {username} }) RETURN usr',
      { username }
    )

    if (result.records.length > 0) {
      return {
        success: false,
        status: 'Username is already taken',
        token: null,
      }
    }

    await session.run(
      'CREATE (n:User { username: {username}, password: {password}, id: {id} }) return n',
      { username, password, id: uuid() }
    )

    session.close()

    return {
      success: true,
      status: 'User created',
      token: 'yay',
    }
  },
}

export default {
  mutation,
}
