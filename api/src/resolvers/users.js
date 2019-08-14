import jwt from 'jsonwebtoken'
import generateID from '../id-generator'
import fs from 'fs-extra'
import { neo4jgraphql } from 'neo4j-graphql-js'

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
      { username, password, id: generateID(), rootPath }
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
  async updateUser(root, args, ctx, info) {
    if (args.rootPath) {
      if (!(await fs.exists(args.rootPath))) {
        throw Error('New root path not found in server filesystem')
      }
    }

    return neo4jgraphql(root, args, ctx, info)
  },
  async createUser(root, args, ctx, info) {
    if (args.rootPath) {
      if (!(await fs.exists(args.rootPath))) {
        throw Error('Root path not found in server filesystem')
      }
    }

    args.id = generateID()

    return neo4jgraphql(root, args, ctx, info)
  },
  async changeUserPassword(root, args, ctx, info) {
    const { newPassword, id } = args

    const session = ctx.driver.session()

    await session.run(
      `MATCH (u:User { id: {id} }) SET u.password = {newPassword}`,
      {
        id,
        newPassword,
      }
    )

    session.close

    return {
      success: true,
      errorMessage: null,
    }
  },
}

export const registerUser = Mutation.registerUser
export const authorizeUser = Mutation.authorizeUser

const Query = {
  myUser(root, args, ctx, info) {
    let customArgs = {
      filter: {},
      ...args,
    }

    customArgs.filter.id = ctx.user.id

    return neo4jgraphql(root, customArgs, ctx, info)
  },
}

export default {
  Mutation,
  Query,
}
