#!/bin/bash
cd /src/ui
npm install
npm start & 
cd /src/api
go install
air
tail -f /dev/null
