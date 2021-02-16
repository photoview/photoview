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
  background-color: #eee;
  position: relative;
  border-radius: 50%;
  width: 150px;
  height: 150px;
  object-fit: fill;
  margin: 12px;
  overflow: hidden;
`

const FaceImage = styled(ProtectedImage)`
  position: absolute;
  width: 100%;
  top: 50%;
  transform: translateY(-50%)
    ${({ origin, scale }) =>
      `translate(${(0.5 - origin.x) * 100}%, ${
        (0.5 - origin.y) * 100
      }%) scale(${scale * 0.8})`};

  transform-origin: ${({ origin }) => `${origin.x * 100}% ${origin.y * 100}%`};
  object-fit: cover;
`

const FaceLabel = styled.div`
  color: ${({ labeled }) => (labeled ? 'black' : '#aaa')};
  margin: 12px 12px 24px;
  text-align: center;
`

const FaceGroup = ({ group }) => {
  const previewFace = group.imageFaces[0]

  const rect = previewFace.rectangle

  let scale = Math.min(1 / (rect.maxX - rect.minX), 1 / (rect.maxY - rect.minY))

  let origin = {
    x: (rect.minX + rect.maxX) / 2,
    y: (rect.minY + rect.maxY) / 2,
  }

  return (
    <Link to={`/people/${group.id}`}>
      <CircleImageWrapper>
        <FaceImage
          scale={scale}
          origin={origin}
          src={previewFace.media.thumbnail.url}
        />
      </CircleImageWrapper>
      <FaceLabel labeled={!!group.label}>
        {group.label ?? 'Unlabeled'}
      </FaceLabel>
    </Link>
  )
}

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
