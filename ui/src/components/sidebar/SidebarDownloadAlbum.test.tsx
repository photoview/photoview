import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SidebarDownloadAlbum, {
  generateDownloadUrl,
} from './SidebarDownloadAlbum'

import * as authentication from '../../helpers/authentication'

vi.mock('../../helpers/authentication.ts')

const authToken = vi.mocked(authentication.authToken)

describe('SidebarDownloadAlbum', () => {
  test('render downloadAlbum, unauthorized', () => {
    authToken.mockImplementation(() => null)

    render(<SidebarDownloadAlbum albumID="30" />)

    // check if the download is the title
    const h2 = document.querySelector('h2')
    expect(h2?.textContent).toMatch(/Download/)
  })

  test('generate correct url, unauthorized as share', () => {
    Object.defineProperty(window, 'location', {
      value: new URL('http://localhost:1234/share/qDSL5I1N'),
      configurable: true,
    })

    authToken.mockImplementation(() => null)
    const url = generateDownloadUrl('30', {
      title: 'testAlbum',
      description: '',
      purpose: 'original',
    })

    expect(url).toContain('/download/album/30/original?token=qDSL5I1N')
  })

  test('generate correct url, authorized as share', () => {
    Object.defineProperty(window, 'location', {
      value: new URL('http://localhost:1234/album/30'),
      configurable: true,
    })

    authToken.mockImplementation(() => 'token-here')
    const url = generateDownloadUrl('30', {
      title: 'testAlbum',
      description: '',
      purpose: 'original',
    })

    expect(url).toContain('/download/album/30/original')
  })

  test('render downloadAlbum and click the original "button", authorized', async () => {
    authToken.mockImplementation(() => 'token-here')

    render(<SidebarDownloadAlbum albumID="30" />)

    await userEvent.click(screen.getByText('Originals'))
    expect(window.location.href).toContain('/download/album/30/original')
  })
})
