import React, { createRef, useEffect, useState } from 'react'
import PropTypes from 'prop-types'
import { gql, useMutation, useQuery } from '@apollo/client'
import Layout from '../../Layout'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import SingleFaceGroup from './SingleFaceGroup/SingleFaceGroup'
import { Button, Icon, Input } from 'semantic-ui-react'
import FaceCircleImage from './FaceCircleImage'

export const MY_FACES_QUERY = gql`
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
          type
          thumbnail {
            url
            width
            height
          }
          highRes {
            url
          }
          favorite
        }
      }
    }
  }
`

export const SET_GROUP_LABEL_MUTATION = gql`
  mutation($groupID: ID!, $label: String) {
    setFaceGroupLabel(faceGroupID: $groupID, label: $label) {
      id
      label
    }
  }
`

const RECOGNIZE_UNLABELED_FACES_MUTATION = gql`
  mutation recognizeUnlabeledFaces {
    recognizeUnlabeledFaces {
      id
    }
  }
`

const FaceDetailsButton = styled.button`
  color: ${({ labeled }) => (labeled ? 'black' : '#aaa')};
  margin: 12px auto 24px;
  text-align: center;
  display: block;
  background: none;
  border: none;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    color: #2683ca;
  }
`

const FaceLabel = styled.span``

const FaceDetails = ({ group }) => {
  const [editLabel, setEditLabel] = useState(false)
  const [inputValue, setInputValue] = useState(group.label ?? '')
  const inputRef = createRef()

  const [setGroupLabel, { loading }] = useMutation(SET_GROUP_LABEL_MUTATION, {
    variables: {
      groupID: group.id,
    },
  })

  const resetLabel = () => {
    setInputValue(group.label ?? '')
    setEditLabel(false)
  }

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus()
    }
  }, [inputRef])

  useEffect(() => {
    if (!loading) {
      resetLabel()
    }
  }, [loading])

  const onKeyUp = e => {
    if (e.key == 'Escape') {
      resetLabel()
      return
    }

    if (e.key == 'Enter') {
      setGroupLabel({
        variables: {
          label: e.target.value == '' ? null : e.target.value,
        },
      })
      return
    }
  }

  let label
  if (!editLabel) {
    label = (
      <FaceDetailsButton
        labeled={!!group.label}
        onClick={() => setEditLabel(true)}
      >
        <FaceImagesCount>{group.imageFaces.length}</FaceImagesCount>
        <FaceLabel>{group.label ?? 'Unlabeled'}</FaceLabel>
        <EditIcon name="pencil" />
      </FaceDetailsButton>
    )
  } else {
    label = (
      <FaceDetailsButton labeled={!!group.label}>
        <Input
          loading={loading}
          ref={inputRef}
          size="mini"
          placeholder="Label"
          icon="arrow right"
          value={inputValue}
          onKeyUp={onKeyUp}
          onChange={e => setInputValue(e.target.value)}
          onBlur={() => {
            resetLabel()
          }}
        />
      </FaceDetailsButton>
    )
  }

  return label
}

FaceDetails.propTypes = {
  group: PropTypes.object.isRequired,
}

const FaceImagesCount = styled.span`
  background-color: #eee;
  color: #222;
  font-size: 0.9em;
  padding: 0 4px;
  margin-right: 6px;
  border-radius: 4px;
`

const EditIcon = styled(Icon)`
  margin-left: 6px !important;
  opacity: 0 !important;

  transition: opacity 100ms;

  ${FaceDetailsButton}:hover &, ${FaceDetailsButton}:focus-visible & {
    opacity: 1 !important;
  }
`

const FaceGroup = ({ group }) => {
  const previewFace = group.imageFaces[0]

  return (
    <div style={{ margin: '12px' }}>
      <Link to={`/people/${group.id}`}>
        <FaceCircleImage imageFace={previewFace} selectable />
      </Link>
      <FaceDetails group={group} />
    </div>
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

  const [
    recognizeUnlabeled,
    { loading: recognizeUnlabeledLoading },
  ] = useMutation(RECOGNIZE_UNLABELED_FACES_MUTATION)

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
      <Button
        loading={recognizeUnlabeledLoading}
        disabled={recognizeUnlabeledLoading}
        onClick={() => {
          recognizeUnlabeled()
        }}
      >
        <Icon name="sync" />
        Recognize unlabeled faces
      </Button>
    </Layout>
  )
}

PeoplePage.propTypes = {
  match: PropTypes.object.isRequired,
}

export default PeoplePage
