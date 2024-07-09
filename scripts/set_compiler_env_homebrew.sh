#!/bin/sh

export CPLUS_INCLUDE_PATH="$(brew --prefix)/opt/jpeg/include:$(brew --prefix)/opt/dlib/include"
export LD_LIBRARY_PATH="$(brew --prefix)/opt/jpeg/lib:$(brew --prefix)/opt/dlib/lib"
export LIBRARY_PATH="$(brew --prefix)/opt/jpeg/lib:$(brew --prefix)/opt/dlib/lib"
