export CPLUS_INCLUDE_PATH="$(brew --prefix)/opt/jpeg/include:$(brew --prefix)/opt/dlib/include:${CPLUS_INCLUDE_PATH:-}"
export C_INCLUDE_PATH="$(brew --prefix)/opt/libmagic/include:$(brew --prefix)/opt/libheif/include:${C_INCLUDE_PATH:-}"
export DYLD_LIBRARY_PATH="$(brew --prefix)/opt/jpeg/lib:$(brew --prefix)/opt/dlib/lib:$(brew --prefix)/opt/libmagic/lib:$(brew --prefix)/opt/libheif/lib:${DYLD_LIBRARY_PATH:-}"
export LIBRARY_PATH="$(brew --prefix)/opt/jpeg/lib:$(brew --prefix)/opt/dlib/lib:$(brew --prefix)/opt/libmagic/lib:$(brew --prefix)/opt/libheif/lib:${LIBRARY_PATH:-}"
export CGO_CFLAGS_ALLOW=-Xpreprocessor