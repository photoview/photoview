import { useMutation } from '@apollo/client'
import PropTypes from 'prop-types'
import React, { useState, useEffect, createRef } from 'react'
import { Dropdown, Input } from 'semantic-ui-react'
import styled from 'styled-components'
import { SET_GROUP_LABEL_MUTATION } from '../PeoplePage'
import MergeFaceGroupsModal from './MergeFaceGroupsModal'

const TitleWrapper = styled.div`
  min-height: 3.5em;
`

const TitleLabel = styled.h1`
  display: inline-block;
  color: ${({ labeled }) => (labeled ? 'black' : '#888')};
  margin-right: 12px;
`

const TitleDropdown = styled(Dropdown)`
  vertical-align: middle;
  margin-top: -10px;
  color: #888;

  &:hover {
    color: #1e70bf;
  }
`

const FaceGroupTitle = ({ faceGroup }) => {
  const [editLabel, setEditLabel] = useState(false)
  const [inputValue, setInputValue] = useState(faceGroup?.label ?? '')
  const inputRef = createRef()
  const [mergeModalOpen, setMergeModalOpen] = useState(false)

  const [setGroupLabel, { loading: setLabelLoading }] = useMutation(
    SET_GROUP_LABEL_MUTATION,
    {
      variables: {
        groupID: faceGroup?.id,
      },
    }
  )

  const resetLabel = () => {
    setInputValue(faceGroup?.label ?? '')
    setEditLabel(false)
  }

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus()
    }
  }, [inputRef])

  useEffect(() => {
    if (!setLabelLoading) {
      resetLabel()
    }
  }, [setLabelLoading])

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

  let title
  if (!editLabel) {
    title = (
      <TitleWrapper>
        <TitleLabel labeled={!!faceGroup?.label}>
          {faceGroup?.label ?? 'Unlabeled person'}
        </TitleLabel>
        <TitleDropdown
          icon={{
            name: 'settings',
            size: 'large',
          }}
        >
          <Dropdown.Menu>
            <Dropdown.Item
              icon="pencil"
              text={faceGroup?.label ? 'Change Label' : 'Add Label'}
              onClick={() => setEditLabel(true)}
            />
            <Dropdown.Item
              icon="object ungroup"
              text="Merge Face"
              alt="Merge this group into another"
              onClick={() => setMergeModalOpen(true)}
            />
          </Dropdown.Menu>
        </TitleDropdown>
      </TitleWrapper>
    )
  } else {
    title = (
      <TitleWrapper>
        <Input
          loading={setLabelLoading}
          ref={inputRef}
          placeholder="Label"
          icon="arrow right"
          value={inputValue}
          onKeyUp={onKeyUp}
          onChange={e => setInputValue(e.target.value)}
          onBlur={() => {
            resetLabel()
          }}
        />
      </TitleWrapper>
    )
  }

  return (
    <>
      {title}
      <MergeFaceGroupsModal
        open={mergeModalOpen}
        setOpen={setMergeModalOpen}
        sourceFaceGroup={faceGroup}
      />
    </>
  )
}

FaceGroupTitle.propTypes = {
  faceGroup: PropTypes.object,
}

export default FaceGroupTitle
