import React from 'react'
import styled from 'styled-components'
import SearchBar from './Searchbar'

import logoPath from '../../assets/photoview-logo.svg'
import { authToken } from '../../authentication'

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
  font-weight: 400;
  padding: 2px 12px;
  flex-grow: 1;
  min-width: 245px;

  @media (max-width: 400px) {
    min-width: auto;

    & span {
      display: none;
    }
  }
`

const Logo = styled.img`
  width: 42px;
  height: 42px;
  display: inline-block;
  vertical-align: middle;
  margin-right: 8px;
`

const LogoText = styled.span`
  vertical-align: middle;
`

const Header = () => (
  <Container>
    <Title>
      <Logo src={logoPath} alt="logo" />
      <LogoText>Photoview</LogoText>
    </Title>
    {authToken() ? <SearchBar /> : null}
  </Container>
)

export default Header
