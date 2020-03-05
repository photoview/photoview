import React from 'react'
import styled from 'styled-components'
import SearchBar from './Searchbar'

const Container = styled.div`
  height: 60px;
  width: 100%;
  display: inline-flex;
  position: fixed;
  background: white;
  top: 0;
  /* border-bottom: 1px solid rgba(0, 0, 0, 0.1); */
  box-shadow: 0 0 2px rgba(0, 0, 0, 0.3);
`

const Title = styled.h1`
  font-size: 36px;
  padding: 5px 12px;
  flex-grow: 1;
`

const Header = () => (
  <Container>
    <Title>Photoview</Title>
    {localStorage.getItem('token') ? <SearchBar /> : null}
  </Container>
)

export default Header
