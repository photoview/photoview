import styled from 'styled-components'

export const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

export const Photo = styled.img`
  margin: 4px;
  background-color: #eee;
  height: 200px;
  flex-grow: 1;
  object-fit: cover;

  ${props =>
    props.active &&
    `
      will-change: transform;
      position: relative;
      z-index: 999;
    `}
`

export const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`
