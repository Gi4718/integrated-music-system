#!/bin/sh
docker exec endfield-music wget -qO- 'http://127.0.0.1:3000/playlist/detail?id=2829816518' > /tmp/pl.json
docker exec endfield-music node -e 'var d=JSON.parse(require("fs").readFileSync("/tmp/pl.json","utf8"));var p=d.playlist;console.log("trackCount:",p.trackCount);console.log("tracks:",p.tracks?p.tracks.length:"MISSING");console.log("trackIds:",p.trackIds?p.trackIds.length:"MISSING");'
