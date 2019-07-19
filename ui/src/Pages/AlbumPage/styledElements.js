import styled from 'styled-components'

export const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
`

export const PhotoContainer = styled.div`
  flex-grow: 1;
  height: 200px;
  margin: 4px;
  background-color: #eee;
  position: relative;
`

export const Photo = styled.img`
  height: 200px;
  min-width: 100%;
  position: relative;
  object-fit: cover;
`

export const PhotoOverlay = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;

  ${props =>
    props.active &&
    `
      border: 4px solid rgba(65, 131, 196, 0.6);
    `}
`

export const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`
