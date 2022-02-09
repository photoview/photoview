import React, { createRef, useEffect, useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import Layout from '../../components/layout/Layout'
import styled from 'styled-components'
import { Link, useParams } from 'react-router-dom'
import SingleFaceGroup from './SingleFaceGroup/SingleFaceGroup'
import { Button, TextField } from '../../primitives/form/Input'
import FaceCircleImage from './FaceCircleImage'
import useScrollPagination from '../../hooks/useScrollPagination'
import PaginateLoader from '../../components/PaginateLoader'
import { useTranslation } from 'react-i18next'
import {
  setGroupLabel,
  setGroupLabelVariables,
} from './__generated__/setGroupLabel'
import {
  myFaces,
  myFacesVariables,
  myFaces_myFaceGroups,
} from './__generated__/myFaces'
import { recognizeUnlabeledFaces } from './__generated__/recognizeUnlabeledFaces'
import { isNil, tailwindClassNames } from '../../helpers/utils'
import classNames from 'classnames'

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
          title
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
  mutation setGroupLabel($groupID: ID!, $label: String) {
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

type FaceDetailsWrapperProps = {
  labeled: boolean
  className?: string
} & React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLSpanElement>,
  HTMLSpanElement
>

const FaceDetailsWrapperInner = ({
  labeled,
  children,
  className,
  ...otherProps
}: FaceDetailsWrapperProps) => (
  <span
    {...otherProps}
    className={classNames(
      className,
      `${labeled ? '' : 'text-gray-400 dark:text-gray-500'}`
    )}
  >
    {children}
  </span>
)

const FaceDetailsWrapper = styled(FaceDetailsWrapperInner)`
  &:hover,
  &:focus-visible {
    color: #2683ca;
  }
`

type FaceDetailsProps = {
  group: {
    __typename: 'FaceGroup'
    id: string
    label: string | null
    imageFaceCount: number
  }
  className?: string
  textFieldClassName?: string
  editLabel: boolean
  setEditLabel: React.Dispatch<React.SetStateAction<boolean>>
}

export const FaceDetails = ({
  group,
  className,
  textFieldClassName,
  editLabel,
  setEditLabel,
}: FaceDetailsProps) => {
  const { t } = useTranslation()
  const [inputValue, setInputValue] = useState(group.label ?? '')
  const inputRef = createRef<HTMLInputElement>()

  const [setGroupLabel, { loading }] = useMutation<
    setGroupLabel,
    setGroupLabelVariables
  >(SET_GROUP_LABEL_MUTATION, {
    variables: {
      groupID: group.id,
    },
  })

  const resetLabel = () => {
    setInputValue(group.label ?? '')
    setEditLabel(false)
  }

  useEffect(() => {
    inputRef.current?.focus()
  }, [inputRef])

  useEffect(() => {
    if (!loading) {
      resetLabel()
    }
  }, [loading])

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key == 'Escape') {
      resetLabel()
      return
    }
  }

  let label
  if (!editLabel) {
    label = (
      <FaceDetailsWrapper
        className={tailwindClassNames(
          className,
          'whitespace-nowrap inline-block overflow-hidden overflow-clip'
        )}
        labeled={!!group.label}
        onClick={() => setEditLabel(true)}
      >
        <FaceImagesCount>{group.imageFaceCount}</FaceImagesCount>
        <button className="">
          {group.label ?? t('people_page.face_group.unlabeled', 'Unlabeled')}
        </button>
        {/* <EditIcon name="pencil" /> */}
      </FaceDetailsWrapper>
    )
  } else {
    label = (
      <FaceDetailsWrapper className={className} labeled={!!group.label}>
        <TextField
          className={textFieldClassName}
          loading={loading}
          ref={inputRef}
          // size="mini"
          placeholder={t('people_page.face_group.label_placeholder', 'Label')}
          // icon="arrow right"
          value={inputValue}
          action={() =>
            setGroupLabel({
              variables: {
                groupID: group.id,
                label: inputValue == '' ? null : inputValue,
              },
            })
          }
          onKeyDown={onKeyDown}
          onChange={e => setInputValue(e.target.value)}
          onBlur={() => {
            resetLabel()
          }}
        />
      </FaceDetailsWrapper>
    )
  }

  return label
}

const FaceImagesCount = styled.span.attrs({
  className:
    'bg-gray-100 text-gray-900 dark:bg-gray-400 dark:text-black text-sm px-1 mr-2 rounded-md',
})``

type FaceGroupProps = {
  group: myFaces_myFaceGroups
}

export const FaceGroup = ({ group }: FaceGroupProps) => {
  const previewFace = group.imageFaces[0]
  const [editLabel, setEditLabel] = useState(false)

  return (
    <div className="m-3">
      <Link to={`/people/${group.id}`}>
        <FaceCircleImage imageFace={previewFace} selectable />
      </Link>
      <FaceDetails
        className="block cursor-pointer text-center w-full mt-3"
        textFieldClassName="w-[140px]"
        group={group}
        editLabel={editLabel}
        setEditLabel={setEditLabel}
      />
    </div>
  )
}

const FaceGroupsWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  margin-top: 24px;
`

/**
 * Renders a page that shows a gallery of all people
 */
export const PeoplePage = () => {
  const { t } = useTranslation()
  const { data, error, loading, fetchMore } = useQuery<
    myFaces,
    myFacesVariables
  >(MY_FACES_QUERY, {
    variables: {
      limit: 50,
      offset: 0,
    },
  })

  const [recognizeUnlabeled, { loading: recognizeUnlabeledLoading }] =
    useMutation<recognizeUnlabeledFaces>(RECOGNIZE_UNLABELED_FACES_MUTATION)

  const { containerElem, finished: finishedLoadingMore } =
    useScrollPagination<myFaces>({
      loading,
      fetchMore,
      data,
      getItems: data => data.myFaceGroups,
    })

  if (error) {
    return <div>{error.message}</div>
  }

  let faces = null
  if (data) {
    faces = data.myFaceGroups.map(faceGroup => (
      <FaceGroup key={faceGroup.id} group={faceGroup} />
    ))
  }

  return (
    <Layout title={t('title.people', 'People')}>
      <Button
        disabled={recognizeUnlabeledLoading}
        onClick={() => {
          recognizeUnlabeled()
        }}
      >
        {t(
          'people_page.recognize_unlabeled_faces_button',
          'Recognize unlabeled faces'
        )}
      </Button>
      <FaceGroupsWrapper ref={containerElem}>{faces}</FaceGroupsWrapper>
      <PaginateLoader
        active={!finishedLoadingMore && !loading}
        text={t('general.loading.paginate.faces', 'Loading more people')}
      />
    </Layout>
  )
}

/**
 * Renders a page for an individual person
 */
export const PersonPage = () => {
  const { t } = useTranslation()
  const { person } = useParams()
  if (isNil(person))
    throw new Error('Expected `person` parameter to be defined')

  return (
    <Layout title={t('title.people', 'People')}>
      <SingleFaceGroup faceGroupID={person} />
    </Layout>
  )
}
