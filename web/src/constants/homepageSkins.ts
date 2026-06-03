export type HomepageSkinKey = 'aurora' | 'sunset' | 'galaxy' | 'mint' | 'peach'

export interface HomepageSkinDefinition {
  key: HomepageSkinKey
  label: string
  previewClass: string
  surfaceClass: string
  accentClass: string
}

export const homepageSkins: HomepageSkinDefinition[] = [
  { key: 'aurora', label: '极光', previewClass: 'skin-aurora', surfaceClass: 'theme-aurora', accentClass: 'theme-accent-aurora' },
  { key: 'sunset', label: '日落', previewClass: 'skin-sunset', surfaceClass: 'theme-sunset', accentClass: 'theme-accent-sunset' },
  { key: 'galaxy', label: '星河', previewClass: 'skin-galaxy', surfaceClass: 'theme-galaxy', accentClass: 'theme-accent-galaxy' },
  { key: 'mint', label: '薄荷', previewClass: 'skin-mint', surfaceClass: 'theme-mint', accentClass: 'theme-accent-mint' },
  { key: 'peach', label: '蜜桃', previewClass: 'skin-peach', surfaceClass: 'theme-peach', accentClass: 'theme-accent-peach' }
]

export function resolveHomepageSkin(key?: string): HomepageSkinDefinition {
  return homepageSkins.find((item) => item.key === key) || homepageSkins[0]
}
