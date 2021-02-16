import React from 'react'
import PropTypes from 'prop-types'
import { gql, useQuery } from '@apollo/client'
import Layout from '../../Layout'
import { ProtectedImage } from '../../components/photoGallery/ProtectedMedia'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import SingleFaceGroup from './SingleFaceGroup'

const MY_FACES_QUERY = gql`
  query myFaces {
    myFaceGroups {
      id
      label
      imageFaces {
        id
        rectangle {
          minX
          maxX
          minY
          maxY
        }
        media {
          id
          thumbnail {
            url
            width
            height
          }
        }
      }
    }
  }
`

const CircleImageWrapper = styled.div`
  border-radius: 50%;
  width: 150px;
  height: 150px;
  object-fit: fill;
  margin: 12px;
  overflow: hidden;
`

const FaceImage = styled(ProtectedImage)`
  width: 100%;
  height: 100%;
  object-fit: cover;
`

const FaceGroup = ({ group }) => (
  <Link to={`/people/${group.id}`}>
    <CircleImageWrapper>
      <FaceImage src={group.imageFaces[0].media.thumbnail.url} />
    </CircleImageWrapper>
  </Link>
)

FaceGroup.propTypes = {
  group: PropTypes.any,
}

const FaceGroupsWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
`

const PeoplePage = ({ match }) => {
  const { data, error } = useQuery(MY_FACES_QUERY)

  if (error) {
    return error.message
  }

  const faceGroup = match.params.person
  if (faceGroup) {
    return (
      <Layout>
        <SingleFaceGroup
          faceGroup={data?.myFaceGroups?.find(x => x.id == faceGroup)}
        />
      </Layout>
    )
  }

  let faces = null
  if (data) {
    faces = data.myFaceGroups.map(faceGroup => (
      <FaceGroup key={faceGroup.id} group={faceGroup} />
    ))
  }

  return (
    <Layout title={'People'}>
      <FaceGroupsWrapper>{faces}</FaceGroupsWrapper>
    </Layout>
  )
}

PeoplePage.propTypes = {
  match: PropTypes.object.isRequired,
}

export default PeoplePage
