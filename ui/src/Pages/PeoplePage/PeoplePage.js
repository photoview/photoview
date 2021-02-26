import React, { createRef, useEffect, useState } from 'react'
import PropTypes from 'prop-types'
import { gql, useMutation, useQuery } from '@apollo/client'
import Layout from '../../Layout'
import styled from 'styled-components'
import { Link } from 'react-router-dom'
import SingleFaceGroup from './SingleFaceGroup/SingleFaceGroup'
import { Button, Icon, Input } from 'semantic-ui-react'
import FaceCircleImage from './FaceCircleImage'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../../components/PaginateLoader'

export const MY_FACES_QUERY = gql`
  query myFaces($limit: Int, $offset: Int) {
    myFaceGroups(paginate: { limit: $limit, offset: $offset }) {
      id
      label
      imageFaceCount
      imageFaces(paginate: { limit: 1 }) {
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
        <FaceImagesCount>{group.imageFaceCount}</FaceImagesCount>
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
  margin-top: 24px;
`

const PeopleGallery = () => {
  const { data, error, loading, fetchMore } = useQuery(MY_FACES_QUERY, {
    variables: {
      limit: 50,
      offset: 0,
    },
  })

  const [
    recognizeUnlabeled,
    { loading: recognizeUnlabeledLoading },
  ] = useMutation(RECOGNIZE_UNLABELED_FACES_MUTATION)

  const { containerElem, finished: finishedLoadingMore } = useScrollPagination({
    loading,
    fetchMore,
    data,
    getItems: data => data.myFaceGroups,
  })

  if (error) {
    return error.message
  }

  let faces = null
  if (data) {
    faces = data.myFaceGroups.map(faceGroup => (
      <FaceGroup key={faceGroup.id} group={faceGroup} />
    ))
  }

  return (
    <Layout title={'People'}>
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
      <FaceGroupsWrapper ref={containerElem}>{faces}</FaceGroupsWrapper>
      <PaginateLoader
        active={!finishedLoadingMore && !loading}
        text="Loading more people"
      />
    </Layout>
  )
}

const PeoplePage = ({ match }) => {
  const faceGroup = match.params.person
  if (faceGroup) {
    return (
      <Layout>
        <SingleFaceGroup faceGroupID={faceGroup} />
      </Layout>
    )
  } else {
    return <PeopleGallery />
  }
}

PeoplePage.propTypes = {
  match: PropTypes.object.isRequired,
}

export default PeoplePage
