declare module '*.svg' {
  const ReactComponent: React.FC<React.SVGProps<SVGSVGElement>>
  // const content: string

  export { ReactComponent }
  // export default content
}

interface ImportMetaEnv {
  readonly REACT_APP_BUILD_VERSION: string | undefined
  readonly REACT_APP_BUILD_DATE: string | undefined
  readonly REACT_APP_BUILD_COMMIT_SHA: string | undefined
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

type Time = string
type Any = object
