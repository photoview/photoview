import jwt from 'jsonwebtoken'

export const getUserFromToken = async function(token, driver) {
  const tokenContent = jwt.verify(token, process.env.JWT_SECRET)
  const userId = tokenContent.id

  const session = driver.session()

  const userResult = await session.run(
    'MATCH (u:User {id: {userId}}) RETURN u',
    {
      userId,
    }
  )

  if (userResult.records.length == 0) {
    throw new Error(`User was not found`)
  }

  let user = userResult.records[0].toObject().u.properties

  session.close()

  return user
}

export const getTokenFromBearer = bearer => {
  let token = bearer

  if (!token) {
    throw new Error('Missing auth token')
  }

  if (!token.toLowerCase().startsWith('bearer ')) {
    throw new Error('Invalid auth token')
  }

  token = token.substr(7)

  return token
}
