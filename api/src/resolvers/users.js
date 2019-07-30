import jwt from 'jsonwebtoken'
import uuid from 'uuid'
import fs from 'fs-extra'

const Mutation = {
  async authorizeUser(root, args, ctx, info) {
    console.log('Authorize user')
    let { username, password } = args

    let session = ctx.driver.session()

    let result = await session.run(
      'MATCH (usr:User {username: {username}, password: {password} }) RETURN usr.id, usr.admin',
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
    const userAdmin = record.get('usr.admin')

    let roles = []

    if (userAdmin) {
      roles.push('admin')
    }

    const token = jwt.sign({ id: userId, roles }, process.env.JWT_SECRET)

    return {
      success: true,
      status: 'Authorized',
      token,
    }
  },
  async registerUser(root, args, ctx, info) {
    let { username, password, rootPath } = args

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

    if (!(await fs.exists(rootPath))) {
      return {
        success: false,
        status: 'Root path does not exist on the server',
        token: null,
      }
    }

    const registerResult = await session.run(
      'CREATE (n:User { username: {username}, password: {password}, id: {id}, admin: false, rootPath: {rootPath} }) return n.id',
      { username, password, id: uuid(), rootPath }
    )

    let id = registerResult.records[0].get('n.id')

    const token = jwt.sign({ id, roles: [] }, process.env.JWT_SECRET)

    session.close()

    return {
      success: true,
      status: 'User created',
      token: token,
    }
  },
}

export const registerUser = Mutation.registerUser

export default {
  Mutation,
}
