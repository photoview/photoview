import { TranslationFn } from '../localization'
import { AlbumPermissionLevel } from '../__generated__/globalTypes'

export const albumPermissionLevelOptions = (t: TranslationFn) => [
  {
    value: AlbumPermissionLevel.READ,
    label: t('album_permission_level.read', 'Read'),
  },
  {
    value: AlbumPermissionLevel.UPLOAD,
    label: t('album_permission_level.upload', 'Read + upload'),
  },
  {
    value: AlbumPermissionLevel.DELETE,
    label: t('album_permission_level.delete', 'Read + upload + delete'),
  },
]
