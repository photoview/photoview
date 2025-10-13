import { vi } from 'vitest'
// Mock react-i18next following the project pattern
const mockUseTranslation = vi.fn()
vi.mock('react-i18next', () => ({
  useTranslation: mockUseTranslation,
}))
mockUseTranslation.mockReturnValue({
  t: (key: string, defaultValue: string) => defaultValue,
  i18n: { language: 'en' },
})

import React from 'react'
import { render, screen } from '@testing-library/react'
import ExifDetails from './MediaSidebarExif'
import { MediaSidebarMedia } from './MediaSidebar'
import { MediaType } from '../../../__generated__/globalTypes'

describe('ExifDetails', () => {
  test('without EXIF information', () => {
    const media: MediaSidebarMedia = {
      id: '1730',
      title: 'media_name.jpg',
      type: MediaType.Photo,
      exif: {
        id: '0',
        description: null,
        camera: null,
        maker: null,
        lens: null,
        dateShotStr: null,
        exposure: null,
        aperture: null,
        iso: null,
        focalLength: null,
        flash: null,
        exposureProgram: null,
        coordinates: null,
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<ExifDetails media={media} />)

    expect(screen.queryByText('Description')).not.toBeInTheDocument()
    expect(screen.queryByText('Camera')).not.toBeInTheDocument()
    expect(screen.queryByText('Maker')).not.toBeInTheDocument()
    expect(screen.queryByText('Lens')).not.toBeInTheDocument()
    expect(screen.queryByText('Program')).not.toBeInTheDocument()
    expect(screen.queryByText('Date shot')).not.toBeInTheDocument()
    expect(screen.queryByText('Exposure')).not.toBeInTheDocument()
    expect(screen.queryByText('Aperture')).not.toBeInTheDocument()
    expect(screen.queryByText('ISO')).not.toBeInTheDocument()
    expect(screen.queryByText('Focal length')).not.toBeInTheDocument()
    expect(screen.queryByText('Flash')).not.toBeInTheDocument()
    expect(screen.queryByText('Coordinates')).not.toBeInTheDocument()
  })

  test('with EXIF information', () => {
    const media: MediaSidebarMedia = {
      id: '1730',
      title: 'media_name.jpg',
      type: MediaType.Photo,
      exif: {
        id: '1666',
        description: 'Media description',
        camera: 'Canon EOS R',
        maker: 'Canon',
        lens: 'TAMRON SP 24-70mm F/2.8',
        dateShotStr: '2021-01-23T20:50:18Z',
        exposure: 0.016666666666666666,
        aperture: 2.8,
        iso: 100,
        focalLength: 24,
        flash: 9,
        exposureProgram: 3,
        coordinates: {
          __typename: 'Coordinates',
          latitude: 41.40338,
          longitude: 2.17403,
        },
        __typename: 'MediaEXIF',
      },
      __typename: 'Media',
    }

    render(<ExifDetails media={media} />)

    expect(screen.getByText('Description')).toBeInTheDocument()

    expect(screen.getByText('Camera')).toBeInTheDocument()
    expect(screen.getByText('Canon EOS R')).toBeInTheDocument()

    expect(screen.getByText('Maker')).toBeInTheDocument()
    expect(screen.getByText('Canon')).toBeInTheDocument()

    expect(screen.getByText('Lens')).toBeInTheDocument()
    expect(screen.getByText('TAMRON SP 24-70mm F/2.8')).toBeInTheDocument()

    expect(screen.getByText('Date shot')).toBeInTheDocument()

    expect(screen.getByText('Exposure')).toBeInTheDocument()
    expect(screen.getByText('1/60')).toBeInTheDocument()

    expect(screen.getByText('Program')).toBeInTheDocument()
    expect(screen.getByText('Aperture priority')).toBeInTheDocument()

    expect(screen.getByText('Aperture')).toBeInTheDocument()
    expect(screen.getByText('f/2.8')).toBeInTheDocument()

    expect(screen.getByText('ISO')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()

    expect(screen.getByText('Focal length')).toBeInTheDocument()
    expect(screen.getByText('24mm')).toBeInTheDocument()

    expect(screen.getByText('Flash')).toBeInTheDocument()
    expect(screen.getByText('On, Fired')).toBeInTheDocument()

    expect(screen.getByText('Coordinates')).toBeInTheDocument()
    expect(screen.getByText('41.40338, 2.17403')).toBeInTheDocument()
  })
})

describe('ExifDetails dateShot formatting', () => {

  const createMediaWithDateShot = (dateShotStr: string | null): MediaSidebarMedia => ({
    id: '1730',
    title: 'media_name.jpg',
    type: MediaType.Photo,
    exif: {
      id: '1666',
      description: null,
      camera: null,
      maker: null,
      lens: null,
      dateShotStr,
      exposure: null,
      aperture: null,
      iso: null,
      focalLength: null,
      flash: null,
      exposureProgram: null,
      coordinates: null,
      __typename: 'MediaEXIF',
    },
    __typename: 'Media',
  })

  describe('RFC3339 with timezone offset', () => {
    test('formats RFC3339 with positive timezone offset', () => {
      const media = createMediaWithDateShot('2023-07-15T14:30:45+02:00')
      render(<ExifDetails media={media} />)

      expect(screen.getByText('Date shot')).toBeInTheDocument()
      // Should show medium date format, time, and timezone offset WITHOUT abbreviation
      expect(screen.getByText(/Jul 15, 2023.*2:30:45 PM.*\+02:00$/)).toBeInTheDocument()
    })

    test('formats RFC3339 with negative timezone offset', () => {
      const media = createMediaWithDateShot('2023-12-15T10:15:30-08:00')
      render(<ExifDetails media={media} />)

      expect(screen.getByText(/Dec 15, 2023.*10:15:30 AM.*-08:00$/)).toBeInTheDocument()
    })

    test('formats RFC3339 with fractional timezone offset', () => {
      const media = createMediaWithDateShot('2023-06-01T16:45:12+05:30')
      render(<ExifDetails media={media} />)

      // India Standard Time with +05:30 offset, no abbreviation
      expect(screen.getByText(/Jun 1, 2023.*4:45:12 PM.*\+05:30$/)).toBeInTheDocument()
    })

    test('formats RFC3339 UTC with Z suffix', () => {
      const media = createMediaWithDateShot('2023-05-10T12:00:00Z')
      render(<ExifDetails media={media} />)

      // Should show UTC offset but no abbreviation
      expect(screen.getByText(/May 10, 2023.*12:00:00 PM.*\+00:00$/)).toBeInTheDocument()
    })

    test('formats RFC3339 with +0000 format', () => {
      const media = createMediaWithDateShot('2023-01-01T12:00:00+0000')
      render(<ExifDetails media={media} />)

      // Should detect +0000 as timezone and show +00:00 offset
      expect(screen.getByText(/Jan 1, 2023.*12:00:00 PM.*\+00:00$/)).toBeInTheDocument()
    })
  })

  describe('RFC3339 without timezone - no offset should be shown', () => {
    test('formats RFC3339 without timezone shows only date and time', () => {
      const media = createMediaWithDateShot('2023-08-25T18:45:30')
      render(<ExifDetails media={media} />)

      expect(screen.getByText('Date shot')).toBeInTheDocument()
      // Should show only date and time, no timezone info
      expect(screen.getByText(/^Aug 25, 2023 6:45:30 PM$/)).toBeInTheDocument()
      expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
    })

    test('formats RFC3339 with milliseconds but no timezone', () => {
      const media = createMediaWithDateShot('2023-03-20T09:30:15.123')
      render(<ExifDetails media={media} />)

      // Should show only date and time, no timezone info
      expect(screen.getByText(/^Mar 20, 2023 9:30:15 AM$/)).toBeInTheDocument()
      expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
    })

    test('handles trimming of whitespace in dateShot', () => {
      const media = createMediaWithDateShot('  2023-04-10T15:20:00  ')
      render(<ExifDetails media={media} />)

      // Should trim whitespace and format correctly without timezone
      expect(screen.getByText(/^Apr 10, 2023 3:20:00 PM$/)).toBeInTheDocument()
      expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
    })
  })

  describe('Translation language-based formatting', () => {
    const mockUseTranslationWithLanguage = (language: string) => {
      mockUseTranslation.mockReturnValue({
        t: (key: string, defaultValue: string) => defaultValue,
        i18n: { language },
      })
    }

    afterEach(() => {
      mockUseTranslation.mockReturnValue({
        t: (key: string, defaultValue: string) => defaultValue,
        i18n: { language: 'en' },
      })
    })

    test('uses English (en) translation language', () => {
      mockUseTranslationWithLanguage('en')

      const mediaWithTz = createMediaWithDateShot('2023-11-05T15:20:10+01:00')
      const { rerender } = render(<ExifDetails media={mediaWithTz} />)

      // English locale DATE_MED format with timezone
      expect(screen.getByText(/Nov 5, 2023.*3:20:10 PM.*\+01:00$/)).toBeInTheDocument()

      // Test without timezone - should use English formatting but no timezone
      const mediaNoTz = createMediaWithDateShot('2023-11-05T15:20:10')
      rerender(<ExifDetails media={mediaNoTz} />)

      expect(screen.getByText(/^Nov 5, 2023 3:20:10 PM$/)).toBeInTheDocument()
      expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
    })

    test('uses German (de) translation language', () => {
      mockUseTranslationWithLanguage('de')

      const mediaWithTz = createMediaWithDateShot('2023-11-05T15:20:10+01:00')
      const { rerender } = render(<ExifDetails media={mediaWithTz} />)

      // German locale DATE_MED format with timezone
      expect(screen.getByText(/5\. Nov\. 2023.*15:20:10.*\+01:00$/)).toBeInTheDocument()

      // Test without timezone - should use German formatting but no timezone
      const mediaNoTz = createMediaWithDateShot('2023-11-05T15:20:10')
      rerender(<ExifDetails media={mediaNoTz} />)

      expect(screen.getByText(/^5\. Nov\. 2023 15:20:10$/)).toBeInTheDocument()
      expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
    })

    test('uses French (fr) translation language', () => {
      mockUseTranslationWithLanguage('fr')

      const mediaWithTz = createMediaWithDateShot('2023-12-25T20:45:30+01:00')
      const { rerender } = render(<ExifDetails media={mediaWithTz} />)

      // French DATE_MED format with timezone
      expect(screen.getByText(/25 déc\. 2023.*20:45:30.*\+01:00$/)).toBeInTheDocument()

      // Test without timezone
      const mediaNoTz = createMediaWithDateShot('2023-12-25T20:45:30')
      rerender(<ExifDetails media={mediaNoTz} />)

      expect(screen.getByText(/^25 déc\. 2023 20:45:30$/)).toBeInTheDocument()
    })

    test('uses Spanish (es) translation language', () => {
      mockUseTranslationWithLanguage('es')

      const media = createMediaWithDateShot('2023-07-20T14:15:30-05:00')
      render(<ExifDetails media={media} />)

      // Spanish locale formatting with timezone
      expect(screen.getByText(/20 jul 2023.*14:15:30.*-05:00$/)).toBeInTheDocument()
    })

    test('uses Danish (da) translation language', () => {
      mockUseTranslationWithLanguage('da')

      const media = createMediaWithDateShot('2023-06-15T12:30:45+02:00')
      render(<ExifDetails media={media} />)

      // Danish locale formatting with timezone
      expect(screen.getByText(/15\. jun\. 2023 12\.30\.45 \+02:00$/)).toBeInTheDocument()
    })

    test('uses Italian (it) translation language', () => {
      mockUseTranslationWithLanguage('it')

      const media = createMediaWithDateShot('2023-09-10T11:20:15+02:00')
      render(<ExifDetails media={media} />)

      // Italian locale formatting with timezone
      expect(screen.getByText(/10 set 2023.*11:20:15.*\+02:00$/)).toBeInTheDocument()
    })

    test('uses Portuguese (pt) translation language', () => {
      mockUseTranslationWithLanguage('pt')

      const media = createMediaWithDateShot('2023-09-10T11:20:15-03:00')
      render(<ExifDetails media={media} />)

      // Portuguese locale formatting with timezone
      expect(screen.getByText(/10 de set\. de 2023.*11:20:15.*-03:00$/)).toBeInTheDocument()
    })

    test('uses Polish (pl) translation language', () => {
      mockUseTranslationWithLanguage('pl')

      const media = createMediaWithDateShot('2023-08-12T16:00:00+02:00')
      render(<ExifDetails media={media} />)

      // Polish locale formatting with timezone
      expect(screen.getByText(/12 sie 2023.*16:00:00.*\+02:00$/)).toBeInTheDocument()
    })

    test('uses Ukrainian (uk) translation language', () => {
      mockUseTranslationWithLanguage('uk')

      const media = createMediaWithDateShot('2023-08-12T16:00:00+03:00')
      render(<ExifDetails media={media} />)

      // Ukrainian locale formatting with timezone
      expect(screen.getByText(/12 серп\. 2023.*16:00:00.*\+03:00$/)).toBeInTheDocument()
    })

    test('uses Swedish (sv) translation language', () => {
      mockUseTranslationWithLanguage('sv')

      const media = createMediaWithDateShot('2023-05-15T14:30:00+02:00')
      render(<ExifDetails media={media} />)

      // Swedish locale formatting with timezone
      expect(screen.getByText(/15 maj 2023.*14:30:00.*\+02:00$/)).toBeInTheDocument()
    })

    test('uses Japanese (ja) translation language', () => {
      mockUseTranslationWithLanguage('ja')

      const media = createMediaWithDateShot('2023-04-12T08:15:45-07:00')
      render(<ExifDetails media={media} />)

      // Japanese locale formatting with timezone
      expect(screen.getByText(/2023年4月12日.*8:15:45.*-07:00$/)).toBeInTheDocument()
    })

    test('uses Basque (eu) translation language', () => {
      mockUseTranslationWithLanguage('eu')

      const media = createMediaWithDateShot('2023-03-20T10:30:00+01:00')
      render(<ExifDetails media={media} />)

      // Basque locale formatting with timezone
      expect(screen.getByText(/2023.*mar.*20.*10:30:00.*\+01:00$/)).toBeInTheDocument()
    })

    test('handles invalid translation language gracefully', () => {
      mockUseTranslationWithLanguage('invalid-locale')

      const media = createMediaWithDateShot('2023-05-15T14:30:00+01:00')
      render(<ExifDetails media={media} />)

      // Should fall back gracefully and still format the date
      expect(screen.getByText('Date shot')).toBeInTheDocument()
      expect(screen.getByText(/May 15, 2023.*2:30:00 PM.*\+01:00$/)).toBeInTheDocument()
    })

    test('handles empty string translation language', () => {
      mockUseTranslationWithLanguage('')

      const media = createMediaWithDateShot('2023-01-15T12:00:00-05:00')
      render(<ExifDetails media={media} />)

      // Should handle empty string gracefully
      expect(screen.getByText('Date shot')).toBeInTheDocument()
      expect(screen.getByText(/Jan 15, 2023.*12:00:00 PM.*-05:00$/)).toBeInTheDocument()
    })

    test('uses Russian (ru) translation language', () => {
      mockUseTranslationWithLanguage('ru')

      const media = createMediaWithDateShot('2023-02-23T12:15:30+03:00')
      render(<ExifDetails media={media} />)

      // Russian locale formatting with timezone
      expect(screen.getByText(/23 февр\. 2023.*12:15:30.*\+03:00$/)).toBeInTheDocument()
    })

    test('uses Traditional Chinese Hong Kong (zh-HK) translation language', () => {
      mockUseTranslationWithLanguage('zh-HK')

      const media = createMediaWithDateShot('2023-10-01T15:30:00+08:00')
      render(<ExifDetails media={media} />)

      // Traditional Chinese HK locale formatting with timezone
      expect(screen.getByText(/2023年10月1日 下午3:30:00 \+08:00$/)).toBeInTheDocument()
    })

    test('uses Traditional Chinese Taiwan (zh-TW) translation language', () => {
      mockUseTranslationWithLanguage('zh-TW')

      const media = createMediaWithDateShot('2023-10-10T16:45:00+08:00')
      render(<ExifDetails media={media} />)

      // Traditional Chinese TW locale formatting with timezone
      expect(screen.getByText(/2023年10月10日 下午4:45:00 \+08:00$/)).toBeInTheDocument()
    })

    test('uses Simplified Chinese (zh-CN) translation language', () => {
      mockUseTranslationWithLanguage('zh-CN')

      const media = createMediaWithDateShot('2023-10-01T15:30:00+08:00')
      render(<ExifDetails media={media} />)

      // Simplified Chinese locale formatting with timezone
      expect(screen.getByText(/2023年10月1日.*15:30:00.*\+08:00$/)).toBeInTheDocument()
    })

    test('uses Turkish (tr) translation language', () => {
      mockUseTranslationWithLanguage('tr')

      const media = createMediaWithDateShot('2023-05-19T13:45:25+03:00')
      render(<ExifDetails media={media} />)

      // Turkish locale formatting with timezone
      expect(screen.getByText(/19 May 2023.*13:45:25.*\+03:00$/)).toBeInTheDocument()
    })
  })

  describe('Timezone detection edge cases', () => {
    test('detects various timezone formats correctly with regex', () => {
      const testCases = [
        { input: '2023-01-01T12:00:00+0000', shouldHaveTz: true }, // +HHMM format
        { input: '2023-01-01T12:00:00-05:00', shouldHaveTz: true }, // -HH:MM format
        { input: '2023-01-01T12:00:00Z', shouldHaveTz: true }, // Z format
        { input: '2023-01-01T12:00:00z', shouldHaveTz: true }, // lowercase z
        { input: '2023-01-01T12:00:00+1400', shouldHaveTz: true }, // +HHMM max offset
        { input: '2023-01-01T12:00:00', shouldHaveTz: false }, // No timezone
        { input: '2023-01-01T12:00:00.123', shouldHaveTz: false }, // Milliseconds, no timezone
        { input: '2023-01-01T12:00:00.456789', shouldHaveTz: false }, // Microseconds, no timezone
      ]

      testCases.forEach(({ input, shouldHaveTz }, index) => {
        const media = createMediaWithDateShot(input)
        const { rerender } = render(<ExifDetails media={media} />)

        if (shouldHaveTz) {
          expect(screen.getByText(/[+-]\d{2}:\d{2}$/)).toBeInTheDocument()
        } else {
          expect(screen.getByText(/^\w+ \d+, \d+ \d+:\d+:\d+ [AP]M$/)).toBeInTheDocument()
          expect(screen.queryByText(/[+-]\d{2}:\d{2}/)).not.toBeInTheDocument()
        }

        if (index < testCases.length - 1) {
          rerender(<div />)
        }
      })
    })

    test('handles edge case timezone offsets', () => {
      const testCases = [
        { input: '2023-01-01T12:00:00+14:00', expected: /Jan 1, 2023.*12:00:00 PM.*\+14:00$/ }, // Max positive
        { input: '2023-01-01T12:00:00-12:00', expected: /Jan 1, 2023.*12:00:00 PM.*-12:00$/ },  // Max negative
        { input: '2023-01-01T12:00:00+05:45', expected: /Jan 1, 2023.*12:00:00 PM.*\+05:45$/ }, // Nepal time
        { input: '2023-01-01T12:00:00+09:30', expected: /Jan 1, 2023.*12:00:00 PM.*\+09:30$/ }, // Adelaide
      ]

      testCases.forEach(({ input, expected }, index) => {
        const media = createMediaWithDateShot(input)
        const { rerender } = render(<ExifDetails media={media} />)

        expect(screen.getByText(expected)).toBeInTheDocument()

        if (index < testCases.length - 1) {
          rerender(<div />)
        }
      })
    })
  })

  describe('Error handling and robustness', () => {
    test('handles invalid dateShot gracefully', () => {
      const media = createMediaWithDateShot('invalid-date-string')
      render(<ExifDetails media={media} />)

      // Should display the original invalid string when parsing fails
      expect(screen.getByText('invalid-date-string')).toBeInTheDocument()
    })

    test('handles malformed but parseable dates', () => {
      // Test dates that might be malformed but still parseable
      const testCases = [
        '2023-02-29T10:00:00', // Invalid leap year date
        '2023-13-01T10:00:00', // Invalid month
        '2023-01-32T10:00:00', // Invalid day
      ]

      testCases.forEach((input, index) => {
        const media = createMediaWithDateShot(input)
        const { rerender } = render(<ExifDetails media={media} />)

        // Should show original string if invalid
        expect(screen.getByText('Date shot')).toBeInTheDocument()
        expect(screen.getByText(input)).toBeInTheDocument()

        if (index < testCases.length - 1) {
          rerender(<div />)
        }
      })
    })
  })

  describe('Format consistency validation', () => {
    test('always follows consistent format structure', () => {
      const testCases = [
        { input: '2023-04-10T15:20:25-07:00', pattern: /^Apr 10, 2023 3:20:25 PM -07:00$/ },
        { input: '2023-04-10T15:20:25', pattern: /^Apr 10, 2023 3:20:25 PM$/ },
        { input: '2023-04-10T15:20:25Z', pattern: /^Apr 10, 2023 3:20:25 PM \+00:00$/ },
      ]

      testCases.forEach(({ input, pattern }, index) => {
        const media = createMediaWithDateShot(input)
        const { rerender } = render(<ExifDetails media={media} />)

        expect(screen.getByText(pattern)).toBeInTheDocument()

        if (index < testCases.length - 1) {
          rerender(<div />)
        }
      })
    })
  })
})
