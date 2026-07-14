const fs = require('fs');
const http = require('http');

const data = fs.readFileSync('/tmp/pl.json', 'utf8');
const d = JSON.parse(data);
const p = d.playlist;
console.log('trackCount:', p.trackCount);
console.log('has tracks:', !!p.tracks);
console.log('tracks len:', p.tracks ? p.tracks.length : 0);
console.log('has trackIds:', !!p.trackIds);
console.log('trackIds len:', p.trackIds ? p.trackIds.length : 0);
if (p.tracks && p.tracks.length > 0) {
  console.log('first track name:', p.tracks[0].name);
}
if (p.trackIds && p.trackIds.length > 0) {
  console.log('first trackId:', JSON.stringify(p.trackIds[0]));
}
